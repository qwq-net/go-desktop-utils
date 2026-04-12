//go:build windows

package widget

import (
	"fmt"
	"math"
	"sync"
	"syscall"
	"unsafe"

	"go-desktop-utils/internal/w32"
)

// Control IDs for gamma dialog
const (
	idcMonitorCombo uint16 = 101
	idcGammaLabel   uint16 = 102
	idcGammaSlider  uint16 = 103
	idcResetBtn     uint16 = 104
)

// Slider range: 50-300 maps to gamma 0.50-3.00
const (
	gammaSliderMin = 50
	gammaSliderMax = 300
	gammaDefault   = 100
)

var (
	gammaClassOnce sync.Once
	gammaState     *gammaDialogState
)

type gammaDialogState struct {
	hwnd     uintptr
	hCombo   uintptr
	hSlider  uintptr
	hLabel   uintptr
	monitors []w32.MonitorDesc
	gammas   []float64
}

// ShowGammaDialog opens the gamma setting dialog.
// If the dialog is already open, it brings it to the foreground.
func ShowGammaDialog() {
	if gammaState != nil && gammaState.hwnd != 0 {
		w32.ProcSetForegroundWindow.Call(gammaState.hwnd)
		return
	}

	hInstance := w32.GetModuleHandle()
	gammaClassOnce.Do(func() {
		registerGammaClass(hInstance)
	})

	monitors := w32.EnumMonitors()
	if len(monitors) == 0 {
		return
	}

	// Preserve gamma values from previous dialog session
	gammas := make([]float64, len(monitors))
	for i := range gammas {
		gammas[i] = 1.0
	}
	if gammaState != nil {
		for i := range gammas {
			if i < len(gammaState.gammas) {
				gammas[i] = gammaState.gammas[i]
			}
		}
	}

	gammaState = &gammaDialogState{
		monitors: monitors,
		gammas:   gammas,
	}

	// Center on primary monitor
	wa := monitors[0].WorkArea
	width, height := 400, 220
	x := int(wa.Left) + (int(wa.Right-wa.Left)-width)/2
	y := int(wa.Top) + (int(wa.Bottom-wa.Top)-height)/2

	hwnd, _, _ := w32.ProcCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(w32.UTF16Ptr("GoDesktopGamma"))),
		uintptr(unsafe.Pointer(w32.UTF16Ptr("Gamma Setting"))),
		uintptr(w32.WS_CAPTION|w32.WS_SYSMENU),
		uintptr(x), uintptr(y),
		uintptr(width), uintptr(height),
		0, 0, hInstance, 0,
	)
	if hwnd == 0 {
		return
	}
	gammaState.hwnd = hwnd

	initGammaControls(hwnd, hInstance)

	w32.ProcShowWindow.Call(hwnd, w32.SW_SHOW)
	w32.ProcUpdateWindow.Call(hwnd)
}

// ResetAllGamma restores gamma to 1.0 for monitors modified by this app.
func ResetAllGamma() {
	if gammaState == nil {
		return
	}
	for i, gamma := range gammaState.gammas {
		if gamma != 1.0 && i < len(gammaState.monitors) {
			applyGamma(gammaState.monitors[i].DeviceName, 1.0)
		}
	}
}

func registerGammaClass(hInstance uintptr) {
	wc := w32.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(w32.WNDCLASSEX{})),
		Style:         w32.CS_HREDRAW | w32.CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(gammaWndProc),
		HInstance:     hInstance,
		HCursor:       w32.LoadCursor(0, w32.IDC_ARROW),
		HbrBackground: w32.COLOR_BTNFACE + 1,
		LpszClassName: w32.UTF16Ptr("GoDesktopGamma"),
		HIcon:         w32.LoadIcon(0, w32.IDI_APPLICATION),
	}
	w32.ProcRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
}

func initGammaControls(hwnd, hInstance uintptr) {
	hFont := w32.GetStockFont()

	// "Monitor:" label
	hLbl := createChildControl("STATIC", "Monitor:", w32.SS_LEFT,
		16, 18, 65, 20, hwnd, 0, hInstance)
	w32.SendMessage(hLbl, w32.WM_SETFONT, hFont, 1)

	// Monitor combo box
	gammaState.hCombo = createChildControl("COMBOBOX", "",
		w32.CBS_DROPDOWNLIST|w32.CBS_HASSTRINGS,
		85, 14, 280, 200, hwnd, int(idcMonitorCombo), hInstance)
	w32.SendMessage(gammaState.hCombo, w32.WM_SETFONT, hFont, 1)

	for i, m := range gammaState.monitors {
		label := fmt.Sprintf("Display %d", i+1)
		if m.IsPrimary {
			label += " (Primary)"
		}
		w32.SendMessage(gammaState.hCombo, w32.CB_ADDSTRING, 0,
			uintptr(unsafe.Pointer(w32.UTF16Ptr(label))))
	}
	w32.SendMessage(gammaState.hCombo, w32.CB_SETCURSEL, 0, 0)

	// "Gamma: X.XX" label
	gammaState.hLabel = createChildControl("STATIC",
		fmt.Sprintf("Gamma: %.2f", gammaState.gammas[0]), w32.SS_LEFT,
		16, 54, 200, 20, hwnd, int(idcGammaLabel), hInstance)
	w32.SendMessage(gammaState.hLabel, w32.WM_SETFONT, hFont, 1)

	// Trackbar slider
	gammaState.hSlider = createChildControl("msctls_trackbar32", "",
		w32.TBS_HORZ|w32.TBS_AUTOTICKS,
		16, 80, 350, 36, hwnd, int(idcGammaSlider), hInstance)
	w32.SendMessage(gammaState.hSlider, w32.TBM_SETRANGE, 1,
		w32.MAKELONG(gammaSliderMin, gammaSliderMax))
	w32.SendMessage(gammaState.hSlider, w32.TBM_SETTICFREQ, 25, 0)
	w32.SendMessage(gammaState.hSlider, w32.TBM_SETPOS, 1,
		uintptr(int(gammaState.gammas[0]*100)))

	// Reset button
	hBtn := createChildControl("BUTTON", "Reset", w32.BS_PUSHBUTTON,
		290, 130, 80, 30, hwnd, int(idcResetBtn), hInstance)
	w32.SendMessage(hBtn, w32.WM_SETFONT, hFont, 1)
}

func createChildControl(className, text string, style uint32,
	x, y, w, h int, parent uintptr, id int, hInstance uintptr) uintptr {
	hwnd, _, _ := w32.ProcCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(w32.UTF16Ptr(className))),
		uintptr(unsafe.Pointer(w32.UTF16Ptr(text))),
		uintptr(style|w32.WS_CHILD|w32.WS_VISIBLE),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		parent, uintptr(id), hInstance, 0,
	)
	return hwnd
}

func gammaWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case w32.WM_COMMAND:
		id := w32.LOWORD(wParam)
		notify := w32.HIWORD(wParam)
		switch {
		case id == idcMonitorCombo && notify == w32.CBN_SELCHANGE:
			onGammaMonitorChanged()
		case id == idcResetBtn:
			onGammaReset()
		}
		return 0

	case w32.WM_HSCROLL:
		if gammaState != nil && lParam == gammaState.hSlider {
			onGammaSliderChanged()
		}
		return 0

	case w32.WM_CLOSE:
		w32.ProcDestroyWindow.Call(hwnd)
		return 0

	case w32.WM_DESTROY:
		if gammaState != nil {
			gammaState.hwnd = 0
		}
		return 0
	}

	ret, _, _ := w32.ProcDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

func onGammaMonitorChanged() {
	if gammaState == nil {
		return
	}
	idx := int(w32.SendMessage(gammaState.hCombo, w32.CB_GETCURSEL, 0, 0))
	if idx < 0 || idx >= len(gammaState.gammas) {
		return
	}
	gamma := gammaState.gammas[idx]
	w32.SendMessage(gammaState.hSlider, w32.TBM_SETPOS, 1, uintptr(int(gamma*100)))
	setGammaLabelText(gamma)
}

func onGammaSliderChanged() {
	if gammaState == nil {
		return
	}
	pos := int(w32.SendMessage(gammaState.hSlider, w32.TBM_GETPOS, 0, 0))
	gamma := float64(pos) / 100.0

	idx := int(w32.SendMessage(gammaState.hCombo, w32.CB_GETCURSEL, 0, 0))
	if idx < 0 || idx >= len(gammaState.gammas) {
		return
	}

	gammaState.gammas[idx] = gamma
	setGammaLabelText(gamma)
	applyGamma(gammaState.monitors[idx].DeviceName, gamma)
}

func onGammaReset() {
	if gammaState == nil {
		return
	}
	idx := int(w32.SendMessage(gammaState.hCombo, w32.CB_GETCURSEL, 0, 0))
	if idx < 0 || idx >= len(gammaState.gammas) {
		return
	}
	gammaState.gammas[idx] = 1.0
	w32.SendMessage(gammaState.hSlider, w32.TBM_SETPOS, 1, uintptr(gammaDefault))
	setGammaLabelText(1.0)
	applyGamma(gammaState.monitors[idx].DeviceName, 1.0)
}

func setGammaLabelText(gamma float64) {
	text := fmt.Sprintf("Gamma: %.2f", gamma)
	w32.ProcSendMessageW.Call(gammaState.hLabel, w32.WM_SETTEXT, 0,
		uintptr(unsafe.Pointer(w32.UTF16Ptr(text))))
}

func calcGammaRamp(gamma float64) w32.GAMMARAMP {
	var ramp w32.GAMMARAMP
	for i := 0; i < 256; i++ {
		value := math.Pow(float64(i)/255.0, 1.0/gamma) * 65535.0
		v := uint16(math.Min(math.Max(value, 0), 65535))
		ramp.Red[i] = v
		ramp.Green[i] = v
		ramp.Blue[i] = v
	}
	return ramp
}

func applyGamma(deviceName string, gamma float64) {
	ramp := calcGammaRamp(gamma)
	hdc, _, _ := w32.ProcCreateDCW.Call(
		0,
		uintptr(unsafe.Pointer(w32.UTF16Ptr(deviceName))),
		0, 0,
	)
	if hdc == 0 {
		return
	}
	defer w32.ProcDeleteDC.Call(hdc)
	w32.ProcSetDeviceGammaRamp.Call(hdc, uintptr(unsafe.Pointer(&ramp)))
}
