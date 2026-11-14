package portablevalue

import (
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/textpolicy"
	"github.com/google/uuid"
)

func IDs(row any) (uuid.UUID, uuid.UUID) {
	switch value := row.(type) {
	case *Parent:
		return value.RecordID, value.IncidentID
	case *Finding:
		return value.RecordID, value.IncidentID
	case *Query:
		return value.RecordID, value.IncidentID
	case *Keyword:
		return value.RecordID, value.IncidentID
	case *Risk:
		return value.RiskRefID, value.IncidentID
	default:
		return uuid.Nil, uuid.Nil
	}
}

func Actors(row any) map[string]uuid.UUID {
	actors := map[string]uuid.UUID{}
	put := func(column string, id *uuid.UUID) {
		if id != nil {
			actors[column] = *id
		}
	}
	switch value := row.(type) {
	case *Parent:
		put("outgoing_owner_user_id", value.OutgoingOwnerUserID)
		put("incoming_owner_user_id", value.IncomingOwnerUserID)
		put("review_owner_user_id", value.ReviewOwnerUserID)
		put("owner_user_id", value.OwnerUserID)
		put("created_by_user_id", value.CreatedByUserID)
	case *Finding:
		put("owner_user_id", &value.OwnerUserID)
	case *Query:
		put("created_by_user_id", &value.CreatedByUserID)
	case *Risk:
		put("created_by_user_id", &value.CreatedByUserID)
		put("deleted_by_user_id", value.DeletedByUserID)
	}
	return actors
}

// WithLocalActor returns a new typed row; prepared historical actors stay intact.
func WithLocalActor(row any, actor uuid.UUID) any {
	local := func(prior *uuid.UUID) *uuid.UUID {
		if prior == nil {
			return nil
		}
		value := actor
		return &value
	}
	switch value := row.(type) {
	case *Parent:
		copy := *value
		copy.OutgoingOwnerUserID = local(value.OutgoingOwnerUserID)
		copy.IncomingOwnerUserID = local(value.IncomingOwnerUserID)
		copy.ReviewOwnerUserID = local(value.ReviewOwnerUserID)
		copy.OwnerUserID = local(value.OwnerUserID)
		copy.CreatedByUserID = local(value.CreatedByUserID)
		return &copy
	case *Finding:
		copy := *value
		copy.OwnerUserID = actor
		return &copy
	case *Query:
		copy := *value
		copy.CreatedByUserID = actor
		return &copy
	case *Keyword:
		copy := *value
		return &copy
	case *Risk:
		copy := *value
		copy.CreatedByUserID = actor
		copy.DeletedByUserID = local(value.DeletedByUserID)
		return &copy
	default:
		return nil
	}
}

func SubtypeLegal(row any) bool {
	switch value := row.(type) {
	case *Parent:
		for _, binding := range []struct {
			values   []*string
			contract string
		}{
			{[]*string{value.Title, value.ChannelOrMeeting, value.Summary, value.PrivilegeTag}, "single_line_title_v1"},
			{[]*string{value.Body, value.CurrentStateSummary, value.NextChecks, value.ActiveRisksSummary}, "multiline_body_v1"},
			{[]*string{value.Audience}, "party_text_v1"},
		} {
			for _, text := range binding.values {
				if text != nil && !exactText(*text, binding.contract) {
					return false
				}
			}
		}
		for _, id := range []*string{value.CommID, value.HandoffID, value.StatusReviewID, value.LessonID} {
			if id != nil && !opaqueID(*id) {
				return false
			}
		}
		if value.CommType != nil && !token(*value.CommType, "meeting", "notification", "approval", "briefing", "handoff") {
			return false
		}
		if value.ClosureState != nil && !token(*value.ClosureState, "open", "closed") {
			return false
		}
		switch value.ArtifactType {
		case "note", "finding", "investigative_query", "forensic_keyword":
			return true
		case "comm_log":
			return value.CommID != nil && value.CommType != nil && value.TimestampUTC != nil && value.Audience != nil && value.ChannelOrMeeting != nil && value.Summary != nil
		case "handoff":
			return value.HandoffID != nil && value.TimestampUTC != nil && value.OutgoingOwnerUserID != nil && value.IncomingOwnerUserID != nil && value.CurrentStateSummary != nil
		case "status_review":
			return value.StatusReviewID != nil && value.TimestampUTC != nil && value.ReviewOwnerUserID != nil && value.CurrentStateSummary != nil
		case "lesson":
			return value.LessonID != nil && value.TimestampUTC != nil && value.Summary != nil && value.OwnerUserID != nil
		default:
			return false
		}
	case *Finding:
		return token(value.Kind, "finding", "hypothesis") && token(value.State, "open", "closed") && exactText(value.Statement, "multiline_body_v1") && (value.ConfidenceScore == nil || (*value.ConfidenceScore >= 0 && *value.ConfidenceScore <= 100))
	case *Query:
		return opaqueID(value.QueryID) && exactText(value.Platform, "single_line_title_v1") && exactText(value.Purpose, "single_line_title_v1") && exactText(value.QueryText, "multiline_body_v1")
	case *Keyword:
		return opaqueID(value.KeywordID) && exactText(value.Pattern, "single_line_title_v1") && exactText(value.Reason, "reason_note_v1") && token(value.MatchMode, "literal", "regex")
	case *Risk:
		return exactText(value.RiskRefText, "single_line_title_v1")
	default:
		return false
	}
}

func LifecycleLegal(row any) bool {
	switch value := row.(type) {
	case *Parent:
		return !value.UpdatedAt.Before(value.CreatedAt)
	case *Finding:
		return !value.UpdatedAt.Before(value.CreatedAt) && (value.State == "closed") == (value.ClosedAt != nil) && (value.ClosedAt == nil || (!value.ClosedAt.Before(value.CreatedAt) && !value.ClosedAt.After(value.UpdatedAt)))
	case *Query:
		return !value.UpdatedAt.Before(value.CreatedAt)
	case *Keyword:
		return !value.UpdatedAt.Before(value.CreatedAt)
	case *Risk:
		return (value.DeletedAt == nil) == (value.DeletedByUserID == nil) && (value.DeletedAt == nil || !value.DeletedAt.Before(value.CreatedAt))
	default:
		return false
	}
}

// UTCClocks makes Go location metadata consistent for already-admitted UTC
// instants. It changes neither an instant nor any retained source bytes.
func UTCClocks(row any) {
	pointer := func(value **time.Time) {
		if *value != nil {
			instant := (*value).UTC()
			*value = &instant
		}
	}
	switch value := row.(type) {
	case *Parent:
		value.CreatedAt, value.UpdatedAt = value.CreatedAt.UTC(), value.UpdatedAt.UTC()
		pointer(&value.TimestampUTC)
		pointer(&value.NextReportAt)
		pointer(&value.AcknowledgedAt)
	case *Finding:
		value.CreatedAt, value.UpdatedAt = value.CreatedAt.UTC(), value.UpdatedAt.UTC()
		pointer(&value.ClosedAt)
	case *Query:
		value.CreatedAt, value.UpdatedAt = value.CreatedAt.UTC(), value.UpdatedAt.UTC()
	case *Keyword:
		value.CreatedAt, value.UpdatedAt = value.CreatedAt.UTC(), value.UpdatedAt.UTC()
	case *Risk:
		value.CreatedAt = value.CreatedAt.UTC()
		pointer(&value.DeletedAt)
	}
}

func exactText(value, contract string) bool {
	normalized, valid := textpolicy.Normalize(value, contract)
	return valid && value != "" && normalized == value
}

func opaqueID(value string) bool {
	return strings.TrimSpace(value) != "" && !strings.ContainsRune(value, 0)
}

func token(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
