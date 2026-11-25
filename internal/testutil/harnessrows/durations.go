package harnessrows

// DurationsSchema identifies the retained test durations document
// (TH-HARNESS-REQ-869). Durations are diagnostics only and never feed
// outcomes or scheduling.
const DurationsSchema = "cartulary.harness_go_test_durations.v1"

const (
	maxDurationTests     = 4096
	maxDurationNameBytes = 512
)

// TestDuration is one test's terminal event.
type TestDuration struct {
	Name       string `json:"name"`
	Action     string `json:"action"`
	DurationMS int64  `json:"duration_ms"`
}

// Durations is one invocation's record: every terminal test in the order
// its first terminal event arrived, and the package's own outcome.
type Durations struct {
	Package          string         `json:"package"`
	Rows             []Row          `json:"rows"`
	PackageAction    *string        `json:"package_action"`
	PackageElapsedMS *int64         `json:"package_elapsed_ms"`
	Tests            []TestDuration `json:"tests"`
}

// DurationsDocument is the retained form of a unit's invocations.
type DurationsDocument struct {
	SchemaID    string         `json:"schema_id"`
	UnitID      *string        `json:"unit_id"`
	Limits      durationLimits `json:"limits"`
	Omitted     int            `json:"omitted"`
	Invocations []Durations    `json:"invocations"`
}

type durationLimits struct {
	MaxTests     int `json:"max_tests"`
	MaxNameBytes int `json:"max_name_bytes"`
}

func (s *Stream) durations(packagePath string) Durations {
	record := Durations{Package: s.packageName, Rows: s.rows, Tests: []TestDuration{}}
	if record.Package == "" {
		record.Package = packagePath
	}
	if s.packageOutcome != nil {
		record.PackageAction = &s.packageOutcome.Action
		record.PackageElapsedMS = s.packageElapsed
	}
	for _, test := range s.order {
		observation := s.terminal[test]
		record.Tests = append(record.Tests, TestDuration{Name: test, Action: observation.Action, DurationMS: observation.DurationMS})
	}
	return record
}

// NewDurationsDocument bounds and redacts a unit's records. A name that is
// empty or too long after redaction, and every test past the bound, is
// counted as omitted, never truncated.
func NewDurationsDocument(unitID string, records []Durations, redact func(string) string) DurationsDocument {
	document := DurationsDocument{SchemaID: DurationsSchema, Limits: durationLimits{MaxTests: maxDurationTests, MaxNameBytes: maxDurationNameBytes}, Invocations: []Durations{}}
	if unitID != "" {
		document.UnitID = &unitID
	}
	remaining := maxDurationTests
	for _, record := range records {
		tests := []TestDuration{}
		for _, test := range record.Tests {
			name := redact(test.Name)
			if remaining == 0 || name == "" || len(name) > maxDurationNameBytes {
				document.Omitted++
				continue
			}
			remaining--
			test.Name = name
			tests = append(tests, test)
		}
		record.Tests = tests
		document.Invocations = append(document.Invocations, record)
	}
	return document
}
