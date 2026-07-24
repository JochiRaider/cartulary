package server

import (
	"context"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestStagedCleanupReadiness_Unit_DegradesAndRecoversServingReadiness(t *testing.T) {
	health := stagedobjects.NewHealth()
	checker := httpapi.NewDependencyReadinessChecker(nil, nil, stagedCleanupReadinessProbe{health: health})
	state := checker.CheckReadiness(context.Background())
	if state.Status != httpapi.ReadinessStatusReady ||
		len(state.Dependencies) != 1 ||
		state.Dependencies[0].Name != "staged_object_cleanup" {
		t.Fatalf("initial readiness = %#v", state)
	}

	health.Unavailable(stagedobjects.NewFailure(stagedobjects.FailureDependency, "object_store_unavailable", nil))
	state = checker.CheckReadiness(context.Background())
	if state.Status != httpapi.ReadinessStatusDegradedDependency ||
		len(state.Dependencies) != 1 ||
		state.Dependencies[0].Status != httpapi.ReadinessStatusDegradedDependency ||
		state.Dependencies[0].ReasonCode != httpapi.ReadinessReasonDependencyUnavailable {
		t.Fatalf("degraded readiness = %#v", state)
	}

	health.Available()
	state = checker.CheckReadiness(context.Background())
	if state.Status != httpapi.ReadinessStatusReady {
		t.Fatalf("recovered readiness = %#v", state)
	}
}

func TestStagedCleanupFatal_Unit_UsesIntegrityExitReason(t *testing.T) {
	err := &stagedobjects.FatalIntegrityError{Cause: errors.New("private contradiction detail")}
	if err.FatalReasonCode() != "staged_object_publication_mismatch" {
		t.Fatalf("fatal reason = %q", err.FatalReasonCode())
	}
}
