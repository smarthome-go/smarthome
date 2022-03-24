package api

import (
	"net/http"

	"github.com/MikMuellerDev/smarthome/services/camera"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
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
