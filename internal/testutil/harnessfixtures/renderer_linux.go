//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

type rendererSlot struct {
	diagnostics                                                    bool
	id                                                             int
	allocation, generation, profile, policy, borrower, file, state string
	physical                                                       map[string]any
	resident                                                       *residentRecord
	key                                                            rendererKey
}

// Image, package and browser identities are part of the admitted profile.
// They and trust are immutable for a physical generation, independently of
// application affinity or the suite that borrows it.
type rendererKey struct {
	Run, Profile, Trust string
	Identity            map[string]any
	Diagnostics         bool
}
type rendererOwner struct {
	mu        sync.Mutex
	options   ProductionOptions
	slots     []*rendererSlot
	changed   chan struct{}
	remaining func(rendererConsumerKey, string) int
	closed    bool
}

func (r *rendererOwner) setState(slot *rendererSlot, state string) {
	r.mu.Lock()
	slot.state = state
	r.mu.Unlock()
}
func (r *rendererOwner) signal() { close(r.changed); r.changed = make(chan struct{}) }
func (r *rendererOwner) reserve(ctx context.Context, unit, profile, policy string, diagnostics bool) (*rendererSlot, error) {
	if r.options.BrowserLaneLimit < 1 || r.options.BrowserLaneLimit > 6 {
		return nil, errors.New("missing or invalid resolved browser lane limit")
	}
	if policy != "pooled" && policy != "dedicated" {
		return nil, errors.New("unknown renderer lifetime")
	}
	for {
		r.mu.Lock()
		if r.closed {
			r.mu.Unlock()
			return nil, context.Canceled
		}
		for _, slot := range r.slots {
			if policy == "pooled" && slot.compatibility() == rendererCompatibility(profile, policy, diagnostics) && slot.borrower == "" && (slot.state == "ready" || slot.state == "unallocated") {
				slot.borrower = unit
				slot.state = "reserved"
				r.mu.Unlock()
				return slot, nil
			}
		}
		// Before acquisition there is no physical profile or trust identity to
		// retire. Keep the bounded slot and its immutable residency proof while
		// an unstarted consumer changes. ensure validates the exact claims.
		for _, slot := range r.slots {
			if slot.borrower == "" && slot.state == "unallocated" && slot.file == "" && slot.physical == nil {
				slot.profile, slot.policy, slot.diagnostics = profile, policy, diagnostics
				slot.borrower, slot.state = unit, "reserved"
				r.mu.Unlock()
				return slot, nil
			}
		}
		available := 0
		for id := 1; id <= r.options.BrowserLaneLimit; id++ {
			if !slices.ContainsFunc(r.slots, func(s *rendererSlot) bool { return s.id == id }) {
				available = id
				break
			}
		}
		if available != 0 {
			slot := &rendererSlot{id: available, allocation: identifier(), generation: identifier(), profile: profile, policy: policy, borrower: unit, state: "unallocated", diagnostics: diagnostics}
			r.slots = append(r.slots, slot)
			slices.SortFunc(r.slots, func(a, b *rendererSlot) int { return a.id - b.id })
			r.mu.Unlock()
			return slot, nil
		}
		changed := r.changed
		r.mu.Unlock()
		// No host activity claim is held while waiting for a logical slot.
		if _, err := r.retireIdle(ctx, nil, false, nil, 1, "renderer_slot_pressure"); err != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-changed:
		}
	}
}
func (r *rendererOwner) evidence(slot *rendererSlot, state, unit, borrow string) error {
	root := filepath.Join(r.options.RunRoot, "_shared", "renderers", slot.allocation)
	directory, err := harnesslocal.PrivateDirectory(root)
	if err != nil {
		return err
	}
	directory.Close()
	record := map[string]any{"schema_id": "cartulary.harness_renderer_event.v2", "run_id": r.options.Runtime.RunID, "allocation_id": slot.allocation, "generation": slot.generation, "slot_id": slot.id, "runtime_profile_id": slot.profile, "state": state, "unit_id": unit, "borrow_id": borrow}
	identity := borrow
	if identity == "" {
		identity = identifier()
	}
	return privateJSON(filepath.Join(root, fmt.Sprintf("%s-%s.json", state, identity)), "cartulary.harness_renderer_event.v2", record, false)
}
func (r *rendererOwner) borrow(ctx context.Context, slot *rendererSlot, environment map[string]string) (map[string]string, func(context.Context, bool) (bool, error), error) {
	p := r.options
	root, err := p.Runtime.RecoveryPath("renderers", slot.allocation)
	if err != nil {
		return nil, nil, err
	}
	directory, err := harnesslocal.PrivateDirectory(root)
	if err != nil {
		return nil, nil, err
	}
	directory.Close()
	slot.file = filepath.Join(root, "physical.json")
	dependencies := []string{}
	if slot.resident != nil {
		dependencies = append(dependencies, p.residency.reference("renderer:"+slot.allocation)...)
	}
	// Acquisition copies the admitted public trust bytes into the renderer.
	// The caller holds its stack through acquisition; the physical renderer then
	// owns that immutable copy and its residency independently of service lifetimes.
	// Depending on live services would let one cluster's retirement stop renderers
	// and clients belonging to unrelated allocations.
	p.Dependencies = dependencies
	env := p.environment()
	maps.Copy(env, environment)
	env["CARTULARY_HARNESS_RESOURCE_DEPENDENCIES"] = resourceDependencies(dependencies)
	delete(env, "CARTULARY_RENDERER_DIAGNOSTICS")
	if slot.diagnostics {
		env["CARTULARY_RENDERER_DIAGNOSTICS"] = "1"
	}
	profileBytes, err := os.ReadFile(filepath.Join(p.Root, "tools/frontend_visual_renderer_profile.json"))
	if err != nil {
		return nil, nil, err
	}
	profile := map[string]any{}
	if json.Unmarshal(profileBytes, &profile) != nil {
		return nil, nil, harnesslocal.ErrUnsafe
	}
	key := rendererKey{Run: p.Runtime.RunID, Profile: slot.profile, Trust: env["CARTULARY_WEB_E2E_TLS_TRUST_SHA256"], Identity: profile, Diagnostics: slot.diagnostics}
	if slot.physical != nil && !reflect.DeepEqual(slot.key, key) {
		return nil, nil, errors.New("renderer compatibility identity changed")
	}
	if slot.physical == nil {
		r.setState(slot, "provisioning")
		if err := r.evidence(slot, "provisioning", slot.borrower, ""); err != nil {
			return nil, nil, err
		}
		physical, err := p.physicalOperation(ctx, "renderer-lifecycle.mjs", "acquire", []string{slot.file}, env)
		if err != nil {
			r.setState(slot, "quarantine")
			if diagnostic, readErr := readObject(filepath.Join(root, "startup-diagnostic.json"), "cartulary.harness_renderer_startup.v1", 4096); readErr == nil {
				writeErr := privateJSON(filepath.Join(p.RunRoot, "_shared", "renderers", slot.allocation, "startup-diagnostic.json"), "cartulary.harness_renderer_startup.v1", diagnostic, false)
				err = errors.Join(err, writeErr)
				if diagnostic["cleanup_failed"] == true {
					err = errors.Join(err, CleanupFailure(errors.New("renderer startup cleanup failed")))
				}
			}
			return nil, nil, err
		}
		if harnessschema.Validate("cartulary.harness_renderer_physical.v1", physical) != nil || physical["run_id"] != p.Runtime.RunID {
			return nil, nil, harnesslocal.ErrUnsafe
		}
		slot.physical = physical
		if !reflect.DeepEqual(physical["attestation"].(map[string]any)["profile"], profile) {
			return nil, nil, errors.New("renderer profile differs from admitted identity")
		}
		slot.key = key
		publicRoot := filepath.Join(p.RunRoot, "_shared", "renderers", slot.allocation)
		for name, value := range map[string]any{"renderer-profile-attestation.json": physical["attestation"], "tls-trust-attestation.json": physical["trust"]} {
			schema := value.(map[string]any)["schema_id"].(string)
			if err := privateJSON(filepath.Join(publicRoot, name), schema, value, false); err != nil {
				return nil, nil, err
			}
		}
	}
	trust := slot.physical["trust"].(map[string]any)
	if "sha256:"+trust["trust_bundle_sha256"].(string) != env["CARTULARY_WEB_E2E_TLS_TRUST_SHA256"] {
		return nil, nil, errors.New("renderer trust generation mismatch")
	}
	borrow := identifier()
	attachment := map[string]any{"schema_id": "cartulary.harness_renderer_attachment.v1", "run_id": p.Runtime.RunID, "unit_id": slot.borrower, "borrow_id": borrow, "allocation_id": slot.allocation, "generation": slot.generation, "runtime_profile_id": slot.profile, "trust_digest": trust["trust_bundle_sha256"], "diagnostic_binding": slot.physical["diagnostic_binding"], "environment": slot.physical["environment"], "trust": trust, "attestation": slot.physical["attestation"], "attestation_ref": "_shared/renderers/" + slot.allocation + "/renderer-profile-attestation.json", "trust_ref": "_shared/renderers/" + slot.allocation + "/tls-trust-attestation.json"}
	file := filepath.Join(root, "borrow-"+borrow+".json")
	if err := privateJSON(file, "cartulary.harness_renderer_attachment.v1", attachment, false); err != nil {
		return nil, nil, err
	}
	r.setState(slot, "borrowed")
	if err := r.evidence(slot, "borrowed", slot.borrower, borrow); err != nil {
		return nil, nil, err
	}
	attached := map[string]string{"CARTULARY_RENDERER_ATTACHMENT": file}
	for key, value := range slot.physical["environment"].(map[string]any) {
		attached[key] = value.(string)
	}
	ref, err := p.Runtime.ResourceReference("browser_renderer", slot.physical["proof"])
	if err != nil {
		return nil, nil, err
	}
	attached["CARTULARY_RENDERER_RESOURCE_REF"] = ref
	if err := p.Runtime.RegisterEnvironment(attached); err != nil {
		return nil, nil, err
	}
	var once sync.Once
	var retained bool
	var closeErr error
	release := func(ctx context.Context, healthy bool) (bool, error) {
		once.Do(func() {
			r.setState(slot, "settling")
			closeErr = errors.Join(closeErr, r.evidence(slot, "settling", slot.borrower, borrow))
			if healthy && closeErr == nil {
				step, done := context.WithTimeout(ctx, 30*time.Second)
				_, err := p.physicalOperation(step, "renderer-lifecycle.mjs", "settle", []string{slot.file}, env)
				done()
				healthy = err == nil
			}
			closeErr = errors.Join(closeErr, removePrivate(file))
			r.mu.Lock()
			keep := healthy && closeErr == nil && slot.policy == "pooled" && r.future(slot, slot.borrower) > 0 && !r.closed
			if keep {
				closeErr = errors.Join(closeErr, r.evidence(slot, "ready", slot.borrower, borrow))
				retained = closeErr == nil
			}
			if retained {
				slot.state = "ready"
				slot.borrower = ""
				r.signal()
			}
			r.mu.Unlock()
			if !retained {
				if !healthy {
					closeErr = errors.Join(closeErr, r.evidence(slot, "quarantine", slot.borrower, borrow))
				}
				closeErr = errors.Join(closeErr, r.retire(ctx, slot))
			}
		})
		return retained, closeErr
	}
	return attached, release, nil
}
func (r *rendererOwner) retire(ctx context.Context, slot *rendererSlot) error {
	p := r.options
	var err error
	if slot.physical != nil {
		err = p.closeResource(ctx, "browser_renderer", slot.physical["proof"], func(ctx context.Context) error {
			return p.Runtime.AttemptCleanup(ctx, "browser_renderer", slot.physical["proof"], func(ctx context.Context) error {
				_, err := p.physicalOperation(ctx, "renderer-lifecycle.mjs", "release", []string{slot.file}, p.environment())
				return err
			})
		})
	}
	if err == nil && slot.resident != nil {
		err = p.residency.close(ctx, "renderer:"+slot.allocation)
	}
	if err != nil {
		r.setState(slot, "quarantine")
		return err
	}
	if slot.file != "" {
		if directory, openErr := harnesslocal.OpenDirectory(filepath.Dir(slot.file), true); openErr == nil {
			err = errors.Join(harnesssession.RemoveOwnedTree(directory, filepath.Dir(slot.file)), directory.Close())
		} else {
			err = openErr
		}
	}
	if slot.file != "" || slot.physical != nil {
		err = errors.Join(err, r.evidence(slot, "retired", slot.borrower, ""))
	}
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.slots = slices.DeleteFunc(r.slots, func(candidate *rendererSlot) bool { return candidate == slot })
	r.signal()
	r.mu.Unlock()
	return nil
}
func (r *rendererOwner) evictIdle(ctx context.Context, exclude *rendererSlot) error {
	_, err := r.retireIdle(ctx, exclude, false, nil, 0, "exclusive_admission")
	return err
}

func (r *rendererOwner) unreserve(slot *rendererSlot) {
	r.mu.Lock()
	defer r.mu.Unlock()
	slot.borrower = ""
	slot.state = "unallocated"
	if slot.physical != nil {
		slot.state = "ready"
	}
	r.signal()
}

func (r *rendererOwner) future(slot *rendererSlot, exclude string) int {
	if r.remaining == nil {
		return 0
	}
	return r.remaining(slot.compatibility(), exclude)
}

func (r *rendererOwner) retireIdle(ctx context.Context, exclude *rendererSlot, unusedOnly bool, pending *harnessadmission.Lease, maximum int, reason string) (int, error) {
	r.mu.Lock()
	slots := slices.Clone(r.slots)
	r.mu.Unlock()
	slices.SortFunc(slots, func(a, b *rendererSlot) int { return a.id - b.id })
	retired := 0
	for _, slot := range slots {
		if slot == exclude {
			continue
		}
		did, err := r.retireSlotIfIdle(ctx, slot, unusedOnly, pending, reason)
		if err != nil {
			return retired, err
		}
		if did {
			retired++
		}
		if maximum > 0 && retired >= maximum {
			break
		}
	}
	return retired, nil
}

func (r *rendererOwner) retireSlotIfIdle(ctx context.Context, slot *rendererSlot, unusedOnly bool, pending *harnessadmission.Lease, reason string) (bool, error) {
	r.mu.Lock()
	if !slices.Contains(r.slots, slot) || slot.borrower != "" || (slot.state != "ready" && slot.state != "unallocated") {
		r.mu.Unlock()
		return false, nil
	}
	unused := r.future(slot, "") == 0
	if unusedOnly && !unused {
		r.mu.Unlock()
		return false, nil
	}
	prior := slot.state
	slot.state = "retiring" // Pin identity and exclude borrowers before evaluation.
	r.mu.Unlock()
	restore := func() { r.mu.Lock(); slot.state = prior; r.signal(); r.mu.Unlock() }
	if pending != nil && !unused {
		if slot.physical == nil || slot.resident == nil {
			restore()
			return false, nil
		}
		useful, err := pending.NeedsRetirement(ctx, []harnessadmission.Resident{slot.resident.Allocation})
		if err != nil || !useful {
			restore()
			return false, err
		}
	}
	if unused && reason == "capacity_pressure" {
		reason = "terminal_borrower_release"
	}
	residents := []harnessadmission.Resident{}
	if slot.resident != nil && slot.physical != nil {
		residents = append(residents, slot.resident.Allocation)
	}
	observe := harnessadmission.Retirement(ctx, residents, reason)
	operation := func(ctx context.Context) error { return r.retire(ctx, slot) }
	var err error
	if token, _ := ctx.Value(activityKey{}).(string); token != "" || slot.resident == nil || slot.physical == nil {
		err = operation(ctx)
	} else {
		err = r.options.residency.drain(ctx, []*residentRecord{slot.resident}, operation)
	}
	observe(err)
	return err == nil, err
}
