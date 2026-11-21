package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommunityManifestCompatibility(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	write := func(s string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"name":"opencode-example","magpie":{"agent":"agent.js","middleware":"middleware.js","hooks":"hooks.js","options":{"enabled":true},"page":{"file":"index.html"},"icon":"https://example.test/icon.png"}}`)
	if f, _ := Agent(dir); f != filepath.Join(dir, "agent.js") {
		t.Fatalf("legacy agent: %q", f)
	}
	if f, _ := Middleware(dir); f != filepath.Join(dir, "middleware.js") {
		t.Fatalf("legacy middleware: %q", f)
	}
	if f := Gateway(dir); f != filepath.Join(dir, "hooks.js") {
		t.Fatalf("legacy hook: %q", f)
	}
	if opts := OptionsExample(dir); opts["enabled"] != true {
		t.Fatalf("legacy options: %v", opts)
	}
	if w, _ := Window(dir); w == nil || w.Page == nil || w.Page.File != filepath.Join(dir, "index.html") {
		t.Fatalf("legacy page: %+v", w)
	}
	var latest npmLatest
	if err := unmarshalManifest([]byte(`{"magpie":{"icon":"https://example.test/icon.png"}}`), &latest); err != nil || latest.Blackpie.Icon != "https://example.test/icon.png" {
		t.Fatalf("legacy market metadata: %+v, %v", latest, err)
	}
	write(`{"magpie":{"agent":"old.js","options":{"enabled":true}},"blackpie":{"agent":"new.js"}}`)
	if f, _ := Agent(dir); f != filepath.Join(dir, "new.js") {
		t.Fatalf("blackpie block did not win: %q", f)
	}
	if opts := OptionsExample(dir); opts != nil {
		t.Fatalf("blocks merged: %v", opts)
	}
	write(`{"magpie":{"agent":"old.js"},"blackpie":null}`)
	if f, _ := Agent(dir); f != "" {
		t.Fatalf("explicit null did not win: %q", f)
	}
	write(`{"magpie":`)
	if f, _ := Agent(dir); f != "" {
		t.Fatalf("malformed manifest accepted: %q", f)
	}
}
