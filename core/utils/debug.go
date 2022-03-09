package utils

import (
	"runtime"
)

type DebugInfo struct {
	CpuCores    uint8
	Goroutines  uint16
	MemoryUsage uint16
}

func SysInfo() DebugInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return DebugInfo{CpuCores: uint8(runtime.NumCPU()), Goroutines: uint16(runtime.NumGoroutine()), MemoryUsage: uint16(m.Alloc / 1024 / 1024)}
}
