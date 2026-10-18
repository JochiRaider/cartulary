package recoverybrowsertest

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

func TestLifecyclePreservesEveryCleanupFailure(t *testing.T) {
	t.Run("drop_wait_diagnostics_exclude_private_values", func(t *testing.T) {
		if got := postgrescleanup.DropWaitDescription("IPC", "ProcSignalBarrier", 0); got != "type=IPC event=ProcSignalBarrier blockers=0" {
			t.Fatal(got)
		}
		for _, input := range [][2]string{{"private_type", "ProcSignalBarrier"}, {"Extension", "PrivateCredentialAlpha123"}, {"Lock", "postgres://private:secret@host"}} {
			got := postgrescleanup.DropWaitDescription(input[0], input[1], 1)
			if strings.Contains(got, "private") || strings.Contains(got, "PrivateCredential") || strings.Contains(got, "secret") {
				t.Fatalf("private wait leaked: %s", got)
			}
		}
		if got := postgrescleanup.DropWaitDescription("IO", "DataFileSync", 0); got != "type=IO event=DataFileSync blockers=0" {
			t.Fatal(got)
		}
		if got := postgrescleanup.CheckpointProgressDescription(-1, 0, 0); got != "progress_unavailable" {
			t.Fatal(got)
		}
		if got := postgrescleanup.CheckpointProgressDescription(7, 6, 200); got != "requested=7 done=6 buffers=200" {
			t.Fatal(got)
		}
	})

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
