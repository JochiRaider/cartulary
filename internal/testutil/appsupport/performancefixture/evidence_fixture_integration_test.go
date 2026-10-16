package performancefixture

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image/png"
	"io"
	"testing"
	"time"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	evidencefixture "github.com/JochiRaider/cartulary/internal/modules/evidence/testsupport/performancefixture"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/google/uuid"
)

func TestEvidenceFixtureUsesRealObjectsAndOwnerMutations_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "evidence_performance_fixture")
	objects, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer objects.Close()
	owners, err := NewOwners(harness.DB, objects, conflicttest.NewCodec("fixture-evidence"))
	if err != nil {
		t.Fatal(err)
	}
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "fixture-evidence@example.test", "Evidence fixture", "EvidenceFixture1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, uuid.NewString(), "IR-EVIDENCE-FIXTURE", "Evidence fixture")
	for _, size := range []int{255, (2 << 20) + 1} {
		if _, err := evidencefixture.ImagePayload(size); err == nil {
			t.Fatal("unbounded fixture PNG size accepted")
		}
	}
	data, err := evidencefixture.ImagePayload(2 << 20)
	if err != nil {
		t.Fatal(err)
	}
	again, err := evidencefixture.ImagePayload(2 << 20)
	if err != nil || len(data) != 2<<20 || !bytes.Equal(data, again) {
		t.Fatal("synthetic PNG physical bytes are not deterministic")
	}
	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil || decoded.Bounds().Dx() != 2 || decoded.Bounds().Dy() != 2 {
		t.Fatal("synthetic payload is not a complete decodable PNG")
	}
	payload := bytes.NewBuffer(data)
	command := evidence.PerformanceFixtureCommand{ActorUserID: actor.ID, IncidentID: incident.ID, Count: 3, TitlePrefix: "Performance Evidence", Payload: payload.Bytes(), ContentType: "image/png", Now: time.Now().UTC()}
	// Reject malformed construction before any object or source mutation.
	for _, mutate := range []func(*evidence.PerformanceFixtureCommand){
		func(c *evidence.PerformanceFixtureCommand) { c.Count = 0 },
		func(c *evidence.PerformanceFixtureCommand) { c.Count = 10001 },
		func(c *evidence.PerformanceFixtureCommand) { c.Payload = []byte("not a PNG") },
		func(c *evidence.PerformanceFixtureCommand) { c.Payload = c.Payload[:len(c.Payload)-1] },
		func(c *evidence.PerformanceFixtureCommand) { c.ContentType = "text/plain" },
	} {
		candidate := command
		mutate(&candidate)
		if _, err := owners.Evidence.CreatePerformanceFixtureRecords(ctx, candidate); err == nil {
			t.Fatal("invalid fixture admitted")
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := owners.Evidence.CreatePerformanceFixtureRecords(cancelled, command); err == nil {
		t.Fatal("cancelled fixture admitted")
	}
	before, err := objects.ListObjects(ctx, "")
	if err != nil || len(before) != 0 {
		t.Fatal("invalid fixture mutated objects")
	}
	result, err := owners.Evidence.CreatePerformanceFixtureRecords(ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	if result.ObjectCount != 3 || len(result.RecordIDs) != 3 || result.TotalBytes != int64(3*payload.Len()) {
		t.Fatal("fixture result has wrong population")
	}
	digest := sha256.Sum256(payload.Bytes())
	for i, id := range result.RecordIDs {
		var title, lifecycle, upload, key, sha, contentType string
		var version, size, observed int64
		if err := harness.DB.QueryRow(ctx, `SELECT e.title,e.lifecycle_state,b.upload_state,r.row_version,b.storage_key,b.byte_size,b.observed_size,b.observed_sha256_hex,b.observed_content_type
 FROM evidence e JOIN records r USING (record_id) JOIN object_blobs b ON b.object_blob_id=e.object_blob_id
 WHERE e.record_id=$1 AND e.incident_id=$2`, id, incident.ID).Scan(&title, &lifecycle, &upload, &version, &key, &size, &observed, &sha, &contentType); err != nil {
			t.Fatal(err)
		}
		if title == "" || lifecycle != "available" || upload != "available" || version != 1 || size != int64(payload.Len()) || observed != size || sha != hex.EncodeToString(digest[:]) || contentType != "image/png" {
			t.Fatalf("record %d is not a coherent available Evidence envelope", i)
		}
		var projectionVersion, revisionVersion int64
		var projectedTitle, projectedLifecycle, projectedUpload, projectedHash string
		if err := harness.DB.QueryRow(ctx, `SELECT p.row_version,p.title,p.lifecycle_state,p.upload_state,p.blob_hash,v.row_version
 FROM evidence_grid_projection p JOIN record_revisions v ON v.record_id=p.record_id
 WHERE p.record_id=$1 AND p.incident_id=$2`, id, incident.ID).Scan(&projectionVersion, &projectedTitle, &projectedLifecycle, &projectedUpload, &projectedHash, &revisionVersion); err != nil {
			t.Fatal(err)
		}
		if projectionVersion != version || revisionVersion != version || projectedTitle != title || projectedLifecycle != lifecycle || projectedUpload != upload || projectedHash != sha {
			t.Fatal("ordinary Evidence projection or revision diverged")
		}
		body, info, err := objects.ReadObject(ctx, key, objectstore.ReadOptions{})
		if err != nil {
			t.Fatal(err)
		}
		actual, err := io.ReadAll(body)
		closeErr := body.Close()
		if err != nil || closeErr != nil || !bytes.Equal(actual, payload.Bytes()) || info.ContentType != contentType {
			t.Fatal("object bytes differ from source metadata")
		}
	}
	var count int
	if err := harness.DB.QueryRow(ctx, `SELECT COUNT(*) FROM evidence_object_upload_leases WHERE incident_id=$1`, incident.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("fixture invented historical upload leases")
	}
	if _, err := owners.Evidence.CreatePerformanceFixtureRecords(ctx, command); err == nil {
		t.Fatal("existing Evidence population was overwritten")
	}
	after, err := objects.ListObjects(ctx, "")
	if err != nil || len(after) != 3 {
		t.Fatal("rejected second construction changed object population")
	}

	second := appsupport.CreateIncidentInStore(t, harness.DB, actor, uuid.NewString(), "IR-EVIDENCE-FAILURE", "Evidence fixture failure")
	failedOwners, err := NewOwners(harness.DB, failedFixtureObjects{TypedStore: objects}, conflicttest.NewCodec("fixture-evidence"))
	if err != nil {
		t.Fatal(err)
	}
	command.IncidentID = second.ID
	if _, err := failedOwners.Evidence.CreatePerformanceFixtureRecords(ctx, command); err == nil {
		t.Fatal("failed object transfer admitted metadata")
	}
	if err := harness.DB.QueryRow(ctx, `SELECT COUNT(*) FROM object_blobs WHERE incident_id=$1`, second.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("failed transfer invented available object metadata")
	}
}

type failedFixtureObjects struct{ objectstore.TypedStore }

func (failedFixtureObjects) Put(context.Context, objectstore.PutObjectRequest) (objectstore.PutObjectResult, error) {
	return objectstore.PutObjectResult{}, errors.New("injected fixture write failure")
}
