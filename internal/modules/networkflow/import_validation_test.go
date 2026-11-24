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
	reader                 func(string) io.Reader
	opens, closes, maxRead int
}

func (s *applySourceFixture) OpenSourceStream(context.Context, string) (imports.ImportSourceStream, error) {
	return imports.ImportSourceStream{}, errors.New("apply attempted an owned source connection")
}

func (s *applySourceFixture) OpenSourceStreamTx(_ context.Context, _ pgx.Tx, _ string) (imports.ImportSourceStream, error) {
	s.opens++
	var reader io.Reader = strings.NewReader(s.source)
	if s.reader != nil {
		reader = s.reader(s.source)
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

// validateImportFixture opens the fixture source and runs apply's single pass
// with an already decoded mapping.
func validateImportFixture(f *importFacade, request imports.ExtensionImportApplyRequest, mapping approvedMapping, visit func(normalizedRow) error) (importValidationSummary, error) {
	stream, err := f.openImportSource(context.Background(), nil, request)
	if err != nil {
		return importValidationSummary{}, err
	}
	defer stream.Reader.Close()
	return f.validateImportSource(context.Background(), stream.Reader, request, mapping, nil, visit)
}

// applyPass prepares an apply and runs its single pass, as CreateTableTx
// does, handing each accepted row to visit.
func applyPass(ctx context.Context, f *importFacade, request imports.ExtensionImportApplyRequest, visit func(normalizedRow) error) (importSummary, error) {
	prepared, err := f.prepareImportApply(ctx, nil, request)
	if err != nil {
		return importSummary{}, err
	}
	defer prepared.close()
	if visit == nil {
		visit = func(normalizedRow) error { return nil }
	}
	return prepared.params.ImportRows(nil, visit)
}

func TestStreamingImportApply(t *testing.T) {
	invalid := strings.Replace(validPreviewRow, "192.0.2.10", "invalid-ip", 1)
	t.Run("rejected rows are independent of diagnostic multiplicity", func(t *testing.T) {
		f, request, _ := streamingApplyFixture(t, validPreviewRow+strings.Repeat("x,x,x,x,x,x,x,x,x\n", 2), defaultLimits())
		summary, err := applyPass(context.Background(), f, request, nil)
		if err != nil || summary.RowCountAccepted != 1 || summary.RowCountRejected != 2 || len(summary.Diagnostics) != 18 {
			t.Fatalf("row and diagnostic counts conflated: %v", err)
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
			visited := 0
			summary, err := applyPass(context.Background(), f, request, func(row normalizedRow) error {
				visited++
				if row.SourceRowNumber != 4 {
					t.Fatal("accepted source locator changed")
				}
				return nil
			})
			if err != nil || visited != 1 {
				t.Fatalf("streaming visitor: count=%d err=%v", visited, err)
			}
			if summary.RowCountAccepted != 1 || summary.RowCountRejected != 2 || len(summary.Diagnostics) != min(int(limit), 3) || summary.DiagnosticsTruncated != (limit < 3) {
				t.Fatalf("streaming retention at %d: %#v", limit, summary)
			}
			expected := preview.OwnerResult["diagnostics"].([]map[string]any)
			for index, diagnostic := range summary.Diagnostics {
				if !canonicalEqual(diagnosticResource(diagnostic), expected[index]) {
					t.Fatal("preview/apply diagnostics or prefix order differ")
				}
			}

			_, rejectedPreview := preparePreviewFixture(t, rejected, limits)
			f, request, _ = streamingApplyFixture(t, rejected, limits)
			request.OwnerMapping, request.MappingFingerprint = rejectedPreview.OwnerMapping, rejectedPreview.MappingFingerprint
			_, err = applyPass(context.Background(), f, request, func(normalizedRow) error {
				t.Fatal("rejected record reached the accepted-row visitor")
				return nil
			})
			var owner *networkFlowImportOwnerError
			if !errors.As(err, &owner) || owner.ownerCode != "network_flow_all_rows_rejected" {
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
		_, err := applyPass(context.Background(), f, request, nil)
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
			count := 0
			summary, err := applyPass(context.Background(), f, request, func(row normalizedRow) error {
				if row.SourceRowNumber != int64(count*2+2) {
					t.Fatal("insertion lost source order")
				}
				count++
				return nil
			})
			if err != nil || count != size || summary.RowCountAccepted != int64(size) || summary.RowCountRejected != int64(size) || len(summary.Diagnostics) != 3 || !summary.DiagnosticsTruncated {
				t.Fatalf("single pass did not retain the bounded summary: count=%d err=%v", count, err)
			}
			if source.opens != 1 || source.closes != 1 || source.maxRead > 65536 {
				t.Fatalf("apply read the source other than once: opens=%d closes=%d max_read=%d", source.opens, source.closes, source.maxRead)
			}
		})
	}
	t.Run("diagnostic retention 0 N Nplus1", func(t *testing.T) {
		for _, limit := range []int64{0, 3} {
			for _, rejected := range []int{0, 3, 4} {
				limits := defaultLimits()
				limits.MaxRejectedRowDiagnostics = limit
				f, request, _ := streamingApplyFixture(t, validPreviewRow+strings.Repeat(invalid, rejected), limits)
				summary, err := applyPass(context.Background(), f, request, nil)
				if err != nil || len(summary.Diagnostics) != min(rejected, int(limit)) || summary.DiagnosticsTruncated != (int64(rejected) > limit) {
					t.Fatalf("retention limit=%d rejected=%d err=%v", limit, rejected, err)
				}
			}
		}
	})
	t.Run("accepted limit and all-rejected admission", func(t *testing.T) {
		limits := defaultLimits()
		limits.MaxAcceptedRowsPerTable = 1
		f, request, source := streamingApplyFixture(t, strings.Repeat(validPreviewRow, 2), limits)
		_, err := applyPass(context.Background(), f, request, nil)
		var owner *networkFlowImportOwnerError
		if !errors.As(err, &owner) || owner.safeDetails["reason_code"] != "network_flow_resource_limit_exceeded" || source.opens != 1 || source.closes != 1 {
			t.Fatalf("accepted-row limit was not enforced during the pass: %v", err)
		}
		f, request, _ = streamingApplyFixture(t, strings.Repeat(invalid, 1000), limits)
		_, err = applyPass(context.Background(), f, request, nil)
		if !errors.As(err, &owner) || owner.ownerCode != "network_flow_all_rows_rejected" || len(owner.safeDetails["diagnostics_sample"].([]map[string]any)) != 50 || owner.safeDetails["diagnostics_truncated"] != true {
			t.Fatalf("all-rejected summary was not bounded: %v", err)
		}
	})
	for _, failure := range []string{"read", "digest", "insert", "cancel"} {
		t.Run("pass "+failure, func(t *testing.T) {
			f, request, source := streamingApplyFixture(t, strings.Repeat(validPreviewRow, 100), defaultLimits())
			injected := errors.New("injected source or sink failure")
			if failure == "read" {
				source.reader = func(s string) io.Reader { return io.MultiReader(strings.NewReader(s), iotest.ErrReader(injected)) }
			}
			if failure == "digest" {
				source.reader = func(s string) io.Reader { return strings.NewReader(strings.Replace(s, "192.0.2.10", "192.0.2.11", 1)) }
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			inserted := 0
			_, err := applyPass(ctx, f, request, func(normalizedRow) error {
				inserted++
				if failure == "insert" {
					return injected
				}
				if failure == "cancel" {
					cancel()
				}
				return nil
			})
			if err == nil || source.opens != 1 || source.closes != 1 {
				t.Fatalf("pass failure accepted: %v", err)
			}
			if inserted == 0 {
				t.Fatal("fixture did not exercise a failure after insertion began")
			}
		})
	}
}
