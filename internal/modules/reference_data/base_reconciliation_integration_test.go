package reference_data

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func TestBaseReconciliationRechecksSelectedBytesBeforePublication_Integration(t *testing.T) {
	for _, mode := range []string{"expired_healthy", "missing", "altered", "operational", "stale", "unrelated"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			f := newCanonicalCoordinatorFixture(t, "reference-pack-startup-"+mode)
			f.coordinator.registryUsage = lifecycleFixtureUsage{}
			accepted, err := f.coordinator.Import(ctx, f.actor, "startup-import", bytes.NewReader(f.container))
			if err != nil {
				t.Fatal(err)
			}
			id := uuid.MustParse(accepted.Job.JobID)
			execution, claimed, err := f.manager.Claim(ctx, id)
			if err != nil || !claimed {
				t.Fatal("claim", err)
			}
			if err := f.coordinator.Execute(ctx, execution); err != nil {
				t.Fatal(err)
			}
			job, err := f.manager.Get(ctx, id)
			if err != nil || job.Status != jobs.StatusSucceeded {
				t.Fatal("import", job.Status, err)
			}
			request, apiErr := DecodeActionRequest(strings.NewReader(`{"client_txn_id":"startup-activate"}`))
			if apiErr != nil {
				t.Fatal(apiErr)
			}
			if _, err := f.coordinator.Activate(ctx, ActionParams{ActorUserID: f.actor, PackKey: "type_registry.host", PackVersion: "signed-fixture.1", Request: request, Now: *f.now}); err != nil {
				t.Fatal(err)
			}
			before, err := f.coordinator.GetVersion(ctx, "type_registry.host", "signed-fixture.1")
			if err != nil {
				t.Fatal(err)
			}
			var reference string
			if err := f.pool.QueryRow(ctx, `SELECT o.storage_ref FROM reference_pack_objects o JOIN reference_pack_object_refs r USING(object_id) WHERE r.owner_kind='version' AND r.owner_id=$1 AND r.logical_path='payload/entries.ndjson'`, versionObjectID("type_registry.host", "signed-fixture.1")).Scan(&reference); err != nil {
				t.Fatal(err)
			}
			storage := &reconciliationTestStorage{ArtifactStorage: f.storage, reference: reference}
			fault := errors.New("injected operational storage failure")
			switch mode {
			case "missing":
				delete(f.storage.objects, reference)
			case "altered":
				data := bytes.Clone(f.storage.objects[reference])
				data[0] ^= 1
				f.storage.objects[reference] = data
			case "operational":
				storage.readErr = fault
			case "stale", "unrelated":
				key := "type_registry.host"
				if mode == "unrelated" {
					key = "framework.attack"
				}
				storage.beforeRead = func() {
					if _, err := f.pool.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, key); err != nil {
						t.Fatal(err)
					}
				}
			}
			// Reconciliation is historical integrity, so current expiry and the
			// fresh-operation clock assertion must not reject healthy content.
			err = ReconcileBaseRelease(ctx, f.pool, storage, BaseReleaseOptions{ProfileClaimed: true, Limits: DefaultLimits()}, f.vector.Expiry.Add(time.Second))
			var rejected *OperationRejection
			switch mode {
			case "operational":
				if !errors.Is(err, fault) {
					t.Fatal("read error became content rejection", err)
				}
			case "stale":
				if !errors.As(err, &rejected) || rejected.Reason != "stale_admission_state" {
					t.Fatal("changed revision published", err)
				}
			default:
				if err != nil {
					t.Fatal("reconcile", err)
				}
			}
			after, err := f.coordinator.GetVersion(ctx, "type_registry.host", "signed-fixture.1")
			if err != nil {
				t.Fatal(err)
			}
			if before.LastVerifiedAt == nil || after.LastVerifiedAt == nil || !before.LastVerifiedAt.Equal(*after.LastVerifiedAt) {
				t.Fatal("startup fabricated a fresh verification")
			}
			if mode == "missing" || mode == "altered" {
				want := "missing"
				if mode == "altered" {
					want = "failed"
				}
				if after.Active || after.Health != want || after.AdministrativelyDisabled {
					t.Fatal("lost imported registry remained effective", after)
				}
				var fallback, invalidation int
				if err := f.pool.QueryRow(ctx, `SELECT count(*) FILTER(WHERE event_kind='safety_fallback'),count(*) FILTER(WHERE event_kind='payload_invalidation') FROM reference_pack_events e JOIN reference_pack_operations o USING(operation_id) WHERE o.kind='reconcile'`).Scan(&fallback, &invalidation); err != nil || fallback != 1 || invalidation != 1 {
					t.Fatal("incomplete atomic startup evidence", fallback, invalidation, err)
				}
				if err := ValidateRequiredState(ctx, f.pool, f.storage, DefaultLimits()); err != nil {
					t.Fatal("unretained invalid former selection blocked recovered Base", err)
				}
			} else if !after.Active || after.Health != before.Health || after.AdministrativelyDisabled {
				t.Fatal("healthy or aborted startup changed selection", after)
			}
		})
	}
}

type reconciliationTestStorage struct {
	ArtifactStorage
	reference  string
	readErr    error
	beforeRead func()
}

func (s *reconciliationTestStorage) OpenPublished(ctx context.Context, ref StorageRef) (ContainerReader, int64, error) {
	if ref.String() == s.reference {
		if s.beforeRead != nil {
			f := s.beforeRead
			s.beforeRead = nil
			f()
		}
		if s.readErr != nil {
			return nil, 0, s.readErr
		}
	}
	return s.ArtifactStorage.OpenPublished(ctx, ref)
}
