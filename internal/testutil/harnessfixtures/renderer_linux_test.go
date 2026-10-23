//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func testRendererOwnership(t *testing.T) {
	t.Run("unstarted_profile_changes_reuse_bounded_logical_ownership", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		owner := NewProduction(ProductionOptions{BrowserLaneLimit: 1, Root: root, RunRoot: run, Runtime: runtime}, Options{})
		t.Cleanup(func() {
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := closeFixtureConsumers(owner, cleanup); err != nil {
				t.Error(err)
			}
		})
		owner.ConfigureAdmission(filepath.Join(root, "admission"), nil, map[string]int64{"cpu": 2, "memory_mb": 512, "process": 8})
		first, err := owner.renderers.reserve(t.Context(), "first", "default", "pooled", false)
		if err != nil {
			t.Fatal(err)
		}
		record, err := owner.Options.residency.ensure("renderer:"+first.allocation, map[string]int64{"process": 1})
		if err != nil {
			t.Fatal(err)
		}
		first.resident = record
		owner.renderers.unreserve(first)
		for i := range 64 {
			profile, policy, diagnostics := "default", "pooled", false
			if i%2 == 0 {
				profile, policy, diagnostics = "network_flow_claimed", "dedicated", true
			}
			slot, err := owner.renderers.reserve(t.Context(), fmt.Sprintf("waiter-%d", i), profile, policy, diagnostics)
			if err != nil {
				t.Fatal(err)
			}
			if slot.allocation != first.allocation || slot.generation != first.generation || slot.resident != record || slot.profile != profile || slot.policy != policy || slot.diagnostics != diagnostics {
				t.Fatal("unstarted reassignment replaced ownership or retained an incompatible preparation")
			}
			owner.renderers.unreserve(slot)
		}
		if _, err := os.Stat(filepath.Join(run, "_shared", "renderers")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("logical reassignment published physical lifecycle evidence", err)
		}
		if _, err := owner.Options.residency.ensure("renderer:"+first.allocation, map[string]int64{"process": 2}); err == nil {
			t.Fatal("reassignment changed immutable residency quantities")
		}
		if err := closeFixtureConsumers(owner, t.Context()); err != nil {
			t.Fatal(err)
		}
		resources, err := runtime.Resources(4096)
		if err != nil || len(resources) != 0 {
			t.Fatal("unstarted ownership did not settle", resources, err)
		}
	})
	pressure := func(t *testing.T, interrupted bool, policy string) {
		runtime, root, run := runtimeFixture(t)
		owner := NewProduction(ProductionOptions{BrowserLaneLimit: 2, Root: root, RunRoot: run, Runtime: runtime}, Options{})
		admissionRoot := filepath.Join(root, "admission")
		requests := []WorkRequirement{}
		for i := range 2 {
			request := WorkRequirement{Request: Request{PostgresIsolation: "exclusive", Services: []string{"postgres", "object_store"}, Capability: "browser_stack", UnitID: fmt.Sprintf("seed-%d", i), Affinity: fmt.Sprintf("stack-%d", i), RuntimeProfile: "default"}, Renderer: "none", Physical: map[string]map[string]int64{"browser_stack": {"port_lane": 2, "memory_mb": 128, "process": 1}, "renderer": {"process": 1}}, Active: map[string]int64{"cpu": 1, "io": 1, "memory_mb": 64, "process": 1}}
			requests = append(requests, request)
			request.Request.UnitID = fmt.Sprintf("waiter-%d", i)
			request.Renderer = policy
			request.Active = map[string]int64{"cpu": 1, "io": 1, "memory_mb": 64, "process": 1, "port_lane": 1}
			requests = append(requests, request)
		}
		owner.ConfigureAdmission(admissionRoot, nil, map[string]int64{"cpu": 2, "io": 2, "memory_mb": 4096, "process": 32, "port_lane": 4})
		registerFixtureConsumers(t, owner, requests)
		var started, released atomic.Int32
		owner.Broker.options.Providers["browser_stack"] = &ProviderFuncs{AcquireFunc: func(_ context.Context, request Request) (*Allocation, error) {
			started.Add(1)
			return ownedAllocation(func(ctx context.Context, _ bool) error {
				released.Add(1)
				return owner.Options.residency.close(ctx, stackKey(request))
			}), nil
		}}
		t.Cleanup(func() {
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := closeFixtureConsumers(owner, cleanup); err != nil {
				t.Error("pressure fixture cleanup", err)
			}
		})
		for _, request := range []WorkRequirement{requests[0], requests[2]} {
			host, finish, err := owner.Admit(t.Context(), request)
			if err != nil {
				t.Fatal(err)
			}
			lease, err := owner.Acquire(t.Context(), request.Request)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lease.Release(t.Context(), true); err != nil {
				t.Fatal(err)
			}
			if err := errors.Join(finish(t.Context()), host.Release(t.Context())); err != nil {
				t.Fatal(err)
			}
		}
		// Hold the existing host mutex until both provisional stack pins exist.
		// This establishes the deadlock precondition without scheduler timing.
		mutex, err := os.OpenFile(filepath.Join(admissionRoot, "mutex"), os.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer mutex.Close()
		if err := syscall.Flock(int(mutex.Fd()), syscall.LOCK_EX); err != nil {
			t.Fatal(err)
		}
		defer syscall.Flock(int(mutex.Fd()), syscall.LOCK_UN)
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		results := make(chan error, 2)
		for _, request := range []WorkRequirement{requests[1], requests[3]} {
			go func() {
				host, finish, err := owner.Admit(ctx, request)
				if err != nil {
					results <- err
					return
				}
				lease, err := owner.Broker.Acquire(WithActivity(ctx, host.Token), request.Request)
				if err == nil {
					_, err = lease.Release(WithActivity(ctx, host.Token), true, false)
				}
				results <- errors.Join(err, finish(ctx), host.Release(ctx))
			}()
		}
		condition := time.NewTicker(time.Millisecond)
		defer condition.Stop()
		bothPinned := false
		for !bothPinned && ctx.Err() == nil {
			owner.Broker.mu.Lock()
			bothPinned = owner.Broker.reserved[sharedKey(requests[1].Request)] == 1 && owner.Broker.reserved[sharedKey(requests[3].Request)] == 1
			owner.Broker.mu.Unlock()
			if !bothPinned {
				select {
				case <-ctx.Done():
				case <-condition.C:
				}
			}
		}
		if interrupted {
			cancel()
		}
		if err := syscall.Flock(int(mutex.Fd()), syscall.LOCK_UN); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			err := <-results
			if interrupted && !errors.Is(err, context.Canceled) {
				t.Error("admission cancellation was not preserved", err)
			} else if !interrupted && err != nil {
				t.Error("pooled admission did not progress", err)
			}
		}
		if !bothPinned {
			t.Fatal("failed to establish concurrent provisional ownership")
		}
		owner.Broker.mu.Lock()
		pins := len(owner.Broker.reserved)
		owner.Broker.mu.Unlock()
		if pins != 0 {
			t.Fatal("provisional pins leaked", pins)
		}
		cleanup, cleanupDone := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupDone()
		if err := closeFixtureConsumers(owner, cleanup); err != nil {
			t.Fatal("pressure settlement", err)
		}
		if (!interrupted && started.Load() < 3) || started.Load() != released.Load() {
			t.Fatal("retirement must reacquire fresh state and settle all physical allocations", started.Load(), released.Load())
		}
	}
	for _, policy := range []string{"pooled", "dedicated"} {
		t.Run(policy+"_waiters_yield_idle_stack_pins_under_pressure", func(t *testing.T) { pressure(t, false, policy) })
		t.Run("cancelled_"+policy+"_waiters_release_provisional_pins", func(t *testing.T) { pressure(t, true, policy) })
	}
	t.Run("six_slots_and_waiting_seventh", func(t *testing.T) {
		owner := &rendererOwner{options: ProductionOptions{BrowserLaneLimit: 6}, changed: make(chan struct{})}
		for i := 0; i < 6; i++ {
			slot, err := owner.reserve(t.Context(), fmt.Sprintf("borrower-%d", i), "default", "pooled", false)
			if err != nil || slot.id != i+1 {
				t.Fatal("six-slot reservation", slot, err)
			}
		}
		ctx, cancel := context.WithCancel(t.Context())
		started := make(chan struct{})
		result := make(chan error, 1)
		go func() {
			close(started)
			_, err := owner.reserve(ctx, "seventh", "default", "pooled", false)
			result <- err
		}()
		<-started
		select {
		case err := <-result:
			t.Fatal("seventh borrower did not wait", err)
		default:
		}
		cancel()
		if err := <-result; !errors.Is(err, context.Canceled) {
			t.Fatal("waiting borrower ignored cancellation", err)
		}
		owner.unreserve(owner.slots[0])
		slot, err := owner.reserve(t.Context(), "replacement", "default", "pooled", false)
		if err != nil || slot.id != 1 || len(owner.slots) != 6 {
			t.Fatal("slot was not reused", slot, err)
		}
	})
	t.Run("resident_stack_pressure_progresses_after_renderers_are_gone", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		owner := NewProduction(ProductionOptions{BrowserLaneLimit: 4, Root: root, RunRoot: run, Runtime: runtime}, Options{})
		requests := []WorkRequirement{}
		for i := range 8 {
			requests = append(requests, WorkRequirement{Request: Request{PostgresIsolation: "exclusive", Services: []string{"object_store", "postgres"}, Capability: "browser_stack", UnitID: fmt.Sprintf("unit-%d", i), Affinity: fmt.Sprintf("stack-%d", i)}, Renderer: "none", Physical: map[string]map[string]int64{"browser_stack": {"port_lane": 2, "memory_mb": 384, "process": 4}}, Active: map[string]int64{"cpu": 1, "io": 1, "process": 1, "memory_mb": 64}})
		}
		requests[7].Active["port_lane"] = 1
		owner.ConfigureAdmission(filepath.Join(root, "admission"), nil, map[string]int64{"cpu": 2, "io": 2, "memory_mb": 8192, "process": 128, "port_lane": 16})
		future := slices.Clone(requests)
		for i := range future {
			future[i].Request.UnitID += "-future"
		}
		registerFixtureConsumers(t, owner, append(slices.Clone(requests), future...))
		released := 0
		releasedKeys := []string{}
		owner.Broker.options.Providers["browser_stack"] = &ProviderFuncs{AcquireFunc: func(_ context.Context, request Request) (*Allocation, error) {
			return ownedAllocation(func(ctx context.Context, _ bool) error {
				released++
				releasedKeys = append(releasedKeys, sharedKey(request))
				return owner.Options.residency.close(ctx, stackKey(request))
			}), nil
		}}
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		for i, request := range requests {
			host, finish, err := owner.Admit(ctx, request)
			if err != nil {
				t.Fatal("admission stalled with only idle stacks resident", i, err)
			}
			if i == 7 {
				keys := []string{}
				for _, prior := range requests[:7] {
					keys = append(keys, sharedKey(prior.Request))
				}
				if released != 1 || releasedKeys[0] != slices.Min(keys) {
					t.Fatal("pressure must retire one stable logical key, independent of allocation UUIDs", releasedKeys)
				}
			}
			lease, err := owner.Acquire(ctx, request.Request)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lease.Release(ctx, true); err != nil {
				t.Fatal(err)
			}
			if err := finish(ctx); err != nil {
				t.Fatal(err)
			}
			if err := host.Release(ctx); err != nil {
				t.Fatal(err)
			}
		}
		if err := closeFixtureConsumers(owner, ctx); err != nil {
			t.Fatal(err)
		}
		if released != 8 {
			t.Fatal("physical cleanup incomplete", released)
		}
	})
	t.Run("renderer_generations_borrows_and_quarantine", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		cwd, _ := os.Getwd()
		profileBytes, err := os.ReadFile(filepath.Join(cwd, "../../../tools/frontend_visual_renderer_profile.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(root, "tools"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "tools/frontend_visual_renderer_profile.json"), profileBytes, 0600); err != nil {
			t.Fatal(err)
		}
		profile := map[string]any{}
		if json.Unmarshal(profileBytes, &profile) != nil {
			t.Fatal("invalid profile")
		}
		starts, settles, removals := 0, 0, 0
		var adapterMu sync.Mutex
		unhealthy := false
		options := ProductionOptions{BrowserLaneLimit: 4, Root: root, RunRoot: run, Runtime: runtime, Environment: map[string]string{"NODE_BIN": "fake-node"}}
		options.Run = func(_ context.Context, _ string, args []string, env map[string]string) (CommandResult, error) {
			adapterMu.Lock()
			defer adapterMu.Unlock()
			switch args[1] {
			case "settle":
				settles++
				if unhealthy {
					return CommandResult{Exit: 12, Stdout: []byte(`{"failure_class":"harness","failure_reason":"fixture_error"}`)}, errors.New("leftover client")
				}
				return CommandResult{Stdout: []byte(`{"server_alive":true,"client_connections_absent":true,"browser_processes_absent":true,"forwarding_listeners_absent":true}`)}, nil
			case "release":
				removals++
				return CommandResult{Stdout: []byte(`{}`)}, nil
			case "acquire":
				starts++
				token := fmt.Sprintf("%032x", starts)
				proof := map[string]any{"daemon_id": "fixture", "image_id": "sha256:" + strings.Repeat("a", 64), "name": "cartulary-visual-123-" + token[:12], "token": token}
				dependencies := []string{}
				if json.Unmarshal([]byte(env["CARTULARY_HARNESS_RESOURCE_DEPENDENCIES"]), &dependencies) != nil {
					t.Fatal("invalid dependencies")
				}
				if err := runtime.RecordDependentResource("browser_renderer", proof, "acquired", dependencies); err != nil {
					t.Fatal(err)
				}
				observed := map[string]any{"image_id": proof["image_id"], "core_version": profile["playwright_version"], "font_files": []any{map[string]any{"path": "fixture.woff2", "bytes": 1, "sha256": strings.Repeat("a", 64)}}}
				for _, key := range []string{"container_image", "platform", "playwright_version", "chromium_revision", "chromium_version", "font_manifest_sha256"} {
					observed[key] = profile[key]
				}
				physical := map[string]any{"schema_id": "cartulary.harness_renderer_physical.v1", "run_id": runtime.RunID, "proof": proof, "container_id": strings.Repeat("a", 64), "diagnostic_binding": nil, "environment": map[string]string{"CARTULARY_VISUAL_RENDERER_ATTESTED": "1", "CARTULARY_VISUAL_RENDERER_PROFILE_ID": profile["profile_id"].(string), "CARTULARY_VISUAL_RENDERER_PROFILE_JSON": string(profileBytes), "CARTULARY_VISUAL_RENDERER_WS_ENDPOINT": "ws://127.0.0.1:1234/private"}, "attestation": map[string]any{"schema_id": "cartulary.frontend_visual_renderer_attestation.v1", "profile": profile, "observed": observed}, "trust": map[string]any{"schema_id": "cartulary.browser_tls_trust.v1", "tool_sha256": strings.Repeat("a", 64), "trust_bundle_sha256": strings.Repeat("b", 64), "authorities_sha256": []string{strings.Repeat("c", 64)}}}
				if err := privateJSON(args[2], physical["schema_id"].(string), physical, false); err != nil {
					t.Fatal(err)
				}
				data, _ := json.Marshal(physical)
				return CommandResult{Stdout: data}, nil
			default:
				t.Fatal(args)
				return CommandResult{}, errors.New("unexpected operation")
			}
		}
		owner := NewProduction(options, Options{})
		owner.renderers.remaining = func(rendererConsumerKey, string) int { return 3 }
		clusters := []string{}
		for _, name := range []string{"cluster-a", "cluster-b"} {
			file := filepath.Join(runtime.Root, name+".json")
			if err := runtime.RecordDependentResource("managed_suite", file, "acquired", []string{}); err != nil {
				t.Fatal(err)
			}
			clusters = append(clusters, file)
			t.Cleanup(func() { _ = runtime.RecordResource("managed_suite", file, "released") })
		}
		environment := map[string]string{"CARTULARY_WEB_E2E_TLS_TRUST_SHA256": "sha256:" + strings.Repeat("b", 64)}
		first, err := owner.renderers.reserve(t.Context(), "suite-a/group-one", "default", "pooled", false)
		if err != nil {
			t.Fatal(err)
		}
		one, release, err := owner.renderers.borrow(t.Context(), first, environment)
		if err != nil {
			t.Fatal(err)
		}
		// Retiring either independent cluster must leave the active renderer
		// untouched. Its admitted trust is an owned immutable copy, not a live
		// service dependency. Exercise the ordinary/recovery cleanup evaluator.
		for _, cluster := range clusters {
			resources, err := runtime.Resources(4096)
			if err != nil {
				t.Fatal(err)
			}
			ref := fixtureReference(t, runtime, "managed_suite", cluster)
			selected := []string{}
			if err := evaluateCleanup(resources, ref, func(resource harnesssession.Resource) error {
				selected = append(selected, resource.Reference())
				return nil
			}, nil); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(selected, []string{ref}) {
				t.Fatal("cluster retirement reached a live renderer", selected)
			}
			if err := runtime.RecordResource("managed_suite", cluster, "released"); err != nil {
				t.Fatal(err)
			}
		}
		if retained, err := release(t.Context(), true); err != nil || !retained {
			t.Fatal(retained, err)
		}
		if retained, err := release(t.Context(), true); err != nil || !retained || settles != 1 {
			t.Fatal("duplicate settlement", retained, err, settles)
		}
		second, err := owner.renderers.reserve(t.Context(), "suite-b/group-two", "default", "pooled", false)
		if err != nil || second != first {
			t.Fatal("compatible suites did not reuse generation", err)
		}
		if _, _, err := owner.renderers.borrow(t.Context(), second, map[string]string{"CARTULARY_WEB_E2E_TLS_TRUST_SHA256": "sha256:" + strings.Repeat("f", 64)}); err == nil || starts != 1 {
			t.Fatal("foreign trust accepted or repaired a shared generation", err)
		}
		two, release, err := owner.renderers.borrow(t.Context(), second, environment)
		if err != nil {
			t.Fatal(err)
		}
		if one["CARTULARY_RENDERER_ATTACHMENT"] == two["CARTULARY_RENDERER_ATTACHMENT"] || starts != 1 {
			t.Fatal("borrow or physical lifetime collapsed")
		}
		unhealthy = true
		if retained, err := release(t.Context(), true); err != nil || retained || removals != 1 {
			t.Fatal("uncertain client did not retire allocation", retained, err, removals)
		}
		third, err := owner.renderers.reserve(t.Context(), "suite-c/group-three", "default", "pooled", false)
		if err != nil || third.generation == first.generation {
			t.Fatal("quarantined generation reused", err)
		}
		foreign, err := owner.renderers.reserve(t.Context(), "foreign/group", "foreign", "pooled", false)
		if err != nil || foreign == third {
			t.Fatal("incompatible profile shared a slot", err)
		}
		dedicated, err := owner.renderers.reserve(t.Context(), "dedicated/group", "default", "dedicated", true)
		if err != nil {
			t.Fatal(err)
		}
		fourth, err := owner.renderers.reserve(t.Context(), "fourth/group", "default", "pooled", false)
		if err != nil {
			t.Fatal(err)
		}
		cancelled, cancel := context.WithCancel(t.Context())
		cancel()
		if _, err := owner.renderers.reserve(cancelled, "overflow", "default", "pooled", false); !errors.Is(err, context.Canceled) {
			t.Fatal("slot exhaustion bypassed cancellation", err)
		}
		for _, slot := range []*rendererSlot{foreign, dedicated} {
			if err := owner.renderers.retire(t.Context(), slot); err != nil {
				t.Fatal(err)
			}
		}
		unhealthy = false
		owner.renderers.remaining = func(rendererConsumerKey, string) int { return 1 }
		_, thirdRelease, err := owner.renderers.borrow(t.Context(), third, environment)
		if err != nil {
			t.Fatal(err)
		}
		_, fourthRelease, err := owner.renderers.borrow(t.Context(), fourth, environment)
		if err != nil {
			t.Fatal(err)
		}
		if retained, err := thirdRelease(t.Context(), true); err != nil || !retained {
			t.Fatal(retained, err)
		}
		owner.renderers.remaining = func(rendererConsumerKey, string) int { return 0 }
		if retained, err := fourthRelease(t.Context(), true); err != nil || retained {
			t.Fatal(retained, err)
		}
		if _, err := owner.renderers.retireIdle(t.Context(), nil, true, nil, 0, "terminal_borrower_release"); err != nil {
			t.Fatal(err)
		}
		if len(owner.renderers.slots) != 0 || removals != 3 {
			t.Fatal("last consumer left idle peers", removals)
		}
		owner.renderers.remaining = func(rendererConsumerKey, string) int { return 0 }
		releases := []func(context.Context, bool) (bool, error){}
		for _, unit := range []string{"concurrent/one", "concurrent/two"} {
			slot, err := owner.renderers.reserve(t.Context(), unit, "default", "pooled", false)
			if err != nil {
				t.Fatal(err)
			}
			_, release, err := owner.renderers.borrow(t.Context(), slot, environment)
			if err != nil {
				t.Fatal(err)
			}
			releases = append(releases, release)
		}
		finished := make(chan error, 2)
		for _, release := range releases {
			go func() { _, err := release(t.Context(), true); finished <- err }()
		}
		for range releases {
			if err := <-finished; err != nil {
				t.Fatal(err)
			}
		}
		if len(owner.renderers.slots) != 0 || removals != 5 {
			t.Fatal("concurrent last consumers leaked a slot", removals)
		}
		idle := []*rendererSlot{}
		for index := range 4 {
			profile := fmt.Sprintf("profile-%d", index)
			owner.renderers.remaining = func(rendererConsumerKey, string) int { return 1 }
			slot, err := owner.renderers.reserve(t.Context(), "idle", profile, "pooled", false)
			if err != nil {
				t.Fatal(err)
			}
			_, release, err := owner.renderers.borrow(t.Context(), slot, environment)
			if err != nil {
				t.Fatal(err)
			}
			if retained, err := release(t.Context(), true); err != nil || !retained {
				t.Fatal("physical slot was not retained", retained, err)
			}
			idle = append(idle, slot)
		}
		newSlot, err := owner.renderers.reserve(t.Context(), "different", "different-profile", "pooled", false)
		if err != nil || len(owner.renderers.slots) != 4 || newSlot.id != idle[0].id {
			t.Fatal("slot pressure did not replace exactly one stable slot", err)
		}
		for _, retained := range idle[1:] {
			if !slices.Contains(owner.renderers.slots, retained) {
				t.Fatal("slot pressure retired an unnecessary peer")
			}
		}
		owner.renderers.unreserve(newSlot)
		if err := owner.renderers.evictIdle(t.Context(), nil); err != nil {
			t.Fatal(err)
		}
		resources, err := runtime.Resources(4096)
		if err != nil || len(resources) != 0 {
			t.Fatal("physical ownership leaked", resources, err)
		}
	})
}

func registerFixtureConsumers(t *testing.T, p *Production, works []WorkRequirement) {
	t.Helper()
	if err := p.RegisterConsumers(works); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := closeFixtureConsumers(p, context.Background()); err != nil {
			t.Error(err)
		}
	})
}
func closeFixtureConsumers(p *Production, ctx context.Context) error {
	p.consumers.mu.Lock()
	ids := []string{}
	for id, value := range p.consumers.units {
		if value.disposition == "" {
			ids = append(ids, id)
		}
	}
	p.consumers.mu.Unlock()
	for _, id := range ids {
		if err := p.ConsumerTerminal(id, "cancelled"); err != nil {
			return err
		}
	}
	return p.Close(ctx)
}
