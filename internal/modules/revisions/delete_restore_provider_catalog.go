package revisions

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	recorddeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"
)

var (
	ErrDuplicateDeleteRestoreProvider  = errors.New("revisions: duplicate delete/restore provider")
	ErrMissingDeleteRestoreProvider    = errors.New("revisions: missing delete/restore provider")
	ErrUnexpectedDeleteRestoreProvider = errors.New("revisions: unexpected delete/restore provider")
)

type DeleteRestoreProviderRegistration struct {
	RecordType string
	Provider   recorddeleterestore.SourceProvider
}

type DeleteRestoreProviderCatalog struct {
	providers map[string]recorddeleterestore.SourceProvider
}

func NewDeleteRestoreProviderCatalog(requiredRecordTypes []string, registrations ...DeleteRestoreProviderRegistration) (*DeleteRestoreProviderCatalog, error) {
	required := make(map[string]struct{}, len(requiredRecordTypes))
	for _, recordType := range requiredRecordTypes {
		if recordType == "" {
			return nil, fmt.Errorf("%w: empty record type", ErrMissingDeleteRestoreProvider)
		}
		required[recordType] = struct{}{}
	}
	catalog := &DeleteRestoreProviderCatalog{providers: make(map[string]recorddeleterestore.SourceProvider, len(registrations))}
	for _, registration := range registrations {
		if registration.RecordType == "" || nilDeleteRestoreProvider(registration.Provider) {
			return nil, fmt.Errorf("%w: %q", ErrMissingDeleteRestoreProvider, registration.RecordType)
		}
		if _, allowed := required[registration.RecordType]; !allowed {
			return nil, fmt.Errorf("%w: %s", ErrUnexpectedDeleteRestoreProvider, registration.RecordType)
		}
		if _, exists := catalog.providers[registration.RecordType]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateDeleteRestoreProvider, registration.RecordType)
		}
		catalog.providers[registration.RecordType] = registration.Provider
	}
	missing := make([]string, 0)
	for recordType := range required {
		if catalog.providers[recordType] == nil {
			missing = append(missing, recordType)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: %v", ErrMissingDeleteRestoreProvider, missing)
	}
	return catalog, nil
}

func (c *DeleteRestoreProviderCatalog) Provider(recordType string) (recorddeleterestore.SourceProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.providers[recordType]
	return provider, ok
}

func nilDeleteRestoreProvider(provider recorddeleterestore.SourceProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
