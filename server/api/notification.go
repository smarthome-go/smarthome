package api

import (
	"encoding/json"
	"net/http"
	"unicode/utf8"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/user/notify"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type NotifyRequest struct {
	Priority    uint8  `json:"priority"` // Includes 1: info, 2: warning, 3: alert
	Name        string `json:"name"`
	Description string `json:"description"`
}

type NotificationCountResponse struct {
	NotificationCount uint16 `json:"count"`
}

type NotificationIdBody struct {
	Id uint `json:"id"`
}

func NotifyUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}

	var request NotifyRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	if request.Priority < 1 || request.Priority > 3 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "illegal request body", Error: "priority must be 1 <= p <= 3"})
		return
	}

	if utf8.RuneCountInString(request.Name) > 100 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "illegal request body", Error: "name must no be longer than 100 characters"})
		return
	}

	newId, err := notify.Manager.Notify(
		username,
		request.Name,
		request.Description,
		notify.NotificationLevel(request.Priority),
		true,
	)

	if err := json.NewEncoder(w).Encode(NotificationIdBody{Id: newId}); err != nil {
		Res(w, Response{Success: false, Message: "failed to add notification", Error: "could not encode response"})
	}
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
		Res(w, Response{Success: false, Message: "failed to get notification count", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(NotificationCountResponse{NotificationCount: notificationCount}); err != nil {
		Res(w, Response{Success: false, Message: "failed to get notification count", Error: "could not encode response"})
	}
}

// Returns a list containing notifications of the current user
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	notifications, err := notify.GetNotifications(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to get notifications", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get notifications", Error: "could not encode response"})
	}
}

// Delete a given notification from the current user
func DeleteUserNotificationById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request NotificationIdBody
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
