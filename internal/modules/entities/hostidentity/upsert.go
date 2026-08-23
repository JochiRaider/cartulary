package hostidentity

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *mutationCore) upsertHostTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, now time.Time) (HostRecord, map[string]any, string, int, *revisions.RecordSnapshot, error) {
	input, err := hostInputFromCreateRequest(request)
	if err != nil {
		return HostRecord{}, nil, "", 0, nil, err
	}
	current, matched, err := matchHostTx(ctx, tx, incidentID, input)
	if err != nil {
		return HostRecord{}, nil, "", 0, nil, err
	}
	var beforeSnapshot *revisions.RecordSnapshot
	if matched {
		snapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
		if err != nil {
			return HostRecord{}, nil, "", 0, nil, err
		}
		beforeSnapshot = &snapshot
	}
	record, before, operation, status, err := s.applyHostUpsertTx(ctx, tx, actor, incidentID, input, current, matched, now)
	return record, before, operation, status, beforeSnapshot, err
}

func (s *mutationCore) upsertIdentityTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, now time.Time) (IdentityRecord, map[string]any, string, int, *revisions.RecordSnapshot, error) {
	input, err := identityInputFromCreateRequest(request)
	if err != nil {
		return IdentityRecord{}, nil, "", 0, nil, err
	}
	current, matched, err := matchIdentityTx(ctx, tx, incidentID, input)
	if err != nil {
		return IdentityRecord{}, nil, "", 0, nil, err
	}
	var beforeSnapshot *revisions.RecordSnapshot
	if matched {
		snapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
		if err != nil {
			return IdentityRecord{}, nil, "", 0, nil, err
		}
		beforeSnapshot = &snapshot
	}
	record, before, operation, status, err := s.applyIdentityUpsertTx(ctx, tx, actor, incidentID, input, current, matched, now)
	return record, before, operation, status, beforeSnapshot, err
}

func (s *mutationCore) applyHostUpsertTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, input hostUpsertInput, current HostRecord, matched bool, now time.Time) (HostRecord, map[string]any, string, int, error) {
	if !matched {
		record := HostRecord{
			IncidentID:    incidentID,
			DisplayName:   input.DisplayName,
			AADDeviceID:   cloneStringPointer(input.AADDeviceID),
			FQDN:          cloneStringPointer(input.FQDN),
			Hostname:      cloneStringPointer(input.Hostname),
			Location:      cloneStringPointer(input.Location),
			OSPlatform:    cloneStringPointer(input.OSPlatform),
			BusinessOwner: cloneStringPointer(input.BusinessOwner),
			Criticality:   cloneStringPointer(input.Criticality),
			ContainmentStatus: cloneStringPointer(
				input.ContainmentStatus,
			),
			HostState:     "stub",
			EntityOrigin:  input.EntityOrigin,
			SeedMentionID: input.SeedMentionID,
			RowVersion:    1,
			CreatedAt:     now.UTC(),
			UpdatedAt:     now.UTC(),
			CreatedByUser: actor.ID,
			UpdatedByUser: actor.ID,
		}
		recordID, err := s.ports.records.InsertTx(ctx, tx, entityRecordInsertParams{
			IncidentID:      incidentID,
			RecordType:      "host",
			CreatedByUserID: actor.ID,
			CreatedAt:       record.CreatedAt,
			UpdatedByUserID: actor.ID,
			UpdatedAt:       record.UpdatedAt,
			RowVersion:      record.RowVersion,
		})
		if err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		record.RecordID = recordID
		if err := insertHostTx(ctx, tx, &record); err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, record.RecordID, "host", hostIdentifierSeeds(input), actor.ID, now); err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		aliasResult, err := syncEntityAliasesTx(ctx, tx, incidentID, record.RecordID, "host", input.AliasAdds, actor.ID, now)
		if err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		record.AliasMutations = aliasResult.Added
		if err := hydrateHostRecordTx(ctx, tx, &record); err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		return record, nil, "create", httpStatusCreated, nil
	}

	if err := hydrateHostRecordTx(ctx, tx, &current); err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	beforeRow := buildHostRow(current)

	next := current
	fieldChanged := false
	if input.AllowDisplayNameUpdate && strings.TrimSpace(input.DisplayName) != "" && input.DisplayName != current.DisplayName {
		next.DisplayName = input.DisplayName
		fieldChanged = true
	}
	if current.AADDeviceID == nil && input.AADDeviceID != nil {
		next.AADDeviceID = cloneStringPointer(input.AADDeviceID)
		fieldChanged = true
	}
	if current.FQDN == nil && input.FQDN != nil {
		next.FQDN = cloneStringPointer(input.FQDN)
		fieldChanged = true
	}
	if current.Hostname == nil && input.Hostname != nil {
		next.Hostname = cloneStringPointer(input.Hostname)
		fieldChanged = true
	}
	if current.Location == nil && input.Location != nil {
		next.Location = cloneStringPointer(input.Location)
		fieldChanged = true
	}
	if current.OSPlatform == nil && input.OSPlatform != nil {
		next.OSPlatform = cloneStringPointer(input.OSPlatform)
		fieldChanged = true
	}
	if current.BusinessOwner == nil && input.BusinessOwner != nil {
		next.BusinessOwner = cloneStringPointer(input.BusinessOwner)
		fieldChanged = true
	}
	if current.Criticality == nil && input.Criticality != nil {
		next.Criticality = cloneStringPointer(input.Criticality)
		fieldChanged = true
	}
	if current.ContainmentStatus == nil && input.ContainmentStatus != nil {
		next.ContainmentStatus = cloneStringPointer(input.ContainmentStatus)
		fieldChanged = true
	}

	identifierChanged, err := hasPendingIdentifierSeedsTx(ctx, tx, current.RecordID, "host", hostIdentifierSeeds(input))
	if err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	aliasChanged, err := hasPendingAliasAddsTx(ctx, tx, current.RecordID, "host", input.AliasAdds)
	if err != nil {
		return HostRecord{}, nil, "", 0, err
	}

	if fieldChanged || identifierChanged || aliasChanged {
		next.RowVersion, err = s.ports.records.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
		if err != nil {
			return HostRecord{}, nil, "", 0, err
		}
		next.UpdatedAt = now.UTC()
		next.UpdatedByUser = actor.ID
		if err := updateHostTx(ctx, tx, next); err != nil {
			return HostRecord{}, nil, "", 0, err
		}
	}
	if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, current.RecordID, "host", hostIdentifierSeeds(input), actor.ID, now); err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	aliasResult, err := syncEntityAliasesTx(ctx, tx, incidentID, current.RecordID, "host", input.AliasAdds, actor.ID, now)
	if err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	next.AliasMutations = aliasResult.Added
	if err := hydrateHostRecordTx(ctx, tx, &next); err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	return next, beforeRow, "patch", httpStatusOK, nil
}

func (s *mutationCore) applyIdentityUpsertTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, input identityUpsertInput, current IdentityRecord, matched bool, now time.Time) (IdentityRecord, map[string]any, string, int, error) {
	if !matched {
		record := IdentityRecord{
			IncidentID:     incidentID,
			DisplayName:    input.DisplayName,
			AADObjectID:    cloneStringPointer(input.AADObjectID),
			SID:            cloneStringPointer(input.SID),
			UPN:            cloneStringPointer(input.UPN),
			Email:          cloneStringPointer(input.Email),
			SamAccountName: cloneStringPointer(input.SamAccountName),
			PrivilegeLevel: cloneStringPointer(input.PrivilegeLevel),
			MFAState:       cloneStringPointer(input.MFAState),
			ResetStatus:    cloneStringPointer(input.ResetStatus),
			IdentityState:  "stub",
			EntityOrigin:   input.EntityOrigin,
			SeedMentionID:  input.SeedMentionID,
			RowVersion:     1,
			CreatedAt:      now.UTC(),
			UpdatedAt:      now.UTC(),
			CreatedByUser:  actor.ID,
			UpdatedByUser:  actor.ID,
		}
		recordID, err := s.ports.records.InsertTx(ctx, tx, entityRecordInsertParams{
			IncidentID:      incidentID,
			RecordType:      "identity",
			CreatedByUserID: actor.ID,
			CreatedAt:       record.CreatedAt,
			UpdatedByUserID: actor.ID,
			UpdatedAt:       record.UpdatedAt,
			RowVersion:      record.RowVersion,
		})
		if err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		record.RecordID = recordID
		if err := insertIdentityTx(ctx, tx, &record); err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, record.RecordID, "identity", identityIdentifierSeeds(input), actor.ID, now); err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		aliasResult, err := syncEntityAliasesTx(ctx, tx, incidentID, record.RecordID, "identity", input.AliasAdds, actor.ID, now)
		if err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		record.AliasMutations = aliasResult.Added
		if err := hydrateIdentityRecordTx(ctx, tx, &record); err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		return record, nil, "create", httpStatusCreated, nil
	}

	if err := hydrateIdentityRecordTx(ctx, tx, &current); err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	beforeRow := buildIdentityRow(current)

	next := current
	fieldChanged := false
	if input.AllowDisplayNameUpdate && strings.TrimSpace(input.DisplayName) != "" && input.DisplayName != current.DisplayName {
		next.DisplayName = input.DisplayName
		fieldChanged = true
	}
	if current.AADObjectID == nil && input.AADObjectID != nil {
		next.AADObjectID = cloneStringPointer(input.AADObjectID)
		fieldChanged = true
	}
	if current.SID == nil && input.SID != nil {
		next.SID = cloneStringPointer(input.SID)
		fieldChanged = true
	}
	if current.UPN == nil && input.UPN != nil {
		next.UPN = cloneStringPointer(input.UPN)
		fieldChanged = true
	}
	if current.Email == nil && input.Email != nil {
		next.Email = cloneStringPointer(input.Email)
		fieldChanged = true
	}
	if current.SamAccountName == nil && input.SamAccountName != nil {
		next.SamAccountName = cloneStringPointer(input.SamAccountName)
		fieldChanged = true
	}
	if current.PrivilegeLevel == nil && input.PrivilegeLevel != nil {
		next.PrivilegeLevel = cloneStringPointer(input.PrivilegeLevel)
		fieldChanged = true
	}
	if current.MFAState == nil && input.MFAState != nil {
		next.MFAState = cloneStringPointer(input.MFAState)
		fieldChanged = true
	}
	if current.ResetStatus == nil && input.ResetStatus != nil {
		next.ResetStatus = cloneStringPointer(input.ResetStatus)
		fieldChanged = true
	}

	identifierChanged, err := hasPendingIdentifierSeedsTx(ctx, tx, current.RecordID, "identity", identityIdentifierSeeds(input))
	if err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	aliasChanged, err := hasPendingAliasAddsTx(ctx, tx, current.RecordID, "identity", input.AliasAdds)
	if err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}

	if fieldChanged || identifierChanged || aliasChanged {
		next.RowVersion, err = s.ports.records.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
		if err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
		next.UpdatedAt = now.UTC()
		next.UpdatedByUser = actor.ID
		if err := updateIdentityTx(ctx, tx, next); err != nil {
			return IdentityRecord{}, nil, "", 0, err
		}
	}
	if _, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, current.RecordID, "identity", identityIdentifierSeeds(input), actor.ID, now); err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	aliasResult, err := syncEntityAliasesTx(ctx, tx, incidentID, current.RecordID, "identity", input.AliasAdds, actor.ID, now)
	if err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	next.AliasMutations = aliasResult.Added
	if err := hydrateIdentityRecordTx(ctx, tx, &next); err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	return next, beforeRow, "patch", httpStatusOK, nil
}
