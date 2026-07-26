package incidentbundles

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	evidencemodule "github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type BundleBuilder struct {
	pool        *pgxpool.Pool
	objectStore objectstore.Store
	portability *PortabilityOrchestrator
}

type Importer struct {
	pool              *pgxpool.Pool
	objectStore       objectstore.Store
	finalizer         incidents.IncidentBundleImportFinalizer
	projectionRebuild importProjectionRebuilder
}

type BuiltIncidentBundle struct {
	Archive        BundleArchive
	IncidentKey    string
	BundleSHA256   string
	BundleByteSize int64
}

type ImportParams struct {
	ActorUserID uuid.UUID
	PublishedAt time.Time
	RequestID   *string
}

type PreparedImport struct {
	IncidentID       uuid.UUID
	files            map[string][]byte
	attributions     importedAttributionBuffer
	stagedObjectKeys []string
	blobPort         evidencemodule.IncidentBundleBlobPortability
}

func (p *PreparedImport) Cleanup(ctx context.Context) {
	if p == nil || len(p.stagedObjectKeys) == 0 {
		return
	}
	p.blobPort.CleanupStagedObjects(context.WithoutCancel(ctx), p.stagedObjectKeys)
	p.stagedObjectKeys = nil
}

type importProjectionRebuilder interface {
	RebuildImportedIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error
}

func (b BundleBuilder) Build(ctx context.Context, incidentID uuid.UUID, request ExportRequest, bundleID uuid.UUID, exportedAt time.Time) (BuiltIncidentBundle, error) {
	tx, err := b.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return BuiltIncidentBundle{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	files := map[string][]byte{}
	incidentJSON, incidentKey, err := incidents.ExportIncidentBundleIncident(ctx, tx, incidentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BuiltIncidentBundle{}, ErrNotFound
		}
		return BuiltIncidentBundle{}, err
	}
	files["data/incident.json"] = incidentJSON
	for _, export := range incidentBundleExportPorts {
		exportedFiles, err := export(ctx, tx, incidentID)
		if err != nil {
			return BuiltIncidentBundle{}, err
		}
		for _, file := range exportedFiles {
			files[file.Path] = file.Payload
		}
	}
	files["data/reference_pack_refs.json"] = []byte("[]\n")
	actors, err := b.exportActors(ctx, tx, files)
	if err != nil {
		return BuiltIncidentBundle{}, err
	}
	files["data/actors.ndjson"] = actors
	blobPort := evidencemodule.IncidentBundleBlobPortability{ObjectStore: b.objectStore}
	if err := blobPort.ExportBlobFiles(ctx, tx, incidentID, files); err != nil {
		return BuiltIncidentBundle{}, verificationErrorFromPort(err)
	}
	if b.portability != nil {
		payloads, err := b.portability.Export(ctx, incidentID)
		if err != nil {
			return BuiltIncidentBundle{}, err
		}
		for _, payload := range payloads {
			filePath, encoded, err := EncodeExtensionPayload(payload)
			if err != nil {
				return BuiltIncidentBundle{}, err
			}
			files[filePath] = encoded
		}
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
	if err := tx.Commit(ctx); err != nil {
		return BuiltIncidentBundle{}, err
	}
	return BuiltIncidentBundle{
		Archive:        archive,
		IncidentKey:    incidentKey,
		BundleSHA256:   hashHex(archive.Bytes),
		BundleByteSize: int64(len(archive.Bytes)),
	}, nil
}

type incidentBundleExportPort func(context.Context, incidentportability.Queryer, uuid.UUID) ([]incidentportability.File, error)

var incidentBundleExportPorts = []incidentBundleExportPort{
	records.ExportIncidentBundleFiles,
	timeline.ExportIncidentBundleFiles,
	parties.ExportIncidentBundleFiles,
	entities.ExportIncidentBundleFiles,
	indicators.ExportIncidentBundleFiles,
	artifacts.ExportIncidentBundleFiles,
	tasksdecisions.ExportIncidentBundleFiles,
	evidencemodule.ExportIncidentBundleFiles,
	assessments.ExportIncidentBundleFiles,
	links.ExportIncidentBundleFiles,
	revisions.ExportIncidentBundleFiles,
	savedviews.ExportIncidentBundleFiles,
}

func (b BundleBuilder) exportActors(ctx context.Context, q incidentportability.Queryer, files map[string][]byte) ([]byte, error) {
	actorIDs := map[string]struct{}{}
	for path, payload := range files {
		if path == "data/actors.ndjson" || path == "data/reference_pack_refs.json" || !strings.HasPrefix(path, "data/") {
			continue
		}
		if path == "data/incident.json" {
			var row map[string]any
			if err := json.Unmarshal(bytes.TrimSpace(payload), &row); err != nil {
				return nil, err
			}
			collectActorIDs(row, actorIDs)
			continue
		}
		rows, err := incidentportability.DecodeNDJSON(payload)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			collectActorIDs(row, actorIDs)
		}
	}
	if len(actorIDs) == 0 {
		return []byte{}, nil
	}
	ids := make([]string, 0, len(actorIDs))
	for id := range actorIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	rows, err := q.Query(ctx, `
SELECT jsonb_build_object(
    'actor_id', u.id::text,
    'display_name', u.display_name,
    'email_hint', u.email::text
)
  FROM users u
 WHERE u.id IN (SELECT unnest($1::text[])::uuid)
 ORDER BY u.id
`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return incidentportability.EncodeRows(rows)
}

func collectActorIDs(row map[string]any, actorIDs map[string]struct{}) {
	for key, value := range row {
		if !strings.HasSuffix(key, "_user_id") || value == nil {
			continue
		}
		id := strings.TrimSpace(incidentportability.StringFromAny(value))
		if _, err := uuid.Parse(id); err == nil {
			actorIDs[id] = struct{}{}
		}
	}
}

func (i Importer) PrepareImport(ctx context.Context, verified VerifiedBundle, params ImportParams) (*PreparedImport, error) {
	incidentID, err := uuid.Parse(verified.Manifest.IncidentID)
	if err != nil {
		return nil, &VerificationError{ReasonCode: "malformed_manifest"}
	}
	var existing bool
	if err := i.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM incidents WHERE id = $1)`, incidentID).Scan(&existing); err != nil {
		return nil, err
	}
	if existing {
		return nil, &VerificationError{ReasonCode: "duplicate_incident_id"}
	}
	attributions := importedAttributionBuffer{IncidentID: incidentID, LocalUserID: params.ActorUserID}
	blobPort := evidencemodule.IncidentBundleBlobPortability{ObjectStore: i.objectStore}
	rewrittenObjectBlobs, writtenObjectKeys, err := blobPort.RewriteAndStageObjectBlobs(ctx, verified.Files, incidentID, params.ActorUserID, &attributions)
	if err != nil {
		blobPort.CleanupStagedObjects(context.WithoutCancel(ctx), writtenObjectKeys)
		return nil, verificationErrorFromPort(err)
	}
	importFiles := make(map[string][]byte, len(verified.Files))
	for path, payload := range verified.Files {
		importFiles[path] = append([]byte(nil), payload...)
	}
	importFiles["data/object_blobs.ndjson"] = rewrittenObjectBlobs
	return &PreparedImport{
		IncidentID: incidentID, files: importFiles, attributions: attributions,
		stagedObjectKeys: writtenObjectKeys, blobPort: blobPort,
	}, nil
}

func (i Importer) ApplyPreparedImportTx(ctx context.Context, tx pgx.Tx, prepared *PreparedImport, params ImportParams) (uuid.UUID, error) {
	if tx == nil || prepared == nil || prepared.IncidentID == uuid.Nil {
		return uuid.UUID{}, errors.New("prepared incident bundle import is required")
	}
	if err := collaboration.SuppressHistoricalIntentsTx(ctx, tx); err != nil {
		return uuid.UUID{}, err
	}
	incidentID := prepared.IncidentID
	var existing int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID).Scan(&existing); err != nil {
		return uuid.UUID{}, err
	}
	if existing > 0 {
		return uuid.UUID{}, &VerificationError{ReasonCode: "duplicate_incident_id"}
	}
	attributions := prepared.attributions
	if err := incidents.ImportIncidentBundleIncidentTx(ctx, tx, prepared.files["data/incident.json"], params.ActorUserID, &attributions); err != nil {
		return uuid.UUID{}, verificationErrorFromPort(err)
	}
	if err := i.importActors(ctx, tx, prepared.files["data/actors.ndjson"], incidentID); err != nil {
		return uuid.UUID{}, verificationErrorFromPort(err)
	}
	for _, importPort := range incidentBundleImportPorts {
		if err := importPort(ctx, tx, prepared.files, params.ActorUserID, &attributions); err != nil {
			return uuid.UUID{}, verificationErrorFromPort(err)
		}
	}
	if err := revisions.RepairIncidentBundleImportedSequencesTx(ctx, tx); err != nil {
		return uuid.UUID{}, err
	}
	if err := attributions.flush(ctx, tx); err != nil {
		return uuid.UUID{}, err
	}
	projectionRebuild := i.projectionRebuild
	if projectionRebuild == nil {
		return uuid.UUID{}, errors.New("incident bundle projection rebuild is required")
	}
	if err := projectionRebuild.RebuildImportedIncidentTx(ctx, tx, incidentID); err != nil {
		return uuid.UUID{}, err
	}
	if i.finalizer == nil {
		return uuid.UUID{}, errors.New("incident bundle import finalizer is required")
	}
	if err := i.finalizer.FinalizeIncidentBundleImportTx(ctx, tx, incidents.IncidentBundleImportFinalizationParams{
		IncidentID:        incidentID,
		SubmittedByUserID: params.ActorUserID,
		PublishedAt:       params.PublishedAt,
		RequestID:         params.RequestID,
	}); err != nil {
		if errors.Is(err, incidents.ErrInitialAdminUnavailable) {
			return uuid.UUID{}, &VerificationError{ReasonCode: "initial_admin_unavailable"}
		}
		return uuid.UUID{}, err
	}
	prepared.attributions = attributions
	return incidentID, nil
}

type incidentBundleImportPort func(context.Context, pgx.Tx, map[string][]byte, uuid.UUID, incidentportability.AttributionRecorder) error

var incidentBundleImportPorts = []incidentBundleImportPort{
	records.ImportIncidentBundleFilesTx,
	timeline.ImportIncidentBundleFilesTx,
	parties.ImportIncidentBundleFilesTx,
	entities.ImportIncidentBundleFilesTx,
	indicators.ImportIncidentBundleFilesTx,
	artifacts.ImportIncidentBundleFilesTx,
	tasksdecisions.ImportIncidentBundleFilesTx,
	evidencemodule.ImportIncidentBundleFilesTx,
	assessments.ImportIncidentBundleFilesTx,
	links.ImportIncidentBundleFilesTx,
	revisions.ImportIncidentBundleFilesTx,
	savedviews.ImportIncidentBundleFilesTx,
}

func verificationErrorFromPort(err error) error {
	var malformed *incidentportability.MalformedPayloadError
	if errors.As(err, &malformed) {
		return &VerificationError{ReasonCode: "malformed_manifest"}
	}
	var verification *incidentportability.VerificationFailure
	if errors.As(err, &verification) {
		return &VerificationError{ReasonCode: verification.ReasonCode}
	}
	return err
}

func (i Importer) importActors(ctx context.Context, tx pgx.Tx, payload []byte, incidentID uuid.UUID) error {
	rows, err := incidentportability.DecodeNDJSON(payload)
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

func (b *importedAttributionBuffer) RecordImportedAttribution(table string, row map[string]any, column string, sourceActorID string) {
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

func sourceRowID(table string, row map[string]any) string {
	if table == "incidents" {
		return incidentportability.StringFromAny(row["id"])
	}
	return incidentportability.SourceRowIDForRelation(table, row)
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
