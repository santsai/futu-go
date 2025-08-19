package client

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

	"github.com/hyperjiang/rsa"
	"github.com/rs/zerolog/log"
	"github.com/santsai/futu-go/infra"
	"github.com/santsai/futu-go/pb"
	"google.golang.org/protobuf/proto"
)

type bodyChanType chan<- []byte

// Client is the client to connect to Futu OpenD.
type Client struct {
	Options

	conn     net.Conn
	sn       atomic.Uint32 // serial number
	resChan  chan response // response channel
	pushChan chan response // push update channel
	closed   chan struct{} // indicate the client is closed
	connID   uint64
	userID   uint64
	crypto   *infra.Crypto
	handlers sync.Map // push notification handlers

	bodyChan      map[uint64]bodyChanType
	bodyChanMutex sync.Mutex
}

// New creates a new client.
func New(opts ...Option) (*Client, error) {
	client := &Client{
		Options:  NewOptions(opts...),
		closed:   make(chan struct{}),
		bodyChan: map[uint64]bodyChanType{},
		handlers: sync.Map{},
	}

	client.resChan = make(chan response, client.ResChanSize)
	client.pushChan = make(chan response, client.ResChanSize)

	if err := client.dial(); err != nil {
		return nil, err
	}

	go client.listen()
	go client.infiniteRead()

	s2c, err := client.initConnect()
	if err != nil {
		client.Close()
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

	if client.PrivateKey != nil || client.PublicKey != nil {
		client.crypto, err = infra.NewCrypto([]byte(s2c.GetConnAESKey()), []byte(s2c.GetAesCBCiv()))
		if err != nil {
			client.Close()
			return nil, err
		}
	}

	if interval := s2c.GetKeepAliveInterval(); interval > 0 {
		go client.heartbeat(time.Second * time.Duration(interval))
	}

	go client.watchNotification()

	return client, nil
}

// GetConnID returns the connection ID.
func (client *Client) GetConnID() uint64 {
	return client.connID
}

// GetUserID returns the user ID.
func (client *Client) GetUserID() uint64 {
	return client.userID
}

// XXX this is temp
func (client *Client) GetTradePacketId() *pb.PacketID {
	return &pb.PacketID{
		ConnID:   proto.Uint64(client.connID),
		SerialNo: proto.Uint32(client.nextSN()),
	}
}

// Close closes the client.
func (client *Client) Close() error {
	close(client.closed)

	client.dispatcherClose()

	if client.conn == nil {
		return nil
	}

	return client.conn.Close()
}

// Request sends a request to the server.
func (client *Client) makeRequest(protoID pb.ProtoId, req proto.Message, bch chan<- []byte) error {
	var buf bytes.Buffer

	b, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	sha1Value := sha1.Sum(b)

	if client.PublicKey != nil {
		if protoID == pb.ProtoId_InitConnect {
			b, err = rsa.EncryptByPublicKey(b, client.PublicKey)
			if err != nil {
				return err
			}
		} else {
			b = client.crypto.Encrypt(b)
		}
	}

	sn := client.nextSN()

	h := futuHeader{
		HeaderFlag:   [2]byte{'F', 'T'},
		ProtoID:      uint32(protoID),
		ProtoFmtType: 0,
		ProtoVer:     0,
		SerialNo:     sn,
		BodyLen:      uint32(len(b)),
		BodySHA1:     sha1Value,
	}

	client.dispatcherRegister(protoID, sn, bch)

	if err := binary.Write(&buf, binary.LittleEndian, &h); err != nil {
		return err
	}

	if _, err := buf.Write(b); err != nil {
		return err
	}

	_, err = buf.WriteTo(client.conn)

	return err
}

func (client *Client) Request(ctx context.Context, id pb.ProtoId, req proto.Message, resp pb.Response) (proto.Message, error) {

	ch := make(chan []byte, 1)
	defer close(ch)

	if err := client.makeRequest(id, req, ch); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-client.closed:
		return nil, ErrInterrupted
	case bs, ok := <-ch:
		if !ok {
			return nil, ErrChannelClosed
		}

		if err := proto.Unmarshal(bs, resp); err != nil {
			return nil, err
		}

		return resp.GetResponsePayload(), pb.ResponseError(resp)
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

// watchNotification watches the push notification.
// no need to close the channels in this function,
// because they will be closed by the dispatcher hub when the client is closed.
func (client *Client) watchNotification() {

	for {
		select {
		case <-client.closed:
			log.Info().Msg("stop watching notification")
			return

		case resp, ok := <-client.pushChan:
			if !ok {
				log.Info().Msg("notification channel closed")
				break
			}

			if err := client.dispatcherPushCall(resp); err != nil {
				log.Error().Err(err).Msg("notification handle error")
			}
		}
	}
}

func (client *Client) dispatcherRegister(protoId pb.ProtoId, sn uint32, ch bodyChanType) {
	// XXX
	var id uint64 = uint64(protoId)<<32 | uint64(sn)

	client.bodyChanMutex.Lock()
	client.bodyChan[id] = ch
	client.bodyChanMutex.Unlock()
}

func (client *Client) dispatcherCall(res response) error {
	// XXX
	var respId = res.respId()

	client.bodyChanMutex.Lock()
	ch, ok := client.bodyChan[respId]
	if ok {
		delete(client.bodyChan, respId)
	}
	client.bodyChanMutex.Unlock()

	if ok {
		ch <- res.Body
		return nil
	}

	if res.SerialNo == 0 {
		client.pushChan <- res
		return nil
	}

	return fmt.Errorf("unexpected dispatch error")
}

func (client *Client) dispatcherPushCall(resp response) error {

	pbResp := pb.GetPushResponseStruct(resp.ProtoID)
	if pbResp == nil {
		return fmt.Errorf("cannot find a response struct for id: %d", resp.ProtoID)
	}

	if err := proto.Unmarshal(resp.Body, pbResp); err != nil {
		return err
	}

	h := client.getHandler(resp.ProtoID)
	h(pbResp.GetResponsePayload())
	return nil
}

func (client *Client) dispatcherClose() {
	client.bodyChanMutex.Lock()
	for mid, ch := range client.bodyChan {
		close(ch)
		delete(client.bodyChan, mid)
	}
	client.bodyChanMutex.Unlock()
}

// nextSN returns the next serial number.
func (client *Client) nextSN() uint32 {
	return client.sn.Add(1)
}

func (client *Client) dial() error {
	conn, err := net.Dial("tcp", client.Addr)
	if err != nil {
		log.Error().Err(err).Msg("dial failed")
		return err
	}

	client.conn = conn

	return nil
}

func (client *Client) listen() {
	for {
		select {
		case <-client.closed:
			return
		case res := <-client.resChan:
			log.Info().Uint32("proto_id", uint32(res.ProtoID)).Uint32("sn", res.SerialNo).Msg("listen")
			if err := client.dispatcherCall(res); err != nil {
				log.Error().Err(err).Msg("dispatch error")
			}
		}
	}
}

func (client *Client) read() error {
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

	if client.PrivateKey != nil {
		if pb.ProtoId(h.ProtoID) == pb.ProtoId_InitConnect {
			var err error
			b, err = rsa.DecryptByPrivateKey(b, client.PrivateKey)
			if err != nil {
				return err
			}
		} else {
			b = client.crypto.Decrypt(b)
		}
	}

	// verify body
	if h.BodySHA1 != sha1.Sum(b) {
		return errors.New("sha1 sum error")
	}

	res := response{
		ProtoID:  pb.ProtoId(h.ProtoID),
		SerialNo: h.SerialNo,
		Body:     b,
	}

	client.resChan <- res

	return nil
}

func (client *Client) infiniteRead() {
	for {
		if err := client.read(); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				// If the connection is closed, stop receiving data.
				// io.EOF: The connection is closed by the remote end.
				// net.ErrClosed: The connection is closed by the local end.
				log.Error().Err(err).Msg("connection closed")
				return
			} else {
				log.Error().Err(err).Msg("decode error")
				return
			}
		}
	}
}

func (client *Client) initConnect() (*pb.InitConnectResponse, error) {
	req := &pb.InitConnectRequest{
		ClientVer:           proto.Int32(ClientVersion),
		ClientID:            proto.String(client.ID),
		RecvNotify:          proto.Bool(client.RecvNotify),
		PacketEncAlgo:       pb.PacketEncAlgo_AES_CBC.Enum(),
		ProgrammingLanguage: proto.String("Go"),
	}

	ctx, cancel := context.WithTimeout(context.TODO(), client.Timeout)
	defer cancel()

	return req.MakeRequest(ctx, client)
}

// XXX disconnect handling
func (client *Client) heartbeat(d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-client.closed:
			log.Info().Msg("heartbeat stopped")
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.TODO(), d)
			req := &pb.KeepAliveRequest{
				Time: proto.Int64(time.Now().Unix()),
			}

			_, err := req.MakeRequest(ctx, client)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
