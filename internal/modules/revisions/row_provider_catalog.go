package revisions

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

var (
	ErrDuplicateRowRollbackProvider  = errors.New("revisions: duplicate row rollback provider")
	ErrMissingRowRollbackProvider    = errors.New("revisions: missing row rollback provider")
	ErrUnexpectedRowRollbackProvider = errors.New("revisions: unexpected row rollback provider")
)

type RowProviderRegistration struct {
	RecordType string
	Provider   rollbackcontract.RowSourceProvider
}

type RowProviderCatalog struct {
	providers map[string]rollbackcontract.RowSourceProvider
}

func NewRowProviderCatalog(requiredRecordTypes []string, registrations ...RowProviderRegistration) (*RowProviderCatalog, error) {
	required := make(map[string]struct{}, len(requiredRecordTypes))
	for _, recordType := range requiredRecordTypes {
		if recordType == "" {
			return nil, fmt.Errorf("%w: empty record type", ErrMissingRowRollbackProvider)
		}
		required[recordType] = struct{}{}
	}
	catalog := &RowProviderCatalog{providers: make(map[string]rollbackcontract.RowSourceProvider, len(registrations))}
	for _, registration := range registrations {
		if registration.RecordType == "" || nilRowProvider(registration.Provider) {
			return nil, fmt.Errorf("%w: %q", ErrMissingRowRollbackProvider, registration.RecordType)
		}
		if _, allowed := required[registration.RecordType]; !allowed {
			return nil, fmt.Errorf("%w: %s", ErrUnexpectedRowRollbackProvider, registration.RecordType)
		}
		if _, exists := catalog.providers[registration.RecordType]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateRowRollbackProvider, registration.RecordType)
		}
		catalog.providers[registration.RecordType] = registration.Provider
	}
	missing := make([]string, 0)
	for recordType := range required {
		if catalog.providers[recordType] == nil {
			missing = append(missing, recordType)
		}
	}
	if len(missing) == 0 {
		return catalog, nil
	}
	sort.Strings(missing)
	return nil, fmt.Errorf("%w: %v", ErrMissingRowRollbackProvider, missing)
}

func nilRowProvider(provider rollbackcontract.RowSourceProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func (c *RowProviderCatalog) Provider(recordType string) (rollbackcontract.RowSourceProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.providers[recordType]
	return provider, ok
}
