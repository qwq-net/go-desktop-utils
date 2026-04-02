//go:build windows

package fonts

import (
	_ "embed"
	"fmt"

	"go-desktop-utils/internal/w32"
)

//go:embed Inter-Regular.ttf
var interRegular []byte

//go:embed Inter-Bold.ttf
var interBold []byte

// handles stores font resource handles for cleanup.
var handles []uintptr

// Install registers the embedded Inter fonts so that GDI can use them
// by family name. Call once at startup before any font creation.
func Install() error {
	for _, data := range [][]byte{interRegular, interBold} {
		h, err := w32.AddFontMemResource(data)
		if err != nil {
			return fmt.Errorf("install font: %w", err)
		}
		handles = append(handles, h)
	}
	return nil
}

// Uninstall removes the registered fonts. Call on shutdown if needed.
func Uninstall() {
	for _, h := range handles {
		w32.ProcRemoveFontMemResourceEx.Call(h)
	}
	handles = nil
}

