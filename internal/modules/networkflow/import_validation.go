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
	Accepted, Rejected, DiagnosticCount        int64
	SourceSHA, ValidationSHA, OriginalFilename string
	Columns                                    []sourceColumnDescriptor
	Diagnostics, Sample                        []rejectedRowDiagnostic
}

func (s importValidationSummary) agrees(other importValidationSummary) bool {
	return s.Accepted == other.Accepted && s.Rejected == other.Rejected && s.DiagnosticCount == other.DiagnosticCount &&
		s.SourceSHA == other.SourceSHA && s.ValidationSHA == other.ValidationSHA && s.OriginalFilename == other.OriginalFilename &&
		canonicalEqual(s.Columns, other.Columns)
}

// validateImportPass retains no accepted rows. The second pass inserts each
// accepted row only after its source query has released the shared connection.
func (f *importFacade) validateImportPass(ctx context.Context, tx pgx.Tx, request imports.ExtensionImportApplyRequest, mapping approvedMapping, retain bool, visit func(normalizedRow) error) (importValidationSummary, error) {
	return f.validateImportPassGated(ctx, tx, request, mapping, retain, visit, nil)
}

// validateImportPassGated runs one source pass. With a gate, the pass shares
// the borrowed connection with a pipelined COPY: it opens and reads the source
// only while the gate holds the connection, reading ahead so each hand-off
// covers a larger source range.
func (f *importFacade) validateImportPassGated(ctx context.Context, tx pgx.Tx, request imports.ExtensionImportApplyRequest, mapping approvedMapping, retain bool, visit func(normalizedRow) error, gate rowSourceGate) (importValidationSummary, error) {
	var summary importValidationSummary
	var stream imports.ImportSourceStream
	open := func() error {
		var err error
		stream, err = f.sourceStore.OpenSourceStreamTx(ctx, tx, request.SourceCapability.SourceStreamRef)
		return err
	}
	var err error
	if gate != nil {
		err = gate.withConnection(open)
	} else {
		err = open()
	}
	if err != nil {
		return summary, err
	}
	defer stream.Reader.Close()
	if stream.ImportSourceCapability != request.SourceCapability {
		return summary, importOwnerError("network_flow_source_changed", nil)
	}
	var reader io.Reader = stream.Reader
	if gate != nil {
		reader = gate.readAhead(stream.Reader, pipelinedSourceReadAhead)
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
			if retain {
				if int64(len(summary.Diagnostics)) < f.limits.MaxRejectedRowDiagnostics {
					summary.Diagnostics = append(summary.Diagnostics, diagnostic)
				}
				if len(summary.Sample) < allRejectedDiagnosticSampleLimit {
					summary.Sample = append(summary.Sample, diagnostic)
				}
			}
		}
	}
	summary.SourceSHA, summary.Columns, summary.OriginalFilename = d.SourceSHA256, d.Columns, stream.OriginalFilename
	summary.ValidationSHA = hex.EncodeToString(digest.Sum(nil))
	return summary, nil
}
