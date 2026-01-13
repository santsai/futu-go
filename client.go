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
	handlers sync.Map // push notification handlers

	wgReader sync.WaitGroup
	wgWorker sync.WaitGroup

	aes *AES
	rsa *RSA

	dispatchMap   map[uint64]*dispatchData
	dispatchMutex sync.Mutex
}

// New creates a new client.
func NewClient(opts ...ClientOption) (*Client, error) {

	client := &Client{
		clientOptions: newClientOptions(opts),
		closed:        make(chan struct{}),
		dispatchMap:   map[uint64]*dispatchData{},
		handlers:      sync.Map{},
	}

	client.respChan = make(chan *response, client.numBuffers)

	// setup rsa
	if client.privateKey != nil {
		rsa, err := NewRSA(client.privateKey)
		if err != nil {
			return nil, err
		}
		client.rsa = rsa
	}

	// connect
	if err := client.dial(); err != nil {
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
		client.aes, err = NewAES([]byte(s2c.GetConnAESKey()), []byte(s2c.GetAesCBCiv()))
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

func (client *Client) getCrypto(id pb.ProtoId) cryptoService {
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

	if client.conn == nil {
		err = client.conn.Close()
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

	if cs := client.getCrypto(protoId); cs != nil {
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

	dida := &dispatchData{
		c:    make(chan *response, 1),
		resp: resp,
	}
	client.dispatchPut(protoId, sn, dida)

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
	case rr, ok := <-dida.c:
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
	client.handlers.Store(protoID, h)
	return client
}

func (client *Client) getHandler(protoID pb.ProtoId) Handler {
	if h, ok := client.handlers.Load(protoID); ok {
		return h.(Handler)
	}

	return defaultHandler
}

func (client *Client) dispatchPut(protoId pb.ProtoId, sn uint32, dida *dispatchData) {
	id := makeRespId(protoId, sn)

	client.dispatchMutex.Lock()
	client.dispatchMap[id] = dida
	client.dispatchMutex.Unlock()
}

func (client *Client) dispatchPop(protoId pb.ProtoId, sn uint32) *dispatchData {
	id := makeRespId(protoId, sn)

	client.dispatchMutex.Lock()
	dida, ok := client.dispatchMap[id]
	if ok {
		delete(client.dispatchMap, id)
	}
	client.dispatchMutex.Unlock()

	// handle push data.
	if dida == nil {
		if resp := pb.GetPushResponseStruct(protoId); resp != nil {
			dida = &dispatchData{resp: resp}
		}
	}

	return dida
}

func (client *Client) dispatchClose() {
	client.dispatchMutex.Lock()
	for id, dida := range client.dispatchMap {
		close(dida.c)
		delete(client.dispatchMap, id)
	}
	client.dispatchMutex.Unlock()
}

// nextSN returns the next serial number.
func (client *Client) nextSN() uint32 {
	return client.sn.Add(1)
}

func (client *Client) dial() error {
	conn, err := net.Dial("tcp", client.openDAddr)
	if err != nil {
		log.Error().Err(err).Msg("dial failed")
		return err
	}

	client.conn = conn

	return nil
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

		case rr := <-client.respChan:

			log.Info().Stringer("protoId", rr.ProtoID).Uint32("sn", rr.SerialNo).Msg("respWorker")

			// decrypt body
			if cs := client.getCrypto(rr.ProtoID); cs != nil {
				if body, err := cs.Decrypt(rr.Body); err != nil {
					rr.Err = err
				} else {
					rr.Body = body
					rr.Encrypted = false
				}
			}

			// verify body
			if rr.Err == nil {
				ssum := sha1.Sum(rr.Body)
				if !bytes.Equal(rr.BodySHA1, ssum[:]) {
					rr.Err = errSHA1Mismatch
				}
			}

			// get dispatchData
			dida := client.dispatchPop(rr.ProtoID, rr.SerialNo)
			if dida == nil {
				// no dispatchData registered
				// dont know how to unmarshal. break.
				log.Error().Uint32("protoId", uint32(rr.ProtoID)).
					Uint32("serialNo", rr.SerialNo).
					Msg("no unmarshal target")
				break
			}

			// proto decode
			if rr.Err == nil {
				if err := proto.Unmarshal(rr.Body, dida.resp); err != nil {
					rr.Err = err
				} else {
					rr.Resp = dida.resp
				}
			}

			// dispatch
			if dida.c != nil {
				dida.c <- rr
				close(dida.c)

			} else {
				if rr.Err == nil {
					h := client.getHandler(rr.ProtoID)
					h(rr.Resp.GetResponsePayload())
				} else {
					log.Error().Err(rr.Err).Msg("push decrypt/decode error ignored")
				}
			}
		}
	}
}

func (client *Client) respRead() error {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("recover", r).Msg("")
		}
	}()

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
		if err := client.respRead(); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				// If the connection is closed, stop receiving data.
				// io.EOF: The connection is closed by the remote end.
				// net.ErrClosed: The connection is closed by the local end.
				log.Error().Err(err).Msg("connection closed")
				return
			} else {
				// XXX should not return on error!
				// how to introduce a error and test?
				log.Error().Err(err).Msg("other read error")
				return
			}
		}
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
			if err != nil {
				return
			}
		}
	}
}
