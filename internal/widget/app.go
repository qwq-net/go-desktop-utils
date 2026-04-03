//go:build windows

package widget

import (
	"sync"
	"time"

	"go-desktop-utils/internal/w32"
)

// App holds the global widget application state.
type App struct {
	Hwnd     uintptr
	Config   *Config
	Colors   ParsedColors
	State    *AppState
	Embedded bool
}

// DiskDriveInfo holds usage info for a single drive.
type DiskDriveInfo struct {
	Letter  string
	Percent float64
	UsedGB  float64
	TotalGB float64
}

// AppState holds all shared data, protected by Mu.
type AppState struct {
	Mu sync.Mutex

	CpuPercent float64
	MemPercent float64
	MemUsedGB  float64
	MemTotalGB float64

	GpuPercent   float64
	VramUsedGB   float64
	VramTotalGB  float64
	GpuAvailable bool

	DiskDrives []DiskDriveInfo

	NetDownBytesPerSec float64
	NetUpBytesPerSec   float64

	ExchangeRates map[string]float64
	ExchangeErr   bool
	ExchangeTime  time.Time

	StockPrices map[string]StockPrice
	StockErr    bool
	StockTime   time.Time
}

func NewApp() *App {
	return &App{
		State: &AppState{
			ExchangeRates: make(map[string]float64),
			StockPrices:   make(map[string]StockPrice),
		},
	}
}

func (a *App) ReloadConfig() {
	cfg, err := LoadConfig()
	if err != nil {
		return
	}
	a.Config = cfg
	a.Colors = cfg.ParseColors()

	// Reposition and resize
	monitors := w32.EnumMonitors()
	if len(monitors) > 0 {
		idx := cfg.Window.Monitor
		if idx < 0 || idx >= len(monitors) {
			idx = 0
		}
		x, y := w32.CalcPosition(
			monitors[idx].WorkArea,
			cfg.Window.Alignment,
			cfg.Window.MarginX,
			cfg.Window.MarginY,
			cfg.Window.Width,
			cfg.Window.Height,
		)
		w32.ProcSetWindowPos.Call(a.Hwnd, 0,
			uintptr(x), uintptr(y),
			uintptr(cfg.Window.Width), uintptr(cfg.Window.Height),
			0,
		)
	}

	// Update opacity
	colorKey := w32.RGB(1, 0, 1)
	alpha := byte(cfg.Window.Opacity)
	w32.ProcSetLayeredWindowAttributes.Call(
		a.Hwnd,
		uintptr(colorKey),
		uintptr(alpha),
		uintptr(w32.LWA_COLORKEY|w32.LWA_ALPHA),
	)
}
