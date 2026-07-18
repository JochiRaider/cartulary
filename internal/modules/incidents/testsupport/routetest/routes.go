package routetest

import (
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

const (
	MutationOwnerIncidentResource   routeinventory.MutationOwner = "incident resource mutation"
	MutationOwnerIncidentMembership routeinventory.MutationOwner = "incident membership mutation"
)

func PublicIncidentCore() []routeinventory.Entry {
	return []routeinventory.Entry{
		{Name: "incident list", Transport: routeinventory.TransportHTTP, Method: http.MethodGet, Template: "/api/v1/incidents", SuccessStatus: http.StatusOK, SuccessEnvelope: true},
		{
			Name: "incident create", Transport: routeinventory.TransportHTTP, Method: http.MethodPost, Template: "/api/v1/incidents",
			SuccessStatus: http.StatusCreated, SuccessEnvelope: true, RequiresCSRF: true,
			MutationOwners: []routeinventory.MutationOwner{MutationOwnerIncidentResource, MutationOwnerIncidentMembership},
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{
					"client_txn_id": "txn-incident-membership-public-incident-create-" + fixture.ClientTxnSuffix,
					"incident_key":  "IR-PUBLIC-" + strings.ToUpper(fixture.ClientTxnSuffix),
					"title":         "Public inventory incident " + fixture.ClientTxnSuffix,
				}
			},
		},
		{Name: "incident get", Transport: routeinventory.TransportHTTP, Method: http.MethodGet, Template: "/api/v1/incidents/{incident_id}", SuccessStatus: http.StatusOK, SuccessEnvelope: true},
		{
			Name: "incident patch", Transport: routeinventory.TransportHTTP, Method: http.MethodPatch, Template: "/api/v1/incidents/{incident_id}",
			SuccessStatus: http.StatusOK, SuccessEnvelope: true, RequiresCSRF: true,
			MutationOwners: []routeinventory.MutationOwner{MutationOwnerIncidentResource},
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"base_incident_version": fixture.BaseIncidentVersion, "tlp": "TLP:AMBER"}
			},
		},
	}
}

func PublicMembershipAdmin() []routeinventory.Entry {
	return []routeinventory.Entry{
		{Name: "memberships list", Transport: routeinventory.TransportHTTP, Method: http.MethodGet, Template: "/api/v1/incidents/{incident_id}/memberships", SuccessStatus: http.StatusOK, SuccessEnvelope: true},
		{
			Name: "membership create", Transport: routeinventory.TransportHTTP, Method: http.MethodPost, Template: "/api/v1/incidents/{incident_id}/memberships",
			SuccessStatus: http.StatusCreated, SuccessEnvelope: true, RequiresCSRF: true,
			MutationOwners: []routeinventory.MutationOwner{MutationOwnerIncidentMembership},
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"client_txn_id": "txn-incident-membership-public-membership-create-" + fixture.ClientTxnSuffix, "user_id": fixture.CandidateUserID, "role": "viewer"}
			},
		},
		{
			Name: "membership patch", Transport: routeinventory.TransportHTTP, Method: http.MethodPatch, Template: "/api/v1/incidents/{incident_id}/memberships/{member_user_id}",
			SuccessStatus: http.StatusOK, SuccessEnvelope: true, RequiresCSRF: true,
			MutationOwners: []routeinventory.MutationOwner{MutationOwnerIncidentMembership},
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"base_membership_version": fixture.BaseMembershipVersion, "role": "reviewer"}
			},
		},
		{
			Name: "membership delete", Transport: routeinventory.TransportHTTP, Method: http.MethodDelete, Template: "/api/v1/incidents/{incident_id}/memberships/{member_user_id}",
			SuccessStatus: http.StatusNoContent, RequiresCSRF: true,
			MutationOwners: []routeinventory.MutationOwner{MutationOwnerIncidentMembership},
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"base_membership_version": fixture.BaseMembershipVersion}
			},
		},
	}
}

func ControlIncidentCore() []routeinventory.Entry {
	return []routeinventory.Entry{
		{Name: "incident get", Transport: routeinventory.TransportHTTP, Method: http.MethodGet, Template: "/api/v1/incidents/{incident_id}", SuccessStatus: http.StatusOK, SuccessEnvelope: true, AllowedRole: routeinventory.ControlRoleMembershipRequired},
		{
			Name: "incident patch", Transport: routeinventory.TransportHTTP, Method: http.MethodPatch, Template: "/api/v1/incidents/{incident_id}",
			SuccessStatus: http.StatusOK, SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleReviewerOrHigher,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"base_incident_version": fixture.BaseIncidentVersion, "tlp": "TLP:AMBER"}
			},
		},
	}
}

func ControlMembershipAdmin() []routeinventory.Entry {
	return []routeinventory.Entry{
		{Name: "memberships list", Transport: routeinventory.TransportHTTP, Method: http.MethodGet, Template: "/api/v1/incidents/{incident_id}/memberships", SuccessStatus: http.StatusOK, SuccessEnvelope: true, AllowedRole: routeinventory.ControlRoleMembershipRequired},
		{
			Name: "membership create", Transport: routeinventory.TransportHTTP, Method: http.MethodPost, Template: "/api/v1/incidents/{incident_id}/memberships",
			SuccessStatus: http.StatusCreated, SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleAdminOnly,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"client_txn_id": "txn-incident-membership-control-membership-create-" + fixture.ClientTxnSuffix, "user_id": fixture.CandidateUserID, "role": "viewer"}
			},
		},
		{
			Name: "membership patch", Transport: routeinventory.TransportHTTP, Method: http.MethodPatch, Template: "/api/v1/incidents/{incident_id}/memberships/{member_user_id}",
			SuccessStatus: http.StatusOK, SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleAdminOnly,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"base_membership_version": fixture.BaseMembershipVersion, "role": "reviewer"}
			},
		},
		{
			Name: "membership delete", Transport: routeinventory.TransportHTTP, Method: http.MethodDelete, Template: "/api/v1/incidents/{incident_id}/memberships/{member_user_id}",
			SuccessStatus: http.StatusNoContent, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleAdminOnly,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"base_membership_version": fixture.BaseMembershipVersion}
			},
		},
	}
}
