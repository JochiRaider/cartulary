package ownerfacade

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestImportScalarValueClosedUnion(t *testing.T) {
	t.Parallel()
	timestamp := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	id := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	token := ImportCollectionToken{RawText: " Raw ", NormalizedText: "Raw"}

	values := []struct {
		name string
		kind ImportScalarKind
		got  ImportScalarValue
	}{
		{name: "null", kind: ImportScalarNull, got: NewNullImportScalar()},
		{name: "text", kind: ImportScalarText, got: NewTextImportScalar("text")},
		{name: "timestamp", kind: ImportScalarTimestamp, got: NewTimestampImportScalar(timestamp)},
		{name: "uuid", kind: ImportScalarUUID, got: NewUUIDImportScalar(id)},
		{name: "number", kind: ImportScalarNumber, got: NewNumberImportScalar(42)},
		{name: "bool", kind: ImportScalarBool, got: NewBoolImportScalar(true)},
		{name: "collection token", kind: ImportScalarCollectionToken, got: NewCollectionTokenImportScalar(token)},
	}
	for _, test := range values {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if !test.got.IsValid() || test.got.Kind() != test.kind {
				t.Fatalf("scalar = %#v, kind = %q", test.got, test.got.Kind())
			}
		})
	}

	if value, ok := values[1].got.Text(); !ok || value != "text" {
		t.Fatalf("text accessor = %q, %t", value, ok)
	}
	if value, ok := values[2].got.Timestamp(); !ok || !value.Equal(timestamp) {
		t.Fatalf("timestamp accessor = %s, %t", value, ok)
	}
	if value, ok := values[3].got.UUID(); !ok || value != id {
		t.Fatalf("UUID accessor = %s, %t", value, ok)
	}
	if value, ok := values[4].got.Number(); !ok || value != 42 {
		t.Fatalf("number accessor = %d, %t", value, ok)
	}
	if value, ok := values[5].got.Bool(); !ok || !value {
		t.Fatalf("bool accessor = %t, %t", value, ok)
	}
	if value, ok := values[6].got.CollectionToken(); !ok || value != token {
		t.Fatalf("collection accessor = %#v, %t", value, ok)
	}
	if _, ok := values[0].got.Text(); ok {
		t.Fatal("null scalar exposed a text payload")
	}
	for name, invalid := range map[string]ImportScalarValue{
		"zero":          {},
		"unknown kind":  {kind: "future", valid: true},
		"unsealed kind": {kind: ImportScalarText},
	} {
		if invalid.IsValid() {
			t.Fatalf("%s scalar unexpectedly valid", name)
		}
	}
}

func TestIndexImportFieldValuesRejectsInvalidStructure(t *testing.T) {
	t.Parallel()
	valid := ImportFieldValue{FieldKey: "indicator.display_value", NormalizedValue: NewTextImportScalar("value")}
	indexed, err := IndexImportFieldValues([]ImportFieldValue{valid})
	if err != nil || len(indexed) != 1 {
		t.Fatalf("index valid field values = %#v, %v", indexed, err)
	}
	for name, fields := range map[string][]ImportFieldValue{
		"empty key":   {{NormalizedValue: NewTextImportScalar("value")}},
		"blank key":   {{FieldKey: "  ", NormalizedValue: NewTextImportScalar("value")}},
		"zero scalar": {{FieldKey: "indicator.display_value"}},
		"duplicate":   {valid, valid},
	} {
		if indexed, err := IndexImportFieldValues(fields); !errors.Is(err, errInvalidImportFieldValues) || indexed != nil {
			t.Fatalf("%s = %#v, %v", name, indexed, err)
		}
	}
}

func TestBoundImportOwnerCreateFacadeRejectsInvalidFieldsBeforeOwner(t *testing.T) {
	t.Parallel()
	called := false
	facade, err := NewImportOwnerCreateFacade(
		ImportOwnerCreateBinding{
			TargetViewSchemaID: "cartulary.view.indicators.v1",
			FacadeID:           "indicators.import_create",
		},
		func(context.Context, pgx.Tx, ImportOwnerCreateCommand) (ImportOwnerCreateResponse, error) {
			called = true
			return ImportOwnerCreateResponse{}, nil
		},
	)
	if err != nil {
		t.Fatalf("construct owner facade: %v", err)
	}
	_, err = facade.CreateImportRowTx(context.Background(), nil, ImportOwnerCreateCommand{
		Request: ImportOwnerCreateRequest{
			TargetViewSchemaID: "cartulary.view.indicators.v1",
			FieldValues: []ImportFieldValue{
				{FieldKey: "indicator.display_value", NormalizedValue: NewTextImportScalar("first")},
				{FieldKey: "indicator.display_value", NormalizedValue: NewTextImportScalar("second")},
			},
		},
	})
	if !errors.Is(err, errInvalidImportFieldValues) || called {
		t.Fatalf("invalid owner dispatch = called:%t error:%v", called, err)
	}
	if detail, ok := ImportOwnerCreateErrorDetail(err); ok || detail != nil {
		t.Fatalf("structural error exposed owner detail: %#v", detail)
	}
}
