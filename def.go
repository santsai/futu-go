package futu

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/santsai/futu-go/pb"
	"google.golang.org/protobuf/proto"
)

const (
	// ClientVersion is the version of the client.
	kClientVersion int32 = 100
)

var (
	ErrChannelClosed = errors.New("channel is closed")
	ErrInterrupted   = errors.New("process is interrupted")
	ErrTimeout       = errors.New("timeout")

	errSHA1Mismatch = errors.New("sha1 mismatch")
)

type cryptoService interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type futuHeader struct {
	HeaderFlag   [2]byte    // Packet header start flag, fixed as "FT"
	ProtoID      pb.ProtoId // Protocol ID
	ProtoFmtType uint8      // Protocol type, 0 for Protobuf, 1 for Json
	ProtoVer     uint8      // Protocol version, used for iterative compatibility, currently 0
	SerialNo     uint32     // Packet serial number, used to correspond to the request packet and return packet, and it is required to be incremented
	BodyLen      uint32     // Body length
	BodySHA1     [20]byte   // SHA1 hash value of the original data of the packet body (after decryption)
	Reserved     [8]byte    // Reserved 8-byte extension
}

type response struct {
	ProtoID   pb.ProtoId
	SerialNo  uint32
	BodySHA1  []byte
	Body      []byte
	Encrypted bool
	Err       error
	Resp      pb.Response
}

type dispatchData struct {
	c    chan *response // c is nil for push
	resp pb.Response
}

func makeRespId(protoId pb.ProtoId, serialNo uint32) uint64 {
	return (uint64(protoId) << 32) | uint64(serialNo)
}

// Handler is the definition of a handler function.
type Handler func(s2c proto.Message) error

func defaultHandler(s2c proto.Message) error {
	log.Info().Interface("s2c", s2c).Msg("notification")
	return nil
}
