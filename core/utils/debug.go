package utils

import (
	"runtime"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/hardware"
)

type DebugInfo struct {
	ServerVersion          string                  `json:"version"`
	DatabaseOnline         bool                    `json:"databaseOnline"`
	DatabaseStats          database.DBStatus       `json:"databaseStats"`
	CpuCores               uint8                   `json:"cpuCores"`
	Goroutines             uint16                  `json:"goroutines"`
	GoVersion              string                  `json:"goVersion"`
	MemoryUsage            uint16                  `json:"memoryUsage"`
	PowerJobCount          uint16                  `json:"powerJobCount"`
	PowerJobWithErrorCount uint16                  `json:"lastPowerJobErrorCount"`
	PowerJobs              []hardware.PowerJob     `json:"powerJobs"`
	PowerJobResults        []hardware.JobResult    `json:"powerJobResults"`
	HardwareNodesCount     uint8                   `json:"hardwareNodesCount"`
	HardwareNodesOnline    uint8                   `json:"hardwareNodesOnline"`
	HardwareNodesEnabled   uint8                   `json:"hardwareNodesEnabled"`
	Nodes                  []database.HardwareNode `json:"hardwareNodes"`
}

func SysInfo() DebugInfo {
	var memoryStats runtime.MemStats
	runtime.ReadMemStats(&memoryStats)

	if err := hardware.RunNodeCheck(); err != nil {
		log.Error("Failed to run node check: ", err.Error())
	}

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
		nodes[index].Token = ""
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
	}
}
