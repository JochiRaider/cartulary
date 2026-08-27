package timeline

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

var _ ownerfacade.ImportOwnerCreateTx = (*Facade)(nil)

func NewImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	owner ownerfacade.ImportOwnerCreateTx,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != TimelineViewSchemaID {
		return nil, fmt.Errorf("timeline import surface %q not mapped", targetViewSchemaID)
	}
	facadeOwner, isFacade := owner.(*Facade)
	if owner == nil || (isFacade && (facadeOwner == nil || facadeOwner.store == nil)) {
		return nil, fmt.Errorf("timeline import facade requires an owner")
	}
	return ownerfacade.NewImportOwnerCreateFacadeWithNormalizer(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: targetViewSchemaID,
			FacadeID:           facadeID,
		},
		normalizeImportField,
		owner.CreateImportRowTx,
	)
}

func normalizeImportField(
	fieldKey string,
	raw string,
	emptyValuePolicy string,
) (ownerfacade.ImportScalarValue, bool, error) {
	if raw == "" {
		switch emptyValuePolicy {
		case "omit_field":
			return ownerfacade.ImportScalarValue{}, false, nil
		case "write_null":
			if !mutationpolicy.IsDirectWritableField(fieldKey) {
				return ownerfacade.ImportScalarValue{}, false, fmt.Errorf(
					"timeline import field %s is not nullable",
					fieldKey,
				)
			}
			return ownerfacade.NewNullImportScalar(), true, nil
		default:
			return ownerfacade.ImportScalarValue{}, false, fmt.Errorf(
				"unsupported Timeline empty value policy %q",
				emptyValuePolicy,
			)
		}
	}
	if mutationpolicy.IsDirectWritableField(fieldKey) {
		if !mutationpolicy.IsValidVisibleText(raw) {
			return ownerfacade.ImportScalarValue{}, false, fmt.Errorf(
				"invalid imported Timeline field %s",
				fieldKey,
			)
		}
		return ownerfacade.NewTextImportScalar(raw), true, nil
	}
	switch fieldKey {
	case "timeline.host_refs", "timeline.identity_refs":
		normalized, ok := fieldnorm.NormalizeMentionToken(raw)
		if !ok {
			return ownerfacade.ImportScalarValue{}, false, fmt.Errorf(
				"invalid imported Timeline field %s",
				fieldKey,
			)
		}
		return importCollectionToken(raw, normalized), true, nil
	case "timeline.tags":
		label, normalized, ok := fieldnorm.NormalizeTagLabel(raw)
		if !ok {
			return ownerfacade.ImportScalarValue{}, false, fmt.Errorf(
				"invalid imported Timeline field %s",
				fieldKey,
			)
		}
		return importCollectionToken(label, normalized), true, nil
	default:
		return ownerfacade.ImportScalarValue{}, false, fmt.Errorf(
			"unsupported imported Timeline field %s",
			fieldKey,
		)
	}
}

func importCollectionToken(
	raw string,
	normalized string,
) ownerfacade.ImportScalarValue {
	return ownerfacade.NewCollectionTokenImportScalar(
		ownerfacade.ImportCollectionToken{
			RawText:        raw,
			NormalizedText: normalized,
		},
	)
}

func (f *Facade) CreateImportRowTx(
	ctx context.Context,
	tx pgx.Tx,
	command ownerfacade.ImportOwnerCreateCommand,
) (ownerfacade.ImportOwnerCreateResponse, error) {
	if f == nil || f.store == nil {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf(
			"timeline import facade is not configured",
		)
	}
	request, err := timelineCreateRequestFromImport(command.Request)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	result, mutationSequence, err := f.store.createRowTx(
		ctx,
		tx,
		command.Request.ActorUserID,
		command.Request.IncidentID,
		request,
		command.ChangeSetID,
		command.AllocateMutationSequence,
		command.Now.UTC(),
		createRowOptions{},
	)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.ImportOwnerCreateResponse{
		RecordID:   result.RecordID,
		RowVersion: result.RowVersion,
		ChangeSetMutationRef: fmt.Sprintf(
			"change_set_mutation:%s:%d",
			command.ChangeSetID,
			mutationSequence,
		),
		CreatedOrReused: "created",
		OwnerResultCode: "created",
		RowRefresh:      buildRow(result.Row),
	}, nil
}

func timelineCreateRequestFromImport(
	request ownerfacade.ImportOwnerCreateRequest,
) (CreateRequest, error) {
	if request.TargetViewSchemaID != TimelineViewSchemaID {
		return CreateRequest{}, fmt.Errorf(
			"timeline import surface %q not mapped",
			request.TargetViewSchemaID,
		)
	}
	result := CreateRequest{ClientTxnID: request.ClientTxnID}
	for _, field := range request.FieldValues {
		value := field.NormalizedValue
		switch field.FieldKey {
		case "timeline.host_refs":
			if err := appendTimelineImportToken(&result.HostRefs, value, "add_token"); err != nil {
				return CreateRequest{}, err
			}
		case "timeline.identity_refs":
			if err := appendTimelineImportToken(&result.IdentityRefs, value, "add_token"); err != nil {
				return CreateRequest{}, err
			}
		case "timeline.tags":
			if err := appendTimelineImportToken(&result.Tags, value, "add_tag"); err != nil {
				return CreateRequest{}, err
			}
		default:
			if err := setTimelineImportScalar(&result, field.FieldKey, value); err != nil {
				return CreateRequest{}, err
			}
		}
	}
	result.RawCaptureColumns = timelineRawCaptureColumns(request)
	return result, nil
}

func appendTimelineImportToken(
	payload **CollectionActionPayload,
	value ownerfacade.ImportScalarValue,
	operation string,
) error {
	token, ok := value.CollectionToken()
	if !ok {
		return fmt.Errorf("invalid imported Timeline collection value")
	}
	if *payload == nil {
		*payload = &CollectionActionPayload{Actions: []CollectionAction{}}
	}
	(*payload).Actions = append((*payload).Actions, CollectionAction{
		Op:             operation,
		RawText:        token.RawText,
		NormalizedText: token.NormalizedText,
	})
	return nil
}

func setTimelineImportScalar(
	request *CreateRequest,
	fieldKey string,
	value ownerfacade.ImportScalarValue,
) error {
	var text *string
	switch value.Kind() {
	case ownerfacade.ImportScalarNull:
	case ownerfacade.ImportScalarText:
		scalar, ok := value.Text()
		if !ok {
			return fmt.Errorf("invalid imported Timeline field %s", fieldKey)
		}
		text = &scalar
	default:
		return fmt.Errorf("invalid imported Timeline field %s", fieldKey)
	}
	switch fieldKey {
	case "timeline.date_entered_text":
		request.DateEnteredText = text
	case "timeline.analyst_text":
		request.AnalystText = text
	case "timeline.mitre_stage_text":
		request.MitreStageText = text
	case "timeline.device_object_text":
		request.DeviceObjectText = text
	case "timeline.ip_address_text":
		request.IPAddressText = text
	case "timeline.activity_utc_text":
		request.ActivityUTCText = text
	case "timeline.activity_local_text":
		request.ActivityLocalText = text
	case "timeline.raw_activity_text":
		request.RawActivityText = text
	case "timeline.activity_synopsis_text":
		request.ActivitySynopsisText = text
	case "timeline.data_source_text":
		request.DataSourceText = text
	default:
		return fmt.Errorf("unsupported imported Timeline field %s", fieldKey)
	}
	return nil
}

func timelineRawCaptureColumns(
	request ownerfacade.ImportOwnerCreateRequest,
) []ClipboardRawImportColumn {
	if len(request.UnknownValues) == 0 {
		return nil
	}
	columns := make([]ClipboardRawImportColumn, 0, len(request.UnknownValues))
	for _, unknown := range request.UnknownValues {
		columns = append(columns, ClipboardRawImportColumn{
			SourceKind:          "file_import",
			ImportSessionID:     request.ImportSessionID.String(),
			ImportUnitID:        request.ImportUnitID.String(),
			MappingFingerprint:  request.MappingFingerprint,
			SourceFileKind:      request.SourceFileKind,
			SourceContentSHA256: request.SourceContentSHA256,
			ParserProfileID:     request.ParserProfileID,
			ParserVersion:       request.ParserVersion,
			LocatorKind:         request.LocatorKind,
			Locator:             request.Locator,
			SourceRectA1:        request.SourceRectA1,
			SourceRowOrdinal:    request.SourceRowRef,
			SourceColumnOrdinal: unknown.SourceColumnOrdinal,
			SourceHeaderText:    unknown.SourceHeaderText,
			RawValue:            unknown.RawValue,
			CellKind:            unknown.CellKind,
		})
	}
	return columns
}
