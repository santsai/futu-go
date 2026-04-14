package futu

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/santsai/futu-go/pb"
	"google.golang.org/protobuf/proto"
)

// Client is the client to connect to Futu OpenD.
type Client struct {
	clientOptions

	conn     net.Conn
	sn       uint32         // packet serial number
	respChan chan *response // response channel
	reqChan  chan *request  // request channel
	closed   chan struct{}  // indicate the client is closed
	connID   uint64
	userID   uint64

	wgWriter sync.WaitGroup
	wgReader sync.WaitGroup
	wgWorker sync.WaitGroup

	*cipherManager

	*dispatcher
}

// New creates a new client.
func NewClient(opts ...ClientOption) (*Client, error) {

	client := &Client{
		clientOptions: newClientOptions(opts),
		closed:        make(chan struct{}),
		dispatcher:    newDispatcher(),
	}

	client.reqChan = make(chan *request, client.numBuffers)
	client.respChan = make(chan *response, client.numBuffers)

	var err error

	// setup cipher manager
	client.cipherManager, err = newCipherManager(client.privateKey)
	if err != nil {
		return nil, err
	}

	// connect
	client.conn, err = net.Dial("tcp", client.openDAddr)
	if err != nil {
		client.conn = nil
		err = fmt.Errorf("dial error: %w", err)
		return nil, err
	}

	// spawn workers
	for i := 0; i < client.numWorkers; i++ {
		client.wgWorker.Add(1)
		go client.respWorker()
	}

	// resp reading
	client.wgReader.Add(1)
	go client.respReadLoop()

	// req writing
	client.wgWriter.Add(1)
	go client.reqWriteLoop()

	s2c, err := client.initConnect()
	if err != nil {
		client.Close()
		err = fmt.Errorf("initConnect error: %w", err)
		return nil, err
	}

	log.Info().
		Int32("server_ver", s2c.GetServerVer()).
		Uint64("conn_id", s2c.GetConnID()).
		Uint64("user_id", s2c.GetLoginUserID()).
		Int32("keep_alive_interval", s2c.GetKeepAliveInterval()).
		Str("user_attr", s2c.GetUserAttribution().String()).
		Str("conn_aes_key", s2c.GetConnAESKey()).
		Str("aes_cbc_iv", s2c.GetAesCBCiv()).
		Msg("init connect success")

	client.connID = s2c.GetConnID()
	client.userID = s2c.GetLoginUserID()

	if client.privateKey != nil {
		key := []byte(s2c.GetConnAESKey())
		iv := []byte(s2c.GetAesCBCiv())
		err = client.UpdateAES(key, iv)
		if err != nil {
			client.Close()
			return nil, err
		}
	}

	if interval := s2c.GetKeepAliveInterval(); interval > 0 {
		client.wgWorker.Add(1)
		go client.heartbeat(time.Second * time.Duration(interval))
	}

	return client, nil
}

func (client *Client) nextTradePacketId() *pb.PacketID {
	return &pb.PacketID{
		ConnID:   proto.Uint64(client.connID),
		SerialNo: proto.Uint32(client.nextSN()),
	}
}

// Close closes the client.
func (client *Client) Close() error {

	var err error = nil

	if client.conn != nil {
		err = client.conn.Close()
		client.conn = nil
		client.wgReader.Wait()
	}

	log.Info().Msg("read loop exited")

	close(client.closed)
	client.wgWorker.Wait()
	client.wgWriter.Wait()
	log.Info().Msg("worker & heartbeat exited")

	client.dispatchClose()

	return err
}

func (client *Client) patchRequest(req pb.Request) {

	payload := req.GetRequestPayload()

	// UserID is no longer needed. but is required in proto.
	if setter, ok := payload.(pb.UserIDSetter); ok {
		setter.SetUserID(client.userID)
	}

	// avoid replay attacks
	if setter, ok := payload.(pb.PacketIDSetter); ok {
		setter.SetPacketID(client.nextTradePacketId())
	}
}

func (client *Client) encodeRequest(protoId pb.ProtoId, req pb.Request) (*bytes.Buffer, uint32, error) {

	// fill in required infomation
	client.patchRequest(req)

	body, err := proto.Marshal(req)
	if err != nil {
		return nil, 0, err
	}

	sha1Value := sha1.Sum(body)

	body, err = client.Encrypt(protoId, body)
	if err != nil {
		return nil, 0, err
	}

	sn := client.nextSN()

	h := futuHeader{
		HeaderFlag:   [2]byte{'F', 'T'},
		ProtoID:      protoId,
		ProtoFmtType: 0,
		ProtoVer:     0,
		SerialNo:     sn,
		BodyLen:      uint32(len(body)),
		BodySHA1:     sha1Value,
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, &h); err != nil {
		return nil, 0, err
	}

	if _, err := buf.Write(body); err != nil {
		return nil, 0, err
	}

	return &buf, sn, nil
}

func (client *Client) writeRequest(r *request) {
	var (
		buf *bytes.Buffer
		sn  uint32
		err error
	)

	// encode
	if buf, sn, err = client.encodeRequest(r.protoId, r.req); err != nil {
		r.respC <- &response{Err: err}
		return
	}

	item := &dispatchItem{
		c:    r.respC,
		resp: r.resp,
	}
	client.dispatchPut(r.protoId, sn, item)

	// write to connection
	if _, err = buf.WriteTo(client.conn); err != nil {
		client.dispatchPop(r.protoId, sn)
		r.respC <- &response{Err: err}
		return
	}
}

func (client *Client) reqWriteLoop() {

	defer client.wgWriter.Done()

	for {
		select {
		case <-client.closed:
			return
		case r := <-client.reqChan:
			client.writeRequest(r)
		}
	}

}

func (client *Client) Request(ctx context.Context, protoId pb.ProtoId, req pb.Request, resp pb.Response) (proto.Message, error) {

	respC := make(chan *response, 1)
	client.reqChan <- &request{
		respC:   respC,
		protoId: protoId,
		req:     req,
		resp:    resp,
	}

	// add timeout to context if not exist.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, client.timeout)
		defer cancel()
	}

	// wait response
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-client.closed:
		return nil, ErrInterrupted
	case rr, ok := <-respC:
		if !ok {
			return nil, ErrChannelClosed
		}

		if rr.Err != nil {
			return nil, rr.Err
		}

		return rr.Resp.GetResponsePayload(), ResponseError(protoId, rr.Resp)
	}
}

// RegisterHandler registers a handler for notifications of a specified protoID.
func (client *Client) RegisterHandler(protoID pb.ProtoId, h Handler) *Client {
	client.registerHandler(protoID, h)
	return client
}

// nextSN returns the next serial number.
func (client *Client) nextSN() uint32 {
	client.sn += 1
	return client.sn
}

func (client *Client) respWork(r *response) {

	defer func() {
		// things can happen during proto unmarshal
		if r := recover(); r != nil {
			log.Error().Interface("recover", r).Msg("panic recovered in respWork")
		}
	}()

	// decrypt body
	if body, err := client.Decrypt(r.ProtoID, r.Body); err != nil {
		r.Err = err
	} else {
		r.Body = body
		r.Encrypted = false
	}

	// verify body
	if r.Err == nil {
		ssum := sha1.Sum(r.Body)
		if !bytes.Equal(r.BodySHA1, ssum[:]) {
			r.Err = errSHA1Mismatch
		}
	}

	// get dispatchItem
	ditem := client.dispatchPop(r.ProtoID, r.SerialNo)
	if ditem == nil {
		// no dispatchItem registered
		// dont know how to unmarshal. break.
		log.Error().Uint32("protoId", uint32(r.ProtoID)).
			Uint32("serialNo", r.SerialNo).
			Msg("no unmarshal target")

		return
	}

	// proto decode
	if r.Err == nil {
		r.Err = proto.Unmarshal(r.Body, ditem.resp)

		if r.Err == nil {
			r.Resp = ditem.resp
		}
	}

	// dispatch
	if ditem.c != nil {
		ditem.c <- r
		close(ditem.c)

	} else {
		if r.Err == nil {
			h := client.getHandler(r.ProtoID)
			h(r.Resp.GetResponsePayload())
		} else {
			log.Error().Err(r.Err).Msg("push decrypt/decode error ignored")
		}
	}

}

func (client *Client) respWorker() {

	defer func() {
		log.Info().Msg("worker exit")
		client.wgWorker.Done()
	}()

	for {
		select {
		case <-client.closed:
			return

		case r := <-client.respChan:

			log.Info().Stringer("protoId", r.ProtoID).Uint32("sn", r.SerialNo).Msg("respWorker")
			client.respWork(r)

		}
	}
}

func (client *Client) respRead() error {
	// read header, it will block until the header is read
	var h futuHeader
	if err := binary.Read(client.conn, binary.LittleEndian, &h); err != nil {
		return err
	}
	if h.HeaderFlag != [2]byte{'F', 'T'} {
		return errors.New("header flag error")
	}

	// read body, it will block until the body is read
	b := make([]byte, h.BodyLen)
	if _, err := io.ReadFull(client.conn, b); err != nil {
		return err
	}

	resp := &response{
		ProtoID:   h.ProtoID,
		SerialNo:  h.SerialNo,
		BodySHA1:  h.BodySHA1[:],
		Body:      b,
		Encrypted: client.privateKey != nil,
	}

	client.respChan <- resp

	return nil
}

func (client *Client) respReadLoop() {

	defer client.wgReader.Done()

	for {
		var err error
		if err = client.respRead(); err == nil {
			continue
		}

		// EOF
		if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
			// If the connection is closed, stop receiving data.
			// io.EOF: The connection is closed by the remote end.
			// net.ErrClosed: The connection is closed by the local end.
			log.Error().Err(err).Msg("respRead: conn closed")
			break
		}

		// XXX ignore other non-fatal? errors
		// XXX should ignore or not? how to test?
		log.Error().Err(err).Msg("respRead: unknown error")
	}
}

func (client *Client) initConnect() (*pb.InitConnectResponse, error) {
	req := &pb.InitConnectRequest{
		ClientVer:           proto.Int32(kClientVersion),
		ClientID:            proto.String(client.clientId),
		RecvNotify:          proto.Bool(client.recvNotify),
		PacketEncAlgo:       pb.PacketEncAlgo_AES_CBC.Enum(),
		ProgrammingLanguage: proto.String("Go"),
	}

	return req.Dispatch(context.TODO(), client)
}

// XXX disconnect/missed ping? handling?
func (client *Client) heartbeat(d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	defer client.wgWorker.Done()

	// take the smaller timeout
	timeout := d
	if timeout > client.timeout {
		timeout = client.timeout
	}

	for {
		select {
		case <-client.closed:
			log.Info().Msg("heartbeat stopped")
			return

		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.TODO(), timeout)
			req := &pb.KeepAliveRequest{
				Time: proto.Int64(time.Now().Unix()),
			}

			_, err := req.Dispatch(ctx, client)
			cancel()
			// XXX is this non-fatal?
			if err != nil {
				log.Error().Err(err).Msg("heartbeat error")
			}
		}
	}
}
