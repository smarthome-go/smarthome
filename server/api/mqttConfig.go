package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core"
	"github.com/smarthome-go/smarthome/core/database"
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
		Res(w, Response{Success: false, Message: "failed to update MQTT config", Error: "backend failure"})
		return
	}

	Res(w, Response{Success: true, Message: "successfully updated MQTT config"})
}
