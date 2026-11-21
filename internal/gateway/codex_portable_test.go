package gateway

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hmjz100/blackpie/internal/codexsummary"
)

// A rollout written portable (sessions.PortCodexRollouts) holds, in place
// of blackpie's compaction item, the message the gateway sends for it: what
// reaches the model through blackpie is the same either way, for a blackpie
// model and for Codex's own, natively compacted or not.
func TestPortedCompactionReadsTheSame(t *testing.T) {
	enc := codexsummary.Encrypted("Progress: done.\nNext: tests.")
	msg, _ := codexsummary.Message(enc)
	body := func(item any) []byte {
		b, _ := json.Marshal(map[string]any{"model": "gpt-5.4", "input": []any{
			userMessage("fix the build"), item, userMessage("go on")}})
		return b
	}
	before := body(map[string]any{"type": "compaction", "id": "cmp_blackpie_1", "encrypted_content": enc})
	after := body(msg)
	for _, c := range []struct{ blackpieModel, native bool }{{true, false}, {true, true}, {false, false}} {
		a, _ := codexInputAs(before, c.blackpieModel, c.native)
		b, _ := codexInputAs(after, c.blackpieModel, c.native)
		var x, y any
		json.Unmarshal(a, &x)
		json.Unmarshal(b, &y)
		if !reflect.DeepEqual(x, y) {
			t.Errorf("%+v: blackpie's item sent as\n%s\nthe ported message as\n%s", c, a, b)
		}
	}
}
