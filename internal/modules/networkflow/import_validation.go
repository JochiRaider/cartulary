package networkflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/jackc/pgx/v5"
)

type importValidationSummary struct {
	Accepted, Rejected, DiagnosticCount int64
	ValidationSHA                       string
	Diagnostics, Sample                 []rejectedRowDiagnostic
}

// openImportSource opens the protected source on the borrowed connection,
// before any COPY holds that connection. The caller closes the stream.
func (f *importFacade) openImportSource(ctx context.Context, tx pgx.Tx, request imports.ExtensionImportApplyRequest) (imports.ImportSourceStream, error) {
	stream, err := f.sourceStore.OpenSourceStreamTx(ctx, tx, request.SourceCapability.SourceStreamRef)
	if err != nil {
		return imports.ImportSourceStream{}, err
	}
	if stream.ImportSourceCapability != request.SourceCapability {
		_ = stream.Reader.Close()
		return imports.ImportSourceStream{}, importOwnerError("network_flow_source_changed", nil)
	}
	return stream, nil
}

// validateImportSource is apply's single pass (§9.2): it validates every
// record, retains the bounded diagnostics and hands each accepted row to
// visit in source order. With a gate, the pass shares the borrowed connection
// with a pipelined COPY: it reads the source only while the gate holds the
// connection, reading ahead so each hand-off covers a larger source range.
func (f *importFacade) validateImportSource(ctx context.Context, source io.Reader, request imports.ExtensionImportApplyRequest, mapping approvedMapping, gate rowSourceGate, visit func(normalizedRow) error) (importValidationSummary, error) {
	var summary importValidationSummary
	reader := source
	if gate != nil {
		reader = gate.readAhead(source, pipelinedSourceReadAhead)
	}
	d, err := newCSVDecoder(reader, request.ExpectedSourceContentSHA256, f.limits, parseModeApply)
	if err != nil {
		return summary, facadeError(err)
	}
	fingerprint, err := mappingFingerprint(mapping, request.ExpectedSourceContentSHA256)
	if err != nil {
		return summary, err
	}
	if !sourceColumnsMatch(mapping.SourceColumns, d.Columns) || fingerprint != request.MappingFingerprint {
		return summary, importOwnerError("network_flow_source_changed", nil)
	}
	digest := sha256.New()
	fields := sourceFieldMappings(mapping)
	summary.Diagnostics = []rejectedRowDiagnostic{}
	summary.Sample = []rejectedRowDiagnostic{}
	for {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		record, err := d.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return summary, facadeError(err)
		}
		// A pass without a visitor keeps only counts and digests.
		row, diagnostics, err := validateCSVRecordWith(record, d.Columns, mapping, request.MappingFingerprint, request.ExpectedSourceContentSHA256, f.limits, fields, visit != nil)
		if err != nil {
			return summary, err
		}
		if len(diagnostics) == 0 {
			summary.Accepted++
			if summary.Accepted > f.limits.MaxAcceptedRowsPerTable {
				return summary, importOwnerError("network_flow_target_unavailable", map[string]any{"reason_code": "network_flow_resource_limit_exceeded"})
			}
			_, _ = digest.Write([]byte(row.SourceRowDigestSHA256 + row.NormalizedRowDigestSHA256))
			if visit != nil {
				if err := visit(row); err != nil {
					return summary, err
				}
			}
			continue
		}
		summary.Rejected++
		for _, diagnostic := range diagnostics {
			summary.DiagnosticCount++
			encoded, err := canonicalJSON(diagnostic)
			if err != nil {
				return summary, err
			}
			_, _ = digest.Write(encoded)
			if int64(len(summary.Diagnostics)) < f.limits.MaxRejectedRowDiagnostics {
				summary.Diagnostics = append(summary.Diagnostics, diagnostic)
			}
			if len(summary.Sample) < allRejectedDiagnosticSampleLimit {
				summary.Sample = append(summary.Sample, diagnostic)
			}
		}
	}
	summary.ValidationSHA = hex.EncodeToString(digest.Sum(nil))
	return summary, nil
}
