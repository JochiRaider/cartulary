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
	mu            sync.RWMutex
	lifecycle     *processlifecycle.Controller
	state         PublicationState
	preparedPlan  *extensions.PublicationPlan
	installedPlan *extensions.PublicationPlan
	expected      map[string]string
	acknowledged  map[string]struct{}
}

func NewPublicationController(lifecycle *processlifecycle.Controller) *PublicationController {
	return &PublicationController{
		lifecycle:    lifecycle,
		state:        PublicationUnpublished,
		expected:     map[string]string{},
		acknowledged: map[string]struct{}{},
	}
}

type publicationOrchestrator struct {
	*PublicationController
}

func preparePublicationOrchestrator(
	lifecycle *processlifecycle.Controller,
	plan extensions.PublicationPlan,
) (*publicationOrchestrator, error) {
	controller := NewPublicationController(lifecycle)
	if err := controller.Prepare(plan); err != nil {
		return nil, err
	}
	return &publicationOrchestrator{PublicationController: controller}, nil
}

func (orchestrator *publicationOrchestrator) HTTPProjections() publicationHTTPProjections {
	if orchestrator == nil {
		return publicationHTTPProjections{}
	}
	return publicationHTTPProjections{publication: orchestrator.PublicationController}
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
	c.preparedPlan = &plan
	c.installedPlan = nil
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
	if c.state != PublicationCommitted || c.installedPlan == nil {
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
	if c.state != PublicationPrepared || c.preparedPlan == nil || c.installedPlan != nil ||
		c.lifecycle == nil || c.lifecycle.AdmissionOpen() {
		return c.failLocked("invalid_commit_state")
	}
	installed := *c.preparedPlan
	c.installedPlan = &installed
	c.preparedPlan = nil
	c.state = PublicationCommitted
	return nil
}

func (c *PublicationController) missingAcknowledgmentsLocked() []string {
	if len(c.acknowledged) == len(c.expected) {
		return nil
	}
	missing := make([]string, 0)
	for componentID := range c.expected {
		if _, ok := c.acknowledged[componentID]; !ok {
			missing = append(missing, componentID)
		}
	}
	sort.Strings(missing)
	return missing
}

func (c *PublicationController) Serve() error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != PublicationCommitted || c.installedPlan == nil || c.lifecycle == nil || c.lifecycle.AdmissionOpen() {
		return c.failLocked("invalid_serving_state")
	}
	if missing := c.missingAcknowledgmentsLocked(); len(missing) > 0 {
		return c.failLocked("missing_acknowledgment:" + fmt.Sprint(missing))
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
	if c.lifecycle == nil || (c.state != PublicationPrepared && c.state != PublicationCommitted && c.state != PublicationServing) {
		c.mu.Unlock()
		return false
	}
	if _, expected := c.expected[componentID]; !expected {
		c.mu.Unlock()
		return false
	}
	wasServing := c.state == PublicationServing
	c.state = PublicationFailed
	c.preparedPlan = nil
	c.installedPlan = nil
	c.expected = nil
	c.acknowledged = nil
	lifecycle := c.lifecycle
	c.mu.Unlock()
	if wasServing {
		return lifecycle.Fatal("published_component_lost")
	}
	return true
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
	if c.installedPlan == nil {
		return nil
	}
	return c.installedPlan.Discovery()
}

func (c *PublicationController) Claims() []extensions.ClaimPublication {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.installedPlan == nil {
		return nil
	}
	return c.installedPlan.Claims()
}

func (c *PublicationController) Routes() []extensions.RouteDispatch {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.installedPlan == nil {
		return nil
	}
	return c.installedPlan.Routes()
}

func (c *PublicationController) Workspaces() []extensions.WorkspacePublication {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.installedPlan == nil {
		return nil
	}
	return c.installedPlan.Workspaces()
}

func (c *PublicationController) Workers() []extensions.WorkerPublication {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.installedPlan == nil {
		return nil
	}
	return c.installedPlan.Workers()
}

func (c *PublicationController) JobKindContracts() []extensions.JobKindContract {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.installedPlan == nil {
		return nil
	}
	return c.installedPlan.JobKindContracts()
}

func (c *PublicationController) Contributions() []extensions.ContributionPublication {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.installedPlan == nil {
		return nil
	}
	return c.installedPlan.Contributions()
}

func (c *PublicationController) ImplementationBindings() []extensions.ImplementationBindingPublication {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.installedPlan == nil {
		return nil
	}
	return c.installedPlan.ImplementationBindings()
}

func (c *PublicationController) ResolvedClaims() extensions.ResolvedClaimSet {
	if c == nil {
		return extensions.ResolvedClaimSet{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.installedPlan == nil {
		return extensions.ResolvedClaimSet{}
	}
	return c.installedPlan.ResolvedClaims()
}

func (c *PublicationController) Summary() (extensions.PublicationPlanSummary, bool) {
	if c == nil {
		return extensions.PublicationPlanSummary{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.installedPlan == nil {
		return extensions.PublicationPlanSummary{}, false
	}
	return c.installedPlan.Summary(), true
}

func (c *PublicationController) failLocked(reason string) error {
	wasServing := c.state == PublicationServing
	c.state = PublicationFailed
	c.preparedPlan = nil
	c.installedPlan = nil
	c.expected = nil
	c.acknowledged = nil
	if wasServing && c.lifecycle != nil {
		c.lifecycle.Fatal("published_component_lost")
	}
	return fmt.Errorf("extension_publication_failed: %s", reason)
}
