package storage

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
)

func TestArtifactPhysicalRowBindingPreservesClosedVariantMatrix(t *testing.T) {
	t.Parallel()
	if got := reflect.TypeOf(artifactPhysicalRow{}).NumField(); got != 55 {
		t.Fatalf("Artifact physical projection fields = %d, want 55", got)
	}

	now := time.Date(2026, 8, 27, 18, 0, 0, 0, time.FixedZone("source", -4*60*60))
	nowUTC := now.UTC()
	nextReportAt := now.Add(2 * time.Hour)
	acknowledgedAt := now.Add(time.Hour)
	closedAt := now.Add(3 * time.Hour)
	day := "2026-08-27"
	nextDay := "2026-08-28"
	title := "Projection title"
	body := "Projection body"
	privilege := "internal"
	nextChecks := "Verify recovery"
	risks := "Persistence risk"
	confidenceScore := 83
	actorID := uuid.New()
	incomingOwnerID := uuid.New()
	envelope, err := artifactprojection.NewProjectionEnvelope(
		uuid.New(), uuid.New(), 4, &title, &body, &now, now.Add(time.Minute), now, &actorID, &day, 2,
	)
	if err != nil {
		t.Fatalf("construct projection envelope: %v", err)
	}
	base := artifactPhysicalRow{
		RecordID: envelope.RecordID(), IncidentID: envelope.IncidentID(), RowVersion: 4,
		Title: &title, Body: &body, TimestampUTC: &nowUTC, UpdatedAt: now.Add(time.Minute).UTC(),
		CreatedAt: nowUTC, CreatedByUserID: &actorID, FindingUpdatedAt: storageTimeValuePointer(now.Add(time.Minute)),
		FindingConfidenceBand: storageStringPointer("unset"), TimestampDay: &day,
		AckState: "pending", LinkedRecordCount: 2,
	}

	tests := []struct {
		name      string
		construct func() (artifactprojection.ProjectionInput, error)
		expected  func() artifactPhysicalRow
	}{
		{name: "note", construct: func() (artifactprojection.ProjectionInput, error) {
			return artifactprojection.NewNoteProjectionInput(envelope, artifactprojection.NoteVariant{})
		}, expected: func() artifactPhysicalRow {
			row := base
			row.ArtifactType = "note"
			return row
		}},
		{name: "communication log", construct: func() (artifactprojection.ProjectionInput, error) {
			return artifactprojection.NewCommunicationLogProjectionInput(envelope, artifactprojection.CommunicationLogVariant{
				CommID: "comm-1", CommType: "briefing", Audience: "responders", ChannelOrMeeting: "bridge",
				Summary: "Situation update", NextReportAt: &nextReportAt, PrivilegeTag: &privilege, NextReportDay: &nextDay,
			})
		}, expected: func() artifactPhysicalRow {
			row := base
			row.ArtifactType = "comm_log"
			row.CommID = storageStringPointer("comm-1")
			row.CommType = storageStringPointer("briefing")
			row.Audience = storageStringPointer("responders")
			row.ChannelOrMeeting = storageStringPointer("bridge")
			row.Summary = storageStringPointer("Situation update")
			row.NextReportAt = storageTimeValuePointer(nextReportAt)
			row.PrivilegeTag = &privilege
			row.NextReportDay = &nextDay
			return row
		}},
		{name: "handoff", construct: func() (artifactprojection.ProjectionInput, error) {
			return artifactprojection.NewHandoffProjectionInput(envelope, artifactprojection.HandoffVariant{
				HandoffID: "handoff-1", OutgoingOwnerUserID: actorID, IncomingOwnerUserID: incomingOwnerID,
				CurrentStateSummary: "Stable", NextChecks: &nextChecks, AcknowledgedAt: &acknowledgedAt, AckState: "acknowledged",
			})
		}, expected: func() artifactPhysicalRow {
			row := base
			row.ArtifactType = "handoff"
			row.HandoffID = storageStringPointer("handoff-1")
			row.OutgoingOwnerUserID = &actorID
			row.IncomingOwnerUserID = &incomingOwnerID
			row.CurrentStateSummary = storageStringPointer("Stable")
			row.NextChecks = &nextChecks
			row.AcknowledgedAt = storageTimeValuePointer(acknowledgedAt)
			row.AckState = "acknowledged"
			return row
		}},
		{name: "status review", construct: func() (artifactprojection.ProjectionInput, error) {
			return artifactprojection.NewStatusReviewProjectionInput(envelope, artifactprojection.StatusReviewVariant{
				StatusReviewID: "review-1", ReviewOwnerUserID: actorID, CurrentStateSummary: "Recovering",
				ActiveRisksSummary: &risks, NextReportAt: &nextReportAt, NextReportDay: &nextDay,
			})
		}, expected: func() artifactPhysicalRow {
			row := base
			row.ArtifactType = "status_review"
			row.StatusReviewID = storageStringPointer("review-1")
			row.ReviewOwnerUserID = &actorID
			row.CurrentStateSummary = storageStringPointer("Recovering")
			row.ActiveRisksSummary = &risks
			row.NextReportAt = storageTimeValuePointer(nextReportAt)
			row.NextReportDay = &nextDay
			return row
		}},
		{name: "lesson", construct: func() (artifactprojection.ProjectionInput, error) {
			return artifactprojection.NewLessonProjectionInput(envelope, artifactprojection.LessonVariant{
				LessonID: "lesson-1", Summary: "Capture early", OwnerUserID: actorID, ClosureState: "open",
			})
		}, expected: func() artifactPhysicalRow {
			row := base
			row.ArtifactType = "lesson"
			row.LessonID = storageStringPointer("lesson-1")
			row.Summary = storageStringPointer("Capture early")
			row.OwnerUserID = &actorID
			row.ClosureState = storageStringPointer("open")
			return row
		}},
		{name: "finding", construct: func() (artifactprojection.ProjectionInput, error) {
			return artifactprojection.NewFindingProjectionInput(envelope, artifactprojection.FindingVariant{
				Statement: "Evidence confirms", Kind: "finding", State: "closed", OwnerUserID: actorID,
				ConfidenceScore: &confidenceScore, ClosedAt: &closedAt, UpdatedAt: now, ConfidenceBand: "high",
			})
		}, expected: func() artifactPhysicalRow {
			row := base
			row.ArtifactType = "finding"
			row.FindingStatement = storageStringPointer("Evidence confirms")
			row.FindingKind = storageStringPointer("finding")
			row.FindingState = storageStringPointer("closed")
			row.FindingOwnerUserID = &actorID
			row.FindingConfidenceScore = &confidenceScore
			row.FindingClosedAt = storageTimeValuePointer(closedAt)
			row.FindingUpdatedAt = storageTimeValuePointer(now)
			row.FindingConfidenceBand = storageStringPointer("high")
			return row
		}},
		{name: "investigative query", construct: func() (artifactprojection.ProjectionInput, error) {
			return artifactprojection.NewInvestigativeQueryProjectionInput(envelope, artifactprojection.InvestigativeQueryVariant{
				QueryID: "query-1", Platform: "edr", Purpose: "scope", QueryText: "search",
				CreatedByUserID: actorID, CreatedAt: now, CreatedDay: day,
			})
		}, expected: func() artifactPhysicalRow {
			row := base
			row.ArtifactType = "investigative_query"
			row.InvestigativeQueryQueryID = storageStringPointer("query-1")
			row.InvestigativeQueryPlatform = storageStringPointer("edr")
			row.InvestigativeQueryPurpose = storageStringPointer("scope")
			row.InvestigativeQueryQueryText = storageStringPointer("search")
			row.InvestigativeQueryCreatedBy = &actorID
			row.InvestigativeQueryCreatedAt = storageTimeValuePointer(now)
			row.InvestigativeQueryCreatedDay = &day
			return row
		}},
		{name: "forensic keyword", construct: func() (artifactprojection.ProjectionInput, error) {
			return artifactprojection.NewForensicKeywordProjectionInput(envelope, artifactprojection.ForensicKeywordVariant{
				KeywordID: "keyword-1", Pattern: "pattern", Reason: "coverage", MatchMode: "regex",
				CaseSensitive: true, CreatedAt: now, CreatedDay: day,
			})
		}, expected: func() artifactPhysicalRow {
			row := base
			row.ArtifactType = "forensic_keyword"
			row.ForensicKeywordKeywordID = storageStringPointer("keyword-1")
			row.ForensicKeywordPattern = storageStringPointer("pattern")
			row.ForensicKeywordReason = storageStringPointer("coverage")
			row.ForensicKeywordMatchMode = storageStringPointer("regex")
			row.ForensicKeywordCaseSensitive = storageBoolPointer(true)
			row.ForensicKeywordCreatedAt = storageTimeValuePointer(now)
			row.ForensicKeywordCreatedDay = &day
			return row
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input, err := tc.construct()
			if err != nil {
				t.Fatalf("construct typed source input: %v", err)
			}
			got, err := artifactPhysicalRowFromInput(input)
			if err != nil {
				t.Fatalf("bind physical row: %v", err)
			}
			if want := tc.expected(); !reflect.DeepEqual(got, want) {
				t.Fatalf("55-column binding mismatch\n got: %#v\nwant: %#v", got, want)
			}
		})
	}

	if _, err := artifactPhysicalRowFromInput(artifactprojection.ProjectionInput{}); err == nil {
		t.Fatal("zero typed source input produced a physical row")
	}
}
