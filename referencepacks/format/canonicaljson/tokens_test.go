package canonicaljson

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestBoundedTokenViews(t *testing.T) {
	for _, tc := range []struct {
		raw                 string
		canonical, integers bool
	}{
		{`{"a":[0,true,null,"é"],"b":{"x":-9007199254740991}}`, true, true},
		{`{"a":1,"a":2}`, false, true},
		{`{"b":1,"a":2}`, false, true},
		{`{"a":"\u0061"}`, false, true},
		{`{"a":"x\ny"}`, true, true},
		{`{"a":1.5}`, true, false},
		{`{"a":9007199254740992}`, true, false},
		{`{"a":-0}`, false, false},
		{`{"a":1e0}`, false, false},
		{`{"😀":1,"\ue000":2}`, false, true},
		{"{\"😀\":1,\"\ue000\":2}", true, true},
	} {
		doc, err := ScanTokens(context.Background(), []byte(tc.raw), 64)
		if err != nil || doc.Canonical != tc.canonical || doc.SafeIntegers != tc.integers {
			t.Fatalf("%q: %+v %v", tc.raw, doc, err)
		}
		count := 0
		if err := doc.Value.Children(context.Background(), func(_ string, _ int, v TokenValue) error {
			count++
			if len(v.Raw()) == 0 {
				t.Fatal("empty child")
			}
			return nil
		}); err != nil || count == 0 {
			t.Fatal(count, err)
		}
	}
	for _, raw := range []string{`{"a":"\ud800"}`, `{"a":[1,]}`, "{\"a\":\"\xff\"}"} {
		if _, err := ScanTokens(context.Background(), []byte(raw), 64); err == nil {
			t.Fatal("invalid token input admitted")
		}
	}
	for _, depth := range []int{64, 65, 10001} {
		_, err := ScanTokens(context.Background(), []byte(strings.Repeat("[", depth)+"0"+strings.Repeat("]", depth)), 64)
		if depth == 64 && err != nil || depth > 64 && !errors.Is(err, ErrDepth) {
			t.Fatal(depth, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ScanTokens(ctx, []byte(`{}`), 64); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
