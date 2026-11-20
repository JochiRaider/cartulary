package imports

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
)

func TestMappingFingerprintUsesCanonicalWriteNullToken(t *testing.T) {
	t.Parallel()

	discovered := []map[string]any{{
		"source_column_ordinal": 1,
		"source_header_text":    "Location",
	}}
	body := func(clientTxnID string, policy string) string {
		return `{
			"client_txn_id":` + quotedJSON(clientTxnID) + `,
			"target_view_schema_id":"cartulary.view.hosts.v1",
			"header_row_ref":1,
			"data_start_row_ref":2,
			"unknown_column_policy":"reject_if_unmapped",
			"source_columns":[{
				"source_column_ordinal":1,
				"source_header_text":"Location",
				"field_key":"host.location",
				"entity_binding_mode":"entity_origin",
				"transform_id":null,
				"transform_options":{},
				"empty_value_policy":` + quotedJSON(policy) + `
			}]
		}`
	}
	first, apiErr := decodeMappingRequest(
		strings.NewReader(body("txn-write-null-a", "write_null")),
		discovered,
	)
	if apiErr != nil {
		t.Fatalf("decode first write_null mapping: %#v", apiErr)
	}
	replay, apiErr := decodeMappingRequest(
		strings.NewReader(body("txn-write-null-b", "write_null")),
		discovered,
	)
	if apiErr != nil {
		t.Fatalf("decode replay write_null mapping: %#v", apiErr)
	}
	if first.Fingerprint == "" || first.Fingerprint != replay.Fingerprint {
		t.Fatalf(
			"write_null fingerprint first=%q replay=%q",
			first.Fingerprint,
			replay.Fingerprint,
		)
	}
	if first.approvedMapping.SourceColumns[0].EmptyValuePolicy != "write_null" ||
		strings.Contains(string(first.Normalized), "use_null") {
		t.Fatalf("non-canonical mapping = %#v", first)
	}
	if _, apiErr := decodeMappingRequest(
		strings.NewReader(body("txn-use-null", "use_null")),
		discovered,
	); apiErr == nil ||
		apiErr.Code != "invalid_import_request" ||
		apiErr.Details["reason_code"] != "invalid_empty_value_policy" {
		t.Fatalf("legacy use_null rejection = %#v", apiErr)
	}
}

func quotedJSON(value string) string {
	return `"` + value + `"`
}

type ownerErrorTestFacade struct {
	translation ExtensionImportErrorTranslation
	translated  bool
	validateErr error
}

func (f ownerErrorTestFacade) Binding() ExtensionImportFacadeBinding {
	return ExtensionImportFacadeBinding{}
}

func (f ownerErrorTestFacade) PrepareImportUnitMapping(
	context.Context,
	ExtensionImportMappingRequest,
) (ExtensionImportMappingResult, error) {
	return ExtensionImportMappingResult{}, nil
}

func (f ownerErrorTestFacade) ValidateImportUnitMappingResult(
	ExtensionImportMappingResult,
) error {
	return nil
}

func (f ownerErrorTestFacade) ImportTransaction(ExtensionImportApplyRequest) (crossownertransaction.Participant, error) {
	return nil, nil
}

func (f ownerErrorTestFacade) TranslateImportUnitError(
	error,
) (ExtensionImportErrorTranslation, bool) {
	return f.translation, f.translated
}

func (f ownerErrorTestFacade) ValidateImportUnitError(ExtensionImportOwnerError) error {
	return f.validateErr
}

func TestOwnerErrorTranslationFailsClosedAndPreservesOnlyRegisteredSafeDetail(t *testing.T) {
	t.Parallel()
	assertCrossOwnerOutcomesUseCoreUnitOutcome(t)

	target := importTarget{
		ErrorSchemaID:      "owner.error.v1",
		ErrorTranslationID: "owner.translation.v1",
	}
	registered := ExtensionImportErrorTranslation{
		ErrorSchemaID:      target.ErrorSchemaID,
		ErrorTranslationID: target.ErrorTranslationID,
		CoreReasonCode:     "owner_apply_validation_failed",
		OwnerError: ExtensionImportOwnerError{
			SchemaID:    target.ErrorSchemaID,
			OwnerCode:   "known_owner_validation",
			Retryable:   false,
			SafeDetails: map[string]any{"reason_code": "known_reason", "field": "safe_field"},
		},
	}
	failure := translateExtensionOwnerFailure(
		target,
		ownerErrorTestFacade{translation: registered, translated: true},
		crossOwnerUnitTransaction{},
		errors.New("raw owner stack and source value"),
	)
	wantOwner := map[string]any{
		"schema_id":  "owner.error.v1",
		"owner_code": "known_owner_validation",
		"retryable":  false,
		"safe_details": map[string]any{
			"reason_code": "known_reason",
			"field":       "safe_field",
		},
	}
	if failure.ErrorCode != "import_apply_blocked" ||
		failure.ReasonCode != "owner_apply_validation_failed" ||
		failure.Retryable ||
		!reflect.DeepEqual(failure.Details["owner_error"], wantOwner) {
		t.Fatalf("registered owner translation = %#v", failure)
	}

	for name, facade := range map[string]ownerErrorTestFacade{
		"unknown error": {},
		"schema mismatch": {
			translation: func() ExtensionImportErrorTranslation {
				value := registered
				value.OwnerError.SchemaID = "attacker.supplied.token"
				return value
			}(),
			translated: true,
		},
		"invalid detail": {
			translation: registered,
			translated:  true,
			validateErr: errors.New("invalid safe detail"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := translateExtensionOwnerFailure(
				target,
				facade,
				crossOwnerUnitTransaction{},
				errors.New("secret=/tmp/source.csv token=unknown_owner_token"),
			)
			if got.ErrorCode != "import_apply_blocked" ||
				got.ReasonCode != "owner_apply_validation_failed" ||
				got.Retryable ||
				len(got.Details) != 1 ||
				got.Details["reason_code"] != "owner_apply_validation_failed" {
				t.Fatalf("closed fallback = %#v", got)
			}
		})
	}
}

// A cross-owner transaction that ends with its commit proven absent by
// timeout or database conflict has the Core unit outcome of REQ-01-620c. The
// owner translation and its fallback are not consulted, even when the owner
// would claim the error.
func assertCrossOwnerOutcomesUseCoreUnitOutcome(t *testing.T) {
	t.Helper()
	target := importTarget{ErrorSchemaID: "owner.error.v1", ErrorTranslationID: "owner.translation.v1"}
	claiming := ownerErrorTestFacade{translated: true, translation: ExtensionImportErrorTranslation{
		ErrorSchemaID: target.ErrorSchemaID, ErrorTranslationID: target.ErrorTranslationID,
		CoreReasonCode: "owner_apply_validation_failed",
		OwnerError:     ExtensionImportOwnerError{SchemaID: target.ErrorSchemaID, OwnerCode: "owner_claims_everything", SafeDetails: map[string]any{}},
	}}
	transaction := crossOwnerUnitTransaction{OperationID: "import-unit:operation", Timeout: 30 * time.Second}
	for _, tc := range []struct {
		name string
		err  error
		want importUnitFailureDetail
	}{
		{
			name: "timeout",
			err:  crossownertransaction.ErrTimeout,
			want: importUnitFailureDetail{ErrorCode: "service_unavailable", ReasonCode: "extension_transaction_timeout", Retryable: true,
				Details: map[string]any{"reason_code": "extension_transaction_timeout", "operation_id": "import-unit:operation", "timeout_seconds": int64(30)}},
		},
		{
			name: "wrapped timeout",
			err:  fmt.Errorf("unit apply: %w", crossownertransaction.ErrTimeout),
			want: importUnitFailureDetail{ErrorCode: "service_unavailable", ReasonCode: "extension_transaction_timeout", Retryable: true,
				Details: map[string]any{"reason_code": "extension_transaction_timeout", "operation_id": "import-unit:operation", "timeout_seconds": int64(30)}},
		},
		{
			name: "deadlock",
			err:  &crossownertransaction.ConflictError{ReasonCode: "deadlock_detected", Cause: errors.New("SQLSTATE 40P01")},
			want: importUnitFailureDetail{ErrorCode: "transaction_conflict", ReasonCode: "deadlock_detected", Retryable: true,
				Details: map[string]any{"reason_code": "deadlock_detected"}},
		},
		{
			name: "serialization failure",
			err:  &crossownertransaction.ConflictError{ReasonCode: "serialization_failure", Cause: errors.New("SQLSTATE 40001")},
			want: importUnitFailureDetail{ErrorCode: "transaction_conflict", ReasonCode: "serialization_failure", Retryable: true,
				Details: map[string]any{"reason_code": "serialization_failure"}},
		},
	} {
		t.Run("cross_owner_"+tc.name, func(t *testing.T) {
			t.Parallel()
			for name, facade := range map[string]ExtensionImportFacade{"claiming owner": claiming, "declining owner": ownerErrorTestFacade{}, "no owner": nil} {
				if got := translateExtensionOwnerFailure(target, facade, transaction, tc.err); !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("%s: unit outcome = %#v, want %#v", name, got, tc.want)
				}
			}
		})
	}
}

func (ownerErrorTestFacade) InspectImportUnitSource(context.Context, ExtensionImportSourceRequest) (ExtensionImportSourceResult, error) {
	return ExtensionImportSourceResult{}, nil
}
func (ownerErrorTestFacade) ValidateImportUnitSourceResult(ExtensionImportSourceResult) error {
	return nil
}
