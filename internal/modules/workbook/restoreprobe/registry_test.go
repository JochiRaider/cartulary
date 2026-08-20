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
)

type recordingQuery struct {
	calls        int
	incidentID   uuid.UUID
	viewSchemaID string
	meta         viewschema.QueryMeta
	rows         []map[string]any
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
	return query.rows, nil
}

func TestRestoreProbeRegistryExecutesExactOwnerRegistration(t *testing.T) {
	query := &recordingQuery{rows: []map[string]any{}}
	registry, err := restoreprobe.NewRegistry(query, timeline.RestoreWorkbookProbeRegistration())
	if err != nil {
		t.Fatalf("new restore probe registry: %v", err)
	}
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000000401")
	result, err := registry.ExecuteDefault(context.Background(), restoreprobe.BaseProfile, incidentID)
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
}

func TestRestoreProbeRegistryRejectsConflicts(t *testing.T) {
	query := &recordingQuery{}
	registration := timeline.RestoreWorkbookProbeRegistration()
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
