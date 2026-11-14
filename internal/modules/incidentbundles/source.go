package incidentbundles

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/artifactport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/importfinalizerport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// BlobPortability is the Incident Bundles-owned consumer port for blob data.
// Application assembly supplies the Evidence-owned implementation.
type BlobPortability interface {
	ExportBlobFiles(context.Context, incidentportability.Queryer, uuid.UUID, incidentportability.MemberWriter) error
	RewriteAndStageObjectBlobs(context.Context, incidentportability.MemberReader, incidentportability.MemberWriter, uuid.UUID, string) (*stagedobjects.PreparedResources, error)
}

// ArtifactReferenceSource exposes exact retained bindings in the same export
// snapshot. The source owner is responsible for its artifact identity checks.
type ArtifactReferenceSource interface {
	ReferenceBindingsTx(context.Context, pgx.Tx, uuid.UUID) ([]reference_data.SetBinding, error)
	ExportArtifactsTx(context.Context, pgx.Tx, uuid.UUID, []byte, artifactport.WriteFile) error
	PrepareArtifactImport(context.Context, artifactport.ImportRequest) (artifactport.Prepared, error)
}

type bundleBuilder struct {
	staging            stagedobjects.Allocator
	limits             Limits
	artifactReferences ArtifactReferenceSource
	pool               *pgxpool.Pool
	blobPort           BlobPortability
	portability        portabilityCoordinator
	sourceCatalog      sourcePortCatalog
	referencePacks     reference_data.IncidentReferences
}

type importer struct {
	artifactReferences ArtifactReferenceSource
	pool               *pgxpool.Pool
	blobPort           BlobPortability
	finalizer          importfinalizerport.Finalizer
	projectionRebuild  ImportProjectionRebuilder
	sourceCatalog      sourcePortCatalog
	referencePacks     reference_data.IncidentReferences
}

type builtIncidentBundle struct {
	scope       *stagedobjects.Scope
	Archive     bundleArchive
	IncidentKey string
}

type importParams struct {
	Execution          jobs.Execution
	ExecutionStarted   time.Time
	ReferenceExecution *reference_data.IncidentReferenceExecution
	ActorUserID        uuid.UUID
	PublishedAt        time.Time
	RequestID          *string
	OperationID        string
	RewrittenMembers   *memberStore
}

type preparedImport struct {
	IncidentID         uuid.UUID
	members            incidentportability.MemberReader
	rewrittenMembers   incidentportability.MemberReader
	stagedObjects      *stagedobjects.PreparedResources
	sourcePreparations []preparedSource
	importContext      sourceport.ImportContext
	referencePacks     *reference_data.PreparedReferenceImport
	artifacts          artifactport.Prepared
}

type preparedSource struct {
	port     sourceport.Port
	prepared sourceport.Prepared
}

func (p *preparedImport) cleanup(ctx context.Context) error {
	if p == nil {
		return nil
	}
	return p.stagedObjects.Abandon(context.WithoutCancel(ctx))
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

	scope, err := stagedobjects.NewScope("incident-bundle-export:"+bundleID.String(), ProfileID, b.staging)
	if err != nil {
		return builtIncidentBundle{}, err
	}
	retained := false
	defer func() {
		if !retained {
			_ = scope.Abandon(context.WithoutCancel(ctx))
		}
	}()
	members, err := newMemberStore("incident-bundle-export:"+bundleID.String(), scope, b.limits)
	if err != nil {
		return builtIncidentBundle{}, err
	}
	if b.sourceCatalog == nil {
		return builtIncidentBundle{}, errors.New("incident bundle source catalog is required")
	}
	actorReferences := newActorReferenceSet()
	for _, port := range b.sourceCatalog.Ports() {
		err := port.Export(ctx, sourceport.ExportContext{
			Members:              members,
			Actors:               actorReferences,
			Query:                snapshotReadOnly{tx},
			IncidentID:           incidentID,
			PortableAttributions: portableAttributionResolver{},
		})
		if err != nil {
			if port.Descriptor().FamilyID == "incident" && errors.Is(err, pgx.ErrNoRows) {
				return builtIncidentBundle{}, errNotFound
			}
			return builtIncidentBundle{}, err
		}
	}
	incidentJSON, err := readControlMember(ctx, members, "data/incident.json", PortabilityParticipantByteLimit)
	if err != nil {
		return builtIncidentBundle{}, err
	}
	var incidentIdentity struct {
		IncidentKey string `json:"incident_key"`
	}
	if err := json.Unmarshal(incidentJSON, &incidentIdentity); err != nil {
		return builtIncidentBundle{}, err
	}
	incidentKey := incidentIdentity.IncidentKey
	if b.referencePacks == nil {
		return builtIncidentBundle{}, errors.New("incident bundle Reference Pack retention is required")
	}
	if b.artifactReferences == nil {
		return builtIncidentBundle{}, errors.New("incident bundle artifact reference source is required")
	}
	bindings, err := b.artifactReferences.ReferenceBindingsTx(ctx, tx, incidentID)
	if err != nil {
		return builtIncidentBundle{}, err
	}
	references, err := b.referencePacks.ExportTx(ctx, tx, incidentID, bindings)
	if err != nil {
		return builtIncidentBundle{}, err
	}
	if _, err := members.WriteMember(ctx, "data/reference_pack_refs.json", bytes.NewReader(references)); err != nil {
		return builtIncidentBundle{}, err
	}
	if slices.Contains(request.OptionalSections, "snapshots") {
		if err := b.artifactReferences.ExportArtifactsTx(ctx, tx, incidentID, references, boundedContainerSink(members)); err != nil {
			return builtIncidentBundle{}, err
		}
	}
	contentManifest, err := b.referencePacks.ExportContentTx(ctx, tx, reference_data.IncidentReferenceExportRequest{
		IncidentID: incidentID, References: references, Embed: request.ReferencePackMode == referencePackModeEmbedded,
		WriteContainer: boundedContainerSink(members),
	})
	if err != nil {
		return builtIncidentBundle{}, err
	}
	if contentManifest != nil {
		if _, err := members.WriteMember(ctx, reference_data.IncidentPackContentPath, bytes.NewReader(contentManifest)); err != nil {
			return builtIncidentBundle{}, err
		}
	}
	if b.portability != nil {
		if err := b.portability.Export(ctx, tx, incidentID, members, actorReferences); err != nil {
			return builtIncidentBundle{}, err
		}
	}
	actors, err := b.exportActors(ctx, tx, incidentID, actorReferences)
	if err != nil {
		return builtIncidentBundle{}, err
	}
	if _, err := members.WriteMember(ctx, "data/actors.ndjson", bytes.NewReader(actors)); err != nil {
		return builtIncidentBundle{}, err
	}
	if b.blobPort == nil {
		return builtIncidentBundle{}, errors.New("incident bundle blob portability is required")
	}
	if err := b.blobPort.ExportBlobFiles(ctx, tx, incidentID, members); err != nil {
		return builtIncidentBundle{}, verificationErrorFromPort(err)
	}
	archive, err := buildBundleArchive(ctx, manifestInput{
		Limits:               &b.limits,
		BundleID:             bundleID.String(),
		IncidentID:           incidentID.String(),
		IncidentKey:          incidentKey,
		ExportedAt:           exportedAt.UTC().Format(time.RFC3339Nano),
		ReferencePackMode:    request.ReferencePackMode,
		OptionalSections:     request.OptionalSections,
		RequiredCapabilities: request.RequiredCapabilities,
	}, members)
	if err != nil {
		return builtIncidentBundle{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return builtIncidentBundle{}, err
	}
	retained = true
	return builtIncidentBundle{
		scope:       scope,
		Archive:     archive,
		IncidentKey: incidentKey,
	}, nil
}

func (b bundleBuilder) exportActors(ctx context.Context, q incidentportability.Queryer, incidentID uuid.UUID, references *actorReferenceSet) ([]byte, error) {
	if references == nil {
		return nil, errors.New("typed actor references are required")
	}
	actorIDs := references.ids

	if len(actorIDs) == 0 {
		return []byte{}, nil
	}
	ids := make([]string, 0, len(actorIDs))
	for id := range actorIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	descriptors := map[string]portableActorDescriptor{}
	var descriptorBytes int
	admitDescriptor := func(descriptor portableActorDescriptor) error {
		if _, exists := descriptors[descriptor.ActorID]; !exists {
			encoded, err := incidentportability.CanonicalJSONString(map[string]any{"actor_id": descriptor.ActorID, "display_name": descriptor.DisplayName, "email_hint": descriptor.EmailHint, "provider_subject_hint": descriptor.ProviderSubjectHint})
			if err != nil {
				return err
			}
			if len(encoded) > PortabilityParticipantByteLimit-descriptorBytes {
				return ErrPortabilityLimit
			}
			descriptorBytes += len(encoded)
		}
		return mergePortableActorDescriptor(descriptors, descriptor)
	}

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
		if err := admitDescriptor(descriptor); err != nil {
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
       email_hint,
       provider_subject_hint
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
			actorID             string
			displayName         sql.NullString
			emailHint           sql.NullString
			providerSubjectHint sql.NullString
		)
		if err := importedRows.Scan(&actorID, &displayName, &emailHint, &providerSubjectHint); err != nil {
			importedRows.Close()
			return nil, err
		}
		if err := admitDescriptor(portableActorDescriptor{
			ActorID:             actorID,
			DisplayName:         displayName.String,
			EmailHint:           emailHint.String,
			ProviderSubjectHint: providerSubjectHint.String,
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
		row := map[string]any{"actor_id": descriptor.ActorID, "display_name": descriptor.DisplayName}
		if descriptor.EmailHint != "" {
			row["email_hint"] = descriptor.EmailHint
		}
		if descriptor.ProviderSubjectHint != "" {
			row["provider_subject_hint"] = descriptor.ProviderSubjectHint
		}
		encoded, err := incidentportability.CanonicalJSONString(row)
		if err != nil {
			return nil, err
		}
		if len(encoded) > PortabilityParticipantByteLimit-payload.Len() {
			return nil, ErrPortabilityLimit
		}
		payload.Write(encoded)
	}
	return payload.Bytes(), nil
}

type portableActorDescriptor struct {
	ActorID             string
	DisplayName         string
	EmailHint           string
	ProviderSubjectHint string
}

func mergePortableActorDescriptor(descriptors map[string]portableActorDescriptor, descriptor portableActorDescriptor) error {
	if id, err := uuid.Parse(descriptor.ActorID); err != nil || id == uuid.Nil || id.String() != descriptor.ActorID {
		return actorReferenceFailure()
	}
	if existing, duplicate := descriptors[descriptor.ActorID]; duplicate && existing != descriptor {
		return actorReferenceFailure()
	}
	descriptors[descriptor.ActorID] = descriptor
	return nil
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
	importContext := sourceport.ImportContext{
		IncidentID: incidentID, ActorUserID: params.ActorUserID,
		BundleVersion: verified.Manifest.BundleVersion, OperationID: params.OperationID,
	}
	actorCatalog, err := validateImportedActors(ctx, verified.Members, incidentID)
	if err != nil {
		return nil, err
	}
	importContext.Actors = actorCatalog
	references, err := incidentportability.ReadControl(ctx, verified.Members, "data/reference_pack_refs.json", PortabilityParticipantByteLimit)
	if err != nil {
		return nil, err
	}
	if err := reference_data.ValidateIncidentBundleReferences(references); err != nil {
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
	if i.referencePacks == nil {
		return nil, errors.New("incident bundle Reference Pack retention is required")
	}
	operationID, err := uuid.Parse(params.OperationID)
	if err != nil {
		return nil, errors.New("incident bundle import operation identity is invalid")
	}
	referenceRequest := reference_data.IncidentReferenceImportRequest{References: references, IncidentID: incidentID, OperationID: operationID, ActorID: params.ActorUserID, At: params.PublishedAt, Execution: params.Execution, ExecutionStarted: params.ExecutionStarted, ExecutionScope: params.ReferenceExecution, OpenContainer: func(readCtx context.Context, path string) (io.ReadCloser, error) {
		if err := readCtx.Err(); err != nil {
			return nil, err
		}
		reader, _, err := verified.Members.OpenMember(readCtx, path)
		return reader, err
	}}
	for _, member := range verified.Members.Members() {
		if member.Path == reference_data.IncidentPackContentPath {
			content, err := incidentportability.ReadControl(ctx, verified.Members, member.Path, PortabilityParticipantByteLimit)
			if err != nil {
				return nil, err
			}
			referenceRequest.ContentManifest = &content
		} else if strings.HasPrefix(member.Path, "ext/reference_packs/") {
			referenceRequest.EmbeddedPaths = append(referenceRequest.EmbeddedPaths, member.Path)
		}
	}
	referencePacks, err := i.referencePacks.PrepareImport(ctx, referenceRequest)
	if err != nil {
		return nil, referenceImportError(err)
	}
	if i.artifactReferences == nil {
		return nil, errors.New("incident bundle artifact source is required")
	}
	artifacts, err := i.artifactReferences.PrepareArtifactImport(ctx, artifactport.ImportRequest{IncidentID: incidentID, OperationID: operationID, References: references, Bundle: verified.Members})
	if err != nil {
		return nil, verificationErrorFromPort(err)
	}
	sourcePreparations := make([]preparedSource, 0, len(i.sourceCatalog.Ports()))
	bundle := verified.Members
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
	if params.RewrittenMembers == nil {
		return nil, errors.New("incident import member staging is required")
	}
	stagedObjects, err := i.blobPort.RewriteAndStageObjectBlobs(ctx, verified.Members, params.RewrittenMembers, incidentID, params.OperationID)
	if err != nil {
		return nil, verificationErrorFromPort(err)
	}
	params.RewrittenMembers.Seal()
	return &preparedImport{
		IncidentID: incidentID, members: verified.Members, rewrittenMembers: params.RewrittenMembers,
		stagedObjects:      stagedObjects,
		sourcePreparations: sourcePreparations, importContext: importContext,
		referencePacks: referencePacks, artifacts: artifacts,
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
	attributions := transactionAttributions{ctx: ctx, tx: tx, incidentID: incidentID, localUserID: params.ActorUserID, actors: prepared.importContext.Actors}
	importContext := prepared.importContext
	importContext.Attributions = &attributions
	importContext.RewrittenObjectBlobs = prepared.rewrittenMembers
	if len(prepared.sourcePreparations) == 0 ||
		prepared.sourcePreparations[0].port.Descriptor().FamilyID != "incident" {
		return uuid.UUID{}, errors.New("incident source port must be first")
	}
	if err := prepared.sourcePreparations[0].port.ApplyImportTx(
		ctx, tx, prepared.sourcePreparations[0].prepared, importContext,
	); err != nil {
		return uuid.UUID{}, verificationErrorFromDeclaredPort(prepared.sourcePreparations[0].port, err)
	}
	if err := i.importActors(ctx, tx, prepared.members, incidentID); err != nil {
		return uuid.UUID{}, verificationErrorFromPort(err)
	}
	if i.referencePacks == nil {
		return uuid.Nil, errors.New("incident bundle Reference Pack retention is required")
	}
	if err := i.referencePacks.ApplyImportTx(ctx, tx, prepared.referencePacks); err != nil {
		return uuid.Nil, referenceImportError(err)
	}
	if prepared.artifacts == nil {
		return uuid.Nil, errors.New("incident bundle artifact preparation is required")
	}
	if err := prepared.artifacts.ApplyTx(ctx, tx); err != nil {
		return uuid.Nil, verificationErrorFromPort(err)
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
	return incidentID, nil
}

func referenceImportError(err error) error {
	if invariant, ok := reference_data.IncidentBundleReferenceInvariant(err); ok {
		return &verificationError{ReasonCode: "source_family_invalid", SourceFamilyID: "reference_pack_refs", InvariantID: invariant}
	}
	return err
}

func revisionsSequenceRepairVerificationError() error {
	return &verificationError{
		ReasonCode:     "source_family_invalid",
		SourceFamilyID: "revisions",
		InvariantID:    "revisions.sequence_repair_after_validation",
	}
}

func verificationErrorFromPort(err error) error {
	if errors.Is(err, artifactport.ErrInvalid) {
		return &verificationError{ReasonCode: "malformed_manifest"}
	}
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

func validateImportedActors(ctx context.Context, members incidentportability.MemberReader, incidentID uuid.UUID) (sourceport.ActorCatalog, error) {
	descriptors := []sourceport.ActorDescriptor{}
	var rosterBytes int64
	var previousActorID string
	err := incidentportability.WalkPath(ctx, members, "data/actors.ndjson", func(row map[string]any) error {
		encoded, err := incidentportability.CanonicalJSONString(row)
		if err != nil {
			return actorReferenceFailure()
		}
		rosterBytes += int64(len(encoded))
		if rosterBytes > PortabilityParticipantByteLimit {
			return actorReferenceFailure()
		}
		descriptor, err := decodePortableActorDescriptor(row)
		if err != nil {
			return err
		}
		if descriptor.ActorID <= previousActorID {
			return actorReferenceFailure()
		}
		previousActorID = descriptor.ActorID
		descriptors = append(descriptors, sourceport.ActorDescriptor{SourceActorID: descriptor.ActorID, DisplayName: descriptor.DisplayName, EmailHint: descriptor.EmailHint, ProviderSubjectHint: descriptor.ProviderSubjectHint})
		return nil
	})
	if err != nil {
		return sourceport.ActorCatalog{}, err
	}
	catalog, err := sourceport.NewActorCatalog(descriptors)
	if err != nil {
		return sourceport.ActorCatalog{}, actorReferenceFailure()
	}
	for _, member := range members.Members() {
		if member.Path == "data/actors.ndjson" || !strings.HasPrefix(member.Path, "data/") || !strings.HasSuffix(member.Path, ".ndjson") {
			continue
		}
		missingActor := false
		err := incidentportability.WalkMember(ctx, members, member, func(row map[string]any) error {
			for key, value := range row {
				if !strings.HasSuffix(key, "_user_id") || value == nil {
					continue
				}
				id := strings.TrimSpace(incidentportability.StringFromAny(value))
				if id != "" {
					if _, ok := catalog.Lookup(id); !ok {
						missingActor = true
						return actorReferenceFailure()
					}
				}
			}
			return nil
		})
		if missingActor {
			return sourceport.ActorCatalog{}, actorReferenceFailure()
		}
		// The source owner supplies its declared row-shape diagnostic. Stream
		// integrity and cancellation failures must never be swallowed here.
		var malformed *incidentportability.StrictNDJSONDecodeError
		if err != nil && !errors.As(err, &malformed) {
			return sourceport.ActorCatalog{}, err
		}
	}
	_ = incidentID
	return catalog, nil
}

func (i importer) importActors(ctx context.Context, tx pgx.Tx, members incidentportability.MemberReader, incidentID uuid.UUID) error {
	return incidentportability.WalkPath(ctx, members, "data/actors.ndjson", func(row map[string]any) error {
		descriptor, err := decodePortableActorDescriptor(row)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `INSERT INTO incident_bundle_imported_actors (incident_id, source_actor_id, display_name, email_hint, provider_subject_hint, local_user_id) VALUES ($1, $2, $3, $4, $5, $6)`, incidentID, descriptor.ActorID, descriptor.DisplayName, nullableString(descriptor.EmailHint), nullableString(descriptor.ProviderSubjectHint), nil)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return actorReferenceFailure()
		}
		return nil
	})
}

func decodePortableActorDescriptor(row map[string]any) (portableActorDescriptor, error) {
	for key := range row {
		switch key {
		case "actor_id", "display_name", "email_hint", "provider_subject_hint":
		default:
			return portableActorDescriptor{}, &verificationError{ReasonCode: "source_family_invalid", SourceFamilyID: "actors", InvariantID: "actors.inert"}
		}
	}
	actorID, ok := row["actor_id"].(string)
	id, err := uuid.Parse(actorID)
	if !ok || err != nil || id == uuid.Nil || id.String() != actorID {
		return portableActorDescriptor{}, actorReferenceFailure()
	}
	displayName, ok := row["display_name"].(string)
	if !ok {
		return portableActorDescriptor{}, actorReferenceFailure()
	}
	descriptor := portableActorDescriptor{ActorID: actorID, DisplayName: displayName}
	for _, hint := range []struct {
		key    string
		target *string
	}{
		{"email_hint", &descriptor.EmailHint},
		{"provider_subject_hint", &descriptor.ProviderSubjectHint},
	} {
		if value, present := row[hint.key]; present {
			text, ok := value.(string)
			if !ok || text == "" {
				return portableActorDescriptor{}, actorReferenceFailure()
			}
			*hint.target = text
		}
	}
	return descriptor, nil
}

func actorReferenceFailure() error {
	return &verificationError{
		ReasonCode:     "source_family_invalid",
		SourceFamilyID: "actors",
		InvariantID:    "actors.reference_complete",
	}
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
