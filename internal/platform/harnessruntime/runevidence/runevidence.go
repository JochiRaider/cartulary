// Package runevidence owns the harness run-root rule and the contract evidence
// stream (TH-HARNESS-REQ-860), including a test server's lifecycle fatal
// (TH-HARNESS-REQ-863). In-process test servers and the server-harness build
// share it; it imports no test helper and no product owner.
package runevidence

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

const (
	// SchemaID identifies one contract evidence record.
	SchemaID = "cartulary.harness_contract_evidence.v1"
	// Directory is the run-root-relative directory holding one file per process.
	Directory = "contract-evidence"

	resultsDirEnv     = "CARTULARY_TEST_RESULTS_DIR"
	runIDEnv          = "CARTULARY_TEST_RUN_ID"
	workUnitEnv       = "CARTULARY_WORK_UNIT_ID"
	defaultResultsDir = ".cartulary/test-results"
)

// Kind is the closed record kind. Only error emission is not a violation; a
// server fatal fails the run with its own reason (TH-HARNESS-REQ-863).
type Kind string

const (
	UndeclaredErrorCode Kind = "undeclared_error_code"
	SecurityMismatch    Kind = "security_mismatch"
	ServerFatal         Kind = "server_fatal"
	ErrorEmission       Kind = "error_emission"
)

// ProcessKind names the writer's process class.
type ProcessKind string

const (
	InProcess     ProcessKind = "in_process"
	ServerHarness ProcessKind = "server_harness"
)

// SecurityFact is the closed declared-versus-observed authentication fact of a
// security mismatch. Values are scheme classes, never credentials: public,
// session (bearer session or session cookie without CSRF), session_csrf
// (session cookie with CSRF verification) or credential_bootstrap.
type SecurityFact struct {
	Declared string `json:"declared"`
	Observed string `json:"observed"`
}

// Evidence is one guard finding or one lifecycle fatal. It has no field for
// request data. A server fatal carries only its closed reason code token; it
// names no operation, status or error code.
type Evidence struct {
	Kind        Kind
	OperationID string
	Status      int
	Code        string
	Security    *SecurityFact
	ReasonCode  string
	Process     ProcessKind
}

type record struct {
	SchemaID    string        `json:"schema_id"`
	Kind        Kind          `json:"kind"`
	OperationID string        `json:"operation_id,omitempty"`
	Status      int           `json:"status,omitempty"`
	Code        string        `json:"code,omitempty"`
	Security    *SecurityFact `json:"security,omitempty"`
	ReasonCode  string        `json:"reason_code,omitempty"`
	ProcessKind ProcessKind   `json:"process_kind"`
	WorkUnitID  *string       `json:"work_unit_id"`
	PID         int           `json:"pid"`
}

var (
	operationPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,127}$`)
	codePattern      = regexp.MustCompile(`^[a-z][a-z0-9_]{0,127}$`)
	unitPattern      = regexp.MustCompile(`^[\x20-\x7e]{1,255}$`)
	securityClasses  = map[string]bool{"public": true, "session": true, "session_csrf": true, "credential_bootstrap": true}
	// One process appends complete lines under one lock; the descriptor is
	// reopened per record so a removed or replaced directory is never reused.
	writeMu sync.Mutex
	// emitted holds each distinct error emission this process has written,
	// keyed by destination and emission (TH-HARNESS-REQ-862). Guarded by
	// writeMu.
	emitted = map[string]struct{}{}
)

// ConfiguredResultsRoot returns the configured result parent, which may be
// relative to the repository root.
func ConfiguredResultsRoot(lookup func(string) (string, bool)) string {
	if value, _ := lookup(resultsDirEnv); strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return filepath.FromSlash(defaultResultsDir)
}

// RepositoryRoot finds the nearest ancestor of the working directory holding go.mod.
func RepositoryRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	for current := wd; ; current = filepath.Dir(current) {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		if filepath.Dir(current) == current {
			return "", fmt.Errorf("resolve repo root from %s: go.mod not found", wd)
		}
	}
}

// ResultsRoot resolves the harness result parent without side effects. A custom
// absolute parent may not be world-writable without the sticky bit. Narrowing
// an existing default parent belongs to the test-support owner of the results
// root (suiteservices.ResolveResultsRoot), not to production code.
func ResultsRoot(lookup func(string) (string, bool)) (string, error) {
	configured := ConfiguredResultsRoot(lookup)
	if filepath.IsAbs(configured) {
		if err := validateResultsRoot(configured, true); err != nil {
			return "", err
		}
		return configured, nil
	}
	repository, err := RepositoryRoot()
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(repository, configured)
	if err := validateResultsRoot(resolved, false); err != nil {
		return "", err
	}
	return resolved, nil
}

func validateResultsRoot(resolved string, custom bool) error {
	existing := resolved
	if _, err := os.Stat(existing); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat result root %s: %w", resolved, err)
		}
		existing = filepath.Dir(resolved)
	}
	info, err := os.Stat(existing)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat result root parent %s: %w", existing, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("result root parent %s is not a directory", existing)
	}
	if custom && info.Mode().Perm()&0o002 != 0 && info.Mode()&os.ModeSticky == 0 {
		return fmt.Errorf("result root %s must not be world-writable without sticky bit", existing)
	}
	return nil
}

// RunRoot resolves the selected run directory. An explicit, single-component
// run identity is mandatory; there is no ad hoc fallback.
func RunRoot(lookup func(string) (string, bool)) (string, error) {
	id, _ := lookup(runIDEnv)
	if id == "" {
		return "", errors.New("fixture execution requires an explicit run ID")
	}
	root, err := ResultsRoot(lookup)
	if err != nil {
		return "", err
	}
	id = strings.TrimSpace(id)
	if id == "" || !fs.ValidPath(id) || id == "." || strings.ContainsAny(id, "/\\:\x00") || !filepath.IsLocal(id) {
		return "", errors.New("invalid run ID")
	}
	return filepath.Join(root, id), nil
}

// Validate admits only the closed record shape before any byte is written.
func (e Evidence) validate() error {
	if e.Process != InProcess && e.Process != ServerHarness {
		return errors.New("invalid contract evidence process kind")
	}
	if e.Kind == ServerFatal {
		// A lifecycle fatal names only its reason code, a closed token.
		if !codePattern.MatchString(e.ReasonCode) || e.OperationID != "" || e.Status != 0 || e.Code != "" || e.Security != nil {
			return errors.New("invalid contract evidence server fatal")
		}
		return nil
	}
	if !operationPattern.MatchString(e.OperationID) || e.Status < 100 || e.Status > 599 || e.ReasonCode != "" {
		return errors.New("invalid contract evidence operation")
	}
	switch e.Kind {
	case UndeclaredErrorCode, ErrorEmission:
		if !codePattern.MatchString(e.Code) || e.Security != nil {
			return errors.New("invalid contract evidence error code")
		}
	case SecurityMismatch:
		if e.Code != "" || e.Security == nil || !securityClasses[e.Security.Declared] || !securityClasses[e.Security.Observed] {
			return errors.New("invalid contract evidence security fact")
		}
	default:
		return errors.New("invalid contract evidence kind")
	}
	return nil
}

// Record appends one finding to <run root>/contract-evidence/<pid>.jsonl using
// the process environment. Callers treat failure as best effort; a missing run
// identity or unsafe run root writes nothing. An error emission is written at
// most once per process for each run root, operation, status and code; a
// violation or server fatal is written every time, and the run gate counts its
// occurrences. A fatal observer writes once per lifecycle fatal.
func Record(e Evidence) error {
	return RecordTo(os.LookupEnv, e)
}

// RecordTo is Record with an explicit environment lookup.
func RecordTo(lookup func(string) (string, bool), e Evidence) error {
	if err := e.validate(); err != nil {
		return err
	}
	// The destination is named by the raw run identity, so a repeated
	// emission costs no filesystem access. Only a written record is marked;
	// a failed write is retried by the next emission.
	var emission string
	if e.Kind == ErrorEmission {
		results, _ := lookup(resultsDirEnv)
		id, _ := lookup(runIDEnv)
		emission = strings.Join([]string{results, id, string(e.Process), e.OperationID, strconv.Itoa(e.Status), e.Code}, "\x00")
		writeMu.Lock()
		_, seen := emitted[emission]
		writeMu.Unlock()
		if seen {
			return nil
		}
	}
	runRoot, err := RunRoot(lookup)
	if err != nil {
		return err
	}
	value := record{SchemaID: SchemaID, Kind: e.Kind, OperationID: e.OperationID, Status: e.Status, Code: e.Code, Security: e.Security, ReasonCode: e.ReasonCode, ProcessKind: e.Process, PID: os.Getpid()}
	if unit, _ := lookup(workUnitEnv); unitPattern.MatchString(unit) {
		value.WorkUnitID = &unit
	}
	line, err := json.Marshal(value)
	if err != nil {
		return err
	}
	writeMu.Lock()
	defer writeMu.Unlock()
	if _, seen := emitted[emission]; seen && emission != "" {
		return nil
	}
	if err := appendLine(runRoot, strconv.Itoa(os.Getpid())+".jsonl", append(line, '\n')); err != nil {
		return err
	}
	if emission != "" {
		emitted[emission] = struct{}{}
	}
	return nil
}

// appendLine confines every operation to an existing owner-controlled run root.
// It never creates the run root and never follows a link out of it.
func appendLine(runRoot, name string, line []byte) error {
	info, err := os.Lstat(runRoot)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || !ownedByCaller(info) {
		return errors.New("unsafe contract evidence run root")
	}
	root, err := os.OpenRoot(runRoot)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Mkdir(Directory, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	directory, err := root.Lstat(Directory)
	if err != nil {
		return err
	}
	if !directory.IsDir() || directory.Mode()&os.ModeSymlink != 0 || directory.Mode().Perm() != 0o700 || !ownedByCaller(directory) {
		return errors.New("unsafe contract evidence directory")
	}
	file, err := root.OpenFile(Directory+"/"+name, os.O_WRONLY|os.O_CREATE|os.O_APPEND|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600)
	if err != nil {
		return err
	}
	opened, err := file.Stat()
	if err == nil && (!opened.Mode().IsRegular() || opened.Mode().Perm() != 0o600 || !ownedByCaller(opened) || links(opened) != 1) {
		err = errors.New("unsafe contract evidence file")
	}
	if err == nil {
		_, err = file.Write(line)
	}
	return errors.Join(err, file.Close())
}

func ownedByCaller(info fs.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && int(stat.Uid) == os.Getuid()
}

func links(info fs.FileInfo) uint64 {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return uint64(stat.Nlink)
	}
	return 0
}
