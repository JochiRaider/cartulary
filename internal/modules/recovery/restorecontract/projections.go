package restorecontract

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const ProviderRegistryRefCodeBacked = "internal/modules/projections/provider_registry.go#projection_provider_descriptor.v1"

type ProjectionRebuildScope string

const (
	ProjectionRebuildScopeAllActiveProviders ProjectionRebuildScope = "all_active_providers"
)

type ProjectionRebuildStatus string

const (
	ProjectionRebuildStatusSucceeded     ProjectionRebuildStatus = "succeeded"
	ProjectionRebuildStatusNotApplicable ProjectionRebuildStatus = "not_applicable"
	ProjectionRebuildStatusFailed        ProjectionRebuildStatus = "failed"
)

type ProjectionReadinessOutcome string

const (
	ProjectionReadinessReady         ProjectionReadinessOutcome = "ready"
	ProjectionReadinessNotApplicable ProjectionReadinessOutcome = "not_applicable"
	ProjectionReadinessIncomplete    ProjectionReadinessOutcome = "incomplete"
	ProjectionReadinessDegraded      ProjectionReadinessOutcome = "degraded"
)

type ProjectionProviderResultStatus string

const (
	ProjectionProviderResultSucceeded               ProjectionProviderResultStatus = "succeeded"
	ProjectionProviderResultSkippedNonparticipating ProjectionProviderResultStatus = "skipped_nonparticipating"
	ProjectionProviderResultFailed                  ProjectionProviderResultStatus = "failed"
)

type ProjectionRebuildRequest struct {
	RestoreOperationID     uuid.UUID
	RestoredSourceStateRef string
	RebuildScope           ProjectionRebuildScope
	ProviderRegistryRef    string
}

func (request ProjectionRebuildRequest) Validate() error {
	if request.RestoreOperationID == uuid.Nil {
		return fmt.Errorf("restore_operation_id is required")
	}
	if strings.TrimSpace(request.RestoredSourceStateRef) == "" {
		return fmt.Errorf("restored_source_state_ref is required")
	}
	if request.RebuildScope != ProjectionRebuildScopeAllActiveProviders {
		return fmt.Errorf("unsupported rebuild_scope %q", request.RebuildScope)
	}
	if strings.TrimSpace(request.ProviderRegistryRef) == "" {
		return fmt.Errorf("provider_registry_ref is required")
	}
	return nil
}

type ProjectionRebuildResult struct {
	RestoreOperationID uuid.UUID
	Status             ProjectionRebuildStatus
	ReadinessOutcome   ProjectionReadinessOutcome
	ProviderResults    []ProjectionProviderResult
	Warnings           []ProjectionRebuildMessage
	Errors             []ProjectionRebuildMessage
}

func (result ProjectionRebuildResult) ReadinessSatisfied() bool {
	return (result.Status == ProjectionRebuildStatusSucceeded && result.ReadinessOutcome == ProjectionReadinessReady) ||
		(result.Status == ProjectionRebuildStatusNotApplicable && result.ReadinessOutcome == ProjectionReadinessNotApplicable)
}

type ProjectionProviderResult struct {
	ProviderKey      string
	Status           ProjectionProviderResultStatus
	IncidentCount    int
	RowCountsByTable map[string]int64
	Warnings         []string
	Error            string
}

type ProjectionRebuildMessage struct {
	Code        string
	Message     string
	ProviderKey string
}

type ProjectionRebuilder interface {
	RebuildRestoreProjections(ctx context.Context, request ProjectionRebuildRequest) (ProjectionRebuildResult, error)
}
