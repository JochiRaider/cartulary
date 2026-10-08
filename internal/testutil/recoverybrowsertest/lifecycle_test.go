package recoverybrowsertest

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestLifecyclePreservesEveryCleanupFailure(t *testing.T) {
	var output bytes.Buffer
	lifecycle := NewLifecycle("attempt", &output)
	original := errors.New("original body failure")
	result := original
	retired := []string{}
	func() {
		defer lifecycle.cleanup("first", time.Second, func() error { retired = append(retired, "first"); return errors.New("first release failed") }, &result)
		defer lifecycle.cleanup("second", time.Second, func() error { retired = append(retired, "second"); return errors.New("second release failed") }, &result)
	}()
	if !errors.Is(result, original) || !strings.Contains(result.Error(), "first: first release failed") || !strings.Contains(result.Error(), "second: second release failed") {
		t.Fatal(result)
	}
	if strings.Join(retired, ",") != "second,first" {
		t.Fatal(retired)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 4 {
		t.Fatal(output.String())
	}
	for i, line := range lines {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if event["attempt_id"] != "attempt" || event["deadline_ms"] != float64(1000) {
			t.Fatal(event)
		}
		if i%2 == 1 && event["outcome"] != "failed" {
			t.Fatal(event)
		}
	}
}
