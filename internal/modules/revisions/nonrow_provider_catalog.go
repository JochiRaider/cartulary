package revisions

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

var (
	ErrDuplicateNonRowRollbackProvider  = errors.New("revisions: duplicate non-row rollback provider")
	ErrMissingNonRowRollbackProvider    = errors.New("revisions: missing non-row rollback provider")
	ErrUnexpectedNonRowRollbackProvider = errors.New("revisions: unexpected non-row rollback provider")
)

type NonRowProviderRegistration struct {
	TargetKind string
	Provider   rollbackcontract.NonRowTargetProvider
}

type NonRowProviderCatalog struct {
	providers map[string]rollbackcontract.NonRowTargetProvider
}

func NewNonRowProviderCatalog(requiredTargetKinds []string, registrations ...NonRowProviderRegistration) (*NonRowProviderCatalog, error) {
	required := make(map[string]struct{}, len(requiredTargetKinds))
	for _, targetKind := range requiredTargetKinds {
		if targetKind == "" {
			return nil, fmt.Errorf("%w: empty target kind", ErrMissingNonRowRollbackProvider)
		}
		required[targetKind] = struct{}{}
	}
	catalog := &NonRowProviderCatalog{providers: make(map[string]rollbackcontract.NonRowTargetProvider, len(registrations))}
	for _, registration := range registrations {
		if registration.TargetKind == "" || nilNonRowProvider(registration.Provider) {
			return nil, fmt.Errorf("%w: %q", ErrMissingNonRowRollbackProvider, registration.TargetKind)
		}
		if _, allowed := required[registration.TargetKind]; !allowed {
			return nil, fmt.Errorf("%w: %s", ErrUnexpectedNonRowRollbackProvider, registration.TargetKind)
		}
		if _, exists := catalog.providers[registration.TargetKind]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateNonRowRollbackProvider, registration.TargetKind)
		}
		catalog.providers[registration.TargetKind] = registration.Provider
	}
	missing := make([]string, 0)
	for targetKind := range required {
		if catalog.providers[targetKind] == nil {
			missing = append(missing, targetKind)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: %v", ErrMissingNonRowRollbackProvider, missing)
	}
	return catalog, nil
}

func (c *NonRowProviderCatalog) Provider(targetKind string) (rollbackcontract.NonRowTargetProvider, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.providers[targetKind]
	return provider, ok
}

func nilNonRowProvider(provider rollbackcontract.NonRowTargetProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func validateNonRowApplyResult(descriptor rollbackcontract.TargetDescriptor, result rollbackcontract.ApplyInverseResult) error {
	expected := canonicalRecordIDs(descriptor.AffectedRecordIDs)
	actual := canonicalRecordIDs(result.AffectedRecordIDs)
	if len(expected) != len(actual) {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	for index := range expected {
		if expected[index] != actual[index] {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	}
	return nil
}

func canonicalRecordIDs(values []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value != uuid.Nil {
			set[value] = struct{}{}
		}
	}
	result := make([]uuid.UUID, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].String() < result[j].String() })
	return result
}
