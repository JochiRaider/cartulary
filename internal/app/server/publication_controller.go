package server

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

type publicationState string

const (
	publicationUnpublished publicationState = "unpublished"
	publicationPrepared    publicationState = "prepared"
	publicationCommitted   publicationState = "committed"
	publicationServing     publicationState = "serving"
	publicationFailed      publicationState = "failed"
)

type publicationController struct {
	mu            sync.RWMutex
	lifecycle     *processlifecycle.Controller
	state         publicationState
	preparedPlan  *extensions.PublicationPlan
	installedPlan *extensions.PublicationPlan
	expected      map[string]string
	acknowledged  map[string]struct{}
}

func newPublicationController(lifecycle *processlifecycle.Controller) *publicationController {
	return &publicationController{
		lifecycle:    lifecycle,
		state:        publicationUnpublished,
		expected:     map[string]string{},
		acknowledged: map[string]struct{}{},
	}
}

type publicationOrchestrator struct {
	controller *publicationController
}

func preparePublicationOrchestrator(
	lifecycle *processlifecycle.Controller,
	plan extensions.PublicationPlan,
) (*publicationOrchestrator, error) {
	controller := newPublicationController(lifecycle)
	if err := controller.prepare(plan); err != nil {
		return nil, err
	}
	return &publicationOrchestrator{controller: controller}, nil
}

func (orchestrator *publicationOrchestrator) httpProjections() publicationHTTPProjections {
	if orchestrator == nil {
		return publicationHTTPProjections{}
	}
	return publicationHTTPProjections{publication: orchestrator.controller}
}

func (c *publicationController) prepare(plan extensions.PublicationPlan) error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != publicationUnpublished || c.lifecycle == nil || c.lifecycle.AdmissionOpen() {
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
	c.state = publicationPrepared
	return nil
}

func (c *publicationController) acknowledge(componentID string, planSHA256 string, preparationErr error) error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != publicationCommitted || c.installedPlan == nil {
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

func (c *publicationController) commit() error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != publicationPrepared || c.preparedPlan == nil || c.installedPlan != nil ||
		c.lifecycle == nil || c.lifecycle.AdmissionOpen() {
		return c.failLocked("invalid_commit_state")
	}
	installed := *c.preparedPlan
	c.installedPlan = &installed
	c.preparedPlan = nil
	c.state = publicationCommitted
	return nil
}

func (c *publicationController) missingAcknowledgmentsLocked() []string {
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

func (c *publicationController) serve() error {
	if c == nil {
		return errors.New("extension_publication_failed")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != publicationCommitted || c.installedPlan == nil || c.lifecycle == nil || c.lifecycle.AdmissionOpen() {
		return c.failLocked("invalid_serving_state")
	}
	if missing := c.missingAcknowledgmentsLocked(); len(missing) > 0 {
		return c.failLocked("missing_acknowledgment:" + fmt.Sprint(missing))
	}
	if err := c.lifecycle.Publish(); err != nil {
		return c.failLocked("admission_gate_failed")
	}
	c.state = publicationServing
	return nil
}

func (c *publicationController) abortStartup() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == publicationServing || c.state == publicationFailed {
		return
	}
	_ = c.failLocked("startup_aborted")
}

func (c *publicationController) componentLost(componentID string) bool {
	if c == nil || componentID == "" {
		return false
	}
	c.mu.Lock()
	if c.lifecycle == nil || (c.state != publicationPrepared && c.state != publicationCommitted && c.state != publicationServing) {
		c.mu.Unlock()
		return false
	}
	if _, expected := c.expected[componentID]; !expected {
		c.mu.Unlock()
		return false
	}
	wasServing := c.state == publicationServing
	c.state = publicationFailed
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

func (c *publicationController) currentState() publicationState {
	if c == nil {
		return publicationFailed
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *publicationController) expectedComponents() map[string]string {
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

func (c *publicationController) discovery() []extensions.DiscoveryProfile {
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

func (c *publicationController) claims() []extensions.ClaimPublication {
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

func (c *publicationController) routes() []extensions.RouteDispatch {
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

func (c *publicationController) workspaces() []extensions.WorkspacePublication {
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

func (c *publicationController) workers() []extensions.WorkerPublication {
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

func (c *publicationController) jobKindContracts() []extensions.JobKindContract {
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

func (c *publicationController) contributions() []extensions.ContributionPublication {
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

func (c *publicationController) implementationBindings() []extensions.ImplementationBindingPublication {
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

func (c *publicationController) summary() (extensions.PublicationPlanSummary, bool) {
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

func (c *publicationController) failLocked(reason string) error {
	wasServing := c.state == publicationServing
	c.state = publicationFailed
	c.preparedPlan = nil
	c.installedPlan = nil
	c.expected = nil
	c.acknowledged = nil
	if wasServing && c.lifecycle != nil {
		c.lifecycle.Fatal("published_component_lost")
	}
	return fmt.Errorf("extension_publication_failed: %s", reason)
}
