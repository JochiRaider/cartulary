package revisions

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type scopeEnvelopeReader struct {
	commandServiceTestEnvelopes
	values map[uuid.UUID]RecordEnvelope
}

func (r scopeEnvelopeReader) LoadEnvelopeTx(_ context.Context, _ pgx.Tx, id uuid.UUID, _ bool) (RecordEnvelope, error) {
	value, ok := r.values[id]
	if !ok {
		return RecordEnvelope{}, ErrEnvelopeNotFound
	}
	return value, nil
}

type scopeDependencyReader func([]uuid.UUID) ([]uuid.UUID, error)

func (r scopeDependencyReader) ProjectionDependentsTx(_ context.Context, _ pgx.Tx, _ uuid.UUID, changed []uuid.UUID) ([]uuid.UUID, error) {
	return r(changed)
}

func TestProjectionScopeIsDirectAndIncidentBound(t *testing.T) {
	incident, other := uuid.New(), uuid.New()
	changed := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	dependent := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	catalog := validCommandServiceDependencies(t).DeleteRestoreSources
	newStore := func() *commandStore {
		return &commandStore{
			deleteRestoreSources: catalog,
			envelopes: scopeEnvelopeReader{values: map[uuid.UUID]RecordEnvelope{
				changed:   {RecordID: changed, IncidentID: incident, RecordType: "host"},
				dependent: {RecordID: dependent, IncidentID: incident, RecordType: "timeline_event"},
			}},
		}
	}
	store := newStore()
	calls := 0
	reader := scopeDependencyReader(func(input []uuid.UUID) ([]uuid.UUID, error) {
		calls++
		if !slices.Equal(input, []uuid.UUID{changed}) {
			t.Fatal("dependency discovery expanded recursively or retained duplicate input")
		}
		return []uuid.UUID{dependent, changed, dependent}, nil
	})
	store.projectionDependencies = []ProjectionDependencyReader{reader, reader}
	scope, err := store.projectionScopeTx(context.Background(), nil, incident, []uuid.UUID{changed, changed})
	if err != nil || calls != 2 || !slices.Equal(scope, []uuid.UUID{changed, dependent}) {
		t.Fatal("direct projection scope did not close exactly", scope, calls, err)
	}
	for _, problem := range []string{"cross incident", "missing envelope", "unsupported owner", "empty identity", "reader failure", "empty mutation"} {
		t.Run(problem, func(t *testing.T) {
			store := newStore()
			values := store.envelopes.(scopeEnvelopeReader).values
			ids := []uuid.UUID{changed}
			readerErr := errors.New("owner unavailable")
			store.projectionDependencies = []ProjectionDependencyReader{scopeDependencyReader(func([]uuid.UUID) ([]uuid.UUID, error) {
				if problem == "reader failure" {
					return nil, readerErr
				}
				if problem == "empty identity" {
					return []uuid.UUID{uuid.Nil}, nil
				}
				return []uuid.UUID{dependent}, nil
			})}
			switch problem {
			case "cross incident":
				value := values[dependent]
				value.IncidentID = other
				values[dependent] = value
			case "missing envelope":
				delete(values, dependent)
			case "unsupported owner":
				value := values[dependent]
				value.RecordType = "unadmitted"
				values[dependent] = value
			case "empty mutation":
				ids = nil
			}
			if scope, err := store.projectionScopeTx(context.Background(), nil, incident, ids); err == nil || scope != nil {
				t.Fatal("invalid dependency exposed a usable partial scope")
			}
		})
	}
}
