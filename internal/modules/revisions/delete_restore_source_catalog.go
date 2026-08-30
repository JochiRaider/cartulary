package revisions

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

var (
	ErrDuplicateDeleteRestoreSource  = errors.New("revisions: duplicate delete/restore source")
	ErrMissingDeleteRestoreSource    = errors.New("revisions: missing delete/restore source")
	ErrUnexpectedDeleteRestoreSource = errors.New("revisions: unexpected delete/restore source")
)

type deleteRestoreSourceRegistration struct {
	RecordType string
	Source     deleterestorecontract.DeleteRestoreSource
}

type DeleteRestoreSourceCatalog struct {
	sources map[string]deleterestorecontract.DeleteRestoreSource
}

func newDeleteRestoreSourceCatalog(requiredRecordTypes []string, registrations ...deleteRestoreSourceRegistration) (*DeleteRestoreSourceCatalog, error) {
	required := make(map[string]struct{}, len(requiredRecordTypes))
	for _, recordType := range requiredRecordTypes {
		if recordType == "" {
			return nil, fmt.Errorf("%w: empty record type", ErrMissingDeleteRestoreSource)
		}
		required[recordType] = struct{}{}
	}
	catalog := &DeleteRestoreSourceCatalog{sources: make(map[string]deleterestorecontract.DeleteRestoreSource, len(registrations))}
	for _, registration := range registrations {
		if registration.RecordType == "" || nilDeleteRestoreSource(registration.Source) {
			return nil, fmt.Errorf("%w: %q", ErrMissingDeleteRestoreSource, registration.RecordType)
		}
		if _, allowed := required[registration.RecordType]; !allowed {
			return nil, fmt.Errorf("%w: %s", ErrUnexpectedDeleteRestoreSource, registration.RecordType)
		}
		if _, exists := catalog.sources[registration.RecordType]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateDeleteRestoreSource, registration.RecordType)
		}
		catalog.sources[registration.RecordType] = registration.Source
	}
	missing := make([]string, 0)
	for recordType := range required {
		if catalog.sources[recordType] == nil {
			missing = append(missing, recordType)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: %v", ErrMissingDeleteRestoreSource, missing)
	}
	return catalog, nil
}

func (c *DeleteRestoreSourceCatalog) Source(recordType string) (deleterestorecontract.DeleteRestoreSource, bool) {
	if c == nil {
		return nil, false
	}
	source, ok := c.sources[recordType]
	return source, ok
}

func nilDeleteRestoreSource(source deleterestorecontract.DeleteRestoreSource) bool {
	if source == nil {
		return true
	}
	value := reflect.ValueOf(source)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
