package update

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBlackpieAssets(t *testing.T) {
	if got, want := AppAsset(), "blackpie-darwin-"+runtime.GOARCH+".zip"; got != want {
		t.Fatalf("app asset %q, want %q", got, want)
	}
	before := GUI
	t.Cleanup(func() { GUI = before })
	for _, gui := range []bool{false, true} {
		GUI = gui
		prefix := "blackpie-cli-"
		if gui && runtime.GOOS != "darwin" {
			prefix = "blackpie-"
		}
		want := prefix + runtime.GOOS + "-" + runtime.GOARCH
		if runtime.GOOS == "windows" {
			want += ".exe"
		}
		if got := BinaryAsset(); got != want {
			t.Fatalf("GUI %v: binary asset %q, want %q", gui, got, want)
		}
	}
}

func TestBlackpieBundleExecutable(t *testing.T) {
	bundle := filepath.Join(t.TempDir(), "blackpie.app")
	if got := bundleExecutable(bundle); got != "blackpie" {
		t.Fatalf("fallback executable %q", got)
	}
	plist, err := os.ReadFile("../../build/darwin/Info.plist")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plist), "<string>io.github.hmjz100.blackpie</string>") || !strings.Contains(string(plist), "<string>blackpie</string></array>") {
		t.Fatal("bundle identity or compatible import scheme missing")
	}
	if err := os.MkdirAll(filepath.Join(bundle, "Contents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle, "Contents", "Info.plist"), plist, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := bundleExecutable(bundle); got != "blackpie" {
		t.Fatalf("bundled executable %q", got)
	}
}

func TestBlackpieAndLegacyCLIVersions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture")
	}
	for _, name := range []string{"blackpie", "blackpie"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), name)
			if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf '"+name+" 0.1.2000\\n'\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			if got := cliVersion(path); got != "0.1.2000" {
				t.Fatalf("version = %q", got)
			}
		})
	}
}

func TestBlackpieKeepsConsoleCompanion(t *testing.T) {
	if comName != "blackpie.com" || ComEnv != "BLACKPIE_COM" {
		t.Fatal("console companion compatibility changed")
	}
	exe := filepath.Join(t.TempDir(), "blackpie-windows-amd64.exe")
	p, err := writeCom([]byte("launcher"), exe)
	if err != nil || filepath.Base(p) != "blackpie.com" || !comRuns(p, exe) {
		t.Fatalf("companion %q, %v, must run renamed app", p, err)
	}
}
