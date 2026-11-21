package evidence_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

// TestEvidencePortableBlobExportMetadata_Integration: export refuses finalized
// Blob metadata that does not describe the bytes it would export, as
// evidence.byte_size_digest_agree, in both finalized states.
func TestEvidencePortableBlobExportMetadata_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "evidence-blob-export")
	actor, incident := uuid.New(), uuid.New()
	if _, err := db.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Blob exporter','test-only',false,true,true)`, actor, actor.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'BLOB-EXPORT','blob-export','Blob export','active',$2,$2)`, incident, actor); err != nil {
		t.Fatal(err)
	}
	content := []byte("exported blob\n")
	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])
	store := &exportBlobStore{objects: map[string][]byte{}}
	portability, err := evidence.NewIncidentBundleBlobPortability(store, unusedResourceAllocator{})
	if err != nil {
		t.Fatal(err)
	}
	// insertBlob seeds one finalized Blob of content and returns its identity.
	insertBlob := func(t testing.TB, tx pgx.Tx, state string) uuid.UUID {
		t.Helper()
		id := uuid.New()
		key := "evidence-export/" + id.String()
		store.objects[key] = bytes.Clone(content)
		if _, err := tx.Exec(ctx, `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state, byte_size,
    expected_sha256_hex, observed_size, observed_content_type, observed_sha256_hex,
    target_expires_at, pending_expires_at, finalized_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $6, 'text/plain', $7, now() + interval '1 hour', now() + interval '24 hours', now())
`, id, incident, actor, key, state, len(content), digest); err != nil {
			t.Fatal(err)
		}
		return id
	}
	export := func(t testing.TB, seed func(pgx.Tx)) (int, error) {
		t.Helper()
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		seed(tx)
		members := portabilitysupport.NewPreparationStore()
		err = portability.ExportBlobFiles(ctx, tx, incident, members)
		return len(members.Members()), err
	}
	for _, state := range []string{"available", "quarantined"} {
		t.Run(state+"/agreeing metadata exports its bytes once", func(t *testing.T) {
			members, err := export(t, func(tx pgx.Tx) {
				insertBlob(t, tx, state)
				insertBlob(t, tx, state)
			})
			if err != nil || members != 1 {
				t.Fatalf("agreeing export = %v with %d members, want one deduplicated member", err, members)
			}
		})
		for _, tc := range []struct {
			name, update string
			duplicate    bool
		}{
			{"absent observed digest", "observed_sha256_hex=NULL", false},
			{"absent observed size", "observed_size=NULL", false},
			{"negative observed size", "observed_size=-1", false},
			{"declared observed size disagreement", "observed_size=byte_size+1", false},
			{"expected observed digest disagreement", "expected_sha256_hex=repeat('0',64)", false},
			{"actual member size disagreement", "byte_size=byte_size+1, observed_size=observed_size+1", false},
			{"duplicate digest size disagreement", "byte_size=byte_size+1, observed_size=observed_size+1", true},
		} {
			t.Run(state+"/"+tc.name, func(t *testing.T) {
				_, err := export(t, func(tx pgx.Tx) {
					target := insertBlob(t, tx, state)
					if tc.duplicate {
						target = insertBlob(t, tx, state)
					}
					if _, err := tx.Exec(ctx, `UPDATE object_blobs SET `+tc.update+` WHERE object_blob_id=$1`, target); err != nil {
						t.Fatal(err)
					}
				})
				requireEvidencePortableFailure(t, err, "byte_size_digest_agree")
			})
		}
	}
}

// exportBlobStore serves seeded object bytes to Evidence export.
type exportBlobStore struct {
	objectstore.TypedStore
	objects map[string][]byte
}

func (s *exportBlobStore) Get(_ context.Context, request objectstore.GetObjectRequest) (io.ReadCloser, objectstore.ObjectInfo, error) {
	content, found := s.objects[request.Key]
	if !found {
		return nil, objectstore.ObjectInfo{}, errors.New("object not seeded")
	}
	return io.NopCloser(bytes.NewReader(content)), objectstore.ObjectInfo{}, nil
}

// unusedResourceAllocator satisfies composition; export never stages.
type unusedResourceAllocator struct {
	stagedobjects.ResourceAllocator
}
