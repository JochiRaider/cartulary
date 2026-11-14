package tasksdecisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

func validateDirectReferenceTx(ctx context.Context, tx pgx.Tx, targets ReferenceTargetCapability, catalog *sourcecatalog.Catalog, incidentID uuid.UUID, fieldKey string, recordID uuid.UUID) error {
	field, ok := catalog.Field(fieldKey)
	if !ok || field.Reference.Role != "same_incident_record" {
		return nil
	}
	return validateTargetRecordTx(ctx, tx, targets, incidentID, recordID, field.Reference.ExpectedTargetRecordType, fieldKey)
}

func isMemberUserReferenceField(catalog *sourcecatalog.Catalog, fieldKey string) bool {
	field, ok := catalog.Field(fieldKey)
	return ok && field.Reference.Role == "incident_member_user"
}

func validateIncidentMemberUserTx(ctx context.Context, tx pgx.Tx, incidentID, userID uuid.UUID, field string) error {
	return tasksource.ValidateMemberUserTx(ctx, tx, incidentID, userID, field)
}

func validateTargetRecordTx(ctx context.Context, tx pgx.Tx, targets ReferenceTargetCapability, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	return tasksource.ValidateTargetRecordTx(ctx, tx, targets, incidentID, recordID, expectedType, field)
}
