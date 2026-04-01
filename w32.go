//go:build windows

package main

import (
	"sort"
	"syscall"
	"unsafe"
)

// DLL handles
var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
)

// user32.dll
var (
	procRegisterClassExW           = user32.NewProc("RegisterClassExW")
	procCreateWindowExW            = user32.NewProc("CreateWindowExW")
	procShowWindow                 = user32.NewProc("ShowWindow")
	procUpdateWindow               = user32.NewProc("UpdateWindow")
	procDestroyWindow              = user32.NewProc("DestroyWindow")
	procDefWindowProcW             = user32.NewProc("DefWindowProcW")
	procGetMessageW                = user32.NewProc("GetMessageW")
	procTranslateMessage           = user32.NewProc("TranslateMessage")
	procDispatchMessageW           = user32.NewProc("DispatchMessageW")
	procPostMessageW               = user32.NewProc("PostMessageW")
	procPostQuitMessage            = user32.NewProc("PostQuitMessage")
	procBeginPaint                 = user32.NewProc("BeginPaint")
	procEndPaint                   = user32.NewProc("EndPaint")
	procInvalidateRect             = user32.NewProc("InvalidateRect")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procSetWindowPos               = user32.NewProc("SetWindowPos")
	procLoadCursorW                = user32.NewProc("LoadCursorW")
	procLoadIconW                  = user32.NewProc("LoadIconW")
	procGetSystemMetrics           = user32.NewProc("GetSystemMetrics")
	procSetProcessDPIAware         = user32.NewProc("SetProcessDPIAware")
	procGetClientRect              = user32.NewProc("GetClientRect")
	procCreatePopupMenu            = user32.NewProc("CreatePopupMenu")
	procAppendMenuW                = user32.NewProc("AppendMenuW")
	procTrackPopupMenu             = user32.NewProc("TrackPopupMenu")
	procDestroyMenu                = user32.NewProc("DestroyMenu")
	procGetCursorPos               = user32.NewProc("GetCursorPos")
	procSetForegroundWindow        = user32.NewProc("SetForegroundWindow")
	procFindWindowW                = user32.NewProc("FindWindowW")
	procFindWindowExW              = user32.NewProc("FindWindowExW")
	procSendMessageTimeoutW        = user32.NewProc("SendMessageTimeoutW")
	procEnumWindows                = user32.NewProc("EnumWindows")
	procSetParent                  = user32.NewProc("SetParent")
	procSetWindowLongPtrW          = user32.NewProc("SetWindowLongPtrW")
	procEnumDisplayMonitors        = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW            = user32.NewProc("GetMonitorInfoW")
)

// gdi32.dll
var (
	procCreateFontW           = gdi32.NewProc("CreateFontW")
	procSelectObject          = gdi32.NewProc("SelectObject")
	procDeleteObject          = gdi32.NewProc("DeleteObject")
	procSetBkMode             = gdi32.NewProc("SetBkMode")
	procSetTextColor          = gdi32.NewProc("SetTextColor")
	procCreateSolidBrush      = gdi32.NewProc("CreateSolidBrush")
	procFillRect              = user32.NewProc("FillRect")
	procDrawTextW             = user32.NewProc("DrawTextW")
	procGetTextExtentPoint32W = gdi32.NewProc("GetTextExtentPoint32W")
)

// shell32.dll
var (
	procShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	procShellExecuteW    = shell32.NewProc("ShellExecuteW")
)

// kernel32.dll
var (
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

// Window styles
const (
	WS_POPUP          = 0x80000000
	WS_CHILD          = 0x40000000
	WS_VISIBLE        = 0x10000000
	WS_EX_LAYERED     = 0x00080000
	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_TOOLWINDOW  = 0x00000080
	WS_EX_TOPMOST     = 0x00000008
	WS_EX_NOACTIVATE  = 0x08000000
)

// SetLayeredWindowAttributes flags
const (
	LWA_COLORKEY = 0x00000001
	LWA_ALPHA    = 0x00000002
)

// SetWindowPos
const (
	HWND_TOPMOST = ^uintptr(0) // -1
	HWND_BOTTOM  = uintptr(1)
	SWP_NOMOVE   = 0x0002
	SWP_NOSIZE   = 0x0001
)

// SetWindowLongPtr index
var (
	gwlStyle = ^uintptr(15) // GWL_STYLE = -16
)

// Window messages
const (
	WM_DESTROY    = 0x0002
	WM_PAINT      = 0x000F
	WM_CLOSE      = 0x0010
	WM_ERASEBKGND = 0x0014
	WM_APP        = 0x8000
	WM_RBUTTONUP  = 0x0205
	WM_LBUTTONUP  = 0x0202

	WM_TRAYICON = WM_APP + 1
	WM_REFRESH  = WM_APP + 2
)

// SendMessageTimeout
const (
	SMTO_NORMAL = 0x0000
)

// Shell_NotifyIcon
const (
	NIM_ADD     = 0x00000000
	NIM_MODIFY  = 0x00000001
	NIM_DELETE  = 0x00000002
	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
)

// GDI
const (
	TRANSPARENT_BK = 1
	DT_LEFT        = 0x00000000
	DT_RIGHT       = 0x00000002
	DT_NOCLIP      = 0x00000100
	DT_SINGLELINE  = 0x00000020
)

// ShowWindow
const (
	SW_SHOW       = 5
	SW_HIDE       = 0
	SW_SHOWNORMAL = 1
)

// LoadCursor / LoadIcon
const (
	IDC_ARROW       = 32512
	IDI_APPLICATION = 32512
)

// Class styles
const (
	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001
)

// TrackPopupMenu / AppendMenu
const (
	TPM_BOTTOMALIGN = 0x0020
	TPM_LEFTALIGN   = 0x0000
	TPM_RETURNCMD   = 0x0100
	MF_STRING       = 0x00000000
	MF_SEPARATOR    = 0x00000800
)

// MonitorInfo
const (
	MONITORINFOF_PRIMARY = 0x00000001
)

// Structs

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type PAINTSTRUCT struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type SIZE struct {
	Cx, Cy int32
}

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

// MonitorDesc holds enumerated monitor info.
type MonitorDesc struct {
	WorkArea  RECT
	IsPrimary bool
}

// Helpers

func rgb(r, g, b byte) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func copyToUint16Array(dst []uint16, src string) {
	s, _ := syscall.UTF16FromString(src)
	copy(dst, s)
}

// Wrappers

func getModuleHandle() uintptr {
	h, _, _ := procGetModuleHandleW.Call(0)
	return h
}

func loadCursor(hInstance uintptr, cursorID uintptr) uintptr {
	h, _, _ := procLoadCursorW.Call(hInstance, cursorID)
	return h
}

func loadIcon(hInstance uintptr, iconID uintptr) uintptr {
	h, _, _ := procLoadIconW.Call(hInstance, iconID)
	return h
}

func createSolidBrush(color uint32) uintptr {
	h, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return h
}

func selectObject(hdc, obj uintptr) uintptr {
	h, _, _ := procSelectObject.Call(hdc, obj)
	return h
}

func deleteObject(obj uintptr) {
	procDeleteObject.Call(obj)
}

func setBkMode(hdc uintptr, mode int) {
	procSetBkMode.Call(hdc, uintptr(mode))
}

func setTextColor(hdc uintptr, color uint32) {
	procSetTextColor.Call(hdc, uintptr(color))
}

func fillRect(hdc uintptr, rc *RECT, brush uintptr) {
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(rc)), brush)
}

func drawText(hdc uintptr, text string, rc *RECT, format uint32) int32 {
	s, _ := syscall.UTF16FromString(text)
	ret, _, _ := procDrawTextW.Call(
		hdc,
		uintptr(unsafe.Pointer(&s[0])),
		uintptr(len(s)-1),
		uintptr(unsafe.Pointer(rc)),
		uintptr(format),
	)
	return int32(ret)
}

func getTextExtent(hdc uintptr, text string) SIZE {
	s, _ := syscall.UTF16FromString(text)
	var sz SIZE
	procGetTextExtentPoint32W.Call(
		hdc,
		uintptr(unsafe.Pointer(&s[0])),
		uintptr(len(s)-1),
		uintptr(unsafe.Pointer(&sz)),
	)
	return sz
}

func createGDIFont(name string, size int, bold bool) uintptr {
	weight := 400
	if bold {
		weight = 700
	}
	h, _, _ := procCreateFontW.Call(
		uintptr(-size),
		0, 0, 0,
		uintptr(weight),
		0, 0, 0,
		1,
		0, 0, 4, 0,
		uintptr(unsafe.Pointer(utf16Ptr(name))),
	)
	return h
}

func postRefresh(hwnd uintptr) {
	procPostMessageW.Call(hwnd, WM_REFRESH, 0, 0)
}

// Desktop embedding: find WorkerW behind desktop icons

func findDesktopWorkerW() uintptr {
	progman, _, _ := procFindWindowW.Call(
		uintptr(unsafe.Pointer(utf16Ptr("Progman"))),
		0,
	)
	if progman == 0 {
		return 0
	}

	// Send undocumented message to spawn WorkerW
	procSendMessageTimeoutW.Call(
		progman, 0x052C, 0, 0,
		SMTO_NORMAL, 1000, 0,
	)

	var workerW uintptr
	procEnumWindows.Call(
		syscall.NewCallback(func(hwnd, lParam uintptr) uintptr {
			defView, _, _ := procFindWindowExW.Call(
				hwnd, 0,
				uintptr(unsafe.Pointer(utf16Ptr("SHELLDLL_DefView"))),
				0,
			)
			if defView != 0 {
				// WorkerW is the next sibling after the window containing SHELLDLL_DefView
				workerW, _, _ = procFindWindowExW.Call(
					0, hwnd,
					uintptr(unsafe.Pointer(utf16Ptr("WorkerW"))),
					0,
				)
			}
			return 1 // continue
		}),
		0,
	)

	return workerW
}

func embedInDesktop(hwnd, workerW uintptr) {
	// Change from WS_POPUP to WS_CHILD
	style, _, _ := procSetWindowLongPtrW.Call(hwnd, gwlStyle,
		uintptr(WS_CHILD|WS_VISIBLE))
	_ = style
	procSetParent.Call(hwnd, workerW)
}

// Monitor enumeration

func enumMonitors() []MonitorDesc {
	var monitors []MonitorDesc

	procEnumDisplayMonitors.Call(0, 0,
		syscall.NewCallback(func(hMonitor, hdcMonitor, lprcMonitor, dwData uintptr) uintptr {
			var mi MONITORINFO
			mi.CbSize = uint32(unsafe.Sizeof(mi))
			procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))
			monitors = append(monitors, MonitorDesc{
				WorkArea:  mi.RcWork,
				IsPrimary: mi.DwFlags&MONITORINFOF_PRIMARY != 0,
			})
			return 1
		}),
		0,
	)

	// Primary monitor first, then left-to-right
	sort.Slice(monitors, func(i, j int) bool {
		if monitors[i].IsPrimary != monitors[j].IsPrimary {
			return monitors[i].IsPrimary
		}
		return monitors[i].WorkArea.Left < monitors[j].WorkArea.Left
	})

	return monitors
}

// Position calculation

func calcPosition(workArea RECT, alignment string, marginX, marginY, width, height int) (int, int) {
	switch alignment {
	case "topLeft":
		return int(workArea.Left) + marginX, int(workArea.Top) + marginY
	case "bottomRight":
		return int(workArea.Right) - width - marginX, int(workArea.Bottom) - height - marginY
	case "bottomLeft":
		return int(workArea.Left) + marginX, int(workArea.Bottom) - height - marginY
	default: // "topRight"
		return int(workArea.Right) - width - marginX, int(workArea.Top) + marginY
	}
}

// ShellExecute wrapper

func shellOpen(path string) {
	procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr("open"))),
		uintptr(unsafe.Pointer(utf16Ptr(path))),
		0, 0,
		SW_SHOWNORMAL,
	)
}
