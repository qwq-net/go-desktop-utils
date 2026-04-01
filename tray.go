//go:build windows

package main

import (
	"fmt"
	"unsafe"
)

const (
	menuIDSettings = 1001
	menuIDReload   = 1002
	menuIDExit     = 1003
)

var trayIconData NOTIFYICONDATA

func addTrayIcon(hwnd, hInstance uintptr) error {
	hIcon := loadIcon(0, IDI_APPLICATION)

	trayIconData = NOTIFYICONDATA{
		HWnd:             hwnd,
		UID:              1,
		UFlags:           NIF_ICON | NIF_MESSAGE | NIF_TIP,
		UCallbackMessage: WM_TRAYICON,
		HIcon:            hIcon,
	}
	trayIconData.CbSize = uint32(unsafe.Sizeof(trayIconData))
	copyToUint16Array(trayIconData.SzTip[:], "Desktop Widget")

	ret, _, err := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&trayIconData)))
	if ret == 0 {
		return fmt.Errorf("Shell_NotifyIcon: %v", err)
	}
	return nil
}

func removeTrayIcon() {
	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&trayIconData)))
}

func showTrayMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer procDestroyMenu.Call(hMenu)

	procAppendMenuW.Call(hMenu, MF_STRING, menuIDSettings, uintptr(unsafe.Pointer(utf16Ptr("Settings"))))
	procAppendMenuW.Call(hMenu, MF_STRING, menuIDReload, uintptr(unsafe.Pointer(utf16Ptr("Reload Config"))))
	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(hMenu, MF_STRING, menuIDExit, uintptr(unsafe.Pointer(utf16Ptr("Exit"))))

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	procSetForegroundWindow.Call(hwnd)

	cmd, _, _ := procTrackPopupMenu.Call(
		hMenu,
		TPM_BOTTOMALIGN|TPM_LEFTALIGN|TPM_RETURNCMD,
		uintptr(pt.X), uintptr(pt.Y),
		0, hwnd, 0,
	)

	switch cmd {
	case menuIDSettings:
		shellOpen(configPath())
	case menuIDReload:
		reloadConfig()
	case menuIDExit:
		procDestroyWindow.Call(hwnd)
	}
}
