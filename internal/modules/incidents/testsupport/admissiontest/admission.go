package admissiontest

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
)

type IncidentCreateInput struct {
	ClientTxnID            string
	IncidentKey            string
	Title                  string
	Description            *string
	Severity               *string
	TLP                    *string
	CurrentPhase           *string
	PrimaryExternalCaseRef *string
}

type OptionalStringInput struct {
	Present bool
	Value   *string
}

type IncidentPatchInput struct {
	BaseIncidentVersion    int64
	Description            OptionalStringInput
	Severity               OptionalStringInput
	TLP                    OptionalStringInput
	CurrentPhase           OptionalStringInput
	PrimaryExternalCaseRef OptionalStringInput
}

type IncidentLifecycleInput struct {
	BaseIncidentVersion int64
	ClientTxnID         string
	Reason              string
}

type MembershipCreateInput struct {
	ClientTxnID string
	UserID      *uuid.UUID
	Email       *string
	Role        string
}

type MembershipPatchInput struct {
	BaseMembershipVersion int64
	Role                  string
}

func IncidentCreate(t testing.TB, input IncidentCreateInput) incidents.IncidentCreateAdmission {
	t.Helper()
	request, admissionErr := incidents.AdmitIncidentCreateJSON(jsonReader(t, map[string]any{
		"client_txn_id": input.ClientTxnID, "incident_key": input.IncidentKey, "title": input.Title,
		"description": input.Description, "severity": input.Severity, "tlp": input.TLP,
		"current_phase": input.CurrentPhase, "primary_external_case_ref": input.PrimaryExternalCaseRef,
	}))
	requireAdmission(t, admissionErr)
	return request
}

func IncidentPatch(t testing.TB, input IncidentPatchInput) incidents.IncidentPatchAdmission {
	t.Helper()
	body := map[string]any{"base_incident_version": input.BaseIncidentVersion}
	putOptional(body, "description", input.Description)
	putOptional(body, "severity", input.Severity)
	putOptional(body, "tlp", input.TLP)
	putOptional(body, "current_phase", input.CurrentPhase)
	putOptional(body, "primary_external_case_ref", input.PrimaryExternalCaseRef)
	request, admissionErr := incidents.AdmitIncidentPatchJSON(jsonReader(t, body))
	requireAdmission(t, admissionErr)
	return request
}

func IncidentLifecycle(
	t testing.TB,
	action incidents.LifecycleAction,
	input IncidentLifecycleInput,
) incidents.IncidentLifecycleAdmission {
	t.Helper()
	request, admissionErr := incidents.AdmitIncidentLifecycleJSON(action, jsonReader(t, map[string]any{
		"base_incident_version": input.BaseIncidentVersion,
		"client_txn_id":         input.ClientTxnID,
		"reason":                input.Reason,
	}))
	requireAdmission(t, admissionErr)
	return request
}

func MembershipCreate(t testing.TB, input MembershipCreateInput) incidents.MembershipCreateAdmission {
	t.Helper()
	body := map[string]any{
		"client_txn_id": input.ClientTxnID,
		"role":          input.Role,
	}
	if input.UserID != nil {
		body["user_id"] = input.UserID.String()
	}
	if input.Email != nil {
		body["email"] = *input.Email
	}
	request, admissionErr := incidents.AdmitMembershipCreateJSON(jsonReader(t, body))
	requireAdmission(t, admissionErr)
	return request
}

func MembershipPatch(t testing.TB, input MembershipPatchInput) incidents.MembershipPatchAdmission {
	t.Helper()
	request, admissionErr := incidents.AdmitMembershipPatchJSON(jsonReader(t, map[string]any{
		"base_membership_version": input.BaseMembershipVersion,
		"role":                    input.Role,
	}))
	requireAdmission(t, admissionErr)
	return request
}

func MembershipDelete(t testing.TB, baseMembershipVersion int64) incidents.MembershipDeleteAdmission {
	t.Helper()
	request, admissionErr := incidents.AdmitMembershipDeleteJSON(jsonReader(t, map[string]any{
		"base_membership_version": baseMembershipVersion,
	}))
	requireAdmission(t, admissionErr)
	return request
}

func jsonReader(t testing.TB, body map[string]any) *bytes.Reader {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal incident mutation admission: %v", err)
	}
	return bytes.NewReader(encoded)
}

func putOptional(body map[string]any, field string, input OptionalStringInput) {
	if input.Present {
		body[field] = input.Value
	}
}

func requireAdmission(t testing.TB, admissionErr *incidents.AdmissionError) {
	t.Helper()
	if admissionErr != nil {
		field, _ := admissionErr.Field()
		t.Fatalf("admit incident mutation fixture: field=%q reason=%q", field, admissionErr.ReasonCode())
	}
}
