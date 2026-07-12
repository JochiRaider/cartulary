package migrationtest

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestApplyThroughRejectsNonPositiveVersionBeforeDatabaseAccess(t *testing.T) {
	if _, err := ApplyThrough(context.Background(), nil, postgres.MigrationSource{}, 0); err == nil {
		t.Fatal("non-positive migration version was accepted")
	}
}
