package cli

import "testing"

// blackpie web on loopback behind a reverse proxy (#1479) prints the proxy's
// link with the key when told the proxy's address, so nobody joins
// "/?k=…" onto it by hand: BLACKPIE_WEB_URL, else BLACKPIE_PUBLIC_URL with no
// port (one with a port isn't a proxy). None is printed without either.
func TestWebProxyLink(t *testing.T) {
	key := "/?k=0123456789abcdef"
	for _, c := range []struct{ env, want string }{
		{"https://blackpie.example.com", "https://blackpie.example.com/?k=0123456789abcdef"},
		{"https://blackpie.example.com/", "https://blackpie.example.com/?k=0123456789abcdef"},
		{"blackpie.example.com", "http://blackpie.example.com/?k=0123456789abcdef"},
		{"https://example.com/blackpie/", "https://example.com/blackpie/?k=0123456789abcdef"},
		{"http://192.168.1.20:3425", ""},
		{"", ""},
	} {
		t.Setenv("BLACKPIE_PUBLIC_URL", c.env)
		if got := proxyLink(key); got != c.want {
			t.Errorf("BLACKPIE_PUBLIC_URL=%q: %q, want %q", c.env, got, c.want)
		}
	}
	// the page's own proxy address, when the gateway's is elsewhere
	t.Setenv("BLACKPIE_PUBLIC_URL", "https://api.example.com")
	for _, c := range []struct{ env, want string }{
		{"https://blackpie.example.com/", "https://blackpie.example.com/?k=0123456789abcdef"},
		{"blackpie.example.com:8443", "https://blackpie.example.com:8443/?k=0123456789abcdef"},
		{"https://blackpie.example.com/?x=1", ""},
		{"ftp://blackpie.example.com", ""},
	} {
		t.Setenv("BLACKPIE_WEB_URL", c.env)
		if got := proxyLink(key); got != c.want {
			t.Errorf("BLACKPIE_WEB_URL=%q: %q, want %q", c.env, got, c.want)
		}
	}
}

// `blackpie web` opens the app's page when the app serves it on the port it
// would take, and starts its own on any other.
func TestWebSamePort(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"127.0.0.1:3430", "127.0.0.1:3430", true},
		{"0.0.0.0:3430", "127.0.0.1:3430", true},
		{"127.0.0.1:3430", "0.0.0.0:3431", false},
		{"", "127.0.0.1:3430", false},
		{"127.0.0.1:3430", "3430", false},
	} {
		if got := samePort(c.a, c.b); got != c.want {
			t.Errorf("samePort(%q, %q) = %v", c.a, c.b, got)
		}
	}
}
