package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/utils"
)

// Runs a healthcheck of most systems on which the appplication relies on, will be used by e.g `Uptime Kuma`, no authentication required
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.CheckDatabase(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Error("Healthcheck failed: ", err.Error())
		Res(w, Response{Success: false, Message: "healthcheck failed: database downtime", Error: err.Error()})
		return
	}
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Error("Healthcheck failed: ", err.Error())
		Res(w, Response{Success: false, Message: "healthcheck failed: failed to get node information", Error: err.Error()})
		return
	}
	for _, node := range nodes {
		if !node.Online && node.Enabled {
			w.WriteHeader(http.StatusServiceUnavailable)
			log.Error(fmt.Sprintf("Healthcheck failed: node %s is offline", node.Url))
			Res(w, Response{Success: false, Message: "healthcheck failed: one or more nodes offline", Error: fmt.Sprintf("Node '%s' %s is offline", node.Name, node.Url)})
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// Reading system debug information, admin authentication required
// Todo: read raspberry pi information
func DebugInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(utils.SysInfo()); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get debug info", Error: "could not encode content"})
	}
}
