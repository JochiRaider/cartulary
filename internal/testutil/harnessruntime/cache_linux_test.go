//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixtureFile(t *testing.T, name string, data []byte, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, data, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(name, mode); err != nil {
		t.Fatal(err)
	}
}
func nativeCacheFixture(t *testing.T, mode string) (*nativeCache, unit) {
	t.Helper()
	root := t.TempDir()
	run := filepath.Join(root, "results")
	if err := os.Mkdir(run, 0700); err != nil {
		t.Fatal(err)
	}
	u := engineUnit("row:fixture", 1)
	u.CachePolicy = "content_addressed"
	u.Digest = "sha256:" + strings.Repeat("a", 64)
	row := "harness.command_surface.unit.fixture"
	u.Evidence = []string{"rows/" + row + ".json", "unit-results/fixture.json"}
	u.Rows = []rowBinding{{ID: row, Runner: "shell"}}
	u.Artifacts = []artifact{{Type: "file", Path: "bin/program", Destination: "repository_artifact", Mode: "0755", Producer: u.ID}, {Type: "directory", Path: "bundle", Destination: "run_root", Mode: "0750", Producer: u.ID}}
	profile, policy, input, dependency, strategy := "test.fixture", "content_addressed", "sha256:"+strings.Repeat("b", 64), "sha256:"+strings.Repeat("c", 64), "broad"
	fact := cacheFact{UnitID: u.ID, Eligible: true, Profile: &profile, Policy: &policy, InputDigest: &input, DependencyDigest: &dependency, Strategy: &strategy}
	source := []byte("current source")
	writeFixtureFile(t, filepath.Join(root, "input.txt"), source, 0644)
	p := enginePlan(u)
	p.facts.Manifest = map[string]any{"cache_mode": mode}
	p.facts.Cache = []cacheFact{fact}
	p.facts.SourceInputs = []sourceInput{{Path: "input.txt", Kind: "file", Mode: "0644", Digest: bytesDigest(source)}}
	p.facts.CacheInputs = map[string][]int{dependency: {0}}
	c := newNativeCache(p, root, run)
	writeFixtureFile(t, filepath.Join(root, "bin/program"), []byte("program bytes"), 0755)
	writeFixtureFile(t, filepath.Join(run, "bundle/sub/content"), []byte("content bytes"), 0644)
	if err := os.Chmod(filepath.Join(run, "bundle"), 0750); err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{"schema_id": "cartulary.harness_row_result.v4", "row_id": row, "terminal_state": "passed", "duration_ms": 19, "exit_code": 0, "failure_class": nil, "failure_reason": nil, "failure_diagnostic": nil, "runner": "shell", "started_at": "2026-10-11T00:00:00.000Z", "finished_at": "2026-10-11T00:00:00.019Z", "wall_duration_ms": 19}
	bytes, err := documentBytes("cartulary.harness_row_result.v4", payload)
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, filepath.Join(run, u.Evidence[0]), bytes, 0600)
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	})
	return c, u
}
func TestRuntimeCacheTransactions(t *testing.T) {
	for _, phase := range []string{"read", "restore"} {
		t.Run("cancelled_"+phase+"_preserves_cache_and_outputs", func(t *testing.T) {
			c, u := nativeCacheFixture(t, "normal")
			if err := c.Store(context.Background(), u); err != nil {
				t.Fatal(err)
			}
			record := filepath.Join(c.directory(c.facts[u.ID]), "record.json")
			before, err := os.ReadFile(record)
			if err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(c.root, "bin/program")
			writeFixtureFile(t, output, []byte("current output"), 0755)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if phase == "read" {
				c.Fresh = func(context.Context, cacheFact) error { cancel(); return nil }
			} else {
				c.BeforeRename = func(int) error { cancel(); return ctx.Err() }
			}
			if _, err = c.Lookup(ctx, u); !errors.Is(err, context.Canceled) {
				t.Fatal("cache cancellation lost", err)
			}
			after, err := os.ReadFile(record)
			if err != nil || string(after) != string(before) {
				t.Fatal("valid cache quarantined or changed", err)
			}
			data, err := os.ReadFile(output)
			if err != nil || string(data) != "current output" {
				t.Fatal("cancelled restore changed output", err)
			}
		})
	}
	t.Run("cancelled_freshness_is_not_configuration_failure", func(t *testing.T) {
		c, u := nativeCacheFixture(t, "normal")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := c.Lookup(ctx, u); !errors.Is(err, context.Canceled) {
			t.Fatal("cache cancellation lost", err)
		}
	})
	ctx := context.Background()
	t.Run("round_trip_and_fresh_semantic_rows", func(t *testing.T) {
		c, u := nativeCacheFixture(t, "normal")
		if err := c.Store(ctx, u); err != nil {
			t.Fatal(err)
		}
		directory := c.directory(c.facts[u.ID])
		if _, err := os.Stat(filepath.Join(directory, "record.json")); err != nil {
			t.Fatal(err)
		}
		writeFixtureFile(t, filepath.Join(c.root, "bin/program"), []byte("damaged"), 0755)
		if err := os.RemoveAll(filepath.Join(c.run, "bundle")); err != nil {
			t.Fatal(err)
		}
		hit, err := c.Lookup(ctx, u)
		if err != nil || hit.Outcome != "hit" {
			t.Fatalf("%+v %v", hit, err)
		}
		bytes, _ := os.ReadFile(filepath.Join(c.root, "bin/program"))
		if string(bytes) != "program bytes" {
			t.Fatal(string(bytes))
		}
		bytes, _ = os.ReadFile(filepath.Join(c.run, "bundle/sub/content"))
		if string(bytes) != "content bytes" {
			t.Fatal(string(bytes))
		}
		bytes, _ = os.ReadFile(filepath.Join(c.run, u.Evidence[0]))
		var row map[string]any
		if err := json.Unmarshal(bytes, &row); err != nil {
			t.Fatal(err)
		}
		if row["duration_ms"] != float64(0) || row["started_at"] == "2026-10-11T00:00:00.000Z" {
			t.Fatal("historical execution evidence replayed")
		}
	})
	t.Run("same_run_is_private_and_drained", func(t *testing.T) {
		c, u := nativeCacheFixture(t, "normal")
		u.CachePolicy = "same_run"
		f := c.facts[u.ID]
		f.Policy = &u.CachePolicy
		c.facts[u.ID] = f
		if err := c.Store(ctx, u); err != nil {
			t.Fatal(err)
		}
		if len(c.staging) != 1 {
			t.Fatal("missing same-run ownership")
		}
		stage := c.staging[0].name
		hit, err := c.Lookup(ctx, u)
		if err != nil || hit.Outcome != "hit" {
			t.Fatal(hit, err)
		}
		other := newNativeCache(c.plan, c.root, c.run)
		other.facts[u.ID] = f
		miss, err := other.Lookup(ctx, u)
		if err != nil || miss.Outcome != "miss" {
			t.Fatal("same-run entry escaped", miss, err)
		}
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Lstat(stage); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("staging survived close", err)
		}
	})
	t.Run("same_run_replacement_is_not_deleted", func(t *testing.T) {
		c, u := nativeCacheFixture(t, "normal")
		u.CachePolicy = "same_run"
		f := c.facts[u.ID]
		f.Policy = &u.CachePolicy
		c.facts[u.ID] = f
		if err := c.Store(ctx, u); err != nil {
			t.Fatal(err)
		}
		stage := c.staging[0].name
		if err := os.Rename(stage, stage+"-original"); err != nil {
			t.Fatal(err)
		}
		writeFixtureFile(t, filepath.Join(stage, "sentinel"), []byte("foreign"), 0600)
		if err := c.Close(); err == nil {
			t.Fatal("replaced stage accepted")
		}
		if bytes, err := os.ReadFile(filepath.Join(stage, "sentinel")); err != nil || string(bytes) != "foreign" {
			t.Fatal("replacement removed", err)
		}
		if err := os.RemoveAll(stage); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(stage+"-original", stage); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("off_skips_all_cache_input_work", func(t *testing.T) {
		c, u := nativeCacheFixture(t, "off")
		c.Fresh = func(context.Context, cacheFact) error {
			t.Error("off inspected inputs")
			return errors.New("unexpected")
		}
		hit, err := c.Lookup(ctx, u)
		if err != nil || hit.Reason != "mode_off" {
			t.Fatal(hit, err)
		}
		if err := c.Store(ctx, u); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Lstat(c.base()); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("off created cache")
		}
	})
	t.Run("cold_reads_bypassed_writes_allowed", func(t *testing.T) {
		c, u := nativeCacheFixture(t, "cold")
		if err := c.Store(ctx, u); err != nil {
			t.Fatal(err)
		}
		hit, err := c.Lookup(ctx, u)
		if err != nil || hit.Reason != "cold_read_bypass" {
			t.Fatal(hit, err)
		}
	})
	t.Run("input_change_blocks_reuse", func(t *testing.T) {
		c, u := nativeCacheFixture(t, "normal")
		if err := c.Store(ctx, u); err != nil {
			t.Fatal(err)
		}
		writeFixtureFile(t, filepath.Join(c.root, "input.txt"), []byte("changed source"), 0644)
		if _, err := c.Lookup(ctx, u); err == nil {
			t.Fatal("stale input reused")
		}
	})
	t.Run("concurrent_valid_entry_is_immutable", func(t *testing.T) {
		c, u := nativeCacheFixture(t, "normal")
		if err := c.Store(ctx, u); err != nil {
			t.Fatal(err)
		}
		name := filepath.Join(c.directory(c.facts[u.ID]), "record.json")
		before, _ := os.ReadFile(name)
		writeFixtureFile(t, filepath.Join(c.root, "bin/program"), []byte("another valid output"), 0755)
		if err := c.Store(ctx, u); err != nil {
			t.Fatal(err)
		}
		after, _ := os.ReadFile(name)
		if string(before) != string(after) {
			t.Fatal("valid concurrent entry replaced")
		}
	})
	for _, corruption := range []string{"identity", "digest", "duplicate", "surplus", "symlink"} {
		t.Run("reject_"+corruption, func(t *testing.T) {
			c, u := nativeCacheFixture(t, "normal")
			if err := c.Store(ctx, u); err != nil {
				t.Fatal(err)
			}
			directory := c.directory(c.facts[u.ID])
			recordPath := filepath.Join(directory, "record.json")
			bytes, err := os.ReadFile(recordPath)
			if err != nil {
				t.Fatal(err)
			}
			switch corruption {
			case "identity":
				bytes = []byte(strings.Replace(string(bytes), `"unit_id":"row:fixture"`, `"unit_id":"foreign"`, 1))
			case "digest":
				bytes = []byte(strings.Replace(string(bytes), `"output_digest":"sha256:`, `"output_digest":"sha256:0`, 1))
			case "duplicate":
				bytes = []byte(strings.Replace(string(bytes), `"schema_id":`, `"schema_id":"cartulary.harness_cache_record.v2","schema_id":`, 1))
			case "surplus":
				writeFixtureFile(t, filepath.Join(directory, "artifacts/0/surplus"), nil, 0600)
			case "symlink":
				name := filepath.Join(directory, "artifacts/0/0")
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(c.root, "bin/program"), name); err != nil {
					t.Fatal(err)
				}
			}
			if corruption == "identity" || corruption == "digest" || corruption == "duplicate" {
				writeFixtureFile(t, recordPath, bytes, 0600)
			}
			hit, err := c.Lookup(ctx, u)
			if err != nil || hit.Reason != "record_invalid" {
				t.Fatalf("%+v %v", hit, err)
			}
			entries, _ := filepath.Glob(filepath.Join(filepath.Dir(directory), ".quarantine-*"))
			if len(entries) != 1 {
				t.Fatal("failed evidence was not quarantined")
			}
		})
	}
}
func TestRuntimePublicationRollback(t *testing.T) {
	t.Run("parent_replacement", func(t *testing.T) {
		base := t.TempDir()
		run := filepath.Join(base, "run")
		moved := filepath.Join(base, "moved")
		writeFixtureFile(t, filepath.Join(run, "a"), []byte("original"), 0600)
		err := publishOutputs(context.Background(), base, run, []publishItem{{Type: "file", Path: "a", Destination: "run_root", Mode: "0600", Bytes: []byte("new")}}, func(int) error {
			if err := os.Rename(run, moved); err != nil {
				return err
			}
			writeFixtureFile(t, filepath.Join(run, "sentinel"), []byte("foreign"), 0600)
			return nil
		})
		if err == nil {
			t.Fatal("replaced parent accepted")
		}
		bytes, _ := os.ReadFile(filepath.Join(run, "sentinel"))
		if string(bytes) != "foreign" {
			t.Fatal("foreign parent modified")
		}
		bytes, _ = os.ReadFile(filepath.Join(moved, "a"))
		if string(bytes) != "original" {
			t.Fatal("original output changed")
		}
	})
	t.Run("replacement_is_not_rollback_authority", func(t *testing.T) {
		run := t.TempDir()
		writeFixtureFile(t, filepath.Join(run, "a"), []byte("original"), 0600)
		items := []publishItem{{Type: "file", Path: "a", Destination: "run_root", Mode: "0600", Bytes: []byte("new")}, {Type: "file", Path: "b", Destination: "run_root", Mode: "0600", Bytes: []byte("new-b")}}
		err := publishOutputs(context.Background(), run, run, items, func(index int) error {
			if index == 1 {
				if err := os.Rename(filepath.Join(run, "a"), filepath.Join(run, "moved")); err != nil {
					return err
				}
				writeFixtureFile(t, filepath.Join(run, "a"), []byte("replacement"), 0600)
				return errors.New("publication interrupted")
			}
			return nil
		})
		if !errors.Is(err, errIncompletePublication) {
			t.Fatal("missing incomplete rollback classification", err)
		}
		bytes, _ := os.ReadFile(filepath.Join(run, "a"))
		if string(bytes) != "replacement" {
			t.Fatal("replacement was deleted")
		}
		backups, _ := filepath.Glob(filepath.Join(run, ".cartulary-cache-backup-*"))
		if len(backups) != 1 {
			t.Fatal("recovery backup lost")
		}
		bytes, _ = os.ReadFile(backups[0])
		if string(bytes) != "original" {
			t.Fatal("original backup lost")
		}
	})
	root := t.TempDir()
	run := t.TempDir()
	writeFixtureFile(t, filepath.Join(run, "a"), []byte("old-a"), 0600)
	writeFixtureFile(t, filepath.Join(run, "b"), []byte("old-b"), 0600)
	items := []publishItem{{Type: "file", Path: "a", Destination: "run_root", Mode: "0600", Bytes: []byte("new-a")}, {Type: "file", Path: "b", Destination: "run_root", Mode: "0600", Bytes: []byte("new-b")}}
	err := publishOutputs(context.Background(), root, run, items, func(index int) error {
		if index == 1 {
			return errors.New("injected disk failure")
		}
		return nil
	})
	if err == nil {
		t.Fatal("failure not propagated")
	}
	for _, name := range []string{"a", "b"} {
		bytes, _ := os.ReadFile(filepath.Join(run, name))
		if string(bytes) != "old-"+name {
			t.Fatal("partial publication", name, string(bytes))
		}
	}
	entries, _ := os.ReadDir(run)
	if len(entries) != 2 {
		t.Fatal("staging leaked", entries)
	}
	if err := os.Remove(filepath.Join(run, "b")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "external"), filepath.Join(run, "b")); err != nil {
		t.Fatal(err)
	}
	if err := publishOutputs(context.Background(), root, run, items, nil); err == nil {
		t.Fatal("symlink destination accepted")
	}
	bytes, _ := os.ReadFile(filepath.Join(run, "a"))
	if string(bytes) != "old-a" {
		t.Fatal("unsafe second output changed first")
	}
}
