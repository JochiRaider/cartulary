package testsupport

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	MentionStatusUnresolved = "unresolved"
	MentionStatusResolved   = "resolved"
	MentionStatusDismissed  = "dismissed"
	MentionActionResolve    = "resolve_item"
	MentionActionDismiss    = "dismiss_item"
)

func HostCreatePayload(clientTxnID string) map[string]any {
	return map[string]any{
		"client_txn_id":     clientTxnID,
		"host.display_name": "VPN Gateway",
		"host.hostname":     "vpn-gateway",
	}
}

func IdentityCreatePayload(clientTxnID string) map[string]any {
	return map[string]any{
		"client_txn_id":             clientTxnID,
		"identity.display_name":     "VPN User",
		"identity.email":            "vpn.user@example.test",
		"identity.sam_account_name": "VPNUSER",
	}
}

var (
	BaseTime                       = time.Date(2026, time.April, 18, 14, 30, 0, 0, time.UTC)
	CanonicalHostRecordID          = uuid.MustParse("40000000-0000-0000-0000-000000000301")
	StubHostRecordID               = uuid.MustParse("40000000-0000-0000-0000-000000000302")
	DuplicateHostRecordID          = uuid.MustParse("40000000-0000-0000-0000-000000000304")
	CanonicalIdentityRecordID      = uuid.MustParse("40000000-0000-0000-0000-000000000401")
	DuplicateIdentityRecordID      = uuid.MustParse("40000000-0000-0000-0000-000000000404")
	HostMentionID                  = uuid.MustParse("40000000-0000-0000-0000-000000000601")
	ResolvedHostMentionID          = uuid.MustParse("40000000-0000-0000-0000-000000000602")
	IdentityMentionID              = uuid.MustParse("40000000-0000-0000-0000-000000000604")
	AutoResolutionSuppressedTokens = []string{
		"WS-023?", "WS-023??", "WS-023 ~", "WS-023 maybe", "WS-023 prob",
		"WS-023 probably", "WS-023 approx", "WS-023 approximately", "(WS-023)",
		"WS-023.", "WS-023,", "WS-023 likely",
	}
)

type MentionFixture struct {
	EntityMentionID  uuid.UUID
	SourceRecordID   uuid.UUID
	EntityType       string
	SourceFieldKey   string
	OriginKind       string
	OriginLocator    string
	RawText          string
	NormalizedText   string
	ResolutionStatus string
	RowVersion       int64
	ResolvedRecordID *uuid.UUID
	ResolvedByUserID *uuid.UUID
	ResolvedAt       *time.Time
	ResolutionMethod *string
	Ordinal          int
}

func RequireMentionStatus(t testing.TB, mention MentionFixture, want string) {
	t.Helper()
	if mention.ResolutionStatus != want {
		t.Fatalf("unexpected mention resolution_status: got %q want %q", mention.ResolutionStatus, want)
	}
}

func RequireRawTextPreserved(t testing.TB, before string, after string) {
	t.Helper()
	if before != after {
		t.Fatalf("expected raw_text to remain unchanged, before=%q after=%q", before, after)
	}
}

func SeedHostRecord(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	recordID uuid.UUID,
	displayName string,
	hostname string,
	fqdn string,
	aadDeviceID string,
) {
	t.Helper()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "host")

	var fqdnValue, aadDeviceValue any
	if fqdn != "" {
		fqdnValue = fqdn
	}
	if aadDeviceID != "" {
		aadDeviceValue = aadDeviceID
	}
	if _, err := execDB(db, `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, fqdn, aad_device_id,
    host_state, created_by_user_id, updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, 'canonical', $7, $7)
`, recordID, incidentID, displayName, hostname, fqdnValue, aadDeviceValue, actorUserID); err != nil {
		t.Fatalf("seed host record: %v", err)
	}
}

func SeedIdentityRecord(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	recordID uuid.UUID,
	displayName string,
	upn string,
	email string,
	samAccountName string,
) {
	t.Helper()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "identity")
	if _, err := execDB(db, `
INSERT INTO identities (
    record_id, incident_id, display_name, upn, email, sam_account_name,
    identity_state, created_by_user_id, updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, 'canonical', $7, $7)
`, recordID, incidentID, displayName, upn, email, samAccountName, actorUserID); err != nil {
		t.Fatalf("seed identity record: %v", err)
	}
}

func SeedEntityAlias(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	recordID uuid.UUID,
	entityType string,
	rawText string,
) {
	t.Helper()
	normalized, ok := fieldnorm.NormalizeLine(rawText)
	if !ok {
		t.Fatalf("normalize entity alias %q", rawText)
	}
	if _, err := execDB(db, `
INSERT INTO entity_aliases (
    incident_id, record_id, entity_type, raw_text, normalized_text,
    classification, created_by_user_id, created_at
)
VALUES ($1, $2, $3, $4, $5, 'suggestion_only', $6, now())
`, incidentID, recordID, entityType, rawText, normalized, actorUserID); err != nil {
		t.Fatalf("seed entity alias: %v", err)
	}
}

func SeedResolvedMention(
	t testing.TB,
	db any,
	actorUserID uuid.UUID,
	mentionID uuid.UUID,
	sourceRecordID uuid.UUID,
	resolvedRecordID uuid.UUID,
	sourceFieldKey string,
	entityType string,
	rawText string,
) {
	t.Helper()
	method := "explicit_resolve_route"
	SeedMention(
		t,
		db,
		actorUserID,
		mentionID,
		sourceRecordID,
		sourceFieldKey,
		entityType,
		rawText,
		MentionStatusResolved,
		&resolvedRecordID,
		&method,
	)
}

func SeedMention(
	t testing.TB,
	db any,
	actorUserID uuid.UUID,
	mentionID uuid.UUID,
	sourceRecordID uuid.UUID,
	sourceFieldKey string,
	entityType string,
	rawText string,
	resolutionStatus string,
	resolvedRecordID *uuid.UUID,
	resolutionMethod *string,
) {
	t.Helper()

	var resolvedByUserID, resolvedAt, methodValue any
	if resolvedRecordID != nil {
		resolvedByUserID = actorUserID
		resolvedAt = time.Now().UTC()
		methodValue = resolutionMethod
	}
	if _, err := execDB(db, `
INSERT INTO entity_mentions (
    entity_mention_id, source_record_id, entity_type, source_field_key,
    origin_kind, origin_locator, raw_text, normalized_text, resolution_status,
    row_version, ordinal, created_by_user_id, resolved_record_id,
    resolved_by_user_id, resolved_at, resolution_method
)
VALUES ($1, $2, $3, $4, 'manual_entry', 'entities-test-support', $5, $6, $7,
        1, 1, $8, $9, $10, $11, $12)
`, mentionID, sourceRecordID, entityType, sourceFieldKey, rawText,
		strings.ToLower(strings.TrimSpace(rawText)), resolutionStatus, actorUserID,
		resolvedRecordID, resolvedByUserID, resolvedAt, methodValue); err != nil {
		t.Fatalf("seed mention: %v", err)
	}
}

func LookupHostState(t testing.TB, db any, recordID uuid.UUID) (string, *uuid.UUID, int64, string) {
	t.Helper()

	var (
		state         string
		mergedIntoRaw sql.NullString
		rowVersion    int64
		fqdn          sql.NullString
	)
	if err := queryRowDB(db, `
SELECT host_state, merged_into_record_id::text, row_version, COALESCE(fqdn, '')
  FROM hosts
 WHERE record_id = $1
`, recordID).Scan(&state, &mergedIntoRaw, &rowVersion, &fqdn); err != nil {
		t.Fatalf("lookup host state: %v", err)
	}
	var mergedInto *uuid.UUID
	if mergedIntoRaw.Valid {
		value := uuid.MustParse(mergedIntoRaw.String)
		mergedInto = &value
	}
	return state, mergedInto, rowVersion, fqdn.String
}

func LookupMention(t testing.TB, db any, mentionID uuid.UUID) MentionFixture {
	t.Helper()

	var (
		mention          MentionFixture
		mentionIDRaw     string
		sourceRecordID   string
		resolvedRecordID sql.NullString
		resolvedByUserID sql.NullString
		resolvedAt       sql.NullTime
		resolutionMethod sql.NullString
	)
	if err := queryRowDB(db, `
SELECT entity_mention_id::text, source_record_id::text, raw_text, resolution_status,
       row_version, resolved_record_id::text, resolved_by_user_id::text,
       resolved_at, resolution_method
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
	mention.EntityMentionID = uuid.MustParse(mentionIDRaw)
	mention.SourceRecordID = uuid.MustParse(sourceRecordID)
	if resolvedRecordID.Valid {
		value := uuid.MustParse(resolvedRecordID.String)
		mention.ResolvedRecordID = &value
	}
	if resolvedByUserID.Valid {
		value := uuid.MustParse(resolvedByUserID.String)
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

func MentionItemRef(mentionID uuid.UUID) string {
	return fmt.Sprintf("entity_mention:%s", mentionID.String())
}

func MentionIDFromItemRef(t testing.TB, itemRef string) uuid.UUID {
	t.Helper()
	const prefix = "entity_mention:"
	if !strings.HasPrefix(itemRef, prefix) {
		t.Fatalf("unexpected mention item_ref: %s", itemRef)
	}
	return uuid.MustParse(strings.TrimPrefix(itemRef, prefix))
}

func ResolveItemAction(itemRef string, resolvedRecordID uuid.UUID) map[string]any {
	return map[string]any{
		"op":                 "resolve_item",
		"item_ref":           itemRef,
		"resolved_record_id": resolvedRecordID.String(),
	}
}

func MentionResolveRoutePayload(
	baseMentionRowVersion int64,
	clientTxnID string,
	action string,
	resolvedRecordID *uuid.UUID,
	reason *string,
) map[string]any {
	payload := map[string]any{
		"base_mention_row_version": baseMentionRowVersion,
		"client_txn_id":            clientTxnID,
		"action":                   action,
	}
	if resolvedRecordID != nil {
		payload["resolved_record_id"] = resolvedRecordID.String()
	}
	if reason != nil {
		payload["reason"] = *reason
	}
	return payload
}

type rowScanner interface {
	Scan(dest ...any) error
}

func queryRowDB(db any, query string, args ...any) rowScanner {
	switch typed := db.(type) {
	case postgres.DB:
		return typed.QueryRow(context.Background(), query, args...)
	case *sql.DB:
		return typed.QueryRowContext(context.Background(), query, args...)
	default:
		panic(fmt.Sprintf("unsupported Entities test database %T", db))
	}
}

func execDB(db any, query string, args ...any) (any, error) {
	switch typed := db.(type) {
	case postgres.DB:
		return typed.Exec(context.Background(), query, args...)
	case *sql.DB:
		return typed.ExecContext(context.Background(), query, args...)
	default:
		return nil, fmt.Errorf("unsupported Entities test database %T", db)
	}
}
