package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #942: the site's Windows download is blackpie-windows-amd64.exe, and Add
// to PATH put its folder on PATH, where a terminal finds no `blackpie`. Add
// writes a blackpie.cmd beside a program not named blackpie.exe that runs it by
// its folder and name, which an update keeps, and the Command line reads
// that blackpie.cmd as this app; one written for another program isn't.
func TestCLIShim(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "blackpie-windows-amd64.exe")
	if err := os.WriteFile(exe, []byte("MZ"), 0o755); err != nil {
		t.Fatal(err)
	}
	p, err := writeShim(exe)
	if err != nil {
		t.Fatal(err)
	}
	if p != filepath.Join(dir, "blackpie.cmd") {
		t.Fatalf("shim at %q", p)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "\"%~dp0blackpie-windows-amd64.exe\" %*\r\n") || !strings.HasPrefix(string(b), "@echo off\r\n") {
		t.Fatalf("blackpie.cmd:\n%q", b)
	}
	if !shimRuns(p, exe) {
		t.Fatal("the blackpie.cmd written isn't read as running this app")
	}
	if shimRuns(p, filepath.Join(dir, "blackpie-old.exe")) {
		t.Fatal("blackpie.cmd read as running another program")
	}
	// renamed, as a second download is: written again for the new name
	exe2 := filepath.Join(dir, "blackpie-windows-amd64 (1).exe")
	if _, err := writeShim(exe2); err != nil {
		t.Fatal(err)
	}
	if !shimRuns(p, exe2) || shimRuns(p, exe) {
		t.Fatal("blackpie.cmd not written again for the renamed program")
	}

	// blackpie.exe is found by its own name: nothing written
	dir2 := t.TempDir()
	if p, err := writeShim(filepath.Join(dir2, "Blackpie.exe")); err != nil || p != "" {
		t.Fatalf("blackpie.exe given a shim: %q, %v", p, err)
	}
	if _, err := os.Stat(filepath.Join(dir2, "blackpie.cmd")); !os.IsNotExist(err) {
		t.Fatal("blackpie.cmd written beside blackpie.exe")
	}
}
