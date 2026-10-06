package reference_data_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/jackc/pgx/v5"
)

type faultingArtifactStorage struct {
	reference_data.ArtifactStorage
	readFailure bool
}

func (s faultingArtifactStorage) OpenPublished(ctx context.Context, ref reference_data.StorageRef) (reference_data.ContainerReader, int64, error) {
	if !s.readFailure {
		return nil, 0, fs.ErrPermission
	}
	reader, size, err := s.ArtifactStorage.OpenPublished(ctx, ref)
	if err != nil {
		return nil, 0, err
	}
	return failedArtifactReader{reader}, size, nil
}

type failedArtifactReader struct{ reference_data.ContainerReader }

func (failedArtifactReader) ReadAt([]byte, int64) (int, error) { return 0, io.ErrClosedPipe }

func TestConsumerLossPublishesInvalidationAndAbortsSourceTransaction_Integration(t *testing.T) {
	ctx := context.Background()
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "reference-pack-consumer-integrity")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	storage, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	codec := pagination.NewCodec([32]byte([]byte(strings.Repeat("i", 32))))
	consumer, err := reference_data.NewConsumer(harness.Pool, storage, codec, time.Now, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	assignments, err := reference_data.NewRegistryAssignments(harness.Pool, storage, codec, time.Now, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	base := consumer.ResolveCurrentPackSet(ctx)
	if base.Error != nil {
		t.Fatal(base.Error)
	}
	for index, tc := range []struct {
		name, health         string
		corrupt, transaction bool
	}{
		{name: "absent authoritative member", health: "missing"},
		{name: "present truncated member", health: "failed", corrupt: true},
		{name: "loss inside guarded source mutation", health: "missing", transaction: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			version := fmt.Sprint(index + 1)
			importReferencePack(t, harness, admin, "type_registry.host", version, "integrity-import-"+version)
			requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/type_registry.host/"+version+"/activate", "integrity-activate-"+version, ""), http.StatusOK)
			set := consumer.ResolveCurrentPackSet(ctx)
			if set.Error != nil {
				t.Fatal(set.Error)
			}
			var originalProvenance []byte
			if err := harness.Pool.QueryRow(ctx, `SELECT canonical_provenance FROM reference_pack_sets WHERE pack_set_id=$1`, set.Value.ID).Scan(&originalProvenance); err != nil {
				t.Fatal(err)
			}
			if index == 0 {
				for _, readFailure := range []bool{false, true} {
					faulted, err := reference_data.NewConsumer(harness.Pool, faultingArtifactStorage{storage, readFailure}, codec, time.Now, consumerIntegrityOptions(t))
					if err != nil {
						t.Fatal(err)
					}
					result := faulted.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: set.Value.ID, PackKey: "type_registry.host", EntryID: "unknown"})
					if result.Error == nil || result.Error.Code != "pack_unavailable" {
						t.Fatal("operational read failure did not fail closed")
					}
					var healthy bool
					if err := harness.Pool.QueryRow(ctx, `SELECT health='verified_available' AND last_failure_code IS NULL FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version=$1`, version).Scan(&healthy); err != nil || !healthy {
						t.Fatal("operational failure condemned retained content", err)
					}
					if current := consumer.ResolveCurrentPackSet(ctx); current.Error != nil || current.Value.ID != set.Value.ID {
						t.Fatal("operational failure changed the effective set")
					}
					if count := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_events WHERE pack_key='type_registry.host' AND pack_version=$1 AND event_kind='payload_invalidation'`, version); count != 0 {
						t.Fatal("operational failure published an invalidity attestation")
					}
				}
			}
			if tc.corrupt {
				if err := os.WriteFile(storedBundlePath(t, harness, "type_registry.host", version), []byte("{}\n"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				removeStoredBundle(t, harness, "type_registry.host", version)
			}
			reader := consumer
			var tx pgx.Tx
			if tc.transaction {
				tx, err = harness.Pool.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback(ctx)
				reader, err = assignments.BeginTx(ctx, tx)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `UPDATE users SET display_name='must roll back after content loss' WHERE is_deployment_admin`); err != nil {
					t.Fatal(err)
				}
			}
			request := reference_data.GetPackEntryRequest{PackSetID: set.Value.ID, PackKey: "type_registry.host", EntryID: "unknown"}
			result := reader.GetPackEntry(ctx, request)
			if result.Error == nil || result.Error.Code != "pack_unavailable" {
				t.Fatal("lost content was consumed", result)
			}
			if tc.transaction {
				if err := tx.Commit(ctx); !errors.Is(err, pgx.ErrTxClosed) {
					t.Fatal("source transaction could commit after integrity loss", err)
				}
				var modified bool
				if err := harness.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE display_name='must roll back after content loss')`).Scan(&modified); err != nil || modified {
					t.Fatal("source write escaped rollback", err)
				}
			}
			var health string
			var missing *string
			if err := harness.Pool.QueryRow(ctx, `SELECT health,missing_reason FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version=$1`, version).Scan(&health, &missing); err != nil {
				t.Fatal(err)
			}
			if health != tc.health || (missing != nil) != (health == "missing") || missing != nil && *missing != "storage_loss" {
				t.Fatal("wrong classification", health, missing)
			}
			current := consumer.ResolveCurrentPackSet(ctx)
			if current.Error != nil || current.Value.ID != base.Value.ID {
				t.Fatal("invalidation did not commit Base fallback", current)
			}
			var first []byte
			if err := harness.Pool.QueryRow(ctx, `SELECT canonical_attestation FROM reference_pack_events WHERE event_kind='payload_invalidation' AND pack_key='type_registry.host' AND pack_version=$1`, version).Scan(&first); err != nil {
				t.Fatal(err)
			}
			for range 2 {
				if result := consumer.GetPackEntry(ctx, request); result.Error == nil {
					t.Fatal("invalidated set was substituted")
				}
			}
			var count int
			var retained []byte
			if err := harness.Pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_events WHERE event_kind='payload_invalidation' AND pack_key='type_registry.host' AND pack_version=$1`, version).Scan(&count); err != nil || count != 1 {
				t.Fatal("duplicate detection attestation", count, err)
			}
			if err := harness.Pool.QueryRow(ctx, `SELECT canonical_attestation FROM reference_pack_events WHERE event_kind='payload_invalidation' AND pack_key='type_registry.host' AND pack_version=$1`, version).Scan(&retained); err != nil || !bytes.Equal(first, retained) {
				t.Fatal("first detection changed", err)
			}
			if err := harness.Pool.QueryRow(ctx, `SELECT canonical_provenance FROM reference_pack_sets WHERE pack_set_id=$1`, set.Value.ID).Scan(&retained); err != nil || !bytes.Equal(originalProvenance, retained) {
				t.Fatal("historical provenance changed", err)
			}
			for _, kind := range []string{"payload_invalidation", "safety_fallback"} {
				if err := harness.Pool.QueryRow(ctx, `SELECT count(*) FROM administrative_audit_projections p JOIN reference_pack_events e ON p.target_id=e.operation_id::text WHERE e.event_kind='payload_invalidation' AND e.pack_key='type_registry.host' AND e.pack_version=$1 AND p.action_code=$2 AND p.actor_kind='system' AND p.actor_user_id IS NULL AND p.source='system'`, version, "reference_pack_"+kind).Scan(&count); err != nil || count != 1 {
					t.Fatal("first detection audit not atomic or idempotent", kind, count, err)
				}
			}
		})
	}
}

func TestRequiredBaseLossClearsSelectionAndBlocksReadiness_Integration(t *testing.T) {
	ctx := context.Background()
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "reference-pack-base-loss")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	storage, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	codec := pagination.NewCodec([32]byte([]byte(strings.Repeat("b", 32))))
	consumer, err := reference_data.NewConsumer(harness.Pool, storage, codec, time.Now, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	base := consumer.ResolveCurrentPackSet(ctx)
	if base.Error != nil {
		t.Fatal(base.Error)
	}
	ready := func(want int) {
		t.Helper()
		response, err := http.Get(harness.Server.HTTP.URL + "/readyz")
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != want {
			t.Fatalf("readiness = %d, want %d", response.StatusCode, want)
		}
	}
	ready(http.StatusOK)
	// A second historical loss must remain attributable after the current set
	// is already absent. Repeated observations must not multiply attestations.
	for _, key := range []string{"type_registry.host", "type_registry.evidence"} {
		version := ""
		for _, member := range base.Value.Members {
			if member.Key == key {
				version = member.Version
			}
		}
		if version == "" {
			t.Fatal("missing Base fixture")
		}
		removeStoredBundle(t, harness, key, version)
		for range 2 {
			result := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: base.Value.ID, PackKey: key, EntryID: "unknown"})
			if result.Error == nil || result.Error.Code != "pack_unavailable" {
				t.Fatal("lost built-in remained usable", result)
			}
		}
		if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2 AND health='missing' AND missing_reason='storage_loss'`, key, version); got != 1 {
			t.Fatal("definitive loss rolled back")
		}
		if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_events WHERE event_kind='payload_invalidation' AND pack_key=$1 AND pack_version=$2 AND convert_from(canonical_attestation,'UTF8')::jsonb->'resulting_pack_set_id'='null'::jsonb`, key, version); got != 1 {
			t.Fatal("loss attestation not durable and idempotent", got)
		}
		ready(http.StatusServiceUnavailable)
	}
	if result := consumer.ResolveCurrentPackSet(ctx); result.Error == nil || result.Error.Code != "pack_unavailable" {
		t.Fatal("incomplete current selection was exposed")
	}
	if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_current_set WHERE pack_set_id IS NULL`) != 1 {
		t.Fatal("incomplete set was retained as current")
	}
	if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_sets WHERE pack_set_id=$1`, base.Value.ID) != 1 {
		t.Fatal("immutable historical set erased")
	}
	if err := reference_data.ValidateRequiredState(ctx, harness.Pool, storage, reference_data.DefaultLimits()); err == nil {
		t.Fatal("startup validation accepted missing current selection")
	}
	requireReasonError(t, postAction(t, harness, admin, "/api/v1/reference-packs/refresh", "blocked-base-loss", ""), http.StatusConflict, "reference_pack_operation_rejected", "required_registry_gap")
	assignments, err := reference_data.NewRegistryAssignments(harness.Pool, storage, codec, time.Now, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	tx, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err := assignments.BeginTx(ctx, tx); err == nil {
		t.Fatal("new registry mutation admitted without a current set")
	}
}

func TestUnavailableBasePreservesHealthyReplacement_Integration(t *testing.T) {
	ctx := context.Background()
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "reference-pack-inactive-base-loss")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	storage, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	consumer, err := reference_data.NewConsumer(harness.Pool, storage, pagination.NewCodec([32]byte([]byte(strings.Repeat("h", 32)))), time.Now, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	base := consumer.ResolveCurrentPackSet(ctx)
	if base.Error != nil {
		t.Fatal(base.Error)
	}
	version := ""
	for _, member := range base.Value.Members {
		if member.Key == "type_registry.host" {
			version = member.Version
		}
	}
	importReferencePack(t, harness, admin, "type_registry.host", "1", "healthy-replacement-import")
	requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/type_registry.host/1/activate", "healthy-replacement-activate", ""), http.StatusOK)
	current := consumer.ResolveCurrentPackSet(ctx)
	if current.Error != nil {
		t.Fatal(current.Error)
	}
	removeStoredBundle(t, harness, "type_registry.host", version)
	result := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: base.Value.ID, PackKey: "type_registry.host", EntryID: "unknown"})
	if result.Error == nil {
		t.Fatal("lost historical Base remained usable")
	}
	if got := consumer.ResolveCurrentPackSet(ctx); got.Error != nil || got.Value.ID != current.Value.ID {
		t.Fatal("inactive Base loss erased healthy replacement")
	}
	if err := reference_data.CheckReadyState(ctx, harness.Pool); err != nil {
		t.Fatal("healthy replacement blocked readiness", err)
	}
	requireReasonError(t, postAction(t, harness, admin, "/api/v1/reference-packs/type_registry.host/1/disable", "unsafe-disable", ""), http.StatusConflict, "reference_pack_operation_rejected", "required_registry_gap")
	if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version='1' AND health='verified_available' AND NOT administratively_disabled`) != 1 {
		t.Fatal("rejected disablement changed state")
	}
	if got := consumer.ResolveCurrentPackSet(ctx); got.Error != nil || got.Value.ID != current.Value.ID {
		t.Fatal("rejected disablement cleared current set")
	}
}
