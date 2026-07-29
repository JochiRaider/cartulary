package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type importFacade struct {
	store        *Store
	sourceStore  ImportSourcePort
	limits       Limits
	now          func() time.Time
	safeDigester SafeDigester
}

const importPreviewResultSchemaID = "cartulary.network_flow.import_preview_result.v1"

func newImportFacade(store *Store, sourceStore ImportSourcePort, limits Limits, now func() time.Time, safeDigester SafeDigester) *importFacade {
	return &importFacade{store: store, sourceStore: sourceStore, limits: limits.normalized(), now: now, safeDigester: safeDigester}
}

func (f *importFacade) PrepareImportUnitMapping(ctx context.Context, request imports.ExtensionImportMappingRequest) (imports.ExtensionImportMappingResult, error) {
	if f == nil || f.sourceStore == nil {
		return imports.ExtensionImportMappingResult{}, applyBlocked("owner_apply_contract_unavailable")
	}
	if request.TargetKind != TargetKindNetworkFlowTable || request.ExtensionProfileID != ProfileID || request.OwnerMappingSchemaID != MappingCandidateSchemaID {
		return imports.ExtensionImportMappingResult{}, applyBlocked("variant_member_conflict")
	}
	stream, err := f.sourceStore.OpenSourceStream(ctx, request.SourceCapability.SourceStreamRef)
	if err != nil {
		return imports.ExtensionImportMappingResult{}, err
	}
	defer func() { _ = stream.Reader.Close() }()
	parsed, err := ParseCSVPreview(stream.Reader, request.SourceCapability.SourceContentSHA256, f.limits)
	if err != nil {
		return imports.ExtensionImportMappingResult{}, facadeError(err)
	}
	mapping, err := MaterializeApprovedMapping(request.OwnerMapping, parsed.SourceColumns)
	if err != nil {
		return imports.ExtensionImportMappingResult{}, facadeError(err)
	}
	fingerprint := MappingFingerprint(mapping, parsed.SourceContentSHA256)
	rows, diagnostics, diagnosticsTruncated, err := ValidateRows(parsed, mapping, fingerprint, f.limits)
	if err != nil {
		return imports.ExtensionImportMappingResult{}, err
	}
	diagnosticResources := make([]map[string]any, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		diagnosticResources = append(diagnosticResources, diagnosticResource(diagnostic))
	}
	ownerResult := map[string]any{
		"schema_id":              importPreviewResultSchemaID,
		"source_content_sha256":  parsed.SourceContentSHA256,
		"source_columns":         mapping.SourceColumns,
		"materialized_mapping":   mapping,
		"mapping_fingerprint":    fingerprint,
		"preview_record_count":   len(parsed.Records),
		"preview_accepted_count": len(rows),
		"preview_rejected_count": len(parsed.Records) - len(rows),
		"diagnostics":            diagnosticResources,
		"diagnostics_truncated":  diagnosticsTruncated,
	}
	return imports.ExtensionImportMappingResult{
		OwnerMapping:        MarshalApprovedMapping(mapping),
		MappingFingerprint:  fingerprint,
		OwnerResultSchemaID: importPreviewResultSchemaID,
		OwnerResult:         ownerResult,
	}, nil
}

func (f *importFacade) ValidateImportUnitMappingResult(result imports.ExtensionImportMappingResult) error {
	if result.OwnerResultSchemaID != importPreviewResultSchemaID || result.OwnerResult == nil {
		return fmt.Errorf("network flow import preview result schema mismatch")
	}
	raw, err := json.Marshal(result.OwnerResult)
	if err != nil {
		return fmt.Errorf("marshal network flow import preview result: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var payload struct {
		SchemaID             string                   `json:"schema_id"`
		SourceContentSHA256  string                   `json:"source_content_sha256"`
		SourceColumns        []SourceColumnDescriptor `json:"source_columns"`
		MaterializedMapping  json.RawMessage          `json:"materialized_mapping"`
		MappingFingerprint   string                   `json:"mapping_fingerprint"`
		PreviewRecordCount   int                      `json:"preview_record_count"`
		PreviewAcceptedCount int                      `json:"preview_accepted_count"`
		PreviewRejectedCount int                      `json:"preview_rejected_count"`
		Diagnostics          []json.RawMessage        `json:"diagnostics"`
		DiagnosticsTruncated bool                     `json:"diagnostics_truncated"`
	}
	if err := decoder.Decode(&payload); err != nil {
		return fmt.Errorf("decode network flow import preview result: %w", err)
	}
	if payload.SchemaID != importPreviewResultSchemaID ||
		!validLowerSHA256(payload.SourceContentSHA256) ||
		!validLowerSHA256(payload.MappingFingerprint) ||
		payload.MappingFingerprint != result.MappingFingerprint ||
		payload.PreviewRecordCount < 0 || payload.PreviewRecordCount > 50 ||
		payload.PreviewAcceptedCount < 0 || payload.PreviewRejectedCount < 0 ||
		payload.PreviewAcceptedCount+payload.PreviewRejectedCount != payload.PreviewRecordCount ||
		payload.DiagnosticsTruncated {
		return fmt.Errorf("network flow import preview result failed semantic validation")
	}
	approved, err := DecodeApprovedMapping(payload.MaterializedMapping)
	if err != nil {
		return fmt.Errorf("validate network flow materialized preview mapping: %w", err)
	}
	if !sourceColumnsMatch(payload.SourceColumns, approved.SourceColumns) || MappingFingerprint(approved, payload.SourceContentSHA256) != payload.MappingFingerprint {
		return fmt.Errorf("network flow import preview result fingerprint mismatch")
	}
	for index, column := range payload.SourceColumns {
		if column.SourceColumnOrdinal != index+1 || !validLowerSHA256(column.RawHeaderSHA256) || column.SampleValues == nil {
			return fmt.Errorf("network flow import preview source column %d is invalid", index+1)
		}
	}
	for _, diagnostic := range payload.Diagnostics {
		if err := validatePreviewDiagnosticShape(diagnostic); err != nil {
			return err
		}
	}
	return nil
}

func validatePreviewDiagnosticShape(raw json.RawMessage) error {
	object, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("network flow import preview diagnostic is not an object")
	}
	required := []string{
		"diagnostic_id", "source_row_number", "source_column_ordinal", "raw_header_sha256", "field_key",
		"error_code", "reason_code", "safe_sample", "raw_value_sha256", "message_key", "message_args", "message",
		"limit_name", "limit_value", "actual_value",
	}
	if len(object) != len(required) {
		return fmt.Errorf("network flow import preview diagnostic shape mismatch")
	}
	for _, member := range required {
		if _, ok := object[member]; !ok {
			return fmt.Errorf("network flow import preview diagnostic missing %s", member)
		}
	}
	return nil
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
		return imports.ExtensionImportApplyResult{}, applyBlocked("owner_apply_contract_unavailable")
	}
	if f.safeDigester == nil {
		return imports.ExtensionImportApplyResult{}, applyBlocked("owner_apply_contract_unavailable")
	}
	if request.TargetKind != TargetKindNetworkFlowTable || request.ExtensionProfileID != ProfileID || request.OwnerMappingSchemaID != MappingCandidateSchemaID {
		return imports.ExtensionImportApplyResult{}, applyBlocked("variant_member_conflict")
	}
	if request.SourceCapability.SourceContentSHA256 != request.ExpectedSourceContentSHA256 {
		return imports.ExtensionImportApplyResult{}, applyBlocked("source_changed")
	}
	prepared, err := f.prepareImportApply(ctx, request)
	if err != nil {
		return imports.ExtensionImportApplyResult{}, err
	}
	importStore, ok := f.sourceStore.(importTransactionPort)
	if !ok {
		return imports.ExtensionImportApplyResult{}, applyBlocked("owner_apply_contract_unavailable")
	}
	capability := &transactionCapability{
		participantID: ImportApplyParticipantID,
		tx:            tx,
		store:         f.store,
		imports:       importStore,
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
	params  CreateTableParams
	request imports.ExtensionImportApplyRequest
	parsed  ParsedCSV
	mapping ApprovedMapping
}

func (p preparedImportApply) result(table TableRecord) imports.ExtensionImportApplyResult {
	ownerResponse := map[string]any{
		"schema_id":             "cartulary.network_flow.import_unit_result.v1",
		"import_session_id":     p.request.ImportSessionID.String(),
		"import_unit_id":        p.request.ImportUnitID.String(),
		"source_content_sha256": p.parsed.SourceContentSHA256,
		"source_profile_id":     p.mapping.SourceProfileID,
		"parser_profile_id":     p.mapping.ParserProfileID,
		"mapping_fingerprint":   p.request.MappingFingerprint,
		"table_ref":             tableResultRef(p.request.IncidentID.String(), table),
	}
	return imports.ExtensionImportApplyResult{
		ResourceRefs: []jobs.ResourceRef{{
			Kind:  TargetKindNetworkFlowTable,
			ID:    table.TableID,
			Route: networkFlowTableRoute(p.request.IncidentID.String(), table.TableID),
		}},
		OwnerResponse: ownerResponse,
	}
}

func (f *importFacade) prepareImportApply(ctx context.Context, request imports.ExtensionImportApplyRequest) (preparedImportApply, error) {
	stream, err := f.sourceStore.OpenSourceStream(ctx, request.SourceCapability.SourceStreamRef)
	if err != nil {
		return preparedImportApply{}, err
	}
	defer func() { _ = stream.Reader.Close() }()
	parsed, err := ParseCSVApply(stream.Reader, request.ExpectedSourceContentSHA256, f.limits)
	if err != nil {
		return preparedImportApply{}, facadeError(err)
	}
	mapping, err := DecodeApprovedMapping(request.OwnerMapping)
	if err != nil {
		return preparedImportApply{}, facadeError(err)
	}
	if !sourceColumnsMatch(mapping.SourceColumns, parsed.SourceColumns) {
		return preparedImportApply{}, applyBlocked("source_changed")
	}
	fingerprint := MappingFingerprint(mapping, parsed.SourceContentSHA256)
	if fingerprint != request.MappingFingerprint {
		return preparedImportApply{}, applyBlocked("source_changed")
	}
	rows, diagnostics, diagnosticsTruncated, err := ValidateRows(parsed, mapping, fingerprint, f.limits)
	if err != nil {
		return preparedImportApply{}, err
	}
	if len(rows) == 0 {
		return preparedImportApply{}, applyBlocked("network_flow_all_rows_rejected")
	}
	if int64(len(rows)) > f.limits.MaxAcceptedRowsPerTable {
		return preparedImportApply{}, applyBlocked("network_flow_resource_limit_exceeded")
	}
	originalFilename, err := f.originalFilename(ctx, request.ImportSessionID)
	if err != nil {
		return preparedImportApply{}, err
	}
	filenameDisplay := SanitizeSourceFilenameDisplay(originalFilename)
	filenameDigest, filenameDigestKeyID, err := f.safeDigester.Digest("source_filename", filenameDisplay)
	if err != nil {
		return preparedImportApply{}, err
	}
	now := f.now()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	params := CreateTableParams{
		IncidentID:                request.IncidentID,
		ActorUserID:               request.ActorUserID,
		ImportSessionID:           request.ImportSessionID,
		ImportUnitID:              request.ImportUnitID,
		ClientTxnID:               request.ClientTxnID,
		SourceContentSHA256:       parsed.SourceContentSHA256,
		OriginalFilename:          originalFilename,
		SourceFilenameDigest:      filenameDigest,
		SourceFilenameDigestKeyID: filenameDigestKeyID,
		MappingFingerprint:        fingerprint,
		SourceProfileID:           mapping.SourceProfileID,
		ParserProfileID:           mapping.ParserProfileID,
		DisplayNameOverride:       mapping.DisplayNameOverride,
		Rows:                      rows,
		Diagnostics:               diagnostics,
		DiagnosticsTruncated:      diagnosticsTruncated,
		Now:                       now,
	}
	return preparedImportApply{params: params, request: request, parsed: parsed, mapping: mapping}, nil
}

func (f *importFacade) originalFilename(ctx context.Context, sessionID uuid.UUID) (string, error) {
	session, _, err := f.sourceStore.GetSession(ctx, sessionID)
	if err != nil {
		return "", err
	}
	value, _ := session["original_filename"].(string)
	return value, nil
}

func sourceColumnsMatch(approved []SourceColumnDescriptor, actual []SourceColumnDescriptor) bool {
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

func tableResultRef(incidentID string, table TableRecord) map[string]any {
	return map[string]any{
		"kind":                  TargetKindNetworkFlowTable,
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
	if errors.Is(err, ErrSourceChanged) {
		return applyBlocked("source_changed")
	}
	var sourceErr *SourceValidationError
	if errors.As(err, &sourceErr) {
		if sourceErr.ReasonCode != "" {
			return applyBlocked(sourceErr.ReasonCode)
		}
		return applyBlocked(sourceErr.Code)
	}
	var mappingErr *MappingValidationError
	if errors.As(err, &mappingErr) {
		if mappingErr.ReasonCode != "" {
			return &imports.ApplyBlockedError{ReasonCode: mappingErr.ReasonCode, Field: mappingErr.FieldKey}
		}
		return &imports.ApplyBlockedError{ReasonCode: mappingErr.Code, Field: mappingErr.FieldKey}
	}
	return err
}

func storeApplyError(err error) error {
	var invalidName *InvalidDisplayNameError
	switch {
	case errors.As(err, &invalidName):
		return applyBlocked(invalidName.ReasonCode)
	case errors.Is(err, ErrTableLimitExceeded):
		return applyBlocked("network_flow_table_limit_exceeded")
	case errors.Is(err, ErrTableNameExhausted):
		return applyBlocked("network_flow_table_name_exhausted")
	case errors.Is(err, ErrIDGenerationFailed):
		return applyBlocked("network_flow_id_generation_failed")
	default:
		return err
	}
}

func applyBlocked(reason string) error {
	return &imports.ApplyBlockedError{ReasonCode: reason}
}
