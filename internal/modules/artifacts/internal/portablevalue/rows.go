// Package portablevalue owns exact Artifact source values independently of
// database-record JSON, derived workbook fields and transport coercion.
package portablevalue

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Identity struct {
	RecordID   uuid.UUID `json:"record_id"`
	IncidentID uuid.UUID `json:"incident_id"`
}

type Parent struct {
	Identity
	ArtifactType        string     `json:"artifact_type"`
	Title               *string    `json:"title"`
	Body                *string    `json:"body"`
	TimestampUTC        *time.Time `json:"timestamp_utc"`
	UpdatedAt           time.Time  `json:"updated_at"`
	CreatedAt           time.Time  `json:"created_at"`
	CommID              *string    `json:"comm_id"`
	CommType            *string    `json:"comm_type"`
	Audience            *string    `json:"audience"`
	ChannelOrMeeting    *string    `json:"channel_or_meeting"`
	Summary             *string    `json:"summary"`
	NextReportAt        *time.Time `json:"next_report_at"`
	PrivilegeTag        *string    `json:"privilege_tag"`
	HandoffID           *string    `json:"handoff_id"`
	OutgoingOwnerUserID *uuid.UUID `json:"outgoing_owner_user_id"`
	IncomingOwnerUserID *uuid.UUID `json:"incoming_owner_user_id"`
	CurrentStateSummary *string    `json:"current_state_summary"`
	NextChecks          *string    `json:"next_checks"`
	AcknowledgedAt      *time.Time `json:"acknowledged_at"`
	StatusReviewID      *string    `json:"status_review_id"`
	ReviewOwnerUserID   *uuid.UUID `json:"review_owner_user_id"`
	ActiveRisksSummary  *string    `json:"active_risks_summary"`
	LessonID            *string    `json:"lesson_id"`
	OwnerUserID         *uuid.UUID `json:"owner_user_id"`
	ClosureState        *string    `json:"closure_state"`
	CreatedByUserID     *uuid.UUID `json:"created_by_user_id"`
}

type Finding struct {
	Identity
	Kind            string     `json:"kind"`
	Statement       string     `json:"statement"`
	State           string     `json:"state"`
	ConfidenceScore *int64     `json:"confidence_score"`
	OwnerUserID     uuid.UUID  `json:"owner_user_id"`
	ClosedAt        *time.Time `json:"closed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Query struct {
	Identity
	QueryID         string    `json:"query_id"`
	Platform        string    `json:"platform"`
	Purpose         string    `json:"purpose"`
	QueryText       string    `json:"query_text"`
	CreatedByUserID uuid.UUID `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Keyword struct {
	Identity
	KeywordID     string    `json:"keyword_id"`
	Pattern       string    `json:"pattern"`
	Reason        string    `json:"reason"`
	MatchMode     string    `json:"match_mode"`
	CaseSensitive bool      `json:"case_sensitive"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Risk struct {
	RiskRefID       uuid.UUID  `json:"risk_ref_id"`
	IncidentID      uuid.UUID  `json:"incident_id"`
	HandoffRecordID uuid.UUID  `json:"handoff_record_id"`
	RiskRefText     string     `json:"risk_ref_text"`
	CreatedByUserID uuid.UUID  `json:"created_by_user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
	DeletedByUserID *uuid.UUID `json:"deleted_by_user_id"`
}

type scalarGroup struct {
	names    string
	kind     string
	nullable bool
}

// Parse checks exact member presence and scalar spelling before constructing a
// typed row. Semantic passes deliberately run separately across the whole family.
func Parse(table string, raw map[string]any) (any, bool) {
	groups := []scalarGroup{{"record_id incident_id", "uuid", false}, {"created_at updated_at", "time", false}}
	var destination any
	switch table {
	case "artifacts":
		destination = &Parent{}
		groups = append(groups,
			scalarGroup{"artifact_type", "text", false},
			scalarGroup{"title body comm_id comm_type audience channel_or_meeting summary privilege_tag handoff_id current_state_summary next_checks status_review_id active_risks_summary lesson_id closure_state", "text", true},
			scalarGroup{"timestamp_utc next_report_at acknowledged_at", "time", true},
			scalarGroup{"outgoing_owner_user_id incoming_owner_user_id review_owner_user_id owner_user_id created_by_user_id", "uuid", true},
		)
	case "artifact_findings":
		destination = &Finding{}
		groups = append(groups,
			scalarGroup{"kind statement state", "text", false},
			scalarGroup{"confidence_score", "integer", true},
			scalarGroup{"owner_user_id", "uuid", false},
			scalarGroup{"closed_at", "time", true},
		)
	case "artifact_investigative_queries":
		destination = &Query{}
		groups = append(groups,
			scalarGroup{"query_id platform purpose query_text", "text", false},
			scalarGroup{"created_by_user_id", "uuid", false},
		)
	case "artifact_forensic_keywords":
		destination = &Keyword{}
		groups = append(groups,
			scalarGroup{"keyword_id pattern reason match_mode", "text", false},
			scalarGroup{"case_sensitive", "boolean", false},
		)
	case "handoff_risk_refs":
		destination = &Risk{}
		groups = []scalarGroup{
			{"risk_ref_id incident_id handoff_record_id created_by_user_id", "uuid", false},
			{"risk_ref_text", "text", false}, {"created_at", "time", false},
			{"deleted_at", "time", true}, {"deleted_by_user_id", "uuid", true},
		}
	default:
		return nil, false
	}
	count := 0
	for _, group := range groups {
		for _, name := range strings.Fields(group.names) {
			count++
			value, present := raw[name]
			if !present || (value == nil && !group.nullable) {
				return nil, false
			}
			if value == nil {
				continue
			}
			valid := false
			switch group.kind {
			case "uuid":
				_, valid = CanonicalUUID(value)
			case "time":
				_, valid = CanonicalTime(value)
			case "text":
				text, ok := value.(string)
				valid = ok && utf8.ValidString(text)
			case "boolean":
				_, valid = value.(bool)
			case "integer":
				if number, ok := value.(json.Number); ok {
					integer, err := number.Int64()
					valid = err == nil && strconv.FormatInt(integer, 10) == number.String()
				}
			}
			if !valid {
				return nil, false
			}
		}
	}
	if count != len(raw) {
		return nil, false
	}
	encoded, err := json.Marshal(raw)
	if err != nil || json.Unmarshal(encoded, destination) != nil {
		return nil, false
	}
	UTCClocks(destination)
	return destination, true
}

func CanonicalUUID(raw any) (uuid.UUID, bool) {
	text, ok := raw.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(text)
	return id, err == nil && id != uuid.Nil && id.String() == text
}

func CanonicalTime(raw any) (time.Time, bool) {
	text, ok := raw.(string)
	if !ok {
		return time.Time{}, false
	}
	value, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || value.Year() < 1 || value.Nanosecond()%1000 != 0 {
		return time.Time{}, false
	}
	canonical := value.UTC().Format(time.RFC3339Nano)
	return value.UTC(), text == canonical || text == strings.TrimSuffix(canonical, "Z")+"+00:00"
}
