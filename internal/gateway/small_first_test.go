package gateway

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/hmjz100/blackpie/internal/provider"
)

func TestSmallContextFirstGroup(t *testing.T) {
	fresh(t)
	up := httptest.NewServer(http.HandlerFunc(streamed))
	t.Cleanup(up.Close)
	to, err := url.Parse(up.URL)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []provider.Provider{
		{ID: "big", Name: "Big", Key: "b", Chat: "https://big.example/v1", Models: []string{"m"}, Contexts: map[string]int{"m": 1000000}},
		{ID: "small", Name: "Small", Key: "s", Chat: "https://small.example/v1", Models: []string{"m"}, Contexts: map[string]int{"m": 4000}},
	} {
		if err := provider.Save(p); err != nil {
			t.Fatal(err)
		}
	}
	group := func(routing string) {
		t.Helper()
		if err := provider.SaveGroup(provider.Group{Name: "G", Members: []string{"big/m", "small/m"}, Routing: routing}); err != nil {
			t.Fatal(err)
		}
	}
	s := New()
	s.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r.URL.Scheme, r.URL.Host = to.Scheme, to.Host
		return http.DefaultTransport.RoundTrip(r)
	})}
	ask := func(text string) *Route {
		t.Helper()
		code, body := postAs(t, s, "", `{"model":"group/g","stream":true,"messages":[{"role":"user","content":"`+text+`"}]}`)
		if code != 200 {
			t.Fatalf("code %d: %s", code, body)
		}
		return s.trace.routes[len(s.trace.routes)-1]
	}
	group(provider.SmallFirst)
	short := ask("hi")
	if short.Order[0].Provider != "small" || short.Order[0].Context != 4000 {
		t.Fatalf("short request: %+v", short.Order)
	}
	long := ask(strings.Repeat("long conversation ", 3000))
	if long.Order[0].Provider != "big" {
		t.Fatalf("long request: %+v", long.Order)
	}
	group(provider.Ordered)
	if order := ask("hi").Order; order[0].Provider != "big" {
		t.Fatalf("ordered group changed: %+v", order)
	}
}
