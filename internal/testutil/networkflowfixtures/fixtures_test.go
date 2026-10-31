package networkflowfixtures

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const testFixtureID = "NF-FIX-999-admission"

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func write(t *testing.T, path string, data []byte) {
	t.Helper()
	requireNoError(t, os.MkdirAll(filepath.Dir(path), 0700))
	requireNoError(t, os.WriteFile(path, data, 0600))
}
func encode(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	requireNoError(t, err)
	return data
}
func read(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	requireNoError(t, err)
	return data
}

// The synthetic corpus isolates admission tests from product scenarios and keeps
// malformed bytes under test control. Digests are authored from the frame rule.
func fixture(t *testing.T) (Config, map[string]any) {
	t.Helper()
	config := Config{CorpusRoot: t.TempDir(), RunRoot: t.TempDir(), FixtureID: testFixtureID, Revision: 3}
	t.Cleanup(func() {
		_ = filepath.WalkDir(config.RunRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return os.Chmod(path, 0700)
			}
			return nil
		})
	})
	dir := filepath.Join(config.CorpusRoot, config.FixtureID)
	data := map[string][]byte{
		"source/bytes.csv":          []byte("a,b\r\n1,2\r\n"),
		"source/test.scenario.json": encode(t, map[string]any{"schema_id": scenarioSchema, "fixture_id": testFixtureID, "summary": "Byte-faithful admission test."}),
		"expected/output.json":      []byte("{\"count\":1}\n"),
		"transcripts/result.json":   []byte("{\"status\":\"accepted\"}\n"),
	}
	file := func(path string) map[string]any {
		write(t, filepath.Join(dir, path), data[path])
		sum := sha256.Sum256(data[path])
		return map[string]any{"logical_path": path, "media_type": "application/octet-stream", "size_bytes": len(data[path]), "sha256": hex.EncodeToString(sum[:])}
	}
	a, b, c, d := file("source/bytes.csv"), file("source/test.scenario.json"), file("expected/output.json"), file("transcripts/result.json")
	a["role"], a["newline_policy"] = "input", "byte_exact"
	b["role"], b["newline_policy"] = "support", "lf_required"
	c["role"], d["transcript_kind"] = "expected_response", "fixture_validation"
	frame := func(entries ...map[string]any) string {
		var input strings.Builder
		for _, entry := range entries {
			fmt.Fprintf(&input, "%s\x00%s\x00%d\n", entry["logical_path"], entry["sha256"], entry["size_bytes"])
		}
		sum := sha256.Sum256([]byte(input.String()))
		return hex.EncodeToString(sum[:])
	}
	m := map[string]any{"schema_id": manifestSchema, "manifest_version": 2, "fixture_id": testFixtureID, "profile_id": "network_flow_activity", "freeze": map[string]any{"status": "frozen", "revision": 3, "change_policy": "new_fixture_revision_required"}, "source_files": []any{a, b}, "expected_artifacts": []any{c}, "transcript_files": []any{d}, "source_bundle_sha256": frame(a, b), "expected_bundle_sha256": frame(c, d)}
	write(t, filepath.Join(dir, "manifest.json"), encode(t, m))
	return config, m
}

func TestFixtureAdmission(t *testing.T) {
	mutations := map[string]func(*testing.T, Config, map[string]any){
		"unknown manifest field": func(t *testing.T, c Config, m map[string]any) { m["instructions"] = "ignored" },
		"unknown nested field":   func(t *testing.T, c Config, m map[string]any) { m["freeze"].(map[string]any)["extra"] = true },
		"unknown file field": func(t *testing.T, c Config, m map[string]any) {
			m["source_files"].([]any)[0].(map[string]any)["extra"] = true
		},
		"draft":            func(t *testing.T, c Config, m map[string]any) { m["freeze"].(map[string]any)["status"] = "draft" },
		"revision":         func(t *testing.T, c Config, m map[string]any) { m["freeze"].(map[string]any)["revision"] = 2 },
		"fixture identity": func(t *testing.T, c Config, m map[string]any) { m["fixture_id"] = "NF-FIX-999-other" },
		"unsorted": func(t *testing.T, c Config, m map[string]any) {
			a := m["source_files"].([]any)
			a[0], a[1] = a[1], a[0]
		},
		"duplicate": func(t *testing.T, c Config, m map[string]any) { a := m["source_files"].([]any); a[1] = a[0] },
		"absolute": func(t *testing.T, c Config, m map[string]any) {
			m["source_files"].([]any)[0].(map[string]any)["logical_path"] = "/tmp/source.csv"
		},
		"traversal": func(t *testing.T, c Config, m map[string]any) {
			m["source_files"].([]any)[0].(map[string]any)["logical_path"] = "source/../bytes.csv"
		},
		"dot segment": func(t *testing.T, c Config, m map[string]any) {
			m["source_files"].([]any)[0].(map[string]any)["logical_path"] = "source/./bytes.csv"
		},
		"backslash": func(t *testing.T, c Config, m map[string]any) {
			m["source_files"].([]any)[0].(map[string]any)["logical_path"] = `source\bytes.csv`
		},
		"source bundle":   func(t *testing.T, c Config, m map[string]any) { m["source_bundle_sha256"] = strings.Repeat("0", 64) },
		"expected bundle": func(t *testing.T, c Config, m map[string]any) { m["expected_bundle_sha256"] = strings.Repeat("0", 64) },
		"missing": func(t *testing.T, c Config, m map[string]any) {
			requireNoError(t, os.Remove(filepath.Join(c.CorpusRoot, c.FixtureID, "source/bytes.csv")))
		},
		"extra": func(t *testing.T, c Config, m map[string]any) {
			write(t, filepath.Join(c.CorpusRoot, c.FixtureID, "source/extra"), []byte("x"))
		},
		"changed hash": func(t *testing.T, c Config, m map[string]any) {
			write(t, filepath.Join(c.CorpusRoot, c.FixtureID, "source/bytes.csv"), []byte("a,b\r\n2,1\r\n"))
		},
		"changed size": func(t *testing.T, c Config, m map[string]any) {
			write(t, filepath.Join(c.CorpusRoot, c.FixtureID, "source/bytes.csv"), []byte("a,b\n"))
		},
		"file symlink": func(t *testing.T, c Config, m map[string]any) {
			p := filepath.Join(c.CorpusRoot, c.FixtureID, "source/bytes.csv")
			requireNoError(t, os.Remove(p))
			requireNoError(t, os.Symlink("test.scenario.json", p))
		},
		"directory symlink": func(t *testing.T, c Config, m map[string]any) {
			dir := filepath.Join(c.CorpusRoot, c.FixtureID)
			requireNoError(t, os.Rename(filepath.Join(dir, "source"), filepath.Join(c.CorpusRoot, "saved")))
			requireNoError(t, os.Symlink(filepath.Join(c.CorpusRoot, "saved"), filepath.Join(dir, "source")))
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			config, manifest := fixture(t)
			mutate(t, config, manifest)
			write(t, filepath.Join(config.CorpusRoot, config.FixtureID, "manifest.json"), encode(t, manifest))
			called := false
			path, err := Run(config, func(*Bundle) error { called = true; return nil })
			if err == nil || called {
				t.Fatalf("invalid admission err=%v callback=%v", err, called)
			}
			var summary Summary
			requireNoError(t, json.Unmarshal(read(t, path), &summary))
			want := "failed"
			if name == "draft" {
				want = "blocked"
				if !errors.Is(err, ErrDraft) {
					t.Fatal(err)
				}
			}
			if summary.Status != want || summary.FailurePhase == nil || *summary.FailurePhase != "admission" {
				t.Fatalf("wrong failure evidence: %+v", summary)
			}
		})
	}
	t.Run("duplicate JSON member", func(t *testing.T) {
		config, _ := fixture(t)
		p := filepath.Join(config.CorpusRoot, config.FixtureID, "manifest.json")
		write(t, p, bytes.Replace(read(t, p), []byte(`"manifest_version":2`), []byte(`"manifest_version":2,"manifest_version":2`), 1))
		_, err := Run(config, func(*Bundle) error { t.Fatal("callback after duplicate key"); return nil })
		if err == nil {
			t.Fatal("duplicate admitted")
		}
	})
	t.Run("manifest symlink", func(t *testing.T) {
		config, _ := fixture(t)
		p := filepath.Join(config.CorpusRoot, config.FixtureID, "manifest.json")
		requireNoError(t, os.Rename(p, p+".saved"))
		requireNoError(t, os.Symlink(p+".saved", p))
		_, err := Run(config, func(*Bundle) error { t.Fatal("callback after manifest symlink"); return nil })
		if err == nil {
			t.Fatal("symlink admitted")
		}
	})
	for name, scenario := range map[string]string{
		"scenario unknown field":  `{"schema_id":"cartulary.network_flow_fixture_scenario.v2","fixture_id":"NF-FIX-999-admission","summary":"valid","selector":"forbidden"}`,
		"scenario wrong identity": `{"schema_id":"cartulary.network_flow_fixture_scenario.v2","fixture_id":"NF-FIX-998-other","summary":"valid"}`,
		"scenario duplicate key":  `{"schema_id":"cartulary.network_flow_fixture_scenario.v2","fixture_id":"NF-FIX-999-admission","summary":"first","summary":"second"}`,
	} {
		t.Run(name, func(t *testing.T) {
			config, manifest := fixture(t)
			dir := filepath.Join(config.CorpusRoot, config.FixtureID)
			write(t, filepath.Join(dir, "source/test.scenario.json"), []byte(scenario))
			entry := manifest["source_files"].([]any)[1].(map[string]any)
			sum := sha256.Sum256([]byte(scenario))
			entry["sha256"], entry["size_bytes"] = hex.EncodeToString(sum[:]), len(scenario)
			var frames strings.Builder
			for _, raw := range manifest["source_files"].([]any) {
				e := raw.(map[string]any)
				fmt.Fprintf(&frames, "%s\x00%s\x00%d\n", e["logical_path"], e["sha256"], e["size_bytes"])
			}
			sum = sha256.Sum256([]byte(frames.String()))
			manifest["source_bundle_sha256"] = hex.EncodeToString(sum[:])
			write(t, filepath.Join(dir, "manifest.json"), encode(t, manifest))
			_, err := Run(config, func(*Bundle) error { t.Fatal("callback after invalid scenario"); return nil })
			if err == nil {
				t.Fatal("scenario admitted")
			}
		})
	}
}

func TestFixtureExecution(t *testing.T) {
	t.Run("verified bytes and evidence", func(t *testing.T) {
		config, _ := fixture(t)
		path, err := Run(config, func(b *Bundle) error {
			original := filepath.Join(config.CorpusRoot, config.FixtureID, "source/bytes.csv")
			write(t, original, []byte("changed after admission"))
			data, err := b.Read("source/bytes.csv")
			if err != nil {
				return err
			}
			if string(data) != "a,b\r\n1,2\r\n" {
				return errors.New("materialized bytes changed")
			}
			input, err := b.InputPath("source/bytes.csv")
			if err != nil {
				return err
			}
			if !strings.HasPrefix(input, b.inputRoot+string(filepath.Separator)) || !bytes.Equal(read(t, input), data) {
				return errors.New("file reader did not receive sealed input bytes")
			}
			for _, unsafe := range []string{"../manifest.json", original, "source/../bytes.csv", "source/unlisted.csv"} {
				if _, err := b.InputPath(unsafe); err == nil {
					return errors.New("unlisted file reader path accepted")
				}
			}
			if _, err := b.Read("../manifest.json"); err == nil {
				return errors.New("unlisted read accepted")
			}
			requireNoError(t, filepath.WalkDir(b.inputRoot, func(p string, e fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				info, err := e.Info()
				if err != nil {
					return err
				}
				if info.Mode().Perm()&0222 != 0 {
					return errors.New("writable materialization")
				}
				return nil
			}))
			return b.Record("actual.json", []byte("{\"count\":1}\n"))
		})
		requireNoError(t, err)
		var s Summary
		requireNoError(t, json.Unmarshal(read(t, path), &s))
		requireNoError(t, validateDocument(read(t, path), summarySchema))
		if s.Status != "passed" || s.FailurePhase != nil || s.InputRoot == nil || s.ManifestSHA256 == nil || len(s.Artifacts) != 1 || s.Artifacts[0].Size != 12 {
			t.Fatalf("incomplete summary: %+v", s)
		}
		if !strings.HasPrefix(filepath.Join(config.RunRoot, *s.InputRoot), config.RunRoot+string(filepath.Separator)) {
			t.Fatal("escaped run root")
		}
	})
	t.Run("failed comparison", func(t *testing.T) {
		config, _ := fixture(t)
		want := errors.New("wrong expectation")
		path, err := Run(config, func(b *Bundle) error { return want })
		if !errors.Is(err, want) {
			t.Fatal(err)
		}
		var s Summary
		requireNoError(t, json.Unmarshal(read(t, path), &s))
		if s.Status != "failed" || *s.FailurePhase != "comparison" {
			t.Fatalf("wrong failure: %+v", s)
		}
	})
	t.Run("concurrent isolation", func(t *testing.T) {
		config, _ := fixture(t)
		var wg sync.WaitGroup
		paths := make(chan string, 2)
		failures := make(chan error, 2)
		for range 2 {
			wg.Go(func() {
				path, err := Run(config, func(b *Bundle) error { return b.Record("actual.json", []byte("{}")) })
				paths <- path
				failures <- err
			})
		}
		wg.Wait()
		requireNoError(t, <-failures)
		requireNoError(t, <-failures)
		firstPath, secondPath := <-paths, <-paths
		if firstPath == secondPath {
			t.Fatal("shared fixture workspace")
		}
	})
	t.Run("publication failure", func(t *testing.T) {
		config, _ := fixture(t)
		_, err := Run(config, func(b *Bundle) error {
			return os.Mkdir(filepath.Join(filepath.Dir(b.outputRoot), "comparison-summary.json"), 0700)
		})
		if err == nil {
			t.Fatal("publication failure swallowed")
		}
	})
	t.Run("artifact bounds and references", func(t *testing.T) {
		config, _ := fixture(t)
		_, err := Run(config, func(b *Bundle) error {
			for _, name := range []string{"../outside", "/tmp/out", "nested/file", "bad\nname", "."} {
				if b.Record(name, nil) == nil {
					return fmt.Errorf("unsafe artifact name accepted: %q", name)
				}
			}
			for i := range maxArtifacts {
				if err := b.Record(fmt.Sprintf("actual-%02d.json", i), []byte("{}")); err != nil {
					return err
				}
			}
			if b.Record("overflow.json", nil) == nil {
				return errors.New("unbounded artifact metadata")
			}
			return nil
		})
		requireNoError(t, err)
	})
	t.Run("run root symlink", func(t *testing.T) {
		config, _ := fixture(t)
		outside := t.TempDir()
		link := filepath.Join(config.RunRoot, "link")
		requireNoError(t, os.Symlink(outside, link))
		config.RunRoot = filepath.Join(link, "must-not-exist")
		_, err := Run(config, func(*Bundle) error { t.Fatal("callback after unsafe run root"); return nil })
		if err == nil {
			t.Fatal("run symlink admitted")
		}
		if _, err := os.Stat(filepath.Join(outside, "must-not-exist")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("unsafe path created before rejection")
		}
	})
}
