package jobs_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
)

func checkCurrentPublicResources(t *testing.T) {
	const id = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	const incident = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	const table = "nft_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const graph = "nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const pack = "/api/v1/reference-packs/example.pack/2026.11+build"
	valid := []jobs.ResourceRef{
		{Kind: "incident", ID: id, Route: "/api/v1/incidents/" + id},
		{Kind: "import_session", ID: id, Route: "/api/v1/import-sessions/" + id},
		{Kind: "snapshot", ID: id, Route: "/api/v1/snapshots/" + id},
		{Kind: "release", ID: id, Route: "/api/v1/releases/" + id},
		{Kind: "incident_bundle", ID: id, Route: "/api/v1/incident-bundles/" + id},
		{Kind: "network_flow_table", ID: table, Route: "/api/v1/incidents/" + incident + "/network-flow/tables/" + table},
		{Kind: "network_flow_graph_view", ID: graph, Route: "/api/v1/incidents/" + incident + "/network-flow/graph-views/" + graph},
		{Kind: "reference_pack_version", ID: pack, Route: pack},
	}
	definition := func(kind string) jobs.Definition {
		return jobs.Definition{JobKind: "test_profile.run_v1", ProgressUnitID: "test_profile.run.attempt.v1", HandlerName: "test_profile.worker_v1", Terminal: jobcontract.TerminalContract{SchemaID: jobcontract.TerminalSuccessSchema, SuccessCodes: []string{"done"}, Resources: []jobcontract.ResourceContract{{Kind: kind, SchemaID: jobcontract.ResourceIDSchema, MinRefs: 1, MaxRefs: 1}}}, Extension: &jobs.ExtensionPolicy{
			OwnerProfileID: "test_profile", OperationKind: "test_profile.run", ContractSHA256: strings.Repeat("a", 64), ProofRequired: true, MaxProofBytes: 4096,
			IdempotencyPolicy: "none", CancellationPolicy: "not_cancelable",
		}}
	}
	check := func(t *testing.T, ref jobs.ResourceRef, accepted bool) {
		t.Helper()
		_, _, _, _, err := jobs.CanonicalTerminalSuccess(definition(ref.Kind), &jobs.ResultSummary{Code: "done", Message: "Done.", ResourceRefs: []jobs.ResourceRef{ref}})
		if (err == nil) != accepted {
			t.Fatalf("resource %+v accepted=%t, want %t: %v", ref, err == nil, accepted, err)
		}
	}
	for _, ref := range valid {
		t.Run(ref.Kind, func(t *testing.T) { check(t, ref, true) })
		t.Run(ref.Kind+" missing route", func(t *testing.T) { ref.Route = ""; check(t, ref, false) })
	}
	invalid := []jobs.ResourceRef{
		{Kind: "future_kind", ID: id, Route: "/api/v1/incidents/" + id},
		{Kind: "incident", ID: "arbitrary", Route: "/api/v1/incidents/arbitrary"},
		{Kind: "incident", ID: strings.ToUpper(id), Route: "/api/v1/incidents/" + strings.ToUpper(id)},
		{Kind: "incident", ID: "00000000-0000-0000-0000-000000000000", Route: "/api/v1/incidents/00000000-0000-0000-0000-000000000000"},
		{Kind: "incident", ID: id, Route: "/api/v1/incidents/" + incident},
		{Kind: "incident", ID: id, Route: "/api/v1/releases/" + id},
		{Kind: "network_flow_table", ID: graph, Route: "/api/v1/incidents/" + incident + "/network-flow/tables/" + graph},
		{Kind: "network_flow_table", ID: table, Route: "/api/v1/incidents/not-an-id/network-flow/tables/" + table},
		{Kind: "reference_pack_version", ID: pack, Route: pack + ".other"},
		{Kind: "reference_pack_version", ID: "/api/v1/reference-packs/BAD/one", Route: "/api/v1/reference-packs/BAD/one"},
		{Kind: "reference_pack_version", ID: pack, Route: pack + strings.Repeat("/", 8192)},
		{Kind: "network_flow_graph_view", ID: graph, Route: "/api/v1/incidents/" + incident + "/network-flow/graph-views/" + graph + strings.Repeat("/", 8192)},
	}
	for _, suffix := range []string{"?download=true", "#fragment", "/", "/..", "%2f", "\x00"} {
		ref := valid[0]
		ref.Route += suffix
		invalid = append(invalid, ref)
	}
	for i, ref := range invalid {
		t.Run(fmt.Sprintf("reject %d %s", i, ref.Kind), func(t *testing.T) { check(t, ref, false) })
	}
	if _, err := jobs.NewCatalog([]jobs.Definition{definition("future_kind")}); err == nil {
		t.Error("catalog accepted a resource kind outside the current Core vocabulary")
	}
	decode, _ := jobcontract.ResolveTerminal(jobcontract.TerminalSuccessSchema)
	optional := valid[0]
	optional.Route = ""
	raw, _ := json.Marshal(jobs.ResultSummary{Code: "done", Message: "Done.", ResourceRefs: []jobs.ResourceRef{optional}})
	if _, err := decode(raw); err != nil {
		t.Fatal("abstract common schema lost its optional route", err)
	}
	if _, err := decode([]byte(`{"code":"done","message":"Done.","resource_refs":[{"kind":"incident","id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","route":""}]}`)); err == nil {
		t.Fatal("abstract schema accepted an explicitly empty route")
	}
}
