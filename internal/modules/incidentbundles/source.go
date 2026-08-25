package incidentbundles

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/importfinalizerport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

// BlobPortability is the Incident Bundles-owned consumer port for blob data.
// Application assembly supplies the Evidence-owned implementation.
type BlobPortability interface {
	ExportBlobFiles(context.Context, incidentportability.Queryer, uuid.UUID, map[string][]byte) error
	RewriteAndStageObjectBlobs(context.Context, map[string][]byte, uuid.UUID, uuid.UUID, incidentportability.AttributionRecorder) ([]byte, []string, error)
	CleanupStagedObjects(context.Context, []string)
}

type bundleBuilder struct {
	pool          *pgxpool.Pool
	blobPort      BlobPortability
	portability   portabilityCoordinator
	sourceCatalog sourcePortCatalog
}

type importer struct {
	pool              *pgxpool.Pool
	blobPort          BlobPortability
	finalizer         importfinalizerport.Finalizer
	projectionRebuild ImportProjectionRebuilder
	sourceCatalog     sourcePortCatalog
}

type builtIncidentBundle struct {
	Archive        bundleArchive
	IncidentKey    string
	BundleSHA256   string
	BundleByteSize int64
}

type importParams struct {
	ActorUserID uuid.UUID
	PublishedAt time.Time
	RequestID   *string
	OperationID string
}

type preparedImport struct {
	IncidentID         uuid.UUID
	files              map[string][]byte
	attributions       importedAttributionBuffer
	stagedObjectKeys   []string
	blobPort           BlobPortability
	sourcePreparations []preparedSource
	importContext      sourceport.ImportContext
}

type preparedSource struct {
	port     sourceport.Port
	prepared sourceport.Prepared
}

func (p *preparedImport) cleanup(ctx context.Context) {
	if p == nil || len(p.stagedObjectKeys) == 0 {
		return
	}
	p.blobPort.CleanupStagedObjects(context.WithoutCancel(ctx), p.stagedObjectKeys)
	p.stagedObjectKeys = nil
}

// ImportProjectionRebuilder reconstructs imported incident projections inside
// the final publication transaction.
type ImportProjectionRebuilder interface {
	RebuildImportedIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error
}

func (b bundleBuilder) build(ctx context.Context, incidentID uuid.UUID, request exportRequest, bundleID uuid.UUID, exportedAt time.Time) (builtIncidentBundle, error) {
	tx, err := b.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return builtIncidentBundle{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	files := map[string][]byte{}
	if b.sourceCatalog == nil {
		return builtIncidentBundle{}, errors.New("incident bundle source catalog is required")
	}
	for _, port := range b.sourceCatalog.Ports() {
		exportedFiles, err := port.Export(ctx, sourceport.ExportContext{
			Query:                tx,
			IncidentID:           incidentID,
			PortableAttributions: portableAttributionResolver{},
		})
		if err != nil {
			if port.Descriptor().FamilyID == "incident" && errors.Is(err, pgx.ErrNoRows) {
				return builtIncidentBundle{}, errNotFound
			}
			return builtIncidentBundle{}, err
		}
		for _, file := range exportedFiles {
			files[file.Path] = file.Payload
		}
	}
	incidentJSON, ok := files["data/incident.json"]
	if !ok {
		return builtIncidentBundle{}, errors.New("incident bundle incident source is required")
	}
	var incidentIdentity struct {
		IncidentKey string `json:"incident_key"`
	}
	if err := json.Unmarshal(incidentJSON, &incidentIdentity); err != nil {
		return builtIncidentBundle{}, err
	}
	incidentKey := incidentIdentity.IncidentKey
	files["data/reference_pack_refs.json"] = []byte("[]\n")
	actors, err := b.exportActors(ctx, tx, incidentID, files)
	if err != nil {
		return builtIncidentBundle{}, err
	}
	files["data/actors.ndjson"] = actors
	if b.blobPort == nil {
		return builtIncidentBundle{}, errors.New("incident bundle blob portability is required")
	}
	if err := b.blobPort.ExportBlobFiles(ctx, tx, incidentID, files); err != nil {
		return builtIncidentBundle{}, verificationErrorFromPort(err)
	}
	if b.portability != nil {
		payloads, err := b.portability.Export(ctx, tx, incidentID)
		if err != nil {
			return builtIncidentBundle{}, err
		}
		for _, payload := range payloads {
			filePath, encoded, err := EncodeExtensionPayload(payload)
			if err != nil {
				return builtIncidentBundle{}, err
			}
			files[filePath] = encoded
		}
	}
	archive, err := buildBundleArchive(manifestInput{
		BundleID:             bundleID.String(),
		IncidentID:           incidentID.String(),
		IncidentKey:          incidentKey,
		ExportedAt:           exportedAt.UTC().Format(time.RFC3339Nano),
		ReferencePackMode:    request.ReferencePackMode,
		OptionalSections:     request.OptionalSections,
		RequiredCapabilities: request.RequiredCapabilities,
	}, files)
	if err != nil {
		return builtIncidentBundle{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return builtIncidentBundle{}, err
	}
	return builtIncidentBundle{
		Archive:        archive,
		IncidentKey:    incidentKey,
		BundleSHA256:   hashHex(archive.Bytes),
		BundleByteSize: int64(len(archive.Bytes)),
	}, nil
}

func (b bundleBuilder) exportActors(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID, files map[string][]byte) ([]byte, error) {
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
	descriptors := map[string]portableActorDescriptor{}
	localRows, err := q.Query(ctx, `
SELECT u.id::text,
       u.display_name,
       u.email::text
  FROM users u
 WHERE u.id IN (SELECT unnest($1::text[])::uuid)
 ORDER BY u.id
`, ids)
	if err != nil {
		return nil, err
	}
	for localRows.Next() {
		var descriptor portableActorDescriptor
		if err := localRows.Scan(&descriptor.ActorID, &descriptor.DisplayName, &descriptor.EmailHint); err != nil {
			localRows.Close()
			return nil, err
		}
		if err := mergePortableActorDescriptor(descriptors, descriptor); err != nil {
			localRows.Close()
			return nil, err
		}
	}
	if err := localRows.Err(); err != nil {
		localRows.Close()
		return nil, err
	}
	localRows.Close()

	importedRows, err := q.Query(ctx, `
SELECT source_actor_id,
       display_name,
       email_hint
  FROM incident_bundle_imported_actors
 WHERE incident_id = $1
   AND source_actor_id IN (SELECT unnest($2::text[]))
 ORDER BY source_actor_id
`, incidentID, ids)
	if err != nil {
		return nil, err
	}
	for importedRows.Next() {
		var (
			actorID     string
			displayName sql.NullString
			emailHint   sql.NullString
		)
		if err := importedRows.Scan(&actorID, &displayName, &emailHint); err != nil {
			importedRows.Close()
			return nil, err
		}
		if err := mergePortableActorDescriptor(descriptors, portableActorDescriptor{
			ActorID:     actorID,
			DisplayName: displayName.String,
			EmailHint:   emailHint.String,
		}); err != nil {
			importedRows.Close()
			return nil, err
		}
	}
	if err := importedRows.Err(); err != nil {
		importedRows.Close()
		return nil, err
	}
	importedRows.Close()

	var payload bytes.Buffer
	for _, actorID := range ids {
		descriptor, ok := descriptors[actorID]
		if !ok {
			return nil, actorReferenceFailure()
		}
		row := map[string]any{"actor_id": descriptor.ActorID}
		if descriptor.DisplayName != "" {
			row["display_name"] = descriptor.DisplayName
		}
		if descriptor.EmailHint != "" {
			row["email_hint"] = descriptor.EmailHint
		}
		encoded, err := incidentportability.CanonicalJSONString(row)
		if err != nil {
			return nil, err
		}
		payload.Write(encoded)
	}
	return payload.Bytes(), nil
}

type portableActorDescriptor struct {
	ActorID     string
	DisplayName string
	EmailHint   string
}

func mergePortableActorDescriptor(descriptors map[string]portableActorDescriptor, descriptor portableActorDescriptor) error {
	if _, err := uuid.Parse(descriptor.ActorID); err != nil {
		return actorReferenceFailure()
	}
	if existing, duplicate := descriptors[descriptor.ActorID]; duplicate && existing != descriptor {
		return actorReferenceFailure()
	}
	descriptors[descriptor.ActorID] = descriptor
	return nil
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

func (i importer) prepareImport(ctx context.Context, verified verifiedBundle, params importParams) (*preparedImport, error) {
	if params.OperationID == "" {
		return nil, errors.New("incident bundle import operation ID is required")
	}
	incidentID, err := uuid.Parse(verified.Manifest.IncidentID)
	if err != nil {
		return nil, &verificationError{ReasonCode: "malformed_manifest"}
	}
	var existing bool
	if err := i.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM incidents WHERE id = $1)`, incidentID).Scan(&existing); err != nil {
		return nil, err
	}
	if existing {
		return nil, &verificationError{ReasonCode: "duplicate_incident_id"}
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
	actorCatalog, err := validateImportedActors(verified.Files, incidentID)
	if err != nil {
		return nil, err
	}
	importContext.Actors = actorCatalog
	if err := reference_data.ValidateIncidentBundleReferences(verified.Files["data/reference_pack_refs.json"]); err != nil {
		invariantID, ok := reference_data.IncidentBundleReferenceInvariant(err)
		if !ok {
			return nil, err
		}
		return nil, &verificationError{
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
			return nil, verificationErrorFromDeclaredPort(port, err)
		}
		sourcePreparations = append(sourcePreparations, preparedSource{port: port, prepared: prepared})
	}
	if i.blobPort == nil {
		return nil, errors.New("incident bundle blob portability is required")
	}
	rewrittenObjectBlobs, writtenObjectKeys, err := i.blobPort.RewriteAndStageObjectBlobs(ctx, verified.Files, incidentID, params.ActorUserID, &attributions)
	if err != nil {
		i.blobPort.CleanupStagedObjects(context.WithoutCancel(ctx), writtenObjectKeys)
		return nil, verificationErrorFromPort(err)
	}
	importFiles := make(map[string][]byte, len(verified.Files))
	for path, payload := range verified.Files {
		importFiles[path] = append([]byte(nil), payload...)
	}
	importFiles["data/object_blobs.ndjson"] = rewrittenObjectBlobs
	return &preparedImport{
		IncidentID: incidentID, files: importFiles, attributions: attributions,
		stagedObjectKeys: writtenObjectKeys, blobPort: i.blobPort,
		sourcePreparations: sourcePreparations, importContext: importContext,
	}, nil
}

func (i importer) applyPreparedImportTx(ctx context.Context, tx pgx.Tx, prepared *preparedImport, params importParams) (uuid.UUID, error) {
	if tx == nil || prepared == nil || prepared.IncidentID == uuid.Nil {
		return uuid.UUID{}, errors.New("prepared incident bundle import is required")
	}
	revisionSequenceOriginalNext, err := revisions.BeginIncidentBundleImportedRevisionSequenceTx(ctx, tx)
	if err != nil {
		return uuid.UUID{}, revisionsSequenceRepairVerificationError()
	}
	incidentID := prepared.IncidentID
	var existing int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID).Scan(&existing); err != nil {
		return uuid.UUID{}, err
	}
	if existing > 0 {
		return uuid.UUID{}, &verificationError{ReasonCode: "duplicate_incident_id"}
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
		return uuid.UUID{}, verificationErrorFromDeclaredPort(prepared.sourcePreparations[0].port, err)
	}
	if err := i.importActors(ctx, tx, prepared.files["data/actors.ndjson"], incidentID); err != nil {
		return uuid.UUID{}, verificationErrorFromPort(err)
	}
	for _, source := range prepared.sourcePreparations[1:] {
		if err := source.port.ApplyImportTx(ctx, tx, source.prepared, importContext); err != nil {
			return uuid.UUID{}, verificationErrorFromDeclaredPort(source.port, err)
		}
	}
	for _, source := range prepared.sourcePreparations {
		if err := source.port.ValidateImportTx(ctx, tx, source.prepared, importContext); err != nil {
			return uuid.UUID{}, verificationErrorFromDeclaredPort(source.port, err)
		}
	}
	if err := revisions.FinishIncidentBundleImportedRevisionSequenceTx(
		ctx,
		tx,
		revisionSequenceOriginalNext,
	); err != nil {
		return uuid.UUID{}, revisionsSequenceRepairVerificationError()
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
	if err := i.finalizer.FinalizeIncidentBundleImportTx(ctx, tx, importfinalizerport.Params{
		IncidentID:        incidentID,
		SubmittedByUserID: params.ActorUserID,
		PublishedAt:       params.PublishedAt,
		RequestID:         params.RequestID,
	}); err != nil {
		if errors.Is(err, importfinalizerport.ErrInitialAdminUnavailable) {
			return uuid.UUID{}, &verificationError{ReasonCode: "initial_admin_unavailable"}
		}
		return uuid.UUID{}, err
	}
	prepared.attributions = attributions
	return incidentID, nil
}

func revisionsSequenceRepairVerificationError() error {
	return &verificationError{
		ReasonCode:     "source_family_invalid",
		SourceFamilyID: "revisions",
		InvariantID:    "revisions.sequence_repair_after_validation",
	}
}

func verificationErrorFromPort(err error) error {
	var malformed *incidentportability.MalformedPayloadError
	if errors.As(err, &malformed) {
		return &verificationError{ReasonCode: "malformed_manifest"}
	}
	var verification *incidentportability.VerificationFailure
	if errors.As(err, &verification) {
		return &verificationError{ReasonCode: verification.ReasonCode}
	}
	var sourceFailure *sourceport.Failure
	if errors.As(err, &sourceFailure) {
		return &verificationError{
			ReasonCode:     "source_family_invalid",
			SourceFamilyID: sourceFailure.FamilyID(),
			InvariantID:    sourceFailure.InvariantID(),
		}
	}
	return err
}

func verificationErrorFromDeclaredPort(port sourceport.Port, err error) error {
	var sourceFailure *sourceport.Failure
	if !errors.As(err, &sourceFailure) {
		return verificationErrorFromPort(err)
	}
	if port == nil {
		return errors.New("incident bundle source port returned a failure without a descriptor")
	}
	descriptor := port.Descriptor()
	familyID, invariantID := sourceFailure.FamilyID(), sourceFailure.InvariantID()
	if familyID != descriptor.FamilyID {
		return errors.New("incident bundle source port returned an undeclared failure")
	}
	if declared := descriptor.DeclaredFailure(invariantID); errors.Is(declared, sourceport.ErrInvalidCatalog) {
		return errors.New("incident bundle source port returned an undeclared failure")
	}
	return &verificationError{
		ReasonCode:     "source_family_invalid",
		SourceFamilyID: descriptor.FamilyID,
		InvariantID:    invariantID,
	}
}

func validateImportedActors(files map[string][]byte, incidentID uuid.UUID) (sourceport.ActorCatalog, error) {
	rows, err := incidentportability.DecodeNDJSON(files["data/actors.ndjson"])
	if err != nil {
		return sourceport.ActorCatalog{}, &verificationError{
			ReasonCode:     "source_family_invalid",
			SourceFamilyID: "actors", InvariantID: "actors.reference_complete",
		}
	}
	descriptors := make([]sourceport.ActorDescriptor, 0, len(rows))
	for _, row := range rows {
		for key := range row {
			switch key {
			case "actor_id", "source_actor_id", "display_name", "email_hint", "provider_subject_hint":
			default:
				return sourceport.ActorCatalog{}, &verificationError{
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
			return sourceport.ActorCatalog{}, &verificationError{
				ReasonCode:     "source_family_invalid",
				SourceFamilyID: "actors", InvariantID: "actors.reference_complete",
			}
		}
		displayName, _ := row["display_name"].(string)
		emailHint, _ := row["email_hint"].(string)
		providerSubjectHint, _ := row["provider_subject_hint"].(string)
		descriptors = append(descriptors, sourceport.ActorDescriptor{
			SourceActorID:       sourceActorID,
			DisplayName:         displayName,
			EmailHint:           emailHint,
			ProviderSubjectHint: providerSubjectHint,
		})
	}
	catalog, err := sourceport.NewActorCatalog(descriptors)
	if err != nil {
		return sourceport.ActorCatalog{}, &verificationError{
			ReasonCode:     "source_family_invalid",
			SourceFamilyID: "actors", InvariantID: "actors.reference_complete",
		}
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
				return sourceport.ActorCatalog{}, &verificationError{ReasonCode: "source_family_invalid", SourceFamilyID: "incident", InvariantID: "incident.exact_shape"}
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
		if _, ok := catalog.Lookup(actorID); !ok {
			return sourceport.ActorCatalog{}, &verificationError{
				ReasonCode:     "source_family_invalid",
				SourceFamilyID: "actors", InvariantID: "actors.reference_complete",
			}
		}
	}
	_ = incidentID
	return catalog, nil
}

func (i importer) importActors(ctx context.Context, tx pgx.Tx, payload []byte, incidentID uuid.UUID) error {
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
			return actorReferenceFailure()
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
			return actorReferenceFailure()
		}
	}
	return nil
}

func actorReferenceFailure() error {
	return &verificationError{
		ReasonCode:     "source_family_invalid",
		SourceFamilyID: "actors",
		InvariantID:    "actors.reference_complete",
	}
}

type importedAttributionBuffer struct {
	IncidentID  uuid.UUID
	LocalUserID uuid.UUID
	rows        []incidentportability.ImportedAttribution
}

func (b *importedAttributionBuffer) RecordImportedAttribution(table string, sourceRowID string, column string, sourceActorID string) error {
	sourceActorID = strings.TrimSpace(sourceActorID)
	sourceRowID = strings.TrimSpace(sourceRowID)
	if b == nil || sourceActorID == "" || sourceRowID == "" ||
		strings.TrimSpace(table) == "" || strings.TrimSpace(column) == "" {
		return errors.New("incident bundle attribution is invalid")
	}
	if _, err := uuid.Parse(sourceActorID); err != nil {
		return errors.New("incident bundle attribution is invalid")
	}
	for _, row := range b.rows {
		if row.SourceTable == table && row.SourceRowID == sourceRowID && row.SourceColumn == column {
			if row.SourceActorID == sourceActorID && row.LocalUserID == b.LocalUserID {
				return nil
			}
			return errors.New("incident bundle attribution conflicts")
		}
	}
	b.rows = append(b.rows, incidentportability.ImportedAttribution{
		SourceTable:   table,
		SourceRowID:   sourceRowID,
		SourceColumn:  column,
		SourceActorID: sourceActorID,
		LocalUserID:   b.LocalUserID,
	})
	return nil
}

func (b *importedAttributionBuffer) ImportedAttributions() []incidentportability.ImportedAttribution {
	if b == nil {
		return nil
	}
	return append([]incidentportability.ImportedAttribution(nil), b.rows...)
}

func (b *importedAttributionBuffer) flush(ctx context.Context, tx pgx.Tx) error {
	if b == nil {
		return nil
	}
	for _, row := range b.rows {
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
`, b.IncidentID, row.SourceTable, row.SourceRowID, row.SourceColumn, row.SourceActorID, row.LocalUserID)
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
