package runevidence

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

func environment(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func privateDirectory(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	if err := os.Chmod(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	return directory
}

// TestContractEvidenceRunRootResolution covers the single run-root rule shared
// with harness fixture artifacts (TH-HARNESS-REQ-860).
func TestContractEvidenceRunRootResolution(t *testing.T) {
	results := privateDirectory(t)
	root, err := RunRoot(environment(map[string]string{resultsDirEnv: results, runIDEnv: "run-1"}))
	if err != nil || root != filepath.Join(results, "run-1") {
		t.Fatalf("absolute results root = %q, %v", root, err)
	}
	repository, err := RepositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	relative := filepath.Join("tmp", "runevidence-unused-"+strconv.Itoa(os.Getpid()))
	root, err = RunRoot(environment(map[string]string{resultsDirEnv: relative, runIDEnv: "run-2"}))
	if err != nil || root != filepath.Join(repository, relative, "run-2") {
		t.Fatalf("relative results root = %q, %v", root, err)
	}
	if ConfiguredResultsRoot(environment(nil)) != filepath.Join(".cartulary", "test-results") {
		t.Fatal("default results root changed")
	}
	for _, id := range []string{"", "..", ".", "a/b", "../escape", `a\b`, "a:b"} {
		if _, err := RunRoot(environment(map[string]string{resultsDirEnv: results, runIDEnv: id})); err == nil {
			t.Fatalf("run ID %q accepted", id)
		}
	}
	open := filepath.Join(results, "world")
	if err := os.Mkdir(open, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(open, 0o777); err != nil {
		t.Fatal(err)
	}
	if _, err := RunRoot(environment(map[string]string{resultsDirEnv: open, runIDEnv: "run-3"})); err == nil {
		t.Fatal("world-writable results root accepted")
	}
}

// TestContractEvidenceRecordsOwnerOnlySchemaLines covers AC-154 and AC-158:
// in-process and server-harness writers emit the same closed schema,
// owner-only, one file per process, with no request data. A server fatal
// carries its reason code and no operation, status or error code.
func TestContractEvidenceRecordsOwnerOnlySchemaLines(t *testing.T) {
	results := privateDirectory(t)
	runRoot := filepath.Join(results, "run")
	if err := os.Mkdir(runRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	withUnit := environment(map[string]string{resultsDirEnv: results, runIDEnv: "run", workUnitEnv: "go:module.auth:1"})
	withoutUnit := environment(map[string]string{resultsDirEnv: results, runIDEnv: "run"})
	for _, write := range []struct {
		lookup   func(string) (string, bool)
		evidence Evidence
	}{
		{withUnit, Evidence{Kind: UndeclaredErrorCode, OperationID: "createIncident", Status: 409, Code: "record_locked", Process: InProcess}},
		{withoutUnit, Evidence{Kind: ErrorEmission, OperationID: "createIncident", Status: 404, Code: "not_found", Process: ServerHarness}},
		{withoutUnit, Evidence{Kind: SecurityMismatch, OperationID: "getCurrentAccountProfile", Status: 200, Security: &SecurityFact{Declared: "session", Observed: "public"}, Process: ServerHarness}},
		{withoutUnit, Evidence{Kind: SecurityMismatch, OperationID: "createIncident", Status: 201, Security: &SecurityFact{Declared: "session_csrf", Observed: "session"}, Process: InProcess}},
		{withUnit, Evidence{Kind: ServerFatal, ReasonCode: "application_process_lease_lost", Process: InProcess}},
		{withoutUnit, Evidence{Kind: ServerFatal, ReasonCode: "indeterminate_database_commit", Process: ServerHarness}},
	} {
		if err := RecordTo(write.lookup, write.evidence); err != nil {
			t.Fatal(err)
		}
	}
	directory := filepath.Join(runRoot, Directory)
	file := filepath.Join(directory, strconv.Itoa(os.Getpid())+".jsonl")
	for path, mode := range map[string]os.FileMode{directory: 0o700, file: 0o600} {
		info, err := os.Lstat(path)
		if err != nil || info.Mode().Perm() != mode {
			t.Fatalf("%s mode = %v, %v", path, info, err)
		}
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSuffix(data, []byte("\n")), []byte("\n"))
	if len(lines) != 6 || !bytes.HasSuffix(data, []byte("\n")) {
		t.Fatalf("records = %q", data)
	}
	for i, line := range lines {
		value, err := harnessschema.Decode(line)
		if err != nil || harnessschema.Validate(SchemaID, value) != nil {
			t.Fatalf("record %d is not %s: %s", i, SchemaID, line)
		}
		record := value.(map[string]any)
		if record["pid"] != float64(os.Getpid()) || (i == 0 || i == 4) != (record["work_unit_id"] != nil) {
			t.Fatalf("record %d identity = %s", i, line)
		}
		_, operation := record["operation_id"]
		_, status := record["status"]
		_, reason := record["reason_code"]
		if fatal := record["kind"] == string(ServerFatal); fatal == operation || fatal == status || fatal != reason {
			t.Fatalf("record %d shape = %s", i, line)
		}
	}
	if fatal := lines[4]; !bytes.Contains(fatal, []byte(`"reason_code":"application_process_lease_lost"`)) || !bytes.Contains(fatal, []byte(`"process_kind":"in_process"`)) || bytes.Contains(fatal, []byte(`"code"`)) {
		t.Fatalf("server fatal record = %s", fatal)
	}
}

// TestContractEvidenceNeverWritesOutsideRunRoot covers AC-154's record outside
// the run root and record carrying request data.
func TestContractEvidenceNeverWritesOutsideRunRoot(t *testing.T) {
	valid := Evidence{Kind: UndeclaredErrorCode, OperationID: "createIncident", Status: 409, Code: "record_locked", Process: InProcess}
	t.Run("missing identity or run root writes nothing", func(t *testing.T) {
		results := privateDirectory(t)
		if err := RecordTo(environment(map[string]string{resultsDirEnv: results}), valid); err == nil {
			t.Fatal("record without run ID accepted")
		}
		if err := RecordTo(environment(map[string]string{resultsDirEnv: results, runIDEnv: "absent"}), valid); err == nil {
			t.Fatal("record created a run root")
		}
		if entries, _ := os.ReadDir(results); len(entries) != 0 {
			t.Fatalf("results root changed: %v", entries)
		}
	})
	t.Run("linked stream directory is refused", func(t *testing.T) {
		results, outside := privateDirectory(t), privateDirectory(t)
		runRoot := filepath.Join(results, "run")
		if err := os.Mkdir(runRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(runRoot, Directory)); err != nil {
			t.Fatal(err)
		}
		if err := RecordTo(environment(map[string]string{resultsDirEnv: results, runIDEnv: "run"}), valid); err == nil {
			t.Fatal("linked stream directory accepted")
		}
		if entries, _ := os.ReadDir(outside); len(entries) != 0 {
			t.Fatalf("record escaped the run root: %v", entries)
		}
	})
	t.Run("linked stream file is refused", func(t *testing.T) {
		results, outside := privateDirectory(t), privateDirectory(t)
		directory := filepath.Join(results, "run", Directory)
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(outside, "target")
		if err := os.WriteFile(target, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(directory, strconv.Itoa(os.Getpid())+".jsonl")); err != nil {
			t.Fatal(err)
		}
		if err := RecordTo(environment(map[string]string{resultsDirEnv: results, runIDEnv: "run"}), valid); err == nil {
			t.Fatal("linked stream file accepted")
		}
		if data, _ := os.ReadFile(target); len(data) != 0 {
			t.Fatal("record escaped through a link")
		}
	})
	t.Run("request data is not admitted", func(t *testing.T) {
		results := privateDirectory(t)
		if err := os.Mkdir(filepath.Join(results, "run"), 0o700); err != nil {
			t.Fatal(err)
		}
		lookup := environment(map[string]string{resultsDirEnv: results, runIDEnv: "run"})
		for _, evidence := range []Evidence{
			{Kind: UndeclaredErrorCode, OperationID: "createIncident", Status: 409, Code: "Bearer sentinel-token", Process: InProcess},
			{Kind: UndeclaredErrorCode, OperationID: "/api/v1/incidents?q=sentinel", Status: 409, Code: "record_locked", Process: InProcess},
			{Kind: UndeclaredErrorCode, OperationID: "createIncident", Status: 9000, Code: "record_locked", Process: InProcess},
			{Kind: SecurityMismatch, OperationID: "createIncident", Status: 200, Security: &SecurityFact{Declared: "session", Observed: "cookie=sentinel"}, Process: ServerHarness},
			{Kind: "free_text", OperationID: "createIncident", Status: 409, Code: "record_locked", Process: InProcess},
			{Kind: UndeclaredErrorCode, OperationID: "createIncident", Status: 409, Code: "record_locked", Process: "browser"},
			{Kind: UndeclaredErrorCode, OperationID: "createIncident", Status: 409, Code: "record_locked", ReasonCode: "published_component_lost", Process: InProcess},
			{Kind: ServerFatal, ReasonCode: "lease lost: Bearer sentinel-token", Process: InProcess},
			{Kind: ServerFatal, ReasonCode: "", Process: InProcess},
			{Kind: ServerFatal, ReasonCode: "published_component_lost", OperationID: "createIncident", Process: InProcess},
			{Kind: ServerFatal, ReasonCode: "published_component_lost", Status: 503, Process: ServerHarness},
			{Kind: ServerFatal, ReasonCode: "published_component_lost", Code: "service_unavailable", Process: ServerHarness},
			{Kind: ServerFatal, ReasonCode: "published_component_lost", Security: &SecurityFact{Declared: "session", Observed: "public"}, Process: ServerHarness},
			{Kind: ServerFatal, ReasonCode: "published_component_lost", Process: "browser"},
		} {
			if err := RecordTo(lookup, evidence); err == nil {
				t.Fatalf("evidence %+v accepted", evidence)
			}
		}
		if _, err := os.Lstat(filepath.Join(results, "run", Directory)); !os.IsNotExist(err) {
			t.Fatal("rejected evidence created the stream")
		}
	})
}

// TestContractEvidenceDeduplicatesEmissionsPerProcess covers
// TH-HARNESS-REQ-862 and TH-HARNESS-AC-157: a process writes one
// error_emission record per distinct operation, status and code for its run
// root; violations are written every time; a failed write is retried rather
// than remembered.
func TestContractEvidenceDeduplicatesEmissionsPerProcess(t *testing.T) {
	results := privateDirectory(t)
	lookup := environment(map[string]string{resultsDirEnv: results, runIDEnv: "run"})
	emission := Evidence{Kind: ErrorEmission, OperationID: "createIncident", Status: 409, Code: "client_txn_conflict", Process: InProcess}
	violation := Evidence{Kind: UndeclaredErrorCode, OperationID: "createIncident", Status: 409, Code: "record_locked", Process: InProcess}
	if err := RecordTo(lookup, emission); err == nil {
		t.Fatal("emission without a run root accepted")
	}
	if err := os.Mkdir(filepath.Join(results, "run"), 0o700); err != nil {
		t.Fatal(err)
	}
	otherStatus, otherCode := emission, emission
	otherStatus.Status = 400
	otherStatus.Code = "invalid_incident_create"
	otherCode.Code = "incident_key_conflict"
	for _, evidence := range []Evidence{emission, emission, violation, otherStatus, emission, violation, otherCode, otherStatus} {
		if err := RecordTo(lookup, evidence); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(results, "run", Directory, strconv.Itoa(os.Getpid())+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, line := range bytes.Split(bytes.TrimSuffix(data, []byte("\n")), []byte("\n")) {
		value, err := harnessschema.Decode(line)
		if err != nil || harnessschema.Validate(SchemaID, value) != nil {
			t.Fatalf("record is not %s: %s", SchemaID, line)
		}
		record := value.(map[string]any)
		counts[record["kind"].(string)+" "+strconv.Itoa(int(record["status"].(float64)))+" "+record["code"].(string)]++
	}
	want := map[string]int{
		"error_emission 409 client_txn_conflict":     1,
		"error_emission 400 invalid_incident_create": 1,
		"error_emission 409 incident_key_conflict":   1,
		"undeclared_error_code 409 record_locked":    2,
	}
	if len(counts) != len(want) {
		t.Fatalf("records = %v, want %v", counts, want)
	}
	for key, count := range want {
		if counts[key] != count {
			t.Fatalf("records = %v, want %v", counts, want)
		}
	}
	t.Run("another run root records the emission again", func(t *testing.T) {
		other := privateDirectory(t)
		if err := os.Mkdir(filepath.Join(other, "run"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := RecordTo(environment(map[string]string{resultsDirEnv: other, runIDEnv: "run"}), emission); err != nil {
			t.Fatal(err)
		}
		if data, err := os.ReadFile(filepath.Join(other, "run", Directory, strconv.Itoa(os.Getpid())+".jsonl")); err != nil || bytes.Count(data, []byte("\n")) != 1 {
			t.Fatalf("emission under another run root = %q, %v", data, err)
		}
	})
}
