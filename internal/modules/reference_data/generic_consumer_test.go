package reference_data

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type genericConsumerRepository struct {
	*consumerMemory
	queried PackSetMember
	query   packformat.LookupInput
}

func (r *genericConsumerRepository) Lookup(_ context.Context, member PackSetMember, query packformat.LookupInput, _ string, _ int) ([]indexedItem, error) {
	r.queried, r.query = member, query
	return []indexedItem{}, nil
}

func TestGenericDescriptorConsumer_Unit(t *testing.T) {
	ctx := context.Background()
	m, base, _ := consumerFixture(t)
	baseDescriptor := base.GetPackDescriptor(ctx, GetPackDescriptorRequest{PackSetID: m.set.ID, PackKey: "type_registry.host"})
	if baseDescriptor.Error != nil || baseDescriptor.Value.Descriptor != nil {
		t.Fatal("Base descriptor must be explicitly nullable", baseDescriptor)
	}
	if raw, err := json.Marshal(baseDescriptor); err != nil || packformat.ValidateConsumerResult("GetPackDescriptor", raw) != nil {
		t.Fatal("nullable descriptor projection", err)
	}
	key := "advisory.unregistered_fifth_dataset"
	m.descriptor = json.RawMessage(`{"lookups":[{"components":[{"name":"event","normalization":"safe_integer","type":"integer"},{"name":"provider","normalization":"ascii_case_insensitive","type":"string"}],"coverage":{"distinct_keys":0,"objects":0,"postings":0},"input_type":"composite","name":"provider_event","unique":false},{"components":[{"name":"value","normalization":"ascii_case_insensitive","type":"string"}],"coverage":{"distinct_keys":0,"objects":0,"postings":0},"input_type":"scalar","name":"service_principal_name","unique":true}],"schema_id":"cartulary.reference_pack.dataset.descriptor.v1"}`)
	member := m.set.Members[0]
	member.Key = key
	member.ProfileID = "cartulary.reference_pack.dataset.v1"
	member.ProfileVersion = "1"
	m.set.Members = append(m.set.Members, member)
	provenance := m.provenance["type_registry.host"]
	provenance.PackKey = key
	provenance.ContentProfileID = member.ProfileID
	provenance.ContentProfileVersion = "1"
	provenance.AuthorityClass = "advisory_enrichment"
	m.provenance[key] = provenance
	r := &genericConsumerRepository{consumerMemory: m}
	codec := pagination.NewCodec([32]byte{1, 2, 3})
	c := newPackConsumer(r, codec, func() time.Time { return time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC) })
	descriptor := c.GetPackDescriptor(ctx, GetPackDescriptorRequest{PackSetID: m.set.ID, PackKey: key})
	if descriptor.Error != nil || string(descriptor.Value.Descriptor) != string(m.descriptor) || descriptor.Value.Provenance.PackKey != key {
		t.Fatal("exact descriptor/provenance not returned", descriptor)
	}
	request := LookupPackEntriesRequest{PackSetID: m.set.ID, PackKey: key, LookupKind: "provider_event", LookupValue: map[string]any{"event": json.Number("42"), "provider": "PROVIDER"}}
	page := c.LookupPackEntries(ctx, request)
	if page.Error != nil || len(page.Value.Items) != 0 || r.query.Kind != "provider_event" || r.query.Value != `[42,"provider"]` || r.queried != member {
		t.Fatal("descriptor-defined exact-set lookup", page, r.query)
	}
	request.LookupKind = "service_principal_name"
	request.LookupValue = "API://MixedCase"
	page = c.LookupPackEntries(ctx, request)
	if page.Error != nil || len(page.Value.Items) != 0 || r.query.Value != `["api://mixedcase"]` {
		t.Fatal("supported zero-coverage lookup is an empty success", page)
	}
	for _, tc := range []struct {
		kind  string
		value any
		code  string
	}{
		{"absent", "x", "lookup_kind_unsupported"},
		{"provider_event", map[string]any{"event": json.Number("42")}, "invalid_pack_request"},
		{"provider_event", map[string]any{"event": json.Number("42"), "provider": "p", "version": json.Number("1")}, "invalid_pack_request"},
		{"service_principal_name", json.Number("42"), "invalid_pack_request"},
	} {
		request.LookupKind, request.LookupValue = tc.kind, tc.value
		got := c.LookupPackEntries(ctx, request)
		if got.Error == nil || got.Error.Code != tc.code {
			t.Fatal(tc, got)
		}
	}
	request.PackKey = "advisory.absent"
	if got := c.LookupPackEntries(ctx, request); got.Error == nil || got.Error.Code != "pack_unavailable" {
		t.Fatal("unavailable result lost", got)
	}
	var decoded GetPackDescriptorRequest
	if json.Unmarshal([]byte(`{"pack_set_id":"x","pack_set_id":"y","pack_key":"advisory.test"}`), &decoded) == nil {
		t.Fatal("duplicate request member admitted")
	}
}
