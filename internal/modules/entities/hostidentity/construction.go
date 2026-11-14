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
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

var (
	ErrInvalidCreateRequest       = errors.New("entities: invalid create request")
	ErrHostIdentityRecordNotFound = errors.New("entities: host/identity record not found")
)

type Store struct {
	*mutationCore
	pool              dbport.DB
	authStore         *authn.Store
	incidentAccess    *admission.Checker
	conflictTokens    conflicts.ConflictTokenCodec
	conflictFields    conflicts.FieldResolver
	conflictSnapshots map[string]conflicts.RevisionSnapshotProjector
	keepSaved         conflicts.IdempotencyPort
	projectionReader  projectionports.QueryReader
}

type mutationCore struct {
	revisionAppender *revisions.Appender
	ports            entityStorePorts
	publications     collaboration.RecordChangedAppender
}

type StoreDependencies struct {
	ConflictTokens         conflicts.ConflictTokenCodec
	ConflictFields         conflicts.FieldResolver
	Postgres               dbport.DB
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
		{name: "ConflictFields", value: dependencies.ConflictFields},
		{name: "Collaboration", value: dependencies.Collaboration},
	} {
		if isNilStoreDependency(dependency.value) {
			return nil, fmt.Errorf("compose Host/Identity store: %s is required", dependency.name)
		}
	}
	if reflect.ValueOf(dependencies.ConflictTokens).IsZero() {
		return nil, fmt.Errorf("compose Host/Identity store: ConflictTokens is required")
	}
	snapshots, err := newEntityConflictSnapshotProjectors()
	if err != nil {
		return nil, err
	}
	return &Store{
		conflictTokens: dependencies.ConflictTokens, conflictFields: dependencies.ConflictFields, conflictSnapshots: snapshots,
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
