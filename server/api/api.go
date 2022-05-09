// Package classification Smarthome API.
//
// Documentation of the Smarthome server API.
//
//     Schemes: http
//     BasePath: /
//     Version: 1.0.0
//     Host: localhost
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - basic
//
//    SecurityDefinitions:
//    basic:
//      type: basic
//
// swagger:meta
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func Res(w http.ResponseWriter, res Response) {
	now := time.Now().Local()
	response := res
	response.Time = fmt.Sprint(now.UnixMilli())
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Could not send response to client: ", err.Error())
		return
	}
}
