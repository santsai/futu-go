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
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/santsai/futu-go/cipher"
	"github.com/santsai/futu-go/pb"
	"google.golang.org/protobuf/proto"
)

// Client is the client to connect to Futu OpenD.
type Client struct {
	clientOptions

	conn     net.Conn
	sn       atomic.Uint32  // serial number
	respChan chan *response // response channel
	closed   chan struct{}  // indicate the client is closed
	connID   uint64
	userID   uint64

	wgReader sync.WaitGroup
	wgWorker sync.WaitGroup

	aes *cipher.AES
	rsa *cipher.RSA

	//
	handlers      map[pb.ProtoId]Handler // push notification handlers
	dispatchMap   map[uint64]*dispatchItem
	dispatchMutex sync.Mutex
}

// New creates a new client.
func NewClient(opts ...ClientOption) (*Client, error) {

	client := &Client{
		clientOptions: newClientOptions(opts),
		closed:        make(chan struct{}),
		dispatchMap:   map[uint64]*dispatchItem{},
		handlers:      map[pb.ProtoId]Handler{},
	}

	client.respChan = make(chan *response, client.numBuffers)

	var err error

	// setup rsa
	if client.privateKey != nil {
		client.rsa, err = cipher.NewRSA(client.privateKey)
		if err != nil {
			return nil, err
		}
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

	client.wgReader.Add(1)
	go client.respReadLoop()

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
		client.aes, err = cipher.NewAES(key, iv)
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

func (client *Client) getCipher(id pb.ProtoId) cipher.Cipher {
	if client.privateKey == nil {
		return nil
	}

	if id == pb.ProtoId_InitConnect {
		return client.rsa
	}

	return client.aes
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

	if cs := client.getCipher(protoId); cs != nil {
		body, err = cs.Encrypt(body)
		if err != nil {
			return nil, 0, err
		}
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

func (client *Client) Request(ctx context.Context, protoId pb.ProtoId, req pb.Request, resp pb.Response) (proto.Message, error) {

	var (
		buf *bytes.Buffer
		sn  uint32
		err error
	)

	// encode
	if buf, sn, err = client.encodeRequest(protoId, req); err != nil {
		return nil, err
	}

	ditem := &dispatchItem{
		c:    make(chan *response, 1),
		resp: resp,
	}
	client.dispatchPut(protoId, sn, ditem)

	// write to connection
	if _, err = buf.WriteTo(client.conn); err != nil {
		client.dispatchPop(protoId, sn)
		return nil, err
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
	case rr, ok := <-ditem.c:
		if !ok {
			return nil, ErrChannelClosed
		}

		if rr.Err != nil {
			return nil, rr.Err
		}

		return rr.Resp.GetResponsePayload(), ResponseError(protoId, rr.Resp)
	}
}

// nextSN returns the next serial number.
func (client *Client) nextSN() uint32 {
	return client.sn.Add(1)
}

func (client *Client) respWork(r *response) {

	defer func() {
		// things can happen during proto unmarshal
		if r := recover(); r != nil {
			log.Error().Interface("recover", r).Msg("panic recovered in respWork")
		}
	}()

	// decrypt body
	if cs := client.getCipher(r.ProtoID); cs != nil {
		if body, err := cs.Decrypt(r.Body); err != nil {
			r.Err = err
		} else {
			r.Body = body
			r.Encrypted = false
		}
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
		ClientVer:           ProtoPtr(kClientVersion),
		ClientID:            ProtoPtr(client.clientId),
		RecvNotify:          ProtoPtr(client.recvNotify),
		PacketEncAlgo:       pb.PacketEncAlgo_AES_CBC.Enum(),
		ProgrammingLanguage: ProtoPtr("Go"),
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
