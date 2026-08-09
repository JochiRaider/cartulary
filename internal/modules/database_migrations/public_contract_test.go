package database_migrations

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPublicSurfaceAndLedgerReaderCapability(t *testing.T) {
	readerType := reflect.TypeOf((*LedgerReader)(nil)).Elem()
	if readerType.NumMethod() != 2 {
		t.Fatalf("LedgerReader method count = %d, want 2", readerType.NumMethod())
	}
	for index, name := range []string{"Query", "QueryRow"} {
		if got := readerType.Method(index).Name; got != name {
			t.Fatalf("LedgerReader method %d = %q, want %q", index, got, name)
		}
	}

	sourceType := reflect.TypeOf(Source{})
	if sourceType.NumField() == 0 {
		t.Fatal("Source must retain private catalog state")
	}
	for index := 0; index < sourceType.NumField(); index++ {
		if sourceType.Field(index).PkgPath == "" {
			t.Fatalf("Source field %q must be unexported", sourceType.Field(index).Name)
		}
	}

	failureType := reflect.TypeOf((*MigrationFailure)(nil)).Elem()
	if failureType.NumMethod() != 2 || failureType.Method(0).Name != "Error" || failureType.Method(1).Name != "ReasonCode" {
		t.Fatalf("unexpected MigrationFailure surface: %v", failureType)
	}
	reporterType := reflect.TypeOf((*RemediationReporter)(nil)).Elem()
	if reporterType.NumMethod() != 3 || reporterType.Method(2).Name != "RemediationReportJSON" {
		t.Fatalf("unexpected RemediationReporter surface: %v", reporterType)
	}
}

var _ LedgerReader = (*pgxpool.Pool)(nil)
