package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/MikMuellerDev/smarthome/services/camera"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func Res(w http.ResponseWriter, res Response) {
	now := time.Now().Local()
	response := res
	response.Time = now.Format(time.UnixDate)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Could not send response to client: ", err.Error())
		return
	}
}

// TEST IMAGE FETCHING MODULE
func TestImageProxy(w http.ResponseWriter, r *http.Request) {
	imageData, err := camera.TestReturn()
	if err != nil {
		log.Error("Failed to test proxy: ", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", http.DetectContentType(imageData))
	if _, err := w.Write(imageData); err != nil {
		log.Error(err.Error())
	}
}
