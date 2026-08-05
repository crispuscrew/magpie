package bench

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTokensRoundTrip(t *testing.T) {
	out, _ := json.Marshal(Row{Task: "x", Tokens: Tokens{In: 10, Out: 41, CacheRead: 17536}})
	if !strings.Contains(string(out), `"tokens":{"in":10,"out":41,"cache_read":17536}`) {
		t.Fatalf("tokens not serialised: %s", out)
	}
	if strings.Contains(string(mustJSON(Row{Task: "x"})), "tokens") {
		t.Error("a run with no usage must not emit an empty tokens object")
	}
	var back Row
	if err := json.Unmarshal(out, &back); err != nil || back.Tokens.Total() != 17587 {
		t.Errorf("round trip: %v total=%d", err, back.Tokens.Total())
	}
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }
