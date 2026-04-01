//go:build windows

package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// AppState holds all shared data, protected by mu.
type AppState struct {
	mu sync.Mutex

	cpuPercent float64
	memPercent float64
	memUsedGB  float64
	memTotalGB float64

	exchangeRates map[string]float64
	exchangeErr   bool
	exchangeTime  time.Time

	stockPrices map[string]StockPrice
	stockErr    bool
	stockTime   time.Time
}

var (
	appHwnd     uintptr
	appConfig   *Config
	appState    *AppState
	appColors   ParsedColors
	appEmbedded bool // true if embedded in desktop WorkerW
)

func main() {
	runtime.LockOSThread()

	procSetProcessDPIAware.Call()

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	appConfig = cfg
	appColors = cfg.ParseColors()

	appState = &AppState{
		exchangeRates: make(map[string]float64),
		stockPrices:   make(map[string]StockPrice),
	}

	hInstance := getModuleHandle()

	if err := registerClass(hInstance); err != nil {
		fmt.Fprintf(os.Stderr, "RegisterClass failed: %v\n", err)
		os.Exit(1)
	}

	// Determine position from monitor + alignment
	x, y := resolvePosition(cfg)

	if err := createWindow(hInstance, cfg, x, y); err != nil {
		fmt.Fprintf(os.Stderr, "CreateWindow failed: %v\n", err)
		os.Exit(1)
	}

	setupLayered(appHwnd, cfg)

	// Try embedding into desktop (behind all windows)
	if workerW := findDesktopWorkerW(); workerW != 0 {
		embedInDesktop(appHwnd, workerW)
		appEmbedded = true
	} else {
		// Fallback: push to bottom of z-order
		procSetWindowPos.Call(appHwnd, HWND_BOTTOM, 0, 0, 0, 0,
			SWP_NOMOVE|SWP_NOSIZE)
	}

	if err := addTrayIcon(appHwnd, hInstance); err != nil {
		fmt.Fprintf(os.Stderr, "warning: tray icon failed: %v\n", err)
	}

	procShowWindow.Call(appHwnd, SW_SHOW)
	procUpdateWindow.Call(appHwnd)

	go sysInfoLoop(appHwnd)
	go marketDataLoop(appHwnd, cfg)

	messageLoop()

	removeTrayIcon()
}

func resolvePosition(cfg *Config) (int, int) {
	monitors := enumMonitors()
	if len(monitors) == 0 {
		return 0, 0
	}
	idx := cfg.Window.Monitor
	if idx < 0 || idx >= len(monitors) {
		idx = 0
	}
	return calcPosition(
		monitors[idx].WorkArea,
		cfg.Window.Alignment,
		cfg.Window.MarginX,
		cfg.Window.MarginY,
		cfg.Window.Width,
		cfg.Window.Height,
	)
}

func registerClass(hInstance uintptr) error {
	className := utf16Ptr("GoDesktopWidget")

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInstance,
		HCursor:       loadCursor(0, IDC_ARROW),
		HbrBackground: 0,
		LpszClassName: className,
	}

	ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		return fmt.Errorf("RegisterClassExW: %v", err)
	}
	return nil
}

func createWindow(hInstance uintptr, cfg *Config, x, y int) error {
	// No TOPMOST — we embed in desktop or use HWND_BOTTOM
	exStyle := uint32(WS_EX_LAYERED | WS_EX_TRANSPARENT | WS_EX_TOOLWINDOW | WS_EX_NOACTIVATE)
	style := uint32(WS_POPUP)

	hwnd, _, err := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(utf16Ptr("GoDesktopWidget"))),
		uintptr(unsafe.Pointer(utf16Ptr(""))),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(cfg.Window.Width),
		uintptr(cfg.Window.Height),
		0, 0,
		hInstance,
		0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW: %v", err)
	}
	appHwnd = hwnd
	return nil
}

func setupLayered(hwnd uintptr, cfg *Config) {
	colorKey := rgb(1, 0, 1) // near-black to eliminate purple anti-aliasing fringe
	alpha := byte(cfg.Window.Opacity)
	procSetLayeredWindowAttributes.Call(
		hwnd,
		uintptr(colorKey),
		uintptr(alpha),
		uintptr(LWA_COLORKEY|LWA_ALPHA),
	)
}

// reloadConfig re-reads config.json and applies visual changes (position, opacity, colors).
func reloadConfig() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "reload config error: %v\n", err)
		return
	}
	appConfig = cfg
	appColors = cfg.ParseColors()

	// Reposition
	x, y := resolvePosition(cfg)
	procSetWindowPos.Call(appHwnd, 0,
		uintptr(x), uintptr(y),
		uintptr(cfg.Window.Width), uintptr(cfg.Window.Height),
		0,
	)

	// Update opacity
	setupLayered(appHwnd, cfg)

	// Redraw
	procInvalidateRect.Call(appHwnd, 0, 1)
}

func wndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		render(hdc)
		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_ERASEBKGND:
		return 1

	case WM_REFRESH:
		procInvalidateRect.Call(hwnd, 0, 1)
		return 0

	case WM_TRAYICON:
		if lParam == WM_RBUTTONUP {
			showTrayMenu(hwnd)
		}
		return 0

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

func messageLoop() {
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if ret == 0 || ret == ^uintptr(0) {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
