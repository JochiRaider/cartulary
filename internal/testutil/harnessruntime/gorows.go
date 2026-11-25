package harnessruntime

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessrows"
)

// goRows is a Go unit's settled rows (TH-HARNESS-REQ-872): the runtime runs
// the planned `go test` command itself, observes its event stream while it
// runs, and publishes each row's result with the unit's evidence.
type goRows struct {
	Outcomes          []harnessrows.Outcome
	Started, Finished time.Time
	// Durations is retained only under basic execution diagnostics.
	Durations *harnessrows.Durations
}

// masterKeyFixture is the explicit test master key. Production key loading
// has no default, so the repository's own Go tests receive this one.
const masterKeyFixture = "internal/testutil/fixtures/auth/master-key.base64"

// bindGo completes a Go unit's environment and returns the go executable:
// the machine Go caches, this invocation's private temporary directory, no
// inherited GOROOT, the test master key for the repository's own module, and
// the go binary the preparation owner selected.
func bindGo(root string, environment map[string]string, args []string) (string, error) {
	for target, source := range map[string]string{"GOCACHE": "GO_CACHE_DIR", "GOMODCACHE": "GO_MOD_CACHE_DIR", "GOTMPDIR": "GO_TMP_DIR"} {
		value := environment[source]
		if value == "" || !filepath.IsAbs(value) || filepath.Clean(value) != value {
			return "", fmt.Errorf("%s must be an absolute clean path", source)
		}
		environment[target] = value
	}
	environment["GOROOT"] = ""
	if len(args) == 0 || args[0] != "-C" {
		key, err := os.ReadFile(filepath.Join(root, masterKeyFixture))
		if err != nil {
			return "", err
		}
		environment["CARTULARY_AUTH_MASTER_KEY"] = strings.TrimSpace(string(key))
	}
	executable := environment["GO"]
	if executable == "" {
		return "go", nil
	}
	if !filepath.IsAbs(executable) {
		return "", errors.New("GO must be an absolute path")
	}
	return executable, nil
}

// settleGoRows turns a finished Go unit's stream into its rows and the
// unit's outcome: the primary failure among its rows, with the bounded
// failure detail and one line per failed row for the unit's log. A stream
// that broke its line contract settles no row.
func settleGoRows(u unit, stream *harnessrows.Stream, started, finished time.Time, durations bool) (*goRows, *failure, []byte) {
	result, err := stream.Finish(u.Command.Args[len(u.Command.Args)-1])
	if err != nil {
		return nil, failed("artifact", "artifact_error"), []byte("[FAIL] runner=go failure_class=artifact failure_reason=artifact_error: the go test event stream broke its line contract\n")
	}
	settled := &goRows{Outcomes: result.Outcomes, Started: started, Finished: finished}
	if durations {
		settled.Durations = &result.Durations
	}
	var failures []*failure
	var report bytes.Buffer
	for _, outcome := range result.Outcomes {
		if outcome.TerminalState == "passed" {
			continue
		}
		failures = append(failures, failed(outcome.FailureClass, outcome.FailureReason))
		fmt.Fprintf(&report, "[FAIL] row=%s runner=go failure_class=%s failure_reason=%s row_reason=%s\n", outcome.RowID, outcome.FailureClass, outcome.FailureReason, outcome.FailureReason)
	}
	if len(failures) == 0 {
		return settled, nil, nil
	}
	return settled, primaryFailure(failures), append([]byte(result.FailureDetails), report.Bytes()...)
}
