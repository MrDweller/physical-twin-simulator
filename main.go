package main

import (
	"bytes"
	"log"

	coap "github.com/plgd-dev/go-coap/v3"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
)

func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		log.Printf("ClientAddress %v, %v\n", w.Conn().RemoteAddr(), r.String())

		next.ServeCOAP(w, r)
	})
}

func handleTemperature(w mux.ResponseWriter, r *mux.Message) {
	err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("27")))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func handleLamp(w mux.ResponseWriter, r *mux.Message) {
	payload, err := r.ReadBody()
	if err != nil {
		log.Fatalf("Error reading payload: %v", err)
	}
	log.Printf("Request payload: %v", string(payload))

	var response []byte
	if string(payload) == "on" {
		response = []byte("the lamp turend on")
	} else if string(payload) == "off" {
		response = []byte("the lamp turend off")
	} else {
		response = []byte("unrecognized command")
	}

	customResp := w.Conn().AcquireMessage(r.Context())
	defer w.Conn().ReleaseMessage(customResp)
	customResp.SetCode(codes.Content)
	customResp.SetToken(r.Token())
	customResp.SetContentFormat(message.TextPlain)
	customResp.SetBody(bytes.NewReader(response))
	err = w.Conn().WriteMessage(customResp)
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func main() {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Handle("/temperature", mux.HandlerFunc(handleTemperature))
	r.Handle("/lamp", mux.HandlerFunc(handleLamp))

	log.Fatal(coap.ListenAndServe("udp", "localhost:5000", r))
}
