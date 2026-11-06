package networkflow

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// normalizedRow is the write input from record normalization. Storage identity,
// import attribution and observation provenance cannot be supplied through it.
type normalizedRow struct {
	Direction                 flowDirection
	DirectionEvidence         json.RawMessage
	SourceRowNumber           int64
	SourceRowDigestSHA256     string
	NormalizedRowDigestSHA256 string
	FlowStartUTC              time.Time
	FlowEndUTC                time.Time
	EndpointAIP               string
	EndpointBIP               string
	EndpointAPort             *int32
	EndpointBPort             *int32
	IPProtocol                int32
	BytesCount                string
	PacketsCount              string
	ExporterID                *string
	InputInterface            *string
	OutputInterface           *string
	TCPFlags                  *int32
	ApplicationLabel          *string
	UnmappedRaw               json.RawMessage
}

var errInvalidImportContext = errors.New("invalid network flow import context")
var errInvalidNormalizedRow = errors.New("invalid normalized network flow row")

type rowImportContext struct {
	incidentID, actorUserID, committedSourceID           uuid.UUID
	sourceSHA, fingerprint, sourceProfile, parserProfile string
	createdAt                                            time.Time
}

// Validate before any table write. Only Imports supplies this context in product
// code; the borrowed transaction remains under its commit/rollback authority.
func validatedRowImportContext(params createTableParams) (rowImportContext, error) {
	profile, knownProfile := mappingSourceProfile(params.SourceProfileID)
	now := params.Now.UTC()
	if params.IncidentID == uuid.Nil || params.ActorUserID == uuid.Nil || params.ImportSessionID == uuid.Nil || params.ImportUnitID == uuid.Nil || params.CommittedSourceID == uuid.Nil ||
		!validLowerSHA256(params.SourceInterpretationSHA256) || params.SourceRawCellScalarLimit < 1 || params.SourceRawCellScalarLimit > 16384 ||
		!validLowerSHA256(params.SourceContentSHA256) || !validLowerSHA256(params.MappingFingerprint) ||
		!validLowerSHA256(params.SourceFilenameDigest) || !safeKeyIDPattern.MatchString(params.SourceFilenameDigestKeyID) ||
		!knownProfile || profile.ConformanceStatus != "required_v1" || params.ParserProfileID == "" || params.ParserProfileID != profile.ParserProfileID ||
		params.Now.IsZero() || now.Year() < 1 || now.Year() > 9999 {
		return rowImportContext{}, errInvalidImportContext
	}
	return rowImportContext{
		incidentID: params.IncidentID, actorUserID: params.ActorUserID,
		committedSourceID: params.CommittedSourceID,
		sourceSHA:         params.SourceContentSHA256, fingerprint: params.MappingFingerprint,
		sourceProfile: params.SourceProfileID, parserProfile: params.ParserProfileID,
		createdAt: now,
	}, nil
}

func (context rowImportContext) persistedRow(tableID string, source normalizedRow) (flowRow, error) {
	if len(tableID) != 36 || !strings.HasPrefix(tableID, "nft_") || strings.ToLower(tableID) != tableID {
		return flowRow{}, errInvalidImportContext
	}
	if _, err := hex.DecodeString(tableID[4:]); err != nil {
		return flowRow{}, errInvalidImportContext
	}
	if source.SourceRowNumber < 1 || !validLowerSHA256(source.SourceRowDigestSHA256) || !validLowerSHA256(source.NormalizedRowDigestSHA256) {
		return flowRow{}, errInvalidNormalizedRow
	}
	if err := validateObservationDirection(flowEndpoint{source.EndpointAIP, source.EndpointAPort}, flowEndpoint{source.EndpointBIP, source.EndpointBPort}, source.Direction, source.DirectionEvidence, context.sourceProfile); err != nil {
		return flowRow{}, err
	}
	var retained retainedRawValues
	if err := json.Unmarshal(source.UnmappedRaw, &retained); err != nil {
		return flowRow{}, fmt.Errorf("%w: unmapped values: %w", errInvalidNormalizedRow, err)
	}
	provenance, err := json.Marshal(map[string]any{
		"committed_source_id":      context.committedSourceID.String(),
		"source_content_sha256":    context.sourceSHA,
		"source_profile_id":        context.sourceProfile,
		"parser_profile_id":        context.parserProfile,
		"mapping_fingerprint":      context.fingerprint,
		"source_row_number":        source.SourceRowNumber,
		"source_row_digest_sha256": source.SourceRowDigestSHA256,
	})
	if err != nil {
		return flowRow{}, fmt.Errorf("encode network flow observation provenance: %w", err)
	}
	return flowRow{
		RowID:              rowID(context.incidentID, tableID, source.SourceRowNumber, source.SourceRowDigestSHA256, source.NormalizedRowDigestSHA256),
		NetworkFlowTableID: tableID, IncidentID: context.incidentID,
		SourceRowNumber:           source.SourceRowNumber,
		SourceRowDigestSHA256:     source.SourceRowDigestSHA256,
		NormalizedRowDigestSHA256: source.NormalizedRowDigestSHA256,
		MappingFingerprint:        context.fingerprint,
		Direction:                 source.Direction, DirectionEvidence: append(json.RawMessage(nil), source.DirectionEvidence...),
		FlowStartUTC: source.FlowStartUTC, FlowEndUTC: source.FlowEndUTC,
		EndpointAIP: source.EndpointAIP, EndpointBIP: source.EndpointBIP,
		EndpointAPort: source.EndpointAPort, EndpointBPort: source.EndpointBPort, IPProtocol: source.IPProtocol,
		BytesCount: source.BytesCount, PacketsCount: source.PacketsCount,
		ExporterID: source.ExporterID, InputInterface: source.InputInterface, OutputInterface: source.OutputInterface,
		TCPFlags: source.TCPFlags, ApplicationLabel: source.ApplicationLabel,
		UnmappedRaw:          append(json.RawMessage(nil), source.UnmappedRaw...),
		ObservationSourceRef: provenance, CreatedAt: context.createdAt, CreatedByUserID: context.actorUserID,
	}, nil
}
