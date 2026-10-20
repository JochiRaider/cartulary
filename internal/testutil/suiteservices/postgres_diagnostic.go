package suiteservices

import (
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

const PostgresReadinessExtensionKey = "cartulary.postgres_readiness"

type PostgresReadinessPhase struct {
	Stage            string `json:"stage"`
	Outcome          string `json:"outcome"`
	ConfiguredMS     int64  `json:"configured_ms"`
	EffectiveMS      int64  `json:"effective_ms"`
	LimitingDeadline string `json:"limiting_deadline"`
	ElapsedMS        int64  `json:"elapsed_ms"`
}
type PostgresReadinessDiagnostic struct {
	SchemaID       string                   `json:"schema_id"`
	Stage          string                   `json:"stage"`
	Outcome        string                   `json:"outcome"`
	ElapsedMS      int64                    `json:"elapsed_ms"`
	SQLWindowMS    int64                    `json:"sql_window_ms"`
	PollMS         int64                    `json:"poll_ms"`
	ProbeCapMS     int64                    `json:"probe_cap_ms"`
	ProbeCount     int                      `json:"probe_count"`
	CleanupOutcome string                   `json:"cleanup_outcome"`
	ContainerState string                   `json:"container_state"`
	Phases         []PostgresReadinessPhase `json:"phases"`
}

func postgresReadinessDiagnosticDetail(details map[string]any) (PostgresReadinessDiagnostic, bool, error) {
	raw, present := details[PostgresReadinessExtensionKey]
	if !present {
		return PostgresReadinessDiagnostic{}, false, nil
	}
	payload, err := json.Marshal(raw)
	var diagnostic PostgresReadinessDiagnostic
	if err == nil {
		var value any
		value, err = harnessschema.Decode(payload)
		if err == nil {
			err = harnessschema.Validate("cartulary.postgres_readiness_diagnostic.v1", value)
		}
	}
	if err == nil {
		err = json.Unmarshal(payload, &diagnostic)
	}
	return diagnostic, true, err
}
