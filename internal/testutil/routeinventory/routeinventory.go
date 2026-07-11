package routeinventory

import (
	"fmt"
	"strings"
)

type Transport string

const (
	TransportHTTP      Transport = "http"
	TransportWebSocket Transport = "websocket"
)

type MutationOwner string

type ControlRoleTier string

const (
	ControlRoleMembershipRequired ControlRoleTier = "membership-required"
	ControlRoleEditorOrHigher     ControlRoleTier = "editor|reviewer|admin"
	ControlRoleReviewerOrHigher   ControlRoleTier = "reviewer|admin"
	ControlRoleAdminOnly          ControlRoleTier = "admin"
)

type Fixture struct {
	IncidentID            string
	AdminUserID           string
	CandidateUserID       string
	MemberUserID          string
	PrimaryRecordID       string
	ReplacementRecordID   string
	BaseIncidentVersion   int64
	BaseRecordVersion     int64
	BaseMembershipVersion int64
	ClientTxnSuffix       string
}

type Entry struct {
	Name            string
	Transport       Transport
	Method          string
	Template        string
	SuccessStatus   int
	SuccessEnvelope bool
	RequiresCSRF    bool
	AllowedRole     ControlRoleTier
	MutationOwners  []MutationOwner
	Body            func(Fixture) map[string]any
}

func BuildPath(template string, fixture Fixture) string {
	replacer := strings.NewReplacer(
		"{incident_id}", fixture.IncidentID,
		"{admin_user_id}", fixture.AdminUserID,
		"{candidate_user_id}", fixture.CandidateUserID,
		"{member_user_id}", fixture.MemberUserID,
		"{record_id}", fixture.PrimaryRecordID,
		"{replacement_record_id}", fixture.ReplacementRecordID,
	)
	return replacer.Replace(template)
}

func Select(inventory []Entry, names ...string) []Entry {
	byName := make(map[string]Entry, len(inventory))
	for _, route := range inventory {
		byName[route.Name] = route
	}

	routes := make([]Entry, 0, len(names))
	for _, name := range names {
		route, ok := byName[name]
		if !ok {
			panic(fmt.Sprintf("missing route inventory entry %q", name))
		}
		routes = append(routes, route)
	}
	return routes
}
