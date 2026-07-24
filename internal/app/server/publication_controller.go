package server

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

type PublicationState string

const (
	PublicationUnpublished PublicationState = "unpublished"
	PublicationPrepared    PublicationState = "prepared"
	PublicationCommitted   PublicationState = "committed"
	PublicationServing     PublicationState = "serving"
	PublicationFailed      PublicationState = "failed"
)

type PublicationController struct {
	mu           sync.RWMutex
	lifecycle    *processlifecycle.Controller
	state        PublicationState
	plan         *extensions.PublicationPlan
	expected     map[string]string
	acknowledged map[string]struct{}
}

func NewPublicationController(lifecycle *processlifecycle.Controller) *PublicationController {
	return &PublicationController{
		lifecycle:    lifecycle,
		state:        PublicationUnpublished,
		expected:     map[string]string{},
		acknowledged: map[string]struct{}{},
	}
}

func (c *PublicationController) Prepare(plan extensions.PublicationPlan) error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != PublicationUnpublished || c.lifecycle == nil || c.lifecycle.AdmissionOpen() {
		return c.failLocked("invalid_prepare_state")
	}
	summary := plan.Summary()
	if summary.SchemaID != "cartulary.extension_publication_plan.v1" {
		return c.failLocked("invalid_plan")
	}
	expected := map[string]string{}
	for _, listener := range plan.Listeners() {
		if listener.ComponentID == "" {
			return c.failLocked("invalid_listener")
		}
		if _, duplicate := expected[listener.ComponentID]; duplicate {
			return c.failLocked("duplicate_component")
		}
		expected[listener.ComponentID] = summary.ListenerPlanSHA256
	}
	for _, worker := range plan.Workers() {
		componentID := "worker:" + worker.ProfileID + ":" + worker.WorkerKind
		if worker.ProfileID == "" || worker.WorkerKind == "" {
			return c.failLocked("invalid_worker")
		}
		if _, duplicate := expected[componentID]; duplicate {
			return c.failLocked("duplicate_component")
		}
		expected[componentID] = summary.WorkerPlanSHA256
	}
	if len(expected) < 3 {
		return c.failLocked("missing_required_components")
	}
	c.plan = &plan
	c.expected = expected
	c.acknowledged = map[string]struct{}{}
	c.state = PublicationPrepared
	return nil
}

func (c *PublicationController) Acknowledge(componentID string, planSHA256 string, preparationErr error) error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != PublicationPrepared {
		return c.failLocked("early_acknowledgment")
	}
	expectedDigest, known := c.expected[componentID]
	if !known {
		return c.failLocked("unknown_acknowledgment")
	}
	if _, duplicate := c.acknowledged[componentID]; duplicate {
		return c.failLocked("duplicate_acknowledgment")
	}
	if preparationErr != nil || planSHA256 != expectedDigest {
		return c.failLocked("failed_acknowledgment")
	}
	c.acknowledged[componentID] = struct{}{}
	return nil
}

func (c *PublicationController) Commit() error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != PublicationPrepared || c.plan == nil || c.lifecycle == nil || c.lifecycle.AdmissionOpen() {
		return c.failLocked("invalid_commit_state")
	}
	if len(c.acknowledged) != len(c.expected) {
		missing := make([]string, 0)
		for componentID := range c.expected {
			if _, ok := c.acknowledged[componentID]; !ok {
				missing = append(missing, componentID)
			}
		}
		sort.Strings(missing)
		return c.failLocked("missing_acknowledgment:" + fmt.Sprint(missing))
	}
	c.state = PublicationCommitted
	return nil
}

func (c *PublicationController) Serve() error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != PublicationCommitted || c.plan == nil || c.lifecycle == nil || c.lifecycle.AdmissionOpen() {
		return c.failLocked("invalid_serving_state")
	}
	if err := c.lifecycle.Publish(); err != nil {
		return c.failLocked("admission_gate_failed")
	}
	c.state = PublicationServing
	return nil
}

func (c *PublicationController) AbortStartup() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == PublicationServing || c.state == PublicationFailed {
		return
	}
	_ = c.failLocked("startup_aborted")
}

func (c *PublicationController) ComponentLost(componentID string) bool {
	if c == nil || componentID == "" {
		return false
	}
	c.mu.Lock()
	if c.state != PublicationServing || c.lifecycle == nil {
		c.mu.Unlock()
		return false
	}
	c.state = PublicationFailed
	c.plan = nil
	c.expected = nil
	c.acknowledged = nil
	lifecycle := c.lifecycle
	c.mu.Unlock()
	return lifecycle.Fatal("published_component_lost")
}

func (c *PublicationController) State() PublicationState {
	if c == nil {
		return PublicationFailed
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *PublicationController) ExpectedComponents() map[string]string {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]string, len(c.expected))
	for componentID, digest := range c.expected {
		result[componentID] = digest
	}
	return result
}

func (c *PublicationController) Discovery() []extensions.DiscoveryProfile {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.plan == nil {
		return nil
	}
	return c.plan.Discovery()
}

func (c *PublicationController) Claims() []extensions.ClaimPublication {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.plan == nil {
		return nil
	}
	return c.plan.Claims()
}

func (c *PublicationController) ResolvedClaims() extensions.ResolvedClaimSet {
	if c == nil {
		return extensions.ResolvedClaimSet{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.plan == nil {
		return extensions.ResolvedClaimSet{}
	}
	return c.plan.ResolvedClaims()
}

func (c *PublicationController) Summary() (extensions.PublicationPlanSummary, bool) {
	if c == nil {
		return extensions.PublicationPlanSummary{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.plan == nil {
		return extensions.PublicationPlanSummary{}, false
	}
	return c.plan.Summary(), true
}

func (c *PublicationController) failLocked(reason string) error {
	wasServing := c.state == PublicationServing
	c.state = PublicationFailed
	c.plan = nil
	c.expected = nil
	c.acknowledged = nil
	if wasServing && c.lifecycle != nil {
		c.lifecycle.Fatal("published_component_lost")
	}
	return fmt.Errorf("extension_publication_failed: %s", reason)
}
