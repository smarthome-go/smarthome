package api

import (
	"encoding/json"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/server/middleware"
)

type NotificationCountResponse struct {
	NotificationCount uint16 `json:"count"`
}

type DeleteNotificationByIdRequest struct {
	Id uint `json:"id"`
}

// Returns a uin16 that indicates the number of notifications the current user has, no authentication required
func GetNotificationCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	notificationCount, err := database.GetUserNotificationCount(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed get notification count", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(NotificationCountResponse{NotificationCount: notificationCount})
}

// Returns a list containing notifications of the current user
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	notifications, err := database.GetUserNotifications(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed get notification count", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(notifications)
}

// Delete a given notification from the current user
func DeleteUserNotificationById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteNotificationByIdRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	if err := database.DeleteNotificationFromUserById(request.Id, username); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to delete notification", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "Successfully sent deletion request"})
}

func DeleteAllUserNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "could not get username from session", Error: "malformed user session"})
		return
	}
	if err := database.DeleteAllNotificationsFromUser(username); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Response{Success: false, Message: "failed to delete all notifications", Error: "database failure"})
		return
	}
	json.NewEncoder(w).Encode(Response{Success: true, Message: "successfully deleted all notifications"})
}
