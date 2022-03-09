package utils

import (
	"runtime"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/hardware"
)

type DebugInfo struct {
	DatabaseOnline         bool                 `json:"databaseOnline"`
	DatabaseStats          database.DBStatus    `json:"databaseStats"`
	CpuCores               uint8                `json:"cpuCores"`
	Goroutines             uint16               `json:"goroutines"`
	MemoryUsage            uint16               `json:"memoryUsage"`
	PowerJobCount          uint16               `json:"powerJobCount"`
	PowerJobWithErrorCount uint16               `json:"lastPowerJobErrorCount"`
	PowerJobs              []hardware.PowerJob  `json:"powerJobs"`
	PowerJobResults        []hardware.JobResult `json:"powerJobResults"`
}

func SysInfo() DebugInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	err := database.CheckDatabase()
	return DebugInfo{
		DatabaseOnline:         err == nil,
		DatabaseStats:          database.GetDatabaseStats(),
		CpuCores:               uint8(runtime.NumCPU()),
		Goroutines:             uint16(runtime.NumGoroutine()),
		MemoryUsage:            uint16(m.Alloc / 1024 / 1024),
		PowerJobCount:          uint16(hardware.GetPendingJobCount()),
		PowerJobs:              hardware.GetPendingJobs(),
		PowerJobResults:        hardware.GetResults(),
		PowerJobWithErrorCount: hardware.GetJobsWithErrorInHandler(),
	}
}
