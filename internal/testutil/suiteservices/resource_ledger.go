package suiteservices

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const resourceLedgerName = "resource-ledger.json"

// ResourceLedger is private cleanup authority. It is never copied into retained
// public diagnostics and is removed after successful cleanup.
type ResourceLedger struct {
	SchemaID        string                 `json:"schema_id"`
	SuiteID         string                 `json:"suite_id"`
	RunID           string                 `json:"run_id"`
	Databases       []string               `json:"databases,omitempty"`
	Buckets         []string               `json:"buckets,omitempty"`
	BrowserFixtures []BrowserE2EFixtureRef `json:"browser_fixtures,omitempty"`
}

func refreshResourceLedger(env map[string]string) error {
	suiteDir, ok, err := ResolveSuiteArtifactDir(env)
	if err != nil || !ok {
		return err
	}
	events, err := readJournalEvents(suiteDir)
	if err != nil {
		return err
	}
	ledger := resourceLedgerFromEvents(env, events)
	payload, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return fmt.Errorf("encode suite-service resource ledger: %w", err)
	}
	if err := writeFileAtomically(filepath.Join(suiteDir, resourceLedgerName), append(payload, '\n'), 0o600); err != nil {
		return fmt.Errorf("write suite-service resource ledger: %w", err)
	}
	return nil
}

func resourceLedgerFromEvents(env map[string]string, events []Event) ResourceLedger {
	databases := make(map[string]struct{})
	buckets := make(map[string]struct{})
	retired := make(map[string]BrowserE2EFixtureRef)
	completed := make(map[string]struct{})
	for _, event := range events {
		switch event.Type {
		case EventPostgresDBCreated:
			if event.Name != "" {
				databases[event.Name] = struct{}{}
			}
		case EventPostgresDBDropped:
			delete(databases, event.Name)
		case EventS3BucketCreated:
			if event.Name != "" {
				buckets[event.Name] = struct{}{}
			}
		case EventS3BucketCleaned:
			delete(buckets, event.Name)
		case EventWebE2EFixtureRetired:
			upsertWebE2EFixture(retired, event)
		case EventWebE2EFixtureCleaned, EventWebE2EFixtureReclaimed:
			fixture := webE2EFixtureFromEvent(event)
			completed[fixtureKey(fixture.DatabaseName, fixture.Bucket, fixture.Target)] = struct{}{}
		}
	}
	for key := range completed {
		delete(retired, key)
	}
	return ResourceLedger{
		SchemaID:        "cartulary.test_services.resource_ledger.v1",
		SuiteID:         SuiteID(env),
		RunID:           ResolveRunID(env),
		Databases:       sortedKeys(databases),
		Buckets:         sortedKeys(buckets),
		BrowserFixtures: sortedWebE2EFixtures(retired),
	}
}

func webE2EFixtureFromEvent(event Event) BrowserE2EFixtureRef {
	return BrowserE2EFixtureRef{
		DatabaseName:    stringDetail(event.Details, "database_name"),
		Bucket:          stringDetail(event.Details, "bucket"),
		Target:          stringDetail(event.Details, "target"),
		ReclaimStrategy: stringDetail(event.Details, "reclaim_strategy"),
		Timestamp:       event.Timestamp,
		PID:             event.PID,
	}
}

func ReadResourceLedger(env map[string]string) (ResourceLedger, bool, error) {
	suiteDir, ok, err := ResolveSuiteArtifactDir(env)
	if err != nil || !ok {
		return ResourceLedger{}, ok, err
	}
	ledger, found, err := ReadResourceLedgerFile(filepath.Join(suiteDir, resourceLedgerName))
	if err != nil || !found {
		return ledger, found, err
	}
	if ledger.SuiteID != SuiteID(env) || ledger.RunID != ResolveRunID(env) {
		return ResourceLedger{}, true, fmt.Errorf("suite-service resource ledger does not match the active suite/run identity")
	}
	return ledger, true, nil
}

func CurrentResourceLedger(env map[string]string) (ResourceLedger, bool, error) {
	suiteDir, ok, err := ResolveSuiteArtifactDir(env)
	if err != nil || !ok {
		return ResourceLedger{}, ok, err
	}
	events, err := readJournalEvents(suiteDir)
	if err != nil {
		return ResourceLedger{}, true, err
	}
	return resourceLedgerFromEvents(env, events), true, nil
}

func ReadResourceLedgerFile(path string) (ResourceLedger, bool, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return ResourceLedger{}, false, nil
	}
	if err != nil {
		return ResourceLedger{}, false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 {
		return ResourceLedger{}, true, fmt.Errorf("suite-service resource ledger must be a mode-0600 non-symlink regular file")
	}
	raw, err := os.ReadFile(path) // #nosec G304 -- callers resolve the path beneath the current run root.
	if err != nil {
		return ResourceLedger{}, true, err
	}
	var ledger ResourceLedger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		return ResourceLedger{}, true, fmt.Errorf("decode suite-service resource ledger: %w", err)
	}
	if ledger.SchemaID != "cartulary.test_services.resource_ledger.v1" || ledger.SuiteID == "" || ledger.RunID == "" {
		return ResourceLedger{}, true, fmt.Errorf("suite-service resource ledger identity is invalid")
	}
	sort.Strings(ledger.Databases)
	sort.Strings(ledger.Buckets)
	return ledger, true, nil
}

func RemoveResourceLedger(env map[string]string) error {
	suiteDir, ok, err := ResolveSuiteArtifactDir(env)
	if err != nil || !ok {
		return err
	}
	err = os.Remove(filepath.Join(suiteDir, resourceLedgerName))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
