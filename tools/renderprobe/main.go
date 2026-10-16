// Package main qualifies the installed renderer inside the exact application
// image. It accepts no commands, credentials, paths or product data. Package
// qualification supplies only the shared, synthetic serializer vectors.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
	"golang.org/x/sys/unix"
)

type observation struct {
	Operation              string `json:"operation"`
	SHA256                 string `json:"sha256"`
	Bytes                  int    `json:"bytes"`
	FreshAttempts          int    `json:"fresh_attempts"`
	NativeLoopbackObserved bool   `json:"native_loopback_observed"`
	CleanupVerified        bool   `json:"cleanup_verified"`
}

func qualify() (map[string]any, error) {
	if len(os.Args) != 1 {
		return nil, errors.New("unexpected_arguments")
	}
	status, err := os.ReadFile("/proc/self/status")
	if err != nil || os.Geteuid() == 0 || !bytes.Contains(status, []byte("CapEff:\t0000000000000000\n")) || !bytes.Contains(status, []byte("NoNewPrivs:\t1\n")) || !bytes.Contains(status, []byte("Seccomp:\t2\n")) {
		return nil, errors.New("outer_process_policy_invalid")
	}
	for _, denied := range []struct {
		call, argument uintptr
		want           unix.Errno
	}{
		{unix.SYS_UNSHARE, unix.CLONE_NEWNET, unix.EPERM},
		{unix.SYS_BPF, 0, unix.EPERM},
		{unix.SYS_CLONE3, 0, unix.ENOSYS},
		{unix.SYS_IO_URING_SETUP, 0, unix.ENOSYS},
	} {
		if _, _, errno := unix.Syscall(denied.call, denied.argument, 0, 0); errno != denied.want {
			return nil, errors.New("outer_syscall_policy_invalid")
		}
	}
	fd, socketErr := unix.Socket(unix.AF_ALG, unix.SOCK_SEQPACKET, 0)
	if socketErr == nil {
		_ = unix.Close(fd)
	}
	if !errors.Is(socketErr, unix.EPERM) {
		return nil, errors.New("outer_socket_policy_invalid")
	}
	installation, err := renderprocess.OpenInstallation("/opt/cartulary/report-renderer")
	if err != nil {
		return nil, errors.New("installation_admission_failed")
	}
	data, err := os.ReadFile("/probe-fixtures/slidev-source.v2.json")
	if err != nil {
		return nil, errors.New("fixture_unavailable")
	}
	var fixture struct {
		SourceLines []string `json:"source_lines"`
		Deck        struct {
			Slides []json.RawMessage `json:"slides"`
		} `json:"deck"`
	}
	if json.Unmarshal(data, &fixture) != nil || len(fixture.Deck.Slides) != 2 || len(fixture.SourceLines) == 0 {
		return nil, errors.New("fixture_invalid")
	}
	temporary, err := os.MkdirTemp("", "qualification-")
	if err != nil {
		return nil, errors.New("temporary_root_unavailable")
	}
	defer os.RemoveAll(temporary)
	observations := []observation{}
	for _, operation := range []string{"mermaid_svg", "slidev_pdf"} {
		source := "flowchart TD\n  n0001[\"First\"]\n  n0002[\"Second\"]\n  n0001 --> n0002\n"
		output, magic := "diagram.svg", []byte("<svg")
		request := map[string]any{"schema_id": "cartulary.reporting_render_stage_request.v5", "operation": operation, "admitted_at": "2026-04-05T12:00:00Z", "identity_sha256": strings.Repeat("a", 64)}
		if operation == "slidev_pdf" {
			source = strings.Join(fixture.SourceLines, "\n")
			output, magic = "deck.pdf", []byte("%PDF-")
			request["expected_pages"] = len(fixture.Deck.Slides)
		}
		request["source"] = source
		input, err := json.Marshal(request)
		if err != nil {
			return nil, errors.New("request_invalid")
		}
		var first []byte
		for attempt := 0; attempt < 2; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			result, err := installation.Runtime(temporary, 32123).Run(ctx, input, []renderprocess.Output{{Path: output, MaxBytes: 32 << 20}})
			cancel()
			if err != nil {
				var execution *renderprocess.ExecutionError
				if errors.As(err, &execution) {
					return nil, fmt.Errorf("%s:%s", operation, execution.Reason)
				}
				if errors.Is(err, context.DeadlineExceeded) {
					return nil, fmt.Errorf("%s:deadline_exceeded", operation)
				}
				return nil, fmt.Errorf("%s:native_execution_failed", operation)
			}
			if len(result.NetworkObservations) != 1 || result.NetworkObservations[0].Kind != "loopback_allowed" || !result.NetworkObservations[0].Allowed || result.NetworkObservations[0].Count < 2 {
				return nil, errors.New("native_observation_missing")
			}
			artifact := result.Files[output]
			if len(result.Files) != 1 || !bytes.Contains(artifact, magic) {
				return nil, errors.New("artifact_invalid")
			}
			if attempt == 0 {
				first = artifact
			} else if !bytes.Equal(first, artifact) {
				return nil, errors.New("fresh_attempts_differ")
			}
			entries, err := os.ReadDir(temporary)
			if err != nil || len(entries) != 0 {
				return nil, errors.New("attempt_cleanup_failed")
			}
		}
		digest := sha256.Sum256(first)
		observations = append(observations, observation{Operation: operation, SHA256: hex.EncodeToString(digest[:]), Bytes: len(first), FreshAttempts: 2, NativeLoopbackObserved: true, CleanupVerified: true})
	}
	if installation.Verify() != nil {
		return nil, errors.New("installation_changed")
	}
	return map[string]any{"schema_id": "cartulary.packaged_renderer_qualification.v1", "status": "pass", "installation_sha256": installation.Digest(), "operations": observations, "outer_policy_verified": true}, nil
}

func main() {
	result, err := qualify()
	if err != nil {
		// Only fixed classifications are returned; native stderr is never emitted.
		_ = json.NewEncoder(os.Stdout).Encode(map[string]string{"schema_id": "cartulary.packaged_renderer_qualification.v1", "status": "fail", "reason_code": err.Error()})
		os.Exit(1)
	}
	if json.NewEncoder(os.Stdout).Encode(result) != nil {
		os.Exit(1)
	}
}
