//go:build darwin && cgo

package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestDockFlapFrames writes the Dock tile's frames, cut from the icon of
// BLACKPIE_DOCK_APP (an installed blackpie.app), into BLACKPIE_DOCK_FRAMES, to
// look at by eye.
func TestDockFlapFrames(t *testing.T) {
	app, dir := os.Getenv("BLACKPIE_DOCK_APP"), os.Getenv("BLACKPIE_DOCK_FRAMES")
	if app == "" || dir == "" {
		t.Skip("set BLACKPIE_DOCK_APP and BLACKPIE_DOCK_FRAMES")
	}
	for f := 0; f <= 30; f++ {
		if !dockFramePNG(app, filepath.Join(dir, fmt.Sprintf("f%02d.png", f)), float64(f)/30) {
			t.Fatalf("frame %d not written", f)
		}
	}
}
