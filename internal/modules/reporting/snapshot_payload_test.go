package reporting

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func TestReportingFrozenSnapshotPayload(t *testing.T) {
	jobID := uuid.MustParse("00000000-0000-0000-0000-000000000410")
	incidentID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("00000000-0000-0000-0000-000000000412")
	at := time.Date(2026, 10, 14, 13, 0, 0, 123456000, time.UTC)
	// Independent Core source-boundary vector, not a digest from the decoder.
	boundary := json.RawMessage(`{"incident_id":"11111111-1111-4111-8111-111111111111","incident_version":1,"latest_change_set_id":null,"latest_change_set_created_at":null}`)
	watermark := "cartulary.source_boundary.v1:7118b27f199de626a091552abc20fb026c5f25f04ca303e0edc790f86ee82478"
	model, digest, err := freezeSnapshotCapture(IncidentMetadataSnapshot{ID: incidentID.String(), Title: "Admitted", Status: "active", Version: 1}, reportingResourceID(jobID, "snapshot").String(), at, watermark, reference_data.SetBinding{}, fixtureSnapshotSource(nil))
	if err != nil {
		t.Fatal(err)
	}
	payload := snapshotCreateJobPayload{
		IncidentID: incidentID.String(), ActorUserID: actorID.String(), ClientTxnID: "frozen-snapshot", SnapshotAt: at, JobAdmittedAt: at, RequestedBoundary: &watermark,
		SourceChangeSetHighWatermark: watermark, SourceBoundaryJSON: boundary, SourceCapture: model, SourceCaptureSHA256: digest,
	}
	raw, err := canonicalJSON(payload)
	if err != nil {
		t.Fatal(err)
	}
	decode := func(raw []byte) (snapshotCreateJobPayload, error) {
		return decodeSnapshotJobPayload(raw, jobID, incidentID, actorID, at)
	}
	if _, err := decode(raw); err != nil {
		t.Fatal("valid frozen snapshot rejected", err)
	}
	uncapturedRequest := payload
	uncapturedRequest.RequestedBoundary = nil
	uncapturedRequest.JobAdmittedAt = at.Add(-time.Second)
	omittedRaw, err := canonicalJSON(uncapturedRequest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeSnapshotJobPayload(omittedRaw, jobID, incidentID, actorID, uncapturedRequest.JobAdmittedAt); err != nil {
		t.Fatal("omitted boundary and distinct capture time were rejected", err)
	}
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, raw, "", "  "); err != nil {
		t.Fatal(err)
	}
	if _, err := decode(formatted.Bytes()); err != nil {
		t.Fatal("JSONB formatting changed admission", err)
	}
	var original map[string]json.RawMessage
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	mutate := func(key string, value any, remove bool) []byte {
		object := make(map[string]json.RawMessage, len(original))
		for k, v := range original {
			object[k] = v
		}
		if remove {
			delete(object, key)
		} else {
			object[key], err = json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
		}
		encoded, err := json.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	for key := range original {
		t.Run("required "+key, func(t *testing.T) {
			candidates := [][]byte{mutate(key, nil, true)}
			if key != "requested_boundary" {
				candidates = append(candidates, mutate(key, nil, false))
			}
			for _, candidate := range candidates {
				if _, err := decode(candidate); err == nil {
					t.Fatal("missing or null admission member was repaired")
				}
			}
		})
	}
	for name, candidate := range map[string][]byte{
		"unknown": mutate("future", true, false), "null": []byte(`null`), "array": []byte(`[]`),
		"duplicate":         append([]byte(`{"actor_user_id":"ignored",`), raw[1:]...),
		"trailing":          append(append([]byte(nil), raw...), []byte(`{}`)...),
		"nested unknown":    mutate("source_capture", json.RawMessage(strings.Replace(string(original["source_capture"]), `"schema_id":`, `"unknown":true,"schema_id":`, 1)), false),
		"retired model":     mutate("source_capture", json.RawMessage(strings.Replace(string(original["source_capture"]), snapshotCaptureSchemaID, "cartulary.reporting_source_capture.v1", 1)), false),
		"actor":             mutate("actor_user_id", incidentID.String(), false),
		"incident":          mutate("incident_id", actorID.String(), false),
		"admission time":    mutate("snapshot_at", at.Add(time.Second), false),
		"empty transaction": mutate("client_txn_id", "", false),
		"boundary token":    mutate("source_change_set_high_watermark", "changed", false),
		"boundary object":   mutate("source_boundary_json", json.RawMessage(`{}`), false),
		"boundary unknown":  mutate("source_boundary_json", json.RawMessage(strings.Replace(string(boundary), `"incident_version":1`, `"incident_version":1,"future":true`, 1)), false),
		"digest":            mutate("source_capture_sha256", strings.Repeat("0", 64), false),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decode(candidate); err == nil {
				t.Fatal("altered retained snapshot was admitted")
			}
		})
	}
	for _, identities := range [][3]uuid.UUID{{uuid.New(), incidentID, actorID}, {jobID, uuid.Nil, actorID}, {jobID, incidentID, uuid.Nil}} {
		if _, err := decodeSnapshotJobPayload(raw, identities[0], identities[1], identities[2], at); err == nil {
			t.Fatal("payload accepted under changed row identity")
		}
	}
	normalized, err := normalizeSnapshotRequest(incidentID, payload.ClientTxnID, &watermark)
	if err != nil {
		t.Fatal(err)
	}
	admission, err := jobs.NewExtensionJobAdmission(ProfileID, jobs.NewRouteIdempotencyKey("snapshots.create", actorID, incidentID.String(), payload.ClientTxnID), jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}, normalized)
	if err != nil {
		t.Fatal(err)
	}
	retained := jobs.RetainedExtensionJob{
		Resource: jobs.Resource{JobID: jobID.String(), Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}, AuthPolicy: jobs.AuthPolicyIncidentMembership, SubmittedByUserID: &payload.ActorUserID, SubmittedAt: at, Status: jobs.StatusSucceeded},
		JobKind:  SnapshotCreateJobKind, HandlerName: JobWorkerKind, OwnerProfileID: ProfileID,
		IdempotencyIdentity: admission.IdempotencyIdentity, RouteKey: admission.IdempotencyRouteKey, ScopeKey: admission.IdempotencyScopeKey, RequestSHA256: admission.NormalizedRequestSHA256,
	}
	if err := validateSnapshotJobAdmission(payload, jobID, retained); err != nil {
		t.Fatal("valid Jobs admission rejected", err)
	}
	// Transaction IDs are opaque. Preserve admitted whitespace rather than
	// silently normalizing it, then compare the exact identity with Jobs.
	padded, err := decode(mutate("client_txn_id", " frozen-snapshot ", false))
	if err != nil || padded.ClientTxnID != " frozen-snapshot " {
		t.Fatal("opaque transaction ID was normalized", err)
	}
	if err := validateSnapshotJobAdmission(padded, jobID, retained); err == nil {
		t.Fatal("different opaque transaction replaced Jobs admission")
	}
	for _, key := range []string{"kind", "handler", "owner", "policy", "scope", "incident", "actor", "time", "route", "scope key", "hash", "identity", "job"} {
		t.Run("Jobs "+key, func(t *testing.T) {
			changed := retained
			switch key {
			case "kind":
				changed.JobKind = ReleaseCreateJobKind
			case "handler":
				changed.HandlerName = "retired"
			case "owner":
				changed.OwnerProfileID = "other"
			case "policy":
				changed.Resource.AuthPolicy = "other"
			case "scope":
				changed.Resource.Scope.Kind = jobs.ScopeKindDeployment
			case "incident":
				changed.Resource.Scope.IncidentID = &actorID
			case "actor":
				changed.Resource.SubmittedByUserID = nil
			case "time":
				changed.Resource.SubmittedAt = at.Add(time.Second)
			case "route":
				changed.RouteKey = "releases.create"
			case "scope key":
				changed.ScopeKey = actorID.String()
			case "hash":
				changed.RequestSHA256 = strings.Repeat("0", 64)
			case "identity":
				changed.IdempotencyIdentity = json.RawMessage(`{}`)
			case "job":
				changed.Resource.JobID = actorID.String()
			}
			if err := validateSnapshotJobAdmission(payload, jobID, changed); err == nil {
				t.Fatal("changed Jobs admission accepted")
			}
		})
	}
	tombstone := retained
	expired := at.Add(48 * time.Hour)
	tombstone.ExpiredAt = &expired
	tombstone.IdempotencyIdentity = nil
	tombstone.RouteKey, tombstone.ScopeKey, tombstone.RequestSHA256 = "", "", ""
	if err := validateSnapshotJobAdmission(payload, jobID, tombstone); err != nil {
		t.Fatal("public response expiry invalidated retained owner history", err)
	}
	tombstone.Resource.Status = jobs.StatusRunning
	if err := validateSnapshotJobAdmission(payload, jobID, tombstone); err == nil {
		t.Fatal("expired pending work accepted")
	}
}
