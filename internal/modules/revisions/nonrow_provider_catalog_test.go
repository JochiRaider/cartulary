package revisions

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type stubNonRowProvider struct{}

func (stubNonRowProvider) DescribeTx(context.Context, pgx.Tx, rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	return rollbackcontract.TargetDescriptor{}, nil
}

func (stubNonRowProvider) ApplyInverseTx(context.Context, pgx.Tx, rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
	return rollbackcontract.ApplyInverseResult{}, nil
}

func TestNonRowProviderCatalogFailsClosed(t *testing.T) {
	t.Parallel()
	required := []string{"record_link", "record_tag"}
	if _, err := NewNonRowProviderCatalog(required, NonRowProviderRegistration{TargetKind: "record_link", Provider: stubNonRowProvider{}}); !errors.Is(err, ErrMissingNonRowRollbackProvider) {
		t.Fatalf("missing provider error = %v", err)
	}
	if _, err := NewNonRowProviderCatalog([]string{"record_link"},
		NonRowProviderRegistration{TargetKind: "record_link", Provider: stubNonRowProvider{}},
		NonRowProviderRegistration{TargetKind: "record_link", Provider: stubNonRowProvider{}},
	); !errors.Is(err, ErrDuplicateNonRowRollbackProvider) {
		t.Fatalf("duplicate provider error = %v", err)
	}
	if _, err := NewNonRowProviderCatalog([]string{"record_link"}, NonRowProviderRegistration{TargetKind: "record_tag", Provider: stubNonRowProvider{}}); !errors.Is(err, ErrUnexpectedNonRowRollbackProvider) {
		t.Fatalf("unexpected provider error = %v", err)
	}
	var typedNil *stubNonRowProvider
	if _, err := NewNonRowProviderCatalog([]string{"record_link"}, NonRowProviderRegistration{TargetKind: "record_link", Provider: typedNil}); !errors.Is(err, ErrMissingNonRowRollbackProvider) {
		t.Fatalf("typed nil provider error = %v", err)
	}
}

func TestValidateNonRowApplyResultRequiresExactCanonicalSet(t *testing.T) {
	t.Parallel()
	first := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	second := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	descriptor := rollbackcontract.TargetDescriptor{AffectedRecordIDs: []uuid.UUID{second, first}}
	if err := validateNonRowApplyResult(descriptor, rollbackcontract.ApplyInverseResult{AffectedRecordIDs: []uuid.UUID{first, second, first}}); err != nil {
		t.Fatalf("canonical equal sets rejected: %v", err)
	}
	if err := validateNonRowApplyResult(descriptor, rollbackcontract.ApplyInverseResult{AffectedRecordIDs: []uuid.UUID{first}}); err == nil {
		t.Fatal("changed affected set was accepted")
	}
}
