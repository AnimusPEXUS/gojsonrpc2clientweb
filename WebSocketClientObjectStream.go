package gojsonrpc2clientweb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"syscall/js"

	"github.com/sourcegraph/jsonrpc2"

	gojstoolsutils "github.com/AnimusPEXUS/gojstools/utils"
	"github.com/AnimusPEXUS/gojswebapi/array"
	"github.com/AnimusPEXUS/gojswebapi/events"
	"github.com/AnimusPEXUS/gojswebapi/ws"
)

/*
type ReadObjectType uint

const (
	ReadObjectTypeNormal ReadObjectType = iota
	ReadObjectTypeError
)
*/

type ReadObjectS struct {
	MessageEvent *events.MessageEvent
	ErrorEvent   *events.ErrorEvent
	CloseEvent   *events.CloseEvent
}

type WebSocketClientObjectStreamOptions struct {
	WebSocket                *ws.WS
	GetCloseCodeAndMessageCB func() (*int, *string)
}

var _ jsonrpc2.ObjectStream = &WebSocketClientObjectStream{}

type WebSocketClientObjectStream struct {
	options           *WebSocketClientObjectStreamOptions
	read_object_queue chan interface{}
	closed            bool
	err               error
}

func NewWebSocketClientObjectStream(
	options *WebSocketClientObjectStreamOptions,
) (
	*WebSocketClientObjectStream,
	error,
) {

	self := &WebSocketClientObjectStream{}
	self.options = options
	self.read_object_queue = make(chan interface{})

	err := self.options.WebSocket.SetOnClose(self.onclose)
	if err != nil {
		return nil, err
	}
	err = self.options.WebSocket.SetOnError(self.onerror)
	if err != nil {
		return nil, err
	}
	err = self.options.WebSocket.SetOnMessage(self.onmessage)
	if err != nil {
		return nil, err
	}
	return self, nil
}

func (self *WebSocketClientObjectStream) WriteObject(obj interface{}) error {

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	// log.Println("WebSocketClientObjectStream.WriteObject marshal result:", string(data))

	len_data := len(data)

	ar, err := array.NewArray(array.ArrayTypeUint8, gojstoolsutils.JSValueLiteralToPointer(js.ValueOf(len_data)), nil, nil)
	if err != nil {
		return err
	}

	ar_js := *ar.JSValue

	// arnp := *ar.JSValue
	// log.Println("arnp:", arnp.Type().String())
	// log.Println("arnp 2:", ar.JSValue.String())

	res := js.CopyBytesToJS(ar_js, data)
	if res != len_data {
		log.Println("res != len_data")
		return errors.New("WebSocketClientObjectStream:WriteObject:res != len_data")
	}

	// log.Println("ar_js", ar_js.Call("toString").String())
	// log.Println("ar_js2", ar_js.Call("toString").String())
	// state, _ := self.options.WebSocket.ReadyStateGet()
	// log.Println("ws state", state)
	// url, _ := self.options.WebSocket.URLGet()
	// log.Println("ws url", url)

	log.Println("WebSocketClientObjectStream : sending ", data)

	err = self.options.WebSocket.Send(&ar_js)
	if err != nil {
		return err
	}
	log.Println("WebSocketClientObjectStream : sent")
	return nil
}

func (self *WebSocketClientObjectStream) Close() error {
	var code *int
	var reason *string
	if self.options.GetCloseCodeAndMessageCB != nil {
		code, reason = self.options.GetCloseCodeAndMessageCB()
	}
	return self.options.WebSocket.Close(code, reason)
}

func (self *WebSocketClientObjectStream) ReadObject(v interface{}) error {
	log.Println("waiting for message from chan")
	e := <-self.read_object_queue
	log.Println("got message from chan")
	var val *ReadObjectS
	val = e.(*ReadObjectS)
	if val.CloseEvent != nil || val.ErrorEvent != nil {
		if val.ErrorEvent != nil {
			e_msg, err := val.ErrorEvent.GetMessage()
			if err != nil {
				return err
			}
			e_file, err := val.ErrorEvent.GetFilename()
			if err != nil {
				return err
			}
			e_line, err := val.ErrorEvent.GetLineno()
			if err != nil {
				return err
			}
			return errors.New(fmt.Sprintf("error: %s ; file: %s ; line: %d", e_msg, e_file, e_line))
		}
		if val.CloseEvent != nil {
			return errors.New("stream closed")
		}
	} else {
		if val.MessageEvent == nil {
			return errors.New("WebSocketClientObjectStream invalid ReadObjectS structure")
		}
		data, err := val.MessageEvent.GetData()
		if err != nil {
			return err
		}

		if data.Type() == js.TypeString {
			err = json.Unmarshal([]byte(data.String()), v)
			if err != nil {
				return err
			}
			return nil
		}

		array_type := array.DetermineArrayType(data)
		if array_type != nil {
			if *array_type != array.ArrayTypeUint8 {
				return errors.New("websocket message.data have not ArrayTypeUint8 array type")
			}
			godata := make([]byte, data.Get("length").Int())
			js.CopyBytesToGo(godata, *data)
			err = json.Unmarshal(godata, v)
			if err != nil {
				return err
			}
			return nil
		}

		return errors.New("unsupported type of message data:" + data.Type().String())

	}
	return nil
}

func (self *WebSocketClientObjectStream) onerror(e *events.ErrorEvent) {
	log.Println("got error:", e)
	self.read_object_queue <- (&ReadObjectS{ErrorEvent: e})
}

func (self *WebSocketClientObjectStream) onclose(e *events.CloseEvent) {
	log.Println("got close:", e)
	self.read_object_queue <- (&ReadObjectS{CloseEvent: e})
}

func (self *WebSocketClientObjectStream) onmessage(e *events.MessageEvent) {
	log.Println("got message:", e)
	self.read_object_queue <- (&ReadObjectS{MessageEvent: e})
}
