package futu

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/santsai/futu-go/pb"
	"google.golang.org/protobuf/proto"
)

type protocolReader interface {
	readFrom(r io.Reader) error
}

type protocolHeader struct {
	HeaderFlag   [2]byte    // Packet header start flag, fixed as "FT"
	ProtoID      pb.ProtoId // Protocol ID
	ProtoFmtType uint8      // Protocol type, 0 for Protobuf, 1 for Json
	ProtoVer     uint8      // Protocol version, used for iterative compatibility, currently 0
	SerialNo     uint32     // Packet serial number, used to correspond to the request packet and return packet, and it is required to be incremented
	BodyLen      uint32     // Body length
	BodySHA1     [20]byte   // SHA1 hash value of the original data of the packet body (after decryption)
	Reserved     [8]byte    // Reserved 8-byte extension
}

// protocol is for frame codec and read/write
type protocol struct {
	*conn

	*cipherManager

	connID uint64
	userID uint64

	serialNo uint32 // packet serial number

	respC chan<- *response
}

func newProtocol(address string, privateKey []byte, respC chan<- *response) (*protocol, error) {

	p := &protocol{}

	var err error

	// newCipherManager
	if p.cipherManager, err = newCipherManager(privateKey); err != nil {
		return nil, fmt.Errorf("newCipherManager error: %w", err)
	}

	// newConn
	if p.conn, err = newConn(address, p); err != nil {
		return nil, fmt.Errorf("newConn error: %w", err)
	}

	p.respC = respC

	return p, nil
}

// nextSN returns the next serial number.
func (p *protocol) nextSerialNo() uint32 {
	p.serialNo += 1
	return p.serialNo
}

func (p *protocol) nextTradePacketId() *pb.PacketID {
	return &pb.PacketID{
		ConnID:   proto.Uint64(p.connID),
		SerialNo: proto.Uint32(p.nextSerialNo()),
	}
}

func (p *protocol) patchRequest(req pb.Request) {

	payload := req.GetRequestPayload()

	// UserID is no longer needed. but is required in proto.
	if setter, ok := payload.(pb.UserIDSetter); ok {
		setter.SetUserID(p.userID)
	}

	// avoid replay attacks
	if setter, ok := payload.(pb.PacketIDSetter); ok {
		setter.SetPacketID(p.nextTradePacketId())
	}
}

func (p *protocol) prepareRequest(r *request) {
	p.patchRequest(r.req)
	r.serialNo = p.nextSerialNo()
}

func (p *protocol) writeRequest(r *request) error {

	body, err := proto.Marshal(r.req)
	if err != nil {
		return err
	}

	bodySum := sha1.Sum(body)

	encryptedBody, err := p.Encrypt(r.protoId, body)
	if err != nil {
		return err
	}

	h := protocolHeader{
		HeaderFlag:   [2]byte{'F', 'T'},
		ProtoID:      r.protoId,
		ProtoFmtType: 0,
		ProtoVer:     0,
		SerialNo:     r.serialNo,
		BodyLen:      uint32(len(encryptedBody)),
		BodySHA1:     bodySum,
	}

	if err := binary.Write(p.conn, binary.LittleEndian, &h); err != nil {
		return err
	}

	if _, err := p.Write(encryptedBody); err != nil {
		return err
	}

	return nil
}

func (p *protocol) updateWithInitResponse(r *pb.InitConnectResponse) error {

	p.connID = r.GetConnID()
	p.userID = r.GetLoginUserID()

	key := []byte(r.GetConnAESKey())
	iv := []byte(r.GetAesCBCiv())

	if err := p.UpdateAES(key, iv); err != nil {
		return err
	}

	return nil
}

func (p *protocol) readFrom(conn io.Reader) error {
	// read header, it will block until the header is read
	var h protocolHeader
	if err := binary.Read(conn, binary.LittleEndian, &h); err != nil {
		return err
	}

	if h.HeaderFlag != [2]byte{'F', 'T'} {
		return errors.New("header flag error")
	}

	// read body, it will block until the body is read
	b := make([]byte, h.BodyLen)
	if _, err := io.ReadFull(conn, b); err != nil {
		return err
	}

	resp := &response{
		ProtoID:   h.ProtoID,
		SerialNo:  h.SerialNo,
		BodySHA1:  h.BodySHA1[:],
		Body:      b,
		Encrypted: p.cipherManager != nil,
	}

	p.respC <- resp

	return nil
}
