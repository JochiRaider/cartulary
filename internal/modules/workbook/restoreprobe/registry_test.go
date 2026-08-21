package restoreprobe_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/restoreprobe"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

type recordingQuery struct {
	calls        int
	incidentID   uuid.UUID
	viewSchemaID string
	meta         viewschema.QueryMeta
	rows         []map[string]any
	err          error
}

func (query *recordingQuery) QueryRows(
	_ context.Context,
	incidentID uuid.UUID,
	viewSchemaID string,
	meta viewschema.QueryMeta,
) ([]map[string]any, error) {
	query.calls++
	query.incidentID = incidentID
	query.viewSchemaID = viewSchemaID
	query.meta = meta
	return query.rows, query.err
}

func TestRestoreProbeRegistryExecutesExactOwnerRegistration(t *testing.T) {
	query := &recordingQuery{rows: []map[string]any{}}
	registry, err := restoreprobe.NewRegistry(query, timeline.RestoreWorkbookProbeRegistration())
	if err != nil {
		t.Fatalf("new restore probe registry: %v", err)
	}
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000000401")
	result, err := registry.ExecuteDefault(context.Background(), workbookprobe.BaseProfile, incidentID)
	if err != nil {
		t.Fatalf("execute default restore probe: %v", err)
	}
	if query.calls != 1 || query.incidentID != incidentID || query.viewSchemaID != "cartulary.view.timeline.v2" {
		t.Fatalf("query identity got calls=%d incident=%s view=%q", query.calls, query.incidentID, query.viewSchemaID)
	}
	wantSort := []viewschema.SortEntry{
		{FieldKey: "timeline.activity_sort_ts", Direction: "asc"},
		{FieldKey: "record_id", Direction: "asc"},
	}
	if len(query.meta.Filters) != 0 || query.meta.GroupBy != nil || !reflect.DeepEqual(query.meta.Sort, wantSort) {
		t.Fatalf("query realization got %#v", query.meta)
	}
	if result.RegistrationID != "timeline.base_restore_probe.v1" ||
		result.ViewSchemaID != "cartulary.view.timeline.v2" ||
		result.RowCount != 0 {
		t.Fatalf("execution result got %#v", result)
	}

	t.Run("nonzero rows are counted without changing registration identity", func(t *testing.T) {
		rowQuery := &recordingQuery{rows: []map[string]any{{"record_id": "first"}, {"record_id": "second"}}}
		rowRegistry, rowErr := restoreprobe.NewRegistry(rowQuery, timeline.RestoreWorkbookProbeRegistration())
		if rowErr != nil {
			t.Fatalf("new row-count registry: %v", rowErr)
		}
		rowResult, rowErr := rowRegistry.ExecuteDefault(context.Background(), workbookprobe.BaseProfile, incidentID)
		if rowErr != nil {
			t.Fatalf("execute row-count registry: %v", rowErr)
		}
		if rowResult.RegistrationID != "timeline.base_restore_probe.v1" ||
			rowResult.ViewSchemaID != "cartulary.view.timeline.v2" ||
			rowResult.RowCount != 2 || rowQuery.calls != 1 {
			t.Fatalf("row-count execution got result=%#v calls=%d", rowResult, rowQuery.calls)
		}
	})

	t.Run("unknown profile and nil incident fail before querying", func(t *testing.T) {
		guardedQuery := &recordingQuery{}
		guardedRegistry, guardedErr := restoreprobe.NewRegistry(guardedQuery, timeline.RestoreWorkbookProbeRegistration())
		if guardedErr != nil {
			t.Fatalf("new guarded registry: %v", guardedErr)
		}
		if _, guardedErr = guardedRegistry.ExecuteDefault(context.Background(), "unknown", incidentID); !errors.Is(guardedErr, restoreprobe.ErrExecutionFailed) {
			t.Fatalf("unknown profile error got %v", guardedErr)
		}
		if _, guardedErr = guardedRegistry.ExecuteDefault(context.Background(), workbookprobe.BaseProfile, uuid.Nil); !errors.Is(guardedErr, restoreprobe.ErrExecutionFailed) {
			t.Fatalf("nil incident error got %v", guardedErr)
		}
		if guardedQuery.calls != 0 {
			t.Fatalf("pre-query failures invoked projection query %d times", guardedQuery.calls)
		}
	})

	t.Run("query failure preserves registration and view identity", func(t *testing.T) {
		queryFailure := errors.New("projection query unavailable")
		failingQuery := &recordingQuery{err: queryFailure}
		failingRegistry, failingErr := restoreprobe.NewRegistry(failingQuery, timeline.RestoreWorkbookProbeRegistration())
		if failingErr != nil {
			t.Fatalf("new failing registry: %v", failingErr)
		}
		failedResult, failingErr := failingRegistry.ExecuteDefault(context.Background(), workbookprobe.BaseProfile, incidentID)
		if !errors.Is(failingErr, restoreprobe.ErrExecutionFailed) ||
			failedResult.RegistrationID != "timeline.base_restore_probe.v1" ||
			failedResult.ViewSchemaID != "cartulary.view.timeline.v2" ||
			failingQuery.calls != 1 {
			t.Fatalf("query failure got result=%#v calls=%d error=%v", failedResult, failingQuery.calls, failingErr)
		}
	})

	t.Run("nil registry fails without panic", func(t *testing.T) {
		var nilRegistry *restoreprobe.Registry
		if _, nilErr := nilRegistry.ExecuteDefault(context.Background(), workbookprobe.BaseProfile, incidentID); !errors.Is(nilErr, restoreprobe.ErrExecutionFailed) {
			t.Fatalf("nil registry error got %v", nilErr)
		}
	})
}

func TestRestoreProbeRegistryRejectsConflicts(t *testing.T) {
	query := &recordingQuery{}
	registration := timeline.RestoreWorkbookProbeRegistration()
	t.Run("nil projection query", func(t *testing.T) {
		_, err := restoreprobe.NewRegistry(nil, registration)
		if !errors.Is(err, restoreprobe.ErrInvalidRegistration) {
			t.Fatalf("nil projection query got %v", err)
		}
	})
	t.Run("typed nil projection query", func(t *testing.T) {
		var typedNil *recordingQuery
		_, err := restoreprobe.NewRegistry(typedNil, registration)
		if !errors.Is(err, restoreprobe.ErrInvalidRegistration) {
			t.Fatalf("typed nil projection query got %v", err)
		}
	})
	t.Run("empty registration set", func(t *testing.T) {
		_, err := restoreprobe.NewRegistry(query)
		if !errors.Is(err, restoreprobe.ErrInvalidRegistration) {
			t.Fatalf("empty registration set got %v", err)
		}
	})
	t.Run("invalid registration schema", func(t *testing.T) {
		invalid := registration
		invalid.SchemaID = "cartulary.restore_workbook_probe_registration.invalid"
		_, err := restoreprobe.NewRegistry(query, invalid)
		if !errors.Is(err, restoreprobe.ErrInvalidRegistration) {
			t.Fatalf("invalid registration schema got %v", err)
		}
	})
	t.Run("unresolved view schema", func(t *testing.T) {
		invalid := registration
		invalid.ViewSchemaID = "cartulary.view.unknown.v1"
		_, err := restoreprobe.NewRegistry(query, invalid)
		if !errors.Is(err, restoreprobe.ErrInvalidRegistration) {
			t.Fatalf("unresolved view schema got %v", err)
		}
	})
	t.Run("duplicate registration", func(t *testing.T) {
		_, err := restoreprobe.NewRegistry(query, registration, registration)
		if !errors.Is(err, restoreprobe.ErrRegistryConflict) {
			t.Fatalf("duplicate registration got %v", err)
		}
	})
	t.Run("multiple profile defaults", func(t *testing.T) {
		second := registration
		second.RegistrationID = "timeline.base_restore_probe.alternate.v1"
		_, err := restoreprobe.NewRegistry(query, registration, second)
		if !errors.Is(err, restoreprobe.ErrRegistryConflict) {
			t.Fatalf("multiple defaults got %v", err)
		}
	})
	t.Run("missing profile default", func(t *testing.T) {
		nonDefault := registration
		nonDefault.IsDefault = false
		_, err := restoreprobe.NewRegistry(query, nonDefault)
		if !errors.Is(err, restoreprobe.ErrRegistryConflict) {
			t.Fatalf("missing default got %v", err)
		}
	})
}
