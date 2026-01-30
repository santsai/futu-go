package futu

import (
	"github.com/rs/zerolog/log"
	"github.com/santsai/futu-go/pb"
	"google.golang.org/protobuf/proto"
	"sync"
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

type dispatcher struct {
	handlers      map[pb.ProtoId]Handler // push notification handlers
	dispatchMap   map[uint64]*dispatchItem
	dispatchMutex sync.Mutex
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		handlers:    make(map[pb.ProtoId]Handler),
		dispatchMap: make(map[uint64]*dispatchItem),
	}
}

func makeDispatchId(protoId pb.ProtoId, serialNo uint32) uint64 {
	return (uint64(protoId) << 32) | uint64(serialNo)
}

// RegisterHandler registers a handler for notifications of a specified protoID.
func (disp *dispatcher) registerHandler(protoID pb.ProtoId, h Handler) {
	disp.dispatchMutex.Lock()
	disp.handlers[protoID] = h
	disp.dispatchMutex.Unlock()
}

func (disp *dispatcher) getHandler(protoID pb.ProtoId) Handler {
	rh := defaultHandler

	disp.dispatchMutex.Lock()
	if h, ok := disp.handlers[protoID]; ok {
		rh = h
	}
	disp.dispatchMutex.Unlock()

	return rh
}

func (disp *dispatcher) dispatchPut(protoId pb.ProtoId, sn uint32, ditem *dispatchItem) {
	id := makeDispatchId(protoId, sn)

	disp.dispatchMutex.Lock()
	disp.dispatchMap[id] = ditem
	disp.dispatchMutex.Unlock()
}

func (disp *dispatcher) dispatchPop(protoId pb.ProtoId, sn uint32) *dispatchItem {
	id := makeDispatchId(protoId, sn)

	disp.dispatchMutex.Lock()
	ditem, ok := disp.dispatchMap[id]
	if ok {
		delete(disp.dispatchMap, id)
	}
	disp.dispatchMutex.Unlock()

	// handle push data.
	if ditem == nil {
		if resp := pb.GetPushResponseStruct(protoId); resp != nil {
			ditem = &dispatchItem{resp: resp}
		}
	}

	return ditem
}

func (disp *dispatcher) dispatchClose() {
	disp.dispatchMutex.Lock()
	for id, ditem := range disp.dispatchMap {
		if ditem.c != nil {  // Only close channels that are not nil
			close(ditem.c)
		}
		delete(disp.dispatchMap, id)
	}
	disp.dispatchMutex.Unlock()
}
