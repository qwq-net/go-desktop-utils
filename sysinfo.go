//go:build windows

package main

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func sysInfoLoop(hwnd uintptr) {
	// First call with 1s interval to get a valid baseline
	collectSysInfo(time.Second)
	postRefresh(hwnd)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		collectSysInfo(0)
		postRefresh(hwnd)
	}
}

func collectSysInfo(cpuInterval time.Duration) {
	cpuVal := 0.0
	if percents, err := cpu.Percent(cpuInterval, false); err == nil && len(percents) > 0 {
		cpuVal = percents[0]
	}

	var memPercent, memUsedGB, memTotalGB float64
	if v, err := mem.VirtualMemory(); err == nil {
		memPercent = v.UsedPercent
		memUsedGB = float64(v.Used) / (1024 * 1024 * 1024)
		memTotalGB = float64(v.Total) / (1024 * 1024 * 1024)
	}

	appState.mu.Lock()
	appState.cpuPercent = cpuVal
	appState.memPercent = memPercent
	appState.memUsedGB = memUsedGB
	appState.memTotalGB = memTotalGB
	appState.mu.Unlock()
}
