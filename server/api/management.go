package api

import (
	"net/http"

	"github.com/smarthome-go/smarthome/core"
	"github.com/smarthome-go/smarthome/server/middleware"
)

func ReloadServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := core.Reload(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to reload server", Error: err.Error()})
		return
	}
	Res(w, Response{Success: true, Message: "server was reloaded successfully"})
	middleware.InitWithRandomKey()
}

func ShutdownServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := core.Shutdown(true); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to shutdown server", Error: err.Error()})
		return
	}

	panic("Unreachable")
}
