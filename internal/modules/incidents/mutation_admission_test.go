package incidents

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestOpaqueMutationAdmissionZeroValuesFailBeforeDependencies_Unit(t *testing.T) {
	t.Parallel()
	application := &Application{}
	ctx := context.Background()
	actor := authn.UserRecord{}
	incidentID := uuid.New()
	userID := uuid.New()

	if _, err := application.CreateIncident(ctx, actor, IncidentCreateAdmission{}, "request"); !errors.Is(err, errInvalidMutationAdmission) {
		t.Fatalf("zero incident create admission error = %v", err)
	}
	if _, _, err := application.UpdateIncident(ctx, actor, incidentID, IncidentPatchAdmission{}, "request"); !errors.Is(err, errInvalidMutationAdmission) {
		t.Fatalf("zero incident patch admission error = %v", err)
	}
	if _, err := application.TransitionIncidentLifecycle(ctx, actor, incidentID, IncidentLifecycleAdmission{}, "request"); !errors.Is(err, errInvalidMutationAdmission) {
		t.Fatalf("zero incident lifecycle admission error = %v", err)
	}
	if _, err := application.CreateMembership(ctx, actor, incidentID, authn.UserRecord{}, MembershipCreateAdmission{}, "request"); !errors.Is(err, errInvalidMutationAdmission) {
		t.Fatalf("zero membership create admission error = %v", err)
	}
	if _, _, err := application.UpdateMembership(ctx, actor, incidentID, userID, MembershipPatchAdmission{}, "request"); !errors.Is(err, errInvalidMutationAdmission) {
		t.Fatalf("zero membership patch admission error = %v", err)
	}
	if _, err := application.DeleteMembership(ctx, actor, incidentID, userID, MembershipDeleteAdmission{}, "request"); !errors.Is(err, errInvalidMutationAdmission) {
		t.Fatalf("zero membership delete admission error = %v", err)
	}
}

func TestMutationAdmissionsShareStrictRootJSONFraming_Unit(t *testing.T) {
	t.Parallel()
	factories := []struct {
		name  string
		admit func(string) *AdmissionError
	}{
		{name: "incident create", admit: func(body string) *AdmissionError {
			_, err := AdmitIncidentCreateJSON(strings.NewReader(body))
			return err
		}},
		{name: "incident patch", admit: func(body string) *AdmissionError {
			_, err := AdmitIncidentPatchJSON(strings.NewReader(body))
			return err
		}},
		{name: "incident lifecycle", admit: func(body string) *AdmissionError {
			_, err := AdmitIncidentLifecycleJSON(LifecycleActionClose, strings.NewReader(body))
			return err
		}},
		{name: "membership create", admit: func(body string) *AdmissionError {
			_, err := AdmitMembershipCreateJSON(strings.NewReader(body))
			return err
		}},
		{name: "membership patch", admit: func(body string) *AdmissionError {
			_, err := AdmitMembershipPatchJSON(strings.NewReader(body))
			return err
		}},
		{name: "membership delete", admit: func(body string) *AdmissionError {
			_, err := AdmitMembershipDeleteJSON(strings.NewReader(body))
			return err
		}},
	}
	invalidBodies := map[string]string{
		"array":          `[]`,
		"duplicate":      `{"duplicate":1,"duplicate":2}`,
		"malformed":      `{"broken":`,
		"trailing value": `{} {}`,
	}
	for _, factory := range factories {
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			t.Parallel()
			for name, body := range invalidBodies {
				admissionErr := factory.admit(body)
				if admissionErr == nil || admissionErr.ReasonCode() != "request_not_object" {
					t.Fatalf("%s body error = %#v", name, admissionErr)
				}
				if _, present := admissionErr.Field(); present {
					t.Fatalf("%s body exposed a field: %#v", name, admissionErr)
				}
			}
		})
	}
}

func TestMutationAdmissionNormalizesAtTheOwnerBoundary_Unit(t *testing.T) {
	create, admissionErr := AdmitIncidentCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01",
		"incident_key":"  IR-E\u0301-2026-001  ",
		"title":"  Incident E\u0301xample  ",
		"description":"  First line\r\nSecond line  ",
		"severity":"  high  ",
		"tlp":"TLP:AMBER",
		"current_phase":"  triage  ",
		"primary_external_case_ref":"  CASE-42  "
	}`))
	if admissionErr != nil {
		t.Fatalf("admit normalized incident create: %v", admissionErr)
	}
	if create.incidentKey != "IR-É-2026-001" || create.title != "Incident Éxample" ||
		create.description == nil || *create.description != "First line\nSecond line" ||
		create.severity == nil || *create.severity != "high" ||
		create.tlp == nil || *create.tlp != "TLP:AMBER" ||
		create.currentPhase == nil || *create.currentPhase != "triage" ||
		create.primaryExternalCaseRef == nil || *create.primaryExternalCaseRef != "CASE-42" {
		t.Fatalf("unexpected normalized incident create admission: %#v", create)
	}

	patch := mustIncidentPatchAdmission(t, `{
		"base_incident_version":7,
		"tlp":null,
		"current_phase":"  containment  ",
		"severity":" \t "
	}`)
	if !patch.tlp.present || patch.tlp.value != nil ||
		!patch.currentPhase.present || patch.currentPhase.value == nil || *patch.currentPhase.value != "containment" ||
		!patch.severity.present || patch.severity.value != nil || patch.primaryExternalCaseRef.present {
		t.Fatalf("unexpected normalized incident patch admission: %#v", patch)
	}

	lifecycle, admissionErr := AdmitIncidentLifecycleJSON(LifecycleActionClose, strings.NewReader(
		`{"base_incident_version":7,"client_txn_id":"txn-lifecycle","reason":"  cafe\u0301\r\nline\t "}`,
	))
	if admissionErr != nil {
		t.Fatalf("admit normalized lifecycle: %v", admissionErr)
	}
	if lifecycle.baseIncidentVersion != 7 || lifecycle.reason != "café\nline" {
		t.Fatalf("unexpected normalized lifecycle admission: %#v", lifecycle)
	}

	emailMembership := mustMembershipCreateAdmission(t, `{
		"client_txn_id":"txn-membership","email":"  Analyst@Example.Test  ","role":"reviewer"
	}`)
	if emailMembership.email == nil || *emailMembership.email != "Analyst@Example.Test" || emailMembership.userID != nil {
		t.Fatalf("unexpected normalized membership admission: %#v", emailMembership)
	}
}

func TestAdmissionErrorsExposeOnlyStableQueries_Unit(t *testing.T) {
	_, admissionErr := AdmitMembershipDeleteJSON(strings.NewReader(`{}`))
	if admissionErr == nil {
		t.Fatal("expected membership delete admission error")
	}
	field, present := admissionErr.Field()
	if !present || field != "base_membership_version" || admissionErr.ReasonCode() != "missing_required_field" {
		t.Fatalf("unexpected admission error queries: field=%q present=%t reason=%q", field, present, admissionErr.ReasonCode())
	}
	if reflect.TypeOf(*admissionErr).NumField() != 2 {
		t.Fatalf("admission error unexpectedly widened: %#v", admissionErr)
	}
}

func mustIncidentPatchAdmission(t testing.TB, body string) IncidentPatchAdmission {
	t.Helper()
	request, admissionErr := AdmitIncidentPatchJSON(strings.NewReader(body))
	if admissionErr != nil {
		field, _ := admissionErr.Field()
		t.Fatalf("admit incident patch: field=%q reason=%q", field, admissionErr.ReasonCode())
	}
	return request
}

func mustMembershipCreateAdmission(t testing.TB, body string) MembershipCreateAdmission {
	t.Helper()
	request, admissionErr := AdmitMembershipCreateJSON(strings.NewReader(body))
	if admissionErr != nil {
		field, _ := admissionErr.Field()
		t.Fatalf("admit membership create: field=%q reason=%q", field, admissionErr.ReasonCode())
	}
	return request
}
