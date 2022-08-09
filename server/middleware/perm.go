package middleware

import (
	"fmt"
	"net/http"

	"github.com/smarthome-go/smarthome/core/database"
)

// Middleware for checking if a user has permission to access given resources
// The permissions to check is given as second or more arguments
func Perm(handler http.HandlerFunc, permissionsToCheck ...database.PermissionType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := GetUserFromCurrentSession(w, r)
		for _, permission := range permissionsToCheck {
			log.Trace(fmt.Sprintf("Checking permission `%s` for user `%s`", permission, username))
		}
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			Res(w, Response{Success: false, Message: "access denied, invalid session", Error: "clear your browser's cookies"})
			return
		}
		for _, permission := range permissionsToCheck {
			hasPermission, err := database.UserHasPermission(username, permission)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				Res(w, Response{Success: false, Message: "database error", Error: "failed to check permission to access this resource"})
				return
			}
			if !hasPermission {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				Res(w, Response{Success: false, Message: "permission denied", Error: "missing permission to access this resource, contact your administrator"})
				return
			}
		}
		handler.ServeHTTP(w, r)
	}
}
