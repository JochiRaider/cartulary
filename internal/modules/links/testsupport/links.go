package testsupport

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const LinkProvenanceAutoMatch = "auto_match"

var (
	ManualLinkID    = uuid.MustParse("40000000-0000-0000-0000-000000000701")
	DuplicateLinkID = uuid.MustParse("40000000-0000-0000-0000-000000000703")
	TagIDSurvivor   = uuid.MustParse("40000000-0000-0000-0000-000000000801")
	TagIDLoser      = uuid.MustParse("40000000-0000-0000-0000-000000000802")
)

type LinkFixture struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SourceID     uuid.UUID
	TargetID     uuid.UUID
	LinkType     string
	Provenance   string
	Confidence   *int
	DeletedAt    *time.Time
}

func RequireActiveLink(
	t testing.TB,
	link LinkFixture,
	sourceID uuid.UUID,
	targetID uuid.UUID,
	linkType string,
	provenance string,
	confidence *int,
) {
	t.Helper()
	if link.DeletedAt != nil {
		t.Fatalf("expected active link, got tombstoned link %#v", link)
	}
	if link.SourceID != sourceID || link.TargetID != targetID {
		t.Fatalf("unexpected link endpoints: got %s -> %s want %s -> %s", link.SourceID, link.TargetID, sourceID, targetID)
	}
	if link.LinkType != linkType || link.Provenance != provenance || !reflect.DeepEqual(link.Confidence, confidence) {
		t.Fatalf("unexpected link attributes: got %#v", link)
	}
}

func SeedRecordLink(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	recordLinkID uuid.UUID,
	srcRecordID uuid.UUID,
	dstRecordID uuid.UUID,
	linkType string,
	provenance string,
	confidence *int,
) {
	t.Helper()
	if _, err := execDB(db, `
INSERT INTO record_links (
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, now(), now())
`, recordLinkID, incidentID, srcRecordID, dstRecordID, linkType, provenance, confidence, actorUserID); err != nil {
		t.Fatalf("seed record link: %v", err)
	}
}

func SeedRecordTag(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	recordTagID uuid.UUID,
	recordID uuid.UUID,
	tagName string,
) {
	t.Helper()
	if _, err := execDB(db, `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6)
`, recordTagID, incidentID, recordID, tagName, tagName, actorUserID); err != nil {
		t.Fatalf("seed record tag: %v", err)
	}
}

func LookupActiveLink(
	t testing.TB,
	db any,
	incidentID uuid.UUID,
	sourceID uuid.UUID,
	targetID uuid.UUID,
	linkType string,
) LinkFixture {
	t.Helper()

	var (
		link        LinkFixture
		confidence  sql.NullInt64
		deletedAt   sql.NullTime
		recordLink  string
		incidentRaw string
		sourceRaw   string
		targetRaw   string
	)
	if err := queryRowDB(db, `
SELECT record_link_id::text, incident_id::text, src_record_id::text, dst_record_id::text,
       link_type, provenance, confidence, deleted_at
  FROM record_links
 WHERE incident_id = $1 AND src_record_id = $2 AND dst_record_id = $3
   AND link_type = $4 AND deleted_at IS NULL
`, incidentID, sourceID, targetID, linkType).Scan(
		&recordLink,
		&incidentRaw,
		&sourceRaw,
		&targetRaw,
		&link.LinkType,
		&link.Provenance,
		&confidence,
		&deletedAt,
	); err != nil {
		t.Fatalf("lookup active link: %v", err)
	}
	link.RecordLinkID = uuid.MustParse(recordLink)
	link.IncidentID = uuid.MustParse(incidentRaw)
	link.SourceID = uuid.MustParse(sourceRaw)
	link.TargetID = uuid.MustParse(targetRaw)
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
		panic(fmt.Sprintf("unsupported Links test database %T", db))
	}
}

func execDB(db any, query string, args ...any) (any, error) {
	switch typed := db.(type) {
	case postgres.DB:
		return typed.Exec(context.Background(), query, args...)
	case *sql.DB:
		return typed.ExecContext(context.Background(), query, args...)
	default:
		return nil, fmt.Errorf("unsupported Links test database %T", db)
	}
}
