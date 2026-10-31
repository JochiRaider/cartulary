package networkflow

import (
	"context"
	"errors"
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
	importOwnerContractRef               = "network_flow_activity@8"
	importMappingSchemaID                = "cartulary.network_flow.approved_mapping.v1"
	importPreviewRequestID               = "cartulary.network_flow.import_preview_request.v1"
	importPreviewResultSchemaID          = "cartulary.network_flow.import_preview_result.v1"
	importSourceInspectionResultSchemaID = "cartulary.network_flow.import_source_inspection_result.v1"
	importSourceInspectionRequestID      = "cartulary.network_flow.import_source_inspection_request.v1"
	importApplyRequestID                 = "cartulary.network_flow.import_apply_request.v1"
	importApplyResultID                  = "cartulary.network_flow.import_unit_result.v1"
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
		ContractMajor:                   8,
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
		"schema_id":              importPreviewResultSchemaID,
		"source_content_sha256":  parsed.SourceContentSHA256,
		"source_columns":         mapping.SourceColumns,
		"materialized_mapping":   mapping,
		"mapping_fingerprint":    fingerprint,
		"preview_record_count":   len(parsed.Records),
		"preview_accepted_count": validation.AcceptedRowCount,
		"preview_rejected_count": validation.RejectedRowCount,
		"diagnostics":            diagnosticResources,
		"diagnostics_truncated":  false,
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

func (f *importFacade) ApplyImportUnitTx(
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
	if request.TargetKind != targetKindNetworkFlowTable || request.ExtensionProfileID != ProfileID || request.OwnerMappingSchemaID != mappingCandidateSchemaID {
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
	capability := &transactionCapability{
		participantID: ImportApplyParticipantID,
		tx:            tx,
		store:         f.store,
		imports:       f.sourceStore,
	}
	if err := capability.ValidateImportApply(ctx, request); err != nil {
		return imports.ExtensionImportApplyResult{}, err
	}
	table, err := capability.CreateImportedTable(ctx, prepared.params)
	if err != nil {
		return imports.ExtensionImportApplyResult{}, storeApplyError(err)
	}
	return prepared.result(table), nil
}

type preparedImportApply struct {
	params    createTableParams
	request   imports.ExtensionImportApplyRequest
	sourceSHA string
	mapping   approvedMapping
}

func (p preparedImportApply) result(table tableRecord) imports.ExtensionImportApplyResult {
	ownerResponse := map[string]any{
		"schema_id":             importApplyResultID,
		"import_session_id":     p.request.ImportSessionID.String(),
		"import_unit_id":        p.request.ImportUnitID.String(),
		"source_content_sha256": p.sourceSHA,
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

func (f *importFacade) prepareImportApply(ctx context.Context, tx pgx.Tx, request imports.ExtensionImportApplyRequest) (preparedImportApply, error) {
	mapping, err := decodeApprovedMapping(request.OwnerMapping, f.limits)
	if err != nil {
		return preparedImportApply{}, facadeError(err)
	}
	first, err := f.validateImportPass(ctx, tx, request, mapping, true, nil)
	if err != nil {
		return preparedImportApply{}, err
	}
	if first.Accepted == 0 {
		return preparedImportApply{}, allRowsRejectedOwnerError(first.Rejected, first.Sample, first.DiagnosticCount > int64(len(first.Sample)))
	}
	fingerprint := request.MappingFingerprint
	originalFilename := first.OriginalFilename
	filenameDisplay := sanitizeSourceFilenameDisplay(originalFilename)
	filenameDigest, filenameDigestKeyID, err := f.safeDigester.Digest("source_filename", filenameDisplay)
	if err != nil {
		return preparedImportApply{}, err
	}
	now := f.now()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	params := createTableParams{
		IncidentID:                request.IncidentID,
		ActorUserID:               request.ActorUserID,
		ImportSessionID:           request.ImportSessionID,
		ImportUnitID:              request.ImportUnitID,
		ClientTxnID:               request.ClientTxnID,
		SourceContentSHA256:       first.SourceSHA,
		OriginalFilename:          originalFilename,
		SourceFilenameDigest:      filenameDigest,
		SourceFilenameDigestKeyID: filenameDigestKeyID,
		MappingFingerprint:        fingerprint,
		SourceProfileID:           mapping.SourceProfileID,
		ParserProfileID:           mapping.ParserProfileID,
		DisplayNameOverride:       mapping.DisplayNameOverride,
		RowCountAccepted:          first.Accepted,
		RowCountRejected:          first.Rejected,
		Diagnostics:               first.Diagnostics,
		DiagnosticsTruncated:      first.DiagnosticCount > int64(len(first.Diagnostics)),
		Now:                       now,
	}
	params.VisitRows = func(visit func(normalizedRow) error) error {
		second, err := f.validateImportPass(ctx, tx, request, mapping, false, visit)
		if err != nil {
			return err
		}
		if !first.agrees(second) {
			return importOwnerError("network_flow_source_changed", nil)
		}
		return nil
	}
	return preparedImportApply{params: params, request: request, sourceSHA: first.SourceSHA, mapping: mapping}, nil
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
