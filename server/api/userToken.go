package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/user"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type TokenResponse struct {
	Response Response `json:"response"`
	Token    string   `json:"token"`
}

type UserTokenDeletionRequest struct {
	Token string `json:"token"`
}

// Generates a new random token for the current user
func GenerateUserToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request database.UserTokenData
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	if utf8.RuneCountInString(request.Label) > 50 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "The `label` must not be longer than 50 characters"})
		return
	}
	token, err := user.AddToken(
		username,
		request.Label,
	)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add token", Error: "backend failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(TokenResponse{
		Response: Response{
			Success: true,
			Message: "generated new random token",
			Time:    fmt.Sprint(time.Now().UnixMilli()),
			Error:   "",
		},
		Token: token,
	}); err != nil {
		log.Error("Could not send response to client: ", err.Error())
		return
	}
}

// Generates a new random token for the current user
func DeleteUserToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request UserTokenDeletionRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	data, found, err := database.GetUserTokenByToken(request.Token)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete token", Error: "database failure"})
		return
	}
	if !found || data.User != username {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete token", Error: "invalid token provided"})
		return
	}
	if err := database.DeleteTokenByToken(request.Token); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete token", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted token"})
}

// Lists each token which belong to the current user
func ListUserTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	// List the tokens
	tokens, err := database.GetUserTokensOfUser(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to list tokens", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		log.Error("Could not send response to client: ", err.Error())
		return
	}
}
