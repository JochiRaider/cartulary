package networkflow

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type ImportFacade struct {
	store           *Store
	sourceStore     *imports.Store
	limits          Limits
	now             func() time.Time
	safeDigestKeyID string
	safeDigestKey   []byte
}

type ImportFacadeOption func(*ImportFacade)

func WithImportFacadeLimits(limits Limits) ImportFacadeOption {
	return func(f *ImportFacade) {
		f.limits = limits.normalized()
	}
}

func WithImportFacadeClock(now func() time.Time) ImportFacadeOption {
	return func(f *ImportFacade) {
		f.now = now
	}
}

func WithImportFacadeSafeDigest(keyID string, key []byte) ImportFacadeOption {
	return func(f *ImportFacade) {
		f.safeDigestKeyID = keyID
		f.safeDigestKey = append([]byte(nil), key...)
	}
}

func NewImportFacade(store *Store, sourceStore *imports.Store, options ...ImportFacadeOption) *ImportFacade {
	facade := &ImportFacade{
		store:       store,
		sourceStore: sourceStore,
		limits:      DefaultLimits(),
		now:         func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(facade)
	}
	facade.limits = facade.limits.normalized()
	return facade
}

func (f *ImportFacade) PrepareImportUnitMapping(ctx context.Context, request imports.ExtensionImportMappingRequest) (imports.ExtensionImportMappingResult, error) {
	if f == nil || f.sourceStore == nil {
		return imports.ExtensionImportMappingResult{}, applyBlocked("owner_apply_contract_unavailable")
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
	ownerResponse := map[string]any{
		"schema_id":              "cartulary.network_flow.import_preview_result.v1",
		"source_content_sha256":  parsed.SourceContentSHA256,
		"source_columns":         mapping.SourceColumns,
		"materialized_mapping":   mapping,
		"mapping_fingerprint":    fingerprint,
		"preview_record_count":   len(parsed.Records),
		"preview_accepted_count": len(rows),
		"preview_rejected_count": len(parsed.Records) - len(rows),
		"diagnostics":            diagnostics,
		"diagnostics_truncated":  diagnosticsTruncated,
	}
	return imports.ExtensionImportMappingResult{
		OwnerMapping:       MarshalApprovedMapping(mapping),
		MappingFingerprint: fingerprint,
		OwnerResponse:      ownerResponse,
	}, nil
}

func (f *ImportFacade) ApplyImportUnitTx(ctx context.Context, tx pgx.Tx, request imports.ExtensionImportApplyRequest) (imports.ExtensionImportApplyResult, error) {
	if f == nil || f.store == nil || f.sourceStore == nil {
		return imports.ExtensionImportApplyResult{}, applyBlocked("owner_apply_contract_unavailable")
	}
	if request.SourceCapability.SourceContentSHA256 != request.ExpectedSourceContentSHA256 {
		return imports.ExtensionImportApplyResult{}, applyBlocked("source_changed")
	}
	stream, err := f.sourceStore.OpenSourceStream(ctx, request.SourceCapability.SourceStreamRef)
	if err != nil {
		return imports.ExtensionImportApplyResult{}, err
	}
	defer func() { _ = stream.Reader.Close() }()
	parsed, err := ParseCSVApply(stream.Reader, request.ExpectedSourceContentSHA256, f.limits)
	if err != nil {
		return imports.ExtensionImportApplyResult{}, facadeError(err)
	}
	mapping, err := DecodeApprovedMapping(request.OwnerMapping)
	if err != nil {
		return imports.ExtensionImportApplyResult{}, facadeError(err)
	}
	if !sourceColumnsMatch(mapping.SourceColumns, parsed.SourceColumns) {
		return imports.ExtensionImportApplyResult{}, applyBlocked("source_changed")
	}
	fingerprint := MappingFingerprint(mapping, parsed.SourceContentSHA256)
	if fingerprint != request.MappingFingerprint {
		return imports.ExtensionImportApplyResult{}, applyBlocked("source_changed")
	}
	rows, diagnostics, diagnosticsTruncated, err := ValidateRows(parsed, mapping, fingerprint, f.limits)
	if err != nil {
		return imports.ExtensionImportApplyResult{}, err
	}
	if len(rows) == 0 {
		return imports.ExtensionImportApplyResult{}, applyBlocked("network_flow_all_rows_rejected")
	}
	if int64(len(rows)) > f.limits.MaxAcceptedRowsPerTable {
		return imports.ExtensionImportApplyResult{}, applyBlocked("network_flow_resource_limit_exceeded")
	}
	originalFilename, err := f.originalFilename(ctx, request.ImportSessionID)
	if err != nil {
		return imports.ExtensionImportApplyResult{}, err
	}
	filenameDisplay := SanitizeSourceFilenameDisplay(originalFilename)
	filenameDigest, filenameDigestKeyID := SafeDigest(f.safeDigestKeyID, f.safeDigestKey, "source_filename", filenameDisplay)
	now := f.now()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	table, err := f.store.CreateTableTx(ctx, tx, CreateTableParams{
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
	})
	if err != nil {
		return imports.ExtensionImportApplyResult{}, storeApplyError(err)
	}
	ownerResponse := map[string]any{
		"schema_id":             "cartulary.network_flow.import_unit_result.v1",
		"import_session_id":     request.ImportSessionID.String(),
		"import_unit_id":        request.ImportUnitID.String(),
		"source_content_sha256": parsed.SourceContentSHA256,
		"source_profile_id":     mapping.SourceProfileID,
		"parser_profile_id":     mapping.ParserProfileID,
		"mapping_fingerprint":   fingerprint,
		"table_ref":             tableResultRef(request.IncidentID.String(), table),
	}
	return imports.ExtensionImportApplyResult{
		ResourceRefs: []jobs.ResourceRef{{
			Kind:  TargetKindNetworkFlowTable,
			ID:    table.TableID,
			Route: networkFlowTableRoute(request.IncidentID.String(), table.TableID),
		}},
		OwnerResponse: ownerResponse,
	}, nil
}

func (f *ImportFacade) originalFilename(ctx context.Context, sessionID uuid.UUID) (string, error) {
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
			return applyBlocked(mappingErr.ReasonCode)
		}
		return applyBlocked(mappingErr.Code)
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
