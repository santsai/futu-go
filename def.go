package futu

import (
	"github.com/santsai/futu-go/pb"
)

const (
	// ClientVersion is the version of the client.
	kClientVersion int32 = 100
)

type response struct {
	ProtoID   pb.ProtoId
	SerialNo  uint32
	BodySHA1  []byte
	Body      []byte
	Encrypted bool
	Err       error
	Resp      pb.Response
}

type request struct {
	respC    chan *response
	protoId  pb.ProtoId
	serialNo uint32
	req      pb.Request
	resp     pb.Response
}
