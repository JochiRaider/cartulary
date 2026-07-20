package httpapi

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const (
	ReadinessSchemaID = "cartulary.readiness.v1"

	ReadinessStatusStartingDependencyProbe = "starting_dependency_probe"
	ReadinessStatusReady                   = "ready"
	ReadinessStatusDegradedDependency      = "degraded_dependency"
	ReadinessStatusRecoveringDependency    = "recovering_dependency"
	ReadinessStatusFatalIntegrityFailure   = "fatal_integrity_failure"

	ReadinessReasonReady                 = "ready"
	ReadinessReasonDependencyUnavailable = "dependency_unavailable"
	ReadinessReasonDeadlineExceeded      = "deadline_exceeded"
)

const readinessCheckTimeout = 2 * time.Second

type ReadinessChecker interface {
	CheckReadiness(ctx context.Context) ReadinessState
}

type ReadinessCheckFunc func(ctx context.Context) ReadinessState

func (fn ReadinessCheckFunc) CheckReadiness(ctx context.Context) ReadinessState {
	if fn == nil {
		return ReadyReadinessState()
	}
	return fn(ctx)
}

type ReadinessState struct {
	SchemaID     string                `json:"schema_id"`
	Status       string                `json:"status"`
	Dependencies []ReadinessDependency `json:"dependencies,omitempty"`
}

type ReadinessDependency struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	ReasonCode string `json:"reason_code,omitempty"`
}

type postgresReadinessPinger interface {
	Ping(ctx context.Context) error
}

type dependencyReadinessChecker struct {
	postgres    postgresReadinessPinger
	objectStore objectstore.Store
}

func ReadyReadinessState() ReadinessState {
	return ReadinessState{
		SchemaID: ReadinessSchemaID,
		Status:   ReadinessStatusReady,
	}
}

func NewDependencyReadinessChecker(postgres postgresReadinessPinger, objectStore objectstore.Store) ReadinessChecker {
	if isNilReadinessDependency(postgres) {
		postgres = nil
	}
	if isNilReadinessDependency(objectStore) {
		objectStore = nil
	}
	return dependencyReadinessChecker{
		postgres:    postgres,
		objectStore: objectStore,
	}
}

func (checker dependencyReadinessChecker) CheckReadiness(ctx context.Context) ReadinessState {
	if checker.postgres == nil && checker.objectStore == nil {
		return ReadyReadinessState()
	}

	checkCtx, cancel := context.WithTimeout(ctx, readinessCheckTimeout)
	defer cancel()

	dependencies := make([]ReadinessDependency, 0, 2)
	ready := true
	if checker.postgres != nil {
		dependency := ReadinessDependency{
			Name:       "postgres",
			Status:     ReadinessStatusReady,
			ReasonCode: ReadinessReasonReady,
		}
		if err := checker.postgres.Ping(checkCtx); err != nil {
			ready = false
			dependency.Status = ReadinessStatusDegradedDependency
			dependency.ReasonCode = readinessReasonCode(err)
		}
		dependencies = append(dependencies, dependency)
	}
	if checker.objectStore != nil {
		dependency := ReadinessDependency{
			Name:       "object_store",
			Status:     ReadinessStatusReady,
			ReasonCode: ReadinessReasonReady,
		}
		if _, err := checker.objectStore.ListObjects(checkCtx, ".cartulary/readiness/"); err != nil {
			ready = false
			dependency.Status = ReadinessStatusDegradedDependency
			dependency.ReasonCode = readinessReasonCode(err)
		}
		dependencies = append(dependencies, dependency)
	}

	status := ReadinessStatusReady
	if !ready {
		status = ReadinessStatusDegradedDependency
	}
	return ReadinessState{
		SchemaID:     ReadinessSchemaID,
		Status:       status,
		Dependencies: dependencies,
	}
}

func isNilReadinessDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func readinessReasonCode(err error) string {
	if err == nil {
		return ReadinessReasonReady
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ReadinessReasonDeadlineExceeded
	}
	if adapterErr, ok := objectstore.AsAdapterError(err); ok {
		return string(adapterErr.Reason)
	}
	return ReadinessReasonDependencyUnavailable
}
