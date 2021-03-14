package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/AnimusPEXUS/gojsonrpc2clientweb"
	"github.com/AnimusPEXUS/gojswebapi/events"
	"github.com/AnimusPEXUS/gojswebapi/ws"
	"github.com/AnimusPEXUS/gojsonrpc2clientweb/examples/serverandclient/types"
	mysync "github.com/AnimusPEXUS/utils/sync"
)

func main() {

	l := mysync.NewMutexCheckable(false)

	op_sig := sync.NewCond(l)

	opts := &ws.WSOptions{
		URL: &([]string{"wss://localhost:9999/ws"}[0]),
		// URL:    "wss://localhost:9999/ws",
		OnOpen: func(*events.Event) {
			log.Println("OnOpen")
			op_sig.Signal()
		},

		OnClose: func(*events.CloseEvent) {
			log.Println("OnClose")
			op_sig.Signal()
		},

		OnError: func(*events.ErrorEvent) {
			log.Println("OnError")
			op_sig.Signal()
		},

		OnMessage: func(*events.MessageEvent) {
			log.Println("OnMessage")
			op_sig.Signal()
		},
	}

	wsoc, err := ws.NewWS(opts)
	if err != nil {
		log.Fatalln("err:", err)
	}

	object_stream_options := &gojsonrpc2clientweb.WebSocketClientObjectStreamOptions{
		WebSocket: wsoc,
	}

	object_stream, err := gojsonrpc2clientweb.NewWebSocketClientObjectStream(object_stream_options)
	if err != nil {
		log.Fatalln("err:", err)
	}

	log.Println("waiting socket to open")
	op_sig.Wait()
	log.Println("  open wait complete")

	jrpc_conn := jsonrpc2.NewConn(
		context.Background(),
		object_stream,
		nil,
	)
	defer jrpc_conn.Close()

	var tp_res *types.TestData

	tp := &types.TestData{
		Command:       "cmd",
		DefaultResult: 100,
		Parameters:    []string{"a", "c", "z"},
	}

	log.Println("calling func")
	err = jrpc_conn.Call(context.Background(), "Test1", tp, &tp_res)
	if err != nil {
		log.Fatalln("err:", err)
	}

	log.Println("tp_res:", tp_res)

	for {
		time.Sleep(time.Second)
		log.Println("sleep iteration")
	}

}
