package futu

import (
	"github.com/rs/zerolog/log"
	"github.com/santsai/futu-go/pb"
	"google.golang.org/protobuf/proto"
)

// Handler is the definition of a handler function.
type Handler func(s2c proto.Message) error

func defaultHandler(s2c proto.Message) error {
	log.Info().Interface("s2c", s2c).Msg("notification (no handler)")
	return nil
}

type dispatchItem struct {
	c    chan *response // c is nil for push
	resp pb.Response
}

func makeDispatchId(protoId pb.ProtoId, serialNo uint32) uint64 {
	return (uint64(protoId) << 32) | uint64(serialNo)
}

// RegisterHandler registers a handler for notifications of a specified protoID.
func (client *Client) RegisterHandler(protoID pb.ProtoId, h Handler) *Client {
	client.dispatchMutex.Lock()
	client.handlers[protoID] = h
	client.dispatchMutex.Unlock()
	return client
}

func (client *Client) getHandler(protoID pb.ProtoId) Handler {
	rh := defaultHandler

	client.dispatchMutex.Lock()
	if h, ok := client.handlers[protoID]; ok {
		rh = h
	}
	client.dispatchMutex.Unlock()

	return rh
}

func (client *Client) dispatchPut(protoId pb.ProtoId, sn uint32, ditem *dispatchItem) {
	id := makeDispatchId(protoId, sn)

	client.dispatchMutex.Lock()
	client.dispatchMap[id] = ditem
	client.dispatchMutex.Unlock()
}

func (client *Client) dispatchPop(protoId pb.ProtoId, sn uint32) *dispatchItem {
	id := makeDispatchId(protoId, sn)

	client.dispatchMutex.Lock()
	ditem, ok := client.dispatchMap[id]
	if ok {
		delete(client.dispatchMap, id)
	}
	client.dispatchMutex.Unlock()

	// handle push data.
	if ditem == nil {
		if resp := pb.GetPushResponseStruct(protoId); resp != nil {
			ditem = &dispatchItem{resp: resp}
		}
	}

	return ditem
}

func (client *Client) dispatchClose() {
	client.dispatchMutex.Lock()
	for id, ditem := range client.dispatchMap {
		close(ditem.c)
		delete(client.dispatchMap, id)
	}
	client.dispatchMutex.Unlock()
}
