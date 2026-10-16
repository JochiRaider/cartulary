package evidence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// PerformanceFixtureContribution is an isolated fixture capability. It is not
// exposed by HTTP or generic Workbook dispatch. The builder owns the entire
// disposable database and object namespace, including partial construction.
type PerformanceFixtureContribution struct {
	postgres    postgres.DB
	objectStore objectstore.TypedStore
	source      *SourceRuntime
}

func NewPerformanceFixtureContribution(db postgres.DB, objects objectstore.TypedStore, source *SourceRuntime) *PerformanceFixtureContribution {
	if isNilMutationCapability(db) || isNilMutationCapability(objects) || source == nil {
		return nil
	}
	return &PerformanceFixtureContribution{postgres: db, objectStore: objects, source: source}
}

type PerformanceFixtureCommand struct {
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	Count       int
	TitlePrefix string
	Payload     []byte
	ContentType string
	Now         time.Time
}

type PerformanceFixtureResult struct {
	RecordIDs   []uuid.UUID
	ObjectCount int
	TotalBytes  int64
}

// CreatePerformanceFixtureRecords materializes upload-complete fixture blobs
// from actual bytes, then uses normal Evidence create admission and mutations
// for every envelope, association, lifecycle, revision and projection. It does
// not simulate a historical upload lease or mint access handles.
func (contribution *PerformanceFixtureContribution) CreatePerformanceFixtureRecords(ctx context.Context, command PerformanceFixtureCommand) (PerformanceFixtureResult, error) {
	if contribution == nil || contribution.source == nil || command.ActorUserID == uuid.Nil || command.IncidentID == uuid.Nil || command.Now.IsZero() || command.Count < 1 || command.Count > 10000 || len(command.Payload) == 0 || len(command.Payload) > 2<<20 || strings.TrimSpace(command.TitlePrefix) == "" || len(command.TitlePrefix) > 160 || command.ContentType != "image/png" {
		return PerformanceFixtureResult{}, errors.New("evidence performance fixture command is invalid")
	}
	runtime := contribution
	var existing int
	if err := runtime.postgres.QueryRow(ctx, `SELECT COUNT(*) FROM evidence WHERE incident_id=$1`, command.IncidentID).Scan(&existing); err != nil {
		return PerformanceFixtureResult{}, err
	}
	if existing != 0 {
		return PerformanceFixtureResult{}, errors.New("evidence performance fixture requires an empty Evidence population")
	}
	// Freeze caller-owned bytes for the complete population.
	payload := bytes.Clone(command.Payload)
	config, err := png.DecodeConfig(bytes.NewReader(payload))
	if err != nil || config.Width != 2 || config.Height != 2 {
		return PerformanceFixtureResult{}, errors.New("evidence fixture payload requires a bounded 2 by 2 PNG image")
	}
	if _, err := png.Decode(bytes.NewReader(payload)); err != nil {
		return PerformanceFixtureResult{}, errors.New("evidence fixture PNG payload is incomplete")
	}
	digest := sha256.Sum256(payload)
	sha := hex.EncodeToString(digest[:])
	now := command.Now.UTC()
	result := PerformanceFixtureResult{RecordIDs: make([]uuid.UUID, 0, command.Count)}
	for i := 0; i < command.Count; i++ {
		if err := ctx.Err(); err != nil {
			return PerformanceFixtureResult{}, err
		}
		blobID := uuid.New()
		key, err := blobref.ObjectBlobStorageKey(command.IncidentID, blobID)
		if err != nil {
			return PerformanceFixtureResult{}, err
		}
		put, err := runtime.objectStore.Put(ctx, objectstore.PutObjectRequest{Key: key, Body: bytes.NewReader(payload), Size: int64(len(payload)), ContentType: command.ContentType, Purpose: objectstore.PurposeProductUpload})
		if err != nil {
			return PerformanceFixtureResult{}, fmt.Errorf("write Evidence fixture object: %w", err)
		}
		if put.SizeBytes != int64(len(payload)) {
			return PerformanceFixtureResult{}, errors.New("evidence fixture object write size diverged")
		}
		// This is an explicit fixture prestate, not fabricated upload provenance.
		// The available metadata is bound to the bytes just written; the shared
		// snapshot owner independently hashes and verifies the complete population.
		_, err = runtime.postgres.Exec(ctx, `INSERT INTO object_blobs (
 object_blob_id,incident_id,created_by_user_id,storage_key,upload_state,
 byte_size,filename_hint,content_type_hint,expected_sha256_hex,
 observed_size,observed_content_type,observed_sha256_hex,
 target_expires_at,pending_expires_at,finalized_at,created_at,updated_at
 ) VALUES ($1,$2,$3,$4,'available',$5,$6,$7,$8,$5,$7,$8,$9,$10,$11,$11,$11)`,
			blobID, command.IncidentID, command.ActorUserID, key, int64(len(payload)), fmt.Sprintf("fixture-%05d.png", i+1), command.ContentType, sha, now.Add(time.Hour), now.Add(24*time.Hour), now)
		if err != nil {
			return PerformanceFixtureResult{}, fmt.Errorf("materialize Evidence fixture blob metadata: %w", err)
		}
		body, err := json.Marshal(map[string]any{"client_txn_id": uuid.NewString(), "evidence.title": fmt.Sprintf("%s %05d", command.TitlePrefix, i+1), "evidence.lifecycle_state": "available", "evidence.initial_object_blob_id": blobID.String()})
		if err != nil {
			return PerformanceFixtureResult{}, err
		}
		admission, failure := AdmitCreateJSON(bytes.NewReader(body))
		if failure != nil {
			return PerformanceFixtureResult{}, errors.New("evidence fixture create admission failed")
		}
		created, err := runtime.source.MutationContribution().Create(ctx, CreateCommand{ActorUserID: command.ActorUserID, IncidentID: command.IncidentID, Admission: admission, RequestID: uuid.NewString(), Now: now})
		if err != nil {
			return PerformanceFixtureResult{}, fmt.Errorf("create Evidence fixture record: %w", err)
		}
		if created.Outcome != MutationOutcomeCreated || created.RecordID == uuid.Nil || created.RowVersion != 1 {
			return PerformanceFixtureResult{}, errors.New("evidence fixture create returned an invalid identity")
		}
		result.RecordIDs = append(result.RecordIDs, created.RecordID)
		result.ObjectCount++
		result.TotalBytes += int64(len(payload))
	}
	return result, nil
}
