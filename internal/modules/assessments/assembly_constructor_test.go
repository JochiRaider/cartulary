package assessments_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestAssessmentAssemblyConstructorsRejectInvalidDependencies(t *testing.T) {
	t.Parallel()

	validDatabase := assessmentAssemblyDatabase{}
	for _, constructor := range []struct {
		name      string
		wantError string
		construct func(postgres.DB) (any, error)
	}{
		{
			name: "subject validator", wantError: "compose assessment subject validator: database is required",
			construct: func(database postgres.DB) (any, error) {
				return assessmentassembly.NewSubjectValidator(database, hostidentity.NewSourceFacts())
			},
		},
		{
			name: "assessor validator", wantError: "compose assessment assessor validator: database is required",
			construct: func(database postgres.DB) (any, error) {
				return assessmentassembly.NewAssessorValidator(database)
			},
		},
		{
			name: "support-target validator", wantError: "compose assessment support-target validator: database is required",
			construct: func(database postgres.DB) (any, error) {
				return assessmentassembly.NewSupportTargetValidator(database)
			},
		},
		{
			name: "record-envelope creator", wantError: "compose assessment record-envelope creator: database is required",
			construct: func(database postgres.DB) (any, error) {
				return assessmentassembly.NewRecordEnvelopeCreator(database)
			},
		},
	} {
		constructor := constructor
		t.Run(constructor.name, func(t *testing.T) {
			t.Parallel()
			for _, database := range []postgres.DB{nil, (*typedNilAssemblyDatabase)(nil)} {
				value, err := constructor.construct(database)
				if value != nil || err == nil || err.Error() != constructor.wantError {
					t.Fatalf("construction = value:%v err:%v, want nil and %q", value, err, constructor.wantError)
				}
			}
			if value, err := constructor.construct(validDatabase); value == nil || err != nil {
				t.Fatalf("valid construction = value:%v err:%v", value, err)
			}
		})
	}

	if value, err := assessmentassembly.NewSubjectValidator(validDatabase, nil); value != nil || err == nil || err.Error() != "compose assessment subject validator: entity source facts are required" {
		t.Fatalf("nil entity source facts = value:%v err:%v", value, err)
	}

	for _, rows := range []assessmentprojection.Rows{nil, (*typedNilAssessmentRows)(nil)} {
		value, err := assessmentassembly.NewProjectionPort(rows)
		if value != nil || err == nil || err.Error() != "compose assessment projection port: rows are required" {
			t.Fatalf("invalid projection rows = value:%v err:%v", value, err)
		}
	}
	if value, err := assessmentassembly.NewProjectionPort(assessmentAssemblyRows{}); value == nil || err != nil {
		t.Fatalf("valid projection port = value:%v err:%v", value, err)
	}
}

func TestAssessmentMergeAssemblyRejectsNilAndTypedNilDependencies(t *testing.T) {
	t.Parallel()

	validRows := assessmentAssemblyRows{}
	validSnapshots := assessmentMergeSnapshotStub{}
	for _, rows := range []assessmentprojection.Rows{nil, (*typedNilAssessmentRows)(nil)} {
		effects, err := assessmentassembly.NewMergeEffects(rows, validSnapshots)
		if effects != nil || err == nil || err.Error() != "compose assessment merge effects: projection rows are required" {
			t.Fatalf("invalid merge rows = effects:%v err:%v", effects, err)
		}
	}
	for _, snapshots := range []assessments.MergeSnapshotCapturePort{nil, (*typedNilAssemblySnapshots)(nil)} {
		effects, err := assessmentassembly.NewMergeEffects(validRows, snapshots)
		if effects != nil || err == nil || err.Error() != "compose assessment merge effects: snapshot capture port is required" {
			t.Fatalf("invalid merge snapshots = effects:%v err:%v", effects, err)
		}
	}
	if effects, err := assessmentassembly.NewMergeEffects(validRows, validSnapshots); effects == nil || err != nil {
		t.Fatalf("valid merge assembly = effects:%v err:%v", effects, err)
	}
}

type assessmentAssemblyDatabase struct{ postgres.DB }
type typedNilAssemblyDatabase struct{ postgres.DB }
type assessmentAssemblyRows struct{ assessmentprojection.Rows }
type typedNilAssessmentRows struct{ assessmentprojection.Rows }
type typedNilAssemblySnapshots struct {
	assessments.MergeSnapshotCapturePort
}
