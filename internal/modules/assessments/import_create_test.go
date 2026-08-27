package assessments

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
)

func TestAssessmentCreateInputFromImportAcceptsExactFieldKinds(t *testing.T) {
	t.Parallel()
	subjectID := uuid.New()
	assessorID := uuid.New()
	assessedAt := time.Date(2026, time.August, 27, 17, 45, 0, 0, time.FixedZone("offset", 3600))
	request := ownerfacade.ImportOwnerCreateRequest{
		ClientTxnID: "txn-assessment-import-exact-kinds",
		FieldValues: []ownerfacade.ImportFieldValue{
			{FieldKey: "assessment.subject_ref", NormalizedValue: ownerfacade.NewUUIDImportScalar(subjectID)},
			{FieldKey: "assessment.subject_type", NormalizedValue: ownerfacade.NewTextImportScalar("host")},
			{FieldKey: "assessment.assessment_state", NormalizedValue: ownerfacade.NewTextImportScalar("confirmed")},
			{FieldKey: "assessment.confidence_score", NormalizedValue: ownerfacade.NewNumberImportScalar(70)},
			{FieldKey: "assessment.rationale", NormalizedValue: ownerfacade.NewTextImportScalar("Exact scalar contract.")},
			{FieldKey: "assessment.assessor", NormalizedValue: ownerfacade.NewUUIDImportScalar(assessorID)},
			{FieldKey: "assessment.assessed_at", NormalizedValue: ownerfacade.NewTimestampImportScalar(assessedAt)},
		},
	}

	input, err := assessmentCreateInputFromImport(request)
	if err != nil {
		t.Fatalf("classify exact assessment import fields: %v", err)
	}
	if input.ClientTxnID != request.ClientTxnID ||
		input.SubjectRef != subjectID ||
		input.SubjectType != "host" ||
		input.AssessmentState != "confirmed" ||
		input.ConfidenceScore == nil || *input.ConfidenceScore != 70 ||
		input.Rationale != "Exact scalar contract." ||
		input.Assessor == nil || *input.Assessor != assessorID ||
		input.AssessedAt == nil || !input.AssessedAt.Equal(assessedAt) {
		t.Fatalf("classified assessment input = %#v", input)
	}
}

func TestAssessmentCreateInputFromImportRejectsUnsafeFieldKinds(t *testing.T) {
	t.Parallel()
	collection := ownerfacade.NewCollectionTokenImportScalar(ownerfacade.ImportCollectionToken{
		RawText:        "record",
		NormalizedText: "record",
	})
	cases := []struct {
		name       string
		field      ownerfacade.ImportFieldValue
		wantReason string
		wantGuard  string
	}{
		{
			name: "unknown field",
			field: ownerfacade.ImportFieldValue{
				FieldKey: "assessment.future", NormalizedValue: ownerfacade.NewTextImportScalar("value"),
			},
			wantReason: "field_not_import_writable", wantGuard: "create_writable",
		},
		{
			name: "support collection",
			field: ownerfacade.ImportFieldValue{
				FieldKey: "assessment.support_refs", NormalizedValue: collection,
			},
			wantReason: "collection_owner_support_required", wantGuard: "collection_review",
		},
		{
			name: "subject ref text",
			field: ownerfacade.ImportFieldValue{
				FieldKey: "assessment.subject_ref", NormalizedValue: ownerfacade.NewTextImportScalar(uuid.NewString()),
			},
			wantReason: "invalid_uuid", wantGuard: "uuid",
		},
		{
			name: "subject type collection",
			field: ownerfacade.ImportFieldValue{
				FieldKey: "assessment.subject_type", NormalizedValue: collection,
			},
			wantReason: "invalid_text", wantGuard: "line_v1",
		},
		{
			name: "state number",
			field: ownerfacade.ImportFieldValue{
				FieldKey: "assessment.assessment_state", NormalizedValue: ownerfacade.NewNumberImportScalar(1),
			},
			wantReason: "invalid_text", wantGuard: "line_v1",
		},
		{
			name: "score text",
			field: ownerfacade.ImportFieldValue{
				FieldKey: "assessment.confidence_score", NormalizedValue: ownerfacade.NewTextImportScalar("70"),
			},
			wantReason: "invalid_integer", wantGuard: "number",
		},
		{
			name: "rationale bool",
			field: ownerfacade.ImportFieldValue{
				FieldKey: "assessment.rationale", NormalizedValue: ownerfacade.NewBoolImportScalar(true),
			},
			wantReason: "invalid_text", wantGuard: "multiline_body_v1",
		},
		{
			name: "assessor timestamp",
			field: ownerfacade.ImportFieldValue{
				FieldKey: "assessment.assessor", NormalizedValue: ownerfacade.NewTimestampImportScalar(time.Now()),
			},
			wantReason: "invalid_uuid", wantGuard: "uuid",
		},
		{
			name: "assessed at uuid",
			field: ownerfacade.ImportFieldValue{
				FieldKey: "assessment.assessed_at", NormalizedValue: ownerfacade.NewUUIDImportScalar(uuid.New()),
			},
			wantReason: "invalid_timestamp", wantGuard: "timestamp_instant_v1",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := assessmentCreateInputFromImport(ownerfacade.ImportOwnerCreateRequest{
				ClientTxnID: "txn-rejected-kind",
				FieldValues: []ownerfacade.ImportFieldValue{test.field},
			})
			requireAssessmentImportSafeError(t, err, test.field.FieldKey, test.wantReason, test.wantGuard)
		})
	}
}

func TestAssessmentCreateInputFromImportRejectsNullForEveryField(t *testing.T) {
	t.Parallel()
	for fieldKey := range assessmentImportFieldKinds {
		fieldKey := fieldKey
		t.Run(fieldKey, func(t *testing.T) {
			t.Parallel()
			_, err := assessmentCreateInputFromImport(ownerfacade.ImportOwnerCreateRequest{
				ClientTxnID: "txn-null",
				FieldValues: []ownerfacade.ImportFieldValue{{
					FieldKey: fieldKey, NormalizedValue: ownerfacade.NewNullImportScalar(),
				}},
			})
			requireAssessmentImportSafeError(t, err, fieldKey, "field_not_nullable", "clearable")
		})
	}
}

func TestAssessmentCreateInputFromImportRetainsSharedStructuralRejection(t *testing.T) {
	t.Parallel()
	valid := ownerfacade.ImportFieldValue{
		FieldKey: "assessment.rationale", NormalizedValue: ownerfacade.NewTextImportScalar("duplicate"),
	}
	for _, test := range []struct {
		name   string
		fields []ownerfacade.ImportFieldValue
	}{
		{name: "duplicate", fields: []ownerfacade.ImportFieldValue{valid, valid}},
		{name: "empty key", fields: []ownerfacade.ImportFieldValue{{
			NormalizedValue: ownerfacade.NewTextImportScalar("empty key"),
		}}},
		{name: "corrupt value", fields: []ownerfacade.ImportFieldValue{{
			FieldKey: "assessment.rationale", NormalizedValue: ownerfacade.ImportScalarValue{},
		}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := assessmentCreateInputFromImport(ownerfacade.ImportOwnerCreateRequest{
				ClientTxnID: "txn-structural", FieldValues: test.fields,
			})
			if err == nil {
				t.Fatal("structurally invalid assessment import unexpectedly succeeded")
			}
			var ownerErr *ownerfacade.ImportOwnerCreateError
			if errors.As(err, &ownerErr) {
				t.Fatalf("shared structural error exposed owner detail: %#v", ownerErr)
			}
			if detail, ok := ownerfacade.ImportOwnerCreateErrorDetail(err); ok || detail != nil {
				t.Fatalf("shared structural detail = %#v, %t", detail, ok)
			}
		})
	}
}

func requireAssessmentImportSafeError(
	t testing.TB,
	err error,
	field string,
	reason string,
	guard string,
) {
	t.Helper()
	var ownerErr *ownerfacade.ImportOwnerCreateError
	if !errors.As(err, &ownerErr) {
		t.Fatalf("assessment import error = %T %v", err, err)
	}
	if errors.Unwrap(ownerErr) != nil {
		t.Fatalf("assessment import error unexpectedly wraps a cause: %v", errors.Unwrap(ownerErr))
	}
	detail, ok := ownerfacade.ImportOwnerCreateErrorDetail(err)
	if !ok {
		t.Fatalf("assessment import error has no safe detail: %#v", ownerErr)
	}
	want := map[string]any{
		"owner_code": ownerfacade.ImportOwnerCreateValidationFailed,
		"retryable":  false,
		"safe_details": map[string]any{
			"reason_code": reason,
			"field":       field,
			"guard":       guard,
		},
	}
	if !reflect.DeepEqual(detail, want) {
		t.Fatalf("assessment import detail = %#v, want %#v", detail, want)
	}
}
