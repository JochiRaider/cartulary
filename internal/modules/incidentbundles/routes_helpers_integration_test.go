package incidentbundles_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

const (
	incidentPortabilityProfileID = "incident_portability"
	incidentBundleExportJobKind  = "incident_portability.export_v1"
	incidentBundleExportedCode   = "incident_bundle_exported"
	incidentBundleImportedCode   = "incident_bundle_imported"
	incidentBundleZIPMediaType   = "application/zip"
)

type incidentBundleManifestMirror struct {
	BundleFormat                 string                     `json:"bundle_format"`
	BundleVersion                int                        `json:"bundle_version"`
	BundleID                     string                     `json:"bundle_id"`
	IncidentID                   string                     `json:"incident_id"`
	IncidentKey                  string                     `json:"incident_key"`
	ExportedAt                   string                     `json:"exported_at"`
	SourceChangeSetHighWatermark string                     `json:"source_change_set_high_watermark"`
	HistoryMode                  string                     `json:"history_mode"`
	BlobMode                     string                     `json:"blob_mode"`
	ReferencePackMode            string                     `json:"reference_pack_mode"`
	OptionalSections             []string                   `json:"optional_sections"`
	RequiredCapabilities         []string                   `json:"required_capabilities"`
	SigningKeyID                 *string                    `json:"signing_key_id,omitempty"`
	Files                        []incidentBundleFileMirror `json:"files"`
}

type incidentBundleFileMirror struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
	Required  bool   `json:"required"`
}

type seededIncidentBundlePortableState struct {
	BlobBytes                []byte
	BlobSHA                  string
	ObjectBlobID             string
	EvidenceRecordID         string
	HistoryHostRecordID      string
	PartyRecordID            string
	FindingArtifactRecordID  string
	QueryArtifactRecordID    string
	KeywordArtifactRecordID  string
	HandoffArtifactRecordID  string
	SavedViewID              string
	ReversibleChangeSetID    string
	NonReversibleChangeSetID string
}

func seedIncidentBundlePortableState(t testing.TB, harness *appsupport.ServerHarness, incidentID string, timelineRecordID string, actorUserID string) seededIncidentBundlePortableState {
	t.Helper()
	ctx := context.Background()
	incidentUUID := uuid.MustParse(incidentID)
	timelineUUID := uuid.MustParse(timelineRecordID)
	actorUUID := uuid.MustParse(actorUserID)
	if _, err := harness.DB.Exec(`
INSERT INTO record_tags (incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, 'ExtensionProfile Portability', 'extensionprofile portability', $3)
`, incidentID, timelineRecordID, actorUserID); err != nil {
		t.Fatalf("seed record tag: %v", err)
	}

	historyHostID := uuid.New()
	insertHostRecord(t, harness.DB, incidentUUID, historyHostID, actorUUID, "portable host before", "portable-host")
	envelopeTx, err := harness.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin portable host envelope normalization: %v", err)
	}
	defer func() { _ = envelopeTx.Rollback() }()
	if _, err := envelopeTx.Exec(`
UPDATE records
   SET created_at = '2026-05-25T16:59:00Z',
       updated_at = '2026-05-25T16:59:00Z'
 WHERE record_id = $1
`, historyHostID); err != nil {
		t.Fatalf("normalize portable host envelope time: %v", err)
	}
	if _, err := envelopeTx.Exec(`
UPDATE hosts
   SET created_at = '2026-05-25T16:59:00Z',
       updated_at = '2026-05-25T16:59:00Z'
 WHERE record_id = $1
`, historyHostID); err != nil {
		t.Fatalf("normalize portable host source time: %v", err)
	}
	if err := envelopeTx.Commit(); err != nil {
		t.Fatalf("commit portable host envelope normalization: %v", err)
	}

	identityID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'identity', $3, $3)
`, identityID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed identity envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO identities (
    record_id, incident_id, display_name, upn, email, sam_account_name,
    entity_origin, identity_state, row_version, created_at, updated_at,
    created_by_user_id, updated_by_user_id
)
SELECT
    envelope.record_id, envelope.incident_id, 'Portable Identity', 'portable.identity@example.test',
    'portable.identity@example.test', 'portable.identity', 'entity_import',
    'canonical', envelope.row_version, envelope.created_at, envelope.updated_at,
    envelope.created_by_user_id, envelope.updated_by_user_id
  FROM records AS envelope
 WHERE envelope.record_id = $1
   AND envelope.incident_id = $2
`, identityID, incidentUUID); err != nil {
		t.Fatalf("seed identity row: %v", err)
	}
	assessmentID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'assessment', $3, $3)
`, assessmentID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed assessment envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO assessments (
    record_id, incident_id, subject_record_id, subject_type,
    assessment_state, confidence_score, assessor_user_id, rationale
)
VALUES ($1, $2, $3, 'host', 'suspected', 70, $4, 'Portable assessment')
`, assessmentID, incidentUUID, historyHostID, actorUUID); err != nil {
		t.Fatalf("seed assessment row: %v", err)
	}

	indicatorID := uuid.New()
	indicatorTimestamp := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := harness.DB.Exec(`
INSERT INTO records (
    record_id, incident_id, record_type, created_at, created_by_user_id,
    updated_at, updated_by_user_id
)
VALUES ($1, $2, 'indicator', $3, $4, $3, $4)
`, indicatorID, incidentUUID, indicatorTimestamp, actorUUID); err != nil {
		t.Fatalf("seed indicator envelope: %v", err)
	}
	indicatortest.SeedSubtype(t, harness.DB, incidentUUID, indicatorID, "domain_name", "atomic", "portable.example.test")
	if _, err := harness.DB.Exec(`
INSERT INTO indicator_observations (
    incident_id, source_record_id, source_field_key, origin_kind, origin_locator,
    observed_text, parsed_indicator_type, normalized_candidate, resolution_status,
    resolved_indicator_record_id, row_version, created_by_user_id, resolved_by_user_id,
    resolved_at, resolution_method, deleted_at, deleted_by_user_id
)
VALUES ($1, $2, 'timeline.activity_synopsis_text', 'extraction', 'extension_profile', 'portable.example.test', 'domain_name', 'portable.example.test', 'resolved', $3, 2, $4, $4, now(), 'fixture', now(), $4)
`, incidentUUID, timelineUUID, indicatorID, actorUUID); err != nil {
		t.Fatalf("seed indicator observation: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO indicator_state_intervals (
    incident_id, indicator_record_id, lifecycle_state, valid_from, support_refs,
    row_version, created_by_user_id, deleted_at, deleted_by_user_id
)
VALUES ($1, $2, 'active', now(), '[]'::jsonb, 2, $3, now(), $3)
`, incidentUUID, indicatorID, actorUUID); err != nil {
		t.Fatalf("seed indicator lifecycle tombstone: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO entity_mentions (
    source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, row_version, ordinal,
    created_by_user_id, resolved_record_id, resolved_by_user_id, resolved_at, resolution_method
)
VALUES ($1, 'host', 'timeline.activity_synopsis_text', 'manual_entry', 'extension_profile', 'portable host', 'portable host', 'resolved', 1, 1, $2, $3, $2, now(), 'fixture')
`, timelineUUID, actorUUID, historyHostID); err != nil {
		t.Fatalf("seed entity mention: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key, provenance,
    owner_user_id, created_by_user_id, decided_at
)
VALUES ($1, $2, $3, 'observed_on_host', 'timeline.host_refs', 'manual', $4, $4, now())
`, incidentUUID, timelineUUID, historyHostID, actorUUID); err != nil {
		t.Fatalf("seed record link: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO timeline_time_conversion_profiles (
    incident_id, enabled, local_offset_minutes, local_label, updated_by_user_id
)
VALUES ($1, true, -300, 'America/New_York', $2)
`, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed timeline time conversion profile: %v", err)
	}
	partyRecordID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'party', $3, $3)
`, partyRecordID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed party envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO parties (
    record_id, incident_id, display_name, party_kind, organization_name,
    role_title, primary_email, timezone_name, external_ref, notes
)
VALUES ($1, $2, 'Portable Party', 'person', 'Cartulary IR', 'Incident lead', 'portable-party@example.test', 'America/New_York', 'party-1', 'portable party note')
`, partyRecordID, incidentUUID); err != nil {
		t.Fatalf("seed party row: %v", err)
	}
	taskRequestID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'task_request', $3, $3)
`, taskRequestID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed task-request envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO task_requests (
    record_id, incident_id, title, status, owner_user_id, priority,
    task_kind, workstream, requester_party_id
)
VALUES (
    $1, $2, 'Portable task request', 'open', $3, 'high',
    'request', 'portable', $4
)
`, taskRequestID, incidentUUID, actorUUID, partyRecordID); err != nil {
		t.Fatalf("seed task-request row: %v", err)
	}
	decisionID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'decision', $3, $3)
`, decisionID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed decision envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO decisions (
    record_id, incident_id, summary, status, owner_user_id,
    decision_type, decided_at, rationale
)
VALUES (
    $1, $2, 'Portable decision', 'approved', $3,
    'containment', '2026-05-25T17:02:00Z', 'Portable rationale'
)
`, decisionID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed decision row: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO entity_preserved_identifiers (
    incident_id, record_id, entity_type, identifier_type, raw_value,
    normalized_value, classification, created_by_user_id
)
VALUES ($1, $2, 'host', 'hostname', 'Portable-Host', 'portable-host', 'exact_match_reuse', $3)
`, incidentUUID, historyHostID, actorUUID); err != nil {
		t.Fatalf("seed entity preserved identifier: %v", err)
	}
	findingArtifactID := uuid.New()
	queryArtifactID := uuid.New()
	keywordArtifactID := uuid.New()
	handoffArtifactID := uuid.New()
	seedPortableArtifactRecord(t, harness.DB, incidentUUID, findingArtifactID, actorUUID, "finding", "Portable finding")
	if _, err := harness.DB.Exec(`
INSERT INTO artifact_findings (
    record_id, incident_id, kind, statement, state, confidence_score, owner_user_id
)
VALUES ($1, $2, 'finding', 'Portable finding statement', 'open', 87, $3)
`, findingArtifactID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed artifact finding: %v", err)
	}
	seedPortableArtifactRecord(t, harness.DB, incidentUUID, queryArtifactID, actorUUID, "investigative_query", "Portable investigative query")
	if _, err := harness.DB.Exec(`
INSERT INTO artifact_investigative_queries (
    record_id, incident_id, query_id, platform, purpose, query_text, created_by_user_id
)
VALUES ($1, $2, 'portable-query-1', 'kusto', 'Find portable events', 'SecurityEvent | take 10', $3)
`, queryArtifactID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed artifact investigative query: %v", err)
	}
	seedPortableArtifactRecord(t, harness.DB, incidentUUID, keywordArtifactID, actorUUID, "forensic_keyword", "Portable forensic keyword")
	if _, err := harness.DB.Exec(`
INSERT INTO artifact_forensic_keywords (
    record_id, incident_id, keyword_id, pattern, reason, match_mode, case_sensitive
)
VALUES ($1, $2, 'portable-keyword-1', 'PortableKeyword', 'Portable keyword reason', 'literal', true)
`, keywordArtifactID, incidentUUID); err != nil {
		t.Fatalf("seed artifact forensic keyword: %v", err)
	}
	seedPortableArtifactRecord(t, harness.DB, incidentUUID, handoffArtifactID, actorUUID, "handoff", "Portable handoff")
	if _, err := harness.DB.Exec(`
UPDATE artifacts
   SET handoff_id = 'portable-handoff-1',
       outgoing_owner_user_id = $3,
       incoming_owner_user_id = $3,
       current_state_summary = 'Portable handoff state'
 WHERE record_id = $1
   AND incident_id = $2
`, handoffArtifactID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed artifact handoff facts: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO handoff_risk_refs (
    incident_id, handoff_record_id, risk_ref_text, normalized_risk_ref_text, created_by_user_id
)
VALUES ($1, $2, 'Portable Risk', 'portable risk', $3)
`, incidentUUID, handoffArtifactID, actorUUID); err != nil {
		t.Fatalf("seed handoff risk ref: %v", err)
	}
	savedViewID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO saved_views (
    saved_view_id, incident_id, view_schema_id, scope, display_name,
    query_json, layout_json, owner_user_id
)
VALUES (
    $1,
    $2,
    $4,
    'private',
    'Portable saved view',
    '{"filters":[{"field_key":"timeline.tags","op":"contains_any","arg":{"values":["portable"]}}]}'::jsonb,
    '{}'::jsonb,
    $3
)
`, savedViewID, incidentUUID, actorUUID, timeline.TimelineViewSchemaID); err != nil {
		t.Fatalf("seed saved view: %v", err)
	}

	reversibleChangeSetID := uuid.New()
	seedPortableRollbackHostPatch(t, harness.DB, incidentUUID, historyHostID, actorUUID, reversibleChangeSetID, time.Date(2026, 5, 25, 17, 0, 0, 0, time.UTC), "portable host before", "portable host after")
	nonReversibleChangeSetID := uuid.New()
	seedPortableRecordTagCreateHistory(t, harness.DB, incidentUUID, historyHostID, actorUUID, nonReversibleChangeSetID, time.Date(2026, 5, 25, 17, 1, 0, 0, time.UTC))

	blobBytes := []byte("extension_profile incident bundle blob\n")
	sum := sha256.Sum256(blobBytes)
	blobSHA := hex.EncodeToString(sum[:])
	sourceStorageKey := "extension_profile/source/" + incidentID + "/" + blobSHA
	if err := harness.ObjectStore.PutObject(ctx, sourceStorageKey, bytes.NewReader(blobBytes), int64(len(blobBytes)), "text/plain"); err != nil {
		t.Fatalf("seed source object bytes: %v", err)
	}

	var objectBlobID string
	if err := harness.DB.QueryRow(`
INSERT INTO object_blobs (
    incident_id,
    created_by_user_id,
    storage_key,
    upload_state,
    byte_size,
    filename_hint,
    content_type_hint,
    expected_sha256_hex,
    observed_size,
    observed_content_type,
    observed_sha256_hex,
    target_expires_at,
    pending_expires_at,
    finalized_at
)
VALUES ($1, $2, $3, 'available', $4, 'extension_profile.txt', 'text/plain', $5, $4, 'text/plain', $5, now() + interval '1 hour', now() + interval '1 hour', now())
RETURNING object_blob_id
`, incidentID, actorUserID, sourceStorageKey, len(blobBytes), blobSHA).Scan(&objectBlobID); err != nil {
		t.Fatalf("seed object blob row: %v", err)
	}

	var evidenceRecordID string
	if err := harness.DB.QueryRow(`
INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, 'evidence', $2, $2)
RETURNING record_id
`, incidentID, actorUserID).Scan(&evidenceRecordID); err != nil {
		t.Fatalf("seed evidence record envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO evidence (
    record_id,
    incident_id,
    title,
    lifecycle_state,
    requested_at,
    received_at,
    storage_ref,
    blob_hash,
    upload_state,
    object_blob_id
)
VALUES ($1, $2, 'Portable evidence', 'received', now(), now(), $3, $4, 'available', $5)
`, evidenceRecordID, incidentID, sourceStorageKey, blobSHA, objectBlobID); err != nil {
		t.Fatalf("seed evidence row: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO evidence_custody_events (
    incident_id,
    evidence_record_id,
    custody_event_type,
    actor_user_id,
    location_text,
    note
)
VALUES ($1, $2, 'made_available', $3, 'source deployment', 'seeded portable custody event')
`, incidentID, evidenceRecordID, actorUserID); err != nil {
		t.Fatalf("seed evidence custody event: %v", err)
	}
	return seededIncidentBundlePortableState{
		BlobBytes:                blobBytes,
		BlobSHA:                  blobSHA,
		ObjectBlobID:             objectBlobID,
		EvidenceRecordID:         evidenceRecordID,
		HistoryHostRecordID:      historyHostID.String(),
		PartyRecordID:            partyRecordID.String(),
		FindingArtifactRecordID:  findingArtifactID.String(),
		QueryArtifactRecordID:    queryArtifactID.String(),
		KeywordArtifactRecordID:  keywordArtifactID.String(),
		HandoffArtifactRecordID:  handoffArtifactID.String(),
		SavedViewID:              savedViewID.String(),
		ReversibleChangeSetID:    reversibleChangeSetID.String(),
		NonReversibleChangeSetID: nonReversibleChangeSetID.String(),
	}
}

func seedPortableArtifactRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, artifactType string, title string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'artifact', $3, $3)
`, recordID, incidentID, actorID); err != nil {
		t.Fatalf("seed artifact envelope: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO artifacts (
    record_id, incident_id, artifact_type, title, body, timestamp_utc, created_by_user_id
)
VALUES ($1, $2, $3, $4, 'portable artifact body', $5, $6)
`, recordID, incidentID, artifactType, title, time.Date(2026, 5, 25, 16, 30, 0, 0, time.UTC), actorID); err != nil {
		t.Fatalf("seed artifact row: %v", err)
	}
}

func insertHostRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, displayName string, hostname string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'host', $3, $3)
`, recordID, incidentID, actorID); err != nil {
		t.Fatalf("seed host record envelope: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state,
    row_version, created_at, updated_at, created_by_user_id, updated_by_user_id
)
SELECT envelope.record_id, envelope.incident_id, $3, $4, 'canonical',
       envelope.row_version, envelope.created_at, envelope.updated_at,
       envelope.created_by_user_id, envelope.updated_by_user_id
  FROM records AS envelope
 WHERE envelope.record_id = $1
   AND envelope.incident_id = $2
`, recordID, incidentID, displayName, hostname); err != nil {
		t.Fatalf("seed host row: %v", err)
	}
}

func seedPortableRollbackHostPatch(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, createdAt time.Time, beforeName string, afterName string) {
	t.Helper()
	ctx := context.Background()
	beforeRecord := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "record_type": "host", "row_version": 1}
	afterRecord := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "record_type": "host", "row_version": 2}
	beforeSource := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "display_name": beforeName, "hostname": "portable-host", "host_state": "canonical", "row_version": 1}
	afterSource := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "display_name": afterName, "hostname": "portable-host", "host_state": "canonical", "row_version": 2}
	beforeValue := map[string]any{"snapshot_schema_id": "cartulary.revisions.snapshot.host.v1", "record": beforeRecord, "source": beforeSource}
	afterValue := map[string]any{"snapshot_schema_id": "cartulary.revisions.snapshot.host.v1", "record": afterRecord, "source": afterSource}
	envelopeTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin portable rollback host transition: %v", err)
	}
	defer func() { _ = envelopeTx.Rollback() }()
	if _, err := envelopeTx.ExecContext(ctx, `
UPDATE records
   SET row_version = 2,
       updated_at = $4,
       updated_by_user_id = $3
 WHERE record_id = $1
   AND incident_id = $2
`, recordID, incidentID, actorID, createdAt); err != nil {
		t.Fatalf("advance portable rollback host envelope: %v", err)
	}
	if _, err := envelopeTx.ExecContext(ctx, `
UPDATE hosts
   SET display_name = $3,
       row_version = 2,
       updated_at = $4,
       updated_by_user_id = $5
 WHERE record_id = $1
   AND incident_id = $2
	`, recordID, incidentID, afterName, createdAt, actorID); err != nil {
		t.Fatalf("advance portable rollback host source: %v", err)
	}
	if err := envelopeTx.Commit(); err != nil {
		t.Fatalf("commit portable rollback host transition: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, reason, client_txn_id, request_id, created_at)
VALUES ($1, $2, $3, 'workbook.records.patch', 'portable rollback seed', 'txn-portable-host-patch', 'req-portable-host-patch', $4)
`, changeSetID, incidentID, actorID, createdAt); err != nil {
		t.Fatalf("seed portable rollback change set: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_set_mutations (
    change_set_id, sequence_no, target_kind, target_id, operation_kind,
    before_version_id, after_version_id, before_value, after_value,
    history_record_ids, history_entry_record_ids
)
VALUES ($1, 1, 'host', $2::text, 'field_update', $3, $4, $5, $6, ARRAY[$2::uuid], ARRAY[$2::uuid])
`, changeSetID, recordID.String(), "host:"+recordID.String()+":1", "host:"+recordID.String()+":2", jsonRaw(t, beforeValue), jsonRaw(t, afterValue)); err != nil {
		t.Fatalf("seed portable rollback mutation: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_revisions (change_set_id, record_id, row_version, before_json, after_json, created_at)
VALUES ($1, $2, 2, $3, $4, $5)
`, changeSetID, recordID, jsonRaw(t, beforeValue), jsonRaw(t, afterValue), createdAt); err != nil {
		t.Fatalf("seed portable rollback revision: %v", err)
	}
}

func seedPortableRecordTagCreateHistory(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, createdAt time.Time) {
	t.Helper()
	recordTagID := uuid.New()
	afterValue := map[string]any{
		"record_tag_id": recordTagID.String(), "incident_id": incidentID.String(),
		"record_id": recordID.String(), "tag_name": "ExtensionProfile History",
		"normalized_tag_name": "extensionprofile history", "created_by_user_id": actorID.String(),
		"created_at": createdAt.UTC().Format(time.RFC3339Nano), "updated_at": createdAt.UTC().Format(time.RFC3339Nano),
		"deleted_at": nil, "deleted_by_user_id": nil,
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id, created_at, updated_at)
VALUES ($1, $2, $3, 'ExtensionProfile History', 'extensionprofile history', $4, $5, $5)
`, recordTagID, incidentID, recordID, actorID, createdAt); err != nil {
		t.Fatalf("seed portable history record tag: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, reason, client_txn_id, request_id, created_at)
VALUES ($1, $2, $3, 'records.tags.create', 'portable tag seed', 'txn-portable-tag-create', 'req-portable-tag-create', $4)
`, changeSetID, incidentID, actorID, createdAt); err != nil {
		t.Fatalf("seed portable record-tag change set: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_set_mutations (
    change_set_id, sequence_no, target_kind, target_id, operation_kind,
    before_value, after_value, history_record_ids, history_entry_record_ids
)
VALUES ($1, 1, 'record_tag', $2, 'create', NULL, $3, ARRAY[$4::uuid], ARRAY[$4::uuid])
`, changeSetID, "record_tag:"+recordID.String()+":"+recordTagID.String(), jsonRaw(t, afterValue), recordID); err != nil {
		t.Fatalf("seed portable record-tag mutation: %v", err)
	}
}

func postExport(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, body map[string]any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/export", body,
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func postImport(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, metadata string, file []byte, filename string) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	metadataPart, err := writer.CreatePart(textprotoMIMEHeader(map[string]string{
		"Content-Disposition": `form-data; name="metadata"`,
		"Content-Type":        "application/json; charset=utf-8",
	}))
	if err != nil {
		t.Fatalf("create metadata part: %v", err)
	}
	if _, err := io.WriteString(metadataPart, metadata); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	filePart, err := writer.CreatePart(textprotoMIMEHeader(map[string]string{
		"Content-Disposition": `form-data; name="file"; filename="` + filename + `"`,
		"Content-Type":        incidentBundleZIPMediaType,
	}))
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := filePart.Write(file); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/import", &body)
	if err != nil {
		t.Fatalf("create import request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(login.SessionCookie)
	req.AddCookie(login.CSRFCookie)
	req.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
	return httptestx.Do(t, http.DefaultClient, req)
}

func waitJob(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	var lastJob map[string]any
	for time.Now().Before(deadline) {
		resp := httptestx.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
		job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		lastJob = job
		status := job["status"].(string)
		switch status {
		case "succeeded":
			return job
		case "failed", "canceled":
			encoded, _ := json.MarshalIndent(job, "", "  ")
			t.Fatalf("job reached terminal non-success state: %s", encoded)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not finish: %#v", jobID, lastJob)
	return nil
}

func waitJobWithStatus(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, jobID string, wantStatus string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp := httptestx.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
		job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		status := job["status"].(string)
		if status == wantStatus {
			return job
		}
		switch status {
		case "succeeded", "failed", "canceled":
			encoded, _ := json.MarshalIndent(job, "", "  ")
			t.Fatalf("job reached terminal status %q while waiting for %q: %s", status, wantStatus, encoded)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not reach status %q", jobID, wantStatus)
	return nil
}

func waitFailedJob(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(40 * time.Second)
	var lastJob map[string]any
	for time.Now().Before(deadline) {
		resp := httptestx.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
		job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		lastJob = job
		status := job["status"].(string)
		switch status {
		case "failed":
			return job
		case "succeeded", "canceled":
			encoded, _ := json.MarshalIndent(job, "", "  ")
			t.Fatalf("job reached unexpected terminal state: %s", encoded)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not fail; last observed state: %#v", jobID, lastJob)
	return nil
}

func corruptZipMember(t testing.TB, bundle []byte, memberPath string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip for corruption: %v", err)
	}
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	found := false
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		if member.Name == memberPath {
			data = append(data, []byte("corrupt\n")...)
			found = true
		}
		w, err := writer.Create(member.Name)
		if err != nil {
			t.Fatalf("create corrupt member %s: %v", member.Name, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write corrupt member %s: %v", member.Name, err)
		}
	}
	if !found {
		t.Fatalf("zip member %s not found", memberPath)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close corrupt zip: %v", err)
	}
	return buf.Bytes()
}

func zipMemberBytes(t testing.TB, bundle []byte, memberPath string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, member := range reader.File {
		if member.Name != memberPath {
			continue
		}
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", memberPath, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", memberPath, err)
		}
		return data
	}
	t.Fatalf("zip member %s not found", memberPath)
	return nil
}

func firstZipMemberWithPrefix(t testing.TB, bundle []byte, prefix string) string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, member := range reader.File {
		if strings.HasPrefix(member.Name, prefix) {
			return member.Name
		}
	}
	t.Fatalf("zip member with prefix %s not found", prefix)
	return ""
}

func removeZipMember(t testing.TB, bundle []byte, memberPath string) []byte {
	t.Helper()
	return rewriteZipMembers(t, bundle, func(path string, data []byte) ([]byte, bool) {
		if path == memberPath {
			return nil, false
		}
		return data, true
	})
}

func replaceZipMember(t testing.TB, bundle []byte, memberPath string, replace func([]byte) []byte) []byte {
	t.Helper()
	found := false
	rewritten := rewriteZipMembers(t, bundle, func(path string, data []byte) ([]byte, bool) {
		if path != memberPath {
			return data, true
		}
		found = true
		return replace(data), true
	})
	if !found {
		t.Fatalf("zip member %s not found", memberPath)
	}
	return rewritten
}

func appendZipMember(t testing.TB, bundle []byte, memberPath string, payload []byte) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		w, err := writer.Create(member.Name)
		if err != nil {
			t.Fatalf("create member %s: %v", member.Name, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write member %s: %v", member.Name, err)
		}
	}
	w, err := writer.Create(memberPath)
	if err != nil {
		t.Fatalf("create appended member %s: %v", memberPath, err)
	}
	if _, err := w.Write(payload); err != nil {
		t.Fatalf("write appended member %s: %v", memberPath, err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func appendZipDirectoryMembers(t testing.TB, bundle []byte, directories ...string) []byte {
	t.Helper()
	result := bundle
	for _, directory := range directories {
		result = appendZipMember(t, result, directory, nil)
	}
	return result
}

func replaceZipMemberAndChecksum(t testing.TB, bundle []byte, memberPath string, replacement []byte) []byte {
	t.Helper()
	replacementSHA := hashHexBytes(replacement)
	rewritten := replaceZipMember(t, bundle, memberPath, func([]byte) []byte {
		return replacement
	})
	return replaceZipMember(t, rewritten, "integrity/checksums.sha256", func(original []byte) []byte {
		lines := strings.Split(strings.TrimSpace(string(original)), "\n")
		for idx, line := range lines {
			if strings.HasSuffix(line, "  "+memberPath) {
				lines[idx] = replacementSHA + "  " + memberPath
			}
		}
		return []byte(strings.Join(lines, "\n") + "\n")
	})
}

func replaceStructuredBundleMember(
	t testing.TB,
	bundle []byte,
	memberPath string,
	replacement []byte,
) []byte {
	t.Helper()
	files := zipMemberMap(t, bundle)
	files[memberPath] = append([]byte(nil), replacement...)
	var manifest incidentBundleManifestMirror
	if err := json.Unmarshal(files["manifest.json"], &manifest); err != nil {
		t.Fatalf("decode manifest for structured replacement: %v", err)
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		if path == "manifest.json" || strings.HasPrefix(path, "integrity/") {
			continue
		}
		paths = append(paths, path)
	}
	slices.Sort(paths)
	manifest.Files = make([]incidentBundleFileMirror, 0, len(paths))
	checksumLines := make([]string, 0, len(paths))
	for _, path := range paths {
		digest := hashHexBytes(files[path])
		manifest.Files = append(manifest.Files, incidentBundleFileMirror{
			Path:      path,
			SHA256:    "sha256:" + digest,
			SizeBytes: int64(len(files[path])),
			Required:  !strings.HasPrefix(path, "ext/"),
		})
		checksumLines = append(checksumLines, digest+"  "+path)
	}
	sourceBoundary, err := json.Marshal(manifest.Files)
	if err != nil {
		t.Fatalf("encode structured replacement boundary: %v", err)
	}
	manifest.SourceChangeSetHighWatermark = "cartulary.source_boundary.v1:" +
		hashHexBytes(sourceBoundary)
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode structured replacement manifest: %v", err)
	}
	files["manifest.json"] = append(manifestBytes, '\n')
	files["integrity/checksums.sha256"] = []byte(
		strings.Join(checksumLines, "\n") + "\n",
	)
	return writeZipMemberMap(t, files)
}

func encodeNDJSONRows(t testing.TB, rows []map[string]any) []byte {
	t.Helper()
	var payload bytes.Buffer
	for _, row := range rows {
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatalf("encode NDJSON row: %v", err)
		}
		payload.Write(encoded)
		payload.WriteByte('\n')
	}
	return payload.Bytes()
}

func mapsClone(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func zipMemberMap(t testing.TB, bundle []byte) map[string][]byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip member map: %v", err)
	}
	files := make(map[string][]byte, len(reader.File))
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open zip member %s: %v", member.Name, err)
		}
		payload, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read zip member %s: %v", member.Name, err)
		}
		files[member.Name] = payload
	}
	return files
}

func writeZipMemberMap(t testing.TB, files map[string][]byte) []byte {
	t.Helper()
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, path := range paths {
		member, err := writer.Create(path)
		if err != nil {
			t.Fatalf("create zip member %s: %v", path, err)
		}
		if _, err := member.Write(files[path]); err != nil {
			t.Fatalf("write zip member %s: %v", path, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return output.Bytes()
}

func decodeNDJSONRows(t testing.TB, payload []byte) []map[string]any {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(payload), []byte("\n"))
	rows := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal(line, &row); err != nil {
			t.Fatalf("decode NDJSON row: %v", err)
		}
		rows = append(rows, row)
	}
	return rows
}

func rewriteZipMembers(t testing.TB, bundle []byte, transform func(path string, data []byte) ([]byte, bool)) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		nextData, keep := transform(member.Name, data)
		if !keep {
			continue
		}
		w, err := writer.Create(member.Name)
		if err != nil {
			t.Fatalf("create member %s: %v", member.Name, err)
		}
		if _, err := w.Write(nextData); err != nil {
			t.Fatalf("write member %s: %v", member.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close rewritten zip: %v", err)
	}
	return buf.Bytes()
}

func postRollback(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, recordID string, body map[string]any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/records/"+recordID+"/rollback", body,
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func getRecordHistoryItems(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, recordID string) []any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/records/"+recordID+"/history", nil, httptestx.WithCookies(login.SessionCookie))
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("history items missing: %#v", data)
	}
	return items
}

func requireHistoryItemForChangeSet(t testing.TB, items []any, changeSetID string) map[string]any {
	t.Helper()
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if item["change_set_id"] == changeSetID {
			return item
		}
	}
	t.Fatalf("history item for change_set_id=%s not found in %#v", changeSetID, items)
	return nil
}

func textprotoMIMEHeader(values map[string]string) textproto.MIMEHeader {
	header := textproto.MIMEHeader{}
	for key, value := range values {
		header.Set(key, value)
	}
	return header
}

func countRows(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

type importFinalizationSideEffects struct {
	MembershipRows           int
	DefaultPreferenceRows    int
	UserPreferenceRows       int
	MembershipAuditRows      int
	MembershipProjectionRows int
}

func (s importFinalizationSideEffects) equal(other importFinalizationSideEffects) bool {
	return s == other
}

func snapshotImportFinalizationSideEffects(t testing.TB, db *sql.DB, incidentID string, userID string) importFinalizationSideEffects {
	t.Helper()
	return importFinalizationSideEffects{
		MembershipRows: countRows(t, db, `
SELECT count(*)
  FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
   AND role = 'admin'
   AND added_by_user_id = $2
   AND updated_by_user_id = $2
   AND membership_version = 1
`, incidentID, userID),
		DefaultPreferenceRows: countRows(t, db, `
SELECT count(*)
  FROM incident_workbook_preferences
 WHERE incident_id = $1
   AND default_sheet_ref IS NULL
   AND updated_by_user_id = $2
`, incidentID, userID),
		UserPreferenceRows: countRows(t, db, `
SELECT count(*)
  FROM user_workbook_preferences
 WHERE incident_id = $1
   AND user_id = $2
   AND home_sheet_ref IS NULL
`, incidentID, userID),
		MembershipAuditRows: countRows(t, db, `
SELECT count(*)
  FROM deployment_admin_audit_events
 WHERE incident_id = $1
   AND actor_user_id = $2
   AND target_user_id = $2
   AND event_source = 'incidents'
   AND event_kind = 'incident_membership_created'
`, incidentID, userID),
		MembershipProjectionRows: countRows(t, db, `
SELECT count(*)
  FROM administrative_audit_projections
 WHERE scope_kind = 'incident'
   AND scope_id = $1
   AND actor_user_id = $2
   AND action_code = 'membership_created'
`, incidentID, userID),
	}
}

func startIsolatedIncidentBundleServer(t testing.TB, runtime *appsupport.Runtime, prefix string) *appsupport.ServerHarness {
	t.Helper()
	return startIsolatedIncidentBundleServerWithEnv(t, runtime, prefix, nil)
}

func startIsolatedIncidentBundleServerWithEnv(t testing.TB, runtime *appsupport.Runtime, prefix string, extraEnv map[string]string) *appsupport.ServerHarness {
	t.Helper()
	testDB := runtime.Postgres.PrepareIsolatedDatabaseT(t, prefix)
	bucket := runtime.S3.BootstrapBucketT(t, prefix)
	env := runtime.S3.Env(bucket)
	for key, value := range extraEnv {
		env[key] = value
	}
	store, err := objectstore.Setup(context.Background(), objectstore.Settings{
		BindingKind: "managed_service",
		Endpoint:    runtime.S3.Endpoint,
		AccessKey:   runtime.S3.AccessKey,
		SecretKey:   runtime.S3.SecretKey,
		Secure:      runtime.S3.Secure,
		Bucket:      bucket,
	}, objectstore.Instrumentation{})
	if err != nil {
		t.Fatalf("open isolated target object store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:        prefix,
		Database:      testDB,
		Env:           env,
		ObjectStore:   store,
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
}

func compareSourceTargetCount(t testing.TB, source *sql.DB, target *sql.DB, query string, incidentID string, label string) {
	t.Helper()
	sourceCount := countRows(t, source, query, incidentID)
	targetCount := countRows(t, target, query, incidentID)
	if targetCount != sourceCount {
		t.Fatalf("imported %s mismatch: target=%d source=%d", label, targetCount, sourceCount)
	}
}

func stringScalar(t testing.TB, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var value string
	if err := db.QueryRow(query, args...).Scan(&value); err != nil {
		t.Fatalf("query scalar: %v", err)
	}
	return value
}

func exportedBundleTestPath(t testing.TB, server *httptestx.Server, rawReference string) string {
	t.Helper()
	reference, err := incidentbundles.ParseBundleStorageRef(rawReference)
	if err != nil {
		t.Fatalf("database export storage reference %q is invalid: %v", rawReference, err)
	}
	if filepath.IsAbs(rawReference) || strings.Contains(rawReference, server.Config.Roots.ExportOutputs.Path) {
		t.Fatalf("database export storage reference disclosed a host root: %q", rawReference)
	}
	return filepath.Join(server.Config.Roots.ExportOutputs.Path, filepath.FromSlash(reference.String()))
}

func stagedBundleTestPath(t testing.TB, server *httptestx.Server, rawReference string) string {
	t.Helper()
	reference, err := incidentbundles.ParseBundleStagingRef(rawReference)
	if err != nil {
		t.Fatalf("database staging reference %q is invalid: %v", rawReference, err)
	}
	if filepath.IsAbs(rawReference) || strings.Contains(rawReference, server.Config.Roots.TemporaryWork.Path) {
		t.Fatalf("database staging reference disclosed a host root: %q", rawReference)
	}
	return filepath.Join(server.Config.Roots.TemporaryWork.Path, filepath.FromSlash(reference.String()))
}

func jsonRaw(t testing.TB, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fixture json: %v", err)
	}
	return payload
}

func exportBundleBytes(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, incidentID string, clientTxnID string) []byte {
	t.Helper()
	job := httptestx.RequireSuccessEnvelope(t, postExport(t, harness.Server, login, map[string]any{
		"incident_id":   incidentID,
		"client_txn_id": clientTxnID,
	}), http.StatusAccepted)["data"].(map[string]any)
	terminal := waitJob(t, harness.Server, login, job["job_id"].(string))
	ref := terminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)
	storageRef := stringScalar(t, harness.DB, `SELECT bundle_storage_ref FROM incident_bundle_exports WHERE bundle_id = $1`, ref["id"].(string))
	bundleBytes, err := os.ReadFile(exportedBundleTestPath(t, harness.Server, storageRef))
	if err != nil {
		t.Fatalf("read exported bundle: %v", err)
	}
	return bundleBytes
}

func importBundleAndWait(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, bundle []byte, clientTxnID string) map[string]any {
	t.Helper()
	resp := postImport(t, server, login, `{"client_txn_id":"`+clientTxnID+`"}`, bundle, "bundle.zip")
	job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	return waitJob(t, server, login, job["job_id"].(string))
}

func requireIncidentPortabilityProof(
	t testing.TB,
	db *sql.DB,
	jobID string,
	operationKind string,
) {
	t.Helper()
	var ownerProfileID string
	var actualOperationKind string
	var finalCommitID string
	var terminalCode string
	if err := db.QueryRow(`
SELECT owner_profile_id, operation_kind, final_commit_id,
       terminal_result->>'code'
  FROM extension_job_commit_proofs
 WHERE job_id::text = $1
`, jobID).Scan(
		&ownerProfileID,
		&actualOperationKind,
		&finalCommitID,
		&terminalCode,
	); err != nil {
		t.Fatalf("read Incident Portability proof for job %s: %v", jobID, err)
	}
	if ownerProfileID != incidentPortabilityProfileID ||
		actualOperationKind != operationKind ||
		finalCommitID == "" ||
		terminalCode == "" {
		t.Fatalf(
			"unexpected Incident Portability proof: owner=%q operation=%q commit=%q code=%q",
			ownerProfileID,
			actualOperationKind,
			finalCommitID,
			terminalCode,
		)
	}
}

func requireFailedJobReason(t testing.TB, job map[string]any, wantCode string, wantReason string) {
	t.Helper()
	errorSummary := job["error_summary"].(map[string]any)
	if errorSummary["code"] != wantCode {
		t.Fatalf("failed job code mismatch: got %#v want %s", errorSummary, wantCode)
	}
	details := errorSummary["details"].(map[string]any)
	if details["reason_code"] != wantReason {
		t.Fatalf("failed job reason mismatch: got %#v want %s", details, wantReason)
	}
}

func requireTimelineQueryRow(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, incidentID string, recordID string) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query", map[string]any{}, httptestx.WithCookies(login.SessionCookie))
	rows := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)["rows"].([]any)
	for _, raw := range rows {
		row := raw.(map[string]any)
		if row["record_id"] == recordID {
			return row
		}
	}
	t.Fatalf("timeline query row %s not found in %#v", recordID, rows)
	return nil
}

func timelineCellValue(t testing.TB, row map[string]any, fieldKey string) any {
	t.Helper()
	cells := row["cells"].(map[string]any)
	cell := cells[fieldKey].(map[string]any)
	return cell["value"]
}

func assertImportFailureLeavesState(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, incidentID string, clientTxnID string, bundle []byte, wantReason string) map[string]any {
	t.Helper()
	before := snapshotImportFailureState(t, harness, incidentID)
	resp := postImport(t, harness.Server, login, `{"client_txn_id":"`+clientTxnID+`"}`, bundle, "bundle.zip")
	job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	terminal := waitFailedJob(t, harness.Server, login, job["job_id"].(string))
	if wantReason == "extension_capability_not_supported" {
		errorSummary := terminal["error_summary"].(map[string]any)
		if errorSummary["code"] != "extension_capability_not_supported" {
			t.Fatalf("capability import code mismatch: %#v", errorSummary)
		}
		details := errorSummary["details"].(map[string]any)
		if len(details) != 1 || details["profile_id"] != incidentPortabilityProfileID {
			t.Fatalf("capability import details mismatch: %#v", details)
		}
	} else {
		requireFailedJobReason(t, terminal, "incident_bundle_import_rejected", wantReason)
	}
	if summary, ok := terminal["result_summary"].(map[string]any); ok {
		if refs, ok := summary["resource_refs"].([]any); ok && len(refs) != 0 {
			t.Fatalf("failed import must not expose imported resource refs: %#v", summary)
		}
	}
	var stagingRef string
	if err := harness.DB.QueryRow(`SELECT bundle_staging_ref FROM incident_bundle_job_payloads WHERE job_id = $1`, job["job_id"].(string)).Scan(&stagingRef); err != nil {
		t.Fatalf("query failed import staging reference: %v", err)
	}
	if _, err := os.Stat(stagedBundleTestPath(t, harness.Server, stagingRef)); !os.IsNotExist(err) {
		t.Fatalf("failed import staging reference must be cleaned up, stat err=%v ref=%s", err, stagingRef)
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_job_payloads WHERE job_id = $1 AND (imported_incident_id IS NOT NULL OR manifest_sha256 IS NOT NULL)`, job["job_id"].(string)) != 0 {
		t.Fatalf("failed import must not persist imported incident id or manifest sha")
	}
	var requestJSON string
	if err := harness.DB.QueryRow(`SELECT request_json::text FROM incident_bundle_job_payloads WHERE job_id = $1`, job["job_id"].(string)).Scan(&requestJSON); err != nil {
		t.Fatalf("query failed import request json: %v", err)
	}
	var normalized map[string]any
	if err := json.Unmarshal([]byte(requestJSON), &normalized); err != nil {
		t.Fatalf("decode failed import request json: %v", err)
	}
	fileSHA, ok := normalized["file_sha256"].(string)
	if len(normalized) != 1 || !ok || fileSHA == "" || strings.Contains(requestJSON, "manifest.json") {
		t.Fatalf("failed import request payload must retain only upload hash, got %s", requestJSON)
	}
	assertNoIncidentBundleStaging(t, harness.Server)
	after := snapshotImportFailureState(t, harness, incidentID)
	if !before.equal(after) {
		t.Fatalf("failed import left partial state: before=%#v after=%#v", before, after)
	}
	return terminal
}

type importFailureState struct {
	IncidentRows            int
	RecordRows              int
	MembershipRows          int
	IncidentPreferenceRows  int
	UserPreferenceRows      int
	ProjectionRows          int
	ImportedActorRows       int
	ImportedAttributionRows int
	SuccessAuditRows        int
	ExportRows              int
	ExtensionStateRows      int
	ExtensionStagedRows     int
	ImportedObjectKeys      []string
}

func (s importFailureState) equal(other importFailureState) bool {
	return s.IncidentRows == other.IncidentRows &&
		s.RecordRows == other.RecordRows &&
		s.MembershipRows == other.MembershipRows &&
		s.IncidentPreferenceRows == other.IncidentPreferenceRows &&
		s.UserPreferenceRows == other.UserPreferenceRows &&
		s.ProjectionRows == other.ProjectionRows &&
		s.ImportedActorRows == other.ImportedActorRows &&
		s.ImportedAttributionRows == other.ImportedAttributionRows &&
		s.SuccessAuditRows == other.SuccessAuditRows &&
		s.ExportRows == other.ExportRows &&
		s.ExtensionStateRows == other.ExtensionStateRows &&
		s.ExtensionStagedRows == other.ExtensionStagedRows &&
		slices.Equal(s.ImportedObjectKeys, other.ImportedObjectKeys)
}

func snapshotImportFailureState(t testing.TB, harness *appsupport.ServerHarness, incidentID string) importFailureState {
	t.Helper()
	return importFailureState{
		IncidentRows:            countRows(t, harness.DB, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID),
		RecordRows:              countRows(t, harness.DB, `SELECT count(*) FROM records WHERE incident_id = $1`, incidentID),
		MembershipRows:          countRows(t, harness.DB, `SELECT count(*) FROM incident_memberships WHERE incident_id = $1`, incidentID),
		IncidentPreferenceRows:  countRows(t, harness.DB, `SELECT count(*) FROM incident_workbook_preferences WHERE incident_id = $1`, incidentID),
		UserPreferenceRows:      countRows(t, harness.DB, `SELECT count(*) FROM user_workbook_preferences WHERE incident_id = $1`, incidentID),
		ProjectionRows:          countRows(t, harness.DB, `SELECT count(*) FROM timeline_grid_projection WHERE incident_id = $1`, incidentID),
		ImportedActorRows:       countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_imported_actors WHERE incident_id = $1`, incidentID),
		ImportedAttributionRows: countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_imported_attributions WHERE incident_id = $1`, incidentID),
		SuccessAuditRows:        countRows(t, harness.DB, `SELECT count(*) FROM deployment_admin_audit_events WHERE incident_id = $1`, incidentID),
		ExportRows:              countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_exports WHERE incident_id = $1`, incidentID),
		ExtensionStateRows:      countRows(t, harness.DB, `SELECT count(*) FROM extension_state_metadata`),
		ExtensionStagedRows:     countRows(t, harness.DB, `SELECT count(*) FROM extension_staged_objects`),
		ImportedObjectKeys:      objectKeysWithPrefix(t, harness.ObjectStore, "incidents/"+incidentID+"/object-blobs/"),
	}
}

func objectKeysWithPrefix(t testing.TB, store objectstore.Store, prefix string) []string {
	t.Helper()
	objects, err := store.ListObjects(context.Background(), prefix)
	if err != nil {
		t.Fatalf("list object store prefix %s: %v", prefix, err)
	}
	keys := make([]string, 0, len(objects))
	for _, object := range objects {
		keys = append(keys, object.Key)
	}
	slices.Sort(keys)
	return keys
}

func seedMissingIncidentBundleBlob(t testing.TB, harness *appsupport.ServerHarness, incidentID string, actorUserID string) {
	t.Helper()
	missingBytes := []byte("incident-bundle missing blob fixture")
	sha := hashHexBytes(missingBytes)
	if _, err := harness.DB.Exec(`
INSERT INTO object_blobs (
    incident_id,
    created_by_user_id,
    storage_key,
    upload_state,
    byte_size,
    filename_hint,
    content_type_hint,
    expected_sha256_hex,
    observed_size,
    observed_content_type,
    observed_sha256_hex,
    target_expires_at,
    pending_expires_at,
    finalized_at
)
VALUES ($1, $2, $3, 'available', $4, 'missing.txt', 'text/plain', $5, $4, 'text/plain', $5, now() + interval '1 hour', now() + interval '1 hour', now())
`, incidentID, actorUserID, "extension_profile/missing/"+incidentID+"/"+sha, len(missingBytes), sha); err != nil {
		t.Fatalf("seed missing object blob row: %v", err)
	}
}

type envelopeDurability struct {
	Jobs        int
	Payloads    int
	Idempotency int
}

func snapshotEnvelopeDurability(t testing.TB, db *sql.DB) envelopeDurability {
	t.Helper()
	return envelopeDurability{
		Jobs:        countRows(t, db, `SELECT count(*) FROM jobs`),
		Payloads:    countRows(t, db, `SELECT count(*) FROM incident_bundle_job_payloads`),
		Idempotency: countRows(t, db, `SELECT count(*) FROM route_idempotency WHERE route_key = 'incident_bundles.import'`),
	}
}

func assertNoIncidentBundleStaging(t testing.TB, server *httptestx.Server) {
	t.Helper()
	stagingDir := filepath.Join(server.Config.Roots.TemporaryWork.Path, "incident-bundles", "imports")
	entries, err := os.ReadDir(stagingDir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("read staging dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("incident bundle staging dir must remain empty, found %d entries in %s", len(entries), stagingDir)
	}
}

type uploadPart struct {
	Name        string
	Filename    string
	ContentType string
	Body        []byte
}

func jsonUploadPart(name string, filename string, body string) uploadPart {
	return uploadPart{Name: name, Filename: filename, ContentType: "application/json; charset=utf-8", Body: []byte(body)}
}

func fileUploadPart(name string, filename string, contentType string, body []byte) uploadPart {
	return uploadPart{Name: name, Filename: filename, ContentType: contentType, Body: body}
}

func newImportEnvelopeRequest(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, parts []uploadPart) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range parts {
		disposition := `form-data; name="` + part.Name + `"`
		if part.Filename != "" {
			disposition += `; filename="` + part.Filename + `"`
		}
		w, err := writer.CreatePart(textprotoMIMEHeader(map[string]string{
			"Content-Disposition": disposition,
			"Content-Type":        part.ContentType,
		}))
		if err != nil {
			t.Fatalf("create multipart part %s: %v", part.Name, err)
		}
		if _, err := w.Write(part.Body); err != nil {
			t.Fatalf("write multipart part %s: %v", part.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/import", &body)
	if err != nil {
		t.Fatalf("create import request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	addImportAuth(req, login)
	return req
}

func addImportAuth(req *http.Request, login flowtest.LoginResult) {
	req.AddCookie(login.SessionCookie)
	req.AddCookie(login.CSRFCookie)
	req.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
}

func stringArray(t testing.TB, raw any) []string {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected JSON array, got %T %#v", raw, raw)
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("expected string array item, got %T %#v", item, item)
		}
		result = append(result, value)
	}
	return result
}

func hashHexBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
