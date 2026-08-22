package mentions

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestNewStoreRequiresLinkOperations(t *testing.T) {
	t.Helper()
	defer func() {
		value := recover()
		if value == nil || !strings.Contains(value.(string), "link operations are required") {
			t.Fatalf("NewStore panic = %#v, want missing link operations", value)
		}
	}()
	_ = NewStore(nil, nil, WithWorkbookProjection(noopProjectionWriter{}))
}

type noopProjectionWriter struct{}

func (noopProjectionWriter) RefreshHostTx(context.Context, pgx.Tx, uuid.UUID) error     { return nil }
func (noopProjectionWriter) RefreshIdentityTx(context.Context, pgx.Tx, uuid.UUID) error { return nil }
func (noopProjectionWriter) DeleteHostTx(context.Context, pgx.Tx, uuid.UUID) error      { return nil }
func (noopProjectionWriter) DeleteIdentityTx(context.Context, pgx.Tx, uuid.UUID) error  { return nil }
func (noopProjectionWriter) RebuildHostsTx(context.Context, pgx.Tx, uuid.UUID) error    { return nil }
func (noopProjectionWriter) RebuildIdentitiesTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}
