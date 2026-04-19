package entities_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/assertx"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

// I-4-01 / REQ-01-196..REQ-01-227, REQ-02-039..REQ-02-044 / AC-188..AC-190, AC-221..AC-225.
func TestPhase4_ResolveRoute_I_4_01_Red(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-i-4-01")
	phase4test.RequireRouteSurface(
		t,
		"I-4-01",
		harness.Server,
		http.MethodPost,
		"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
		fixtures.MentionResolveRoutePayload(7, "txn-phase4-i-4-01", golden.Phase4MentionActionResolve, uuidPointer(golden.Phase4CanonicalHostRecordID), nil),
	)
}

// I-4-02 / REQ-02-035..REQ-02-036, REQ-02-054..REQ-02-055, REQ-02-059..REQ-02-063 / AC-022, AC-186.
func TestPhase4_EntityOriginUpsert_I_4_02_Red(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-i-4-02")
	resp := phase4test.RequireRouteSurface(
		t,
		"I-4-02",
		harness.Server,
		http.MethodPost,
		"/api/v1/incidents/"+golden.Phase4IncidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
		fixtures.HostCreatePayload("txn-phase4-i-4-02"),
	)
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] == "unknown_view_schema" {
		t.Fatalf("Phase 4 I-4-02 expected Hosts entity_origin surface %s to be active, got reason_code=%v", golden.Phase4HostsViewSchemaID, details["reason_code"])
	}
}

// I-4-03 / REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
func TestPhase4_ExplicitMergeRoute_I_4_03_Red(t *testing.T) {
	t.Run("host merge repoints live fan-out and preserves survivor reuse", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-03")
		phase4test.RequireSchemaTables(t, harness.DB, "I-4-03", "hosts", "identities", "entity_mentions", "record_tags", "compromise_assessments")

		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-03-incident",
			"incident_key":  "IR-I403",
			"title":         "Entity merge",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		hubChanges, unsubscribe := harness.Server.Runtime.WSHub.SubscribeRecordChanges(16)
		defer unsubscribe()

		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "ws-023.corp.example.test", "")
		seedEntityAlias(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateHostRecordID, "host", "Workstation 23")
		seedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
		seedResolvedMention(t, harness.DB, adminUserID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateHostRecordID, golden.Phase4FieldTimelineHostRefs, "WS-023")
		seedRecordLink(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateLinkID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateHostRecordID, "observed_on_host", "manual", nil)
		seedRecordTag(t, harness.DB, incidentID, adminUserID, golden.Phase4TagIDSurvivor, golden.Phase4CanonicalHostRecordID, "critical-host")
		seedRecordTag(t, harness.DB, incidentID, adminUserID, golden.Phase4TagIDLoser, golden.Phase4DuplicateHostRecordID, "critical-host")
		seedAssessment(t, harness.DB, incidentID, adminUserID, golden.Phase4AssessmentHostID, golden.Phase4DuplicateHostRecordID, "host", "confirmed")

		mergeResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.Phase4CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.Phase4DuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-phase4-i-4-03-merge",
				"reason":                    "  merge duplicate host  ",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		if mergeResp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: got %d want %d body=%#v", mergeResp.StatusCode, http.StatusOK, httptestx.ReadJSONBody(t, mergeResp))
		}
		mergeData := httptestx.RequireSuccessEnvelope(t, mergeResp, http.StatusOK)["data"].(map[string]any)
		if mergeData["survivor_record_id"] != golden.Phase4CanonicalHostRecordID.String() {
			t.Fatalf("unexpected survivor_record_id: %#v", mergeData)
		}
		if mergeData["loser_record_id"] != golden.Phase4DuplicateHostRecordID.String() {
			t.Fatalf("unexpected loser_record_id: %#v", mergeData)
		}
		if got := int64(mergeData["survivor_row_version"].(float64)); got != 2 {
			t.Fatalf("expected survivor_row_version=2, got %d", got)
		}
		if got := int64(mergeData["loser_row_version"].(float64)); got != 2 {
			t.Fatalf("expected loser_row_version=2, got %d", got)
		}
		if mergeData["merged_into_record_id"] != golden.Phase4CanonicalHostRecordID.String() {
			t.Fatalf("expected merged_into_record_id to echo survivor, got %#v", mergeData)
		}

		summary := mergeData["merge_summary"].(map[string]any)
		if summary["record_type"] != "host" {
			t.Fatalf("unexpected merge summary record_type: %#v", summary)
		}
		if got := int(summary["repointed_mention_resolution_count"].(float64)); got != 1 {
			t.Fatalf("expected one repointed mention resolution, got %d", got)
		}
		if got := int(summary["repointed_link_count"].(float64)); got != 1 {
			t.Fatalf("expected one repointed link, got %d", got)
		}
		if got := int(summary["deduped_tag_count"].(float64)); got != 1 {
			t.Fatalf("expected one deduped tag, got %d", got)
		}
		exactMatchClasses := summary["exact_match_classes"].([]any)
		if len(exactMatchClasses) != 3 {
			t.Fatalf("expected three host exact-match classes, got %#v", exactMatchClasses)
		}
		if exactMatchClasses[0].(map[string]any)["identifier_class"] != "aad_device_id" {
			t.Fatalf("unexpected host exact-match precedence: %#v", exactMatchClasses)
		}
		if exactMatchClasses[1].(map[string]any)["identifier_class"] != "fqdn" {
			t.Fatalf("unexpected host exact-match precedence: %#v", exactMatchClasses)
		}
		if got := int(exactMatchClasses[1].(map[string]any)["promoted_count"].(float64)); got != 1 {
			t.Fatalf("expected fqdn promoted_count=1, got %#v", exactMatchClasses[1])
		}

		survivorState, survivorMergedInto, survivorRowVersion, survivorFQDN := lookupHostState(t, harness.DB, golden.Phase4CanonicalHostRecordID)
		if survivorState != "canonical" || survivorMergedInto != nil || survivorRowVersion != 2 || survivorFQDN != "ws-023.corp.example.test" {
			t.Fatalf("unexpected survivor host state after merge: state=%s merged_into=%v row_version=%d fqdn=%q", survivorState, survivorMergedInto, survivorRowVersion, survivorFQDN)
		}
		loserState, loserMergedInto, loserRowVersion, _ := lookupHostState(t, harness.DB, golden.Phase4DuplicateHostRecordID)
		if loserState != "merged" || loserMergedInto == nil || *loserMergedInto != golden.Phase4CanonicalHostRecordID || loserRowVersion != 2 {
			t.Fatalf("unexpected loser host state after merge: state=%s merged_into=%v row_version=%d", loserState, loserMergedInto, loserRowVersion)
		}

		mention := lookupMention(t, harness.DB, golden.Phase4HostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected merge to repoint mention to survivor, got %#v", mention)
		}
		if mention.RowVersion != 2 {
			t.Fatalf("expected merge to increment mention row_version, got %#v", mention)
		}

		link := lookupActiveLink(t, harness.DB, incidentID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(t, link, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host", "manual", nil)
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, golden.Phase4DuplicateLinkID); got != 0 {
			t.Fatalf("expected loser-targeted active link to disappear, got %d active rows", got)
		}

		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'critical-host'
   AND deleted_at IS NULL
`, incidentID, golden.Phase4CanonicalHostRecordID); got != 1 {
			t.Fatalf("expected one active survivor tag after dedupe, got %d", got)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE record_id = $1
   AND deleted_at IS NULL
`, golden.Phase4DuplicateHostRecordID); got != 0 {
			t.Fatalf("expected loser active tags to be cleared, got %d", got)
		}
		if got := lookupAssessmentSubject(t, harness.DB, golden.Phase4AssessmentHostID); got != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected loser assessment to repoint to survivor, got %s", got)
		}

		createResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-phase4-i-4-03-create-after-merge",
				"host.fqdn":     "ws-023.corp.example.test",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusOK)["data"].(map[string]any)
		row := createData["row"].(map[string]any)
		if row["record_id"] != golden.Phase4CanonicalHostRecordID.String() {
			t.Fatalf("expected carried-forward exact match to reuse survivor, got %#v", createData)
		}

		changes := collectRecordChanges(t, hubChanges, 3, 5*time.Second)
		requireRecordChange(t, changes, golden.Phase4CanonicalHostRecordID, golden.Phase4HostsViewSchemaID)
		requireRecordChange(t, changes, golden.Phase4DuplicateHostRecordID, golden.Phase4HostsViewSchemaID)
		requireRecordChange(t, changes, golden.Phase4TimelineRecordID, golden.Phase4TimelineViewSchemaID)
		_ = link
	})
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

type indicatorRecordRow struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DedupeKey       string
	DefangedValue   *string
	HashAlgorithm   *string
	HashValue       *string
	STIXPattern     *string
	RowVersion      int64
	CreatedByUser   uuid.UUID
	UpdatedByUser   uuid.UUID
}

type indicatorProjectionRow struct {
	RecordID            uuid.UUID
	RowVersion          int64
	IndicatorType       string
	ValueKind           string
	DisplayValue        string
	NormalizedValue     *string
	DefangedValue       *string
	HashAlgorithm       *string
	HashValue           *string
	STIXPattern         *string
	FirstObservedAt     *time.Time
	LastObservedAt      *time.Time
	ObservationCount    int
	LifecycleSummary    *string
	SupportingLinkCount int
}

type indicatorObservationRow struct {
	ObservationID             uuid.UUID
	SourceRecordID            uuid.UUID
	SourceFieldKey            string
	OriginKind                string
	OriginLocator             string
	ObservedText              string
	ParsedIndicatorType       *string
	NormalizedCandidate       *string
	ResolutionStatus          string
	ResolvedIndicatorRecordID *uuid.UUID
	RowVersion                int64
}

type indicatorLifecycleIntervalRow struct {
	IntervalID     uuid.UUID
	IndicatorID    uuid.UUID
	LifecycleState string
	ValidFrom      time.Time
	ValidTo        *time.Time
}

func provisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (loginResult, uuid.UUID) {
	t.Helper()

	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-phase4-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	completeInitialEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-phase4-bootstrap-admin-complete")
	login := loginLocalUserWithSecondFactor(t, server, "bootstrap-admin@example.test", "BootstrapPass1!", generateTOTPCode(t, secretBase32))

	sessionResp := doEntitiesJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(login.sessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, mustUUID(t, sessionData["user_id"].(string))
}

func createIncident(t testing.TB, server *httptestx.Server, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := doEntitiesJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := doEntitiesJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	return details["bootstrap_token"].(string)
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := doEntitiesJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, withHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := doEntitiesJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func loginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) loginResult {
	t.Helper()

	resp := doEntitiesJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
		"second_factor": map[string]any{
			"kind": "totp",
			"assertion": map[string]any{
				"code": code,
			},
		},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login with second factor failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return loginResult{sessionCookie: sessionCookie, csrfCookie: csrfCookie}
}

func generateTOTPCode(t testing.TB, secretBase32 string) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(secretBase32, time.Now().UTC(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	return code
}

func doEntitiesJSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func withHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func seedHostRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, hostname string, fqdn string, aadDeviceID string) {
	t.Helper()

	var (
		fqdnValue      any
		aadDeviceValue any
	)
	if fqdn != "" {
		fqdnValue = fqdn
	}
	if aadDeviceID != "" {
		aadDeviceValue = aadDeviceID
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, fqdn, aad_device_id, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, 'canonical', $7, $7)
`, recordID, incidentID, displayName, hostname, fqdnValue, aadDeviceValue, actorUserID); err != nil {
		t.Fatalf("seed host record: %v", err)
	}
}

func seedEntityAlias(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, entityType string, rawText string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO entity_aliases (incident_id, record_id, entity_type, raw_text, normalized_text, classification, created_by_user_id, created_at)
VALUES ($1, $2, $3, $4, $5, 'suggestion_only', $6, now())
`, incidentID, recordID, entityType, rawText, rawText, actorUserID); err != nil {
		t.Fatalf("seed entity alias: %v", err)
	}
}

func seedTimelineRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO timeline_events (record_id, incident_id, summary, capture_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'merge-source-row', 'reviewed', $3, $3)
`, recordID, incidentID, actorUserID); err != nil {
		t.Fatalf("seed timeline record: %v", err)
	}
}

func seedResolvedMention(t testing.TB, db *sql.DB, actorUserID uuid.UUID, mentionID uuid.UUID, sourceRecordID uuid.UUID, resolvedRecordID uuid.UUID, sourceFieldKey string, rawText string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO entity_mentions (
    entity_mention_id,
    source_record_id,
    entity_type,
    source_field_key,
    origin_kind,
    origin_locator,
    raw_text,
    normalized_text,
    resolution_status,
    row_version,
    ordinal,
    created_by_user_id,
    resolved_record_id,
    resolved_by_user_id,
    resolved_at,
    resolution_method
)
VALUES ($1, $2, 'host', $3, 'interactive_cell', 'merge-test', $4, $4, 'resolved', 1, 1, $5, $6, $5, now(), 'explicit_resolve_route')
`, mentionID, sourceRecordID, sourceFieldKey, rawText, actorUserID, resolvedRecordID); err != nil {
		t.Fatalf("seed resolved mention: %v", err)
	}
}

func seedRecordLink(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordLinkID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string, provenance string, confidence *int) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_links (
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    decided_at,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
`, recordLinkID, incidentID, srcRecordID, dstRecordID, linkType, provenance, confidence, actorUserID); err != nil {
		t.Fatalf("seed record link: %v", err)
	}
}

func seedRecordTag(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordTagID uuid.UUID, recordID uuid.UUID, tagName string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6)
`, recordTagID, incidentID, recordID, tagName, tagName, actorUserID); err != nil {
		t.Fatalf("seed record tag: %v", err)
	}
}

func seedAssessment(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, assessmentID uuid.UUID, subjectID uuid.UUID, subjectType string, state string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO compromise_assessments (compromise_assessment_id, incident_id, subject_id, subject_type, state, assessed_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6)
`, assessmentID, incidentID, subjectID, subjectType, state, actorUserID); err != nil {
		t.Fatalf("seed assessment: %v", err)
	}
}

func lookupHostState(t testing.TB, db *sql.DB, recordID uuid.UUID) (string, *uuid.UUID, int64, string) {
	t.Helper()

	var (
		state         string
		mergedIntoRaw sql.NullString
		rowVersion    int64
		fqdn          sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT host_state, merged_into_record_id::text, row_version, COALESCE(fqdn, '')
  FROM hosts
 WHERE record_id = $1
`, recordID).Scan(&state, &mergedIntoRaw, &rowVersion, &fqdn); err != nil {
		t.Fatalf("lookup host state: %v", err)
	}
	var mergedInto *uuid.UUID
	if mergedIntoRaw.Valid {
		value := mustUUID(t, mergedIntoRaw.String)
		mergedInto = &value
	}
	return state, mergedInto, rowVersion, fqdn.String
}

func lookupMention(t testing.TB, db *sql.DB, mentionID uuid.UUID) fixtures.EntityMentionFixture {
	t.Helper()

	var mention fixtures.EntityMentionFixture
	var (
		sourceRecordID   string
		resolvedRecordID sql.NullString
		resolvedByUserID sql.NullString
		resolvedAt       sql.NullTime
		resolutionMethod sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT source_record_id::text, raw_text, resolution_status, row_version, resolved_record_id::text, resolved_by_user_id::text, resolved_at, resolution_method
  FROM entity_mentions
 WHERE entity_mention_id = $1
`, mentionID).Scan(
		&sourceRecordID,
		&mention.RawText,
		&mention.ResolutionStatus,
		&mention.RowVersion,
		&resolvedRecordID,
		&resolvedByUserID,
		&resolvedAt,
		&resolutionMethod,
	); err != nil {
		t.Fatalf("lookup mention: %v", err)
	}
	mention.EntityMentionID = mentionID
	mention.SourceRecordID = mustUUID(t, sourceRecordID)
	if resolvedRecordID.Valid {
		value := mustUUID(t, resolvedRecordID.String)
		mention.ResolvedRecordID = &value
	}
	if resolvedByUserID.Valid {
		value := mustUUID(t, resolvedByUserID.String)
		mention.ResolvedByUserID = &value
	}
	if resolvedAt.Valid {
		value := resolvedAt.Time.UTC()
		mention.ResolvedAt = &value
	}
	if resolutionMethod.Valid {
		value := resolutionMethod.String
		mention.ResolutionMethod = &value
	}
	return mention
}

func lookupActiveLink(t testing.TB, db *sql.DB, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, linkType string) fixtures.LinkFixture {
	t.Helper()

	var (
		link        fixtures.LinkFixture
		confidence  sql.NullInt64
		deletedAt   sql.NullTime
		recordLink  string
		incidentRaw string
		sourceRaw   string
		targetRaw   string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT record_link_id::text, incident_id::text, src_record_id::text, dst_record_id::text, link_type, provenance, confidence, deleted_at
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
   AND deleted_at IS NULL
`, incidentID, sourceID, targetID, linkType).Scan(&recordLink, &incidentRaw, &sourceRaw, &targetRaw, &link.LinkType, &link.Provenance, &confidence, &deletedAt); err != nil {
		t.Fatalf("lookup active link: %v", err)
	}
	link.RecordLinkID = mustUUID(t, recordLink)
	link.IncidentID = mustUUID(t, incidentRaw)
	link.SourceID = mustUUID(t, sourceRaw)
	link.TargetID = mustUUID(t, targetRaw)
	if confidence.Valid {
		value := int(confidence.Int64)
		link.Confidence = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		link.DeletedAt = &value
	}
	return link
}

func lookupAssessmentSubject(t testing.TB, db *sql.DB, assessmentID uuid.UUID) uuid.UUID {
	t.Helper()

	var subjectID string
	if err := db.QueryRowContext(context.Background(), `
SELECT subject_id::text
  FROM compromise_assessments
 WHERE compromise_assessment_id = $1
`, assessmentID).Scan(&subjectID); err != nil {
		t.Fatalf("lookup assessment subject: %v", err)
	}
	return mustUUID(t, subjectID)
}

func collectRecordChanges(t testing.TB, changes <-chan platformws.RecordChange, want int, timeout time.Duration) []platformws.RecordChange {
	t.Helper()

	deadline := time.After(timeout)
	collected := make([]platformws.RecordChange, 0, want)
	for len(collected) < want {
		select {
		case change := <-changes:
			collected = append(collected, change)
		case <-deadline:
			t.Fatalf("timed out waiting for %d record changes, got %#v", want, collected)
		}
	}
	return collected
}

func requireRecordChange(t testing.TB, changes []platformws.RecordChange, recordID uuid.UUID, viewSchemaID string) {
	t.Helper()

	for _, change := range changes {
		if change.RecordID == recordID && change.ViewSchemaID == viewSchemaID {
			return
		}
	}
	payload, _ := json.Marshal(changes)
	t.Fatalf("expected record change for record=%s view=%s, got %s", recordID, viewSchemaID, string(payload))
}

func httptestSuccess(t testing.TB, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()
	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected status: got %d want %d body=%#v", resp.StatusCode, wantStatus, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.RequireSuccessEnvelope(t, resp, wantStatus)["data"].(map[string]any)
}

func httptestError(t testing.TB, resp *http.Response, wantStatus int, wantCode string) map[string]any {
	t.Helper()
	return httptestx.RequireErrorEnvelope(t, resp, wantStatus, wantCode)
}

func requireIndicatorCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		t.Fatalf("missing indicator cell %s in %#v", fieldKey, row)
	}
	if got := cell["value"]; got != want {
		t.Fatalf("unexpected indicator cell %s: got %#v want %#v", fieldKey, got, want)
	}
}

func lookupIndicatorRecord(t testing.TB, db *sql.DB, recordID uuid.UUID) indicatorRecordRow {
	t.Helper()

	var (
		row             indicatorRecordRow
		recordIDRaw     string
		incidentIDRaw   string
		normalizedValue sql.NullString
		defangedValue   sql.NullString
		hashAlgorithm   sql.NullString
		hashValue       sql.NullString
		stixPattern     sql.NullString
		createdByRaw    string
		updatedByRaw    string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT
    record_id::text,
    incident_id::text,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    dedupe_key,
    defanged_value,
    hash_algorithm,
    hash_value,
    stix_pattern,
    row_version,
    created_by_user_id::text,
    updated_by_user_id::text
  FROM indicators
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &incidentIDRaw, &row.IndicatorType, &row.ValueKind, &row.DisplayValue, &normalizedValue, &row.DedupeKey, &defangedValue, &hashAlgorithm, &hashValue, &stixPattern, &row.RowVersion, &createdByRaw, &updatedByRaw); err != nil {
		t.Fatalf("lookup indicator record: %v", err)
	}
	row.RecordID = mustUUID(t, recordIDRaw)
	row.IncidentID = mustUUID(t, incidentIDRaw)
	row.NormalizedValue = nullStringPointer(normalizedValue)
	row.DefangedValue = nullStringPointer(defangedValue)
	row.HashAlgorithm = nullStringPointer(hashAlgorithm)
	row.HashValue = nullStringPointer(hashValue)
	row.STIXPattern = nullStringPointer(stixPattern)
	row.CreatedByUser = mustUUID(t, createdByRaw)
	row.UpdatedByUser = mustUUID(t, updatedByRaw)
	return row
}

func lookupIndicatorProjection(t testing.TB, db *sql.DB, recordID uuid.UUID) indicatorProjectionRow {
	t.Helper()

	var (
		row              indicatorProjectionRow
		recordIDRaw      string
		normalizedValue  sql.NullString
		defangedValue    sql.NullString
		hashAlgorithm    sql.NullString
		hashValue        sql.NullString
		stixPattern      sql.NullString
		firstObservedAt  sql.NullTime
		lastObservedAt   sql.NullTime
		lifecycleSummary sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT
    record_id::text,
    row_version,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    defanged_value,
    hash_algorithm,
    hash_value,
    stix_pattern,
    first_observed_at,
    last_observed_at,
    observation_count,
    lifecycle_summary,
    supporting_link_count
  FROM indicator_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &row.RowVersion, &row.IndicatorType, &row.ValueKind, &row.DisplayValue, &normalizedValue, &defangedValue, &hashAlgorithm, &hashValue, &stixPattern, &firstObservedAt, &lastObservedAt, &row.ObservationCount, &lifecycleSummary, &row.SupportingLinkCount); err != nil {
		t.Fatalf("lookup indicator projection: %v", err)
	}
	row.RecordID = mustUUID(t, recordIDRaw)
	row.NormalizedValue = nullStringPointer(normalizedValue)
	row.DefangedValue = nullStringPointer(defangedValue)
	row.HashAlgorithm = nullStringPointer(hashAlgorithm)
	row.HashValue = nullStringPointer(hashValue)
	row.STIXPattern = nullStringPointer(stixPattern)
	row.FirstObservedAt = nullTimePointer(firstObservedAt)
	row.LastObservedAt = nullTimePointer(lastObservedAt)
	row.LifecycleSummary = nullStringPointer(lifecycleSummary)
	return row
}

func listIndicatorObservations(t testing.TB, db *sql.DB, incidentID uuid.UUID) []indicatorObservationRow {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `
SELECT
    indicator_observation_id::text,
    source_record_id::text,
    source_field_key,
    origin_kind,
    origin_locator,
    observed_text,
    parsed_indicator_type,
    normalized_candidate,
    resolution_status,
    resolved_indicator_record_id::text,
    row_version
  FROM indicator_observations
 WHERE incident_id = $1
 ORDER BY created_at ASC, indicator_observation_id ASC
`, incidentID)
	if err != nil {
		t.Fatalf("list indicator observations: %v", err)
	}
	defer rows.Close()

	result := make([]indicatorObservationRow, 0)
	for rows.Next() {
		var (
			row               indicatorObservationRow
			observationIDRaw  string
			sourceRecordIDRaw string
			parsedType        sql.NullString
			normalized        sql.NullString
			resolvedID        sql.NullString
		)
		if err := rows.Scan(&observationIDRaw, &sourceRecordIDRaw, &row.SourceFieldKey, &row.OriginKind, &row.OriginLocator, &row.ObservedText, &parsedType, &normalized, &row.ResolutionStatus, &resolvedID, &row.RowVersion); err != nil {
			t.Fatalf("scan indicator observation: %v", err)
		}
		row.ObservationID = mustUUID(t, observationIDRaw)
		row.SourceRecordID = mustUUID(t, sourceRecordIDRaw)
		row.ParsedIndicatorType = nullStringPointer(parsedType)
		row.NormalizedCandidate = nullStringPointer(normalized)
		if resolvedID.Valid {
			value := mustUUID(t, resolvedID.String)
			row.ResolvedIndicatorRecordID = &value
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate indicator observations: %v", err)
	}
	return result
}

func lookupIndicatorLifecycleInterval(t testing.TB, db *sql.DB, intervalID uuid.UUID) indicatorLifecycleIntervalRow {
	t.Helper()

	var (
		row            indicatorLifecycleIntervalRow
		intervalIDRaw  string
		indicatorIDRaw string
		validTo        sql.NullTime
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT
    indicator_state_interval_id::text,
    indicator_record_id::text,
    lifecycle_state,
    valid_from,
    valid_to
  FROM indicator_state_intervals
 WHERE indicator_state_interval_id = $1
`, intervalID).Scan(&intervalIDRaw, &indicatorIDRaw, &row.LifecycleState, &row.ValidFrom, &validTo); err != nil {
		t.Fatalf("lookup indicator lifecycle interval: %v", err)
	}
	row.IntervalID = mustUUID(t, intervalIDRaw)
	row.IndicatorID = mustUUID(t, indicatorIDRaw)
	row.ValidTo = nullTimePointer(validTo)
	return row
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func mustUUID(t testing.TB, value string) uuid.UUID {
	t.Helper()

	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", value, err)
	}
	return parsed
}

func nullStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func nullTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}

func derefStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
