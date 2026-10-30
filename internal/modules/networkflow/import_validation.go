package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/jackc/pgx/v5"
)

type importValidationSummary struct {
	Accepted, Rejected, DiagnosticCount        int64
	SourceSHA, ValidationSHA, OriginalFilename string
	Columns                                    []sourceColumnDescriptor
	Diagnostics, Sample                        []rejectedRowDiagnostic
}

func (s importValidationSummary) agrees(other importValidationSummary) bool {
	return s.Accepted == other.Accepted && s.Rejected == other.Rejected && s.DiagnosticCount == other.DiagnosticCount &&
		s.SourceSHA == other.SourceSHA && s.ValidationSHA == other.ValidationSHA && s.OriginalFilename == other.OriginalFilename &&
		bytes.Equal(canonicalJSON(s.Columns), canonicalJSON(other.Columns))
}

// validateImportPass retains no accepted rows. The second pass inserts each
// accepted row only after its source query has released the shared connection.
func (f *importFacade) validateImportPass(ctx context.Context, tx pgx.Tx, request imports.ExtensionImportApplyRequest, mapping approvedMapping, retain bool, visit func(flowRow) error) (importValidationSummary, error) {
	var summary importValidationSummary
	stream, err := f.sourceStore.OpenSourceStreamTx(ctx, tx, request.SourceCapability.SourceStreamRef)
	if err != nil {
		return summary, err
	}
	defer stream.Reader.Close()
	if stream.ImportSourceCapability != request.SourceCapability {
		return summary, importOwnerError("network_flow_source_changed", nil)
	}
	d, err := newCSVDecoder(stream.Reader, request.ExpectedSourceContentSHA256, f.limits, parseModeApply)
	if err != nil {
		return summary, facadeError(err)
	}
	if !sourceColumnsMatch(mapping.SourceColumns, d.Columns) || mappingFingerprint(mapping, request.ExpectedSourceContentSHA256) != request.MappingFingerprint {
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
		row, diagnostics := validateCSVRecord(record, d.Columns, mapping, request.MappingFingerprint, request.ExpectedSourceContentSHA256, f.limits, fields)
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
			_, _ = digest.Write(canonicalJSON(diagnostic))
			if retain {
				if int64(len(summary.Diagnostics)) < f.limits.MaxRejectedRowDiagnostics {
					summary.Diagnostics = append(summary.Diagnostics, diagnostic)
				}
				if len(summary.Sample) < previewRecordLimit {
					summary.Sample = append(summary.Sample, diagnostic)
				}
			}
		}
	}
	summary.SourceSHA, summary.Columns, summary.OriginalFilename = d.SourceSHA256, d.Columns, stream.OriginalFilename
	summary.ValidationSHA = hex.EncodeToString(digest.Sum(nil))
	return summary, nil
}
