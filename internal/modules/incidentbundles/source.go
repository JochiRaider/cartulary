package incidentbundles

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type BundleBuilder struct {
	pool        *pgxpool.Pool
	objectStore objectstore.Store
}

type Importer struct {
	pool        *pgxpool.Pool
	objectStore objectstore.Store
}

type BuiltIncidentBundle struct {
	Archive        BundleArchive
	IncidentKey    string
	BundleSHA256   string
	BundleByteSize int64
}

type exportNDJSONSpec struct {
	Path  string
	Query string
}

var exportSpecs = []exportNDJSONSpec{
	{"data/records.ndjson", `SELECT to_jsonb(t) FROM records t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/timeline_events.ndjson", `SELECT to_jsonb(t) FROM timeline_events t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/entity_mentions.ndjson", `SELECT to_jsonb(t) FROM entity_mentions t JOIN records r ON r.record_id = t.source_record_id WHERE r.incident_id = $1 ORDER BY t.entity_mention_id`},
	{"data/hosts.ndjson", `SELECT to_jsonb(t) FROM hosts t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/identities.ndjson", `SELECT to_jsonb(t) FROM identities t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/entity_aliases.ndjson", `SELECT to_jsonb(t) FROM entity_aliases t WHERE incident_id = $1 ORDER BY entity_alias_id`},
	{"data/indicators.ndjson", `SELECT to_jsonb(t) FROM indicators t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/indicator_observations.ndjson", `SELECT to_jsonb(t) FROM indicator_observations t WHERE incident_id = $1 ORDER BY indicator_observation_id`},
	{"data/indicator_state_intervals.ndjson", `SELECT to_jsonb(t) FROM indicator_state_intervals t WHERE incident_id = $1 ORDER BY indicator_state_interval_id`},
	{"data/artifacts.ndjson", `SELECT to_jsonb(t) FROM artifacts t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/task_requests.ndjson", `SELECT to_jsonb(t) FROM task_requests t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/decisions.ndjson", `SELECT to_jsonb(t) FROM decisions t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/evidence_records.ndjson", `SELECT to_jsonb(t) FROM evidence t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/evidence_custody_events.ndjson", `SELECT to_jsonb(t) FROM evidence_custody_events t WHERE incident_id = $1 ORDER BY evidence_record_id, occurred_at, custody_event_id`},
	{"data/object_blobs.ndjson", `SELECT to_jsonb(t) FROM object_blobs t WHERE incident_id = $1 ORDER BY object_blob_id`},
	{"data/compromise_assessments.ndjson", `SELECT to_jsonb(t) FROM assessments t WHERE incident_id = $1 ORDER BY record_id`},
	{"data/record_links.ndjson", `SELECT to_jsonb(t) FROM record_links t WHERE incident_id = $1 ORDER BY record_link_id`},
	{"data/tags.ndjson", `SELECT jsonb_build_object('tag_name', tag_name, 'normalized_tag_name', normalized_tag_name) FROM (SELECT DISTINCT tag_name, normalized_tag_name FROM record_tags WHERE incident_id = $1 ORDER BY normalized_tag_name, tag_name) tags`},
	{"data/record_tags.ndjson", `SELECT to_jsonb(t) FROM record_tags t WHERE incident_id = $1 ORDER BY record_id, normalized_tag_name, record_tag_id`},
	{"data/change_sets.ndjson", `SELECT to_jsonb(t) FROM change_sets t WHERE incident_id = $1 ORDER BY created_at, change_set_id`},
	{"data/change_set_mutations.ndjson", `SELECT to_jsonb(t) FROM change_set_mutations t JOIN change_sets c ON c.change_set_id = t.change_set_id WHERE c.incident_id = $1 ORDER BY t.change_set_id, t.sequence_no`},
	{"data/record_revisions.ndjson", `SELECT to_jsonb(t) FROM record_revisions t JOIN change_sets c ON c.change_set_id = t.change_set_id WHERE c.incident_id = $1 ORDER BY t.record_id, t.row_version`},
	{"data/saved_views.ndjson", `SELECT to_jsonb(t) FROM saved_views t WHERE incident_id = $1 ORDER BY saved_view_id`},
}

var importSpecs = []struct {
	Path  string
	Table string
}{
	{"data/records.ndjson", "records"},
	{"data/timeline_events.ndjson", "timeline_events"},
	{"data/entity_mentions.ndjson", "entity_mentions"},
	{"data/hosts.ndjson", "hosts"},
	{"data/identities.ndjson", "identities"},
	{"data/entity_aliases.ndjson", "entity_aliases"},
	{"data/indicators.ndjson", "indicators"},
	{"data/indicator_observations.ndjson", "indicator_observations"},
	{"data/indicator_state_intervals.ndjson", "indicator_state_intervals"},
	{"data/artifacts.ndjson", "artifacts"},
	{"data/task_requests.ndjson", "task_requests"},
	{"data/decisions.ndjson", "decisions"},
	{"data/object_blobs.ndjson", "object_blobs"},
	{"data/evidence_records.ndjson", "evidence"},
	{"data/evidence_custody_events.ndjson", "evidence_custody_events"},
	{"data/compromise_assessments.ndjson", "assessments"},
	{"data/record_links.ndjson", "record_links"},
	{"data/record_tags.ndjson", "record_tags"},
	{"data/change_sets.ndjson", "change_sets"},
	{"data/change_set_mutations.ndjson", "change_set_mutations"},
	{"data/record_revisions.ndjson", "record_revisions"},
	{"data/saved_views.ndjson", "saved_views"},
}

func (b BundleBuilder) Build(ctx context.Context, incidentID uuid.UUID, request ExportRequest, bundleID uuid.UUID, exportedAt time.Time) (BuiltIncidentBundle, error) {
	files := map[string][]byte{}
	incidentJSON, incidentKey, err := b.exportIncident(ctx, incidentID)
	if err != nil {
		return BuiltIncidentBundle{}, err
	}
	files["data/incident.json"] = incidentJSON
	actors, err := b.exportActors(ctx, incidentID)
	if err != nil {
		return BuiltIncidentBundle{}, err
	}
	files["data/actors.ndjson"] = actors
	for _, spec := range exportSpecs {
		payload, err := b.exportNDJSON(ctx, incidentID, spec.Query)
		if err != nil {
			return BuiltIncidentBundle{}, err
		}
		files[spec.Path] = payload
	}
	files["data/reference_pack_refs.json"] = []byte("[]\n")
	if err := b.exportBlobs(ctx, incidentID, files); err != nil {
		return BuiltIncidentBundle{}, err
	}
	archive, err := BuildBundleArchive(ManifestInput{
		BundleID:             bundleID.String(),
		IncidentID:           incidentID.String(),
		IncidentKey:          incidentKey,
		ExportedAt:           exportedAt.UTC().Format(time.RFC3339Nano),
		ReferencePackMode:    request.ReferencePackMode,
		OptionalSections:     request.OptionalSections,
		RequiredCapabilities: request.RequiredCapabilities,
	}, files)
	if err != nil {
		return BuiltIncidentBundle{}, err
	}
	return BuiltIncidentBundle{
		Archive:        archive,
		IncidentKey:    incidentKey,
		BundleSHA256:   hashHex(archive.Bytes),
		BundleByteSize: int64(len(archive.Bytes)),
	}, nil
}

func (b BundleBuilder) exportIncident(ctx context.Context, incidentID uuid.UUID) ([]byte, string, error) {
	var raw []byte
	var incidentKey string
	err := b.pool.QueryRow(ctx, `SELECT to_jsonb(i), incident_key FROM incidents i WHERE id = $1`, incidentID).Scan(&raw, &incidentKey)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	canonical, err := canonicalRawJSON(raw)
	return canonical, incidentKey, err
}

func (b BundleBuilder) exportActors(ctx context.Context, incidentID uuid.UUID) ([]byte, error) {
	rows, err := b.pool.Query(ctx, `
WITH scoped_rows AS (
    SELECT to_jsonb(t) AS row_json FROM incidents t WHERE id = $1
    UNION ALL SELECT to_jsonb(t) FROM records t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM timeline_events t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM entity_mentions t JOIN records r ON r.record_id = t.source_record_id WHERE r.incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM hosts t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM identities t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM entity_aliases t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM indicators t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM indicator_observations t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM indicator_state_intervals t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM artifacts t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM task_requests t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM decisions t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM evidence t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM evidence_custody_events t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM object_blobs t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM assessments t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM record_links t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM record_tags t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM change_sets t WHERE incident_id = $1
    UNION ALL SELECT to_jsonb(t) FROM saved_views t WHERE incident_id = $1
), actor_ids AS (
    SELECT DISTINCT kv.value::uuid AS user_id
      FROM scoped_rows
      CROSS JOIN LATERAL jsonb_each_text(row_json) AS kv(key, value)
     WHERE kv.key LIKE '%\_user_id' ESCAPE '\'
       AND kv.value IS NOT NULL
       AND kv.value <> ''
       AND kv.value <> 'null'
       AND kv.value ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
)
SELECT jsonb_build_object(
    'actor_id', u.id::text,
    'display_name', u.display_name,
    'email_hint', u.email::text
)
  FROM users u
  JOIN actor_ids a ON a.user_id = u.id
 ORDER BY u.id
`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return encodeRows(rows)
}

func (b BundleBuilder) exportNDJSON(ctx context.Context, incidentID uuid.UUID, query string) ([]byte, error) {
	rows, err := b.pool.Query(ctx, query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return encodeRows(rows)
}

func encodeRows(rows pgx.Rows) ([]byte, error) {
	var buf bytes.Buffer
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		canonical, err := canonicalRawJSON(raw)
		if err != nil {
			return nil, err
		}
		buf.Write(canonical)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b BundleBuilder) exportBlobs(ctx context.Context, incidentID uuid.UUID, files map[string][]byte) error {
	rows, err := b.pool.Query(ctx, `
SELECT storage_key, observed_sha256_hex
  FROM object_blobs
 WHERE incident_id = $1
   AND upload_state = 'available'
   AND observed_sha256_hex IS NOT NULL
 ORDER BY observed_sha256_hex
`, incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var storageKey string
		var sha string
		if err := rows.Scan(&storageKey, &sha); err != nil {
			return err
		}
		rc, _, err := b.objectStore.ReadObject(ctx, storageKey, objectstore.ReadOptions{})
		if err != nil {
			return &VerificationError{ReasonCode: "missing_required_blob"}
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return &VerificationError{ReasonCode: "missing_required_blob"}
		}
		if hashHex(data) != sha {
			return &VerificationError{ReasonCode: "missing_required_blob"}
		}
		files["blobs/sha256/"+sha] = data
	}
	return rows.Err()
}

func (i Importer) Import(ctx context.Context, verified VerifiedBundle, actorUserID uuid.UUID) (uuid.UUID, error) {
	incidentID, err := uuid.Parse(verified.Manifest.IncidentID)
	if err != nil {
		return uuid.UUID{}, &VerificationError{ReasonCode: "malformed_manifest"}
	}
	tx, err := i.pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var existing int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID).Scan(&existing); err != nil {
		return uuid.UUID{}, err
	}
	if existing > 0 {
		return uuid.UUID{}, &VerificationError{ReasonCode: "duplicate_incident_id"}
	}
	attributions := importedAttributionBuffer{IncidentID: incidentID, LocalUserID: actorUserID}
	if err := i.importIncident(ctx, tx, verified.Files["data/incident.json"], actorUserID, &attributions); err != nil {
		return uuid.UUID{}, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO incident_memberships (incident_id, user_id, role, joined_at, added_by_user_id, updated_at, updated_by_user_id)
VALUES ($1, $2, 'admin', now(), $2, now(), $2)
ON CONFLICT (incident_id, user_id) DO NOTHING
`, incidentID, actorUserID); err != nil {
		return uuid.UUID{}, err
	}
	if err := i.importActors(ctx, tx, verified.Files["data/actors.ndjson"], incidentID); err != nil {
		return uuid.UUID{}, err
	}
	rewrittenObjectBlobs, writtenObjectKeys, err := i.rewriteAndImportObjectBlobBytes(ctx, verified, incidentID, actorUserID, &attributions)
	if err != nil {
		return uuid.UUID{}, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		for _, key := range writtenObjectKeys {
			_ = i.objectStore.DeleteObject(ctx, key)
		}
	}()
	for _, spec := range importSpecs {
		payload := verified.Files[spec.Path]
		if spec.Path == "data/object_blobs.ndjson" {
			payload = rewrittenObjectBlobs
		}
		if len(bytes.TrimSpace(payload)) == 0 {
			continue
		}
		if err := i.importNDJSON(ctx, tx, spec.Table, payload, actorUserID, &attributions); err != nil {
			return uuid.UUID{}, err
		}
	}
	if err := attributions.flush(ctx, tx); err != nil {
		return uuid.UUID{}, err
	}
	projectionStore := projections.NewStore(i.pool)
	if err := projectionStore.RebuildIncidentTimelineTx(ctx, tx, incidentID); err != nil {
		return uuid.UUID{}, err
	}
	if err := projectionStore.RebuildIncidentHostsTx(ctx, tx, incidentID); err != nil {
		return uuid.UUID{}, err
	}
	if err := projectionStore.RebuildIncidentIdentitiesTx(ctx, tx, incidentID); err != nil {
		return uuid.UUID{}, err
	}
	if err := projectionStore.RebuildIncidentIndicatorsTx(ctx, tx, incidentID); err != nil {
		return uuid.UUID{}, err
	}
	if err := projectionStore.RebuildIncidentAssessmentsTx(ctx, tx, incidentID); err != nil {
		return uuid.UUID{}, err
	}
	if err := projectionStore.RebuildIncidentTaskRequestsTx(ctx, tx, incidentID); err != nil {
		return uuid.UUID{}, err
	}
	if err := projectionStore.RebuildIncidentDecisionsTx(ctx, tx, incidentID); err != nil {
		return uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, err
	}
	committed = true
	return incidentID, nil
}

func (i Importer) importIncident(ctx context.Context, tx pgx.Tx, payload []byte, actorUserID uuid.UUID, attributions *importedAttributionBuffer) error {
	var row map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(payload), &row); err != nil {
		return &VerificationError{ReasonCode: "malformed_manifest"}
	}
	remapTopLevelUserFields(row, "incidents", actorUserID, attributions)
	raw, err := json.Marshal(row)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO incidents SELECT * FROM jsonb_populate_record(NULL::incidents, $1::jsonb)`, raw)
	return err
}

func (i Importer) importActors(ctx context.Context, tx pgx.Tx, payload []byte, incidentID uuid.UUID) error {
	rows, err := decodeNDJSON(payload)
	if err != nil {
		return err
	}
	for _, row := range rows {
		sourceActorID, _ := row["actor_id"].(string)
		if strings.TrimSpace(sourceActorID) == "" {
			sourceActorID, _ = row["source_actor_id"].(string)
		}
		if strings.TrimSpace(sourceActorID) == "" {
			continue
		}
		displayName, _ := row["display_name"].(string)
		emailHint, _ := row["email_hint"].(string)
		_, err := tx.Exec(ctx, `
INSERT INTO incident_bundle_imported_actors (incident_id, source_actor_id, display_name, email_hint, local_user_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (incident_id, source_actor_id) DO NOTHING
`, incidentID, sourceActorID, nullableString(displayName), nullableString(emailHint), nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i Importer) rewriteAndImportObjectBlobBytes(ctx context.Context, verified VerifiedBundle, incidentID uuid.UUID, actorUserID uuid.UUID, attributions *importedAttributionBuffer) ([]byte, []string, error) {
	rows, err := decodeNDJSON(verified.Files["data/object_blobs.ndjson"])
	if err != nil {
		return nil, nil, err
	}
	writtenKeys := make([]string, 0, len(rows))
	var buf bytes.Buffer
	for _, row := range rows {
		remapTopLevelUserFields(row, "object_blobs", actorUserID, attributions)
		state, _ := row["upload_state"].(string)
		if state != "available" {
			line, err := canonicalJSONString(row)
			if err != nil {
				return nil, writtenKeys, err
			}
			buf.Write(line)
			continue
		}
		sha, _ := row["observed_sha256_hex"].(string)
		if sha == "" {
			sha, _ = row["expected_sha256_hex"].(string)
		}
		data, ok := verified.Files["blobs/sha256/"+sha]
		if !ok {
			return nil, writtenKeys, &VerificationError{ReasonCode: "missing_required_blob"}
		}
		if hashHex(data) != sha {
			return nil, writtenKeys, &VerificationError{ReasonCode: "blob_hash_mismatch"}
		}
		storageKey := "incident-bundles/imported/" + incidentID.String() + "/sha256/" + sha
		row["storage_key"] = storageKey
		contentType, _ := row["observed_content_type"].(string)
		if contentType == "" {
			contentType, _ = row["content_type_hint"].(string)
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		if err := i.objectStore.PutObject(ctx, storageKey, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
			return nil, writtenKeys, err
		}
		writtenKeys = append(writtenKeys, storageKey)
		line, err := canonicalJSONString(row)
		if err != nil {
			return nil, writtenKeys, err
		}
		buf.Write(line)
	}
	return buf.Bytes(), writtenKeys, nil
}

func (i Importer) importNDJSON(ctx context.Context, tx pgx.Tx, table string, payload []byte, actorUserID uuid.UUID, attributions *importedAttributionBuffer) error {
	rows, err := decodeNDJSON(payload)
	if err != nil {
		return err
	}
	for _, row := range rows {
		remapTopLevelUserFields(row, table, actorUserID, attributions)
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		query := fmt.Sprintf("INSERT INTO %s SELECT * FROM jsonb_populate_record(NULL::%s, $1::jsonb) ON CONFLICT DO NOTHING", table, table)
		if _, err := tx.Exec(ctx, query, raw); err != nil {
			return err
		}
	}
	return nil
}

func decodeNDJSON(payload []byte) ([]map[string]any, error) {
	scanner := bufio.NewScanner(bytes.NewReader(payload))
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	var rows []map[string]any
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var row map[string]any
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.UseNumber()
		if err := decoder.Decode(&row); err != nil {
			return nil, &VerificationError{ReasonCode: "malformed_manifest"}
		}
		rows = append(rows, row)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

type importedAttribution struct {
	SourceTable   string
	SourceRowID   string
	SourceColumn  string
	SourceActorID string
}

type importedAttributionBuffer struct {
	IncidentID  uuid.UUID
	LocalUserID uuid.UUID
	rows        []importedAttribution
}

func (b *importedAttributionBuffer) add(table string, row map[string]any, column string, sourceActorID string) {
	sourceActorID = strings.TrimSpace(sourceActorID)
	if b == nil || sourceActorID == "" {
		return
	}
	rowID := sourceRowID(table, row)
	if rowID == "" {
		return
	}
	b.rows = append(b.rows, importedAttribution{
		SourceTable:   table,
		SourceRowID:   rowID,
		SourceColumn:  column,
		SourceActorID: sourceActorID,
	})
}

func (b *importedAttributionBuffer) flush(ctx context.Context, tx pgx.Tx) error {
	if b == nil {
		return nil
	}
	for _, row := range b.rows {
		_, err := tx.Exec(ctx, `
INSERT INTO incident_bundle_imported_attributions (
    incident_id,
    source_table,
    source_row_id,
    source_column,
    source_actor_id,
    local_user_id
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (incident_id, source_table, source_row_id, source_column) DO NOTHING
`, b.IncidentID, row.SourceTable, row.SourceRowID, row.SourceColumn, row.SourceActorID, b.LocalUserID)
		if err != nil {
			return err
		}
	}
	return nil
}

func remapTopLevelUserFields(row map[string]any, table string, actorUserID uuid.UUID, attributions *importedAttributionBuffer) {
	for key, value := range row {
		if !strings.HasSuffix(key, "_user_id") || value == nil {
			continue
		}
		sourceActorID := stringFromAny(value)
		if strings.TrimSpace(sourceActorID) == "" {
			continue
		}
		if attributions != nil {
			attributions.add(table, row, key, sourceActorID)
		}
		row[key] = actorUserID.String()
	}
}

func sourceRowID(table string, row map[string]any) string {
	switch table {
	case "incidents":
		return stringFromAny(row["id"])
	case "records", "timeline_events", "hosts", "identities", "indicators", "artifacts", "task_requests", "decisions", "evidence", "assessments", "saved_views":
		return stringFromAny(row["record_id"])
	case "entity_mentions":
		return stringFromAny(row["entity_mention_id"])
	case "entity_aliases":
		return stringFromAny(row["entity_alias_id"])
	case "indicator_observations":
		return stringFromAny(row["indicator_observation_id"])
	case "indicator_state_intervals":
		return stringFromAny(row["indicator_state_interval_id"])
	case "object_blobs":
		return stringFromAny(row["object_blob_id"])
	case "record_links":
		return stringFromAny(row["record_link_id"])
	case "record_tags":
		return stringFromAny(row["record_tag_id"])
	case "change_sets":
		return stringFromAny(row["change_set_id"])
	case "change_set_mutations":
		changeSetID := stringFromAny(row["change_set_id"])
		sequenceNo := stringFromAny(row["sequence_no"])
		if changeSetID == "" || sequenceNo == "" {
			return ""
		}
		return changeSetID + ":" + sequenceNo
	case "record_revisions":
		return stringFromAny(row["revision_id"])
	case "evidence_custody_events":
		return stringFromAny(row["custody_event_id"])
	default:
		return ""
	}
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func canonicalRawJSON(raw []byte) ([]byte, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return canonicalJSONString(value)
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
