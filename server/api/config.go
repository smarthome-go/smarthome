package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core/config"
	"github.com/smarthome-go/smarthome/core/database"
)

type UpdateLocationRequest struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

// Can be used to update the server's latitude and longitude
func UpdateLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UpdateLocationRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if err := database.UpdateLocation(request.Latitude, request.Longitude); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to update location", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated location"})
}

// Is used to request an export of the server's configuration
func ExportConfiguration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	export, err := config.Export()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to perform configuration export", Error: "internal server error"})
		return
	}
	if err := json.NewEncoder(w).Encode(export); err != nil {
		Res(w, Response{Success: false, Message: "failed to export server configuration", Error: "could not encode content"})
	}
}
