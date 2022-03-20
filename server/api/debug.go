package api

import (
	"encoding/json"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/utils"
)

// Runs a healthcheck of most systems on which the appplication relies on, will be used by e.g `Uptime Kuma`, no authentication required
// TODO: also check the hardware nodes
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.CheckDatabase(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Error("Healthcheck failed: ", err.Error())
		json.NewEncoder(w).Encode(Response{Success: false, Message: "healthcheck failed: database downtime", Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Reading system debug information, admin authentication required
// Todo: read raspberry pi information
func DebugInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utils.SysInfo())
}
