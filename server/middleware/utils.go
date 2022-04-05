package middleware

import (
	"encoding/json"
	"net/http"
	"time"
)

func Res(w http.ResponseWriter, res Response) {
	now := time.Now().Local()
	response := res
	response.Time = now.Format(time.UnixDate)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("internal server error")); err != nil {
			log.Error("Could not send response to client")
		}
	}
}
