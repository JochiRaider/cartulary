package admission

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Role uint8

const (
	RoleViewer Role = iota + 1
	RoleEditor
	RoleReviewer
	RoleAdmin
)

type RoleSet uint8

func (role Role) String() string {
	switch role {
	case RoleViewer:
		return "viewer"
	case RoleEditor:
		return "editor"
	case RoleReviewer:
		return "reviewer"
	case RoleAdmin:
		return "admin"
	default:
		return ""
	}
}

const (
	roleViewerBit RoleSet = 1 << iota
	roleEditorBit
	roleReviewerBit
	roleAdminBit
	knownRoleBits = roleViewerBit | roleEditorBit | roleReviewerBit | roleAdminBit
)

const (
	RolesMember              RoleSet = knownRoleBits
	RolesEditorReviewerAdmin RoleSet = roleEditorBit | roleReviewerBit | roleAdminBit
	RolesReviewerAdmin       RoleSet = roleReviewerBit | roleAdminBit
	RolesEditorAdmin         RoleSet = roleEditorBit | roleAdminBit
	RolesAdmin               RoleSet = roleAdminBit
)

type Lifecycle uint8

const (
	LifecycleAny Lifecycle = iota + 1
	LifecycleOpen
)

type IncidentStatus uint8

const (
	IncidentStatusActive IncidentStatus = iota + 1
	IncidentStatusClosed
)

type Requirement struct {
	AllowedRoles RoleSet
	Lifecycle    Lifecycle
}

type Grant struct {
	Role           Role
	IncidentStatus IncidentStatus
}

type DenialCode string

const (
	DenialNotVisible       DenialCode = "not_visible"
	DenialInsufficientRole DenialCode = "insufficient_role"
	DenialIncidentClosed   DenialCode = "incident_closed"
)

type Denied struct {
	Code DenialCode
}

func (e *Denied) Error() string {
	return "incident admission denied: " + string(e.Code)
}

func IsDenied(err error, code DenialCode) bool {
	var denied *Denied
	return errors.As(err, &denied) && denied.Code == code
}

type Checker struct {
	db postgres.DB
}

func NewChecker(db postgres.DB) *Checker {
	return &Checker{db: db}
}

func (c *Checker) Check(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	requirement Requirement,
) (Grant, error) {
	if err := validateChecker(c); err != nil {
		return Grant{}, err
	}
	if err := validateRequirement(requirement); err != nil {
		return Grant{}, err
	}
	var storedRole string
	var storedStatus string
	err := c.db.QueryRow(ctx, `
SELECT m.role, i.status
  FROM incidents i
  JOIN incident_memberships m
    ON m.incident_id = i.id
   AND m.user_id = $2
 WHERE i.id = $1
`, incidentID, userID).Scan(&storedRole, &storedStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Grant{}, &Denied{Code: DenialNotVisible}
	}
	if err != nil {
		return Grant{}, fmt.Errorf("check incident admission: %w", err)
	}
	return decide(storedRole, storedStatus, requirement)
}

func (c *Checker) CheckTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	userID uuid.UUID,
	requirement Requirement,
) (Grant, error) {
	if err := validateChecker(c, tx); err != nil {
		return Grant{}, err
	}
	if err := validateRequirement(requirement); err != nil {
		return Grant{}, err
	}
	storedStatus, err := readStatusForShare(ctx, tx, incidentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Grant{}, &Denied{Code: DenialNotVisible}
	}
	if err != nil {
		return Grant{}, fmt.Errorf("lock incident admission: %w", err)
	}
	var storedRole string
	err = tx.QueryRow(ctx, `
SELECT role
  FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
 FOR SHARE
`, incidentID, userID).Scan(&storedRole)
	if errors.Is(err, pgx.ErrNoRows) {
		return Grant{}, &Denied{Code: DenialNotVisible}
	}
	if err != nil {
		return Grant{}, fmt.Errorf("lock incident membership admission: %w", err)
	}
	return decide(storedRole, storedStatus, requirement)
}

func (c *Checker) RequireOpenTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if err := validateChecker(c, tx); err != nil {
		return err
	}
	storedStatus, err := readStatusForShare(ctx, tx, incidentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Denied{Code: DenialNotVisible}
	}
	if err != nil {
		return fmt.Errorf("lock incident lifecycle admission: %w", err)
	}
	status, err := parseStatus(storedStatus)
	if err != nil {
		return err
	}
	if status == IncidentStatusClosed {
		return &Denied{Code: DenialIncidentClosed}
	}
	return nil
}

func readStatusForShare(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (string, error) {
	var storedStatus string
	err := tx.QueryRow(ctx, `
SELECT status
  FROM incidents
 WHERE id = $1
 FOR SHARE
`, incidentID).Scan(&storedStatus)
	return storedStatus, err
}

func decide(storedRole string, storedStatus string, requirement Requirement) (Grant, error) {
	role, err := parseRole(storedRole)
	if err != nil {
		return Grant{}, err
	}
	status, err := parseStatus(storedStatus)
	if err != nil {
		return Grant{}, err
	}
	if !requirement.AllowedRoles.includes(role) {
		return Grant{}, &Denied{Code: DenialInsufficientRole}
	}
	if requirement.Lifecycle == LifecycleOpen && status == IncidentStatusClosed {
		return Grant{}, &Denied{Code: DenialIncidentClosed}
	}
	return Grant{Role: role, IncidentStatus: status}, nil
}

func validateChecker(checker *Checker, values ...any) error {
	if checker == nil {
		return errors.New("incident admission checker is required")
	}
	values = append([]any{checker.db}, values...)
	for _, value := range values {
		if value == nil || isNilDependency(value) {
			return errors.New("incident admission dependency is required")
		}
	}
	return nil
}

func isNilDependency(value any) bool {
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func validateRequirement(requirement Requirement) error {
	if requirement.AllowedRoles == 0 || requirement.AllowedRoles&^knownRoleBits != 0 {
		return fmt.Errorf("invalid incident admission role set: %08b", requirement.AllowedRoles)
	}
	if requirement.Lifecycle != LifecycleAny && requirement.Lifecycle != LifecycleOpen {
		return fmt.Errorf("invalid incident admission lifecycle: %d", requirement.Lifecycle)
	}
	return nil
}

func (roles RoleSet) includes(role Role) bool {
	bit, err := roleBit(role)
	return err == nil && roles&bit != 0
}

func roleBit(role Role) (RoleSet, error) {
	switch role {
	case RoleViewer:
		return roleViewerBit, nil
	case RoleEditor:
		return roleEditorBit, nil
	case RoleReviewer:
		return roleReviewerBit, nil
	case RoleAdmin:
		return roleAdminBit, nil
	default:
		return 0, fmt.Errorf("invalid incident admission role: %d", role)
	}
}

func parseRole(stored string) (Role, error) {
	switch stored {
	case "viewer":
		return RoleViewer, nil
	case "editor":
		return RoleEditor, nil
	case "reviewer":
		return RoleReviewer, nil
	case "admin":
		return RoleAdmin, nil
	default:
		return 0, fmt.Errorf("invalid stored incident role %q", stored)
	}
}

func parseStatus(stored string) (IncidentStatus, error) {
	switch stored {
	case "active":
		return IncidentStatusActive, nil
	case "closed":
		return IncidentStatusClosed, nil
	default:
		return 0, fmt.Errorf("invalid stored incident status %q", stored)
	}
}
