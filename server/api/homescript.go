package api

import (
	"encoding/json"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/homescript"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type HomescriptResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
	Error   string `json:"error"`
}

type HomescriptRequest struct {
	Code string `json:"code"`
}

// Runs any given Homescript as a string
func RunHomescriptString(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request HomescriptRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	output, hmsErrors := homescript.Run(username, "live", request.Code)
	if len(hmsErrors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(HomescriptResponse{
			Success: false,
			Message: "Homescript terminated abnormally",
			Output:  output,
			// TODO: handle multiple errors
			Error: hmsErrors[0].Error(),
		})
		return
	}
	json.NewEncoder(w).Encode(HomescriptResponse{
		Success: true,
		Message: "Homescript ran successfully",
		Output:  output,
	})
}

// Returns a list of homescripts which are owned by the current user
func ListPersonalHomescripts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	homescriptList, err := database.ListHomescriptOfUser(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to list personal homescript", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(homescriptList)
}
