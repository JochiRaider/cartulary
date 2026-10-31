package networkflow_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/google/uuid"
)

type linkFixtureIndicator struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	ValueKind string `json:"value_kind"`
	Value     string `json:"value"`
}
type linkFixtureStep struct {
	Name    string          `json:"name"`
	Request json.RawMessage `json:"request"`
	Replay  bool            `json:"replay"`
	Fault   string          `json:"fault,omitempty"`
	Actor   string          `json:"actor,omitempty"`
}
type linkFixtureCase struct {
	publicationFixtureCase
	ExistingIndicators []linkFixtureIndicator `json:"existing_indicators"`
	Links              []linkFixtureStep      `json:"links"`
}

// Raw opaque IDs remain in transport and must be stable across replay and every
// durable reference. Only frozen comparisons use aliases, as in fixture 027.
type linkFixtureAliases struct {
	values               map[string]string
	bindings, indicators int
}

func (a *linkFixtureAliases) admit(t *testing.T, response map[string]any, known map[string]bool) {
	t.Helper()
	if response["data"] == nil {
		return
	}
	binding := response["data"].(map[string]any)["binding"].(map[string]any)
	id := binding["network_flow_indicator_binding_id"].(string)
	if !strings.HasPrefix(id, "nfb_") || len(id) != 36 {
		t.Fatal("invalid binding identity")
	}
	if _, err := hex.DecodeString(id[4:]); err != nil || id != strings.ToLower(id) {
		t.Fatal("invalid binding identity bytes")
	}
	if _, ok := a.values[id]; !ok {
		a.bindings++
		a.values[id] = fmt.Sprintf("$binding:%02d", a.bindings)
	}
	indicator := binding["target_indicator_ref"].(map[string]any)["indicator_id"].(string)
	parsed, err := uuid.Parse(indicator)
	if err != nil || parsed == uuid.Nil || parsed.String() != indicator {
		t.Fatal("invalid Core indicator identity")
	}
	if _, ok := a.values[indicator]; !ok && !known[indicator] {
		a.indicators++
		a.values[indicator] = fmt.Sprintf("$indicator:%02d", a.indicators)
	}
}
func (a *linkFixtureAliases) apply(t *testing.T, value any) any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		t.Fatal(err)
	}
	var walk func(any) any
	walk = func(value any) any {
		switch value := value.(type) {
		case string:
			if replacement, ok := a.values[value]; ok {
				return replacement
			}
		case []any:
			for i, item := range value {
				value[i] = walk(item)
			}
		case map[string]any:
			for key, item := range value {
				value[key] = walk(item)
			}
		}
		return value
	}
	return walk(result)
}

func (r *publicationFixtureRuntime) linkState() map[string]any {
	return map[string]any{
		"network_flow": r.durableState(),
		"bindings":     r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(b) ORDER BY network_flow_indicator_binding_id),'[]') FROM network_flow_indicator_bindings b`),
		"indicators":   r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(i)||jsonb_build_object('record',to_jsonb(r)) ORDER BY i.record_id),'[]') FROM indicators i JOIN records r USING(record_id)`),
		"receipts":     r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(r)||jsonb_build_object('request_hash',encode(request_hash,'hex')) ORDER BY client_txn_id),'[]') FROM route_idempotency r WHERE route_key='nf.indicator_links.create'`),
		"counts":       r.jsonQuery(`SELECT jsonb_build_object('observations',(SELECT count(*) FROM indicator_observations),'lifecycle_intervals',(SELECT count(*) FROM indicator_state_intervals),'indicators',(SELECT count(*) FROM indicators),'bindings',(SELECT count(*) FROM network_flow_indicator_bindings),'link_receipts',(SELECT count(*) FROM route_idempotency WHERE route_key='nf.indicator_links.create'))`),
	}
}

func (r *publicationFixtureRuntime) comparableLinkState(raw map[string]any, aliases *linkFixtureAliases, known map[string]bool) any {
	state := map[string]any{}
	for k, v := range raw {
		state[k] = v
	}
	state["network_flow"] = r.comparableState(raw["network_flow"].(map[string]any))
	state["receipts"] = r.comparableState(map[string]any{"audit": raw["receipts"]})["audit"]
	indicators := []any{}
	for _, item := range raw["indicators"].([]any) {
		original := item.(map[string]any)
		indicator := map[string]any{}
		for k, v := range original {
			indicator[k] = v
		}
		// Dedupe-key representation belongs to Reference Data. Compare its public
		// canonical identity, exact dedupe effects and the complete raw row on replay.
		delete(indicator, "dedupe_key")
		record := map[string]any{}
		for k, v := range original["record"].(map[string]any) {
			record[k] = v
		}
		for _, key := range []string{"created_at", "updated_at"} {
			at, err := time.Parse(time.RFC3339Nano, record[key].(string))
			if err != nil {
				r.t.Fatal(err)
			}
			if known[indicator["record_id"].(string)] {
				if at.Before(r.started) || at.After(time.Now().UTC().Add(time.Second)) {
					r.t.Fatal("seeded Core timestamp outside setup interval")
				}
				record[key] = "$seeded-at"
			} else if !at.Equal(time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)) {
				r.t.Fatal("Core participant lost pinned operation time")
			}
		}
		indicator["record"] = record
		indicators = append(indicators, indicator)
	}
	state["indicators"] = indicators
	aliased := aliases.apply(r.t, state).(map[string]any)
	for _, collection := range []struct{ name, key string }{{"bindings", "network_flow_indicator_binding_id"}, {"indicators", "record_id"}} {
		rows := aliased[collection.name].([]any)
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].(map[string]any)[collection.key].(string) < rows[j].(map[string]any)[collection.key].(string)
		})
	}
	return aliased
}

func runLinkFixtureFamily(t *testing.T, bundle *networkflowfixtures.Bundle, after func(*publicationFixtureRuntime, linkFixtureCase, map[string]any, *linkFixtureAliases, map[string]bool) error) ([]map[string]any, error) {
	raw, err := bundle.Read("source/cases.json")
	if err != nil {
		return nil, err
	}
	var inputs []linkFixtureCase
	if err := json.Unmarshal(raw, &inputs); err != nil {
		return nil, err
	}
	outcomes := []map[string]any{}
	for _, input := range inputs {
		var outcome map[string]any
		var failure error
		passed := t.Run(input.Name, func(t *testing.T) {
			probe := &linkFixtureDependencyProbe{}
			configure := func(d ModuleDependencies) ModuleDependencies {
				probe.IndicatorParticipationPort = d.Indicators
				d.Indicators = probe
				return d
			}
			outcome, failure = runPublicationFixture(t, bundle, input.publicationFixtureCase, nil, func(r *publicationFixtureRuntime, result map[string]any) error {
				known := map[string]bool{}
				for _, indicator := range input.ExistingIndicators {
					known[indicator.ID] = true
					indicatortest.SeedRecord(t, r.h.DB, uuid.MustParse(input.IncidentID), uuid.MustParse(input.ActorID), uuid.MustParse(indicator.ID), indicator.Type, indicator.ValueKind, indicator.Value)
				}
				aliases := &linkFixtureAliases{values: map[string]string{}}
				steps := []any{}
				for _, step := range input.Links {
					if err := probe.arm(step.Fault); err != nil {
						return err
					}
					restore := func() {}
					switch step.Actor {
					case "":
					case "viewer":
						restore = r.alternateActor(input.IncidentID)
					default:
						return fmt.Errorf("unknown link fixture actor")
					}
					before := r.linkState()
					response := r.observedPost("/api/v1/incidents/"+input.IncidentID+"/network-flow/indicator-links", step.Request)
					state := r.linkState()
					if response["error"] != nil && !reflect.DeepEqual(before, state) {
						return fmt.Errorf("%s: rejected link mutated durable state", step.Name)
					}
					aliases.admit(t, response, known)
					if step.Replay {
						replay := r.observedPost("/api/v1/incidents/"+input.IncidentID+"/network-flow/indicator-links", step.Request)
						if !reflect.DeepEqual(response, replay) || !reflect.DeepEqual(state, r.linkState()) {
							return fmt.Errorf("%s: replay changed response, identities, audit or durable effects", step.Name)
						}
					}
					restore()
					observed := map[string]any{"name": step.Name, "request": step.Request, "response": aliases.apply(t, response), "replayed": step.Replay, "state": r.comparableLinkState(state, aliases, known)}
					if step.Actor != "" {
						observed["actor"] = step.Actor
					}
					if step.Fault != "" {
						summary, raw, err := probe.observation()
						if err != nil {
							return err
						}
						observed["fault"] = summary
						r.transport = append(r.transport, map[string]any{"injected_core_participant": raw})
					}
					steps = append(steps, observed)
				}
				result["link_steps"] = steps
				state := r.linkState()
				result["link_final_state"] = r.comparableLinkState(state, aliases, known)
				r.transport = append(r.transport, map[string]any{"link_final_state": state})
				if after != nil {
					if err := after(r, input, result, aliases, known); err != nil {
						return err
					}
				}
				return nil
			}, configure)
		})
		if failure != nil {
			return nil, failure
		}
		if !passed {
			return nil, fmt.Errorf("%s: link fixture failed", input.Name)
		}
		outcomes = append(outcomes, outcome)
	}
	return outcomes, nil
}

func TestNetworkFlowIndicatorLinkingFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-007-indicator-linking", 4, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return runLinkFixtureFamily(t, bundle, nil)
	})
}

func TestNetworkFlowIndicatorMismatchFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-017-indicator-link-mismatch", 3, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return runLinkFixtureFamily(t, bundle, nil)
	})
}
