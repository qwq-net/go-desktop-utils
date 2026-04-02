//go:build windows

package widget

import (
	"time"

	"go-desktop-utils/internal/w32"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func (a *App) SysInfoLoop() {
	collectSysInfo(a, time.Second)
	w32.PostRefresh(a.Hwnd)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		collectSysInfo(a, 0)
		w32.PostRefresh(a.Hwnd)
	}
}

func collectSysInfo(a *App, cpuInterval time.Duration) {
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

	a.State.Mu.Lock()
	a.State.CpuPercent = cpuVal
	a.State.MemPercent = memPercent
	a.State.MemUsedGB = memUsedGB
	a.State.MemTotalGB = memTotalGB
	a.State.Mu.Unlock()
}
