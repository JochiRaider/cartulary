package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	contractreporting "github.com/JochiRaider/cartulary/internal/gen/contractreporting"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func testIndependentRenderFailureRecord(t *testing.T) {
	t.Run("typed companion failures preserve transport errors", func(t *testing.T) {
		for _, sample := range []struct {
			err    error
			reason string
		}{
			{reportcomposition.ErrNotFound, "retained_input_missing"},
			{&reportcomposition.ReleaseTupleError{ReasonCode: reportcomposition.ReleaseTupleReasonVersionNotFound}, "retained_input_missing"},
			{&reportcomposition.ReleaseTupleError{ReasonCode: reportcomposition.ReleaseTupleReasonDigestMismatch}, "retained_input_mismatch"},
			{&reportcomposition.RenderSourceError{ReasonCode: "composition_schema_invalid"}, "retained_input_mismatch"},
		} {
			got := renderFailureDetails(retainedCompositionInputError(sample.err))
			if got.FailureCode != "retained_input_invalid" || got.ReasonCode != sample.reason || len(got.SafeDetails) != 0 {
				t.Fatal("companion classification drift", got)
			}
		}
		transport := errors.New("transport fixture")
		if retainedCompositionInputError(transport) != transport {
			t.Fatal("transport error became input corruption")
		}
	})

	incident := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	actor := "00000000-0000-0000-0000-000000000003"
	at := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	retained := jobs.RetainedExtensionJob{OwnerProfileID: ProfileID, HandlerName: JobWorkerKind, JobKind: ReleaseCreateJobKind,
		RequestSHA256: strings.Repeat("a", 64), Resource: jobs.Resource{JobID: "00000000-0000-0000-0000-000000000001",
			AuthPolicy: jobs.AuthPolicyIncidentMembership, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incident}, SubmittedByUserID: &actor, SubmittedAt: at}}
	summary, err := encodeFailedRenderSummary(nil, nil, "2026-11-02T12:00:00.000000Z", RenderFailure{Stage: "release_state", FailureCode: "retained_input_invalid", ReasonCode: "retained_input_missing", SafeDetails: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	record, err := bindRenderFailureRecord(retained, summary)
	if err != nil {
		t.Fatal(err)
	}
	raw, digest, err := encodeRenderFailureRecord(record)
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	var projected map[string]any
	artifact, ok := contractreporting.Index["contracts/reporting/render-failure-record.v1.schema.json"]
	if !ok || json.Unmarshal([]byte(artifact.JSON), &projected) != nil {
		t.Fatal("failure record schema is not packaged")
	}
	location := "https://reporting.invalid/render-failure-record.v1.schema.json"
	if err := compiler.AddResource(location, projected); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(location)
	if err != nil {
		t.Fatal(err)
	}
	var emitted any
	if err := json.Unmarshal(raw, &emitted); err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(emitted); err != nil {
		t.Fatal("emitted failure violates packaged schema", err)
	}

	t.Run("pinned-reference failure projection", func(t *testing.T) {
		releaseID, snapshotID := "00000000-0000-0000-0000-000000000004", "00000000-0000-0000-0000-000000000005"
		summary, err := encodeFailedRenderSummary(&releaseID, &snapshotID, record.AdmittedAt, RenderFailure{Stage: "export_model_materialization", FailureCode: "export_model_invalid", ReasonCode: "required_reference_pack_unavailable", SafeDetails: map[string]any{}})
		if err != nil {
			t.Fatal(err)
		}
		bound := record
		bound.Summary = summary
		bound.InputBinding = &renderFailureInputBinding{AdmissionSHA256: strings.Repeat("a", 64)}
		encoded, digest, err := encodeRenderFailureRecord(bound)
		if err != nil {
			t.Fatal(err)
		}
		var value any
		if err := json.Unmarshal(encoded, &value); err != nil {
			t.Fatal(err)
		}
		if err := schema.Validate(value); err != nil {
			t.Fatal("pinned-reference failure violates schema", err)
		}
		if _, _, err := decodeRenderFailureRecord(encoded, digest, retained); err != nil {
			t.Fatal("bound failure rejected", err)
		}
	})

	if _, got, err := decodeRenderFailureRecord(raw, digest, retained); err != nil || got.ReleaseID != nil || got.SnapshotID != nil {
		t.Fatal("unbound failure rejected or invented identity", err)
	}
	for name, mutate := range map[string]func(map[string]any){
		"job":                      func(v map[string]any) { v["job_id"] = actor },
		"kind":                     func(v map[string]any) { v["job_kind"] = CompositionPreviewJobKind },
		"incident":                 func(v map[string]any) { v["incident_id"] = actor },
		"actor":                    func(v map[string]any) { v["actor_user_id"] = incident.String() },
		"request":                  func(v map[string]any) { v["request_sha256"] = strings.Repeat("b", 64) },
		"time":                     func(v map[string]any) { v["admitted_at"] = "2026-11-03T12:00:00.000000Z" },
		"missing nullable binding": func(v map[string]any) { delete(v, "input_binding") },
		"fabricated binding": func(v map[string]any) {
			v["input_binding"] = map[string]any{"admission_sha256": strings.Repeat("a", 64), "preparation_sha256": nil}
		},
		"invented snapshot":      func(v map[string]any) { v["summary"].(map[string]any)["snapshot_id"] = incident.String() },
		"unknown member":         func(v map[string]any) { v["raw_error"] = "private-diagnostic-sentinel" },
		"unknown summary member": func(v map[string]any) { v["summary"].(map[string]any)["raw_error"] = "private-diagnostic-sentinel" },
		"known wrong integrity stage": func(v map[string]any) {
			summary := v["summary"].(map[string]any)
			summary["terminal_stage"] = "redaction"
			summary["first_failure"].(map[string]any)["stage"] = "redaction"
			summary["issues"].([]any)[0].(map[string]any)["stage"] = "redaction"
		},
		"unsupported schema": func(v map[string]any) { v["schema_id"] = "cartulary.reporting_render_failure_record.v0" },
	} {
		t.Run(name, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			mutate(value)
			altered, err := canonicalJSON(value)
			if err != nil {
				t.Fatal(err)
			}
			_, _, err = decodeRenderFailureRecord(altered, reportingObjectDigest(renderFailureRecordSchema, altered), retained)
			if err == nil || strings.Contains(err.Error(), "private-diagnostic-sentinel") {
				t.Fatal("rehashed substitution accepted or disclosed", err)
			}
		})
	}
	for _, altered := range [][]byte{nil, append(append([]byte(nil), raw...), '\n'), bytes.Repeat([]byte("x"), maxRenderFailureRecordBytes+1), []byte{0xff}} {
		if _, _, err := decodeRenderFailureRecord(altered, reportingObjectDigest(renderFailureRecordSchema, altered), retained); err == nil {
			t.Fatal("invalid byte boundary accepted")
		}
	}
	expired := retained
	expired.ExpiredAt = &at
	expired.RequestSHA256 = ""
	if _, _, err := decodeRenderFailureRecord(raw, digest, expired); err != nil {
		t.Fatal("original record lost at public-response expiry", err)
	}
	if _, err := bindRenderFailureRecord(expired, summary); err == nil {
		t.Fatal("expired admission reconstructed new failure evidence")
	}
	if _, _, err := decodeRenderFailureRecord(nil, "", expired); err == nil {
		t.Fatal("expired missing record reconstructed")
	}
}
