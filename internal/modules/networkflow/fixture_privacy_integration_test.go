package networkflow_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	logglobal "go.opentelemetry.io/otel/log/global"
)

type privacyFixtureCase struct {
	publicationFixtureCase
	ForbiddenValues              []string `json:"forbidden_values"`
	ForbiddenObservabilityValues []string `json:"forbidden_observability_values"`
}

// The TLS collector receives records exported by the real application observer.
// The decorator independently captures the complete semantic telemetry payload.
// Both are retained; neither replaces the production observer's work.
type privacyFixtureCapture struct {
	GraphTelemetryObserver
	mu             sync.Mutex
	records        []map[string]any
	exports        []json.RawMessage
	collectorError error
	observations   []any
	overflow       bool
}

func (p *privacyFixtureCapture) observe(value any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.observations) >= 1024 {
		p.overflow = true
		return
	}
	p.observations = append(p.observations, value)
}
func (p *privacyFixtureCapture) ObserveGraphPhase(ctx context.Context, v GraphPhaseTelemetryObservation) {
	p.observe(v)
	p.GraphTelemetryObserver.ObserveGraphPhase(ctx, v)
}
func (p *privacyFixtureCapture) ObserveGraphResult(ctx context.Context, v GraphResultTelemetryObservation) {
	p.observe(v)
	p.GraphTelemetryObserver.ObserveGraphResult(ctx, v)
}
func (p *privacyFixtureCapture) ObserveGraphCleanup(ctx context.Context, v GraphCleanupTelemetryObservation) {
	p.observe(v)
	p.GraphTelemetryObserver.ObserveGraphCleanup(ctx, v)
}
func privacyJSON(value any) (any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	err = decoder.Decode(&result)
	return result, err
}
func (p *privacyFixtureCapture) snapshot(started time.Time) (map[string]any, map[string]any, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.overflow {
		return nil, nil, fmt.Errorf("privacy capture overflow")
	}
	if p.collectorError != nil {
		return nil, nil, p.collectorError
	}
	rawLogs := []any{}
	logs := []any{}
	cleanupLogs := []any{}
	observations := []any{}
	cleanups := []any{}
	for _, entry := range p.records {
		rawLogs = append(rawLogs, entry)
		if entry["scope"] != "cartulary.network_flow" {
			continue
		}
		for _, key := range []string{"timestamp", "observed_timestamp"} {
			timestamp, err := time.Parse(time.RFC3339Nano, entry[key].(string))
			if err != nil || timestamp.Before(started) || timestamp.After(time.Now()) {
				return nil, nil, fmt.Errorf("invalid emitted log timestamp")
			}
		}
		normalized, err := privacyJSON(entry)
		if err != nil {
			return nil, nil, err
		}
		item := normalized.(map[string]any)
		item["timestamp"] = "$emitted-at"
		item["observed_timestamp"] = "$observed-at"
		if entry["attributes"].(map[string]any)["cartulary.operation"] == "cleanup_sweep" {
			cleanupLogs = append(cleanupLogs, item)
		} else {
			logs = append(logs, item)
		}
	}
	for _, observation := range p.observations {
		normalized, err := privacyJSON(observation)
		if err != nil {
			return nil, nil, err
		}
		item := normalized.(map[string]any)
		if duration, ok := item["Duration"]; ok {
			n, err := duration.(json.Number).Int64()
			if err != nil || n < 0 {
				return nil, nil, fmt.Errorf("invalid telemetry duration")
			}
			item["Duration"] = "$nonnegative-duration"
		}
		if item["Operation"] == "cleanup_sweep" {
			cleanups = append(cleanups, item)
		} else {
			observations = append(observations, item)
		}
	}
	return map[string]any{"logs": logs, "cleanup_logs": cleanupLogs, "telemetry": observations, "cleanup_telemetry": cleanups}, map[string]any{"logs": rawLogs, "telemetry": p.observations, "otlp_exports": p.exports}, nil
}

type privacyLogBuffer struct {
	mu sync.Mutex
	bytes.Buffer
}

func (b *privacyLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.Write(p)
}
func (b *privacyLogBuffer) text() string { b.mu.Lock(); defer b.mu.Unlock(); return b.Buffer.String() }

func rejectPrivacyValues(value any, forbidden []string) error {
	switch v := value.(type) {
	case string:
		for _, raw := range forbidden {
			if raw == "" {
				return fmt.Errorf("empty forbidden privacy value")
			}
			if strings.Contains(v, raw) {
				return fmt.Errorf("restricted channel contains forbidden fixture value")
			}
		}
	case []any:
		for _, item := range v {
			if err := rejectPrivacyValues(item, forbidden); err != nil {
				return err
			}
		}
	case map[string]any:
		for key, item := range v {
			if err := rejectPrivacyValues(key, forbidden); err != nil {
				return err
			}
			if err := rejectPrivacyValues(item, forbidden); err != nil {
				return err
			}
		}
	}
	return nil
}
func privacyProjectionMetadata(result map[string]any) (map[string]any, error) {
	graph, ok := result["graph"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("privacy fixture requires a public graph")
	}
	projection := graph["graph_projection_result"].(map[string]any)
	metadata := map[string]any{"mapped_metadata": projection["mapped_metadata"]}
	for _, kind := range []string{"vertices", "edges"} {
		values := []any{}
		for _, raw := range projection[kind].([]any) {
			values = append(values, raw.(map[string]any)["metadata"])
		}
		metadata[kind] = values
	}
	return metadata, nil
}
func TestNetworkFlowPrivacyFixtureComparison(t *testing.T) {
	if runPrivacyFixtureChild(t) {
		return
	}
	RunFixtureComparison(t, "NF-FIX-016-redaction", 9, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		data, err := bundle.Read("source/cases.json")
		if err != nil {
			return nil, err
		}
		var inputs []privacyFixtureCase
		if err := json.Unmarshal(data, &inputs); err != nil {
			return nil, err
		}
		outcomes := []map[string]any{}
		for _, input := range inputs {
			var outcome map[string]any
			var failure error
			passed := t.Run(input.Name, func(t *testing.T) {
				started := time.Now()
				capture := &privacyFixtureCapture{}
				input.environment = capture.startCollector(t)
				standard := &privacyLogBuffer{}
				previousOutput := log.Writer()
				log.SetOutput(standard)
				t.Cleanup(func() { log.SetOutput(previousOutput) })
				configure := func(d ModuleDependencies) ModuleDependencies {
					if d.GraphTelemetry == nil {
						t.Fatal("missing application telemetry observer")
					}
					capture.GraphTelemetryObserver = d.GraphTelemetry
					d.GraphTelemetry = capture
					return d
				}
				outcome, failure = runPublicationFixture(t, bundle, input.publicationFixtureCase, nil, func(r *publicationFixtureRuntime, result map[string]any) error {
					provider, ok := logglobal.GetLoggerProvider().(interface{ ForceFlush(context.Context) error })
					if !ok {
						return fmt.Errorf("configured log exporter is not active")
					}
					ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
					defer cancel()
					if err := provider.ForceFlush(ctx); err != nil {
						return err
					}
					normalized, raw, err := capture.snapshot(started)
					if err != nil {
						return err
					}
					raw["standard_log"] = standard.text()
					normalized["standard_log"] = standard.text()
					r.transport = append(r.transport, map[string]any{"privacy_capture": raw})
					metadata, err := privacyProjectionMetadata(result)
					if err != nil {
						return err
					}
					state := r.comparableState(r.durableState())
					restricted := map[string]any{"capture": raw, "audit": state["audit"], "projection_metadata": metadata}
					r.transport = append(r.transport, map[string]any{"privacy_restricted_channels": restricted})
					value, err := privacyJSON(restricted)
					if err != nil {
						return err
					}
					if err := rejectPrivacyValues(value, input.ForbiddenValues); err != nil {
						return err
					}
					observability, err := privacyJSON(map[string]any{"capture": raw, "projection_metadata": metadata})
					if err != nil {
						return err
					}
					if err := rejectPrivacyValues(observability, input.ForbiddenObservabilityValues); err != nil {
						return err
					}
					// A negative control proves the scan cannot quietly ignore forbidden text.
					if len(input.ForbiddenValues) == 0 || rejectPrivacyValues(map[string]any{"log": input.ForbiddenValues[0]}, input.ForbiddenValues) == nil {
						return fmt.Errorf("privacy negative control failed")
					}
					result["privacy"] = map[string]any{"capture": normalized, "projection_metadata": metadata, "final_state": state, "forbidden_values_checked": input.ForbiddenValues, "forbidden_observability_values_checked": input.ForbiddenObservabilityValues}
					return nil
				}, configure)
			})
			if failure != nil {
				return nil, failure
			}
			if !passed {
				return nil, fmt.Errorf("%s: privacy fixture failed", input.Name)
			}
			outcomes = append(outcomes, outcome)
		}
		return outcomes, nil
	})
}
