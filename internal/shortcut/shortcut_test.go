package shortcut

import "testing"

func TestTarget(t *testing.T) {
	for in, want := range map[string]string{
		`C:\Tools\blackpie-windows-amd64.exe`:          `C:\Tools\blackpie-windows-amd64.exe`,
		`C:\Tools\blackpie.exe.old`:                    `C:\Tools\blackpie.exe`,
		`C:\Tools\blackpie.exe.old-3`:                  `C:\Tools\blackpie.exe`,
		`C:\Tools\blackpie.exe.new`:                    `C:\Tools\blackpie.exe`,
		`C:\Tools\BLACKPIE.EXE.OLD`:                    `C:\Tools\BLACKPIE.EXE`,
		`C:\old\blackpie.exe`:                          `C:\old\blackpie.exe`,
		`C:\Tools\blackpie.exe.older`:                  `C:\Tools\blackpie.exe.older`,
		`/usr/local/bin/blackpie`:                      `/usr/local/bin/blackpie`,
		`C:\Users\me\Downloads\blackpie (1).exe.old-2`: `C:\Users\me\Downloads\blackpie (1).exe`,
	} {
		if got := Target(in); got != want {
			t.Errorf("Target(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStale(t *testing.T) {
	exe := "/tools/blackpie.exe"
	for _, c := range []struct {
		target, dir string
		stale       bool
	}{
		{"", "", true},                            // no shortcut yet
		{"/tools/blackpie.exe", "/tools", false},    // already this one
		{"/TOOLS/Blackpie.EXE", "/Tools/", false},   // the same, in another case
		{"/downloads/blackpie.exe", "/tools", true}, // the exe was moved
		{"/tools/blackpie.exe", "/downloads", true}, // starts in another folder
		{"/tools/blackpie.exe", "", true},           // in no folder
	} {
		if got := Stale(c.target, c.dir, exe); got != c.stale {
			t.Errorf("Stale(%q, %q) = %v, want %v", c.target, c.dir, got, c.stale)
		}
	}
}
