package reference_data

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type consumerMemory struct {
	set        PackSet
	items      map[string]json.RawMessage
	lookups    int
	limit      int
	provenance map[string]PackProvenance
	hostItem   json.RawMessage
}

func (m *consumerMemory) CurrentSet(context.Context) (PackSet, error) { return m.set, nil }
func (m *consumerMemory) RetainedSet(_ context.Context, id string) (PackSet, error) {
	if id != m.set.ID {
		return PackSet{}, consumerError("pack_set_not_found")
	}
	return m.set, nil
}
func (m *consumerMemory) Provenance(_ context.Context, set, key string) (PackProvenance, error) {
	return m.provenance[key], nil
}
func (m *consumerMemory) Entry(_ context.Context, _ PackSetMember, id string) (json.RawMessage, error) {
	if v, ok := m.items[id]; ok {
		return v, nil
	}
	return nil, consumerError("entry_not_found")
}
func (m *consumerMemory) Lookup(_ context.Context, _ PackSetMember, _ packformat.LookupInput, after string, limit int) ([]indexedItem, error) {
	m.lookups++
	m.limit = limit
	if after != "" {
		return []indexedItem{}, nil
	}
	rows := make([]indexedItem, limit)
	for i := range rows {
		rows[i] = indexedItem{Item: m.hostItem, Position: "private-position"}
	}
	return rows, nil
}
func consumerFixture(t *testing.T) (*consumerMemory, Consumer, *time.Time) {
	t.Helper()
	_, versions, err := loadBuiltinRelease(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	m := &consumerMemory{items: map[string]json.RawMessage{}, provenance: map[string]PackProvenance{}}
	members := []PackSetMember{}
	for _, v := range versions {
		members = append(members, memberFor(v.Content))
		if v.Content.Manifest.Key == "type_registry.host" {
			m.hostItem = v.Rows[0].Canonical
		}
		if v.Content.Manifest.Key == "type_registry.indicator" {
			for _, r := range v.Rows {
				m.items[r.ID] = r.Canonical
			}
		}
	}
	m.set, err = packformat.BuildSet(members)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	for _, v := range versions {
		m.provenance[v.Content.Manifest.Key] = provenanceFor(m.set.ID, v.Content.Manifest, successfulEnvelope{ManifestSHA256: v.Content.ManifestSHA256, PayloadSHA256: v.Content.PayloadSHA256, VerifiedAt: now})
	}
	c := newPackConsumer(m, pagination.NewCodec([]byte(strings.Repeat("x", 32))), func() time.Time { return now })
	return m, c, &now
}

func TestCanonicalBaseReleaseBindingsAndConsumerEvaluation_Unit(t *testing.T) {
	t.Run("machine contracts", testConsumerMachineContracts)
	m, c, _ := consumerFixture(t)
	if len(m.set.Members) != 3 || len(m.items) != 9 {
		t.Fatal("incomplete release-bound Base vocabulary")
	}
	for typ, raw := range map[string]string{"domain_name": "EXAMPLE.COM", "ipv4_addr": "192.0.2.1", "ipv6_addr": "2001:db8::1", "url": "https://EXAMPLE.COM/a/%2E%2E/b", "email_addr": "User@EXAMPLE.COM", "sha256": strings.Repeat("A", 64), "process_name": "Example.EXE", "registry_key": `HKEY_LOCAL_MACHINE\Software\Example`, "text": "Example"} {
		t.Run(typ, func(t *testing.T) {
			row, ok := m.items[typ]
			if !ok {
				t.Fatalf("missing built-in type %s", typ)
			}
			var policy struct {
				Kinds []string `json:"allowed_value_kinds"`
			}
			if err := json.Unmarshal(row, &policy); err != nil {
				t.Fatal(err)
			}
			result := c.EvaluateIndicatorValue(context.Background(), EvaluateIndicatorRequest{PackSetID: m.set.ID, IndicatorTypeID: typ, ValueKind: policy.Kinds[0], RawValue: raw})
			if result.Error != nil || result.Value == nil || !result.Value.Valid {
				t.Fatalf("evaluation: %#v", result)
			}
		})
	}
	unknown := c.EvaluateIndicatorValue(context.Background(), EvaluateIndicatorRequest{PackSetID: m.set.ID, IndicatorTypeID: "unrecognized", ValueKind: "text", RawValue: "value"})
	if unknown.Error == nil || unknown.Error.Code != "indicator_type_unsupported" || unknown.Value != nil {
		t.Fatal("unknown type fabricated evaluation")
	}
	for _, length := range []int{512, 513} {
		request := GetPackEntryRequest{PackSetID: "rpset_" + strings.Repeat("0", 64), PackKey: "type_registry.host", EntryID: strings.Repeat("x", length)}
		result := c.GetPackEntry(context.Background(), request)
		expected := "pack_set_not_found"
		if length > 512 {
			expected = "invalid_pack_request"
		}
		if result.Error == nil || result.Error.Code != expected {
			t.Fatal("generic entry bound precedence", length, result.Error)
		}
	}
	request := GetPackEntryRequest{PackSetID: "rpset_" + strings.Repeat("0", 64), PackKey: "type_registry.host", EntryID: strings.Repeat("x", 513)}
	result := c.GetPackEntry(context.Background(), request)
	if result.Error == nil || result.Error.Code != "invalid_pack_request" {
		t.Fatal("request bounds did not precede set resolution")
	}
}

func testConsumerEnvelopePrecedesSetResolution(t *testing.T) {
	t.Helper()
	_, c, _ := consumerFixture(t)
	unknown := "rpset_" + strings.Repeat("0", 64)
	malformed := string([]byte{255})
	for _, value := range []any{true, []any{}, json.Number("1e0"), json.Number("65536"), json.Number("-1"), map[string]any{"provider_name": "p", "event_id": json.Number("1"), "event_version": nil}, map[string]any{"provider_name": "p", "event_id": json.Number("1"), "hostile": true}, map[string]any{"algorithm": "sha256", "value": "x", "hostile": true}, map[string]any{"algorithm": "sha256", "value": strings.Repeat("x", 8193)}} {
		result := c.LookupPackEntries(context.Background(), LookupPackEntriesRequest{PackSetID: unknown, PackKey: "enrichment.windows_event_ids", LookupKind: "provider_event_id", LookupValue: value})
		if result.Error == nil || result.Error.Code != "invalid_pack_request" || result.Value != nil {
			t.Fatal("malformed lookup reached set resolution", result.Error)
		}
	}
	for _, value := range []any{json.Number("65535"), map[string]any{"provider_name": "p", "event_id": json.Number("1")}, map[string]any{"algorithm": "sha256", "value": "not-a-hash"}} {
		result := c.LookupPackEntries(context.Background(), LookupPackEntriesRequest{PackSetID: unknown, PackKey: "enrichment.windows_event_ids", LookupKind: "provider_event_id", LookupValue: value})
		if result.Error == nil || result.Error.Code != "pack_set_not_found" {
			t.Fatal("profile semantics preceded set resolution", result.Error)
		}
	}
	for _, field := range []string{"type", "kind"} {
		request := EvaluateIndicatorRequest{PackSetID: unknown, IndicatorTypeID: "text", ValueKind: "atomic", RawValue: strings.Repeat("x", 8193)}
		if field == "type" {
			request.IndicatorTypeID = malformed
		} else {
			request.ValueKind = malformed
		}
		result := c.EvaluateIndicatorValue(context.Background(), request)
		if result.Error == nil || result.Error.Code != "invalid_pack_request" || result.Error.ReasonCode != nil {
			t.Fatal("scalar field order drift", result.Error)
		}
	}
}

func TestCanonicalConsumerCursorBindingOmissionAndExactExpiry_Unit(t *testing.T) {
	testConsumerEnvelopePrecedesSetResolution(t)
	for _, input := range []string{
		`{"pack_set_id":"set","pack_key":"key","lookup_kind":"alias","lookup_value":"unknown","limit":null}`,
		`{"pack_set_id":"set","pack_key":"key","lookup_kind":"alias","lookup_value":"unknown","cursor":null}`,
		`{"pack_set_id":"set","pack_key":"key","lookup_kind":"alias","lookup_value":"unknown","limit":1.5}`,
		`{"pack_set_id":"set","pack_key":"key","lookup_kind":"alias","lookup_value":"unknown","limit":1,"limit":2}`,
		`{"pack_set_id":"set","pack_key":"key","lookup_kind":"alias","lookup_value":"unknown","hostile":true}`,
		`{"pack_set_id":"set","pack_key":"key","lookup_kind":"alias","lookup_value":"\ud800"}`,
		`{"pack_set_id":"set","pack_key":"key","lookup_kind":"alias"}`,
	} {
		request := LookupPackEntriesRequest{PackSetID: "unchanged"}
		if err := json.Unmarshal([]byte(input), &request); err == nil || err.Error() != "reference pack consumer: invalid_pack_request" || request.PackSetID != "unchanged" {
			t.Fatal("request decoder admitted invalid bytes, leaked details, or partially mutated input")
		}
	}
	var decoded LookupPackEntriesRequest
	if err := json.Unmarshal([]byte(`{"pack_set_id":"rpset_`+strings.Repeat("0", 64)+`","pack_key":"enrichment.windows_event_ids","lookup_kind":"provider_event_id","lookup_value":{"provider_name":"provider","event_id":1}}`), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Limit != nil || decoded.Cursor != nil {
		t.Fatal("omission did not survive decoding")
	}
	if _, ok := decoded.LookupValue.(map[string]any)["event_id"].(json.Number); !ok {
		t.Fatal("lookup numeric representation lost")
	}
	for _, target := range []any{&GetPackEntryRequest{}, &GetPackProvenanceRequest{}, &EvaluateIndicatorRequest{}} {
		if err := json.Unmarshal([]byte(`{}`), target); err == nil {
			t.Fatal("omitted required request members accepted")
		}
	}
	m, c, now := consumerFixture(t)
	for _, limit := range []int{0, 1, 200, 201} {
		result := c.LookupPackEntries(context.Background(), LookupPackEntriesRequest{PackSetID: m.set.ID, PackKey: "type_registry.host", LookupKind: "alias", LookupValue: "unknown", Limit: &limit})
		if limit == 0 || limit == 201 {
			if result.Error == nil || result.Error.Code != "invalid_pack_request" {
				t.Fatal("lookup range rejection", limit, result.Error)
			}
		} else if result.Error != nil || len(result.Value.Items) != limit {
			t.Fatal("lookup range equality", limit, result.Error)
		}
	}
	defaultPage := c.LookupPackEntries(context.Background(), LookupPackEntriesRequest{PackSetID: m.set.ID, PackKey: "type_registry.host", LookupKind: "alias", LookupValue: "unknown"})
	if defaultPage.Error != nil || len(defaultPage.Value.Items) != 50 {
		t.Fatal("lookup default", defaultPage.Error)
	}
	*now = now.Add(750 * time.Millisecond)
	issued := *now
	limit := 2
	request := LookupPackEntriesRequest{PackSetID: m.set.ID, PackKey: "type_registry.host", LookupKind: "alias", LookupValue: "unknown", Limit: &limit}
	first := c.LookupPackEntries(context.Background(), request)
	if first.Error != nil || first.Value == nil || first.Value.NextCursor == nil || len(first.Value.Items) != 2 {
		t.Fatalf("first page: %#v", first)
	}
	request.Cursor = first.Value.NextCursor
	request.Limit = nil
	next := c.LookupPackEntries(context.Background(), request)
	if next.Error != nil || m.limit != 3 {
		t.Fatal("continuation did not inherit cursor limit")
	}
	wrong := 3
	request.Limit = &wrong
	if r := c.LookupPackEntries(context.Background(), request); r.Error == nil || r.Error.Code != "cursor_query_mismatch" {
		t.Fatal("cursor accepted limit change")
	}
	request.Limit = nil
	request.LookupValue = "different"
	if r := c.LookupPackEntries(context.Background(), request); r.Error == nil || r.Error.Code != "cursor_query_mismatch" {
		t.Fatal("cursor accepted query change")
	}
	request.LookupValue = "unknown"
	*now = issued.Add(900*time.Second - time.Nanosecond)
	if r := c.LookupPackEntries(context.Background(), request); r.Error != nil {
		t.Fatal("cursor expired before its exact 900-second lifetime", r.Error)
	}
	*now = issued.Add(900 * time.Second)
	if r := c.LookupPackEntries(context.Background(), request); r.Error == nil || r.Error.Code != "cursor_invalid" {
		t.Fatal("cursor valid at expiry equality")
	}
}
