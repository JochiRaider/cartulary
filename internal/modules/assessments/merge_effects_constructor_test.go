package assessments_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestAssessmentMergeEffectsConstruction(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		projections assessments.MergeProjectionPort
		snapshots   assessments.MergeSnapshotCapturePort
		wantError   string
	}{
		{
			name:      "projection required",
			snapshots: assessmentMergeSnapshotStub{},
			wantError: "construct assessment merge effects: projection port is required",
		},
		{
			name:        "typed nil projection required",
			projections: (*typedNilMergeProjection)(nil),
			snapshots:   assessmentMergeSnapshotStub{},
			wantError:   "construct assessment merge effects: projection port is required",
		},
		{
			name:        "snapshot capture required",
			projections: assessmentMergeProjectionStub{},
			wantError:   "construct assessment merge effects: snapshot capture port is required",
		},
		{
			name:        "typed nil snapshot capture required",
			projections: assessmentMergeProjectionStub{},
			snapshots:   (*typedNilMergeSnapshot)(nil),
			wantError:   "construct assessment merge effects: snapshot capture port is required",
		},
		{
			name:        "valid",
			projections: assessmentMergeProjectionStub{},
			snapshots:   assessmentMergeSnapshotStub{},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			effects, err := assessments.NewMergeEffects(test.projections, test.snapshots)
			if test.wantError != "" {
				if err == nil || err.Error() != test.wantError || effects != nil {
					t.Fatalf("construction = effects:%v err:%v; want nil and %q", effects, err, test.wantError)
				}
				return
			}
			if err != nil || effects == nil {
				t.Fatalf("valid construction = effects:%v err:%v", effects, err)
			}
		})
	}
}

type assessmentMergeProjectionStub struct{}

func (assessmentMergeProjectionStub) RefreshAssessmentProjectionTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

type assessmentMergeSnapshotStub struct{}

func (assessmentMergeSnapshotStub) CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error) {
	return revisions.RecordSnapshot{}, nil
}

type typedNilMergeProjection struct {
	assessments.MergeProjectionPort
}
type typedNilMergeSnapshot struct {
	assessments.MergeSnapshotCapturePort
}
