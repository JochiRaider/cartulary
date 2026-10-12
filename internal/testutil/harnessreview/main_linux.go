//go:build linux

package harnessreview

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func installedEnvironment(repository string) (map[string]string, error) {
	env := map[string]string{}
	for _, entry := range os.Environ() {
		k, v, ok := strings.Cut(entry, "=")
		if ok {
			env[k] = v
		}
	}
	data, err := os.ReadFile(filepath.Join(repository, "tools/workspace_layout.json"))
	if err != nil {
		return nil, err
	}
	var layout struct{ Directories []struct{ ID, Path string } }
	if json.Unmarshal(data, &layout) != nil {
		return nil, harnessschema.ErrInvalid
	}
	paths := map[string]string{}
	for _, d := range layout.Directories {
		paths[d.ID] = filepath.Join(repository, d.Path)
	}
	for key, value := range map[string]string{"NODE_BIN": filepath.Join(paths["node_runtime"], "bin/node"), "CARTULARY_TEST_SERVICES_BIN": filepath.Join(paths["toolbin"], "cartulary-test-services"), "CARTULARY_HARNESS_RUNTIME_BIN": filepath.Join(paths["toolbin"], "cartulary-harness"), "CARTULARY_SERVER_HARNESS_BIN": filepath.Join(paths["binaries"], "server-harness"), "CARTULARY_MIGRATE_BIN": filepath.Join(paths["binaries"], "migrate")} {
		if env[key] == "" {
			env[key] = value
		}
	}
	return env, nil
}

// Main reads a bounded private command, emitting only structural progress and
// normalized public outcomes. Status takes the reader-only path without writers.
func Main(args []string, input io.Reader, output io.Writer) int {
	if len(args) != 0 {
		return 2
	}
	repository, err := os.Getwd()
	if err != nil {
		return 2
	}
	data, err := io.ReadAll(io.LimitReader(input, harnesssession.MaximumFrame+1))
	if err != nil || len(data) > harnesssession.MaximumFrame {
		return 2
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate(commandSchema, value) != nil {
		_ = json.NewEncoder(output).Encode(map[string]any{"kind": "terminal", "value": errorResult("ui-review", failure("invalid_request"))})
		return 2
	}
	var q commandRequest
	if json.Unmarshal(data, &q) != nil {
		return 2
	}
	environment, err := installedEnvironment(repository)
	if err != nil {
		_ = json.NewEncoder(output).Encode(map[string]any{"kind": "terminal", "value": errorResult(q.Command, failure("tool_configuration"))})
		return 2
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	var mu sync.Mutex
	cause := error(failure("session_lost"))
	interrupted := func() error { mu.Lock(); defer mu.Unlock(); return cause }
	parent, parentErr := harnessadmission.Identity(os.Getppid())
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-signals:
				e := failure("interrupted")
				if s == syscall.SIGINT {
					e.Exit = 130
				} else {
					e.Exit = 143
				}
				mu.Lock()
				cause = e
				mu.Unlock()
				cancel()
				return
			case <-ticker.C:
				alive, err := harnessadmission.Alive(parent)
				if parentErr != nil || err != nil || !alive {
					cancel()
					return
				}
			}
		}
	}()
	var response Result
	if q.Command == "ui-review" || q.Command == "browser-design-review" || q.Command == "browser-design-review-smoke" {
		response, err = start(ctx, repository, q, environment, output, interrupted)
	} else {
		response, err = control(ctx, repository, q, environment)
	}
	if err != nil {
		command := q.Command
		if strings.HasPrefix(command, "browser-design-review") {
			command = "ui-review"
		}
		response = errorResult(command, err)
	}
	if _, err = jsonDocument(response.Schema, response); err != nil {
		response = errorResult(q.Command, failure("invalid_artifact"))
	}
	_ = json.NewEncoder(output).Encode(map[string]any{"kind": "terminal", "value": response})
	return response.Exit
}
