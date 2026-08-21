package workbookassembly

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type recordTargetAdapter struct{ owner *records.RouteTargetResolver }

func NewRecordTargetResolver(pool postgres.DB) workbook.RecordTargetResolver {
	return recordTargetAdapter{owner: records.NewRouteTargetResolver(pool)}
}

func (adapter recordTargetAdapter) ResolveRecordTarget(ctx context.Context, recordID uuid.UUID) (workbook.RecordTarget, error) {
	target, err := adapter.owner.Resolve(ctx, recordID)
	if errors.Is(err, records.ErrEnvelopeNotFound) || errors.Is(err, pgx.ErrNoRows) {
		return workbook.RecordTarget{}, workbook.ErrRecordTargetNotFound
	}
	if err != nil {
		return workbook.RecordTarget{}, err
	}
	lifecycle := workbook.RecordLifecycleActive
	if target.Deleted {
		lifecycle = workbook.RecordLifecycleDeleted
	}
	return workbook.RecordTarget{
		RecordID: recordID, IncidentID: target.IncidentID, RecordType: target.RecordType,
		LifecycleState: lifecycle,
	}, nil
}

type conflictTokenAdapter struct {
	owner conflicttokens.ConflictTokenCodec
}

func NewConflictTokenDecoder(owner conflicttokens.ConflictTokenCodec) workbook.ConflictTokenDecoder {
	return conflictTokenAdapter{owner: owner}
}

func (adapter conflictTokenAdapter) DecodeConflictToken(token string) (workbook.ConflictClaims, bool) {
	claims, ok := adapter.owner.Parse(token)
	if !ok {
		return workbook.ConflictClaims{}, false
	}
	recordID, err := uuid.Parse(claims.RecordID)
	if err != nil {
		return workbook.ConflictClaims{}, false
	}
	return workbook.ConflictClaims{
		Version: claims.Version, RecordID: recordID, ViewSchemaID: claims.ViewSchemaID,
		RouteKey: claims.RouteKey, FieldKey: claims.FieldKey,
		ConflictResolutionClass: claims.ConflictResolutionClass,
		BaseRowVersion:          claims.BaseRowVersion, CurrentRowVersion: claims.CurrentRowVersion,
		RequestHash: claims.RequestHash, IssuedAt: claims.IssuedAt, ExpiresAt: claims.ExpiresAt,
	}, true
}
