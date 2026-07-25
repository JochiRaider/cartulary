package migrationevidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	SchemaID            = "cartulary.migration_history_evidence.v1"
	DefaultManifestPath = "tools/migration_history_manifest.json"
)

var filenamePattern = regexp.MustCompile(`^([0-9]{5})_[a-z0-9]+(?:_[a-z0-9]+)*\.sql$`)
var phaseNamePattern = regexp.MustCompile(`(^|_)phase[0-9]+(_|$)`)

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
	Path                    string `json:"path"`
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

func Build(ctx context.Context, binding DatabaseBinding, pool postgres.DB, collectedAt time.Time, manifestPath string, sourceFS fs.FS) (Result, error) {
	manifest, manifestSummary, manifestFindings, err := loadManifest(manifestPath)
	if err != nil {
		return Result{}, err
	}
	manifestByVersion := map[int64]manifestEntry{}
	for _, entry := range manifest.Entries {
		manifestByVersion[entry.Version] = entry
	}

	sourceAudit, sourceFindings, err := auditSource(sourceFS, manifest, manifestByVersion)
	if err != nil {
		return Result{}, err
	}
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
		DatabaseBinding: DatabaseBinding{
			BindingKind: strings.TrimSpace(binding.BindingKind),
			ServiceRef:  strings.TrimSpace(binding.ServiceRef),
		},
		Manifest:    manifestSummary,
		SourceAudit: sourceAudit,
		GooseLedger: ledger,
		Findings:    findings,
	}, nil
}

func loadManifest(path string) (manifestDocument, ManifestSummary, []Finding, error) {
	if strings.TrimSpace(path) == "" {
		return manifestDocument{}, ManifestSummary{}, nil, errors.New("migration evidence manifest path is required")
	}
	body, err := os.ReadFile(filepath.Clean(path)) // #nosec G304 -- operator supplied local manifest path is intentionally read by deployment-local CLI.
	if err != nil {
		return manifestDocument{}, ManifestSummary{}, nil, fmt.Errorf("load migration evidence manifest: %w", err)
	}
	var manifest manifestDocument
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifestDocument{}, ManifestSummary{}, nil, fmt.Errorf("decode migration evidence manifest: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return manifestDocument{}, ManifestSummary{}, nil, errors.New("decode migration evidence manifest: trailing JSON content")
	}

	findings := auditManifest(manifest)
	minVersion, maxVersion := versionRange(manifest.Entries)
	summary := ManifestSummary{
		SchemaID:                manifest.SchemaID,
		Path:                    filepath.Clean(path),
		SHA256:                  sha256Hex(body),
		MigrationRoot:           manifest.MigrationRoot,
		ImmutableThroughVersion: manifest.ImmutableThroughVersion,
		ExpectedMinVersion:      minVersion,
		ExpectedMaxVersion:      maxVersion,
		ExpectedVersionCount:    len(manifest.Entries),
	}
	return manifest, summary, findings, nil
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

func auditSource(sourceFS fs.FS, manifest manifestDocument, manifestByVersion map[int64]manifestEntry) ([]SourceAudit, []Finding, error) {
	files, err := fs.ReadDir(sourceFS, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("inspect embedded migration source: %w", err)
	}

	sourceByVersion := map[int64]SourceAudit{}
	audits := []SourceAudit{}
	findings := []Finding{}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		version, ok := parseFilenameVersion(file.Name())
		if !ok {
			findings = append(findings, finding("blocking", "source_filename_invalid", 0, file.Name(), "migration filename does not match exact 5-digit lower-snake SQL pattern"))
			continue
		}
		body, err := fs.ReadFile(sourceFS, file.Name())
		if err != nil {
			return nil, nil, fmt.Errorf("read embedded migration %s: %w", file.Name(), err)
		}
		bodyText := string(body)
		entry, inManifest := manifestByVersion[version]
		audit := SourceAudit{
			Version:             version,
			Filename:            file.Name(),
			SHA256:              sha256Hex(body),
			HasGooseUp:          strings.Contains(bodyText, "-- +goose Up"),
			HasGooseDown:        strings.Contains(bodyText, "-- +goose Down"),
			PhaseShapedName:     phaseNamePattern.MatchString(strings.TrimSuffix(file.Name(), ".sql")),
			ImmutabilityClass:   immutabilityClass(version, manifest.ImmutableThroughVersion),
			ManifestHashMatches: false,
		}
		if inManifest {
			audit.ManifestFilename = entry.Filename
			audit.ManifestSHA256 = entry.SHA256
			audit.ManifestHashMatches = audit.SHA256 == entry.SHA256
		}
		if prior, duplicate := sourceByVersion[version]; duplicate {
			findings = append(findings, finding("blocking", "source_duplicate_version", version, file.Name(), fmt.Sprintf("version also appears as %s", prior.Filename)))
		} else {
			sourceByVersion[version] = audit
		}
		audits = append(audits, audit)

		switch {
		case !inManifest:
			findings = append(findings, finding("blocking", "source_version_not_in_manifest", version, file.Name(), "embedded migration version is absent from the manifest"))
		case entry.Filename != file.Name():
			findings = append(findings, finding("blocking", "manifest_filename_mismatch", version, file.Name(), fmt.Sprintf("manifest filename is %s", entry.Filename)))
		case entry.SHA256 != audit.SHA256:
			findings = append(findings, finding("blocking", "manifest_hash_mismatch", version, file.Name(), "embedded migration hash differs from manifest"))
		}
		if !audit.HasGooseUp || !audit.HasGooseDown {
			findings = append(findings, finding("blocking", "source_marker_missing", version, file.Name(), "migration must include both goose Up and Down markers"))
		}
		if version > manifest.ImmutableThroughVersion && audit.PhaseShapedName {
			findings = append(findings, finding("warning", "future_phase_shaped_filename", version, file.Name(), "migration after immutable history uses a phase-shaped filename"))
		}
	}

	for _, entry := range manifest.Entries {
		if _, ok := sourceByVersion[entry.Version]; !ok {
			findings = append(findings, finding("blocking", "manifest_version_not_in_source", entry.Version, entry.Filename, "manifest entry has no embedded migration source"))
		}
	}

	versions := make([]int64, 0, len(sourceByVersion))
	for version := range sourceByVersion {
		versions = append(versions, version)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	for index, version := range versions {
		expected := int64(index + 1)
		audit := sourceByVersion[version]
		if version != expected {
			findings = append(findings, finding("blocking", "source_version_gap", version, audit.Filename, fmt.Sprintf("expected contiguous version %05d", expected)))
			break
		}
	}
	sort.Slice(audits, func(i, j int) bool {
		if audits[i].Version != audits[j].Version {
			return audits[i].Version < audits[j].Version
		}
		return audits[i].Filename < audits[j].Filename
	})
	return audits, findings, nil
}

func collectGooseLedger(ctx context.Context, pool postgres.DB, manifestByVersion map[int64]manifestEntry, immutableThroughVersion int64) (GooseLedger, []Finding, error) {
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

func parseFilenameVersion(filename string) (int64, bool) {
	matches := filenamePattern.FindStringSubmatch(filename)
	if len(matches) != 2 {
		return 0, false
	}
	var version int64
	for _, r := range matches[1] {
		version = version*10 + int64(r-'0')
	}
	return version, true
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
