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

func TestTargetSemanticsCatalogRejectsInvalidNonRowProviders(t *testing.T) {
	t.Parallel()
	requirement := func(targetKind string) targetSemanticsRequirement {
		return targetSemanticsRequirement{
			TargetKind:            targetKind,
			SourceOwnerID:         "links",
			DispatchClass:         rollbackcontract.DispatchNonRow,
			HistoryRecordIDFields: []string{"record_id"},
			Addressability:        HistorySingleEntry,
		}
	}
	contribution := func(targetKinds ...string) []ProviderContribution {
		targets := make([]NonRowProviderContribution, 0, len(targetKinds))
		for _, targetKind := range targetKinds {
			targets = append(targets, NonRowProviderContribution{
				SourceOwnerModule: SourceOwnerLinks,
				TargetKind:        targetKind,
				HistoryFacet:      NewFieldAssociationHistoryFacet([]string{"record_id"}, HistorySingleEntry),
				RollbackProvider:  stubNonRowProvider{},
			})
		}
		return []ProviderContribution{{SourceOwnerModule: SourceOwnerLinks, NonRowTargets: targets}}
	}

	if _, err := compileTargetSemanticsCatalog(
		[]targetSemanticsRequirement{requirement("record_link"), requirement("record_tag")},
		contribution("record_link"),
	); !errors.Is(err, ErrMissingTargetSemantics) {
		t.Fatalf("missing non-row provider error = %v", err)
	}
	if _, err := compileTargetSemanticsCatalog(
		[]targetSemanticsRequirement{requirement("record_link")},
		contribution("record_link", "record_link"),
	); !errors.Is(err, ErrDuplicateTargetSemantics) {
		t.Fatalf("duplicate non-row provider error = %v", err)
	}
	if _, err := compileTargetSemanticsCatalog(
		[]targetSemanticsRequirement{requirement("record_link")},
		contribution("record_tag"),
	); !errors.Is(err, ErrUnexpectedTargetSemantics) {
		t.Fatalf("unexpected non-row provider error = %v", err)
	}
	values := contribution("record_link")
	var typedNil *stubNonRowProvider
	values[0].NonRowTargets[0].RollbackProvider = typedNil
	if _, err := compileTargetSemanticsCatalog([]targetSemanticsRequirement{requirement("record_link")}, values); !errors.Is(err, ErrInvalidTargetSemantics) {
		t.Fatalf("typed nil non-row provider error = %v", err)
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
