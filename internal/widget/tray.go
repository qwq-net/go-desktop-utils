//go:build windows

package widget

import (
	"fmt"
	"unsafe"

	"go-desktop-utils/internal/w32"
)

const (
	menuIDSettings = 1001
	menuIDReload   = 1002
	menuIDExit     = 1003
)

var trayIconData w32.NOTIFYICONDATA

func AddTrayIcon(hwnd, hInstance uintptr) error {
	hIcon := w32.LoadIcon(0, w32.IDI_APPLICATION)

	trayIconData = w32.NOTIFYICONDATA{
		HWnd:             hwnd,
		UID:              1,
		UFlags:           w32.NIF_ICON | w32.NIF_MESSAGE | w32.NIF_TIP,
		UCallbackMessage: w32.WM_TRAYICON,
		HIcon:            hIcon,
	}
	trayIconData.CbSize = uint32(unsafe.Sizeof(trayIconData))
	w32.CopyToUint16Array(trayIconData.SzTip[:], "Desktop Widget")

	ret, _, err := w32.ProcShellNotifyIconW.Call(w32.NIM_ADD, uintptr(unsafe.Pointer(&trayIconData)))
	if ret == 0 {
		return fmt.Errorf("Shell_NotifyIcon: %v", err)
	}
	return nil
}

func RemoveTrayIcon() {
	w32.ProcShellNotifyIconW.Call(w32.NIM_DELETE, uintptr(unsafe.Pointer(&trayIconData)))
}

func ShowTrayMenu(app *App, hwnd uintptr) {
	hMenu, _, _ := w32.ProcCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer w32.ProcDestroyMenu.Call(hMenu)

	w32.ProcAppendMenuW.Call(hMenu, w32.MF_STRING, menuIDSettings, uintptr(unsafe.Pointer(w32.UTF16Ptr("Settings"))))
	w32.ProcAppendMenuW.Call(hMenu, w32.MF_STRING, menuIDReload, uintptr(unsafe.Pointer(w32.UTF16Ptr("Reload Config"))))
	w32.ProcAppendMenuW.Call(hMenu, w32.MF_SEPARATOR, 0, 0)
	w32.ProcAppendMenuW.Call(hMenu, w32.MF_STRING, menuIDExit, uintptr(unsafe.Pointer(w32.UTF16Ptr("Exit"))))

	var pt w32.POINT
	w32.ProcGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	w32.ProcSetForegroundWindow.Call(hwnd)

	cmd, _, _ := w32.ProcTrackPopupMenu.Call(
		hMenu,
		w32.TPM_BOTTOMALIGN|w32.TPM_LEFTALIGN|w32.TPM_RETURNCMD,
		uintptr(pt.X), uintptr(pt.Y),
		0, hwnd, 0,
	)

	switch cmd {
	case menuIDSettings:
		w32.ShellOpen(ConfigPath())
	case menuIDReload:
		app.ReloadConfig()
		w32.ProcInvalidateRect.Call(hwnd, 0, 1)
	case menuIDExit:
		w32.ProcDestroyWindow.Call(hwnd)
	}
}
