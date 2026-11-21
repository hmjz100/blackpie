package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hmjz100/blackpie/internal/access"
	"github.com/hmjz100/blackpie/internal/provider"
)

// codexOwnOffBlackpieSetup: Codex signed in to ChatGPT, its models cache
// listing gpt-6.1-sol, and no ChatGPT account in blackpie serving it: the
// sign-in removed from blackpie's Providers (hidden), or with discoveryOff
// subscription accounts turned off. The ChatGPT backend stands in and
// notes what it was sent.
func codexOwnOffBlackpieSetup(t *testing.T, discoveryOff bool) (got *http.Header, models *[]string) {
	t.Helper()
	codexSignedIn(t)
	if discoveryOff {
		t.Setenv("BLACKPIE_NO_ACCOUNTS", "1")
	} else if err := provider.Delete("codex"); err != nil {
		t.Fatal(err)
	}
	provider.ForgetAccounts()
	home := os.Getenv("HOME")
	os.WriteFile(filepath.Join(home, ".codex", "models_cache.json"),
		[]byte(`{"etag":"e1","models":[{"slug":"gpt-6.1-sol","display_name":"gpt-6.1-sol","priority":1}]}`), 0o600)
	got, models = &http.Header{}, &[]string{}
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*got = r.Header.Clone()
		*models = append(*models, modelOf(b))
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, sse(
			`data: {"type":"response.created","response":{"id":"r1","model":"gpt-6.1-sol"}}`,
			`data: {"type":"response.output_text.delta","delta":"pong"}`,
			`data: {"type":"response.completed","response":{"id":"r1","usage":{"input_tokens":7,"output_tokens":1}}}`))
	}))
	t.Cleanup(up.Close)
	was := provider.CodexBase
	provider.CodexBase = up.URL + "/backend-api/codex"
	t.Cleanup(func() { provider.CodexBase = was })
	return got, models
}

func codexOnBlackpieRequest(from string) *http.Request {
	req := httptest.NewRequest("POST", "/v1/responses",
		strings.NewReader(`{"model":"gpt-6.1-sol","stream":true,"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"ping"}]}]}`))
	req.RemoteAddr = from
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "codex_cli_rs/0.164.0 (Mac OS 26.2.0; arm64) Terminal")
	req.Header.Set("Originator", "codex_cli_rs")
	req.Header.Set("Authorization", "Bearer "+Token)
	req.Header.Set("X-Openai-Actor-Authorization", "blackpie")
	return req
}

// A thread Codex started while blackpie was its provider keeps sending to
// blackpie's /v1 after Codex is disconnected or set back to ChatGPT; one of
// Codex's own models picked in it, which no account in blackpie serves,
// goes to the ChatGPT backend on Codex's own sign-in, blackpie's key and
// header left behind, where it was answered 404 "blackpie knows no model
// \"gpt-6.1-sol\"" (iuengei on Discord).
func TestCodexOwnModelOnBlackpieThreadGoesOnSignIn(t *testing.T) {
	got, models := codexOwnOffBlackpieSetup(t, false)
	rec := httptest.NewRecorder()
	New().Handler().ServeHTTP(rec, codexOnBlackpieRequest("127.0.0.1:51000"))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "pong") {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
	if len(*models) != 1 || (*models)[0] != "gpt-6.1-sol" {
		t.Fatalf("backend asked for %v", *models)
	}
	auth := got.Get("Authorization")
	if auth == "" || strings.Contains(auth, Token) {
		t.Fatalf("Authorization %q: want Codex's sign-in, not blackpie's key", auth)
	}
	if got.Get("Chatgpt-Account-Id") != "acct-1" {
		t.Fatalf("chatgpt-account-id %q, want acct-1", got.Get("Chatgpt-Account-Id"))
	}
	if v := got.Get("X-Openai-Actor-Authorization"); v != "" {
		t.Fatalf("blackpie's provider header went to OpenAI: %q", v)
	}
}

// Not for a request from another computer, nor a gateway key held to some
// models, nor a model that isn't OpenAI's: those are answered as before,
// nothing sent on Codex's sign-in.
func TestCodexOwnModelOnBlackpieThreadStaysHere(t *testing.T) {
	t.Run("another computer", func(t *testing.T) {
		_, models := codexOwnOffBlackpieSetup(t, false)
		rec := httptest.NewRecorder()
		New().Handler().ServeHTTP(rec, codexOnBlackpieRequest("192.0.2.7:51000"))
		if len(*models) != 0 {
			t.Fatalf("a request from another computer went on Codex's sign-in: %d %s", rec.Code, rec.Body.String())
		}
	})
	t.Run("held key", func(t *testing.T) {
		_, models := codexOwnOffBlackpieSetup(t, false)
		req := codexOnBlackpieRequest("127.0.0.1:51000")
		req = req.WithContext(access.WithIdentity(req.Context(), access.Identity{KeyID: "k1", Models: []string{"fake/*"}}))
		rec := httptest.NewRecorder()
		New().handle(provider.Responses)(rec, req)
		if len(*models) != 0 {
			t.Fatalf("a held key went on Codex's sign-in: %d %s", rec.Code, rec.Body.String())
		}
	})
	t.Run("not OpenAI's", func(t *testing.T) {
		_, models := codexOwnOffBlackpieSetup(t, false)
		req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(`{"model":"no-such-model-xyz","input":"hi"}`))
		req.RemoteAddr = "127.0.0.1:51000"
		req.Header = codexOnBlackpieRequest("127.0.0.1:51000").Header
		rec := httptest.NewRecorder()
		New().Handler().ServeHTTP(rec, req)
		if len(*models) != 0 || rec.Code != 404 {
			t.Fatalf("an unknown model: %d %s, sent %v", rec.Code, rec.Body.String(), *models)
		}
	})
}

// With subscription accounts off, blackpie reads no sign-in, Codex's
// neither: the request is answered why it can't be served and what to do,
// not with the bare "blackpie knows no model".
func TestCodexOwnModelOnBlackpieThreadSaysWhy(t *testing.T) {
	_, models := codexOwnOffBlackpieSetup(t, true)
	rec := httptest.NewRecorder()
	New().Handler().ServeHTTP(rec, codexOnBlackpieRequest("127.0.0.1:51000"))
	body := rec.Body.String()
	if rec.Code != 404 || len(*models) != 0 {
		t.Fatalf("got %d %s, sent %v", rec.Code, body, *models)
	}
	for _, want := range []string{"one of Codex's own models", "started while blackpie was Codex's model provider", "subscription accounts are off", "Start a new thread"} {
		if !strings.Contains(body, want) {
			t.Errorf("answer lacks %q: %s", want, body)
		}
	}
}
