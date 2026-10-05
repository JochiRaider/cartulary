package reference_data

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepackfixtures"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type lifecycleFixtureStep struct {
	ID          string  `json:"fixture_id"`
	Operation   string  `json:"operation"`
	Transaction string  `json:"client_txn_id"`
	Input       *string `json:"input_ref"`
	Active      *bool   `json:"expected_active"`
	Disabled    *bool   `json:"expected_disabled"`
	Removed     *bool   `json:"expected_removed"`
	Status      *string `json:"expected_job_status"`
	Replayed    bool    `json:"expected_replayed"`
}

type lifecycleFixtureExpectation struct {
	ID        string   `json:"fixture_id"`
	Manifest  *string  `json:"expected_manifest_sha256"`
	Payload   *string  `json:"expected_payload_sha256"`
	Set       *string  `json:"expected_pack_set_sha256"`
	Condition *string  `json:"expected_condition"`
	Effects   []string `json:"expected_side_effects"`
	Forbidden []string `json:"forbidden_side_effects"`
}

// The canonical scenario and its manifest expectations are authored inputs.
// UUIDs, physical storage identities and SQL row order are intentionally not
// fixture identity. Logical set/digest expectations were computed in Python.
func TestCanonicalLifecycleFixtureManifests_Integration(t *testing.T) {
	f := newCanonicalCoordinatorFixture(t, "reference-pack-canonical-lifecycle")
	// This fresh fixture has no source records. Replacement still invokes the
	// participating owner's port under the real transaction/usage guard.
	f.coordinator.registryUsage = lifecycleFixtureUsage{}
	ctx := context.Background()
	var scenario struct {
		Schema string                 `json:"schema_id"`
		Steps  []lifecycleFixtureStep `json:"steps"`
	}
	raw, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/lifecycle.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&scenario); err != nil || scenario.Schema != "cartulary.reference_pack_lifecycle_fixture.v1" {
		t.Fatal("invalid scenario", err)
	}
	manifests := map[string]lifecycleFixtureExpectation{}
	for _, artifact := range contractreferencepackfixtures.Artifacts {
		if !strings.HasPrefix(artifact.Path, "contracts/reference-pack-fixtures/fixture-manifests/lifecycle/") {
			continue
		}
		var expected lifecycleFixtureExpectation
		if err := json.Unmarshal([]byte(artifact.JSON), &expected); err != nil {
			t.Fatal(err)
		}
		if _, exists := manifests[expected.ID]; exists {
			t.Fatal("duplicate lifecycle manifest")
		}
		manifests[expected.ID] = expected
	}
	if len(manifests) != len(scenario.Steps) || len(manifests) != 14 {
		t.Fatal("missing lifecycle execution or manifest")
	}
	observer := &lifecycleFixtureObserver{t: t}
	f.coordinator.observer = observer
	network := &lifecycleFixtureNetwork{}
	previousTransport := http.DefaultTransport
	http.DefaultTransport = network
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	for index, step := range scenario.Steps {
		if !t.Run(step.ID, func(t *testing.T) {
			expected, ok := manifests[step.ID]
			if !ok {
				t.Fatal("unregistered scenario")
			}
			delete(manifests, step.ID)
			*f.now = f.vector.At.Add(time.Duration(index) * time.Minute)
			before := captureLifecycleFixture(t, f, network)
			replayed := false
			var accepted JobAcceptedResult
			switch step.Operation {
			case "reconcile_claimed", "reconcile_unclaimed":
				err = ReconcileBaseRelease(ctx, f.pool, f.storage, BaseReleaseOptions{ProfileClaimed: step.Operation == "reconcile_claimed", ClockTrusted: true, Limits: DefaultLimits(), Observer: observer}, *f.now)
			case "import", "renew":
				data := f.container
				if step.Input != nil {
					// The closed scenario permits only this independently produced archive.
					if *step.Input != "contracts/reference-pack-fixtures/fixtures/admission/equivalent_tar.v1.json" {
						t.Fatal("unknown lifecycle input")
					}
					raw, e := os.ReadFile("../../../" + *step.Input)
					if e != nil {
						t.Fatal(e)
					}
					var archive struct {
						Bytes string `json:"bytes_base64"`
					}
					if e := json.Unmarshal(raw, &archive); e != nil {
						t.Fatal(e)
					}
					data, e = base64.StdEncoding.Strict().DecodeString(archive.Bytes)
					if e != nil {
						t.Fatal(e)
					}
				}
				accepted, err = f.coordinator.Import(ctx, f.actor, step.Transaction, bytes.NewReader(data))
				replayed = accepted.Replayed
			case "reverify", "refresh", "empty_refresh", "cancel_reverify":
				kind := step.Operation
				key := "type_registry.host"
				version := "signed-fixture.1"
				if kind == "cancel_reverify" {
					kind = "reverify"
				}
				if kind == "empty_refresh" {
					kind = "refresh"
					key = "framework.attack"
				}
				if kind == "refresh" {
					version = ""
				}
				accepted, err = f.coordinator.VerifyRetained(ctx, VerificationRequest{Kind: kind, ActorUserID: f.actor, ClientTxnID: step.Transaction, PackKeys: []string{key}, KeysProvided: true, PackVersion: version})
			case "activate", "disable", "remove":
				raw, _ := json.Marshal(map[string]any{"client_txn_id": step.Transaction, "reason": "private lifecycle fixture reason"})
				request, apiErr := DecodeActionRequest(bytes.NewReader(raw))
				if apiErr != nil {
					t.Fatal(apiErr)
				}
				p := ActionParams{ActorUserID: f.actor, PackKey: "type_registry.host", PackVersion: "signed-fixture.1", Request: request, Now: *f.now}
				var result ActionResult
				switch step.Operation {
				case "activate":
					result, err = f.coordinator.Activate(ctx, p)
				case "disable":
					result, err = f.coordinator.Disable(ctx, p)
				case "remove":
					result, err = f.coordinator.Remove(ctx, p)
				}
				replayed = result.Replayed
			default:
				t.Fatal("unknown scenario operation")
			}
			if err != nil {
				t.Fatal("lifecycle operation", err)
			}
			if replayed != step.Replayed {
				t.Fatal("wrong replay disposition")
			}
			if step.Status != nil {
				id, e := uuid.Parse(accepted.Job.JobID)
				if e != nil {
					t.Fatal(e)
				}
				if !replayed {
					if step.Operation == "cancel_reverify" {
						_, e = f.manager.Cancel(ctx, jobs.CancelParams{JobID: id, ActorUserID: f.actor, ClientTxnID: "cancel-canonical", NormalizedRequest: []byte(`{"client_txn_id":"cancel-canonical"}`)})
						if e != nil {
							t.Fatal(e)
						}
					}
					// Admitted asynchronous work has not published a set or usable index.
					admitted := captureLifecycleFixture(t, f, network)
					if admitted.set != before.set || admitted.index != before.index {
						t.Fatal("admission published verification outputs")
					}
					execution, claimed, e := f.manager.Claim(ctx, id)
					if e != nil || !claimed {
						t.Fatal("claim", e)
					}
					if e := f.coordinator.Execute(ctx, execution); e != nil {
						t.Fatal("execute", e)
					}
				}
				job, e := f.manager.Get(ctx, id)
				if e != nil || string(job.Status) != *step.Status {
					t.Fatal("terminal disposition", job.Status, job.ErrorSummary, e)
				}
			}
			after := captureLifecycleFixture(t, f, network)
			if expected.Condition == nil {
				if after.version != nil {
					t.Fatal("reconciliation fabricated operator candidate")
				}
			} else {
				v := after.version
				if v == nil || v.Condition != *expected.Condition || v.Active != *step.Active || v.AdministrativelyDisabled != *step.Disabled || v.Removed != *step.Removed || v.PendingWork {
					t.Fatal("administrative projection differs", v)
				}
				if expected.Manifest == nil || expected.Payload == nil || v.ManifestSHA256 == nil || v.PayloadSHA256 == nil || *v.ManifestSHA256 != *expected.Manifest || *v.PayloadSHA256 != *expected.Payload {
					t.Fatal("logical identity changed")
				}
			}
			if expected.Set == nil || after.set != "rpset_"+*expected.Set {
				t.Fatal("exact set identity changed", after.set)
			}
			effects := map[string]bool{
				"job_admitted": after.jobs > before.jobs, "candidate_created": after.candidates > before.candidates,
				"condition_changed": after.condition != before.condition, "attestation_appended": after.events > before.events,
				"audit_event_appended": after.audits > before.audits, "trust_state_changed": after.trust != before.trust,
				"index_published": after.index != "" && after.index != before.index, "active_pointer_changed": after.set != before.set,
				"pack_set_published": after.set != before.set, "payload_bytes_deleted": after.objects < before.objects,
				"network_request_attempted": after.network > before.network,
			}
			for _, token := range expected.Effects {
				if got, ok := effects[token]; !ok || !got {
					t.Fatal("missing expected side effect", token)
				}
			}
			for _, token := range expected.Forbidden {
				if got, ok := effects[token]; !ok || got {
					t.Fatal("forbidden or unobserved side effect", token)
				}
			}
			if len(f.storage.staged) != 0 {
				t.Fatal("retained staging after terminal operation")
			}
		}) {
			break
		}
	}
	if len(manifests) != 0 {
		t.Fatal("unexecuted lifecycle fixtures", len(manifests))
	}
	if observer.completed == 0 || network.calls.Load() != 0 {
		t.Fatal("missing observation or attempted network fetch")
	}
}

type lifecycleFixtureSnapshot struct {
	jobs, candidates, events, audits, objects int
	set, index, condition, trust              string
	version                                   *AdministrativeVersion
	network                                   int64
}

func captureLifecycleFixture(t *testing.T, f canonicalCoordinatorFixture, network *lifecycleFixtureNetwork) lifecycleFixtureSnapshot {
	t.Helper()
	ctx := context.Background()
	s := lifecycleFixtureSnapshot{objects: len(f.storage.objects), network: network.calls.Load()}
	err := f.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM jobs),(SELECT count(*) FROM reference_pack_candidates),(SELECT count(*) FROM reference_pack_events),(SELECT count(*) FROM administrative_audit_projections),(SELECT pack_set_id FROM reference_pack_current_set WHERE singleton),coalesce((SELECT current_index_id::text FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version='signed-fixture.1'),'')`).Scan(&s.jobs, &s.candidates, &s.events, &s.audits, &s.set, &s.index)
	if err != nil {
		t.Fatal(err)
	}
	err = f.pool.QueryRow(ctx, `SELECT jsonb_build_object('roots',(SELECT jsonb_agg(jsonb_build_array(repository_id,root_version,sha256) ORDER BY repository_id,root_version) FROM reference_pack_roots),'metadata',(SELECT jsonb_agg(jsonb_build_array(repository_id,pack_key,pack_version,role,metadata_version,encode(canonical_bytes,'hex')) ORDER BY repository_id,pack_key,pack_version,role) FROM reference_pack_metadata_versions))::text`).Scan(&s.trust)
	if err != nil {
		t.Fatal(err)
	}
	v, err := f.coordinator.GetVersion(ctx, "type_registry.host", "signed-fixture.1")
	if err == nil {
		s.version = &v
		s.condition = v.Condition
	} else if !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	return s
}

type lifecycleFixtureNetwork struct{ calls atomic.Int64 }

type lifecycleFixtureUsage struct{}

func (lifecycleFixtureUsage) ReferencedRegistryEntriesTx(context.Context, pgx.Tx, string) ([]string, error) {
	return []string{}, nil
}

func (n *lifecycleFixtureNetwork) RoundTrip(*http.Request) (*http.Response, error) {
	n.calls.Add(1)
	return nil, errors.New("fixture denies outbound HTTP")
}

type lifecycleFixtureObserver struct {
	t         *testing.T
	completed int
}

func (o *lifecycleFixtureObserver) BeginReferenceOperation(ctx context.Context, operation string) (context.Context, func(string)) {
	if !slices.Contains([]string{"reference_pack.import", "reference_pack.verify", "reference_pack.reverify", "reference_pack.activate", "reference_pack.disable", "reference_pack.refresh", "reference_pack.reconcile", "reference_pack.remove"}, operation) {
		o.t.Error("unregistered observer operation")
	}
	return ctx, func(result string) {
		if !slices.Contains([]string{"success", "rejected", "conflict", "canceled", "failed", "timeout"}, result) {
			o.t.Error("unregistered observer result")
		}
		o.completed++
	}
}
