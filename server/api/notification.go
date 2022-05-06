package api

import (
	"encoding/json"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/server/middleware"
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
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	notificationCount, err := database.GetUserNotificationCount(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed get notification count", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(NotificationCountResponse{NotificationCount: notificationCount}); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed get notification count", Error: "could not encode response"})
	}
}

// Returns a list containing notifications of the current user
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	notifications, err := database.GetUserNotifications(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed get notifications", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed get notifications", Error: "could not encode response"})
	}
}

// Delete a given notification from the current user
func DeleteUserNotificationById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteNotificationByIdRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	if err := database.DeleteNotificationFromUserById(request.Id, username); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete notification", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "Successfully sent deletion request"})
}

func DeleteAllUserNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	if err := database.DeleteAllNotificationsFromUser(username); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete all notifications", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted all notifications"})
}
