package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	coap "github.com/plgd-dev/go-coap/v3"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
)

type ResponseInfo struct {
	Result string `json:"result"`
}

func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		log.Printf("ClientAddress %v, %v\n", w.Conn().RemoteAddr(), r.String())

		next.ServeCOAP(w, r)
	})
}

type TemperatureResponse struct {
	Temperature float64 `json:"temperature"`
}

func handleTemperature(w mux.ResponseWriter, r *mux.Message) {
	responseInfo := TemperatureResponse{
		Temperature: 10 + rand.Float64()*(20-10),
	}
	response, err := json.Marshal(responseInfo)
	if err != nil {
		log.Printf("cannot parse response: %v", err)
		return
	}
	err = w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader(response))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

type LampCommand struct {
	LampOn bool `json:"lampOn"`
}

func handleLamp(w mux.ResponseWriter, r *mux.Message) {
	payload, err := r.ReadBody()
	if err != nil {
		log.Fatalf("Error reading payload: %v", err)
	}
	log.Printf("Request payload: %v", string(payload))

	var command LampCommand

	var response ResponseInfo
	err = json.Unmarshal(payload, &command)
	if err != nil {
		response = ResponseInfo{Result: "unrecognized command"}
	} else {
		if command.LampOn {
			response = ResponseInfo{Result: "the lamp turend on"}
		} else if !command.LampOn {
			response = ResponseInfo{Result: "the lamp turend off"}
		} else {
			response = ResponseInfo{Result: "unrecognized command"}
		}

	}
	responseByte, err := json.Marshal(response)
	if err != nil {
		log.Printf("cannot parse response: %v", err)
		return
	}

	customResp := w.Conn().AcquireMessage(r.Context())
	defer w.Conn().ReleaseMessage(customResp)
	customResp.SetCode(codes.Content)
	customResp.SetToken(r.Token())
	customResp.SetContentFormat(message.TextPlain)
	customResp.SetBody(bytes.NewReader(responseByte))
	err = w.Conn().WriteMessage(customResp)
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Handle("/temperature", mux.HandlerFunc(handleTemperature))
	r.Handle("/lamp", mux.HandlerFunc(handleLamp))

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Panic(err)
	}

	addr := fmt.Sprintf("%s:%d", os.Getenv("ADDRESS"), port)
	log.Printf("Starting CoAP simulator on %s", addr)
	log.Fatal(coap.ListenAndServe("udp", addr, r))
}
