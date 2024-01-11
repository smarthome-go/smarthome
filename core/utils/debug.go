package utils

import (
	"runtime"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type DebugInfo struct {
	ServerVersion          string                     `json:"version"`
	DatabaseOnline         bool                       `json:"databaseOnline"`
	DatabaseStats          database.DBStatus          `json:"databaseStats"`
	CpuCores               uint8                      `json:"cpuCores"`
	Goroutines             uint16                     `json:"goroutines"`
	GoVersion              string                     `json:"goVersion"`
	MemoryUsage            uint16                     `json:"memoryUsage"`
	PowerJobCount          uint16                     `json:"powerJobCount"`
	PowerJobWithErrorCount uint16                     `json:"lastPowerJobErrorCount"`
	PowerJobs              []hardware.DeviceOutputJob `json:"powerJobs"`
	PowerJobResults        []hardware.JobResult       `json:"powerJobResults"`
	HardwareNodesCount     uint8                      `json:"hardwareNodesCount"`
	HardwareNodesOnline    uint8                      `json:"hardwareNodesOnline"`
	HardwareNodesEnabled   uint8                      `json:"hardwareNodesEnabled"`
	Nodes                  []database.HardwareNode    `json:"hardwareNodes"`
	HomescriptJobCount     uint                       `json:"homescriptJobCount"`
	Time                   serverTime                 `json:"time"`
}

type serverTime struct {
	Hours   uint `json:"hours"`
	Minutes uint `json:"minutes"`
	Seconds uint `json:"seconds"`
	Unix    uint `json:"unix"`
}

func SysInfo() DebugInfo {
	var memoryStats runtime.MemStats
	runtime.ReadMemStats(&memoryStats)

	// TODO: also include driver health check if supported

	// if err := hardware.RunNodeCheck(); err != nil {
	// 	log.Error("Failed to run node check: ", err.Error())
	// }

	nodes, err := database.GetHardwareNodes()
	if err != nil {
		log.Error("Failed to obtain node information while getting debug info: ", err.Error())
	}

	nodesOnline := 0
	nodesEnabled := 0
	for index, node := range nodes {
		if node.Online {
			nodesOnline += 1
		}
		if node.Enabled {
			nodesEnabled += 1
		}
		// Remove token visibility from debug info
		nodes[index].Token = "redacted"
	}

	err = database.CheckDatabase()
	return DebugInfo{
		ServerVersion:          Version,
		DatabaseOnline:         err == nil,
		DatabaseStats:          database.GetDatabaseStats(),
		CpuCores:               uint8(runtime.NumCPU()),
		Goroutines:             uint16(runtime.NumGoroutine()),
		GoVersion:              runtime.Version(),
		MemoryUsage:            uint16(memoryStats.Alloc / 1024 / 1024),
		PowerJobCount:          uint16(hardware.GetPendingJobCount()),
		PowerJobs:              hardware.GetPendingJobs(),
		PowerJobResults:        hardware.GetResults(),
		PowerJobWithErrorCount: hardware.GetJobsWithErrorInHandler(),
		HardwareNodesCount:     uint8(len(nodes)),
		HardwareNodesOnline:    uint8(nodesOnline),
		HardwareNodesEnabled:   uint8(nodesEnabled),
		Nodes:                  nodes,
		HomescriptJobCount:     uint(len(homescript.HmsManager.GetJobList())),
		Time: serverTime{
			Hours:   uint(time.Now().Hour()),
			Minutes: uint(time.Now().Minute()),
			Seconds: uint(time.Now().Second()),
			Unix:    uint(time.Now().UnixMilli()),
		},
	}
}
