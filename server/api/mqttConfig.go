package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher"
)

// Can be used to update the server's MQTT settings
func UpdateMQTTConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request database.MqttConfig
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	reloadErr, dbErr := core.UpdateMqttConfig(request)
	if dbErr != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to update MQTT config", Error: "database failure"})
		return
	}

	if reloadErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to update MQTT config", Error: reloadErr.Error()})
		return
	}

	Res(w, Response{Success: true, Message: "successfully updated MQTT config"})
}

type MqttStatus struct {
	Working bool    `json:"working"`
	Error   *string `json:"error"`
}

func GetMQTTStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := dispatcher.Instance.MQTTStatus()

	var errMsg *string
	if err != nil {
		errS := err.Error()
		errMsg = &errS
	}

	if err := json.NewEncoder(w).Encode(MqttStatus{
		Working: err == nil,
		Error:   errMsg,
	}); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get MQTT status", Error: "could not encode response"})
	}
}
