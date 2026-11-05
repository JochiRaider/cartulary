package networkflow

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/google/uuid"
)

func validateOrientationAdmission(ctx context.Context, reader extensionstore.Querier) error {
	if err := imports.ValidateRetainedAnalyticalState(ctx, reader, ProfileID, func(value imports.RetainedAnalyticalPayload) error {
		switch value.Kind {
		case "mapping":
			if value.SchemaID != importMappingSchemaID {
				return errSavedGraphCutoverIncompatible
			}
			_, err := decodeApprovedMapping(value.Payload, defaultEffectiveLimits())
			return err
		case "applied_result":
			if value.SchemaID != importApplyResultID {
				return errSavedGraphCutoverIncompatible
			}
			return validateRetainedImportResult(value.Payload)
		default:
			return errSavedGraphCutoverIncompatible
		}
	}); err != nil {
		return err
	}
	after := uuid.Nil
	for {
		records, next, err := authn.ReadRouteIdempotencyPage(ctx, reader, []string{routeKeyTablesPatch, routeKeyTablesDelete}, after)
		if err != nil {
			return err
		}
		for _, record := range records {
			payload, err := decodeStoredNetworkFlowResponse(record.ResponseJSON)
			if err != nil {
				return errSavedGraphCutoverIncompatible
			}
			if len(record.RequestHash) != 32 || validateTableReceipt(savedGraphReceiptKey(record), record.StatusCode, payload) != nil {
				return errSavedGraphCutoverIncompatible
			}
		}
		if len(records) < 128 {
			return nil
		}
		after = next
	}
}
