package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
)

type UpdateLocationRequest struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

// Admin endpoints for changing the servers global configuration
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
