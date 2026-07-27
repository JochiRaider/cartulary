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

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidencemodule "github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type BundleBuilder struct {
	pool          *pgxpool.Pool
	objectStore   objectstore.Store
	portability   *PortabilityOrchestrator
	sourceCatalog *sourceport.Catalog
}

type Importer struct {
	pool              *pgxpool.Pool
	objectStore       objectstore.Store
	finalizer         incidents.IncidentBundleImportFinalizer
	projectionRebuild importProjectionRebuilder
	sourceCatalog     *sourceport.Catalog
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
	OperationID string
}

type PreparedImport struct {
	IncidentID         uuid.UUID
	files              map[string][]byte
	attributions       importedAttributionBuffer
	stagedObjectKeys   []string
	blobPort           evidencemodule.IncidentBundleBlobPortability
	sourcePreparations []preparedSource
	importContext      sourceport.ImportContext
}

type preparedSource struct {
	port     sourceport.Port
	prepared sourceport.Prepared
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
	if b.sourceCatalog == nil {
		return BuiltIncidentBundle{}, errors.New("incident bundle source catalog is required")
	}
	for _, port := range b.sourceCatalog.Ports() {
		exportedFiles, err := port.Export(ctx, tx, incidentID)
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
	if params.OperationID == "" {
		return nil, errors.New("incident bundle import operation ID is required")
	}
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
	if i.sourceCatalog == nil {
		return nil, errors.New("incident bundle source catalog is required")
	}
	attributions := importedAttributionBuffer{IncidentID: incidentID, LocalUserID: params.ActorUserID}
	importContext := sourceport.ImportContext{
		IncidentID: incidentID, ActorUserID: params.ActorUserID,
		BundleVersion: verified.Manifest.BundleVersion, OperationID: params.OperationID,
		Attributions: &attributions,
	}
	if err := validateImportedActors(verified.Files, incidentID); err != nil {
		return nil, err
	}
	if err := reference_data.ValidateIncidentBundleReferences(verified.Files["data/reference_pack_refs.json"]); err != nil {
		invariantID, ok := reference_data.IncidentBundleReferenceInvariant(err)
		if !ok {
			return nil, err
		}
		return nil, &VerificationError{
			ReasonCode:     "source_family_invalid",
			SourceFamilyID: "reference_pack_refs",
			InvariantID:    invariantID,
		}
	}
	sourcePreparations := make([]preparedSource, 0, len(i.sourceCatalog.Ports()))
	bundle := sourceport.MapBundle(verified.Files)
	for _, port := range i.sourceCatalog.Ports() {
		prepared, err := port.PrepareImport(ctx, bundle, importContext)
		if err != nil {
			return nil, verificationErrorFromPort(err)
		}
		sourcePreparations = append(sourcePreparations, preparedSource{port: port, prepared: prepared})
	}
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
		sourcePreparations: sourcePreparations, importContext: importContext,
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
	importContext := prepared.importContext
	importContext.Attributions = &attributions
	importContext.RewrittenObjectBlobs = prepared.files["data/object_blobs.ndjson"]
	if len(prepared.sourcePreparations) == 0 ||
		prepared.sourcePreparations[0].port.Descriptor().FamilyID != "incident" {
		return uuid.UUID{}, errors.New("incident source port must be first")
	}
	if err := prepared.sourcePreparations[0].port.ApplyImportTx(
		ctx, tx, prepared.sourcePreparations[0].prepared, importContext,
	); err != nil {
		return uuid.UUID{}, verificationErrorFromPort(err)
	}
	if err := i.importActors(ctx, tx, prepared.files["data/actors.ndjson"], incidentID); err != nil {
		return uuid.UUID{}, verificationErrorFromPort(err)
	}
	for _, source := range prepared.sourcePreparations[1:] {
		if err := source.port.ApplyImportTx(ctx, tx, source.prepared, importContext); err != nil {
			return uuid.UUID{}, verificationErrorFromPort(err)
		}
	}
	for _, source := range prepared.sourcePreparations {
		if err := source.port.ValidateImportTx(ctx, tx, importContext); err != nil {
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

func verificationErrorFromPort(err error) error {
	var malformed *incidentportability.MalformedPayloadError
	if errors.As(err, &malformed) {
		return &VerificationError{ReasonCode: "malformed_manifest"}
	}
	var verification *incidentportability.VerificationFailure
	if errors.As(err, &verification) {
		return &VerificationError{ReasonCode: verification.ReasonCode}
	}
	var sourceFailure *sourceport.Failure
	if errors.As(err, &sourceFailure) {
		return &VerificationError{
			ReasonCode:     "source_family_invalid",
			SourceFamilyID: sourceFailure.FamilyID,
			InvariantID:    sourceFailure.InvariantID,
		}
	}
	return err
}

func validateImportedActors(files map[string][]byte, incidentID uuid.UUID) error {
	rows, err := incidentportability.DecodeNDJSON(files["data/actors.ndjson"])
	if err != nil {
		return &VerificationError{
			ReasonCode:     "source_family_invalid",
			SourceFamilyID: "actors", InvariantID: "actors.reference_complete",
		}
	}
	descriptors := map[string]struct{}{}
	for _, row := range rows {
		for key := range row {
			switch key {
			case "actor_id", "source_actor_id", "display_name", "email_hint", "provider_subject_hint":
			default:
				return &VerificationError{
					ReasonCode:     "source_family_invalid",
					SourceFamilyID: "actors", InvariantID: "actors.inert",
				}
			}
		}
		sourceActorID := strings.TrimSpace(incidentportability.StringFromAny(row["actor_id"]))
		if sourceActorID == "" {
			sourceActorID = strings.TrimSpace(incidentportability.StringFromAny(row["source_actor_id"]))
		}
		if _, parseErr := uuid.Parse(sourceActorID); parseErr != nil {
			return &VerificationError{
				ReasonCode:     "source_family_invalid",
				SourceFamilyID: "actors", InvariantID: "actors.reference_complete",
			}
		}
		if _, duplicate := descriptors[sourceActorID]; duplicate {
			return &VerificationError{
				ReasonCode:     "source_family_invalid",
				SourceFamilyID: "actors", InvariantID: "actors.reference_complete",
			}
		}
		descriptors[sourceActorID] = struct{}{}
	}
	referenced := map[string]struct{}{}
	for filePath, payload := range files {
		if filePath == "data/actors.ndjson" || filePath == "data/reference_pack_refs.json" ||
			!strings.HasPrefix(filePath, "data/") {
			continue
		}
		var sourceRows []map[string]any
		if filePath == "data/incident.json" {
			var row map[string]any
			if err := json.Unmarshal(bytes.TrimSpace(payload), &row); err != nil {
				return &VerificationError{ReasonCode: "source_family_invalid", SourceFamilyID: "incident", InvariantID: "incident.exact_shape"}
			}
			sourceRows = []map[string]any{row}
		} else {
			decoded, err := incidentportability.DecodeNDJSON(payload)
			if err != nil {
				continue
			}
			sourceRows = decoded
		}
		for _, row := range sourceRows {
			for key, value := range row {
				if !strings.HasSuffix(key, "_user_id") || value == nil {
					continue
				}
				actorID := strings.TrimSpace(incidentportability.StringFromAny(value))
				if actorID != "" {
					referenced[actorID] = struct{}{}
				}
			}
		}
	}
	for actorID := range referenced {
		if _, ok := descriptors[actorID]; !ok {
			return &VerificationError{
				ReasonCode:     "source_family_invalid",
				SourceFamilyID: "actors", InvariantID: "actors.reference_complete",
			}
		}
	}
	_ = incidentID
	return nil
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
			return &sourceport.Failure{FamilyID: "actors", InvariantID: "actors.reference_complete"}
		}
		displayName, _ := row["display_name"].(string)
		emailHint, _ := row["email_hint"].(string)
		tag, err := tx.Exec(ctx, `
INSERT INTO incident_bundle_imported_actors (incident_id, source_actor_id, display_name, email_hint, local_user_id)
VALUES ($1, $2, $3, $4, $5)
`, incidentID, sourceActorID, nullableString(displayName), nullableString(emailHint), nil)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return &sourceport.Failure{FamilyID: "actors", InvariantID: "actors.reference_complete"}
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

func (b *importedAttributionBuffer) RecordImportedAttribution(table string, sourceRowID string, column string, sourceActorID string) {
	sourceActorID = strings.TrimSpace(sourceActorID)
	sourceRowID = strings.TrimSpace(sourceRowID)
	if b == nil || sourceActorID == "" || sourceRowID == "" {
		return
	}
	b.rows = append(b.rows, importedAttribution{
		SourceTable:   table,
		SourceRowID:   sourceRowID,
		SourceColumn:  column,
		SourceActorID: sourceActorID,
	})
}

func (b *importedAttributionBuffer) flush(ctx context.Context, tx pgx.Tx) error {
	if b == nil {
		return nil
	}
	seen := map[string]struct{}{}
	for _, row := range b.rows {
		key := row.SourceTable + "\x00" + row.SourceRowID + "\x00" + row.SourceColumn
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		tag, err := tx.Exec(ctx, `
INSERT INTO incident_bundle_imported_attributions (
    incident_id,
    source_table,
    source_row_id,
    source_column,
    source_actor_id,
    local_user_id
)
VALUES ($1, $2, $3, $4, $5, $6)
`, b.IncidentID, row.SourceTable, row.SourceRowID, row.SourceColumn, row.SourceActorID, b.LocalUserID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return &incidentportability.VerificationFailure{ReasonCode: "duplicate_source_row"}
		}
	}
	return nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
