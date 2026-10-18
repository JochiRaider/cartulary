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
	"testing"
	"time"
)

func testRendererOwnership(t *testing.T) {
	t.Run("six_slots_and_waiting_seventh", func(t *testing.T) {
		owner := &rendererOwner{options: ProductionOptions{BrowserLaneLimit: 6}, changed: make(chan struct{}), remaining: map[string]int{}}
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
			requests = append(requests, WorkRequirement{Request: Request{Capability: "browser_stack", UnitID: fmt.Sprintf("unit-%d", i), Affinity: fmt.Sprintf("stack-%d", i)}, Renderer: "none", Physical: map[string]map[string]int64{"browser_stack": {"port_lane": 2, "memory_mb": 384, "process": 4}}, Active: map[string]int64{"cpu": 1, "io": 1, "process": 1, "memory_mb": 64}})
		}
		requests[7].Active["port_lane"] = 1
		owner.ConfigureAdmission(filepath.Join(root, "admission"), nil, map[string]int64{"cpu": 2, "io": 2, "memory_mb": 8192, "process": 128, "port_lane": 16}, requests)
		released := 0
		owner.Broker.options.Providers["browser_stack"] = &ProviderFuncs{AcquireFunc: func(_ context.Context, request Request) (*Allocation, error) {
			return ownedAllocation(func(ctx context.Context, _ bool) error {
				released++
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
			if i == 7 && released != 1 {
				t.Fatal("pressure must retire exactly one stack", released)
			}
			lease, err := owner.Acquire(ctx, request.Request)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lease.Release(ctx, true, true); err != nil {
				t.Fatal(err)
			}
			if err := finish(ctx); err != nil {
				t.Fatal(err)
			}
			if err := host.Release(ctx); err != nil {
				t.Fatal(err)
			}
		}
		if err := owner.Close(ctx); err != nil {
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
		owner.renderers.remaining["default"] = 4
		environment := map[string]string{"CARTULARY_WEB_E2E_TLS_TRUST_SHA256": "sha256:" + strings.Repeat("b", 64)}
		first, err := owner.renderers.reserve(t.Context(), "suite-a/group-one", "default", "pooled", false)
		if err != nil {
			t.Fatal(err)
		}
		one, release, err := owner.renderers.borrow(t.Context(), first, environment)
		if err != nil {
			t.Fatal(err)
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
		owner.renderers.remaining["default"] = 2
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
		if retained, err := fourthRelease(t.Context(), true); err != nil || retained {
			t.Fatal(retained, err)
		}
		if len(owner.renderers.slots) != 0 || removals != 3 {
			t.Fatal("last consumer left idle peers", removals)
		}
		owner.renderers.remaining["default"] = 2
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
			slot, err := owner.renderers.reserve(t.Context(), "idle", fmt.Sprintf("profile-%d", index), "pooled", false)
			if err != nil {
				t.Fatal(err)
			}
			owner.renderers.unreserve(slot)
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
