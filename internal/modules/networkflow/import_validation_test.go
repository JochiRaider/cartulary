package networkflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/jackc/pgx/v5"
)

type applySourceFixture struct {
	ImportSourcePort
	capability             imports.ImportSourceCapability
	source                 string
	second                 func(string) io.Reader
	opens, closes, maxRead int
}

func (s *applySourceFixture) OpenSourceStream(context.Context, string) (imports.ImportSourceStream, error) {
	return imports.ImportSourceStream{}, errors.New("apply attempted an owned source connection")
}

func (s *applySourceFixture) OpenSourceStreamTx(_ context.Context, _ pgx.Tx, _ string) (imports.ImportSourceStream, error) {
	s.opens++
	var reader io.Reader = strings.NewReader(s.source)
	if s.opens == 2 && s.second != nil {
		reader = s.second(s.source)
	}
	return imports.ImportSourceStream{ImportSourceCapability: s.capability, OriginalFilename: "source.csv", Reader: &observedApplyReader{Reader: reader, source: s}}, nil
}

type observedApplyReader struct {
	io.Reader
	source *applySourceFixture
}

func (r *observedApplyReader) Read(dst []byte) (int, error) {
	r.source.maxRead = max(r.source.maxRead, len(dst))
	return r.Reader.Read(dst)
}
func (r *observedApplyReader) Close() error { r.source.closes++; return nil }

func streamingApplyFixture(t *testing.T, data string, limits EffectiveLimits) (*importFacade, imports.ExtensionImportApplyRequest, *applySourceFixture) {
	t.Helper()
	return streamingApplySourceFixture(t, strings.Join(requiredFields(sourceProfileCiscoSNANetFlowCSV), ",")+"\n"+data, limits)
}

func streamingApplySourceFixture(t *testing.T, source string, limits EffectiveLimits) (*importFacade, imports.ExtensionImportApplyRequest, *applySourceFixture) {
	t.Helper()
	sha := sha256Hex([]byte(source))
	parsed, err := parseCSVPreview(strings.NewReader(source), sha, limits)
	if err != nil {
		t.Fatal(err)
	}
	mapping := approvedMappingFixture(sourceProfileCiscoSNANetFlowCSV)
	mapping.SourceColumns = parsed.SourceColumns
	fixture := &applySourceFixture{source: source, capability: imports.ImportSourceCapability{SourceStreamRef: "impsrc_" + strings.Repeat("1", 32), SourceContentSHA256: sha, SourceMediaType: "text/csv", SourceByteSize: int64(len(source))}}
	facade := &importFacade{sourceStore: fixture, limits: limits, safeDigester: &queryCompletionProbe{}, now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }}
	request := imports.ExtensionImportApplyRequest{SourceCapability: fixture.capability, ExpectedSourceContentSHA256: sha, MappingFingerprint: mustMappingFingerprint(mapping, sha), OwnerMapping: marshalApprovedMapping(mapping)}
	return facade, request, fixture
}

func TestStreamingImportApply(t *testing.T) {
	invalid := strings.Replace(validPreviewRow, "192.0.2.10", "invalid-ip", 1)
	t.Run("rejected rows are independent of diagnostic multiplicity", func(t *testing.T) {
		f, request, _ := streamingApplyFixture(t, validPreviewRow+strings.Repeat("x,x,x,x,x,x,x,x,x\n", 2), defaultLimits())
		prepared, err := f.prepareImportApply(context.Background(), nil, request)
		if err != nil || prepared.params.RowCountAccepted != 1 || prepared.params.RowCountRejected != 2 || len(prepared.params.Diagnostics) != 18 {
			t.Fatalf("row and diagnostic counts conflated: %v", err)
		}
		if err := prepared.params.VisitRows(func(normalizedRow) error { return nil }); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("preview agreement and all-rejected sampling ignore retention", func(t *testing.T) {
		invalid := strings.ReplaceAll(strings.ReplaceAll(validPreviewRow, "443", "99999"), "192.0.2.10", "secret-host")
		rejected := invalid + "short,row\n"
		for _, limit := range []int64{0, 2, 3, 4} {
			limits := defaultLimits()
			limits.MaxRejectedRowDiagnostics = limit
			_, preview := preparePreviewFixture(t, rejected+validPreviewRow, limits)
			f, request, _ := streamingApplyFixture(t, rejected+validPreviewRow, limits)
			request.OwnerMapping, request.MappingFingerprint = preview.OwnerMapping, preview.MappingFingerprint
			prepared, err := f.prepareImportApply(context.Background(), nil, request)
			if err != nil {
				t.Fatal(err)
			}
			if prepared.params.RowCountAccepted != 1 || prepared.params.RowCountRejected != 2 || len(prepared.params.Diagnostics) != min(int(limit), 3) || prepared.params.DiagnosticsTruncated != (limit < 3) {
				t.Fatalf("streaming retention at %d: %#v", limit, prepared.params)
			}
			expected := preview.OwnerResult["diagnostics"].([]map[string]any)
			for index, diagnostic := range prepared.params.Diagnostics {
				if !canonicalEqual(diagnosticResource(diagnostic), expected[index]) {
					t.Fatal("preview/apply diagnostics or prefix order differ")
				}
			}
			visited := 0
			if err := prepared.params.VisitRows(func(row normalizedRow) error {
				visited++
				if row.SourceRowNumber != 4 {
					t.Fatal("accepted source locator changed")
				}
				return nil
			}); err != nil || visited != 1 {
				t.Fatalf("streaming visitor: count=%d err=%v", visited, err)
			}

			_, rejectedPreview := preparePreviewFixture(t, rejected, limits)
			f, request, _ = streamingApplyFixture(t, rejected, limits)
			request.OwnerMapping, request.MappingFingerprint = rejectedPreview.OwnerMapping, rejectedPreview.MappingFingerprint
			prepared, err = f.prepareImportApply(context.Background(), nil, request)
			var owner *networkFlowImportOwnerError
			if !errors.As(err, &owner) || owner.ownerCode != "network_flow_all_rows_rejected" || prepared.params.VisitRows != nil {
				t.Fatalf("all-rejected source admitted: %v", err)
			}
			if owner.safeDetails["row_count_rejected"] != int64(2) || owner.safeDetails["diagnostics_truncated"] != false || !canonicalEqual(owner.safeDetails["diagnostics_sample"], rejectedPreview.OwnerResult["diagnostics"]) {
				t.Fatalf("all-rejected accounting/sample depends on retention %d: %#v", limit, owner.safeDetails)
			}
		}
	})
	t.Run("valid zero instant participates in cross-field validation", func(t *testing.T) {
		data := strings.Replace(validPreviewRow, "2026-07-13T12:01:00Z", "0001-01-01T00:00:00Z", 1)
		f, request, _ := streamingApplyFixture(t, data, defaultLimits())
		_, err := f.prepareImportApply(context.Background(), nil, request)
		var owner *networkFlowImportOwnerError
		if !errors.As(err, &owner) || owner.ownerCode != "network_flow_all_rows_rejected" {
			t.Fatalf("end before start accepted at valid zero instant: %v", err)
		}
		sample := owner.safeDetails["diagnostics_sample"].([]map[string]any)
		if len(sample) != 1 || sample[0]["error_code"] != "network_flow_end_before_start" {
			t.Fatal("valid timestamp did not participate in cross-field validation")
		}
	})
	for _, size := range []int{10, 1000, 10000} {
		t.Run(fmt.Sprintf("bounded summary rows_%d", size), func(t *testing.T) {
			limits := defaultLimits()
			limits.MaxRejectedRowDiagnostics = 3
			f, request, source := streamingApplyFixture(t, strings.Repeat(validPreviewRow+invalid, size), limits)
			prepared, err := f.prepareImportApply(context.Background(), nil, request)
			if err != nil {
				t.Fatal(err)
			}
			if prepared.params.RowCountAccepted != int64(size) || prepared.params.RowCountRejected != int64(size) || len(prepared.params.Diagnostics) != 3 || !prepared.params.DiagnosticsTruncated || source.opens != 1 {
				t.Fatal("first pass did not retain the bounded validation summary")
			}
			count := 0
			err = prepared.params.VisitRows(func(row normalizedRow) error {
				if row.SourceRowNumber != int64(count*2+2) {
					t.Fatal("insertion lost source order")
				}
				count++
				return nil
			})
			if err != nil || count != size || source.opens != 2 || source.closes != 2 || source.maxRead > 65536 {
				t.Fatalf("streamed passes disagree: count=%d opens=%d closes=%d max_read=%d err=%v", count, source.opens, source.closes, source.maxRead, err)
			}
		})
	}
	t.Run("diagnostic retention 0 N Nplus1", func(t *testing.T) {
		for _, limit := range []int64{0, 3} {
			for _, rejected := range []int{0, 3, 4} {
				limits := defaultLimits()
				limits.MaxRejectedRowDiagnostics = limit
				f, request, _ := streamingApplyFixture(t, validPreviewRow+strings.Repeat(invalid, rejected), limits)
				p, err := f.prepareImportApply(context.Background(), nil, request)
				if err != nil || len(p.params.Diagnostics) != min(rejected, int(limit)) || p.params.DiagnosticsTruncated != (int64(rejected) > limit) {
					t.Fatalf("retention limit=%d rejected=%d err=%v", limit, rejected, err)
				}
			}
		}
	})
	t.Run("accepted limit and all-rejected admission", func(t *testing.T) {
		limits := defaultLimits()
		limits.MaxAcceptedRowsPerTable = 1
		f, request, source := streamingApplyFixture(t, strings.Repeat(validPreviewRow, 2), limits)
		p, err := f.prepareImportApply(context.Background(), nil, request)
		if err == nil || p.params.VisitRows != nil || source.opens != 1 || source.closes != 1 {
			t.Fatal("accepted-row limit was enforced after allocation")
		}
		f, request, _ = streamingApplyFixture(t, strings.Repeat(invalid, 1000), limits)
		p, err = f.prepareImportApply(context.Background(), nil, request)
		var owner *networkFlowImportOwnerError
		if !errors.As(err, &owner) || owner.ownerCode != "network_flow_all_rows_rejected" || p.params.VisitRows != nil || len(owner.safeDetails["diagnostics_sample"].([]map[string]any)) != 50 || owner.safeDetails["diagnostics_truncated"] != true {
			t.Fatalf("all-rejected summary was not bounded: %v", err)
		}
	})
	for _, failure := range []string{"read", "digest", "insert", "cancel", "summary"} {
		t.Run("second pass "+failure, func(t *testing.T) {
			f, request, source := streamingApplyFixture(t, strings.Repeat(validPreviewRow, 100), defaultLimits())
			injected := errors.New("injected source or sink failure")
			if failure == "read" {
				source.second = func(s string) io.Reader { return io.MultiReader(strings.NewReader(s), iotest.ErrReader(injected)) }
			}
			if failure == "digest" {
				source.second = func(s string) io.Reader { return strings.NewReader(strings.Replace(s, "192.0.2.10", "192.0.2.11", 1)) }
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			prepared, err := f.prepareImportApply(ctx, nil, request)
			if err != nil {
				t.Fatal(err)
			}
			if failure == "summary" {
				f.limits.MaxRawCellScalarLength = 1
			}
			inserted := 0
			err = prepared.params.VisitRows(func(normalizedRow) error {
				inserted++
				if failure == "insert" {
					return injected
				}
				if failure == "cancel" {
					cancel()
				}
				return nil
			})
			if err == nil || source.opens != 2 || source.closes != 2 {
				t.Fatalf("second-pass failure accepted: %v", err)
			}
			if failure != "summary" && inserted == 0 {
				t.Fatal("fixture did not exercise a failure after insertion began")
			}
		})
	}
}
