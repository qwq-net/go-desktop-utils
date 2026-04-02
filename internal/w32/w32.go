//go:build windows

// Package w32 provides low-level Win32 API declarations and helpers.
package w32

import (
	"fmt"
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
	ProcRegisterClassExW           = user32.NewProc("RegisterClassExW")
	ProcCreateWindowExW            = user32.NewProc("CreateWindowExW")
	ProcShowWindow                 = user32.NewProc("ShowWindow")
	ProcUpdateWindow               = user32.NewProc("UpdateWindow")
	ProcDestroyWindow              = user32.NewProc("DestroyWindow")
	ProcDefWindowProcW             = user32.NewProc("DefWindowProcW")
	ProcGetMessageW                = user32.NewProc("GetMessageW")
	ProcTranslateMessage           = user32.NewProc("TranslateMessage")
	ProcDispatchMessageW           = user32.NewProc("DispatchMessageW")
	ProcPostMessageW               = user32.NewProc("PostMessageW")
	ProcPostQuitMessage            = user32.NewProc("PostQuitMessage")
	ProcBeginPaint                 = user32.NewProc("BeginPaint")
	ProcEndPaint                   = user32.NewProc("EndPaint")
	ProcInvalidateRect             = user32.NewProc("InvalidateRect")
	ProcSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	ProcSetWindowPos               = user32.NewProc("SetWindowPos")
	ProcLoadCursorW                = user32.NewProc("LoadCursorW")
	ProcLoadIconW                  = user32.NewProc("LoadIconW")
	ProcSetProcessDPIAware         = user32.NewProc("SetProcessDPIAware")
	ProcCreatePopupMenu            = user32.NewProc("CreatePopupMenu")
	ProcAppendMenuW                = user32.NewProc("AppendMenuW")
	ProcTrackPopupMenu             = user32.NewProc("TrackPopupMenu")
	ProcDestroyMenu                = user32.NewProc("DestroyMenu")
	ProcGetCursorPos               = user32.NewProc("GetCursorPos")
	ProcSetForegroundWindow        = user32.NewProc("SetForegroundWindow")
	ProcFindWindowW                = user32.NewProc("FindWindowW")
	ProcFindWindowExW              = user32.NewProc("FindWindowExW")
	ProcSendMessageTimeoutW        = user32.NewProc("SendMessageTimeoutW")
	ProcEnumWindows                = user32.NewProc("EnumWindows")
	ProcSetParent                  = user32.NewProc("SetParent")
	ProcSetWindowLongPtrW          = user32.NewProc("SetWindowLongPtrW")
	ProcEnumDisplayMonitors        = user32.NewProc("EnumDisplayMonitors")
	ProcGetMonitorInfoW            = user32.NewProc("GetMonitorInfoW")
)

// gdi32.dll
var (
	ProcCreateFontW           = gdi32.NewProc("CreateFontW")
	ProcSelectObject          = gdi32.NewProc("SelectObject")
	ProcDeleteObject          = gdi32.NewProc("DeleteObject")
	ProcSetBkMode             = gdi32.NewProc("SetBkMode")
	ProcSetTextColor          = gdi32.NewProc("SetTextColor")
	ProcCreateSolidBrush      = gdi32.NewProc("CreateSolidBrush")
	ProcAddFontMemResourceEx  = gdi32.NewProc("AddFontMemResourceEx")
	ProcRemoveFontMemResourceEx = gdi32.NewProc("RemoveFontMemResourceEx")
	ProcFillRect              = user32.NewProc("FillRect")
	ProcDrawTextW             = user32.NewProc("DrawTextW")
)

// shell32.dll
var (
	ProcShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	ProcShellExecuteW    = shell32.NewProc("ShellExecuteW")
)

// kernel32.dll
var (
	ProcGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

// Window styles
const (
	WS_POPUP          = 0x80000000
	WS_CHILD          = 0x40000000
	WS_VISIBLE        = 0x10000000
	WS_EX_LAYERED     = 0x00080000
	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_TOOLWINDOW  = 0x00000080
	WS_EX_NOACTIVATE  = 0x08000000
)

// SetLayeredWindowAttributes flags
const (
	LWA_COLORKEY = 0x00000001
	LWA_ALPHA    = 0x00000002
)

// SetWindowPos
const (
	HWND_BOTTOM = uintptr(1)
	SWP_NOMOVE  = 0x0002
	SWP_NOSIZE  = 0x0001
)

// GWL_STYLE
var GWL_STYLE = ^uintptr(15) // -16

// Window messages
const (
	WM_DESTROY    = 0x0002
	WM_PAINT      = 0x000F
	WM_ERASEBKGND = 0x0014
	WM_APP        = 0x8000
	WM_RBUTTONUP  = 0x0205

	WM_TRAYICON = WM_APP + 1
	WM_REFRESH  = WM_APP + 2
)

// SendMessageTimeout
const SMTO_NORMAL = 0x0000

// Shell_NotifyIcon
const (
	NIM_ADD     = 0x00000000
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
const MONITORINFOF_PRIMARY = 0x00000001

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

type MonitorDesc struct {
	WorkArea  RECT
	IsPrimary bool
}

// Helpers

func RGB(r, g, b byte) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func UTF16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func CopyToUint16Array(dst []uint16, src string) {
	s, _ := syscall.UTF16FromString(src)
	copy(dst, s)
}

// Wrappers

func GetModuleHandle() uintptr {
	h, _, _ := ProcGetModuleHandleW.Call(0)
	return h
}

func LoadCursor(hInstance uintptr, cursorID uintptr) uintptr {
	h, _, _ := ProcLoadCursorW.Call(hInstance, cursorID)
	return h
}

func LoadIcon(hInstance uintptr, iconID uintptr) uintptr {
	h, _, _ := ProcLoadIconW.Call(hInstance, iconID)
	return h
}

func CreateSolidBrush(color uint32) uintptr {
	h, _, _ := ProcCreateSolidBrush.Call(uintptr(color))
	return h
}

func SelectObject(hdc, obj uintptr) uintptr {
	h, _, _ := ProcSelectObject.Call(hdc, obj)
	return h
}

func DeleteObject(obj uintptr) {
	ProcDeleteObject.Call(obj)
}

func SetBkMode(hdc uintptr, mode int) {
	ProcSetBkMode.Call(hdc, uintptr(mode))
}

func SetTextColor(hdc uintptr, color uint32) {
	ProcSetTextColor.Call(hdc, uintptr(color))
}

func FillRect(hdc uintptr, rc *RECT, brush uintptr) {
	ProcFillRect.Call(hdc, uintptr(unsafe.Pointer(rc)), brush)
}

func DrawText(hdc uintptr, text string, rc *RECT, format uint32) int32 {
	s, _ := syscall.UTF16FromString(text)
	ret, _, _ := ProcDrawTextW.Call(
		hdc,
		uintptr(unsafe.Pointer(&s[0])),
		uintptr(len(s)-1),
		uintptr(unsafe.Pointer(rc)),
		uintptr(format),
	)
	return int32(ret)
}

func AddFontMemResource(data []byte) (uintptr, error) {
	var numFonts uint32
	handle, _, err := ProcAddFontMemResourceEx.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		0,
		uintptr(unsafe.Pointer(&numFonts)),
	)
	if handle == 0 {
		return 0, fmt.Errorf("AddFontMemResourceEx: %v", err)
	}
	return handle, nil
}

func CreateGDIFont(name string, size int, bold bool) uintptr {
	weight := 400
	if bold {
		weight = 700
	}
	h, _, _ := ProcCreateFontW.Call(
		uintptr(-size),
		0, 0, 0,
		uintptr(weight),
		0, 0, 0,
		1,
		0, 0, 4, 0,
		uintptr(unsafe.Pointer(UTF16Ptr(name))),
	)
	return h
}

func PostRefresh(hwnd uintptr) {
	ProcPostMessageW.Call(hwnd, WM_REFRESH, 0, 0)
}

// Desktop embedding

func FindDesktopWorkerW() uintptr {
	progman, _, _ := ProcFindWindowW.Call(
		uintptr(unsafe.Pointer(UTF16Ptr("Progman"))),
		0,
	)
	if progman == 0 {
		return 0
	}

	ProcSendMessageTimeoutW.Call(
		progman, 0x052C, 0, 0,
		SMTO_NORMAL, 1000, 0,
	)

	var workerW uintptr
	ProcEnumWindows.Call(
		syscall.NewCallback(func(hwnd, lParam uintptr) uintptr {
			defView, _, _ := ProcFindWindowExW.Call(
				hwnd, 0,
				uintptr(unsafe.Pointer(UTF16Ptr("SHELLDLL_DefView"))),
				0,
			)
			if defView != 0 {
				workerW, _, _ = ProcFindWindowExW.Call(
					0, hwnd,
					uintptr(unsafe.Pointer(UTF16Ptr("WorkerW"))),
					0,
				)
			}
			return 1
		}),
		0,
	)

	return workerW
}

func EmbedInDesktop(hwnd, workerW uintptr) {
	ProcSetWindowLongPtrW.Call(hwnd, GWL_STYLE,
		uintptr(WS_CHILD|WS_VISIBLE))
	ProcSetParent.Call(hwnd, workerW)
}

// Monitor enumeration

func EnumMonitors() []MonitorDesc {
	var monitors []MonitorDesc

	ProcEnumDisplayMonitors.Call(0, 0,
		syscall.NewCallback(func(hMonitor, hdcMonitor, lprcMonitor, dwData uintptr) uintptr {
			var mi MONITORINFO
			mi.CbSize = uint32(unsafe.Sizeof(mi))
			ProcGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))
			monitors = append(monitors, MonitorDesc{
				WorkArea:  mi.RcWork,
				IsPrimary: mi.DwFlags&MONITORINFOF_PRIMARY != 0,
			})
			return 1
		}),
		0,
	)

	sort.Slice(monitors, func(i, j int) bool {
		if monitors[i].IsPrimary != monitors[j].IsPrimary {
			return monitors[i].IsPrimary
		}
		return monitors[i].WorkArea.Left < monitors[j].WorkArea.Left
	})

	return monitors
}

func CalcPosition(workArea RECT, alignment string, marginX, marginY, width, height int) (int, int) {
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

func ShellOpen(path string) {
	ProcShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(UTF16Ptr("open"))),
		uintptr(unsafe.Pointer(UTF16Ptr(path))),
		0, 0,
		SW_SHOWNORMAL,
	)
}
