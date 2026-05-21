package viewschema

// ArtifactVariant records the closed current-profile artifact-backed variant
// family owned by Core 02 section 10.4.4A.
type ArtifactVariant struct {
	ArtifactVariantID         string
	DurableDiscriminatorKind  string
	DurableDiscriminatorField string
	DurableDiscriminatorValue string
	SubkindDimension          string
	PublicSurfaceRef          string
	SurfaceStatus             string
	IdentifierField           string
	RequiredStructuredState   []string
	OptionalStructuredState   []string
	LifecycleNotes            string
}

var artifactVariants = []ArtifactVariant{
	{
		ArtifactVariantID:         "note",
		DurableDiscriminatorKind:  "artifact_type",
		DurableDiscriminatorField: "artifact_type",
		DurableDiscriminatorValue: "note",
		PublicSurfaceRef:          "cartulary.view.notes.v1",
		SurfaceStatus:             "required built-in sheet",
		IdentifierField:           "record_id",
		RequiredStructuredState:   []string{"title", "body"},
		OptionalStructuredState:   []string{"tags via record_tags"},
		LifecycleNotes:            "shared artifact lifecycle only; Notes-sheet create and add-linked-note create use the same underlying artifact shape",
	},
	{
		ArtifactVariantID:         "comm_log",
		DurableDiscriminatorKind:  "artifact_type",
		DurableDiscriminatorField: "artifact_type",
		DurableDiscriminatorValue: "comm_log",
		PublicSurfaceRef:          "cartulary.view.comm_log.v1",
		SurfaceStatus:             "required workbook-native coordination surface",
		IdentifierField:           "comm_id",
		RequiredStructuredState:   []string{"comm_id", "comm_type", "timestamp_utc", "audience", "channel_or_meeting", "summary"},
		OptionalStructuredState:   []string{"decision_ids[]", "action_task_ids[]", "next_report_at", "privilege_tag", "audience_party_ids[]", "attendee_party_ids[]"},
		LifecycleNotes:            "audience text remains required source-preserving text even when supplemental party refs are present",
	},
	{
		ArtifactVariantID:         "handoff",
		DurableDiscriminatorKind:  "artifact_type",
		DurableDiscriminatorField: "artifact_type",
		DurableDiscriminatorValue: "handoff",
		PublicSurfaceRef:          "cartulary.view.handoff.v1",
		SurfaceStatus:             "required workbook-native coordination surface",
		IdentifierField:           "handoff_id",
		RequiredStructuredState:   []string{"handoff_id", "timestamp_utc", "outgoing_owner_user_id", "incoming_owner_user_id", "current_state_summary"},
		OptionalStructuredState:   []string{"open_task_ids[]", "open_decision_ids[]", "open_risk_refs[]", "next_checks", "acknowledged_at"},
		LifecycleNotes:            "derived ack_state uses exact tokens pending|acknowledged; risk refs are child rows rather than first-class risk records",
	},
	{
		ArtifactVariantID:         "status_review",
		DurableDiscriminatorKind:  "artifact_type",
		DurableDiscriminatorField: "artifact_type",
		DurableDiscriminatorValue: "status_review",
		PublicSurfaceRef:          "cartulary.view.status_review.v1",
		SurfaceStatus:             "required workbook-native coordination surface",
		IdentifierField:           "status_review_id",
		RequiredStructuredState:   []string{"status_review_id", "timestamp_utc", "review_owner_user_id", "current_state_summary"},
		OptionalStructuredState:   []string{"blocked_task_ids[]", "pending_evidence_ids[]", "open_decision_ids[]", "active_risks_summary", "next_report_at"},
		LifecycleNotes:            "coordination artifact only; no separate subtype lifecycle machine beyond ordinary artifact lifecycle",
	},
	{
		ArtifactVariantID:         "lesson",
		DurableDiscriminatorKind:  "artifact_type",
		DurableDiscriminatorField: "artifact_type",
		DurableDiscriminatorValue: "lesson",
		PublicSurfaceRef:          "cartulary.view.lesson.v1",
		SurfaceStatus:             "required workbook-native coordination surface",
		IdentifierField:           "lesson_id",
		RequiredStructuredState:   []string{"lesson_id", "timestamp_utc", "summary", "owner_user_id"},
		OptionalStructuredState:   []string{"follow_up_task_ids[]", "closure_state", "evidence_refs[]"},
		LifecycleNotes:            "closure_state uses the exact closed vocabulary defined in Core 02 section 18; lessons remain artifact-backed and reuse shared history and links",
	},
	{
		ArtifactVariantID:         "finding",
		DurableDiscriminatorKind:  "artifact_type",
		DurableDiscriminatorField: "artifact_type",
		DurableDiscriminatorValue: "finding",
		SubkindDimension:          "finding.kind",
		PublicSurfaceRef:          "cartulary.view.findings.v1",
		SurfaceStatus:             "standardized optional workbook surface",
		IdentifierField:           "record_id",
		RequiredStructuredState:   []string{"finding.kind", "statement", "state", "confidence_score", "owner_user_id"},
		OptionalStructuredState:   []string{"closed_at", "supporting_refs[]", "contradictory_refs[]"},
		LifecycleNotes:            "this is the only current-profile row that covers both findings and hypotheses; finding.kind is required structured state; finding.state uses the exact open|closed vocabulary defined in Core 02 section 18; closed_at is server-managed",
	},
}

func ListArtifactVariants() []ArtifactVariant {
	variants := make([]ArtifactVariant, len(artifactVariants))
	for index, variant := range artifactVariants {
		variants[index] = cloneArtifactVariant(variant)
	}
	return variants
}

func LookupArtifactVariant(artifactVariantID string) (ArtifactVariant, bool) {
	for _, variant := range artifactVariants {
		if variant.ArtifactVariantID == artifactVariantID {
			return cloneArtifactVariant(variant), true
		}
	}
	return ArtifactVariant{}, false
}

func LookupArtifactVariantByArtifactType(artifactType string) (ArtifactVariant, bool) {
	for _, variant := range artifactVariants {
		if variant.DurableDiscriminatorKind == "artifact_type" &&
			variant.DurableDiscriminatorField == "artifact_type" &&
			variant.DurableDiscriminatorValue == artifactType {
			return cloneArtifactVariant(variant), true
		}
	}
	return ArtifactVariant{}, false
}

func cloneArtifactVariant(variant ArtifactVariant) ArtifactVariant {
	variant.RequiredStructuredState = cloneStrings(variant.RequiredStructuredState)
	variant.OptionalStructuredState = cloneStrings(variant.OptionalStructuredState)
	return variant
}
