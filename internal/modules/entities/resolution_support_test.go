package entities_test

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

func provisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (loginResult, uuid.UUID) {
	t.Helper()

	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-entity_linking-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	completeInitialEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-entity_linking-bootstrap-admin-complete")
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

func createEntityRow(t testing.TB, serverURL string, incidentID string, viewSchemaID string, actor appsupport.LoginResult, body map[string]any, wantStatus int) map[string]any {
	t.Helper()

	resp := appsupport.DoJSON(
		t,
		http.MethodPost,
		serverURL+"/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/rows",
		body,
		appsupport.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return appsupport.RequireSuccessData(t, resp, wantStatus)
}

func loginLocalUser(t testing.TB, server *httptestx.Server, username string, password string) appsupport.LoginResult {
	t.Helper()

	resp := appsupport.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
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
	return appsupport.LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
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

func doEntitiesRawJSON(t testing.TB, method string, url string, body string, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("create raw json request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
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
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "host")

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
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "timeline_event")

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO timeline_events (record_id, incident_id, activity_synopsis_text, capture_state, created_by_user_id, updated_by_user_id)
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
    created_by_user_id,
    decided_at,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, now(), now())
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
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, assessmentID, "assessment")

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO assessments (record_id, incident_id, subject_record_id, subject_type, assessment_state, rationale, assessor_user_id)
VALUES ($1, $2, $3, $4, $5, 'Seeded test assessment rationale.', $6)
`, assessmentID, incidentID, subjectID, subjectType, state, actorUserID); err != nil {
		t.Fatalf("seed assessment: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO assessment_grid_projection (
    record_id,
    incident_id,
    row_version,
    subject_ref,
    subject_type,
    assessment_state,
    confidence_band,
    rationale,
    assessor,
    assessed_at,
    supporting_link_count
)
SELECT a.record_id, a.incident_id, r.row_version, a.subject_record_id, a.subject_type, a.assessment_state, 'unset', a.rationale, a.assessor_user_id, a.assessed_at, 0
  FROM assessments a
  JOIN records r ON r.record_id = a.record_id
 WHERE a.record_id = $1
ON CONFLICT (record_id) DO NOTHING
`, assessmentID); err != nil {
		t.Fatalf("seed assessment projection: %v", err)
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

func lookupMention(t testing.TB, db *sql.DB, mentionID uuid.UUID) entitytest.MentionFixture {
	t.Helper()

	var mention entitytest.MentionFixture
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

func lookupActiveLink(t testing.TB, db *sql.DB, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, linkType string) linktest.LinkFixture {
	t.Helper()

	var (
		link        linktest.LinkFixture
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
SELECT subject_record_id::text
  FROM assessments
 WHERE record_id = $1
`, assessmentID).Scan(&subjectID); err != nil {
		t.Fatalf("lookup assessment subject: %v", err)
	}
	return mustUUID(t, subjectID)
}

func requireViewRowFieldSurface(t testing.TB, testID string, row map[string]any, viewSchemaID string) {
	t.Helper()

	contractassert.RequireFieldKeyConformance(
		t,
		workbookscenariotest.SortedRowFieldKeys(t, row),
		workbookscenariotest.AllowedFieldKeys(t, testID, viewSchemaID),
	)
}

type hostProjectionSnapshot struct {
	RecordID    uuid.UUID
	IncidentID  uuid.UUID
	RowVersion  int64
	DisplayName string
	Hostname    string
	HostState   string
	EditedAt    time.Time
}

type identityProjectionSnapshot struct {
	RecordID       uuid.UUID
	IncidentID     uuid.UUID
	RowVersion     int64
	DisplayName    string
	Email          *string
	SamAccountName *string
	IdentityState  string
	EditedAt       time.Time
}

func lookupHostProjectionSnapshot(t testing.TB, db *sql.DB, recordID uuid.UUID) hostProjectionSnapshot {
	t.Helper()

	var (
		snapshot      hostProjectionSnapshot
		recordIDRaw   string
		incidentIDRaw string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT record_id::text, incident_id::text, row_version, display_name, hostname, host_state, edited_at
  FROM host_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &incidentIDRaw, &snapshot.RowVersion, &snapshot.DisplayName, &snapshot.Hostname, &snapshot.HostState, &snapshot.EditedAt); err != nil {
		t.Fatalf("lookup host projection snapshot: %v", err)
	}
	snapshot.RecordID = mustUUID(t, recordIDRaw)
	snapshot.IncidentID = mustUUID(t, incidentIDRaw)
	snapshot.EditedAt = snapshot.EditedAt.UTC()
	return snapshot
}

func lookupIdentityProjectionSnapshot(t testing.TB, db *sql.DB, recordID uuid.UUID) identityProjectionSnapshot {
	t.Helper()

	var (
		snapshot       identityProjectionSnapshot
		recordIDRaw    string
		incidentIDRaw  string
		email          *string
		samAccountName *string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT record_id::text, incident_id::text, row_version, display_name, email::text, sam_account_name, identity_state, edited_at
  FROM identity_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &incidentIDRaw, &snapshot.RowVersion, &snapshot.DisplayName, &email, &samAccountName, &snapshot.IdentityState, &snapshot.EditedAt); err != nil {
		t.Fatalf("lookup identity projection snapshot: %v", err)
	}
	snapshot.RecordID = mustUUID(t, recordIDRaw)
	snapshot.IncidentID = mustUUID(t, incidentIDRaw)
	snapshot.Email = email
	snapshot.SamAccountName = samAccountName
	snapshot.EditedAt = snapshot.EditedAt.UTC()
	return snapshot
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func lookupIdentityState(t testing.TB, db *sql.DB, recordID uuid.UUID) (string, *uuid.UUID, int64, string) {
	t.Helper()

	var (
		state         string
		mergedIntoRaw sql.NullString
		rowVersion    int64
		email         sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT identity_state, merged_into_record_id::text, row_version, email::text
  FROM identities
 WHERE record_id = $1
`, recordID).Scan(&state, &mergedIntoRaw, &rowVersion, &email); err != nil {
		t.Fatalf("lookup identity state: %v", err)
	}
	var mergedInto *uuid.UUID
	if mergedIntoRaw.Valid {
		value := mustUUID(t, mergedIntoRaw.String)
		mergedInto = &value
	}
	return state, mergedInto, rowVersion, email.String
}

func mustUUID(t testing.TB, value string) uuid.UUID {
	t.Helper()

	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", value, err)
	}
	return parsed
}
