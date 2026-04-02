//go:build windows

package widget

import (
	"sync"
	"time"
)

// App holds the global widget application state.
type App struct {
	Hwnd     uintptr
	Config   *Config
	Colors   ParsedColors
	State    *AppState
	Embedded bool
}

// AppState holds all shared data, protected by Mu.
type AppState struct {
	Mu sync.Mutex

	CpuPercent float64
	MemPercent float64
	MemUsedGB  float64
	MemTotalGB float64

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
}
