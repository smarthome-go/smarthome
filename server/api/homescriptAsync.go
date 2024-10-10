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
	"github.com/smarthome-go/smarthome/core/homescript/types"
	"github.com/smarthome-go/smarthome/server/middleware"
)

// BUG: is this even correct?
const megabyte = 1000000
const wsTimeout = 10 * time.Second

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
	Kind         HMSMessageKindTX  `json:"kind"`
	Errors       []types.HmsError  `json:"errors"`
	FileContents map[string]string `json:"fileContents"`
	Success      bool              `json:"success"`
}

type HMSMessageKindTX string

const (
	// Sent by the server
	MessageKindErr     HMSMessageKindTX = "err"
	MessageKindStdOut  HMSMessageKindTX = "out"
	MessageKindResults HMSMessageKindTX = "res"
)

// Messages sent by the client
type HmsMessageRX struct {
	Kind HMSMessageKindRX `json:"kind"`

	// For init.
	HMSID    string          `json:"hmsID"`
	IsDriver bool            `json:"isDriver"`
	Args     []HomescriptArg `json:"args"`

	// Relatively generic, mosyly used for STDIN messages.
	Payload string `json:"payload"`
}

type HMSMessageKindRX string

const (
	// Sent by the client
	MessageKindInit  HMSMessageKindRX = "init"
	MessageKindKill  HMSMessageKindRX = "kill"
	MessageKindStdin HMSMessageKindRX = "stdin"
)

func BufioScanAll(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	return len(data), data, nil
}

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
	if err := ws.SetReadDeadline(time.Time{}); err != nil {
		return
	}
	ws.SetPongHandler(func(string) error {
		return ws.SetReadDeadline(time.Time{})
	})
	var request HmsMessageRX
	if err := ws.ReadJSON(&request); err != nil {
		wsMutex.Lock()
		if err := ws.WriteJSON(HMSMessageTXErr{
			Kind:    MessageKindErr,
			Message: fmt.Sprintf("invalid init request: %s", err.Error()),
		}); err != nil {
			wsMutex.Unlock()
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
			wsMutex.Unlock()
			return
		}
		wsMutex.Unlock()
		return
	}
	args := make(map[string]string)
	for _, arg := range request.Args {
		args[arg.Key] = arg.Value
	}

	// Start running the code.
	res := make(chan types.HmsRes)
	idChan := make(chan uint64)
	stdin := types.NewStdinBuffer()

	go func(writer io.Writer, results *chan types.HmsRes, idChan *chan uint64) {
		ctx, cancel := context.WithCancel(context.Background())

		res, err := homescript.HmsManager.RunUserScript(
			request.HMSID,
			username,
			nil,
			types.Cancelation{
				Context:    ctx,
				CancelFunc: cancel,
			},
			outWriter,
			idChan,
			stdin,
		)

		log.Tracef("WS homescript (%s) finished.", request.HMSID)

		if err != nil {
			wsMutex.Lock()
			if err := ws.WriteJSON(HMSMessageTXErr{
				Kind:    MessageKindErr,
				Message: fmt.Sprintf("Could not run Homescript: internal error: %s", err.Error()),
			}); err != nil {
				wsMutex.Unlock()
				return
			}
			ws.Close()
			wsMutex.Unlock()
			return
		}
		outWriter.Close()

		fmt.Println("block before res")
		*results <- res
		fmt.Println("block finished res")
	}(outWriter, &res, &idChan)

	jobId := <-idChan

	go func() {
		// Check if the script should be killed.
		ws.SetReadLimit(100 * megabyte)
		if err := ws.SetReadDeadline(time.Time{}); err != nil {
			return
		}
		ws.SetPongHandler(func(string) error { return ws.SetReadDeadline(time.Time{}) })

		for {
			var request HmsMessageRX
			if err := ws.ReadJSON(&request); err != nil {
				wsMutex.Lock()
				if err := ws.WriteJSON(HMSMessageTXErr{
					Kind:    MessageKindErr,
					Message: fmt.Sprintf("invalid kill message: `%s`\n", err.Error()),
				}); err != nil {
					wsMutex.Unlock()
					return
				}
				wsMutex.Unlock()
				return
			}

			switch request.Kind {
			case MessageKindKill:
				log.Trace("Killing script from Websocket...")
				homescript.HmsManager.Kill(jobId)
				log.Trace("Killed script from Websocket...")
				return
			case MessageKindStdin:
				log.Tracef("Got HMS stdin: `%s`", request.Payload)
				stdin.Send(request.Payload)
				continue
			default:
				wsMutex.Lock()
				if err := ws.WriteJSON(HMSMessageTXErr{
					Kind:    MessageKindErr,
					Message: fmt.Sprintf("invalid WS HMS request kind: `%s`\n", request.Kind),
				}); err != nil {
					wsMutex.Unlock()
					return
				}
				wsMutex.Unlock()
				return
			}
		}
	}()

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
				wsMutex.Unlock()
				return
			}

			wsMutex.Unlock()
		}
		if scanner.Err() != nil {
			log.Error("Scanner failed: ", scanner.Err())
		}
		<-kill
	}(killPipe)

outer:
	for {
		select {
		case res := <-res:
			killPipe <- true
			wsMutex.Lock()
			if err := ws.SetWriteDeadline(time.Now().Add(wsTimeout)); err != nil {
				wsMutex.Unlock()
				return
			}

			if err := ws.WriteJSON(HMSMessageTXRes{
				Kind:         MessageKindResults,
				Errors:       res.Errors.Diagnostics,
				FileContents: res.Errors.FileContents,
				Success:      !res.Errors.ContainsError,
			}); err != nil {
				wsMutex.Unlock()
				return
			}
			wsMutex.Unlock()
			break outer
		default:
			time.Sleep(time.Millisecond)
		}
	}
	wsMutex.Lock()
	if err := ws.SetWriteDeadline(time.Now().Add(wsTimeout)); err != nil {
		wsMutex.Unlock()
		return
	}
	if err := ws.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	); err != nil {
		wsMutex.Unlock()
		return
	}

	// Give the client a grace period to close the connection.
	time.Sleep(time.Second)
	ws.Close()

	wsMutex.Unlock()
}
