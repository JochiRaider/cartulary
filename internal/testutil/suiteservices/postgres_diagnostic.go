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
type PostgresStartupLogEvent struct {
	Stream             string  `json:"stream"`
	StreamSequence     int     `json:"stream_sequence"`
	Signal             string  `json:"signal"`
	Phase              string  `json:"phase"`
	Timestamp          *string `json:"timestamp"`
	ContainerElapsedMS *int64  `json:"container_elapsed_ms"`
}
type PostgresEndpoint struct {
	Ref      string `json:"ref"`
	Family   string `json:"family"`
	Category string `json:"category"`
}
type PostgresBindingObservation struct {
	Endpoint            PostgresEndpoint `json:"endpoint"`
	MatchesSelectedPort bool             `json:"matches_selected_port"`
}
type PostgresLookupObservation struct {
	ElapsedMS    int64              `json:"elapsed_ms"`
	SelectedHost bool               `json:"selected_host"`
	Outcome      string             `json:"outcome"`
	Addresses    []PostgresEndpoint `json:"addresses"`
}
type PostgresDialObservation struct {
	ElapsedMS           int64             `json:"elapsed_ms"`
	Endpoint            PostgresEndpoint  `json:"endpoint"`
	Peer                *PostgresEndpoint `json:"peer"`
	MatchesSelectedPort bool              `json:"matches_selected_port"`
	FromLookup          bool              `json:"from_lookup"`
	Outcome             string            `json:"outcome"`
}
type PostgresEndpointDiagnostic struct {
	Selected                 *PostgresEndpoint            `json:"selected"`
	InitialStatus            string                       `json:"initial_status"`
	FailureStatus            string                       `json:"failure_status"`
	InitialBindings          []PostgresBindingObservation `json:"initial_bindings"`
	FailureBindings          []PostgresBindingObservation `json:"failure_bindings"`
	BindingChanged           *bool                        `json:"binding_changed"`
	SelectedPortPublished    *bool                        `json:"selected_port_published"`
	ContainerIdentityMatches *bool                        `json:"container_identity_matches"`
	Lookups                  []PostgresLookupObservation  `json:"lookups"`
	Dials                    []PostgresDialObservation    `json:"dials"`
	InternalListener         string                       `json:"internal_listener"`
	PublishedDial            *PostgresDialObservation     `json:"published_dial"`
	Truncated                bool                         `json:"truncated"`
}
type PostgresProbeObservation struct {
	ElapsedMS int64  `json:"elapsed_ms"`
	Cause     string `json:"cause"`
}
type PostgresPublicationBinding struct {
	ContainerPort string `json:"container_port"`
	AddressScope  string `json:"address_scope"`
	PortKind      string `json:"port_kind"`
}

type PostgresPublicationObservation struct {
	Status          string                       `json:"status"`
	RequestedPorts  []string                     `json:"requested_ports"`
	EffectivePorts  []string                     `json:"effective_ports"`
	Bindings        []PostgresPublicationBinding `json:"bindings"`
	PublishAllPorts *bool                        `json:"publish_all_ports"`
}

type PostgresReadinessDiagnostic struct {
	Publication                PostgresPublicationObservation `json:"publication"`
	Endpoint                   PostgresEndpointDiagnostic     `json:"endpoint"`
	SchemaID                   string                         `json:"schema_id"`
	Stage                      string                         `json:"stage"`
	Outcome                    string                         `json:"outcome"`
	ElapsedMS                  int64                          `json:"elapsed_ms"`
	PollMS                     int64                          `json:"poll_ms"`
	ProbeCapMS                 int64                          `json:"probe_cap_ms"`
	ProbeCount                 int                            `json:"probe_count"`
	CleanupOutcome             string                         `json:"cleanup_outcome"`
	ContainerState             string                         `json:"container_state"`
	Phases                     []PostgresReadinessPhase       `json:"phases"`
	ProbeFailures              map[string]int                 `json:"probe_failures"`
	LastProbeFailure           string                         `json:"last_probe_failure,omitempty"`
	LastSQLState               string                         `json:"last_sqlstate,omitempty"`
	ContainerExitCode          *int                           `json:"container_exit_code,omitempty"`
	ContainerOOMKilled         *bool                          `json:"container_oom_killed,omitempty"`
	StartupLogStatus           string                         `json:"startup_log_status"`
	StartupLogSignals          []string                       `json:"startup_log_signals"`
	StartupLogClock            string                         `json:"startup_log_clock"`
	StartupLogEvents           []PostgresStartupLogEvent      `json:"startup_log_events"`
	ContainerStartedAt         *string                        `json:"container_started_at"`
	ProbeObservations          []PostgresProbeObservation     `json:"probe_observations"`
	ProbeObservationsTruncated bool                           `json:"probe_observations_truncated"`
	FirstSQLConnectionMS       *int64                         `json:"first_sql_connection_ms"`
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
			err = harnessschema.Validate("cartulary.postgres_readiness_diagnostic.v6", value)
		}
	}
	if err == nil {
		err = json.Unmarshal(payload, &diagnostic)
	}
	return diagnostic, true, err
}
