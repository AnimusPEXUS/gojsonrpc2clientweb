package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/AnimusPEXUS/gojsonrpc2clientweb/examples/serverandclient/types"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	jsonrpc2websocket "github.com/sourcegraph/jsonrpc2/websocket"
)

type H struct {
}

func (self *H) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	log.Println("handling")
	switch req.Method {
	case "Test1":
		log.Println("server Test1")
		d, err := req.Params.MarshalJSON()
		if err != nil {
			log.Fatalln("err", err)
			return
		}

		log.Println("parameters json:", string(d))

		var td *types.TestData

		err = json.Unmarshal(d, &td)
		if err != nil {
			log.Fatalln("err", err)
			return
		}

		td.Result = "result 123"

		log.Println("test data", td)

		err = conn.Reply(ctx, req.ID, td)
		if err != nil {
			log.Fatalln("err", err)
			return
		}
	}
}

func main() {

	log.SetFlags(log.Lshortfile)

	h := &H{}

	mux_router := mux.NewRouter()

	mux_router.PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	mux_router.Path("/ws").Methods("GET").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			log.Println("NewFastHTTPHandlerFunc")
			defer log.Println("NewFastHTTPHandlerFunc exited")

			up := &websocket.Upgrader{}
			ws, err := up.Upgrade(w, r, nil)
			if err != nil {
				log.Println("err:", err)
				return
			}

			// ws_net_conn := gorilla.NewGorillaWSConn(ws)

			object_stream := jsonrpc2websocket.NewObjectStream(ws)

			// bs := jsonrpc2.NewBufferedStream(ws_net_conn, &jsonrpc2.VarintObjectCodec{})

			ctx := context.Background()

			jr2_conn := jsonrpc2.NewConn(ctx, object_stream, h)

			<-jr2_conn.DisconnectNotify()
		},
	)

	certificate, err := tls.LoadX509KeyPair("./tls/cert.pem", "./tls/key.pem")
	if err != nil {
		log.Fatal("error loading keyfile or certificate: ", err)
	}

	tls_cfg := &tls.Config{
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: true,
	}

	s := &http.Server{
		Addr:           "0.0.0.0:9999",
		Handler:        mux_router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      tls_cfg,
	}

	for {
		log.Println("Server settings:", s)
		log.Println("ws: listening..")
		err = s.ListenAndServeTLS("", "")
		if err != nil {
			log.Println("ws: ListenAndServeTLS: ", err)
			time.Sleep(time.Second)
			continue
		}
	}

}
