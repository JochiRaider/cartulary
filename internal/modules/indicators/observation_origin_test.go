package indicators

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestObservationOriginRegistryAndProducerMapping(t *testing.T) {
	t.Parallel()
	valid := []string{
		"manual_entry",
		"clipboard_paste",
		"csv_import",
		"xlsx_import",
		"api_import",
		"extraction",
		"system",
	}
	for _, raw := range valid {
		parsed, err := ParseObservationOrigin(raw)
		if err != nil || parsed.String() != raw {
			t.Fatalf("parse exact origin %q = %q, %v", raw, parsed, err)
		}
	}
	for _, raw := range []string{
		"",
		" manual_entry",
		"manual_entry ",
		"Manual_Entry",
		"interactive_cell",
		"auto_extract",
		"domain",
		"extension:manual_entry",
		"unknown",
	} {
		if _, err := ParseObservationOrigin(raw); !errors.Is(err, ErrInvalidObservationOrigin) {
			t.Fatalf("parse invalid origin %q = %v", raw, err)
		}
	}

	producers := []struct {
		context ObservationProducerContext
		want    string
	}{
		{ManualEntryObservationProducer(), "manual_entry"},
		{ClipboardPasteObservationProducer(), "clipboard_paste"},
		{CSVImportObservationProducer(), "csv_import"},
		{XLSXImportObservationProducer(), "xlsx_import"},
		{APIImportObservationProducer(), "api_import"},
		{ExtractionObservationProducer(), "extraction"},
	}
	for _, producer := range producers {
		origin, err := producer.context.originForWrite()
		if err != nil || origin.String() != producer.want {
			t.Fatalf("producer origin = %q, %v; want %q", origin, err, producer.want)
		}
	}
	if _, err := (ObservationProducerContext{}).originForWrite(); !errors.Is(err, ErrInvalidObservationOrigin) {
		t.Fatalf("zero producer context = %v", err)
	}
}

func TestObservationProducerSurfaceHasNoSystemConstructor(t *testing.T) {
	t.Parallel()
	parsed, err := parser.ParseFile(token.NewFileSet(), "observation_origin.go", nil, 0)
	if err != nil {
		t.Fatalf("parse observation producer surface: %v", err)
	}
	constructors := []string{}
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || !function.Name.IsExported() || function.Type.Results == nil {
			continue
		}
		for _, result := range function.Type.Results.List {
			identifier, ok := result.Type.(*ast.Ident)
			if ok && identifier.Name == "ObservationProducerContext" {
				constructors = append(constructors, function.Name.Name)
			}
		}
	}
	slices.Sort(constructors)
	want := []string{
		"APIImportObservationProducer",
		"CSVImportObservationProducer",
		"ClipboardPasteObservationProducer",
		"ExtractionObservationProducer",
		"ManualEntryObservationProducer",
		"XLSXImportObservationProducer",
	}
	if !reflect.DeepEqual(constructors, want) {
		t.Fatalf("ordinary producer constructors = %v, want %v", constructors, want)
	}
}

func TestInvalidObservationProducerFailsBeforeTransaction(t *testing.T) {
	t.Parallel()
	db := &rejectOriginBeginDB{}
	store := &Store{pool: db}
	_, _, err := store.CreateIndicatorObservation(context.Background(), authn.UserRecord{}, IndicatorObservationCreateParams{
		IncidentID:     uuid.New(),
		SourceRecordID: uuid.New(),
		SourceFieldKey: "timeline.raw_activity_text",
		OriginLocator:  "origin-prewrite-test",
		ObservedText:   "192.0.2.10",
	})
	if !errors.Is(err, ErrInvalidObservationOrigin) {
		t.Fatalf("invalid producer error = %v", err)
	}
	if db.began {
		t.Fatal("invalid observation producer began a transaction")
	}
}

type rejectOriginBeginDB struct {
	postgres.DB
	began bool
}

func (db *rejectOriginBeginDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	db.began = true
	return nil, errors.New("transaction must not start")
}
