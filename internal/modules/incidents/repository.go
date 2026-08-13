package incidents

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
)

type repository struct {
	queries *sqlc.Queries
}

type createIncidentPersistenceParams struct {
	IncidentKey            string
	Title                  string
	Description            *string
	Severity               *string
	TLP                    *string
	CurrentPhase           *string
	PrimaryExternalCaseRef *string
	CreatedByUserID        uuid.UUID
	CreatedAt              time.Time
}

type createBootstrapMembershipPersistenceParams struct {
	IncidentID  uuid.UUID
	UserID      uuid.UUID
	JoinedAt    time.Time
	Role        string
	DisplayName string
}

type createMembershipPersistenceParams struct {
	IncidentID    uuid.UUID
	UserID        uuid.UUID
	Role          string
	JoinedAt      time.Time
	AddedByUserID uuid.UUID
	DisplayName   string
}

type updateIncidentLifecyclePersistenceParams struct {
	IncidentID      uuid.UUID
	Status          string
	ClosedAt        *time.Time
	UpdatedAt       time.Time
	UpdatedByUserID uuid.UUID
}

type updateMembershipPersistenceParams struct {
	IncidentID      uuid.UUID
	UserID          uuid.UUID
	Role            string
	UpdatedAt       time.Time
	UpdatedByUserID uuid.UUID
	DisplayName     string
}

type incidentBundleInitialAdmin struct {
	DisplayName       string
	IsActive          bool
	IsDeploymentAdmin bool
}

func newRepository(db sqlc.DBTX) *repository {
	return &repository{queries: sqlc.New(db)}
}

func (r *repository) listVisibleIncidentCandidates(
	ctx context.Context,
	userID uuid.UUID,
	page IncidentListPageRequest,
	after *IncidentListPosition,
	limit int,
) ([]IncidentRecord, error) {
	var afterUpdatedAt *time.Time
	var afterID *uuid.UUID
	if after != nil {
		updatedAt := after.UpdatedAt.UTC()
		afterUpdatedAt = &updatedAt
		id := after.ID
		afterID = &id
	}
	filterByStatus := false
	status := ""
	if page.Status != nil {
		filterByStatus = true
		status = *page.Status
	}
	rows, err := r.queries.ListVisibleIncidents(ctx, sqlc.ListVisibleIncidentsParams{
		UserID:  pgUUID(userID),
		Column2: pgOptionalTimestamptzPtr(page.AnchorUpdatedAt),
		Column3: pgOptionalTimestamptzPtr(afterUpdatedAt),
		Column4: pgOptionalUUIDPtr(afterID),
		Limit:   int32(limit),
		Column6: filterByStatus,
		Status:  status,
	})
	if err != nil {
		return nil, fmt.Errorf("list visible incidents: %w", err)
	}

	records := make([]IncidentRecord, 0, page.Limit)
	for _, row := range rows {
		record, err := incidentRecordFromSQL(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (r *repository) getVisibleIncident(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (IncidentRecord, error) {
	row, err := r.queries.GetVisibleIncidentByID(ctx, sqlc.GetVisibleIncidentByIDParams{
		ID:     pgUUID(incidentID),
		UserID: pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return IncidentRecord{}, ErrIncidentNotFound
	}
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("get visible incident: %w", err)
	}
	return incidentRecordFromSQL(row)
}

func (r *repository) getMembership(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (MembershipRecord, error) {
	row, err := r.queries.GetIncidentMembershipForActor(ctx, sqlc.GetIncidentMembershipForActorParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipRecord{}, ErrMembershipNotFound
	}
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("get incident membership: %w", err)
	}
	return membershipRecordFromSQL(row)
}

func (r *repository) listMemberships(
	ctx context.Context,
	incidentID uuid.UUID,
) ([]MembershipRecord, error) {
	rows, err := r.queries.ListAllIncidentMemberships(ctx, pgUUID(incidentID))
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}
	records := make([]MembershipRecord, 0, 16)
	for _, row := range rows {
		record, err := membershipRecordFromSQL(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (r *repository) createIncident(
	ctx context.Context,
	params createIncidentPersistenceParams,
) (IncidentRecord, error) {
	row, err := r.queries.CreateIncident(ctx, sqlc.CreateIncidentParams{
		IncidentKey:            params.IncidentKey,
		IncidentKeyCanonical:   params.IncidentKey,
		Title:                  params.Title,
		Description:            pgTextPtr(params.Description),
		Severity:               pgTextPtr(params.Severity),
		Tlp:                    pgTextPtr(params.TLP),
		CurrentPhase:           pgTextPtr(params.CurrentPhase),
		PrimaryExternalCaseRef: pgTextPtr(params.PrimaryExternalCaseRef),
		CreatedByUserID:        pgUUID(params.CreatedByUserID),
		CreatedAt:              pgTimestamptz(params.CreatedAt),
	})
	if isIncidentKeyConflict(err) {
		return IncidentRecord{}, ErrIncidentKeyConflict
	}
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("insert incident: %w", err)
	}
	return incidentRecordFromSQL(row)
}

func (r *repository) createBootstrapMembership(
	ctx context.Context,
	params createBootstrapMembershipPersistenceParams,
) (MembershipRecord, error) {
	row, err := r.queries.CreateBootstrapIncidentMembership(ctx, sqlc.CreateBootstrapIncidentMembershipParams{
		IncidentID: pgUUID(params.IncidentID),
		UserID:     pgUUID(params.UserID),
		JoinedAt:   pgTimestamptz(params.JoinedAt),
		Role:       params.Role,
		Column5:    params.DisplayName,
	})
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("insert bootstrap membership: %w", err)
	}
	return membershipRecordFromSQL(row)
}

func (r *repository) getIncidentForUpdate(
	ctx context.Context,
	incidentID uuid.UUID,
) (IncidentRecord, error) {
	row, err := r.queries.GetIncidentForUpdate(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IncidentRecord{}, ErrIncidentNotFound
	}
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("query incident for update: %w", err)
	}
	return incidentRecordFromSQL(row)
}

func (r *repository) updateIncidentMetadata(
	ctx context.Context,
	incidentID uuid.UUID,
	next IncidentRecord,
	actorUserID uuid.UUID,
) (IncidentRecord, error) {
	row, err := r.queries.UpdateIncidentMetadata(ctx, sqlc.UpdateIncidentMetadataParams{
		ID:                     pgUUID(incidentID),
		Description:            pgTextPtr(next.Description),
		Severity:               pgTextPtr(next.Severity),
		Tlp:                    pgTextPtr(next.TLP),
		CurrentPhase:           pgTextPtr(next.CurrentPhase),
		PrimaryExternalCaseRef: pgTextPtr(next.PrimaryExternalCaseRef),
		UpdatedAt:              pgTimestamptz(next.UpdatedAt),
		UpdatedByUserID:        pgUUID(actorUserID),
	})
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("update incident: %w", err)
	}
	return incidentRecordFromSQL(row)
}

func (r *repository) updateIncidentLifecycle(
	ctx context.Context,
	params updateIncidentLifecyclePersistenceParams,
) (IncidentRecord, error) {
	row, err := r.queries.UpdateIncidentLifecycle(ctx, sqlc.UpdateIncidentLifecycleParams{
		ID:              pgUUID(params.IncidentID),
		Status:          params.Status,
		ClosedAt:        pgOptionalTimestamptzPtr(params.ClosedAt),
		UpdatedAt:       pgTimestamptz(params.UpdatedAt),
		UpdatedByUserID: pgUUID(params.UpdatedByUserID),
	})
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("update incident lifecycle: %w", err)
	}
	return incidentRecordFromSQL(row)
}

func (r *repository) getMembershipForUpdate(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (MembershipRecord, error) {
	row, err := r.queries.GetIncidentMembershipForUpdate(ctx, sqlc.GetIncidentMembershipForUpdateParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipRecord{}, ErrMembershipNotFound
	}
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("query membership for update: %w", err)
	}
	return membershipRecordFromSQL(row)
}

func (r *repository) createMembership(
	ctx context.Context,
	params createMembershipPersistenceParams,
) (MembershipRecord, error) {
	row, err := r.queries.CreateIncidentMembership(ctx, sqlc.CreateIncidentMembershipParams{
		IncidentID:    pgUUID(params.IncidentID),
		UserID:        pgUUID(params.UserID),
		Role:          params.Role,
		JoinedAt:      pgTimestamptz(params.JoinedAt),
		AddedByUserID: pgUUID(params.AddedByUserID),
		Column6:       params.DisplayName,
	})
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("insert membership: %w", err)
	}
	return membershipRecordFromSQL(row)
}

func (r *repository) updateMembership(
	ctx context.Context,
	params updateMembershipPersistenceParams,
) (MembershipRecord, error) {
	row, err := r.queries.UpdateIncidentMembershipRole(ctx, sqlc.UpdateIncidentMembershipRoleParams{
		IncidentID:      pgUUID(params.IncidentID),
		UserID:          pgUUID(params.UserID),
		Role:            params.Role,
		UpdatedAt:       pgTimestamptz(params.UpdatedAt),
		UpdatedByUserID: pgUUID(params.UpdatedByUserID),
		Column6:         params.DisplayName,
	})
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("update membership: %w", err)
	}
	return membershipRecordFromSQL(row)
}

func (r *repository) deleteMembership(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) error {
	if err := r.queries.DeleteIncidentMembership(ctx, sqlc.DeleteIncidentMembershipParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	}); err != nil {
		return fmt.Errorf("delete membership: %w", err)
	}
	return nil
}

func (r *repository) countIncidentAdmins(ctx context.Context, incidentID uuid.UUID) (int, error) {
	count, err := r.queries.CountIncidentAdmins(ctx, pgUUID(incidentID))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *repository) ensureOpen(ctx context.Context, incidentID uuid.UUID) error {
	status, err := r.queries.EnsureIncidentOpenForMutation(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrIncidentNotFound
	}
	if err != nil {
		return fmt.Errorf("ensure incident open: %w", err)
	}
	if status == "closed" {
		return ErrIncidentClosed
	}
	return nil
}

func (r *repository) getIncidentBundleInitialAdminForUpdate(
	ctx context.Context,
	userID uuid.UUID,
) (incidentBundleInitialAdmin, error) {
	row, err := r.queries.GetIncidentBundleInitialAdminForUpdate(ctx, pgUUID(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return incidentBundleInitialAdmin{}, ErrInitialAdminUnavailable
	}
	if err != nil {
		return incidentBundleInitialAdmin{}, fmt.Errorf("read incident bundle initial admin: %w", err)
	}
	return incidentBundleInitialAdmin{
		DisplayName:       row.DisplayName,
		IsActive:          row.IsActive,
		IsDeploymentAdmin: row.IsDeploymentAdmin,
	}, nil
}

func incidentRecordFromSQL(row sqlc.Incident) (IncidentRecord, error) {
	id, err := uuidFromPG(row.ID)
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("incident id: %w", err)
	}
	createdBy, err := uuidFromPG(row.CreatedByUserID)
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("incident created by: %w", err)
	}
	createdAt, err := timeFromPG(row.CreatedAt)
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("incident created at: %w", err)
	}
	updatedAt, err := timeFromPG(row.UpdatedAt)
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("incident updated at: %w", err)
	}
	return IncidentRecord{
		ID:                     id,
		IncidentKey:            row.IncidentKey,
		Title:                  row.Title,
		Description:            optionalStringFromPG(row.Description),
		Status:                 row.Status,
		Severity:               optionalStringFromPG(row.Severity),
		TLP:                    optionalStringFromPG(row.Tlp),
		CurrentPhase:           optionalStringFromPG(row.CurrentPhase),
		PrimaryExternalCaseRef: optionalStringFromPG(row.PrimaryExternalCaseRef),
		CreatedByUserID:        createdBy,
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
		UpdatedByUserID:        optionalUUIDFromPG(row.UpdatedByUserID),
		IncidentVersion:        row.IncidentVersion,
		ClosedAt:               optionalTimeFromPG(row.ClosedAt),
	}, nil
}

func membershipRecordFromSQL(row any) (MembershipRecord, error) {
	switch typed := row.(type) {
	case sqlc.GetIncidentMembershipForActorRow:
		return membershipRecordFromSQLFields(typed.IncidentID, typed.UserID, typed.DisplayName, typed.Role, typed.JoinedAt, typed.AddedByUserID, typed.UpdatedAt, typed.UpdatedByUserID, typed.MembershipVersion)
	case sqlc.GetIncidentMembershipForUpdateRow:
		return membershipRecordFromSQLFields(typed.IncidentID, typed.UserID, typed.DisplayName, typed.Role, typed.JoinedAt, typed.AddedByUserID, typed.UpdatedAt, typed.UpdatedByUserID, typed.MembershipVersion)
	case sqlc.ListAllIncidentMembershipsRow:
		return membershipRecordFromSQLFields(typed.IncidentID, typed.UserID, typed.DisplayName, typed.Role, typed.JoinedAt, typed.AddedByUserID, typed.UpdatedAt, typed.UpdatedByUserID, typed.MembershipVersion)
	case sqlc.ListIncidentMembershipsRow:
		return membershipRecordFromSQLFields(typed.IncidentID, typed.UserID, typed.DisplayName, typed.Role, typed.JoinedAt, typed.AddedByUserID, typed.UpdatedAt, typed.UpdatedByUserID, typed.MembershipVersion)
	case sqlc.CreateBootstrapIncidentMembershipRow:
		return membershipRecordFromSQLFields(typed.IncidentID, typed.UserID, typed.DisplayName, typed.Role, typed.JoinedAt, typed.AddedByUserID, typed.UpdatedAt, typed.UpdatedByUserID, typed.MembershipVersion)
	case sqlc.CreateIncidentMembershipRow:
		return membershipRecordFromSQLFields(typed.IncidentID, typed.UserID, typed.DisplayName, typed.Role, typed.JoinedAt, typed.AddedByUserID, typed.UpdatedAt, typed.UpdatedByUserID, typed.MembershipVersion)
	case sqlc.UpdateIncidentMembershipRoleRow:
		return membershipRecordFromSQLFields(typed.IncidentID, typed.UserID, typed.DisplayName, typed.Role, typed.JoinedAt, typed.AddedByUserID, typed.UpdatedAt, typed.UpdatedByUserID, typed.MembershipVersion)
	default:
		return MembershipRecord{}, fmt.Errorf("unsupported membership SQL row %T", row)
	}
}

func membershipRecordFromSQLFields(
	rowIncidentID pgtype.UUID,
	rowUserID pgtype.UUID,
	rowDisplayName string,
	rowRole string,
	rowJoinedAt pgtype.Timestamptz,
	rowAddedByUserID pgtype.UUID,
	rowUpdatedAt pgtype.Timestamptz,
	rowUpdatedByUserID pgtype.UUID,
	rowMembershipVersion int64,
) (MembershipRecord, error) {
	incidentID, err := uuidFromPG(rowIncidentID)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership incident id: %w", err)
	}
	userID, err := uuidFromPG(rowUserID)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership user id: %w", err)
	}
	addedBy, err := uuidFromPG(rowAddedByUserID)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership added by: %w", err)
	}
	joinedAt, err := timeFromPG(rowJoinedAt)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership joined at: %w", err)
	}
	updatedAt, err := timeFromPG(rowUpdatedAt)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership updated at: %w", err)
	}
	return MembershipRecord{
		IncidentID:        incidentID,
		UserID:            userID,
		DisplayName:       rowDisplayName,
		Role:              rowRole,
		JoinedAt:          joinedAt,
		AddedByUserID:     addedBy,
		UpdatedAt:         updatedAt,
		UpdatedByUserID:   optionalUUIDFromPG(rowUpdatedByUserID),
		MembershipVersion: rowMembershipVersion,
	}, nil
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func pgOptionalUUIDPtr(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return pgUUID(*value)
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
}

func optionalUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	id, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return nil
	}
	return &id
}

func pgTextPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalStringFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func pgTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func pgOptionalTimestamptzPtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgTimestamptz(*value)
}

func timeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, errors.New("missing timestamp")
	}
	return value.Time.UTC(), nil
}

func optionalTimeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	timestamp := value.Time.UTC()
	return &timestamp
}

func isIncidentKeyConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == "incidents_incident_key_canonical_key"
}
