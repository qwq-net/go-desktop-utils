//go:build windows

package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"go-desktop-utils/internal/w32"
	"go-desktop-utils/internal/widget"
)

var app *widget.App

func main() {
	runtime.LockOSThread()

	w32.ProcSetProcessDPIAware.Call()

	cfg, err := widget.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	app = widget.NewApp()
	app.Config = cfg
	app.Colors = cfg.ParseColors()

	hInstance := w32.GetModuleHandle()

	if err := registerClass(hInstance); err != nil {
		fmt.Fprintf(os.Stderr, "RegisterClass failed: %v\n", err)
		os.Exit(1)
	}

	x, y := resolvePosition(cfg)

	if err := createWindow(hInstance, cfg, x, y); err != nil {
		fmt.Fprintf(os.Stderr, "CreateWindow failed: %v\n", err)
		os.Exit(1)
	}

	setupLayered(app.Hwnd, cfg)

	if workerW := w32.FindDesktopWorkerW(); workerW != 0 {
		w32.EmbedInDesktop(app.Hwnd, workerW)
		app.Embedded = true
	} else {
		w32.ProcSetWindowPos.Call(app.Hwnd, w32.HWND_BOTTOM, 0, 0, 0, 0,
			w32.SWP_NOMOVE|w32.SWP_NOSIZE)
	}

	if err := widget.AddTrayIcon(app.Hwnd, hInstance); err != nil {
		fmt.Fprintf(os.Stderr, "warning: tray icon failed: %v\n", err)
	}

	w32.ProcShowWindow.Call(app.Hwnd, w32.SW_SHOW)
	w32.ProcUpdateWindow.Call(app.Hwnd)

	if cfg.System.Enabled {
		go app.SysInfoLoop()
	}
	if cfg.Exchange.Enabled || cfg.Stocks.Enabled {
		go app.MarketDataLoop()
	}

	messageLoop()

	widget.RemoveTrayIcon()
}

func resolvePosition(cfg *widget.Config) (int, int) {
	monitors := w32.EnumMonitors()
	if len(monitors) == 0 {
		return 0, 0
	}
	idx := cfg.Window.Monitor
	if idx < 0 || idx >= len(monitors) {
		idx = 0
	}
	return w32.CalcPosition(
		monitors[idx].WorkArea,
		cfg.Window.Alignment,
		cfg.Window.MarginX,
		cfg.Window.MarginY,
		cfg.Window.Width,
		cfg.Window.Height,
	)
}

func registerClass(hInstance uintptr) error {
	className := w32.UTF16Ptr("GoDesktopWidget")

	wc := w32.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(w32.WNDCLASSEX{})),
		Style:         w32.CS_HREDRAW | w32.CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInstance,
		HCursor:       w32.LoadCursor(0, w32.IDC_ARROW),
		HbrBackground: 0,
		LpszClassName: className,
	}

	ret, _, err := w32.ProcRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		return fmt.Errorf("RegisterClassExW: %v", err)
	}
	return nil
}

func createWindow(hInstance uintptr, cfg *widget.Config, x, y int) error {
	exStyle := uint32(w32.WS_EX_LAYERED | w32.WS_EX_TRANSPARENT | w32.WS_EX_TOOLWINDOW | w32.WS_EX_NOACTIVATE)
	style := uint32(w32.WS_POPUP)

	hwnd, _, err := w32.ProcCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(w32.UTF16Ptr("GoDesktopWidget"))),
		uintptr(unsafe.Pointer(w32.UTF16Ptr(""))),
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
	app.Hwnd = hwnd
	return nil
}

func setupLayered(hwnd uintptr, cfg *widget.Config) {
	colorKey := w32.RGB(1, 0, 1)
	alpha := byte(cfg.Window.Opacity)
	w32.ProcSetLayeredWindowAttributes.Call(
		hwnd,
		uintptr(colorKey),
		uintptr(alpha),
		uintptr(w32.LWA_COLORKEY|w32.LWA_ALPHA),
	)
}

func wndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case w32.WM_PAINT:
		var ps w32.PAINTSTRUCT
		hdc, _, _ := w32.ProcBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		app.Render(hdc)
		w32.ProcEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case w32.WM_ERASEBKGND:
		return 1

	case w32.WM_REFRESH:
		w32.ProcInvalidateRect.Call(hwnd, 0, 1)
		return 0

	case w32.WM_TRAYICON:
		if lParam == w32.WM_RBUTTONUP {
			widget.ShowTrayMenu(app, hwnd)
		}
		return 0

	case w32.WM_DESTROY:
		w32.ProcPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := w32.ProcDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

func messageLoop() {
	var msg w32.MSG
	for {
		ret, _, _ := w32.ProcGetMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if ret == 0 || ret == ^uintptr(0) {
			break
		}
		w32.ProcTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		w32.ProcDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
