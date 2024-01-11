package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/server/middleware"
)

const megabyte = 1000000

// Messages sent by the server
type HMSMessageTXErr struct {
	Kind    HMSMessageKindTX `json:"kind"`
	Message string           `json:"message"`
}
type HMSMessageTXOut struct {
	Kind    HMSMessageKindTX `json:"kind"`
	Payload string           `json:"payload"`
}
type HMSMessageTXRes struct {
	Kind         HMSMessageKindTX      `json:"kind"`
	Errors       []homescript.HmsError `json:"errors"`
	FileContents map[string]string     `json:"fileContents"`
	Success      bool                  `json:"success"`
}
type HMSMessageKindTX string

const (
	// Sent by the server
	MessageKindErr     HMSMessageKindTX = "err"
	MessageKindStdOut  HMSMessageKindTX = "out"
	MessageKindResults HMSMessageKindTX = "res"
)

// Messages sent by the client
type HmsMessageRXInit struct {
	Kind    HMSMessageKindRX `json:"kind"`
	Payload string           `json:"payload"`
	Args    []HomescriptArg  `json:"args"`
}

type HmsMessageRXKill struct {
	Kind HMSMessageKindRX `json:"kind"`
}

type HMSMessageKindRX string

const (
	// Sent by the client
	MessageKindInit HMSMessageKindRX = "init"
	MessageKindKill HMSMessageKindRX = "kill"
)

// Runs any given Homescript by its ID
// The output is in realtime
func RunHomescriptByIDAsync(w http.ResponseWriter, r *http.Request) {
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}

	// Upgrade the connection
	upgrader := websocket.Upgrader{}
	wsMutex := sync.Mutex{}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Could not upgrade connection to WS: ", err.Error())
		return
	}
	defer ws.Close()

	outReader, outWriter, err := os.Pipe()
	if err != nil {
		return
	}
	defer outReader.Close()
	defer outWriter.Close()

	// Receive the code and args to run
	ws.SetReadLimit(100 * megabyte)
	if err := ws.SetReadDeadline(time.Now().Add(time.Minute)); err != nil {
		return
	}
	ws.SetPongHandler(func(string) error {
		return ws.SetReadDeadline(time.Now().Add(time.Minute))
	})
	var request HmsMessageRXInit
	if err := ws.ReadJSON(&request); err != nil {
		wsMutex.Lock()
		if err := ws.WriteJSON(HMSMessageTXErr{
			Kind:    MessageKindErr,
			Message: fmt.Sprintf("invalid init request: %s", err.Error()),
		}); err != nil {
			return
		}
		wsMutex.Unlock()
		return
	}
	if request.Kind != MessageKindInit {
		wsMutex.Lock()
		if err := ws.WriteJSON(HMSMessageTXErr{
			Kind:    MessageKindErr,
			Message: fmt.Sprintf("invalid init request kind: %s", request.Kind),
		}); err != nil {
			return
		}
		wsMutex.Unlock()
		return
	}
	args := make(map[string]string)
	for _, arg := range request.Args {
		args[arg.Key] = arg.Value
	}

	// Start running the code
	res := make(chan homescript.HmsRes)

	idChan := make(chan uint64)

	go func(writer io.Writer, results *chan homescript.HmsRes, idChan *chan uint64) {
		ctx, cancel := context.WithCancel(context.Background())

		res, err := homescript.HmsManager.RunById(
			homescript.HMS_PROGRAM_KIND_NORMAL,
			request.Payload,
			username,
			homescript.InitiatorAPI,
			ctx,
			cancel,
			idChan,
			args,
			outWriter,
			nil,
		)
		if err != nil {
			wsMutex.Lock()
			if err := ws.WriteJSON(HMSMessageTXErr{
				Kind:    MessageKindErr,
				Message: fmt.Sprintf("Could not run Homescript: internal error: %s", err.Error()),
			}); err != nil {
				return
			}
			ws.Close()
			wsMutex.Unlock()
			return
		}
		outWriter.Close()

		*results <- res
	}(outWriter, &res, &idChan)

	jobId := <-idChan

	go func() {
		// Check if the script should be killed
		ws.SetReadLimit(100 * megabyte)
		if err := ws.SetReadDeadline(time.Now().Add(time.Minute)); err != nil {
			return
		}
		ws.SetPongHandler(func(string) error { return ws.SetReadDeadline(time.Now().Add(time.Minute)) })
		var request HmsMessageRXKill
		if err := ws.ReadJSON(&request); err != nil {
			wsMutex.Lock()
			if err := ws.WriteJSON(HMSMessageTXErr{
				Kind:    MessageKindErr,
				Message: fmt.Sprintf("invalid kill message: %s", err.Error()),
			}); err != nil {
				return
			}
			wsMutex.Unlock()
			return
		}
		if request.Kind != MessageKindKill {
			wsMutex.Lock()
			if err := ws.WriteJSON(HMSMessageTXErr{
				Kind:    MessageKindErr,
				Message: fmt.Sprintf("invalid kill request kind: %s", request.Kind),
			}); err != nil {
				return
			}
			wsMutex.Unlock()
		}
		// Kill the Homescript
		log.Trace("Killing script via Websocket")
		// cancel()
		homescript.HmsManager.Kill(jobId)
		log.Trace("Killed script via Websocket")
	}()

	// Stream the stdout
	scanner := bufio.NewScanner(outReader)

	killPipe := make(chan bool)
	go func(kill chan bool) {
		scanner.Split(bufio.ScanRunes)
		for scanner.Scan() {
			wsMutex.Lock()
			if err := ws.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			if err := ws.WriteJSON(HMSMessageTXOut{
				Kind:    MessageKindStdOut,
				Payload: string(scanner.Bytes()),
			}); err != nil {
				return
			}

			wsMutex.Unlock()
		}
		if scanner.Err() != nil {
			log.Error("Scanner failed: ", err.Error())
		}
		<-kill
	}(killPipe)

outer:
	for {
		select {
		case res := <-res:
			killPipe <- true
			wsMutex.Lock()
			if err := ws.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}

			if err := ws.WriteJSON(HMSMessageTXRes{
				Kind:         MessageKindResults,
				Errors:       res.Errors,
				FileContents: res.FileContents,
				Success:      res.Success,
			}); err != nil {
				return
			}
			wsMutex.Unlock()
			break outer
		default:
			time.Sleep(time.Millisecond)
		}
	}
	wsMutex.Lock()
	if err := ws.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return
	}
	if err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		return
	}
	wsMutex.Unlock()
	// Give the client a grace period to close the connection
	time.Sleep(300 * time.Millisecond)
	ws.Close()
}
