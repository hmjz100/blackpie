package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hmjz100/blackpie/internal/gwtool"
	"github.com/hmjz100/blackpie/internal/provider"
)

// The tool point's web search with plugins/search on and off, for a user
// with a search API and a Kimi Code plan that searches by itself. On,
// blackpie offers its web_search to a model that can't search and answers
// it with the API or the plan's own search service. Off, it offers none,
// the request goes to the vendor without it, and a search asked of blackpie
// anyway (the search MCP server, Codex's in ChatGPT's place) says the
// plugin is off without asking the API or the plan. A vendor that
// searches by itself is sent its own web_search either way.
func TestSearchOnAndOff(t *testing.T) {
	fresh(t)
	relay, seen := relayOfEvery(t)
	kimi := newKimiPlan(t)
	plan := provider.Provider{ID: "kimi", Name: "Kimi Code", Preset: "kimi-code-cn", Key: "sk-kimi",
		Chat: kimi.URL + "/coding/v1", Models: []string{"kimi-for-coding"}}
	if err := provider.Save(plan); err != nil {
		t.Fatal(err)
	}
	f := newSearchAPIs(t)
	if err := provider.SetSearchAPI(f.api("tavily", "tvly-k")); err != nil {
		t.Fatal(err)
	}
	s := New()
	ask := func(native bool) string {
		t.Helper()
		p := provider.Provider{ID: "relay", Name: "Relay", Key: "k", Models: []string{"custom-model"}, Searches: native}
		if native {
			p.Anthropic = relay.URL
		} else {
			p.Chat = relay.URL + "/v1"
		}
		if err := provider.Save(p); err != nil {
			t.Fatal(err)
		}
		_, before := seen()
		body := `{"model":"relay/custom-model","max_tokens":100,"messages":[{"role":"user","content":"news?"}],"tools":[{"name":"Read","input_schema":{"type":"object"}},{"type":"web_search_20250305","name":"web_search"}]}`
		if code, reply := post(t, "/v1/messages", body); code != 200 {
			t.Fatalf("native %v: %d %s", native, code, reply)
		}
		paths, bodies := seen()
		if len(bodies) != len(before)+1 {
			t.Fatalf("native %v: upstream calls %v", native, paths[len(before):])
		}
		sent := bodies[len(bodies)-1]
		if !strings.Contains(sent, "Read") {
			t.Fatalf("native %v: the agent's own tool was dropped: %s", native, sent)
		}
		return sent
	}
	mcp := func() (string, error) {
		return s.searchMCPRun(httptest.NewRequest("POST", "/mcp/search", nil), "latest go")
	}

	// on
	if !CanSearch() || !canSearchFor(provider.Provider{}) || !searchesOwn(plan) {
		t.Fatal("on: blackpie's search not offered")
	}
	if !strings.Contains(ask(false), "web_search") {
		t.Error("on: a model that can't search wasn't given blackpie's web_search")
	}
	if !strings.Contains(ask(true), "web_search_20250305") {
		t.Error("on: a vendor that searches by itself wasn't sent its own web_search")
	}
	if _, err := mcp(); err != nil {
		t.Errorf("on: MCP search: %v", err)
	}
	if _, _, err := s.webSearch(searchingOn(context.Background(), plan), "latest go"); err != nil {
		t.Errorf("on: the plan's own search: %v", err)
	}
	f.mu.Lock()
	asked := len(f.asked)
	f.mu.Unlock()
	kimi.mu.Lock()
	searched := len(kimi.searched)
	kimi.mu.Unlock()
	if asked == 0 || searched == 0 {
		t.Fatalf("on: the API was asked %d times, the plan %d", asked, searched)
	}

	// off
	withoutSearch(t)
	if CanSearch() || canSearchFor(provider.Provider{}) || canSearchFor(plan) || searchesOwn(plan) {
		t.Error("off: blackpie's search still offered")
	}
	if sent := ask(false); strings.Contains(sent, "web_search") {
		t.Errorf("off: blackpie's web_search offered: %s", sent)
	}
	if !strings.Contains(ask(true), "web_search_20250305") {
		t.Error("off: a vendor that searches by itself lost its own web_search")
	}
	if _, err := mcp(); err == nil || !strings.Contains(err.Error(), "Web search plugin") {
		t.Errorf("off: MCP search said %v, not that the plugin is off", err)
	}
	for _, ctx := range []context.Context{context.Background(), searchingOn(context.Background(), plan)} {
		if _, _, err := s.webSearch(ctx, "latest go"); !errors.Is(err, errSearchOff) {
			t.Errorf("off: search said %v", err)
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	kimi.mu.Lock()
	defer kimi.mu.Unlock()
	if len(f.asked) != asked || len(kimi.searched) != searched {
		t.Errorf("off: the API was asked %v, the plan %v", f.asked[asked:], kimi.searched[searched:])
	}
}

// The blackpie serving the gateway, and only while it serves, offers
// blackpie's tools only as their plugins serve them (gwtool.Gate): with
// plugins/search off, not its web search.
func TestServingGatesTools(t *testing.T) {
	fresh(t)
	t.Setenv("BLACKPIE_ADDR", "127.0.0.1:0") // never where a blackpie runs
	searchOn()
	defer func() { searchOn = gwtool.Serve(gwtool.WebSearch) }()
	if !gwtool.Served(gwtool.WebSearch) {
		t.Fatal("gated before serving")
	}
	s := New()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.ListenAndServe(ctx) }()
	defer func() {
		cancel()
		if done != nil {
			if err := <-done; err != nil {
				t.Error(err)
			}
		}
	}()
	deadline := time.Now().Add(30 * time.Second)
	for {
		s.lnMu.Lock()
		ln := s.ln
		s.lnMu.Unlock()
		if ln != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("gateway did not bind")
		}
		time.Sleep(10 * time.Millisecond)
	}
	s.lnMu.Lock()
	addr := s.ln.Addr().String()
	s.lnMu.Unlock()
	// Binding precedes installing the tool gate; a response proves Serve began.
	readyCtx, readyCancel := context.WithTimeout(ctx, 30*time.Second)
	defer readyCancel()
	req, err := http.NewRequestWithContext(readyCtx, http.MethodGet, "http://"+addr+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("gateway did not start serving: %v", err)
	}
	res.Body.Close()
	if gwtool.Served(gwtool.WebSearch) {
		t.Error("serving, with the plugin off: web search still served")
	}
	cancel()
	err = <-done
	done = nil
	if err != nil {
		t.Fatal(err)
	}
	if !gwtool.Served(gwtool.WebSearch) {
		t.Error("gate held after serving ended")
	}
}
