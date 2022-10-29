package api

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/server/middleware"
)

const megabyte = 1000000

type HMSMessageOut struct {
	Kind    HMSMessageKind `json:"kind"`
	Payload string         `json:"payload"`
}
type HMSResWs struct {
	Kind     HMSMessageKind        `json:"kind"`
	Exitcode int                   `json:"exitCode"`
	Errors   []homescript.HmsError `json:"errors"`
}

type HMSMessageKind string

const (
	MessageKindStdOut  HMSMessageKind = "out"
	MessageKindResults HMSMessageKind = "res"
)

// Runs any given Homescript as a string
// The output is in realtime
func RunHomescriptStringAsync(w http.ResponseWriter, r *http.Request) {
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}

	// Upgrade the connection
	upgrader := websocket.Upgrader{}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Could not upgrade connection to WS: ", err.Error())
		return
	}
	defer ws.Close()

	outReader, outWriter, err := os.Pipe()
	if err != nil {
		log.Error("Could not open OS pipe: ", err.Error())
		if err := ws.WriteJSON(Response{Success: false, Message: "could not run Homescript", Error: "could not open os pipe"}); err != nil {
			log.Error("Cannot write to websocket: ", err.Error())
			return
		}
		return
	}
	defer outReader.Close()
	defer outWriter.Close()

	// Receive the code and args to run
	ws.SetReadLimit(100 * megabyte)
	if err := ws.SetWriteDeadline(time.Now().Add(time.Minute)); err != nil {
		return
	}
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(time.Minute)); return nil })
	var request HomescriptLiveRunRequest
	if err := ws.ReadJSON(&request); err != nil {
		log.Error("Cannot receive JSON: ", err.Error())
		return
	}

	// Start running the code
	res := make(chan homescript.HmsExecRes)
	go func(writer io.Writer, results *chan homescript.HmsExecRes) {
		res := homescript.HmsManager.Run(
			username,
			"live",
			request.Code,
			make(map[string]string),
			make([]string, 0),
			homescript.InitiatorAPI,
			make(chan int),
			writer,
		)
		*results <- res
	}(outWriter, &res)

	// Stream the stdout
	scanner := bufio.NewScanner(outReader)

	killPipe := make(chan bool)
	go func(kill chan bool) {
		for scanner.Scan() {
			out := scanner.Bytes()
			fmt.Println(string(out))

			select {
			case <-kill:
				return
			default:
				time.Sleep(time.Millisecond)
			}

			ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			ws.WriteJSON(HMSMessageOut{
				Kind:    MessageKindStdOut,
				Payload: string(out),
			})
			if scanner.Err() != nil {
				log.Error("Scanner failed: ", err.Error())
			}
		}
	}(killPipe)

outer:
	for {
		select {
		case res := <-res:
			killPipe <- true
			ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			ws.WriteJSON(HMSResWs{
				Kind:     MessageKindResults,
				Exitcode: res.ExitCode,
				Errors:   res.Errors,
			})
			break outer
		default:
			time.Sleep(time.Millisecond)
		}
	}
	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(10 * time.Second)
	ws.Close()
}
