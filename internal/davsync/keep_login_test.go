package davsync

import (
	"testing"

	"github.com/hmjz100/blackpie/internal/backup"
	"github.com/hmjz100/blackpie/internal/provider"
)

// Which account an agent is kept signed in to is this computer's own
// (provider.Mirror keeps it), so changing it uploads nothing, and two
// computers that differ only in it don't trade the providers part.
func TestKeepLoginIsNoChange(t *testing.T) {
	p := provider.Provider{ID: "codex", Name: "Codex"}
	home := backup.Bundle{Providers: []provider.Provider{p}}
	p.KeepLogin, p.KeepLoginAs, p.SwitchOnly = true, "work@example.com", true
	office := backup.Bundle{Providers: []provider.Provider{p}}
	if hashes(home)["providers"] != hashes(office)["providers"] {
		t.Fatal("the account kept signed in changed the providers hash")
	}
	if office.Providers[0].KeepLoginAs != "work@example.com" {
		t.Fatal("hashing changed the bundle")
	}
	p.Name = "Codex renamed"
	if hashes(home)["providers"] == hashes(backup.Bundle{Providers: []provider.Provider{p}})["providers"] {
		t.Fatal("a rename no longer changes the providers hash")
	}
}
