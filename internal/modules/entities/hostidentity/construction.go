package hostidentity

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var (
	ErrInvalidCreateRequest       = errors.New("entities: invalid create request")
	ErrHostIdentityRecordNotFound = errors.New("entities: host/identity record not found")
)

type Store struct {
	*mutationCore
	pool             postgres.DB
	authStore        *authn.Store
	incidentAccess   *admission.Checker
	keepSaved        conflicts.IdempotencyPort
	projectionReader projectionports.QueryReader
}

type mutationCore struct {
	revisionAppender *revisions.Appender
	ports            entityStorePorts
	publications     collaboration.RecordChangedAppender
}

type StoreDependencies struct {
	Postgres               postgres.DB
	Revisions              *revisions.Appender
	ProjectionMutationRows projectionports.MutationRows
	ProjectionQueryReader  projectionports.QueryReader
	KeepSavedIdempotency   conflicts.IdempotencyPort
	Collaboration          collaboration.RecordChangedAppender
}

func NewStore(dependencies StoreDependencies) (*Store, error) {
	for _, dependency := range []struct {
		name  string
		value any
	}{
		{name: "Postgres", value: dependencies.Postgres},
		{name: "Revisions", value: dependencies.Revisions},
		{name: "ProjectionMutationRows", value: dependencies.ProjectionMutationRows},
		{name: "ProjectionQueryReader", value: dependencies.ProjectionQueryReader},
		{name: "KeepSavedIdempotency", value: dependencies.KeepSavedIdempotency},
		{name: "Collaboration", value: dependencies.Collaboration},
	} {
		if isNilStoreDependency(dependency.value) {
			return nil, fmt.Errorf("compose Host/Identity store: %s is required", dependency.name)
		}
	}
	return &Store{
		mutationCore:     newMutationCore(dependencies.Revisions, dependencies.ProjectionMutationRows, dependencies.Collaboration),
		pool:             dependencies.Postgres,
		authStore:        authn.NewStore(dependencies.Postgres),
		incidentAccess:   admission.NewChecker(dependencies.Postgres),
		keepSaved:        dependencies.KeepSavedIdempotency,
		projectionReader: dependencies.ProjectionQueryReader,
	}, nil
}

func newMutationCore(appender *revisions.Appender, projectionRows projectionports.MutationRows, publications collaboration.RecordChangedAppender) *mutationCore {
	return &mutationCore{
		revisionAppender: appender,
		ports:            newEntityStorePorts(appender, projectionRows),
		publications:     publications,
	}
}

func isNilStoreDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
