package hostidentity

import (
	"context"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OptionalConflictValue struct {
	Present bool
	Value   any
}
type SameFieldConflict struct {
	ConflictToken                     string
	RecordID                          uuid.UUID
	FieldKey, ConflictResolutionClass string
	BaseRowVersion, CurrentRowVersion int64
	ClientValue, ServerValue          any
	BaseValue                         OptionalConflictValue
	ServerUpdatedBy                   uuid.UUID
	ServerUpdatedAt                   time.Time
	SuggestedMergedValue              OptionalConflictValue
}
type SameFieldConflictError struct{ Conflict SameFieldConflict }

func (*SameFieldConflictError) Error() string { return "entities: same field conflict" }

func newEntityConflictSnapshotProjectors() (map[string]conflicts.RevisionSnapshotProjector, error) {
	result := make(map[string]conflicts.RevisionSnapshotProjector, 2)
	for _, entry := range []struct {
		view, schema string
		fields       map[string]string
	}{
		{entitycontract.HostsViewSchemaID, "cartulary.revisions.snapshot.host.v1", map[string]string{
			"host.display_name": "display_name", "host.hostname": "hostname", "host.location": "location", "host.os_platform": "os_platform", "host.business_owner": "business_owner", "host.criticality": "criticality", "host.containment_status": "containment_status",
		}},
		{entitycontract.IdentitiesViewSchemaID, "cartulary.revisions.snapshot.identity.v1", map[string]string{
			"identity.display_name": "display_name", "identity.upn": "upn", "identity.email": "email", "identity.sam_account_name": "sam_account_name", "identity.privilege_level": "privilege_level", "identity.mfa_state": "mfa_state", "identity.reset_status": "reset_status",
		}},
	} {
		projector, err := conflicts.NewRevisionSnapshotProjector(entry.schema, entry.fields)
		if err != nil {
			return nil, err
		}
		result[entry.view] = projector
	}
	return result, nil
}

func (s *Store) admitPatchRebaseTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, meta entityRecordMeta, request PatchRequest, requestHash []byte) error {
	rows, err := s.ports.revisions.LoadRevisionWindowTx(ctx, tx, id, request.BaseRowVersion, meta.RowVersion)
	if err != nil {
		return err
	}
	fields, err := s.conflictFields.ResolveViewSchema(request.ViewSchemaID)
	if err != nil {
		return err
	}
	window, err := conflicts.BuildCanonicalPatchConflictWindow(id, request.BaseRowVersion, meta.RowVersion, rows, fields, s.conflictSnapshots[request.ViewSchemaID])
	if err != nil {
		return err
	}
	for _, change := range request.Changes {
		changed, overlap := window.ChangedFields[change.FieldKey]
		if !overlap {
			continue
		}
		var current map[string]any
		switch meta.RecordType {
		case "host":
			record, err := loadHostByRecordIDTx(ctx, tx, id)
			if err != nil {
				return err
			}
			if err := hydrateHostRecordTx(ctx, tx, &record); err != nil {
				return err
			}
			current = buildHostRow(record)
		case "identity":
			record, err := loadIdentityByRecordIDTx(ctx, tx, id)
			if err != nil {
				return err
			}
			if err := hydrateIdentityRecordTx(ctx, tx, &record); err != nil {
				return err
			}
			current = buildIdentityRow(record)
		default:
			return fmt.Errorf("unsupported entity conflict source")
		}
		base, baseOK := entityConflictCell(window.BaseRow, change.FieldKey)
		saved, savedOK := entityConflictCell(current, change.FieldKey)
		if !baseOK || !savedOK {
			return &conflicts.RevisionWindowError{RecordID: id, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
		}
		class := fields.ConflictResolutionClass(change.FieldKey)
		if class == "" {
			class = "atomic_replace"
		}
		token, err := s.conflictTokens.Issue(conflicts.ConflictTokenClaims{RouteKey: "workbook.records.conflicts.resolve", RecordID: id.String(), ViewSchemaID: request.ViewSchemaID, FieldKey: change.FieldKey, ConflictResolutionClass: class, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion, RequestHash: conflicts.RequestHashTokenValue(requestHash)})
		if err != nil {
			return err
		}
		value := canonicalPatchValue(change.Value)
		if len(change.CollectionActions) > 0 {
			collection, ok := canonicalEntityConflictCollection(base)
			if !ok {
				return &conflicts.RevisionWindowError{RecordID: id, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
			}
			base = collection
			value, err = previewEntityCollectionTx(ctx, tx, meta.IncidentID, id, change, collection, requestHash)
			if err != nil {
				return err
			}
		}
		result := SameFieldConflict{ConflictToken: token, RecordID: id, FieldKey: change.FieldKey, ConflictResolutionClass: class, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion, ClientValue: value, ServerValue: saved, BaseValue: OptionalConflictValue{Present: true, Value: base}, ServerUpdatedBy: changed.ServerUpdatedBy, ServerUpdatedAt: changed.ServerUpdatedAt.UTC()}
		if class == "text_compare_merge" {
			if merged, ok := conflicts.SuggestedTextMergeValue(base, saved, value); ok {
				result.SuggestedMergedValue = OptionalConflictValue{Present: true, Value: merged}
			}
		}
		return &SameFieldConflictError{Conflict: result}
	}
	return nil
}

func entityConflictCell(row map[string]any, key string) (any, bool) {
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return nil, false
	}
	cell, ok := cells[key].(map[string]any)
	if !ok {
		return nil, false
	}
	value, present := cell["value"]
	return value, present
}
