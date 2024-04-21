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
type HmsMessageRXInit struct {
	Kind     HMSMessageKindRX `json:"kind"`
	HMSID    string           `json:"hmsID"`
	IsDriver bool             `json:"isDriver"`
	Args     []HomescriptArg  `json:"args"`
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

	// Start running the code.
	res := make(chan types.HmsRes)
	idChan := make(chan uint64)

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
		// Check if the script should be killed.
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
				Message: fmt.Sprintf("invalid kill message: `%s`\n", err.Error()),
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
				Message: fmt.Sprintf("invalid kill request kind: `%s`\n", request.Kind),
			}); err != nil {
				return
			}
			wsMutex.Unlock()
		}

		log.Trace("Killing script from Websocket...")
		homescript.HmsManager.Kill(jobId)
	}()

	scanner := bufio.NewScanner(outReader)
	killPipe := make(chan bool)

	go func(kill chan bool) {
		scanner.Split(bufio.ScanRunes)

		// lastBufioInput := time.Now()

		// scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// 	timeSinceLast := time.Since(lastBufioInput)
		// 	lastBufioInput = time.Now()
		//
		// 	throughput := float64(len(data)) / timeSinceLast.Seconds()
		// 	fmt.Printf("=====  len: %d | time since: %f | throughput: %f ====\n", len(data), timeSinceLast.Seconds(), throughput)
		//
		// 	if timeSinceLast.Milliseconds() < 1 {
		// 		fmt.Println("time since last is sub milli")
		// 		return bufio.ScanRunes(data, atEOF)
		// 	}
		//
		// 	// TODO: use another split function dynamically based on the "throughput"
		// 	// if throughput < 10 {
		// 	// 	return bufio.ScanRunes(data, atEOF)
		// 	// }
		//
		// 	// if len(data) > 10000 {
		// 	// 	fmt.Println("=== emergency dump ===")
		// 	// 	// return BufioScanAll(data, atEOF)
		// 	// }
		//
		// 	if len(data) < 100 {
		// 		return 0, nil, nil
		// 	}
		//
		// 	advance = 0
		// 	token = make([]byte, 0)
		// 	for i := 0; i < 100; i++ {
		// 		advanceTemp, tokenTemp, err := bufio.ScanRunes(data[i:], atEOF)
		// 		if err != nil {
		// 			return advanceTemp, tokenTemp, err
		// 		}
		//
		// 		advance += advanceTemp
		// 		token = append(token, tokenTemp...)
		// 	}
		//
		// 	return advance, token, nil
		// })

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
				return
			}

			if err := ws.WriteJSON(HMSMessageTXRes{
				Kind:         MessageKindResults,
				Errors:       res.Errors.Diagnostics,
				FileContents: res.Errors.FileContents,
				Success:      !res.Errors.ContainsError,
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
	if err := ws.SetWriteDeadline(time.Now().Add(wsTimeout)); err != nil {
		return
	}
	if err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		return
	}
	wsMutex.Unlock()
	// Give the client a grace period to close the connection.
	time.Sleep(time.Second)
	ws.Close()
}
