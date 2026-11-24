package networkflow

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type importFacade struct {
	store        *store
	sourceStore  ImportSourcePort
	limits       EffectiveLimits
	now          func() time.Time
	safeDigester safeDigester
}

const (
	importFacadeBindingSchemaID          = "cartulary.imports.analytical_facade_binding.v1"
	importFacadeID                       = "network_flow_import_facade_v1"
	importOwnerContractRef               = "network_flow_activity@10"
	importMappingSchemaID                = "cartulary.network_flow.approved_mapping.v2"
	importPreviewRequestID               = "cartulary.network_flow.import_preview_request.v2"
	importPreviewResultSchemaID          = "cartulary.network_flow.import_preview_result.v2"
	importSourceInspectionResultSchemaID = "cartulary.network_flow.import_source_inspection_result.v2"
	importSourceInspectionRequestID      = "cartulary.network_flow.import_source_inspection_request.v1"
	importApplyRequestID                 = "cartulary.network_flow.import_apply_request.v2"
	importApplyResultID                  = "cartulary.network_flow.import_unit_result.v2"
	importCommitProtocolID               = "cartulary.imports.unit_commit.v1"
)

func newImportFacade(store *store, sourceStore ImportSourcePort, limits EffectiveLimits, now func() time.Time, safeDigester safeDigester) *importFacade {
	return &importFacade{store: store, sourceStore: sourceStore, limits: limits, now: now, safeDigester: safeDigester}
}

func (f *importFacade) Binding() imports.ExtensionImportFacadeBinding {
	return imports.ExtensionImportFacadeBinding{
		SchemaID:                        importFacadeBindingSchemaID,
		TargetKind:                      targetKindNetworkFlowTable,
		ExtensionProfileID:              ProfileID,
		OwnerContractRef:                importOwnerContractRef,
		FacadeID:                        importFacadeID,
		ContractMajor:                   10,
		MappingSchemaID:                 importMappingSchemaID,
		PreviewRequestSchemaID:          importPreviewRequestID,
		SourceInspectionRequestSchemaID: importSourceInspectionRequestID,
		PreviewResultSchemaID:           importPreviewResultSchemaID,
		SourceInspectionResultSchemaID:  importSourceInspectionResultSchemaID,
		ApplyRequestSchemaID:            importApplyRequestID,
		ApplyResultSchemaID:             importApplyResultID,
		ErrorSchemaID:                   networkFlowImportOwnerErrorSchemaID,
		ErrorTranslationID:              networkFlowImportErrorTranslationID,
		CommitProtocolID:                importCommitProtocolID,
	}
}

func (f *importFacade) PrepareImportUnitMapping(ctx context.Context, request imports.ExtensionImportMappingRequest) (imports.ExtensionImportMappingResult, error) {
	if f == nil || f.sourceStore == nil {
		return imports.ExtensionImportMappingResult{}, importOwnerError(
			"network_flow_target_unavailable",
			map[string]any{"reason_code": "owner_apply_contract_unavailable"},
		)
	}
	if request.TargetKind != targetKindNetworkFlowTable || request.ExtensionProfileID != ProfileID || request.OwnerMappingSchemaID != mappingCandidateSchemaID {
		return imports.ExtensionImportMappingResult{}, importOwnerError(
			"network_flow_mapping_invalid",
			map[string]any{"reason_code": "variant_member_conflict"},
		)
	}
	stream, err := f.sourceStore.OpenSourceStream(ctx, request.SourceCapability.SourceStreamRef)
	if err != nil {
		return imports.ExtensionImportMappingResult{}, err
	}
	defer func() { _ = stream.Reader.Close() }()
	parsed, err := parseCSVPreview(stream.Reader, request.SourceCapability.SourceContentSHA256, f.limits)
	if err != nil {
		return imports.ExtensionImportMappingResult{}, facadeError(err)
	}
	mapping, err := materializeApprovedMapping(request.OwnerMapping, parsed.SourceColumns)
	if err != nil {
		return imports.ExtensionImportMappingResult{}, facadeError(err)
	}
	fingerprint, err := mappingFingerprint(mapping, parsed.SourceContentSHA256)
	if err != nil {
		return imports.ExtensionImportMappingResult{}, err
	}
	validation, err := validatePreview(parsed, mapping, fingerprint, f.limits)
	if err != nil {
		return imports.ExtensionImportMappingResult{}, err
	}
	diagnosticResources := make([]map[string]any, 0, len(validation.Diagnostics))
	for _, diagnostic := range validation.Diagnostics {
		diagnosticResources = append(diagnosticResources, diagnosticResource(diagnostic))
	}
	ownerResult := map[string]any{
		"schema_id":                        importPreviewResultSchemaID,
		"source_content_sha256":            parsed.SourceContentSHA256,
		"source_columns":                   mapping.SourceColumns,
		"materialized_mapping":             mapping,
		"mapping_fingerprint":              fingerprint,
		"preview_record_count":             len(parsed.Records),
		"preview_accepted_count":           validation.AcceptedRowCount,
		"preview_rejected_count":           validation.RejectedRowCount,
		"preview_resolved_direction_count": validation.ResolvedDirectionCount,
		"preview_unknown_direction_count":  validation.UnknownDirectionCount,
		"diagnostics":                      diagnosticResources,
		"diagnostics_truncated":            false,
	}
	return imports.ExtensionImportMappingResult{
		SourceColumns:       importSourceColumns(mapping.SourceColumns),
		HeaderRowRef:        1,
		DataStartRowRef:     2,
		OwnerMapping:        marshalApprovedMapping(mapping),
		MappingFingerprint:  fingerprint,
		OwnerResultSchemaID: importPreviewResultSchemaID,
		OwnerResult:         ownerResult,
	}, nil
}

func importSourceColumns(columns []sourceColumnDescriptor) []imports.ExtensionImportSourceColumn {
	result := make([]imports.ExtensionImportSourceColumn, len(columns))
	for i, column := range columns {
		result[i] = imports.ExtensionImportSourceColumn{Ordinal: column.SourceColumnOrdinal, HeaderText: column.RawHeaderText}
	}
	return result
}

func validLowerSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

func (f *importFacade) applyImportUnitTx(
	ctx context.Context,
	tx pgx.Tx,
	request imports.ExtensionImportApplyRequest,
) (imports.ExtensionImportApplyResult, error) {
	if f == nil || f.store == nil || f.sourceStore == nil || tx == nil {
		return imports.ExtensionImportApplyResult{}, importOwnerError(
			"network_flow_target_unavailable",
			map[string]any{"reason_code": "owner_apply_contract_unavailable"},
		)
	}
	if f.safeDigester == nil {
		return imports.ExtensionImportApplyResult{}, importOwnerError(
			"network_flow_target_unavailable",
			map[string]any{"reason_code": "owner_apply_contract_unavailable"},
		)
	}
	if request.TargetKind != targetKindNetworkFlowTable || request.ExtensionProfileID != ProfileID || request.OwnerMappingSchemaID != importMappingSchemaID {
		return imports.ExtensionImportApplyResult{}, importOwnerError(
			"network_flow_mapping_invalid",
			map[string]any{"reason_code": "variant_member_conflict"},
		)
	}
	if request.SourceCapability.SourceContentSHA256 != request.ExpectedSourceContentSHA256 {
		return imports.ExtensionImportApplyResult{}, importOwnerError(
			"network_flow_source_changed",
			nil,
		)
	}
	prepared, err := f.prepareImportApply(ctx, tx, request)
	if err != nil {
		return imports.ExtensionImportApplyResult{}, err
	}
	defer prepared.close()
	table, err := f.store.CreateTableTx(ctx, tx, prepared.params)
	if err != nil {
		return imports.ExtensionImportApplyResult{}, storeApplyError(err)
	}
	return prepared.result(table), nil
}

type preparedImportApply struct {
	params  createTableParams
	request imports.ExtensionImportApplyRequest
	mapping approvedMapping
	source  io.Closer
}

// close releases the protected source. It never ends the borrowed transaction.
func (p preparedImportApply) close() {
	if p.source != nil {
		_ = p.source.Close()
	}
}

func (p preparedImportApply) result(table tableRecord) imports.ExtensionImportApplyResult {
	ownerResponse := map[string]any{
		"schema_id":             importApplyResultID,
		"import_session_id":     p.request.ImportSessionID.String(),
		"import_unit_id":        p.request.ImportUnitID.String(),
		"source_content_sha256": p.request.ExpectedSourceContentSHA256,
		"source_profile_id":     p.mapping.SourceProfileID,
		"parser_profile_id":     p.mapping.ParserProfileID,
		"mapping_fingerprint":   p.request.MappingFingerprint,
		"table_ref":             tableResultRef(p.request.IncidentID.String(), table),
	}
	return imports.ExtensionImportApplyResult{
		ResourceRefs: []jobs.ResourceRef{{
			Kind:  targetKindNetworkFlowTable,
			ID:    table.TableID,
			Route: networkFlowTableRoute(p.request.IncidentID.String(), table.TableID),
		}},
		OwnerResponse: ownerResponse,
	}
}

// prepareImportApply opens the protected source and fixes everything the
// table needs before its single pass (§9.2). The pass runs later, inside
// CreateTableTx, while COPY inserts the rows it accepts; the source digest it
// verifies at end of source equals the expected digest stored here.
func (f *importFacade) prepareImportApply(ctx context.Context, tx pgx.Tx, request imports.ExtensionImportApplyRequest) (preparedImportApply, error) {
	mapping, err := decodeApprovedMapping(request.OwnerMapping, f.limits)
	if err != nil {
		return preparedImportApply{}, facadeError(err)
	}
	stream, err := f.openImportSource(ctx, tx, request)
	if err != nil {
		return preparedImportApply{}, err
	}
	prepared := preparedImportApply{request: request, mapping: mapping, source: stream.Reader}
	filenameDisplay := sanitizeSourceFilenameDisplay(stream.OriginalFilename)
	filenameDigest, filenameDigestKeyID, err := f.safeDigester.Digest("source_filename", filenameDisplay)
	if err != nil {
		prepared.close()
		return preparedImportApply{}, err
	}
	now := f.now()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	prepared.params = createTableParams{
		IncidentID:                request.IncidentID,
		ActorUserID:               request.ActorUserID,
		ImportSessionID:           request.ImportSessionID,
		ImportUnitID:              request.ImportUnitID,
		CommittedSourceID:         imports.CommittedSourceID(request.ImportUnitID),
		SourceRawCellScalarLimit:  f.limits.MaxRawCellScalarLength,
		ClientTxnID:               request.ClientTxnID,
		SourceContentSHA256:       request.ExpectedSourceContentSHA256,
		OriginalFilename:          stream.OriginalFilename,
		SourceFilenameDigest:      filenameDigest,
		SourceFilenameDigestKeyID: filenameDigestKeyID,
		MappingFingerprint:        request.MappingFingerprint,
		SourceProfileID:           mapping.SourceProfileID,
		ParserProfileID:           mapping.ParserProfileID,
		DisplayNameOverride:       mapping.DisplayNameOverride,
		Now:                       now,
		ImportRows: func(gate rowSourceGate, visit func(normalizedRow) error) (importSummary, error) {
			pass, err := f.validateImportSource(ctx, stream.Reader, request, mapping, gate, visit)
			if err != nil {
				return importSummary{}, err
			}
			if pass.Accepted == 0 {
				return importSummary{}, allRowsRejectedOwnerError(pass.Rejected, pass.Sample, pass.DiagnosticCount > int64(len(pass.Sample)))
			}
			return importSummary{
				RowCountAccepted:           pass.Accepted,
				RowCountRejected:           pass.Rejected,
				SourceInterpretationSHA256: pass.ValidationSHA,
				Diagnostics:                pass.Diagnostics,
				DiagnosticsTruncated:       pass.DiagnosticCount > int64(len(pass.Diagnostics)),
			}, nil
		},
	}
	return prepared, nil
}

func sourceColumnsMatch(approved []sourceColumnDescriptor, actual []sourceColumnDescriptor) bool {
	if len(approved) != len(actual) {
		return false
	}
	for index := range approved {
		if approved[index].SourceColumnOrdinal != actual[index].SourceColumnOrdinal ||
			approved[index].RawHeaderText != actual[index].RawHeaderText ||
			approved[index].NormalizedHeaderForSuggestion != actual[index].NormalizedHeaderForSuggestion ||
			approved[index].RawHeaderSHA256 != actual[index].RawHeaderSHA256 {
			return false
		}
	}
	return true
}

func tableResultRef(incidentID string, table tableRecord) map[string]any {
	return map[string]any{
		"kind":                  targetKindNetworkFlowTable,
		"id":                    table.TableID,
		"route":                 networkFlowTableRoute(incidentID, table.TableID),
		"display_name":          table.DisplayName,
		"row_count_accepted":    table.RowCountAccepted,
		"row_count_rejected":    table.RowCountRejected,
		"diagnostics_truncated": table.DiagnosticsTruncated,
		"table_version":         table.TableVersion,
	}
}

func networkFlowTableRoute(incidentID string, tableID string) string {
	return "/api/v1/incidents/" + incidentID + "/network-flow/tables/" + tableID
}

func facadeError(err error) error {
	var invalidName *invalidDisplayNameError
	if errors.As(err, &invalidName) {
		return importDisplayNameError(invalidName)
	}
	if errors.Is(err, errSourceChanged) {
		return importOwnerError("network_flow_source_changed", nil)
	}
	var sourceErr *sourceValidationError
	if errors.As(err, &sourceErr) {
		if sourceErr.Code == "network_flow_no_data_rows" {
			return importOwnerError("network_flow_no_data_rows", nil)
		}
		reasonCode := sourceErr.ReasonCode
		if reasonCode == "" {
			reasonCode = sourceErr.Code
		}
		return importOwnerError(
			"network_flow_mapping_invalid",
			map[string]any{"reason_code": reasonCode},
		)
	}
	var mappingErr *mappingValidationError
	if errors.As(err, &mappingErr) {
		reasonCode := mappingErr.ReasonCode
		if reasonCode == "" {
			reasonCode = mappingErr.Code
		}
		details := map[string]any{"reason_code": reasonCode}
		if mappingErr.FieldKey != "" {
			details["field"] = mappingErr.FieldKey
		}
		return importOwnerError("network_flow_mapping_invalid", details)
	}
	return err
}

func storeApplyError(err error) error {
	switch {
	case errors.Is(err, errTableLimitExceeded):
		return importOwnerError(
			"network_flow_target_unavailable",
			map[string]any{"reason_code": "network_flow_table_limit_exceeded"},
		)
	case errors.Is(err, errTableNameExhausted):
		return importOwnerError(
			"network_flow_target_unavailable",
			map[string]any{"reason_code": "network_flow_table_name_exhausted"},
		)
	case errors.Is(err, errIDGenerationFailed):
		return importOwnerError("network_flow_internal_failure", nil)
	default:
		return facadeError(err)
	}
}
