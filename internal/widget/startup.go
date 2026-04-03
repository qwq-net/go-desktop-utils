//go:build windows

package widget

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	startupRegKey  = `Software\Microsoft\Windows\CurrentVersion\Run`
	startupRegName = "GoDesktopWidget"
)

// IsStartupEnabled checks whether the auto-start registry entry exists.
func IsStartupEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, startupRegKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(startupRegName)
	return err == nil
}

// SetStartupEnabled adds or removes the auto-start registry entry.
func SetStartupEnabled(enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, startupRegKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if enabled {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		return key.SetStringValue(startupRegName, exe)
	}
	return key.DeleteValue(startupRegName)
}
