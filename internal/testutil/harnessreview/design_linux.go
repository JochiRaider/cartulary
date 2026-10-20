//go:build linux

package harnessreview

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

// Design review shares the native session/fixture lifecycle. Its presentation
// remains helper evidence and never claims a passing product row or graph.
type designReady struct {
	Origin    string `json:"public_origin"`
	Samples   string `json:"samples"`
	Private   string `json:"private_directory"`
	Scenarios []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"scenarios"`
}

func writeDesign(file string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	dir, err := harnesslocal.PrivateDirectory(filepath.Dir(file))
	if err != nil {
		return err
	}
	defer dir.Close()
	return harnesslocal.ReplaceFileAt(dir, filepath.Base(file), append(data, '\n'))
}
func publishSource(root string, facts *adapterFacts, profile string) error {
	return writeDesign(filepath.Join(root, "run-manifest.json"), map[string]any{"purpose": "interactive_design_review", "source_digest": facts.Source, "source_commit": facts.Commit, "toolchain_digest": facts.Toolchain, "runtime_profile_id": profile})
}

type designPublisher struct {
	*diskPublisher
	profile string
	driver  *productionDriver
	ready   *designReady
}

func (p *designPublisher) Locator(v Locator) error {
	if err := p.diskPublisher.Locator(v); err != nil {
		return err
	}
	if p.record.Publication != nil {
		return nil
	}
	return p.details(v.State, nil)
}
func (p *designPublisher) detailValue(state string, failures []Failure) map[string]any {
	v := map[string]any{"purpose": "interactive_design_review", "run_id": p.record.Runtime.RunID, "runtime_profile_id": p.profile, "state": state}
	if p.ready != nil {
		v["public_origin"] = p.ready.Origin
		v["samples"] = p.ready.Samples
		v["scenarios"] = p.ready.Scenarios
	}
	if len(failures) > 0 {
		v["failure_class"] = failures[0].Class
		v["failure_reason"] = failures[0].Reason
	}
	if state == "closed" || state == "failed" {
		v["artifact_refs"] = []string{"cleanup-results.json"}
	}
	return v
}
func (p *designPublisher) details(state string, failures []Failure) error {
	return writeDesign(filepath.Join(p.runRoot, "review-session.json"), p.detailValue(state, failures))
}
func (p *designPublisher) Terminal(locator Locator, receipt Receipt, cleanupFailed bool) (*Reference, error) {
	if p.record.Publication != nil {
		return p.diskPublisher.Terminal(locator, receipt, cleanupFailed)
	}
	cleanup := make([]map[string]any, 0, len(p.driver.cleanup))
	for i, step := range p.driver.cleanup {
		outcome := "completed"
		var class, reason any
		var diagnostic *harnesssession.CleanupDiagnostic
		failures := []map[string]any{}
		if step.err != nil {
			outcome = "failed"
			if errors.Is(step.err, harnessfixtures.ErrCleanupBlocked) {
				outcome = "blocked"
			}
			if outcome == "failed" {
				diagnostic = harnesssession.DescribeCleanupFailure(harnesssession.AtCleanupStage(step.operation, step.err))
				class = "harness"
				reason = "cleanup_error"
				failures = append(failures, map[string]any{"failure_class": class, "failure_reason": reason})
			}
		}
		cleanup = append(cleanup, map[string]any{"sequence": i + 1, "operation": step.operation, "unit_id": nil, "fixture_lease_id": nil, "outcome": outcome, "failure_class": class, "failure_reason": reason, "failures": failures, "artifact_refs": []string{}, "diagnostic": diagnostic})
	}
	cleanupBytes, err := jsonDocument("cartulary.harness_cleanup_results.v2", map[string]any{"schema_id": "cartulary.harness_cleanup_results.v2", "run_id": p.record.Runtime.RunID, "results": cleanup})
	if err != nil {
		return nil, err
	}
	detailBytes, err := json.Marshal(p.detailValue(locator.State, receipt.Failures))
	if err != nil {
		return nil, err
	}
	p.extras = []publicationExtra{{Path: "review-session.json", Content: string(append(detailBytes, '\n'))}, {Path: "cleanup-results.json", Content: string(cleanupBytes)}}
	return p.diskPublisher.Terminal(locator, receipt, cleanupFailed)
}
