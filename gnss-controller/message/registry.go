package message

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/daedaleanai/ublox/ubx"
)

type UBXMessageType reflect.Type

var UbxMsgNavPvt = reflect.TypeOf(&ubx.NavPvt{})
var UbxMsgNavDop = reflect.TypeOf(&ubx.NavDop{})
var UbxMsgNavSat = reflect.TypeOf(&ubx.NavSat{})
var UbxMsgNavSig = reflect.TypeOf(&ubx.NavSig{})
var UbxMsgNavCov = reflect.TypeOf(&ubx.NavCov{})
var UbxMsgNavPosecef = reflect.TypeOf(&ubx.NavPosecef{})
var UbxMsgNavTimegps = reflect.TypeOf(&ubx.NavTimegps{})
var UbxMsgNavVelecef = reflect.TypeOf(&ubx.NavVelecef{})
var UbxMsgNavStatus = reflect.TypeOf(&ubx.NavStatus{})
var UbxMsgNavEoe = reflect.TypeOf(&ubx.NavEoe{})
var UbxMsgMgaAckData = reflect.TypeOf(&ubx.MgaAckData0{})
var UbxMsgMonRf = reflect.TypeOf(&ubx.MonRf{})
var UbxRxmMeasx = reflect.TypeOf(&ubx.RxmMeasx{})
var UbxRxmRawx = reflect.TypeOf(&ubx.RxmRawx{})
var UbxRxmSfrbx = reflect.TypeOf(&ubx.RxmSfrbx{})
var UbxTimTp = reflect.TypeOf(&ubx.TimTp{})
var UbxSecEcsignWithBuffer = reflect.TypeOf(&SecEcsignWithBuffer{})

type HandlerRegistry struct {
	lock     sync.Mutex
	Handlers map[reflect.Type][]UbxMessageHandler
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		Handlers: map[reflect.Type][]UbxMessageHandler{},
	}
}

func (r *HandlerRegistry) RegisterHandler(msgType UBXMessageType, handler UbxMessageHandler) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Handlers[msgType] = append(r.Handlers[msgType], handler)
}

func (r *HandlerRegistry) UnregisterHandler(msgType reflect.Type, handler UbxMessageHandler) {
	r.lock.Lock()
	defer r.lock.Unlock()
	var newHandlers []UbxMessageHandler
	handlers := r.Handlers[msgType]
	for _, h := range handlers {
		if h == handler {
			fmt.Printf("unregistering handler %v\n", handler)
			continue
		}
		newHandlers = append(newHandlers, h)
	}
	r.Handlers[msgType] = newHandlers
}

func (r *HandlerRegistry) ForEachHandler(msgType reflect.Type, f func(handler UbxMessageHandler)) {
	r.lock.Lock()
	defer r.lock.Unlock()
	handlers := r.Handlers[msgType]
	for _, h := range handlers {
		f(h)
	}
}
