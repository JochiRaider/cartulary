package storetest

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/fixtures"
	workbookstartupbootstrap "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type StoreHarness struct {
	DB postgres.DB
}

type LoginResult struct {
	SessionCookie *http.Cookie
	CSRFCookie    *http.Cookie
}

func StartStore(t testing.TB, prefix string) *StoreHarness {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	if pgtest.ExplicitPostgresFixturePolicyT(t) == pgtest.PostgresFixturePolicyTemplateClone {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
		pool, err := pgxpool.New(context.Background(), testDB.DSN)
		if err != nil {
			t.Fatalf("open template-clone postgres pool: %v", err)
		}
		t.Cleanup(pool.Close)
		return &StoreHarness{DB: pool}
	}
	return &StoreHarness{DB: postgresHarness.BeginRollbackDBT(t, prefix)}
}

func DoJSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func WithCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func WithHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func ProvisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (LoginResult, uuid.UUID) {
	t.Helper()

	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-records-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	completeInitialEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-records-bootstrap-admin-complete")
	login := LoginLocalUserWithSecondFactor(t, server, "bootstrap-admin@example.test", "BootstrapPass1!", generateTOTPCode(t, secretBase32))

	sessionResp := DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, WithCookies(login.SessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, MustUUID(t, sessionData["user_id"].(string))
}

func CreateIncident(t testing.TB, server *httptestx.Server, admin LoginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		WithCookies(admin.SessionCookie, admin.CSRFCookie),
		WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func CreateIncidentInStore(t testing.TB, pool postgres.DB, actor authn.UserRecord, clientTxnID string, incidentKey string, title string) incidents.IncidentRecord {
	t.Helper()

	store := incidents.NewStoreWithOptions(pool, incidents.StoreOptions{
		WorkbookBootstrap: workbookstartupbootstrap.NewIncidentCreatePreferencesPort(),
	})
	result, err := store.CreateIncident(context.Background(), actor, incidents.CreateIncidentRequest{
		ClientTxnID: clientTxnID,
		IncidentKey: incidentKey,
		Title:       title,
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Now().UTC())
	if err != nil {
		t.Fatalf("create incident in store: %v", err)
	}
	return result.Incident
}

func SeedLocalUserFlags(t testing.TB, db postgres.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) authn.UserRecord {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var record authn.UserRecord
	if err := db.QueryRow(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin, created_at, updated_at, user_version
`, email, displayName, hash, mfaRequired, isActive, isDeploymentAdmin).Scan(
		&record.ID,
		&record.Email,
		&record.DisplayName,
		&record.PasswordHash,
		&record.MFARequired,
		&record.IsActive,
		&record.IsDeploymentAdmin,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.UserVersion,
	); err != nil {
		t.Fatalf("seed local user with flags: %v", err)
	}
	return record
}

func SeedIncidentMembership(t testing.TB, db postgres.DB, incidentID uuid.UUID, userID uuid.UUID, displayName string, role string, addedByUserID uuid.UUID) {
	t.Helper()

	if _, err := db.Exec(context.Background(), `
INSERT INTO incident_memberships (
    incident_id,
    user_id,
    role,
    joined_at,
    added_by_user_id,
    updated_at,
    updated_by_user_id,
    membership_version
)
VALUES ($1, $2, $3, now(), $4, now(), $4, 1)
ON CONFLICT (incident_id, user_id) DO UPDATE
SET role = EXCLUDED.role,
    updated_at = now(),
    updated_by_user_id = EXCLUDED.updated_by_user_id
`, incidentID, userID, role, addedByUserID); err != nil {
		t.Fatalf("seed incident membership: %v", err)
	}

	if _, err := db.Exec(context.Background(), `
INSERT INTO user_workbook_preferences (incident_id, user_id, home_sheet_ref, created_at, updated_at)
VALUES ($1, $2, NULL, now(), now())
ON CONFLICT (incident_id, user_id) DO NOTHING
`, incidentID, userID); err != nil {
		t.Fatalf("seed user workbook preferences: %v", err)
	}

	_ = displayName
}

func SeedHostRecord(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, hostname string, fqdn string, aadDeviceID string) {
	t.Helper()
	SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "host")

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
	if _, err := db.Exec(context.Background(), `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, fqdn, aad_device_id, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, 'canonical', $7, $7)
`, recordID, incidentID, displayName, hostname, fqdnValue, aadDeviceValue, actorUserID); err != nil {
		t.Fatalf("seed host record: %v", err)
	}
}

func SeedIdentityRecord(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, upn string, email string, samAccountName string) {
	t.Helper()
	SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "identity")

	if _, err := db.Exec(context.Background(), `
INSERT INTO identities (record_id, incident_id, display_name, upn, email, sam_account_name, identity_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, 'canonical', $7, $7)
`, recordID, incidentID, displayName, upn, email, samAccountName, actorUserID); err != nil {
		t.Fatalf("seed identity record: %v", err)
	}
}

func SeedEntityAlias(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, entityType string, rawText string) {
	t.Helper()

	normalized, ok := fieldnorm.NormalizeLine(rawText)
	if !ok {
		t.Fatalf("normalize entity alias %q", rawText)
	}
	if _, err := db.Exec(context.Background(), `
INSERT INTO entity_aliases (incident_id, record_id, entity_type, raw_text, normalized_text, classification, created_by_user_id, created_at)
VALUES ($1, $2, $3, $4, $5, 'suggestion_only', $6, now())
`, incidentID, recordID, entityType, rawText, normalized, actorUserID); err != nil {
		t.Fatalf("seed entity alias: %v", err)
	}
}

func SeedTimelineRecord(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID) {
	t.Helper()
	SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "timeline_event")

	if _, err := db.Exec(context.Background(), `
INSERT INTO timeline_events (record_id, incident_id, activity_synopsis_text, capture_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'record-support-source-row', 'reviewed', $3, $3)
`, recordID, incidentID, actorUserID); err != nil {
		t.Fatalf("seed timeline record: %v", err)
	}
}

func SeedResolvedMention(t testing.TB, db postgres.DB, actorUserID uuid.UUID, mentionID uuid.UUID, sourceRecordID uuid.UUID, resolvedRecordID uuid.UUID, sourceFieldKey string, entityType string, rawText string) {
	t.Helper()

	SeedMention(t, db, actorUserID, mentionID, sourceRecordID, sourceFieldKey, entityType, rawText, "resolved", &resolvedRecordID, resolutionMethodPointer("explicit_resolve_route"))
}

func SeedMention(t testing.TB, db postgres.DB, actorUserID uuid.UUID, mentionID uuid.UUID, sourceRecordID uuid.UUID, sourceFieldKey string, entityType string, rawText string, resolutionStatus string, resolvedRecordID *uuid.UUID, resolutionMethod *string) {
	t.Helper()

	var (
		resolvedByUserID any
		resolvedAt       any
		methodValue      any
	)
	if resolvedRecordID != nil {
		resolvedByUserID = actorUserID
		resolvedAt = time.Now().UTC()
		methodValue = resolutionMethod
	}
	if _, err := db.Exec(context.Background(), `
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
VALUES ($1, $2, $3, $4, 'manual_entry', 'records-store-support', $5, $6, $7, 1, 1, $8, $9, $10, $11, $12)
`, mentionID, sourceRecordID, entityType, sourceFieldKey, rawText, strings.ToLower(strings.TrimSpace(rawText)), resolutionStatus, actorUserID, resolvedRecordID, resolvedByUserID, resolvedAt, methodValue); err != nil {
		t.Fatalf("seed mention: %v", err)
	}
}

func SeedRecordLink(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordLinkID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string, provenance string, confidence *int) {
	t.Helper()

	if _, err := db.Exec(context.Background(), `
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

func SeedRecordTag(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordTagID uuid.UUID, recordID uuid.UUID, tagName string) {
	t.Helper()

	if _, err := db.Exec(context.Background(), `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6)
`, recordTagID, incidentID, recordID, tagName, tagName, actorUserID); err != nil {
		t.Fatalf("seed record tag: %v", err)
	}
}

func SeedAssessment(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, assessmentID uuid.UUID, subjectID uuid.UUID, subjectType string, state string) {
	t.Helper()
	SeedRecordEnvelope(t, db, incidentID, actorUserID, assessmentID, "assessment")

	if _, err := db.Exec(context.Background(), `
INSERT INTO assessments (record_id, incident_id, subject_record_id, subject_type, assessment_state, rationale, assessor_user_id)
VALUES ($1, $2, $3, $4, $5, 'Seeded test assessment rationale.', $6)
`, assessmentID, incidentID, subjectID, subjectType, state, actorUserID); err != nil {
		t.Fatalf("seed assessment: %v", err)
	}
	if _, err := db.Exec(context.Background(), `
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

func SeedRecordEnvelope(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, recordType string) {
	t.Helper()

	if _, err := db.Exec(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $4)
ON CONFLICT (record_id) DO NOTHING
`, recordID, incidentID, recordType, actorUserID); err != nil {
		t.Fatalf("seed record envelope: %v", err)
	}
}

func LookupHostState(t testing.TB, db postgres.DB, recordID uuid.UUID) (string, *uuid.UUID, int64, string) {
	t.Helper()

	var (
		state         string
		mergedIntoRaw sql.NullString
		rowVersion    int64
		fqdn          sql.NullString
	)
	if err := db.QueryRow(context.Background(), `
SELECT host_state, merged_into_record_id::text, row_version, COALESCE(fqdn, '')
  FROM hosts
 WHERE record_id = $1
`, recordID).Scan(&state, &mergedIntoRaw, &rowVersion, &fqdn); err != nil {
		t.Fatalf("lookup host state: %v", err)
	}
	var mergedInto *uuid.UUID
	if mergedIntoRaw.Valid {
		value := MustUUID(t, mergedIntoRaw.String)
		mergedInto = &value
	}
	return state, mergedInto, rowVersion, fqdn.String
}

func LookupMention(t testing.TB, db postgres.DB, mentionID uuid.UUID) fixtures.EntityMentionFixture {
	t.Helper()

	var mention fixtures.EntityMentionFixture
	var (
		mentionIDRaw     string
		sourceRecordID   string
		resolvedRecordID sql.NullString
		resolvedByUserID sql.NullString
		resolvedAt       sql.NullTime
		resolutionMethod sql.NullString
	)
	if err := db.QueryRow(context.Background(), `
SELECT entity_mention_id::text, source_record_id::text, raw_text, resolution_status, row_version, resolved_record_id::text, resolved_by_user_id::text, resolved_at, resolution_method
  FROM entity_mentions
 WHERE entity_mention_id = $1
`, mentionID).Scan(
		&mentionIDRaw,
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

	mention.EntityMentionID = MustUUID(t, mentionIDRaw)
	mention.SourceRecordID = MustUUID(t, sourceRecordID)
	if resolvedRecordID.Valid {
		value := MustUUID(t, resolvedRecordID.String)
		mention.ResolvedRecordID = &value
	}
	if resolvedByUserID.Valid {
		value := MustUUID(t, resolvedByUserID.String)
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

func LookupActiveLink(t testing.TB, db postgres.DB, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, linkType string) fixtures.LinkFixture {
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
	if err := db.QueryRow(context.Background(), `
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
	link.RecordLinkID = MustUUID(t, recordLink)
	link.IncidentID = MustUUID(t, incidentRaw)
	link.SourceID = MustUUID(t, sourceRaw)
	link.TargetID = MustUUID(t, targetRaw)
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

func LookupAssessmentSubject(t testing.TB, db postgres.DB, assessmentID uuid.UUID) uuid.UUID {
	t.Helper()

	var subjectID string
	if err := db.QueryRow(context.Background(), `
SELECT subject_record_id::text
  FROM assessments
 WHERE record_id = $1
`, assessmentID).Scan(&subjectID); err != nil {
		t.Fatalf("lookup assessment subject: %v", err)
	}
	return MustUUID(t, subjectID)
}

type IndicatorProjectionRow struct {
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

type IndicatorLifecycleIntervalRow struct {
	IntervalID     uuid.UUID
	IndicatorID    uuid.UUID
	LifecycleState string
	ValidFrom      time.Time
	ValidTo        *time.Time
}

func LookupIndicatorProjection(t testing.TB, db postgres.DB, recordID uuid.UUID) IndicatorProjectionRow {
	t.Helper()

	var (
		row              IndicatorProjectionRow
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
	if err := db.QueryRow(context.Background(), `
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
	row.RecordID = MustUUID(t, recordIDRaw)
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

func LookupIndicatorLifecycleInterval(t testing.TB, db postgres.DB, intervalID uuid.UUID) IndicatorLifecycleIntervalRow {
	t.Helper()

	var (
		row            IndicatorLifecycleIntervalRow
		intervalIDRaw  string
		indicatorIDRaw string
		validTo        sql.NullTime
	)
	if err := db.QueryRow(context.Background(), `
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
	row.IntervalID = MustUUID(t, intervalIDRaw)
	row.IndicatorID = MustUUID(t, indicatorIDRaw)
	row.ValidTo = nullTimePointer(validTo)
	return row
}

func RequireSuccessData(t testing.TB, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()

	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected status: got %d want %d body=%#v", resp.StatusCode, wantStatus, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.RequireSuccessEnvelope(t, resp, wantStatus)["data"].(map[string]any)
}

func RequireErrorBody(t testing.TB, resp *http.Response, wantStatus int, wantCode string) map[string]any {
	t.Helper()

	return httptestx.RequireErrorEnvelope(t, resp, wantStatus, wantCode)
}

func MustUUID(t testing.TB, value string) uuid.UUID {
	t.Helper()

	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", value, err)
	}
	return parsed
}

func QueryCount(t testing.TB, db postgres.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func nullStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	timestamp := value.Time.UTC()
	return &timestamp
}

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	return details["bootstrap_token"].(string)
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, WithHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, WithHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func LoginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) LoginResult {
	t.Helper()

	resp := DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
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
	return LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
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

func resolutionMethodPointer(value string) *string {
	return &value
}

func SerializeActiveLinkRead(t testing.TB, link fixtures.LinkFixture) map[string]any {
	t.Helper()

	data, err := json.Marshal(map[string]any{
		"record_link_id": link.RecordLinkID.String(),
		"provenance":     link.Provenance,
		"confidence":     link.Confidence,
	})
	if err != nil {
		t.Fatalf("marshal active link helper read: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal active link helper read: %v", err)
	}
	return payload
}
