//go:build linux

package harnesssession

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testActiveResourceAdmission(t *testing.T) {
	t.Helper()
	t.Run("package_deployment_workspace_is_bound_to_exact_native_private_allocation", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		const token = "11111111-2222-4333-8444-555555555555"
		workspace, err := r.PrivatePath("package-qualification", token)
		if err != nil {
			t.Fatal(err)
		}
		for _, scope := range []string{"source", "destination"} {
			for _, family := range []string{"container", "volume", "network"} {
				target := map[string]any{"daemon_id": "fixture-engine", "token": token, "scope": scope, "family": family}
				if err := r.RecordDependentResource("package_deployment", target, "pending", []string{}); err != nil {
					t.Fatal(err)
				}
				if err := r.RecordResource("package_deployment", target, "released"); err != nil {
					t.Fatal(err)
				}
			}
		}
		for _, candidate := range []string{workspace, filepath.Join(r.Root, "operator"), r.Root, filepath.Join(r.Root, "recovery", token), filepath.Join(filepath.Dir(workspace), "foreign"), "relative", workspace + "/../" + token} {
			target := map[string]any{"daemon_id": "fixture-engine", "token": token, "scope": "shared", "family": "workspace", "workspace": candidate, "helper_image": "sha256:" + strings.Repeat("a", 64)}
			err := r.RecordDependentResource("package_deployment", target, "pending", []string{})
			if candidate == workspace {
				if err != nil {
					t.Fatal(err)
				}
				if err := r.RecordResource("package_deployment", target, "released"); err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("foreign workspace admitted", candidate)
			}
		}
		resources, err := r.Resources(64)
		if err != nil || len(resources) != 0 {
			t.Fatal(resources, err)
		}
		if _, err := os.Stat(workspace); !os.IsNotExist(err) {
			t.Fatal("proof registration created a workspace", err)
		}
	})
	t.Run("active_resource_lookup_is_exact_nonmutating_and_current", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		target := json.RawMessage(`{"daemon_id":"fixture-engine","token":"11111111-2222-4333-8444-555555555555","role":"application"}`)
		selected := Resource{Kind: "package_resource", Target: target, State: "pending", RunID: r.RunID, LeaseID: r.LeaseID, Dependencies: []string{}}
		if _, err := r.InspectResource(selected); err == nil {
			t.Fatal("unregistered identity admitted")
		}
		if _, err := os.Stat(filepath.Join(r.Root, "recovery")); !os.IsNotExist(err) {
			t.Fatal("lookup created a journal", err)
		}
		if err := r.RecordDependentResource(selected.Kind, selected.Target, "pending", []string{}); err != nil {
			t.Fatal(err)
		}
		defer r.RecordResource(selected.Kind, selected.Target, "released")
		entries, err := os.ReadDir(filepath.Join(r.Root, "recovery"))
		if err != nil || len(entries) != 1 {
			t.Fatal(entries, err)
		}
		file := filepath.Join(r.Root, "recovery", entries[0].Name())
		for _, state := range []string{"pending", "acquired"} {
			if err := r.RecordDependentResource(selected.Kind, selected.Target, state, []string{}); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			admitted, err := r.InspectResource(selected)
			if err != nil || admitted.Reference() != selected.Reference() || admitted.State != state || admitted.Dependencies == nil || len(admitted.Dependencies) != 0 {
				t.Fatal("incorrect admission", admitted, err)
			}
			after, err := os.ReadFile(file)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("lookup changed proof", err)
			}
		}
		for _, change := range []func(*Resource){
			func(v *Resource) { v.RunID = "foreign" }, func(v *Resource) { v.LeaseID = "00000000-0000-4000-8000-000000000000" }, func(v *Resource) { v.State = "released" },
			func(v *Resource) {
				v.Target = json.RawMessage(`{"daemon_id":"fixture-engine","token":"11111111-2222-4333-8444-666666666666","role":"application"}`)
			},
		} {
			candidate := selected
			change(&candidate)
			if _, err := r.InspectResource(candidate); err == nil {
				t.Fatal("unowned lookup admitted", candidate)
			}
		}
		if err := r.RecordResource(selected.Kind, selected.Target, "released"); err != nil {
			t.Fatal(err)
		}
		if _, err := r.InspectResource(selected); err == nil {
			t.Fatal("released identity admitted")
		}
	})
	t.Run("active_resource_lookup_rejects_invalid_stored_authority", func(t *testing.T) {
		for _, scenario := range []string{"malformed", "duplicate", "foreign_run", "foreign_target", "released", "extra", "symlink", "hardlink", "oversize"} {
			t.Run(scenario, func(t *testing.T) {
				r, _, _ := runtimeFixture(t)
				target := json.RawMessage(`{"daemon_id":"fixture-engine","token":"11111111-2222-4333-8444-555555555555","role":"application"}`)
				selected := Resource{Kind: "package_resource", Target: target, State: "pending", RunID: r.RunID, LeaseID: r.LeaseID, Dependencies: []string{}}
				if err := r.RecordDependentResource(selected.Kind, selected.Target, "pending", []string{}); err != nil {
					t.Fatal(err)
				}
				defer r.RecordResource(selected.Kind, selected.Target, "released")
				entries, err := os.ReadDir(filepath.Join(r.Root, "recovery"))
				if err != nil || len(entries) != 1 {
					t.Fatal(entries, err)
				}
				file := filepath.Join(r.Root, "recovery", entries[0].Name())
				original, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := os.Remove(file); err != nil {
						t.Error(err)
					}
					if err := os.WriteFile(file, original, 0600); err != nil {
						t.Error(err)
					}
				}()
				var value map[string]any
				if err := json.Unmarshal(original, &value); err != nil {
					t.Fatal(err)
				}
				switch scenario {
				case "malformed":
					err = os.WriteFile(file, []byte("{"), 0600)
				case "duplicate":
					err = os.WriteFile(file, append([]byte(`{"state":"acquired",`), original[1:]...), 0600)
				case "foreign_run":
					value["run_id"] = "foreign"
				case "foreign_target":
					value["target"].(map[string]any)["role"] = "postgres"
				case "released":
					value["state"] = "released"
				case "extra":
					value["unsafe"] = true
				case "symlink", "hardlink":
					source := filepath.Join(t.TempDir(), "proof.json")
					if err = os.WriteFile(source, original, 0600); err != nil {
						t.Fatal(err)
					}
					if err = os.Remove(file); err != nil {
						t.Fatal(err)
					}
					if scenario == "symlink" {
						err = os.Symlink(source, file)
					} else {
						err = os.Link(source, file)
					}
				case "oversize":
					err = os.Truncate(file, MaximumFrame+1)
				}
				if err != nil {
					t.Fatal(err)
				}
				if scenario == "foreign_run" || scenario == "foreign_target" || scenario == "released" || scenario == "extra" {
					changed, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					if err = os.WriteFile(file, changed, 0600); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := r.InspectResource(selected); err == nil {
					t.Fatal("invalid stored authority admitted")
				}
			})
		}
	})
}
