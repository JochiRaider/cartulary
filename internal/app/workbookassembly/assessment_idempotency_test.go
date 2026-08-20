package workbookassembly

import (
	"encoding/json"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
)

func TestAssessmentCreateResultCodec(t *testing.T) {
	t.Parallel()
	incidentID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	recordID := uuid.MustParse("2aaaaaaa-0000-4000-8000-000000000001")
	changeSetID := uuid.MustParse("3bbbbbbb-0000-4000-8000-000000000001")
	scopeKey := incidentID.String() + ":" + assessments.AssessmentsViewSchemaID
	requestHash := []byte("request-hash")
	valid := `{
  "schema_id":"cartulary.assessments.create_result.v1",
  "record_id":"2aaaaaaa-0000-4000-8000-000000000001",
  "change_set_id":"3bbbbbbb-0000-4000-8000-000000000001",
  "row_version":7,
  "row":{
    "record_id":"2aaaaaaa-0000-4000-8000-000000000001",
    "incident_id":"10000000-0000-4000-8000-000000000001",
    "row_version":7,
    "future_field":{"nested":true}
  }
}`

	result, err := decodeAssessmentCreateResult([]byte(valid), scopeKey, http.StatusCreated, requestHash)
	if err != nil {
		t.Fatalf("decode valid canonical result: %v", err)
	}
	if result.RecordID != recordID || result.ChangeSetID != changeSetID || result.RowVersion != 7 {
		t.Fatalf("decoded identity/version = %#v", result)
	}
	if result.Outcome != assessments.CreateOutcomeCommitted || result.CanonicalRow["future_field"] == nil {
		t.Fatalf("decoded result lost committed outcome or extensible row: %#v", result)
	}

	encoded, err := encodeAssessmentCreateResult(scopeKey, requestHash, assessments.CreateResult{
		RecordID:    recordID,
		ChangeSetID: changeSetID,
		RowVersion:  7,
		CanonicalRow: map[string]any{
			"record_id":   recordID.String(),
			"incident_id": incidentID.String(),
			"row_version": int64(7),
		},
	})
	if err != nil {
		t.Fatalf("encode canonical result: %v", err)
	}
	var encodedObject map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &encodedObject); err != nil {
		t.Fatalf("decode encoded object for inspection: %v", err)
	}
	if !hasExactJSONMembers(encodedObject, assessmentCreateResultMembers) {
		t.Fatalf("encoded members = %v", encodedObject)
	}

	for _, test := range []struct {
		name       string
		payload    string
		scope      string
		status     int
		request    []byte
		wantReason string
	}{
		{name: "legacy", payload: `{"view_schema_id":"cartulary.view.assessments.v1","change_set_id":"3bbbbbbb-0000-4000-8000-000000000001","row":{"record_id":"2aaaaaaa-0000-4000-8000-000000000001","incident_id":"10000000-0000-4000-8000-000000000001","row_version":7}}`, scope: scopeKey, status: 201, request: requestHash, wantReason: "invalid members"},
		{name: "extra top member", payload: strings.Replace(valid, `"row":{`, `"extra":true,"row":{`, 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "invalid members"},
		{name: "duplicate nested member", payload: strings.Replace(valid, "\"row_version\":7,\n    \"future_field\"", "\"row_version\":7,\n    \"row_version\":7,\n    \"future_field\"", 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "duplicate object member"},
		{name: "unknown schema", payload: strings.Replace(valid, assessmentCreateResultSchemaID, "cartulary.assessments.create_result.v2", 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "invalid schema_id"},
		{name: "noncanonical record uuid", payload: strings.Replace(valid, recordID.String(), strings.ToUpper(recordID.String()), 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "invalid record_id"},
		{name: "nil change set uuid", payload: strings.Replace(valid, changeSetID.String(), uuid.Nil.String(), 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "invalid change_set_id"},
		{name: "fractional top version", payload: strings.Replace(valid, `"row_version":7,`, `"row_version":7.5,`, 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "invalid row_version"},
		{name: "overflow top version", payload: strings.Replace(valid, `"row_version":7,`, `"row_version":9223372036854775808,`, 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "invalid row_version"},
		{name: "row record mismatch", payload: strings.Replace(valid, "\"row\":{\n    \"record_id\":\""+recordID.String(), "\"row\":{\n    \"record_id\":\"2aaaaaaa-0000-4000-8000-000000000002", 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "row record_id mismatch"},
		{name: "row incident mismatch", payload: strings.Replace(valid, incidentID.String(), "10000000-0000-4000-8000-000000000002", 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "row incident_id mismatch"},
		{name: "row version mismatch", payload: strings.Replace(valid, "\"row_version\":7,\n    \"future_field\"", "\"row_version\":8,\n    \"future_field\"", 1), scope: scopeKey, status: 201, request: requestHash, wantReason: "row_version mismatch"},
		{name: "invalid scope", payload: valid, scope: "not-an-incident:" + assessments.AssessmentsViewSchemaID, status: 201, request: requestHash, wantReason: "invalid scope"},
		{name: "wrong status", payload: valid, scope: scopeKey, status: 200, request: requestHash, wantReason: "invalid stored status"},
		{name: "empty request hash", payload: valid, scope: scopeKey, status: 201, request: nil, wantReason: "request hash is required"},
		{name: "not object", payload: `[]`, scope: scopeKey, status: 201, request: requestHash, wantReason: "value is not one object"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeAssessmentCreateResult([]byte(test.payload), test.scope, test.status, test.request)
			if err == nil || !strings.Contains(err.Error(), test.wantReason) {
				t.Fatalf("decode error = %v; want reason %q", err, test.wantReason)
			}
			if strings.Contains(err.Error(), recordID.String()) || strings.Contains(err.Error(), incidentID.String()) {
				t.Fatalf("decode error leaked stored identity: %v", err)
			}
		})
	}

	if _, err := encodeAssessmentCreateResult(scopeKey, requestHash, assessments.CreateResult{
		RecordID: recordID, ChangeSetID: changeSetID, RowVersion: math.MaxInt64,
		CanonicalRow: map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "row_version": int64(0)},
	}); err == nil {
		t.Fatal("encode accepted a canonical row that disagrees with the result")
	}
}
