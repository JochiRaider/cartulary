package incidents

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestNewApplicationValidatesDependenciesInOrder_Unit(t *testing.T) {
	validPostgres := &pgxpool.Pool{}
	validBootstrap := workbookstartuppostgres.NewWriter()
	var typedNilPostgres *pgxpool.Pool
	var typedNilBootstrap *workbookstartuppostgres.Writer

	tests := []struct {
		name         string
		dependencies ApplicationDependencies
		wantError    string
	}{
		{name: "missing Postgres", dependencies: ApplicationDependencies{}, wantError: "Postgres dependency is required"},
		{name: "typed nil Postgres", dependencies: ApplicationDependencies{Postgres: typedNilPostgres, PreferenceBootstrap: validBootstrap}, wantError: "Postgres dependency is required"},
		{name: "missing bootstrap", dependencies: ApplicationDependencies{Postgres: validPostgres}, wantError: "workbook preference bootstrap dependency is required"},
		{name: "typed nil bootstrap", dependencies: ApplicationDependencies{Postgres: validPostgres, PreferenceBootstrap: typedNilBootstrap}, wantError: "workbook preference bootstrap dependency is required"},
		{name: "valid", dependencies: ApplicationDependencies{Postgres: validPostgres, PreferenceBootstrap: validBootstrap}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			application, err := NewApplication(test.dependencies)
			if test.wantError == "" {
				if err != nil || application == nil {
					t.Fatalf("valid construction = (%#v, %v)", application, err)
				}
				return
			}
			if application != nil || err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("invalid construction = (%#v, %v), want error containing %q", application, err, test.wantError)
			}
		})
	}
}

func TestIncidentPatchKeepsNoOpStableAndAdvancesMaterialChange_Unit(t *testing.T) {
	current := IncidentRecord{
		ID:                     uuid.MustParse("00000000-0000-0000-0000-000000000905"),
		TLP:                    stringPointerForTest("TLP:AMBER"),
		CurrentPhase:           stringPointerForTest("containment"),
		PrimaryExternalCaseRef: stringPointerForTest("CASE-1"),
		UpdatedAt:              time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC),
		IncidentVersion:        7,
	}
	actorID := uuid.MustParse("00000000-0000-0000-0000-000000000777")
	updatedAt := time.Date(2026, 4, 17, 13, 0, 0, 0, time.UTC)

	noOp, changed := applyIncidentPatch(current, IncidentPatchRequest{BaseIncidentVersion: 7}, actorID, updatedAt)
	if changed || noOp.IncidentVersion != current.IncidentVersion || !noOp.UpdatedAt.Equal(current.UpdatedAt) {
		t.Fatalf("no-op patch changed stable state: before=%#v after=%#v changed=%t", current, noOp, changed)
	}

	material, changed := applyIncidentPatch(current, IncidentPatchRequest{
		BaseIncidentVersion:    7,
		TLP:                    OptionalNullableString{Present: true, Value: stringPointerForTest("TLP:GREEN")},
		PrimaryExternalCaseRef: OptionalNullableString{Present: true},
	}, actorID, updatedAt)
	if !changed || material.IncidentVersion != 8 || material.TLP == nil || *material.TLP != "TLP:GREEN" ||
		material.PrimaryExternalCaseRef != nil || material.UpdatedByUserID == nil || *material.UpdatedByUserID != actorID ||
		!material.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("material patch projection drifted: %#v changed=%t", material, changed)
	}
}

func TestLastAdminGuardDetectsSoleAdministrator_Unit(t *testing.T) {
	nextRole := "reviewer"
	if !wouldLeaveNoIncidentAdmins("admin", 1, &nextRole, false) {
		t.Fatal("demoting the last admin must be rejected")
	}
	if wouldLeaveNoIncidentAdmins("admin", 2, &nextRole, false) {
		t.Fatal("demoting one of two admins must be allowed")
	}
	if !wouldLeaveNoIncidentAdmins("admin", 1, nil, true) {
		t.Fatal("deleting the last admin must be rejected")
	}
}

func TestValidateMembershipCreateTarget_Unit(t *testing.T) {
	targetID := uuid.MustParse("00000000-0000-0000-0000-000000000707")
	otherID := uuid.MustParse("00000000-0000-0000-0000-000000000708")
	email := "Analyst@Example.Test"
	target := authn.UserRecord{ID: targetID, Email: "analyst@example.test"}

	if err := validateMembershipCreateTarget(MembershipCreateRequest{UserID: &targetID}, target); err != nil {
		t.Fatalf("matching user_id target failed: %v", err)
	}
	if err := validateMembershipCreateTarget(MembershipCreateRequest{Email: &email}, target); err != nil {
		t.Fatalf("matching normalized email target failed: %v", err)
	}
	if err := validateMembershipCreateTarget(MembershipCreateRequest{UserID: &otherID}, target); err == nil {
		t.Fatal("mismatched user_id target was admitted")
	}
	otherEmail := "other@example.test"
	if err := validateMembershipCreateTarget(MembershipCreateRequest{Email: &otherEmail}, target); err == nil {
		t.Fatal("mismatched email target was admitted")
	}
	if err := validateMembershipCreateTarget(MembershipCreateRequest{}, target); err == nil {
		t.Fatal("missing selector was admitted")
	}
}

func stringPointerForTest(value string) *string {
	return &value
}
