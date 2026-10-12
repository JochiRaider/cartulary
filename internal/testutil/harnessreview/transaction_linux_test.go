//go:build linux

package harnessreview

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

var publicationStages = []string{"before_intent", "after_intent", "after_seal", "after_pre_scan", "after_detail_purge", "after_extras", "after_terminal", "after_locator", "after_scan", "after_runtime_removal", "before_registry_removal"}

func testTerminalReceipt(l Locator) Receipt {
	return Receipt{Schema: "cartulary.ui_review_receipt.v2", Command: commandID("ui-review"), Session: l.Session, Mode: l.Mode, State: "closed", Status: "ok", Started: l.Created, Finished: l.Updated, Failures: []Failure{}, Cleanup: "complete"}
}
func retirePublicationTestOwner(t *testing.T, p *diskPublisher) {
	t.Helper()
	dead := harnessadmission.Proof{Boot: "independent-dead-publication-fixture", PID: 12345, Start: "1"}
	data, _ := json.Marshal(dead)
	if _, err := os.Stat(p.runtime.Root); err == nil {
		if err := harnesslocal.ReplaceFile(filepath.Join(p.runtime.Root, "runtime-process.json"), data); err != nil {
			t.Fatal(err)
		}
	}
	p.record.Process = dead
	if err := writeRegistry(p.record, true); err != nil {
		t.Fatal(err)
	}
}

func testPublicationTransactions(t *testing.T) {
	for _, presentation := range []string{"ui", "design"} {
		for _, stage := range publicationStages {
			t.Run(presentation+"_resumable_"+stage, func(t *testing.T) {
				p, l := publicationFixture(t)
				var publisher Publisher = p
				if presentation == "design" {
					p.record.Presentation = "design"
					p.record.Profile = "default"
					if err := writeRegistry(p.record, true); err != nil {
						t.Fatal(err)
					}
					publisher = &designPublisher{diskPublisher: p, profile: "default", driver: &productionDriver{}}
					if err := publisher.Locator(l); err != nil {
						t.Fatal(err)
					}
				}
				l.State = "closed"
				detail := filepath.Join(p.runtime.Root, "private-detail.log")
				if err := os.WriteFile(detail, []byte("disposable private material"), 0600); err != nil {
					t.Fatal(err)
				}
				p.fault = func(current string) error {
					if current == stage {
						return errors.New("independent publication fault")
					}
					return nil
				}
				if _, err := publisher.Terminal(l, testTerminalReceipt(l), false); err == nil {
					t.Fatal("fault did not fail")
				}
				if _, err := os.Stat(detail); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("independent detail purge skipped", err)
				}
				terminal := filepath.Join(p.runRoot, "ui-review/terminal.json")
				prior, _ := os.ReadFile(terminal)
				retirePublicationTestOwner(t, p)
				registryBefore, err := os.ReadFile(registryPath(l.Session))
				if err != nil {
					t.Fatal(err)
				}
				observed, observeErr := control(context.Background(), p.record.Repository, commandRequest{Command: "ui-review-status", Locator: p.record.Locator}, nil)
				if observeErr == nil && observed.Status == "ok" {
					t.Fatal("pending transaction reported successful closure")
				}
				registryAfter, err := os.ReadFile(registryPath(l.Session))
				if err != nil || string(registryBefore) != string(registryAfter) {
					t.Fatal("status mutated recovery", err)
				}
				q := commandRequest{Command: "ui-review-stop", Locator: p.record.Locator}
				value, err := control(context.Background(), p.record.Repository, q, nil)
				if err != nil || value.Receipt == nil {
					t.Fatal("resume", value, err)
				}
				after, err := os.ReadFile(terminal)
				if err != nil {
					t.Fatal(err)
				}
				if len(prior) > 0 && string(prior) != string(after) {
					t.Fatal("terminal bytes changed during recovery")
				}
				var receipt Receipt
				if err := json.Unmarshal(after, &receipt); err != nil {
					t.Fatal(err)
				}
				if stage != "before_intent" && receipt.Finished != l.Updated {
					t.Fatal("recovery regenerated timestamp")
				}
				second, err := control(context.Background(), p.record.Repository, q, nil)
				if err != nil || second.Receipt == nil || *second.Receipt != *value.Receipt {
					t.Fatal("nonconvergent repeated recovery", second, err)
				}
				if _, err := os.Stat(p.runtime.Root); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("private runtime remains", err)
				}
				if _, err := os.Stat(registryPath(l.Session)); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("registry remains", err)
				}
			})
		}
	}
	t.Run("post_terminal_scan_retains_authority_and_history", func(t *testing.T) {
		p, l := publicationFixture(t)
		l.State = "closed"
		const sentinel = "private-post-terminal-scan-sentinel"
		if err := p.runtime.RegisterSecret(sentinel); err != nil {
			t.Fatal(err)
		}
		p.fault = func(stage string) error {
			if stage == "after_locator" {
				return os.WriteFile(filepath.Join(p.runRoot, "unsafe.log"), []byte(sentinel), 0600)
			}
			return nil
		}
		if _, err := p.Terminal(l, testTerminalReceipt(l), false); err == nil {
			t.Fatal("post-terminal scan accepted secret")
		}
		secrets, err := p.runtime.ForbiddenValues()
		if err != nil || len(secrets) != 1 || secrets[0] != sentinel {
			t.Fatal("scanner inputs lost", err)
		}
		data, err := os.ReadFile(filepath.Join(p.runRoot, "ui-review/terminal.json"))
		if err != nil {
			t.Fatal(err)
		}
		retirePublicationTestOwner(t, p)
		value, err := control(context.Background(), p.record.Repository, commandRequest{Command: "ui-review-stop", Locator: p.record.Locator}, nil)
		if err != nil || value.Receipt == nil {
			t.Fatal(value, err)
		}
		after, err := os.ReadFile(filepath.Join(p.runRoot, "ui-review/terminal.json"))
		if err != nil || string(data) != string(after) {
			t.Fatal("scan recovery rewrote terminal", err)
		}
	})
	t.Run("conflicting_terminal_rejected_with_safe_detail_cleanup", func(t *testing.T) {
		p, l := publicationFixture(t)
		l.State = "closed"
		p.fault = func(stage string) error {
			if stage == "after_detail_purge" {
				return errors.New("interrupt before bytes")
			}
			return nil
		}
		if _, err := p.Terminal(l, testTerminalReceipt(l), false); err == nil {
			t.Fatal("fault missing")
		}
		if !p.record.Publication.DetailPurged {
			t.Fatal("fixture did not record prior purge")
		}
		detail, err := p.runtime.PrivatePath("new-recovery-detail.log")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(detail, []byte("new attempt private detail"), 0600); err != nil {
			t.Fatal(err)
		}
		foreign := []byte("foreign immutable outcome\n")
		terminal := filepath.Join(p.runRoot, "ui-review/terminal.json")
		if err := os.WriteFile(terminal, foreign, 0600); err != nil {
			t.Fatal(err)
		}
		retirePublicationTestOwner(t, p)
		value, err := control(context.Background(), p.record.Repository, commandRequest{Command: "ui-review-stop", Locator: p.record.Locator}, nil)
		if err == nil && value.Status == "ok" {
			t.Fatal("foreign terminal adopted")
		}
		after, err := os.ReadFile(terminal)
		if err != nil || string(after) != string(foreign) {
			t.Fatal("foreign bytes changed", err)
		}
		if _, err := os.Stat(registryPath(l.Session)); err != nil {
			t.Fatal("conflict lost recovery state", err)
		}
		if _, err := os.Stat(detail); !os.IsNotExist(err) {
			t.Fatal("historical purge suppressed current cleanup", err)
		}
		// Reject inconsistent saved intent before terminal publication, while still
		// removing detail from this later recovery attempt.
		detail, err = p.runtime.PrivatePath("early-intent-detail.log")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(detail, []byte("later private detail"), 0600); err != nil {
			t.Fatal(err)
		}
		p.record.Publication.Locator.Receipt.Bytes++
		if err = writeRegistry(p.record, true); err != nil {
			t.Fatal(err)
		}
		if _, err = control(context.Background(), p.record.Repository, commandRequest{Command: "ui-review-stop", Locator: p.record.Locator}, nil); err == nil {
			t.Fatal("inconsistent intent accepted")
		}
		if _, err = os.Stat(detail); !os.IsNotExist(err) {
			t.Fatal("early recovery exit retained detail", err)
		}
	})
	t.Run("concurrent_recovery_converges", func(t *testing.T) {
		p, l := publicationFixture(t)
		l.State = "closed"
		p.fault = func(stage string) error {
			if stage == "after_terminal" {
				return errors.New("interrupted after bytes")
			}
			return nil
		}
		if _, err := p.Terminal(l, testTerminalReceipt(l), false); err == nil {
			t.Fatal("fault missing")
		}
		retirePublicationTestOwner(t, p)
		type reply struct {
			v   Result
			err error
		}
		replies := make(chan reply, 2)
		for i := 0; i < 2; i++ {
			go func() {
				v, err := control(context.Background(), p.record.Repository, commandRequest{Command: "ui-review-stop", Locator: p.record.Locator}, nil)
				replies <- reply{v, err}
			}()
		}
		first, second := <-replies, <-replies
		if first.err != nil || second.err != nil || first.v.Receipt == nil || second.v.Receipt == nil || *first.v.Receipt != *second.v.Receipt {
			t.Fatal(first, second)
		}
	})
}

func publicationCrashHelper(t *testing.T, stage string) {
	p, l := publicationFixture(t)
	l.State = "closed"
	handoff, _ := json.Marshal(map[string]string{"locator": p.record.Locator, "repository": p.record.Repository, "runtime": p.runtime.Root, "registry": registryPath(l.Session), "socket": p.record.Socket})
	if err := os.WriteFile(filepath.Join(os.Getenv("CARTULARY_REVIEW_PUBLICATION_CRASH_BASE"), "handoff.json"), handoff, 0600); err != nil {
		t.Fatal(err)
	}
	if err := p.runtime.RegisterSecret("private-crash-publication-sentinel"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.runtime.Root, "private.log"), []byte("private-crash-publication-sentinel"), 0600); err != nil {
		t.Fatal(err)
	}
	p.fault = func(current string) error {
		if current == stage {
			os.Exit(19)
		}
		return nil
	}
	_, err := p.Terminal(l, testTerminalReceipt(l), false)
	t.Fatal("crash boundary not reached", err)
}

func testPublicationCrashRecovery(t *testing.T) {
	for _, stage := range append(append([]string{}, publicationStages...), "after_registry_removal") {
		t.Run("process_death_"+stage, func(t *testing.T) {
			base := t.TempDir()
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			child := exec.Command(executable, "-test.run=^TestReviewPublication$")
			child.Env = append(os.Environ(), "CARTULARY_REVIEW_PUBLICATION_CRASH_BASE="+base, "CARTULARY_REVIEW_PUBLICATION_CRASH_STAGE="+stage)
			output, err := child.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 19 {
				t.Fatalf("child did not die at boundary: %v %s", err, output)
			}
			data, err := os.ReadFile(filepath.Join(base, "handoff.json"))
			if err != nil {
				t.Fatal(err)
			}
			handoff := map[string]string{}
			if err := json.Unmarshal(data, &handoff); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				os.Remove(handoff["registry"])
				os.Remove(handoff["socket"])
				os.Remove(handoff["registry"] + ".lock")
			})
			terminal := filepath.Join(base, "fixture/ui-review/terminal.json")
			prior, _ := os.ReadFile(terminal)
			q := commandRequest{Command: "ui-review-stop", Locator: handoff["locator"]}
			value, err := control(context.Background(), handoff["repository"], q, nil)
			if err != nil || value.Receipt == nil {
				t.Fatal("new-process recovery", value, err)
			}
			after, err := os.ReadFile(terminal)
			if err != nil || len(prior) > 0 && string(prior) != string(after) {
				t.Fatal("crash changed terminal bytes", err)
			}
			if stage != "before_intent" && value.Exit != 0 {
				t.Fatal("crash replaced a persisted successful intent", value)
			}
			if _, err := os.Stat(handoff["runtime"]); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("runtime survived completed recovery", err)
			}
		})
	}
}
