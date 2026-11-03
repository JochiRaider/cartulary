package jobs_test

import (
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
	"testing"
)

func checkTerminalCanonicalBytes(t *testing.T) {
	definition := jobs.Definition{Terminal: jobcontract.TerminalContract{SchemaID: jobcontract.TerminalSuccessSchema, SuccessCodes: []string{"done"}, Resources: []jobcontract.ResourceContract{{Kind: "incident", SchemaID: jobcontract.ResourceIDSchema, MaxRefs: 1}}}}
	reference := jobs.ResourceRef{Kind: "incident", ID: "11111111-1111-4111-8111-111111111111", Route: "/api/v1/incidents/11111111-1111-4111-8111-111111111111"}
	for _, tc := range []struct {
		name, message, terminal, refs, digest string
		resources                             []jobs.ResourceRef
	}{
		{name: "UTF8 and HTML without Go escaping", message: "café <&>\n\u2028", terminal: `{"code":"done","message":"café <&>\n` + "\u2028" + `","resource_refs":[{"id":"11111111-1111-4111-8111-111111111111","kind":"incident","route":"/api/v1/incidents/11111111-1111-4111-8111-111111111111"}]}`, refs: `[{"id":"11111111-1111-4111-8111-111111111111","kind":"incident","route":"/api/v1/incidents/11111111-1111-4111-8111-111111111111"}]`, digest: "e5b48117a6d359ca123857df9abb68247907bf5270c5e2f814ed9baab263d06d", resources: []jobs.ResourceRef{reference}},
		{name: "controls astral and preserved decomposition", message: "\x01\b\t\n\f\r\"\\😀e\u0301", terminal: `{"code":"done","message":"\u0001\b\t\n\f\r\"\\😀é"}`, refs: `[]`, digest: "45f43b1e92a9f9db40cd5494a6b2e8103e4375cb33876c1f832d0fad6c010405"},
		{name: "empty reference omission and array", message: "Done.", terminal: `{"code":"done","message":"Done."}`, refs: `[]`, digest: "00152aec3eab5a420636515f594a3560311c609cc67d9592b26a80fdfbc48231"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, terminal, refs, digest, err := jobs.CanonicalTerminalSuccess(definition, &jobs.ResultSummary{Code: "done", Message: tc.message, ResourceRefs: tc.resources})
			if err != nil {
				t.Fatal(err)
			}
			if string(terminal) != tc.terminal || string(refs) != tc.refs || digest != tc.digest {
				t.Fatalf("canonical byte vector mismatch: terminal=%q refs=%q digest=%s", terminal, refs, digest)
			}
		})
	}
	t.Run("current schema only", func(t *testing.T) {
		if jobcontract.TerminalSuccessSchema != "cartulary.common_job_terminal_success.v2" {
			t.Error("current terminal schema does not bind the adopted encoding")
		}
		if _, ok := jobcontract.ResolveTerminal("cartulary.common_job_terminal_success.v1"); ok {
			t.Error("retired terminal schema resolves")
		}
	})
}
