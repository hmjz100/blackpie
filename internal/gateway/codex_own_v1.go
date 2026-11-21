package gateway

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hmjz100/blackpie/internal/catalog"
	"github.com/hmjz100/blackpie/internal/codexcat"
	"github.com/hmjz100/blackpie/internal/provider"
)

// codexSignInToken is Codex's ChatGPT sign-in as Codex keeps it, which
// tests set.
var codexSignInToken = provider.CodexSignInAsKept

// codexOwnOffBlackpie is the request a Codex on this computer sent to /v1
// for one of its own models that blackpie serves nothing for, put on the
// ChatGPT backend's route (CodexPath) and signed with Codex's own ChatGPT
// sign-in, as Codex's built-in provider would have sent it. A thread
// keeps the provider it was started on: one started while blackpie was
// Codex's provider (its API, or the takeover of a used-up account) goes
// on sending to /v1 after Codex is disconnected or set back to ChatGPT,
// and an official model picked in it (gpt-6.1-sol) met "blackpie knows no
// model" whenever no ChatGPT account in blackpie serves it (iuengei on
// Discord). nil when that isn't the case: a model blackpie serves, switched
// off, disabled or a group; a model not OpenAI's; a request from another
// computer or held by its gateway key. Codex's sign-in not to be had
// (subscription accounts off in blackpie, signed out, expired), why is what
// the request is answered instead of the bare 404: what sends it here and
// what to do.
func codexOwnOffBlackpie(r *http.Request, body []byte, model string) (to *http.Request, why string) {
	if model == "" || strings.Contains(model, "/") || !local(r) {
		return nil, ""
	}
	if !fromCodex(r.Header) && agentOf(r) != "codex" {
		return nil, ""
	}
	if _, _, ok := provider.Resolve(model); ok {
		return nil, ""
	}
	if provider.IsGroupRef(model) {
		return nil, ""
	}
	if _, ok := provider.GroupFor(model); ok {
		return nil, ""
	}
	if _, ok := provider.AutoStandIn(model); ok {
		return nil, ""
	}
	if _, off := idOff(model); off {
		return nil, ""
	}
	if _, off := provider.SwitchedOff(model); off {
		return nil, ""
	}
	if _, keyHeld := keyHolds(r); keyHeld {
		return nil, ""
	}
	if _, accHeld := accountHolds(r); accHeld {
		return nil, ""
	}
	if !codexOwnModel(model) {
		return nil, ""
	}
	tok, acct, err := codexSignInToken()
	if err != nil || tok == "" {
		if err == nil {
			err = errors.New("Codex is signed out")
		}
		return nil, fmt.Sprintf("%q is one of Codex's own models, and this Codex thread was started while blackpie was Codex's model provider, so it is still sent to blackpie whatever Codex is set to now. "+
			"No ChatGPT account in blackpie serves it, and blackpie can't send it on Codex's own ChatGPT sign-in: %v. "+
			"Start a new thread in Codex for its own models, or add Codex's ChatGPT account in blackpie (Providers)", model, err)
	}
	r2 := r.Clone(r.Context())
	r2.URL.Path = CodexPath + "/responses"
	r2.URL.RawPath = ""
	r2.RequestURI = ""
	r2.Body = io.NopCloser(bytes.NewReader(body))
	r2.ContentLength = int64(len(body))
	r2.Header.Del("Content-Encoding")
	r2.Header.Del("Content-Length")
	// blackpie's key and the header that names blackpie Codex's provider are
	// blackpie's own, never sent on to OpenAI
	r2.Header.Del("X-Api-Key")
	r2.Header.Del("X-Openai-Actor-Authorization")
	r2.Header.Set("Authorization", "Bearer "+tok)
	if acct != "" {
		r2.Header.Set("Chatgpt-Account-Id", acct)
	}
	return r2, ""
}

// codexOwnModel: model is one of Codex's own, as the ChatGPT backend last
// listed it to Codex (models_cache.json, blackpie's entries left out) or as
// the catalog has it for OpenAI.
func codexOwnModel(model string) bool {
	if _, ok := codexcat.CacheEntries()[model]; ok {
		return true
	}
	for _, m := range catalog.Provider("openai") {
		if m.ID == model {
			return true
		}
	}
	return false
}
