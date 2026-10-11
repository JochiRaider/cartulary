//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"golang.org/x/sys/unix"
)

const cacheSchema = "cartulary.harness_cache_record.v2"

type cachedArtifact struct {
	artifact
	Digest  string           `json:"digest"`
	Input   string           `json:"semantic_input_digest"`
	Members []artifactMember `json:"members"`
}
type cachedRow struct {
	ID     string `json:"row_id"`
	Runner string `json:"runner"`
}
type semanticRows struct {
	Status string      `json:"status"`
	Rows   []cachedRow `json:"rows"`
}
type cacheRecord struct {
	Schema        string           `json:"schema_id"`
	Profile       string           `json:"profile_id"`
	Policy        string           `json:"policy"`
	Producer      string           `json:"producer_identity"`
	UnitID        string           `json:"unit_id"`
	UnitDigest    string           `json:"unit_digest"`
	Input         string           `json:"input_digest"`
	Semantic      semanticRows     `json:"semantic_result"`
	Output        string           `json:"output_digest"`
	Artifacts     []cachedArtifact `json:"artifacts"`
	Vulnerability string           `json:"vulnerability_database_revision,omitempty"`
	Created       string           `json:"created_at"`
}
type cacheEntry struct {
	record    cacheRecord
	items     []publishItem
	directory string
}
type ownedCacheStage struct {
	name     string
	identity unix.Stat_t
}

type nativeCache struct {
	mu              sync.Mutex
	plan            *Plan
	root, run, mode string
	facts           map[string]cacheFact
	sameRun         map[string]string
	staging         []ownedCacheStage
	Now             func() time.Time
	BeforeRename    func(int) error
	Fresh           func(context.Context, cacheFact) error
}

func newNativeCache(plan *Plan, root, run string) *nativeCache {
	c := &nativeCache{plan: plan, root: root, run: run, mode: plan.facts.Manifest["cache_mode"].(string), facts: map[string]cacheFact{}, sameRun: map[string]string{}, Now: time.Now}
	for _, fact := range plan.facts.Cache {
		c.facts[fact.UnitID] = fact
	}
	c.Fresh = c.freshInputs
	return c
}
func textValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func (c *nativeCache) base() string { return filepath.Join(c.root, ".cache/cartulary/graph-v2") }
func (c *nativeCache) directory(f cacheFact) string {
	return filepath.Join(c.base(), textValue(f.Profile), strings.TrimPrefix(textValue(f.InputDigest), "sha256:"))
}
func (c *nativeCache) freshInputs(ctx context.Context, f cacheFact) error {
	for _, index := range c.plan.facts.CacheInputs[textValue(f.DependencyDigest)] {
		entry := c.plan.facts.SourceInputs[index]
		if !safeRelative(entry.Path) {
			return errPlan
		}
		name := filepath.Join(c.root, entry.Path)
		if entry.Kind == "symlink" {
			parent, err := openOutputParent(name, false)
			if err != nil {
				return err
			}
			var before, after unix.Stat_t
			err = unix.Fstatat(int(parent.file.Fd()), parent.leaf, &before, unix.AT_SYMLINK_NOFOLLOW)
			if err != nil || before.Mode&unix.S_IFMT != unix.S_IFLNK {
				parent.close()
				return harnesslocal.ErrUnsafe
			}
			buffer := make([]byte, 4097)
			n, err := unix.Readlinkat(int(parent.file.Fd()), parent.leaf, buffer)
			if err != nil || n >= len(buffer) {
				parent.close()
				return harnesslocal.ErrUnsafe
			}
			err = unix.Fstatat(int(parent.file.Fd()), parent.leaf, &after, unix.AT_SYMLINK_NOFOLLOW)
			same := err == nil && stableStat(before, after) && parent.check() == nil
			parent.close()
			if !same || bytesDigest(buffer[:n]) != entry.Digest || fmt.Sprintf("%04o", before.Mode&07777) != entry.Mode {
				return harnesslocal.ErrUnsafe
			}
			continue
		}
		digest, mode, err := streamRegular(ctx, name, nil)
		if err != nil {
			return err
		}
		if digest != entry.Digest || modeString(mode) != entry.Mode {
			return harnesslocal.ErrUnsafe
		}
	}
	return nil
}
func (c *nativeCache) Lookup(ctx context.Context, u unit) (cacheResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	f, ok := c.facts[u.ID]
	if !ok {
		return cacheResult{}, errPlan
	}
	result := cacheResult{Profile: textValue(f.Profile)}
	if u.CachePolicy == "none" {
		result.Outcome = "bypass"
		result.Reason = "policy_none"
		return result, nil
	}
	if !f.Eligible || c.mode == "off" {
		result.Outcome = "bypass"
		result.Reason = textValue(f.Reason)
		if c.mode == "off" {
			result.Reason = "mode_off"
		}
		return result, nil
	}
	if err := c.Fresh(ctx, f); err != nil {
		return cacheResult{}, failed("config", "configuration_error")
	}
	result.Outcome = "miss"
	result.Reason = "record_missing"
	if c.mode == "cold" {
		result.Reason = "cold_read_bypass"
		return result, nil
	}
	directory := c.directory(f)
	if u.CachePolicy == "same_run" {
		directory = c.sameRun[textValue(f.InputDigest)]
		if directory == "" {
			return result, nil
		}
	}
	if _, err := os.Lstat(directory); errors.Is(err, os.ErrNotExist) {
		return result, nil
	} else if err != nil {
		return cacheResult{}, err
	}
	entry, err := c.readEntry(ctx, directory, u, f)
	if err != nil {
		result.Reason = "record_invalid"
		if u.CachePolicy != "same_run" {
			if err := c.quarantine(directory, f); err != nil {
				return cacheResult{}, err
			}
		}
		return result, nil
	}
	items, err := semanticEvidence(u, entry.record.Semantic, c.Now())
	if err != nil {
		return cacheResult{}, err
	}
	if err := publishOutputs(ctx, c.root, c.run, append(entry.items, items...), c.BeforeRename); err != nil {
		if errors.Is(err, errIncompletePublication) {
			return cacheResult{}, failed("artifact", "artifact_error")
		}
		result.Reason = "restore_rejected"
		return result, nil
	}
	result.Outcome = "hit"
	result.Reason = "record_valid"
	result.OutputDigest = entry.record.Output
	return result, nil
}
func (c *nativeCache) readEntry(ctx context.Context, directory string, u unit, f cacheFact) (cacheEntry, error) {
	root, err := harnesslocal.OpenDirectory(directory, true)
	if err != nil {
		return cacheEntry{}, err
	}
	defer root.Close()
	bytes, err := harnesslocal.ReadFileAt(root, "record.json", 32*1024*1024, false)
	if err != nil {
		return cacheEntry{}, err
	}
	value, err := harnessschema.Decode(bytes)
	if err != nil || harnessschema.Validate(cacheSchema, value) != nil {
		return cacheEntry{}, harnesslocal.ErrUnsafe
	}
	var record cacheRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return cacheEntry{}, err
	}
	if record.Schema != cacheSchema || record.Profile != textValue(f.Profile) || record.Policy != u.CachePolicy || record.Producer != u.ID || record.UnitID != u.ID || record.UnitDigest != u.Digest || record.Input != textValue(f.InputDigest) || record.Vulnerability != textValue(f.Vulnerability) || len(record.Artifacts) != len(u.Artifacts) {
		return cacheEntry{}, harnesslocal.ErrUnsafe
	}
	if err := validateSemanticRows(u, record.Semantic); err != nil {
		return cacheEntry{}, err
	}
	artifactsRoot := filepath.Join(directory, "artifacts")
	members, err := walkArtifact(ctx, artifactsRoot)
	if err != nil {
		return cacheEntry{}, err
	}
	expected := map[string]string{}
	items := []publishItem{}
	for i, manifest := range record.Artifacts {
		if manifest.artifact != u.Artifacts[i] || manifest.Input != record.Input {
			return cacheEntry{}, harnesslocal.ErrUnsafe
		}
		prefix := strconv.Itoa(i)
		expected[prefix] = "directory"
		item := publishItem{Path: manifest.Path, Destination: manifest.Destination, Mode: manifest.Mode, Type: manifest.Type, Digest: manifest.Digest, Members: manifest.Members, Payloads: map[string]string{}}
		if manifest.Type == "file" {
			if len(manifest.Members) != 0 {
				return cacheEntry{}, harnesslocal.ErrUnsafe
			}
			item.Source = filepath.Join(artifactsRoot, prefix, "0")
			expected[prefix+"/0"] = "file"
			digest, mode, err := streamRegular(ctx, item.Source, nil)
			if err != nil || mode != 0600 || digest != manifest.Digest {
				return cacheEntry{}, harnesslocal.ErrUnsafe
			}
		} else {
			byPath := map[string]artifactMember{}
			previous := ""
			for j, member := range manifest.Members {
				if !safeRelative(member.Path) || member.Path <= previous {
					return cacheEntry{}, harnesslocal.ErrUnsafe
				}
				previous = member.Path
				parent := filepath.ToSlash(filepath.Dir(member.Path))
				if parent != "." && byPath[parent].Type != "directory" {
					return cacheEntry{}, harnesslocal.ErrUnsafe
				}
				byPath[member.Path] = member
				if member.Type == "directory" {
					digest, err := directoryDigest(member)
					if err != nil || digest != member.Digest {
						return cacheEntry{}, harnesslocal.ErrUnsafe
					}
				} else {
					relative := prefix + "/" + strconv.Itoa(j)
					expected[relative] = "file"
					source := filepath.Join(artifactsRoot, relative)
					digest, mode, err := streamRegular(ctx, source, nil)
					if err != nil || mode != 0600 || digest != member.Digest {
						return cacheEntry{}, harnesslocal.ErrUnsafe
					}
					item.Payloads[member.Path] = source
				}
			}
			digest, err := semanticDigest(manifest.Members)
			if err != nil || digest != manifest.Digest {
				return cacheEntry{}, harnesslocal.ErrUnsafe
			}
		}
		items = append(items, item)
	}
	if len(expected) != len(members) {
		return cacheEntry{}, harnesslocal.ErrUnsafe
	}
	for _, member := range members {
		if expected[member.Path] != member.Type || member.Type == "directory" && member.Mode != "0700" || member.Type == "file" && member.Mode != "0600" {
			return cacheEntry{}, harnesslocal.ErrUnsafe
		}
	}
	digest, err := semanticDigest(map[string]any{"semantic_result": record.Semantic, "artifacts": record.Artifacts})
	if err != nil || digest != record.Output {
		return cacheEntry{}, harnesslocal.ErrUnsafe
	}
	if harnesslocal.CheckDirectory(root, directory) != nil {
		return cacheEntry{}, harnesslocal.ErrUnsafe
	}
	return cacheEntry{record: record, items: items, directory: directory}, nil
}
func validateSemanticRows(u unit, result semanticRows) error {
	ids := []string{}
	for _, row := range result.Rows {
		ids = append(ids, row.ID)
	}
	slices.Sort(ids)
	if result.Status != "passed" || !slices.Equal(ids, rowIDs(u)) {
		return harnesslocal.ErrUnsafe
	}
	return nil
}
func semanticEvidence(u unit, result semanticRows, now time.Time) ([]publishItem, error) {
	if err := validateSemanticRows(u, result); err != nil {
		return nil, err
	}
	items := []publishItem{}
	stamp := now.UTC().Format("2006-01-02T15:04:05.000Z")
	for _, row := range result.Rows {
		value := map[string]any{"schema_id": "cartulary.harness_row_result.v3", "row_id": row.ID, "runner": row.Runner, "terminal_state": "passed", "duration_ms": 0, "wall_duration_ms": 0, "exit_code": 0, "failure_class": nil, "failure_reason": nil, "failure_diagnostic": nil, "started_at": stamp, "finished_at": stamp}
		bytes, err := documentBytes("cartulary.harness_row_result.v3", value)
		if err != nil {
			return nil, err
		}
		items = append(items, publishItem{Type: "file", Path: "rows/" + row.ID + ".json", Destination: "run_root", Mode: "0600", Bytes: append(bytes, '\n')})
	}
	for _, name := range u.Evidence {
		if strings.HasPrefix(name, "unit-results/") {
			value := map[string]any{"schema_id": "cartulary.harness_unit_result.v1", "unit_id": u.ID, "semantic_digest": u.Digest, "status": "passed", "exit_code": nil, "signal": nil, "failure_class": nil, "failure_reason": nil, "evidence_outputs": u.Evidence, "missing_outputs": []string{}}
			bytes, err := documentBytes("cartulary.harness_unit_result.v1", value)
			if err != nil {
				return nil, err
			}
			items = append(items, publishItem{Type: "file", Path: name, Destination: "run_root", Mode: "0600", Bytes: append(bytes, '\n')})
			return items, nil
		}
	}
	return nil, harnesslocal.ErrUnsafe
}
func (c *nativeCache) snapshot(ctx context.Context, u unit, f cacheFact) (cacheRecord, []publishItem, error) {
	record := cacheRecord{Schema: cacheSchema, Profile: textValue(f.Profile), Policy: u.CachePolicy, Producer: u.ID, UnitID: u.ID, UnitDigest: u.Digest, Input: textValue(f.InputDigest), Semantic: semanticRows{Status: "passed", Rows: []cachedRow{}}, Artifacts: []cachedArtifact{}, Vulnerability: textValue(f.Vulnerability), Created: c.Now().UTC().Format("2006-01-02T15:04:05.000Z")}
	items := []publishItem{}
	for _, row := range rowIDs(u) {
		bytes, err := harnesslocal.ReadFile(filepath.Join(c.run, "rows", row+".json"), 16*1024*1024, false)
		if err != nil {
			return cacheRecord{}, nil, err
		}
		value, err := harnessschema.Decode(bytes)
		if err != nil || harnessschema.Validate("cartulary.harness_row_result.v3", value) != nil {
			return cacheRecord{}, nil, harnesslocal.ErrUnsafe
		}
		object := value.(map[string]any)
		if object["row_id"] != row || object["terminal_state"] != "passed" {
			return cacheRecord{}, nil, harnesslocal.ErrUnsafe
		}
		record.Semantic.Rows = append(record.Semantic.Rows, cachedRow{ID: row, Runner: object["runner"].(string)})
	}
	for i, declared := range u.Artifacts {
		base := c.root
		if declared.Destination == "run_root" {
			base = c.run
		}
		name := filepath.Join(base, declared.Path)
		info, err := os.Lstat(name)
		if err != nil || modeString(info.Mode().Perm()) != declared.Mode {
			return cacheRecord{}, nil, harnesslocal.ErrUnsafe
		}
		manifest := cachedArtifact{artifact: declared, Input: record.Input, Members: []artifactMember{}}
		if declared.Type == "file" {
			digest, _, err := streamRegular(ctx, name, nil)
			if err != nil {
				return cacheRecord{}, nil, err
			}
			manifest.Digest = digest
			items = append(items, publishItem{Type: "file", Path: fmt.Sprintf("artifacts/%d/0", i), Destination: "run_root", Mode: "0600", Source: name, Digest: digest})
		} else {
			if !info.IsDir() {
				return cacheRecord{}, nil, harnesslocal.ErrUnsafe
			}
			members, err := walkArtifact(ctx, name)
			if err != nil {
				return cacheRecord{}, nil, err
			}
			manifest.Members = members
			manifest.Digest, err = semanticDigest(members)
			if err != nil {
				return cacheRecord{}, nil, err
			}
			for j, member := range members {
				if member.Type == "file" {
					items = append(items, publishItem{Type: "file", Path: fmt.Sprintf("artifacts/%d/%d", i, j), Destination: "run_root", Mode: "0600", Source: filepath.Join(name, member.Path), Digest: member.Digest})
				}
			}
		}
		record.Artifacts = append(record.Artifacts, manifest)
	}
	var err error
	record.Output, err = semanticDigest(map[string]any{"semantic_result": record.Semantic, "artifacts": record.Artifacts})
	return record, items, err
}
func (c *nativeCache) Store(ctx context.Context, u unit) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	f, ok := c.facts[u.ID]
	if !ok {
		return errPlan
	}
	if c.mode == "off" || !f.Eligible {
		return nil
	}
	if c.Fresh(ctx, f) != nil {
		return nil
	}
	record, items, err := c.snapshot(ctx, u, f)
	if err != nil {
		return nil
	}
	bytes, err := documentBytes(cacheSchema, record)
	if err != nil {
		return err
	}
	base := filepath.Join(c.base(), textValue(f.Profile))
	parent, err := harnesslocal.PrivateDirectory(base)
	if err != nil {
		return err
	}
	defer parent.Close()
	name := ".staging-" + uuid()
	if err := unix.Mkdirat(int(parent.Fd()), name, 0700); err != nil {
		return err
	}
	staging := filepath.Join(base, name)
	var identity unix.Stat_t
	if err := unix.Fstatat(int(parent.Fd()), name, &identity, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err
	}
	owned := true
	defer func() {
		if owned {
			_ = removeOwnedOutput(parent, name, identity)
		}
	}()
	artifacts, err := harnesslocal.PrivateDirectory(filepath.Join(staging, "artifacts"))
	if err != nil {
		return err
	}
	artifacts.Close()
	for i := range record.Artifacts {
		directory, err := harnesslocal.PrivateDirectory(filepath.Join(staging, "artifacts", strconv.Itoa(i)))
		if err != nil {
			return err
		}
		directory.Close()
	}
	if err := publishOutputs(ctx, c.root, staging, items, nil); err != nil {
		return err
	}
	if err := harnesslocal.PublishNew(filepath.Join(staging, "record.json"), append(bytes, '\n')); err != nil {
		return err
	}
	if _, err := c.readEntry(ctx, staging, u, f); err != nil {
		return err
	}
	if c.Fresh(ctx, f) != nil {
		return nil
	}
	if u.CachePolicy == "same_run" {
		c.sameRun[record.Input] = staging
		c.staging = append(c.staging, ownedCacheStage{staging, identity})
		owned = false
		return nil
	}
	destination := filepath.Base(c.directory(f))
	for attempt := 0; attempt < 2; attempt++ {
		if err := harnesslocal.CheckDirectory(parent, base); err != nil {
			return err
		}
		err = unix.Renameat2(int(parent.Fd()), name, int(parent.Fd()), destination, unix.RENAME_NOREPLACE)
		if err == nil {
			owned = false
			return parent.Sync()
		}
		if !errors.Is(err, unix.EEXIST) && !errors.Is(err, unix.ENOTEMPTY) {
			return err
		}
		if _, err := c.readEntry(ctx, c.directory(f), u, f); err == nil {
			return nil
		}
		if err := c.quarantine(c.directory(f), f); err != nil {
			return err
		}
	}
	return harnesslocal.ErrUnsafe
}
func (c *nativeCache) quarantine(directory string, f cacheFact) error {
	parent, err := openOutputParent(directory, false)
	if err != nil {
		return err
	}
	defer parent.close()
	if parent.check() != nil {
		return harnesslocal.ErrUnsafe
	}
	// Quarantine in the same proven parent; the closed prefix cannot be selected
	// as a content-addressed key and preserves corrupt bytes for diagnosis.
	err = unix.Renameat2(int(parent.file.Fd()), parent.leaf, int(parent.file.Fd()), ".quarantine-"+parent.leaf+"-"+uuid(), unix.RENAME_NOREPLACE)
	if errors.Is(err, unix.ENOENT) {
		return nil
	}
	return err
}
func (c *nativeCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var errs []error
	for _, stage := range c.staging {
		parent, err := openOutputParent(stage.name, false)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if parent.check() == nil {
			errs = append(errs, removeOwnedOutput(parent.file, parent.leaf, stage.identity))
		} else {
			errs = append(errs, harnesslocal.ErrUnsafe)
		}
		parent.close()
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}
	c.staging = nil
	c.sameRun = map[string]string{}
	return nil
}
