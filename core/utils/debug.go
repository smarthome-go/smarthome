package utils

import (
	"runtime"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type DebugInfo struct {
	ServerVersion      string            `json:"version"`
	DatabaseOnline     bool              `json:"databaseOnline"`
	DatabaseStats      database.DBStatus `json:"databaseStats"`
	CpuCores           uint8             `json:"cpuCores"`
	Goroutines         uint16            `json:"goroutines"`
	GoVersion          string            `json:"goVersion"`
	MemoryUsage        uint16            `json:"memoryUsage"`
	HomescriptJobCount uint              `json:"homescriptJobCount"`
	Time               serverTime        `json:"time"`
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

	err := database.CheckDatabase()
	return DebugInfo{
		ServerVersion:      Version,
		DatabaseOnline:     err == nil,
		DatabaseStats:      database.GetDatabaseStats(),
		CpuCores:           uint8(runtime.NumCPU()),
		Goroutines:         uint16(runtime.NumGoroutine()),
		GoVersion:          runtime.Version(),
		MemoryUsage:        uint16(memoryStats.Alloc / 1024 / 1024),
		HomescriptJobCount: uint(len(homescript.HmsManager.GetJobList())),
		Time: serverTime{
			Hours:   uint(time.Now().Hour()),
			Minutes: uint(time.Now().Minute()),
			Seconds: uint(time.Now().Second()),
			Unix:    uint(time.Now().UnixMilli()),
		},
	}
}
