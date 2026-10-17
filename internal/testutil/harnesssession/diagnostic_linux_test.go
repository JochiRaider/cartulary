//go:build linux

package harnesssession

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

func testDiagnosticAuthority(t *testing.T) {
	t.Run("proc_read_failure_requires_renewed_settlement_proof", func(t *testing.T) {
		cmd := exec.Command("/bin/sleep", "60")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
		proof, err := readProcess(cmd.Process.Pid)
		if err != nil {
			t.Fatal(err)
		}
		if diagnosticReadSettled(proof) {
			t.Fatal("live process treated as settled")
		}
		if err := cmd.Process.Kill(); err != nil {
			t.Fatal(err)
		}
		_ = cmd.Wait()
		if !diagnosticReadSettled(proof) {
			t.Fatal("verified absence did not resolve proc read race")
		}
		self, err := readProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		self.Start = "0"
		if diagnosticReadSettled(self) {
			t.Fatal("reused live identity treated as settled")
		}
	})
	t.Run("scope_requires_exact_journal_and_current_boot", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		scope, err := r.CreateDiagnosticScope([]string{})
		if err != nil {
			t.Fatal(err)
		}
		defer r.RecordResource("diagnostic_scope", scope, "released")
		raw, _ := json.Marshal(scope)
		if _, err := r.InspectDiagnosticScope(raw); err != nil {
			t.Fatal(err)
		}
		forged := scope
		forged.Token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		foreign, _ := json.Marshal(forged)
		if _, err := r.InspectDiagnosticScope(foreign); !errors.Is(err, harnesslocal.ErrUnsafe) {
			t.Fatal("unrecorded scope admitted", err)
		}
		forged = scope
		forged.Boot = "previous-boot"
		if err := r.RecordDependentResource("diagnostic_scope", forged, "pending", []string{}); err != nil {
			t.Fatal(err)
		}
		defer r.RecordResource("diagnostic_scope", forged, "released")
		foreign, _ = json.Marshal(forged)
		if _, err := r.InspectDiagnosticScope(foreign); !errors.Is(err, ErrUnresolved) {
			t.Fatal("conflicting boot treated as absence", err)
		}
	})
	t.Run("inaccessible_relevant_proof_is_unresolved_without_signaling", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		scope, err := r.CreateDiagnosticScope([]string{})
		if err != nil {
			t.Fatal(err)
		}
		defer r.RecordResource("diagnostic_scope", scope, "released")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSessionRuntimeProof$")
		cmd.Env = append(os.Environ(), "CARTULARY_SCOPE_TEST_PROTECTED=1", "CARTULARY_DIAGNOSTIC_TOKEN="+scope.Token)
		output, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
		line, err := bufio.NewReader(output).ReadString('\n')
		if err != nil || line != "ready\n" {
			t.Fatal("protected fixture failed", line, err)
		}
		if _, marked, _, err := diagnosticMember(scope, cmd.Process.Pid); marked || !errors.Is(err, os.ErrPermission) {
			t.Fatal("protected fixture did not exercise inaccessible private proof", marked, err)
		}
		raw, _ := json.Marshal(scope)
		if _, err := r.InspectDiagnosticScope(raw); !errors.Is(err, ErrUnresolved) {
			t.Fatal("inaccessible proof ignored", err)
		}
		settlement, cancelSettlement := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancelSettlement()
		if err := r.SettleDiagnosticScope(settlement, raw); !errors.Is(err, ErrUnresolved) || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("inaccessible scope did not retain uncertainty until its deadline", err)
		}
		current, err := readProcess(cmd.Process.Pid)
		if err != nil || current.state == "Z" || current.state == "X" {
			t.Fatal("uncertain process signaled", err)
		}
	})
	t.Run("temporary_proof_denial_settles_within_the_same_deadline", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		scope, err := r.CreateDiagnosticScope([]string{})
		if err != nil {
			t.Fatal(err)
		}
		defer r.RecordResource("diagnostic_scope", scope, "released")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSessionRuntimeProof$")
		cmd.Env = append(os.Environ(), "CARTULARY_SCOPE_TEST_PROTECTED=1", "CARTULARY_SCOPE_TEST_RESUME=1", "CARTULARY_DIAGNOSTIC_TOKEN="+scope.Token)
		input, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		defer input.Close()
		output, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
		if line, err := bufio.NewReader(output).ReadString('\n'); err != nil || line != "ready\n" {
			t.Fatal("protected fixture failed", line, err)
		}
		if _, marked, _, err := diagnosticMember(scope, cmd.Process.Pid); marked || !errors.Is(err, os.ErrPermission) {
			t.Fatal("fixture proof was not inaccessible", marked, err)
		}
		raw, _ := json.Marshal(scope)
		done := make(chan error, 1)
		go func() { done <- r.SettleDiagnosticScope(ctx, raw) }()
		select {
		case err := <-done:
			t.Fatal("temporary denial prematurely ended settlement", err)
		case <-time.After(50 * time.Millisecond):
		}
		if _, err := input.Write([]byte{'r'}); err != nil {
			t.Fatal("release proof-read barrier", err)
		}
		if err := <-done; err != nil {
			t.Fatal("readable marked process did not settle", err)
		}
		current, err := readProcess(cmd.Process.Pid)
		if !processGone(err) && (err != nil || current.state != "Z" && current.state != "X") {
			t.Fatal("settlement returned before marked process exited", err)
		}
	})
}
