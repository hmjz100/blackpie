package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/hmjz100/blackpie/internal/catalog"
	"github.com/hmjz100/blackpie/internal/provider"
)

const openRouterRefusal = "This endpoint's maximum context length is 1048576 tokens. However, you requested about 1050086 tokens (101187 of text input, 5181 of tool input, 943718 in the output). Please reduce the length of either one, or use the context-compression plugin to compress your prompt automatically."

func TestReplyLeftOpenRouter(t *testing.T) {
	if got := replyLeft([]byte(openRouterRefusal)); got != 942208 {
		t.Fatalf("got %d, want 942208", got)
	}
}

func TestOpenRouterReplyFitsTheWindow(t *testing.T) {
	fresh(t)
	var mu sync.Mutex
	var asked []int64
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}

		n := gjson.GetBytes(b, "max_tokens").Int()
		mu.Lock()
		asked = append(asked, n)
		mu.Unlock()
		if n+106368 > 1048576 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			fmt.Fprintf(w, `{"error":{"message":%q,"code":400}}`, openRouterRefusal)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, sse(
			`data: {"id":"test","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"ok"}}]}`,
			`data: {"id":"test","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		))
	}))
	t.Cleanup(up.Close)
	if err := provider.Save(provider.Provider{ID: "orx", Name: "OR", Key: "k", Chat: up.URL + "/v1", Models: []string{"deepseek/deepseek-v4-pro"}}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.SaveLive("orx", up.URL+"/v1", []catalog.Model{{ID: "deepseek/deepseek-v4-pro", Context: 1048576, Output: 943718}}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(catalog.Reset)
	code, body := post(t, "/v1/responses", `{"model":"orx/deepseek/deepseek-v4-pro","stream":true,"input":"hi"}`)
	if code != 200 || !strings.Contains(body, "ok") {
		t.Fatalf("HTTP %d: %s", code, body)
	}
	mu.Lock()
	defer mu.Unlock()
	if fmt.Sprint(asked) != "[943718 942208]" {
		t.Fatalf("max_tokens requests = %v", asked)
	}
}

func TestOpenRouterOverflowFallsBackInGroups(t *testing.T) {
	for _, routing := range []string{provider.Ordered, provider.SmallFirst} {
		t.Run(routing, func(t *testing.T) {
			fresh(t)
			refusal := fmt.Sprintf(`{"error":{"message":%q,"code":400}}`, openRouterRefusal)
			first := &scripted{replies: []reply{{400, "", refusal}, {400, "", refusal}}}
			second := &scripted{replies: []reply{{200, "", chatOK}}}
			scriptedOn(t, "a", provider.Chat, first)
			scriptedOn(t, "b", provider.Chat, second)
			if err := provider.SaveGroup(provider.Group{Name: "G", Members: []string{"a/m", "b/m"}, Routing: routing}); err != nil {
				t.Fatal(err)
			}
			code, body := postAs(t, New(), "", `{"model":"group/g","max_tokens":943718,"messages":[{"role":"user","content":"hi"}]}`)
			if code != 200 || !strings.Contains(body, "hello") || first.n != 2 || second.n != 1 {
				t.Fatalf("HTTP %d: %s; first %d, second %d", code, body, first.n, second.n)
			}
		})
	}
}
