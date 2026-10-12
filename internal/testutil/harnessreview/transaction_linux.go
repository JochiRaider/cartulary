//go:build linux

package harnessreview

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

// Publication is a durable transaction. Progress is a recovery aid, never
// permission to trust an existing artifact without checking its exact bytes.
type publicationState struct {
	Locator           Locator            `json:"locator"`
	DetailPurged      bool               `json:"detail_purged"`
	TerminalPublished bool               `json:"terminal_published"`
	LocatorPublished  bool               `json:"locator_published"`
	Scanned           bool               `json:"scanned"`
	RuntimeRemoved    bool               `json:"runtime_removed"`
	Extras            []publicationExtra `json:"extras"`
}
type publicationExtra struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func receiptReference(receipt Receipt) (*Reference, []byte, error) {
	data, err := jsonDocument(receipt.Schema, receipt)
	if err != nil {
		return nil, nil, err
	}
	return &Reference{Path: "ui-review/terminal.json", Bytes: len(data), SHA256: fmt.Sprintf("%x", sha256.Sum256(data)), MediaType: "application/json"}, data, nil
}

func (p *diskPublisher) checkpoint(stage string) error {
	if p.fault != nil {
		return p.fault(stage)
	}
	return nil
}

func (p *diskPublisher) validateIntent() error {
	r, tx := p.record.Outcome, p.record.Publication
	if r == nil || tx == nil || r.Session != p.record.Session || r.Command != commandID("ui-review") || r.Operation != nil || r.Cleanup == "not_terminal" || r.Mode != tx.Locator.Mode || r.State != tx.Locator.State || (r.State == "closed") != (r.Status == "ok") || r.Started != tx.Locator.Created || (r.Exit == 0) != (r.Status == "ok") || (len(r.Failures) == 0) != (r.Status == "ok") {
		return failure("invalid_artifact")
	}
	for _, f := range r.Failures {
		if err := f.validate(); err != nil {
			return err
		}
	}
	ref, _, err := receiptReference(*r)
	if err != nil || tx.Locator.Receipt == nil || *ref != *tx.Locator.Receipt || tx.Locator.Session != p.record.Session || tx.Locator.RunID != p.record.Runtime.RunID || tx.Locator.Updated != r.Finished {
		return failure("invalid_artifact")
	}
	if p.record.Locator != filepath.Join(p.runRoot, "ui-review/session.json") || tx.RuntimeRemoved && (!tx.Scanned || !p.record.Released) {
		return failure("session_mismatch")
	}
	if (p.record.Presentation == "design" && len(tx.Extras) != 2) || (p.record.Presentation == "ui" && len(tx.Extras) != 0) {
		return failure("invalid_artifact")
	}
	seen := map[string]bool{}
	for _, extra := range tx.Extras {
		if seen[extra.Path] {
			return failure("invalid_artifact")
		}
		seen[extra.Path] = true
		value, err := harnessschema.Decode([]byte(extra.Content))
		if err != nil {
			return err
		}
		object, ok := value.(map[string]any)
		if !ok || object["run_id"] != p.record.Runtime.RunID {
			return failure("session_mismatch")
		}
		switch extra.Path {
		case "cleanup-results.json":
			if harnessschema.Validate("cartulary.harness_cleanup_results.v1", value) != nil {
				return failure("invalid_artifact")
			}
		case "review-session.json":
			if object["purpose"] != "interactive_design_review" || object["runtime_profile_id"] != p.record.Profile || object["state"] != r.State {
				return failure("session_mismatch")
			}
		default:
			return failure("invalid_artifact")
		}
	}
	return nil
}

// Only mutable, session-bound locators/presentation detail may be replaced.
// Terminal receipts and terminal design artifacts are adoptable only byte-for-byte.
func publishMatching(file string, data []byte, mutable func([]byte) bool) error {
	prior, err := harnesslocal.ReadFile(file, LimitRequest, false)
	if err == nil {
		if string(prior) == string(data) {
			return nil
		}
		if mutable == nil || !mutable(prior) {
			return failure("session_mismatch")
		}
		return harnesslocal.ReplaceFile(file, data)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	directory, err := harnesslocal.PrivateDirectory(filepath.Dir(file))
	if err != nil {
		return err
	}
	defer directory.Close()
	if err = harnesslocal.PublishNew(file, data); errors.Is(err, os.ErrExist) {
		prior, err = harnesslocal.ReadFile(file, LimitRequest, false)
		if err == nil && string(prior) != string(data) {
			return failure("session_mismatch")
		}
	}
	return err
}

func (p *diskPublisher) mutableLocator(data []byte) bool {
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate("cartulary.ui_review_session.v1", value) != nil {
		return false
	}
	var prior Locator
	if json.Unmarshal(data, &prior) != nil {
		return false
	}
	l := p.record.Publication.Locator
	return prior.Session == l.Session && prior.RunID == l.RunID && prior.Mode == l.Mode && prior.Created == l.Created && prior.Receipt == nil
}
func (p *diskPublisher) mutableDesign(data []byte) bool {
	var prior map[string]any
	if json.Unmarshal(data, &prior) != nil {
		return false
	}
	return prior["purpose"] == "interactive_design_review" && prior["run_id"] == p.record.Runtime.RunID && prior["runtime_profile_id"] == p.record.Profile && prior["state"] != "closed" && prior["state"] != "failed"
}

func (p *diskPublisher) publicationFailure(code string) error {
	tx := p.record.Publication
	if tx == nil || tx.TerminalPublished {
		return nil
	}
	// A crash may have published bytes before persisting progress. Their presence
	// freezes the intent even when the subsequent registry update failed.
	if _, err := harnesslocal.ReadFile(filepath.Join(p.runRoot, "ui-review/terminal.json"), LimitRequest, false); !errors.Is(err, os.ErrNotExist) {
		return nil
	}
	for _, extra := range tx.Extras {
		if existing, err := harnesslocal.ReadFile(filepath.Join(p.runRoot, extra.Path), LimitRequest, false); err == nil && string(existing) == extra.Content {
			return nil
		}
	}
	r := p.record.Outcome
	f := failure(code)
	r.Failures = appendFailure(append([]Failure{}, r.Failures...), f.Record)
	if r.Exit == 0 {
		r.Exit = f.Exit
	}
	r.State = "failed"
	r.Status = "error"
	if code == "cleanup_failed" {
		r.Cleanup = "failed"
	}
	tx.Locator.State = r.State
	ref, _, err := receiptReference(*r)
	if err != nil {
		return err
	}
	tx.Locator.Receipt = ref
	for i, extra := range tx.Extras {
		if extra.Path != "review-session.json" {
			continue
		}
		var v map[string]any
		if err := json.Unmarshal([]byte(extra.Content), &v); err != nil {
			return err
		}
		v["state"] = r.State
		v["failure_class"] = r.Failures[0].Class
		v["failure_reason"] = r.Failures[0].Reason
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		tx.Extras[i].Content = string(append(data, '\n'))
	}
	return writeRegistry(p.record, true)
}

func (p *diskPublisher) Terminal(locator Locator, receipt Receipt, cleanupFailed bool) (ref *Reference, result error) {
	intentSaved := p.record.Publication != nil
	defer func() {
		if result == nil {
			return
		}
		if p.runtime != nil && !p.runtimeClosed {
			resources, err := p.runtime.Resources(4096)
			if err == nil && len(resources) == 0 {
				_, err = p.runtime.BeginFinalization()
			}
			result = errors.Join(result, err, p.runtime.PreserveRecovery())
		}
		if intentSaved {
			result = errors.Join(result, p.publicationFailure("unsafe_artifact"))
		}
	}()
	if p.record.Publication == nil {
		if _, err := harnesslocal.ReadFile(filepath.Join(p.runRoot, "ui-review/terminal.json"), LimitRequest, false); !errors.Is(err, os.ErrNotExist) {
			return nil, failure("session_mismatch")
		}
		if err := p.checkpoint("before_intent"); err != nil {
			return nil, err
		}
		ref, _, err := receiptReference(receipt)
		if err != nil {
			return nil, err
		}
		locator.Receipt = ref
		locator.Updated = receipt.Finished
		receipt.Failures = append([]Failure{}, receipt.Failures...)
		p.record.Outcome = &receipt
		p.record.Publication = &publicationState{Locator: locator, Extras: append([]publicationExtra{}, p.extras...)}
		if err := p.validateIntent(); err != nil {
			return nil, err
		}
		if err := writeRegistry(p.record, true); err != nil {
			return nil, err
		}
		intentSaved = true
	}
	if err := p.validateIntent(); err != nil {
		return nil, err
	}
	if err := p.checkpoint("after_intent"); err != nil {
		return nil, err
	}
	tx := p.record.Publication
	_, expected, err := receiptReference(*p.record.Outcome)
	if err != nil {
		return nil, err
	}
	if existing, err := harnesslocal.ReadFile(filepath.Join(p.runRoot, "ui-review/terminal.json"), LimitRequest, false); err == nil {
		if string(existing) != string(expected) {
			return nil, failure("session_mismatch")
		}
		tx.TerminalPublished = true
		if err := writeRegistry(p.record, true); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if p.runtime != nil && !p.runtimeClosed {
		p.record.Released = false
		tx.Scanned = false
		resources, resourceErr := p.runtime.Resources(4096)
		var sealErr, errorScan, errorPurge error
		if resourceErr == nil && len(resources) == 0 {
			p.secrets, sealErr = p.runtime.BeginFinalization()
			if sealErr == nil {
				p.record.Released = !cleanupFailed
				sealErr = writeRegistry(p.record, true)
			}
		} else if resourceErr == nil {
			p.secrets, sealErr = p.runtime.ForbiddenValues()
		}
		sealErr = errors.Join(sealErr, p.checkpoint("after_seal"))
		if resourceErr == nil && sealErr == nil {
			_, errorScan = harnesssession.ScanRetained(p.runRoot, p.secrets)
		}
		errorScan = errors.Join(errorScan, p.checkpoint("after_pre_scan"))
		// Independent of scan/publication success. Scanner authority remains under
		// the private runtime until required scans and artifact validation finish.
		errorPurge = p.runtime.PreserveRecovery()
		if errorPurge == nil {
			tx.DetailPurged = true
			errorPurge = writeRegistry(p.record, true)
		}
		errorPurge = errors.Join(errorPurge, p.checkpoint("after_detail_purge"))
		if err := errors.Join(resourceErr, sealErr, errorScan, errorPurge); err != nil {
			return nil, err
		}
		if cleanupFailed || len(resources) > 0 {
			if err := p.publicationFailure("cleanup_failed"); err != nil {
				return nil, err
			}
		}
	} else if !tx.Scanned || !p.record.Released {
		return nil, failure("unsafe_artifact")
	}
	for _, extra := range tx.Extras {
		var mutable func([]byte) bool
		if extra.Path == "review-session.json" {
			mutable = p.mutableDesign
		}
		if err := publishMatching(filepath.Join(p.runRoot, extra.Path), []byte(extra.Content), mutable); err != nil {
			return nil, err
		}
	}
	if err := p.checkpoint("after_extras"); err != nil {
		return nil, err
	}
	ref, data, err := receiptReference(*p.record.Outcome)
	if err != nil {
		return nil, err
	}
	if err := publishMatching(filepath.Join(p.runRoot, ref.Path), data, nil); err != nil {
		return nil, err
	}
	if err := p.checkpoint("after_terminal"); err != nil {
		return nil, err
	}
	tx.TerminalPublished = true
	if err := writeRegistry(p.record, true); err != nil {
		return nil, err
	}
	data, err = jsonDocument(tx.Locator.Schema, tx.Locator)
	if err != nil {
		return nil, err
	}
	if err := publishMatching(p.record.Locator, data, p.mutableLocator); err != nil {
		return nil, err
	}
	if err := p.checkpoint("after_locator"); err != nil {
		return nil, err
	}
	tx.LocatorPublished = true
	if err := writeRegistry(p.record, true); err != nil {
		return nil, err
	}
	if p.runtime != nil && !p.runtimeClosed {
		if _, err := harnesssession.ScanRetained(p.runRoot, p.secrets); err != nil {
			return nil, err
		}
		tx.Scanned = p.record.Released
		if err := writeRegistry(p.record, true); err != nil {
			return nil, err
		}
		if err := p.checkpoint("after_scan"); err != nil {
			return nil, err
		}
		if !p.record.Released {
			if p.record.Outcome.Exit == 0 {
				return ref, failure("cleanup_failed")
			}
			return ref, nil
		}
		if err := p.runtime.Close(); err != nil {
			return nil, err
		}
		p.runtimeClosed = true
	}
	if err := p.checkpoint("after_runtime_removal"); err != nil {
		return nil, err
	}
	tx.RuntimeRemoved = true
	if err := writeRegistry(p.record, true); err != nil {
		return nil, err
	}
	if err := p.checkpoint("before_registry_removal"); err != nil {
		return nil, err
	}
	if err := removeRegistry(p.record); err != nil {
		return nil, err
	}
	if err := p.checkpoint("after_registry_removal"); err != nil {
		return nil, err
	}
	return ref, nil
}
