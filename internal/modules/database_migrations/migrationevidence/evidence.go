package migrationevidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
)

const (
	SchemaID            = "cartulary.migration_history_evidence.v2"
	DefaultManifestPath = "tools/migration_history_manifest.json"
)

var filenamePattern = regexp.MustCompile(`^([0-9]{5})_[a-z0-9]+(?:_[a-z0-9]+)*\.sql$`)
var phaseNamePattern = regexp.MustCompile(`(^|_)phase[0-9]+(_|$)`)
var manifestLogicalIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
var sha256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var bindingKindPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
var serviceRefPattern = regexp.MustCompile(`^[A-Za-z0-9._:@+-]+$`)

type Result struct {
	SchemaID          string          `json:"schema_id"`
	CollectedAt       time.Time       `json:"collected_at"`
	EvidenceOnly      bool            `json:"evidence_only"`
	RewriteAuthorized bool            `json:"rewrite_authorized"`
	DatabaseBinding   DatabaseBinding `json:"database_binding"`
	Manifest          ManifestSummary `json:"manifest"`
	SourceAudit       []SourceAudit   `json:"source_audit"`
	GooseLedger       GooseLedger     `json:"goose_ledger"`
	Findings          []Finding       `json:"findings"`
}

type DatabaseBinding struct {
	BindingKind string `json:"binding_kind"`
	ServiceRef  string `json:"service_ref,omitempty"`
}

type ManifestSummary struct {
	SchemaID                string `json:"schema_id"`
	SHA256                  string `json:"sha256"`
	MigrationRoot           string `json:"migration_root"`
	ImmutableThroughVersion int64  `json:"immutable_through_version"`
	ExpectedMinVersion      int64  `json:"expected_min_version"`
	ExpectedMaxVersion      int64  `json:"expected_max_version"`
	ExpectedVersionCount    int    `json:"expected_version_count"`
}

type SourceAudit struct {
	Version             int64  `json:"version"`
	Filename            string `json:"filename"`
	SHA256              string `json:"sha256"`
	HasGooseUp          bool   `json:"has_goose_up"`
	HasGooseDown        bool   `json:"has_goose_down"`
	PhaseShapedName     bool   `json:"phase_shaped_name"`
	ImmutabilityClass   string `json:"immutability_class"`
	ManifestFilename    string `json:"manifest_filename,omitempty"`
	ManifestSHA256      string `json:"manifest_sha256,omitempty"`
	ManifestHashMatches bool   `json:"manifest_hash_matches"`
}

type GooseLedger struct {
	MetadataPresent                bool         `json:"metadata_present"`
	RowCount                       int64        `json:"row_count"`
	CurrentEffectiveAppliedVersion int64        `json:"current_effective_applied_version"`
	LatestEffectiveStates          []GooseState `json:"latest_effective_states"`
}

type GooseState struct {
	Version   int64     `json:"version"`
	IsApplied bool      `json:"is_applied"`
	TStamp    time.Time `json:"tstamp"`
}

type Finding struct {
	Severity   string `json:"severity"`
	ReasonCode string `json:"reason_code"`
	Version    *int64 `json:"version,omitempty"`
	Filename   string `json:"filename,omitempty"`
	Detail     string `json:"detail"`
}

type manifestDocument struct {
	SchemaID                string          `json:"schema_id"`
	MigrationRoot           string          `json:"migration_root"`
	ImmutableThroughVersion int64           `json:"immutable_through_version"`
	Entries                 []manifestEntry `json:"entries"`
}

type manifestEntry struct {
	Version               int64  `json:"version"`
	Filename              string `json:"filename"`
	SHA256                string `json:"sha256"`
	HistoricalPhaseShaped bool   `json:"historical_phase_shaped"`
}

type manifestFailureReason string

const (
	manifestFailurePathRequired manifestFailureReason = "migration evidence manifest path is required"
	manifestFailureUnavailable  manifestFailureReason = "migration evidence manifest unavailable"
	manifestFailureInvalid      manifestFailureReason = "migration evidence manifest invalid"
)

type manifestFailureError struct {
	reason manifestFailureReason
}

func (failure manifestFailureError) Error() string {
	return string(failure.reason)
}

func newManifestFailure(reason manifestFailureReason) error {
	return manifestFailureError{reason: reason}
}

func Build(ctx context.Context, binding DatabaseBinding, pool database_migrations.LedgerReader, collectedAt time.Time, manifestPath string, source *database_migrations.Source) (Result, error) {
	normalizedBinding := normalizeDatabaseBinding(binding)
	if !safeDatabaseBinding(normalizedBinding) {
		return Result{}, errors.New("migration evidence database binding is invalid")
	}
	manifest, manifestSummary, manifestFindings, err := loadManifest(manifestPath)
	if err != nil {
		return Result{}, err
	}
	manifestByVersion := map[int64]manifestEntry{}
	for _, entry := range manifest.Entries {
		manifestByVersion[entry.Version] = entry
	}

	inspection, err := database_migrations.InspectSource(source)
	if err != nil {
		return Result{}, fmt.Errorf("inspect embedded migration source: %w", err)
	}
	sourceAudit, sourceFindings := auditSource(inspection, manifest, manifestByVersion)
	ledger, ledgerFindings, err := collectGooseLedger(ctx, pool, manifestByVersion, manifest.ImmutableThroughVersion)
	if err != nil {
		return Result{}, err
	}

	findings := append([]Finding{}, manifestFindings...)
	findings = append(findings, sourceFindings...)
	findings = append(findings, ledgerFindings...)
	sortFindings(findings)

	return Result{
		SchemaID:          SchemaID,
		CollectedAt:       collectedAt,
		EvidenceOnly:      true,
		RewriteAuthorized: false,
		DatabaseBinding:   normalizedBinding,
		Manifest:          manifestSummary,
		SourceAudit:       sourceAudit,
		GooseLedger:       ledger,
		Findings:          findings,
	}, nil
}

func normalizeDatabaseBinding(binding DatabaseBinding) DatabaseBinding {
	return DatabaseBinding{
		BindingKind: strings.TrimSpace(binding.BindingKind),
		ServiceRef:  strings.TrimSpace(binding.ServiceRef),
	}
}

func safeDatabaseBinding(binding DatabaseBinding) bool {
	if !bindingKindPattern.MatchString(binding.BindingKind) {
		return false
	}
	return binding.ServiceRef == "" || serviceRefPattern.MatchString(binding.ServiceRef)
}

func loadManifest(path string) (manifestDocument, ManifestSummary, []Finding, error) {
	if strings.TrimSpace(path) == "" {
		return manifestDocument{}, ManifestSummary{}, nil, newManifestFailure(manifestFailurePathRequired)
	}
	body, err := os.ReadFile(filepath.Clean(path)) // #nosec G304 -- operator supplied local manifest path is intentionally read by deployment-local CLI.
	if err != nil {
		return manifestDocument{}, ManifestSummary{}, nil, newManifestFailure(manifestFailureUnavailable)
	}
	var manifest manifestDocument
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifestDocument{}, ManifestSummary{}, nil, newManifestFailure(manifestFailureInvalid)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return manifestDocument{}, ManifestSummary{}, nil, newManifestFailure(manifestFailureInvalid)
	}
	if !safeManifestDocument(manifest) {
		return manifestDocument{}, ManifestSummary{}, nil, newManifestFailure(manifestFailureInvalid)
	}

	findings := auditManifest(manifest)
	minVersion, maxVersion := versionRange(manifest.Entries)
	summary := ManifestSummary{
		SchemaID:                manifest.SchemaID,
		SHA256:                  sha256Hex(body),
		MigrationRoot:           manifest.MigrationRoot,
		ImmutableThroughVersion: manifest.ImmutableThroughVersion,
		ExpectedMinVersion:      minVersion,
		ExpectedMaxVersion:      maxVersion,
		ExpectedVersionCount:    len(manifest.Entries),
	}
	return manifest, summary, findings, nil
}

func safeManifestDocument(manifest manifestDocument) bool {
	if !manifestLogicalIDPattern.MatchString(manifest.SchemaID) || manifest.MigrationRoot != "db/migrations" {
		return false
	}
	for _, entry := range manifest.Entries {
		if !filenamePattern.MatchString(entry.Filename) || !sha256Pattern.MatchString(entry.SHA256) {
			return false
		}
	}
	return true
}

func auditManifest(manifest manifestDocument) []Finding {
	findings := []Finding{}
	if manifest.SchemaID != "cartulary.migration_history_manifest.v1" {
		findings = append(findings, finding("blocking", "manifest_schema_unsupported", 0, "", fmt.Sprintf("manifest schema_id %q is not supported", manifest.SchemaID)))
	}
	seen := map[int64]string{}
	versions := make([]int64, 0, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		if prior, ok := seen[entry.Version]; ok {
			findings = append(findings, finding("blocking", "manifest_duplicate_version", entry.Version, entry.Filename, fmt.Sprintf("version also appears as %s", prior)))
			continue
		}
		seen[entry.Version] = entry.Filename
		versions = append(versions, entry.Version)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	for index, version := range versions {
		expected := int64(index + 1)
		if version != expected {
			findings = append(findings, finding("blocking", "manifest_version_gap", version, seen[version], fmt.Sprintf("expected contiguous version %05d", expected)))
			break
		}
	}
	return findings
}

func auditSource(inspection database_migrations.SourceInspection, manifest manifestDocument, manifestByVersion map[int64]manifestEntry) ([]SourceAudit, []Finding) {
	sourceVersions := make(map[int64]struct{}, len(inspection.Entries))
	audits := []SourceAudit{}
	findings := []Finding{}
	for _, sourceEntry := range inspection.Entries {
		version := sourceEntry.Version
		entry, inManifest := manifestByVersion[version]
		audit := SourceAudit{
			Version:             version,
			Filename:            sourceEntry.Filename,
			SHA256:              sourceEntry.SHA256,
			HasGooseUp:          sourceEntry.HasGooseUp,
			HasGooseDown:        sourceEntry.HasGooseDown,
			PhaseShapedName:     phaseNamePattern.MatchString(strings.TrimSuffix(sourceEntry.Filename, ".sql")),
			ImmutabilityClass:   immutabilityClass(version, manifest.ImmutableThroughVersion),
			ManifestHashMatches: false,
		}
		if inManifest {
			audit.ManifestFilename = entry.Filename
			audit.ManifestSHA256 = entry.SHA256
			audit.ManifestHashMatches = audit.SHA256 == entry.SHA256
		}
		sourceVersions[version] = struct{}{}
		audits = append(audits, audit)

		switch {
		case !inManifest:
			findings = append(findings, finding("blocking", "source_version_not_in_manifest", version, sourceEntry.Filename, "embedded migration version is absent from the manifest"))
		case entry.Filename != sourceEntry.Filename:
			findings = append(findings, finding("blocking", "manifest_filename_mismatch", version, sourceEntry.Filename, fmt.Sprintf("manifest filename is %s", entry.Filename)))
		case entry.SHA256 != audit.SHA256:
			findings = append(findings, finding("blocking", "manifest_hash_mismatch", version, sourceEntry.Filename, "embedded migration hash differs from manifest"))
		}
		if version > manifest.ImmutableThroughVersion && audit.PhaseShapedName {
			findings = append(findings, finding("warning", "future_phase_shaped_filename", version, sourceEntry.Filename, "migration after immutable history uses a phase-shaped filename"))
		}
	}

	for _, entry := range manifest.Entries {
		if _, ok := sourceVersions[entry.Version]; !ok {
			findings = append(findings, finding("blocking", "manifest_version_not_in_source", entry.Version, entry.Filename, "manifest entry has no embedded migration source"))
		}
	}
	return audits, findings
}

func collectGooseLedger(ctx context.Context, pool database_migrations.LedgerReader, manifestByVersion map[int64]manifestEntry, immutableThroughVersion int64) (GooseLedger, []Finding, error) {
	var metadataPresent bool
	if err := pool.QueryRow(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&metadataPresent); err != nil {
		return GooseLedger{}, nil, fmt.Errorf("inspect goose metadata table: %w", err)
	}
	if !metadataPresent {
		return GooseLedger{MetadataPresent: false}, []Finding{
			finding("blocking", "migration_metadata_missing", 0, "", "goose_db_version is missing; no applied-version evidence is available"),
		}, nil
	}

	var rowCount int64
	if err := pool.QueryRow(ctx, `SELECT COUNT(*)::bigint FROM goose_db_version`).Scan(&rowCount); err != nil {
		return GooseLedger{}, nil, fmt.Errorf("count goose metadata rows: %w", err)
	}

	rows, err := pool.Query(ctx, `
SELECT version_id::bigint, is_applied, tstamp
  FROM (
        SELECT version_id, is_applied, tstamp,
               row_number() OVER (PARTITION BY version_id ORDER BY id DESC) AS rn
          FROM goose_db_version
       ) latest
 WHERE rn = 1
 ORDER BY version_id ASC
`)
	if err != nil {
		return GooseLedger{}, nil, fmt.Errorf("inspect goose effective states: %w", err)
	}
	defer rows.Close()

	states := []GooseState{}
	stateByVersion := map[int64]GooseState{}
	currentApplied := int64(0)
	for rows.Next() {
		var state GooseState
		if err := rows.Scan(&state.Version, &state.IsApplied, &state.TStamp); err != nil {
			return GooseLedger{}, nil, fmt.Errorf("scan goose effective state: %w", err)
		}
		states = append(states, state)
		stateByVersion[state.Version] = state
		if state.IsApplied && state.Version > currentApplied {
			currentApplied = state.Version
		}
	}
	if err := rows.Err(); err != nil {
		return GooseLedger{}, nil, fmt.Errorf("scan goose effective states: %w", err)
	}

	findings := []Finding{}
	for _, state := range states {
		if state.Version == 0 {
			continue
		}
		if _, ok := manifestByVersion[state.Version]; !ok {
			findings = append(findings, finding("blocking", "db_version_not_in_manifest", state.Version, "", "database ledger references a version absent from the manifest"))
		}
	}
	for version := int64(1); version <= currentApplied; version++ {
		if state, ok := stateByVersion[version]; !ok || !state.IsApplied {
			findings = append(findings, finding("blocking", "db_applied_version_gap", version, "", "effective applied ledger has a gap at or below current schema version"))
			break
		}
	}
	if currentApplied >= immutableThroughVersion && immutableThroughVersion > 0 {
		findings = append(findings, finding("info", "protected_boundary_applied", immutableThroughVersion, "", "database has applied the immutable historical migration boundary; rewrite/rebaseline remains blocked without owner approval"))
	}

	return GooseLedger{
		MetadataPresent:                true,
		RowCount:                       rowCount,
		CurrentEffectiveAppliedVersion: currentApplied,
		LatestEffectiveStates:          states,
	}, findings, nil
}

func versionRange(entries []manifestEntry) (int64, int64) {
	if len(entries) == 0 {
		return 0, 0
	}
	minVersion := entries[0].Version
	maxVersion := entries[0].Version
	for _, entry := range entries[1:] {
		if entry.Version < minVersion {
			minVersion = entry.Version
		}
		if entry.Version > maxVersion {
			maxVersion = entry.Version
		}
	}
	return minVersion, maxVersion
}

func immutabilityClass(version int64, immutableThroughVersion int64) string {
	if immutableThroughVersion > 0 && version <= immutableThroughVersion {
		return "protected"
	}
	return "current"
}

func finding(severity string, reasonCode string, version int64, filename string, detail string) Finding {
	var versionPointer *int64
	if version > 0 {
		versionPointer = &version
	}
	return Finding{
		Severity:   severity,
		ReasonCode: reasonCode,
		Version:    versionPointer,
		Filename:   filename,
		Detail:     detail,
	}
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		leftVersion := int64(0)
		rightVersion := int64(0)
		if findings[i].Version != nil {
			leftVersion = *findings[i].Version
		}
		if findings[j].Version != nil {
			rightVersion = *findings[j].Version
		}
		if leftVersion != rightVersion {
			return leftVersion < rightVersion
		}
		if findings[i].ReasonCode != findings[j].ReasonCode {
			return findings[i].ReasonCode < findings[j].ReasonCode
		}
		return findings[i].Filename < findings[j].Filename
	})
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
