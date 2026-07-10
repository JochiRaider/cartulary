package revisions

import (
	"errors"
	"testing"

	recorddeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"
)

func TestDeleteRestoreProviderCatalogFailsClosed(t *testing.T) {
	t.Parallel()
	provider := recorddeleterestore.TableProvider{}
	if _, err := NewDeleteRestoreProviderCatalog([]string{"host", "party"},
		DeleteRestoreProviderRegistration{RecordType: "host", Provider: provider},
	); !errors.Is(err, ErrMissingDeleteRestoreProvider) {
		t.Fatalf("missing provider error = %v", err)
	}
	if _, err := NewDeleteRestoreProviderCatalog([]string{"host"},
		DeleteRestoreProviderRegistration{RecordType: "host", Provider: provider},
		DeleteRestoreProviderRegistration{RecordType: "host", Provider: provider},
	); !errors.Is(err, ErrDuplicateDeleteRestoreProvider) {
		t.Fatalf("duplicate provider error = %v", err)
	}
	if _, err := NewDeleteRestoreProviderCatalog([]string{"host"},
		DeleteRestoreProviderRegistration{RecordType: "party", Provider: provider},
	); !errors.Is(err, ErrUnexpectedDeleteRestoreProvider) {
		t.Fatalf("unexpected provider error = %v", err)
	}
	var typedNil *recorddeleterestore.TableProvider
	if _, err := NewDeleteRestoreProviderCatalog([]string{"host"},
		DeleteRestoreProviderRegistration{RecordType: "host", Provider: typedNil},
	); !errors.Is(err, ErrMissingDeleteRestoreProvider) {
		t.Fatalf("typed nil provider error = %v", err)
	}
}
