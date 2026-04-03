//go:build windows

package widget

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go-desktop-utils/internal/w32"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// diskTargetLetters defines the drive letters to probe.
var diskTargetLetters = []string{"C", "D", "E", "F"}

// SysInfoLoop collects CPU, MEM, GPU, and NET every 5 seconds.
func (a *App) SysInfoLoop() {
	cfg := a.Config

	// NET baseline
	var prevRecv, prevSent uint64
	var prevTime time.Time
	if cfg.System.Network {
		if counters, err := net.IOCounters(false); err == nil && len(counters) > 0 {
			prevRecv = counters[0].BytesRecv
			prevSent = counters[0].BytesSent
			prevTime = time.Now()
		}
	}

	// GPU: track availability to avoid repeated exec on failure
	gpuFailed := false

	collectSysInfo(a, time.Second)
	if cfg.System.GPU && !gpuFailed {
		if !collectGpuInfo(a) {
			gpuFailed = true
		}
	}
	w32.PostRefresh(a.Hwnd)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		collectSysInfo(a, 0)

		if cfg.System.GPU && !gpuFailed {
			if !collectGpuInfo(a) {
				gpuFailed = true
			}
		}

		if cfg.System.Network {
			if counters, err := net.IOCounters(false); err == nil && len(counters) > 0 {
				now := time.Now()
				elapsed := now.Sub(prevTime).Seconds()
				if elapsed > 0 && !prevTime.IsZero() {
					var down, up float64
					if counters[0].BytesRecv >= prevRecv {
						down = float64(counters[0].BytesRecv-prevRecv) / elapsed
					}
					if counters[0].BytesSent >= prevSent {
						up = float64(counters[0].BytesSent-prevSent) / elapsed
					}
					a.State.Mu.Lock()
					a.State.NetDownBytesPerSec = down
					a.State.NetUpBytesPerSec = up
					a.State.Mu.Unlock()
				}
				prevRecv = counters[0].BytesRecv
				prevSent = counters[0].BytesSent
				prevTime = now
			}
		}

		w32.PostRefresh(a.Hwnd)
	}
}

// DiskInfoLoop collects disk usage every 60 seconds.
func (a *App) DiskInfoLoop() {
	collectDiskInfo(a)
	w32.PostRefresh(a.Hwnd)

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		collectDiskInfo(a)
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

// collectGpuInfo runs nvidia-smi and parses GPU utilization and VRAM.
// Returns false if nvidia-smi is unavailable (caller should stop retrying).
func collectGpuInfo(a *App) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		"nvidia-smi",
		"--query-gpu=utilization.gpu,memory.used,memory.total",
		"--format=csv,noheader,nounits",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		a.State.Mu.Lock()
		a.State.GpuAvailable = false
		a.State.Mu.Unlock()
		return false
	}

	line := strings.TrimSpace(string(out))
	// Handle multi-GPU: use first GPU only
	if idx := strings.Index(line, "\n"); idx >= 0 {
		line = line[:idx]
	}
	parts := strings.Split(line, ", ")
	if len(parts) < 3 {
		a.State.Mu.Lock()
		a.State.GpuAvailable = false
		a.State.Mu.Unlock()
		return false
	}

	gpuPct, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	vramUsedMB, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	vramTotalMB, err3 := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err1 != nil || err2 != nil || err3 != nil {
		a.State.Mu.Lock()
		a.State.GpuAvailable = false
		a.State.Mu.Unlock()
		return false
	}

	a.State.Mu.Lock()
	a.State.GpuAvailable = true
	a.State.GpuPercent = gpuPct
	a.State.VramUsedGB = vramUsedMB / 1024.0
	a.State.VramTotalGB = vramTotalMB / 1024.0
	a.State.Mu.Unlock()
	return true
}

func collectDiskInfo(a *App) {
	var drives []DiskDriveInfo
	for _, letter := range diskTargetLetters {
		path := fmt.Sprintf("%s:\\", letter)
		usage, err := disk.Usage(path)
		if err != nil || usage.Total == 0 {
			continue
		}
		drives = append(drives, DiskDriveInfo{
			Letter:  letter,
			Percent: usage.UsedPercent,
			UsedGB:  float64(usage.Used) / (1024 * 1024 * 1024),
			TotalGB: float64(usage.Total) / (1024 * 1024 * 1024),
		})
	}

	a.State.Mu.Lock()
	a.State.DiskDrives = drives
	a.State.Mu.Unlock()
}
