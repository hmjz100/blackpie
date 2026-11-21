//go:build !windows

package gui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// A second blackpie on the same config finds the first one's lock held, a
// copy elsewhere (blackpie-dev) has its own, and a lock whose holder ended,
// as when the Mac shuts down with blackpie running, is free again (Crispin
// on Discord: two blackpies after each boot).
func TestOneBlackpieAtATime(t *testing.T) {
	dir := t.TempDir()
	if os.Getenv("BLACKPIE_HOLD_LOCK") != "" {
		if held, err := holdInstance(os.Getenv("BLACKPIE_HOLD_LOCK")); err != nil || !held {
			os.Exit(3)
		}
		os.Stdout.WriteString("held\n")
		select {}
	}
	lock := instanceLock(dir, "/Applications/Blackpie.app/Contents/MacOS/blackpie")
	if dev := instanceLock(dir, "/src/blackpie/blackpie-dev"); dev == lock {
		t.Fatal("blackpie-dev shares the app's lock")
	}
	// every installed copy shares the lock: another blackpie.app (one the
	// Mac reopens, one run translocated from Downloads) and Homebrew's of
	// any version (inaction on Discord: two birds after a restart)
	for _, other := range []string{
		"/Users/u/Downloads/blackpie.app/Contents/MacOS/blackpie",
		"/private/var/folders/x/T/AppTranslocation/1F2E/d/blackpie.app/Contents/MacOS/blackpie",
		"/opt/homebrew/Cellar/blackpie/0.1.900/bin/blackpie",
	} {
		if instanceLock(dir, other) != lock {
			t.Errorf("%s has a lock of its own", other)
		}
	}
	if filepath.Dir(lock) != dir {
		t.Fatalf("lock %s not in the config dir", lock)
	}

	// the first blackpie: another process, killed as a shutdown would
	cmd := exec.Command(os.Args[0], "-test.run=^TestOneBlackpieAtATime$")
	cmd.Env = append(os.Environ(), "BLACKPIE_HOLD_LOCK="+lock)
	out, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 5)
	if _, err := out.Read(buf); err != nil || string(buf[:4]) != "held" {
		cmd.Process.Kill()
		t.Fatalf("the first blackpie didn't take the lock: %q %v", buf, err)
	}
	if held, err := holdInstance(lock); err != nil || held {
		cmd.Process.Kill()
		t.Fatalf("a second blackpie took the held lock: %v %v", held, err)
	}
	cmd.Process.Kill()
	cmd.Wait()
	if held, err := holdInstance(lock); err != nil || !held {
		t.Fatalf("the lock of a blackpie that was killed stayed held: %v %v", held, err)
	}
	heldInstance.Close()
	heldInstance = nil
}
