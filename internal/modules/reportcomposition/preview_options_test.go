package reportcomposition

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompositionPreviewOutputOptionsAdmission(t *testing.T) {
	const defaults = `{"pdf":true,"png":false,"pptx":false,"rendered_diagrams":true,"schema_id":"cartulary.reporting_render_request_options.v1","source_only":false,"svg":false}`
	const prefix = `{"client_txn_id":"preview-options","snapshot_id":"snapshot","derivation_version":"cartulary.reporting_derivation_profile.v2","template_id":"cartulary.report.default","template_version":"1","redaction_profile_id":"profile","redaction_profile_version":"1","redaction_profile_sha256":"`
	base := prefix + strings.Repeat("a", 64) + `","render_environment_profile_id":"local","output_kind":"slidev"`
	decode := func(options string) (PreviewRequest, bool) {
		t.Helper()
		raw := base
		if options != "" {
			raw += `,"output_options":` + options
		}
		request, err := DecodePreviewRequest(strings.NewReader(raw + `}`))
		return request, err == nil
	}
	omitted, ok := decode("")
	if !ok || string(omitted.OutputOptions) != defaults {
		t.Fatal("preview did not materialize exact Reporting defaults before admission")
	}
	for _, options := range []string{`{}`, defaults} {
		explicit, ok := decode(options)
		if !ok || !bytes.Equal(explicit.Normalized, omitted.Normalized) {
			t.Fatal("equivalent preview defaults changed the idempotency request")
		}
	}
	sourceOnly, ok := decode(`{"source_only":true}`)
	const sourceOnlyOptions = `{"pdf":false,"png":false,"pptx":false,"rendered_diagrams":false,"schema_id":"cartulary.reporting_render_request_options.v1","source_only":true,"svg":false}`
	if !ok || string(sourceOnly.OutputOptions) != sourceOnlyOptions {
		t.Fatal("source-only preview did not freeze disabled rendered outputs")
	}
	for _, retained := range []PreviewRequest{omitted, sourceOnly} {
		decoded, err := DecodeNormalizedPreviewRequest(retained.Normalized)
		if err != nil || !bytes.Equal(decoded.Normalized, retained.Normalized) {
			t.Fatal("private draft normalization did not round trip", err)
		}
		if _, err := DecodePreviewRequest(bytes.NewReader(retained.Normalized)); err == nil {
			t.Fatal("private null version weakened public admission")
		}
	}
	version, apiErr := DecodePreviewRequest(strings.NewReader(base + `,"source_kind":"version","composition_version":"v1"}`))
	if apiErr != nil {
		t.Fatal(apiErr)
	}
	if _, err := DecodeNormalizedPreviewRequest(version.Normalized); err != nil {
		t.Fatal("private version normalization", err)
	}
	for _, raw := range [][]byte{
		bytes.Replace(omitted.Normalized, []byte(`"composition_version":null,`), nil, 1),
		append([]byte(" "), omitted.Normalized...),
		bytes.Replace(omitted.Normalized, []byte(`"composition_version":null`), []byte(`"composition_version":null,"unadmitted":true`), 1),
	} {
		if _, err := DecodeNormalizedPreviewRequest(raw); err == nil {
			t.Fatal("invalid normalized preview repaired")
		}
	}
	for _, options := range []string{`null`, `[]`, `{"pdf":null}`, `{"future":true}`, `{"pdf":"true"}`, `{"pdf":true,"pdf":false}`, `{"schema_id":"retired"}`, `{"source_only":true,"rendered_diagrams":true}`, `{"png":true}`, `{"pptx":true}`} {
		if _, ok := decode(options); ok {
			t.Fatal("preview admitted malformed output options")
		}
	}
}
