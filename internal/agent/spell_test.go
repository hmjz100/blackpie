package agent

import (
	"os"
	"testing"

	"github.com/hmjz100/blackpie/internal/provider"
	"github.com/hmjz100/blackpie/internal/settings"
)

func TestSpell(t *testing.T) {
	opts := func(prefix string) func(map[string]string) []Option {
		return func(map[string]string) []Option {
			return []Option{
				{Value: "gpt-6"},
				{Value: prefix + "copilot/gpt-6-sol", Ref: "copilot/gpt-6-sol"},
			}
		}
	}
	get := func() string { return "" }
	bare := &Agent{Fields: []Field{{Key: "model", Get: get, Options: opts("")}}}
	prefixed := &Agent{Fields: []Field{{Key: "model", Get: get, Options: opts("blackpie/")}}}
	for _, c := range []struct {
		a       *Agent
		in, out string
		err     bool
	}{
		{bare, "copilot/gpt-6-sol", "copilot/gpt-6-sol", false},
		{bare, "blackpie/copilot/gpt-6-sol", "copilot/gpt-6-sol", false},
		{prefixed, "copilot/gpt-6-sol", "blackpie/copilot/gpt-6-sol", false},
		{prefixed, "blackpie/copilot/gpt-6-sol", "blackpie/copilot/gpt-6-sol", false},
		{bare, "gpt-6", "gpt-6", false},
		{bare, "some-unlisted-model", "some-unlisted-model", false},
		{bare, "", "", false},
		{bare, "blackpie/copilot/nope", "", true},
		{prefixed, "blackpie/copilot/nope", "", true},
	} {
		got, err := c.a.Spell("model", c.in)
		if got != c.out || (err != nil) != c.err {
			t.Errorf("%q: got %q, %v; want %q, err %v", c.in, got, err, c.out, c.err)
		}
	}
}

// #2376: a model of a remote blackpie that the agent's picker doesn't offer,
// kept from it by blackpie visible or not on the remote's list yet, was
// taken as the agent's own: Pi read remote/group/smart as its provider
// "remote" and model "group/smart", and was taken out of blackpie. It is
// refused, and one the picker offers, or the agent's own, is spelled as
// before.
func TestSpellRefusesBlackpieModelNotOffered(t *testing.T) {
	syncHome(t)
	if err := provider.Save(provider.Provider{ID: "remote", Name: "Remote blackpie", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"group/smart", "codex/gpt-5.5"}}); err != nil {
		t.Fatal(err)
	}
	if err := settings.Save(settings.Settings{Visible: map[string][]string{"pi": {"relay"}, "codex": {"relay"}}}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { settings.Save(settings.Settings{}) })
	home := os.Getenv("HOME")
	for _, a := range []*Agent{pi(home), codex(home)} {
		for _, v := range []string{"remote/group/smart", "remote/codex/gpt-5.5", "remote/group/new-pool", "group/nope"} {
			if got, err := a.Spell(a.Fields[0].Key, v); err == nil {
				t.Errorf("%s: %s spelled %q, want refused", a.ID, v, got)
			}
		}
		want := "relay/glm-4.6"
		if a.ID == "pi" {
			want = "blackpie/" + want
		}
		if got, err := a.Spell(a.Fields[0].Key, "relay/glm-4.6"); err != nil || got != want {
			t.Errorf("%s: relay/glm-4.6 spelled %q, %v; want %q", a.ID, got, err, want)
		}
	}
	// a provider of the agent's own that blackpie hasn't is the agent's
	if got, err := pi(home).Spell("model", "my-own/some-model"); err != nil || got != "my-own/some-model" {
		t.Errorf("pi's own: %q, %v", got, err)
	}
	// shown it every model, the remote's are offered and spelled as blackpie's
	if err := settings.Save(settings.Settings{}); err != nil {
		t.Fatal(err)
	}
	if got, err := pi(home).Spell("model", "remote/group/smart"); err != nil || got != "blackpie/remote/group/smart" {
		t.Errorf("pi, shown every model: %q, %v", got, err)
	}
}
