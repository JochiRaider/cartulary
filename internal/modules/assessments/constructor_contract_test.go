package assessments_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestAssessmentFacadeConstructionRejectsNilAndTypedNilDependencies(t *testing.T) {
	t.Parallel()

	for _, dependency := range []struct {
		name      string
		wantError string
		clear     func(*assessments.FacadeDependencies, bool)
	}{
		{
			name: "idempotency", wantError: "construct assessment facade: idempotency port is required",
			clear: func(dependencies *assessments.FacadeDependencies, typed bool) {
				if typed {
					var value *typedNilCreateIdempotency
					dependencies.Idempotency = value
				} else {
					dependencies.Idempotency = nil
				}
			},
		},
		{
			name: "subjects", wantError: "construct assessment facade: subject validator is required",
			clear: func(dependencies *assessments.FacadeDependencies, typed bool) {
				if typed {
					var value *typedNilSubjectValidator
					dependencies.Subjects = value
				} else {
					dependencies.Subjects = nil
				}
			},
		},
		{
			name: "assessors", wantError: "construct assessment facade: assessor validator is required",
			clear: func(dependencies *assessments.FacadeDependencies, typed bool) {
				if typed {
					var value *typedNilAssessorValidator
					dependencies.Assessors = value
				} else {
					dependencies.Assessors = nil
				}
			},
		},
		{
			name: "support targets", wantError: "construct assessment facade: support-target validator is required",
			clear: func(dependencies *assessments.FacadeDependencies, typed bool) {
				if typed {
					var value *typedNilSupportTargetValidator
					dependencies.SupportTargets = value
				} else {
					dependencies.SupportTargets = nil
				}
			},
		},
		{
			name: "records", wantError: "construct assessment facade: record-envelope creator is required",
			clear: func(dependencies *assessments.FacadeDependencies, typed bool) {
				if typed {
					var value *typedNilRecordEnvelopeCreator
					dependencies.Records = value
				} else {
					dependencies.Records = nil
				}
			},
		},
		{
			name: "support links", wantError: "construct assessment facade: support-link applier is required",
			clear: func(dependencies *assessments.FacadeDependencies, typed bool) {
				if typed {
					var value *typedNilSupportLinkApplier
					dependencies.SupportLinks = value
				} else {
					dependencies.SupportLinks = nil
				}
			},
		},
		{
			name: "revisions", wantError: "construct assessment facade: revision appender is required",
			clear: func(dependencies *assessments.FacadeDependencies, typed bool) {
				if typed {
					var value *typedNilCreateRevisionAppender
					dependencies.Revisions = value
				} else {
					dependencies.Revisions = nil
				}
			},
		},
		{
			name: "projections", wantError: "construct assessment facade: projection port is required",
			clear: func(dependencies *assessments.FacadeDependencies, typed bool) {
				if typed {
					var value *typedNilAssessmentProjection
					dependencies.Projections = value
				} else {
					dependencies.Projections = nil
				}
			},
		},
	} {
		dependency := dependency
		for _, typed := range []bool{false, true} {
			typed := typed
			label := "nil"
			if typed {
				label = "typed nil"
			}
			t.Run(dependency.name+"/"+label, func(t *testing.T) {
				t.Parallel()
				dependencies := validAssessmentFacadeDependencies()
				dependency.clear(&dependencies, typed)
				facade, err := assessments.NewFacade(assessmentConstructorDB{}, dependencies)
				if facade != nil || err == nil || err.Error() != dependency.wantError {
					t.Fatalf("construction = facade:%v err:%v, want nil and %q", facade, err, dependency.wantError)
				}
			})
		}
	}

	for _, database := range []postgres.DB{nil, (*typedNilDatabase)(nil)} {
		facade, err := assessments.NewFacade(database, validAssessmentFacadeDependencies())
		if facade != nil || err == nil || err.Error() != "construct assessment facade: database is required" {
			t.Fatalf("database construction = facade:%v err:%v", facade, err)
		}
	}
	if facade, err := assessments.NewFacade(assessmentConstructorDB{}, validAssessmentFacadeDependencies()); err != nil || facade == nil {
		t.Fatalf("valid assessment facade construction = facade:%v err:%v", facade, err)
	}
}

func TestAssessmentImportFacadeConstructionRejectsNilAndTypedNilDependencies(t *testing.T) {
	t.Parallel()

	for _, dependency := range []struct {
		name      string
		wantError string
		clear     func(*assessments.ImportCreateDependencies, bool)
	}{
		{
			name: "subjects", wantError: "construct assessment import facade: subject validator is required",
			clear: func(dependencies *assessments.ImportCreateDependencies, typed bool) {
				if typed {
					var value *typedNilSubjectValidator
					dependencies.Subjects = value
				} else {
					dependencies.Subjects = nil
				}
			},
		},
		{
			name: "assessors", wantError: "construct assessment import facade: assessor validator is required",
			clear: func(dependencies *assessments.ImportCreateDependencies, typed bool) {
				if typed {
					var value *typedNilAssessorValidator
					dependencies.Assessors = value
				} else {
					dependencies.Assessors = nil
				}
			},
		},
		{
			name: "records", wantError: "construct assessment import facade: record-envelope creator is required",
			clear: func(dependencies *assessments.ImportCreateDependencies, typed bool) {
				if typed {
					var value *typedNilRecordEnvelopeCreator
					dependencies.Records = value
				} else {
					dependencies.Records = nil
				}
			},
		},
		{
			name: "revisions", wantError: "construct assessment import facade: revision appender is required",
			clear: func(dependencies *assessments.ImportCreateDependencies, typed bool) {
				if typed {
					var value *typedNilLiveRevisionAppender
					dependencies.Revisions = value
				} else {
					dependencies.Revisions = nil
				}
			},
		},
		{
			name: "projections", wantError: "construct assessment import facade: projection port is required",
			clear: func(dependencies *assessments.ImportCreateDependencies, typed bool) {
				if typed {
					var value *typedNilAssessmentProjection
					dependencies.Projections = value
				} else {
					dependencies.Projections = nil
				}
			},
		},
		{
			name: "collaboration", wantError: "construct assessment import facade: collaboration publication appender is required",
			clear: func(dependencies *assessments.ImportCreateDependencies, typed bool) {
				if typed {
					var value *typedNilRecordChangedAppender
					dependencies.Collaboration = value
				} else {
					dependencies.Collaboration = nil
				}
			},
		},
	} {
		dependency := dependency
		for _, typed := range []bool{false, true} {
			typed := typed
			label := "nil"
			if typed {
				label = "typed nil"
			}
			t.Run(dependency.name+"/"+label, func(t *testing.T) {
				t.Parallel()
				dependencies := validAssessmentImportDependencies()
				dependency.clear(&dependencies, typed)
				facade, err := assessments.NewImportCreateFacade(
					assessments.AssessmentsViewSchemaID,
					"assessments.import_create",
					dependencies,
				)
				if facade != nil || err == nil || err.Error() != dependency.wantError {
					t.Fatalf("construction = facade:%v err:%v, want nil and %q", facade, err, dependency.wantError)
				}
			})
		}
	}
	if facade, err := assessments.NewImportCreateFacade(
		assessments.AssessmentsViewSchemaID,
		"assessments.import_create",
		validAssessmentImportDependencies(),
	); err != nil || facade == nil {
		t.Fatalf("valid assessment import construction = facade:%v err:%v", facade, err)
	}
}

func TestAssessmentProjectionConstructionRejectsNilAndTypedNilDependencies(t *testing.T) {
	t.Parallel()

	valid := assessments.ProjectionContributionDependencies{
		Envelopes: assessmentProjectionEnvelopeStub{},
		Support:   assessmentProjectionSupportStub{},
	}
	for _, test := range []struct {
		name         string
		dependencies assessments.ProjectionContributionDependencies
		wantError    string
	}{
		{
			name: "nil envelopes", dependencies: assessments.ProjectionContributionDependencies{Support: valid.Support},
			wantError: "construct assessment projection contribution: envelope reader is required",
		},
		{
			name: "typed nil envelopes", dependencies: assessments.ProjectionContributionDependencies{
				Envelopes: (*typedNilEnvelopeReader)(nil), Support: valid.Support,
			},
			wantError: "construct assessment projection contribution: envelope reader is required",
		},
		{
			name: "nil support", dependencies: assessments.ProjectionContributionDependencies{Envelopes: valid.Envelopes},
			wantError: "construct assessment projection contribution: support fact reader is required",
		},
		{
			name: "typed nil support", dependencies: assessments.ProjectionContributionDependencies{
				Envelopes: valid.Envelopes, Support: (*typedNilSupportFactReader)(nil),
			},
			wantError: "construct assessment projection contribution: support fact reader is required",
		},
	} {
		contribution, err := assessments.NewProjectionContribution(test.dependencies)
		if err == nil || err.Error() != test.wantError || contribution.ProjectionContribution().SourceOwnerModule() != "" {
			t.Fatalf("%s construction = contribution:%#v err:%v", test.name, contribution, err)
		}
	}
	if contribution, err := assessments.NewProjectionContribution(valid); err != nil || contribution.ProjectionContribution().SourceOwnerModule() != "assessments" {
		t.Fatalf("valid projection contribution = %#v err:%v", contribution, err)
	}
}

func validAssessmentFacadeDependencies() assessments.FacadeDependencies {
	ports := &assessmentFacadePorts{}
	return assessments.FacadeDependencies{
		Idempotency:    ports,
		Subjects:       ports,
		Assessors:      ports,
		SupportTargets: ports,
		Records:        ports,
		SupportLinks:   ports,
		Revisions:      ports,
		Projections:    ports,
	}
}

func validAssessmentImportDependencies() assessments.ImportCreateDependencies {
	ports := &assessmentFacadePorts{}
	return assessments.ImportCreateDependencies{
		Subjects:      ports,
		Assessors:     ports,
		Records:       ports,
		Revisions:     assessmentConstructorLiveRevisions{},
		Projections:   ports,
		Collaboration: assessmentConstructorPublications{},
	}
}

type assessmentConstructorDB struct{ postgres.DB }
type assessmentConstructorLiveRevisions struct {
	ownerfacade.LiveRecordRevisionAppender
}
type assessmentConstructorPublications struct {
	collaboration.RecordChangedAppender
}

type typedNilDatabase struct{ postgres.DB }
type typedNilCreateIdempotency struct {
	assessments.CreateIdempotencyPort
}
type typedNilSubjectValidator struct{ assessments.SubjectValidator }
type typedNilAssessorValidator struct{ assessments.AssessorValidator }
type typedNilSupportTargetValidator struct {
	assessments.SupportTargetValidator
}
type typedNilRecordEnvelopeCreator struct {
	assessments.RecordEnvelopeCreator
}
type typedNilSupportLinkApplier struct {
	assessments.InitialSupportLinkApplier
}
type typedNilCreateRevisionAppender struct {
	assessments.CreateRevisionAppender
}
type typedNilAssessmentProjection struct {
	assessments.AssessmentProjectionPort
}
type typedNilLiveRevisionAppender struct {
	ownerfacade.LiveRecordRevisionAppender
}
type typedNilRecordChangedAppender struct {
	collaboration.RecordChangedAppender
}
type typedNilEnvelopeReader struct {
	workbookprojection.EnvelopeReader
}
type typedNilSupportFactReader struct {
	workbookprojection.SupportFactReader
}
