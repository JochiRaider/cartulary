package networkflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
)

// A single CSV decoder is paired with the source-ordered observation stream.
// The retained diagnostic set is bounded separately by its owner limit. Neither
// accepted records nor complete source files accumulate in preparation memory.
type portableSourceVerifier struct {
	table                               tableRecord
	reader                              io.ReadCloser
	decoder                             *csvDecoder
	mapping                             approvedMapping
	fields                              map[string]fieldMapping
	limits                              EffectiveLimits
	digest                              hash.Hash
	diagnostics                         map[string]rejectedRowDiagnostic
	retainedDiagnostics                 int64
	accepted, rejected, diagnosticCount int64
}

func (p *preparedNetworkFlow) openSourceVerifier(ctx context.Context, sources committedPortabilitySources, table tableRecord) (_ *portableSourceVerifier, err error) {
	source, reader, err := sources.OpenSource(ctx, table.CommittedSourceID)
	if err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, errPortableNetworkFlow
	}
	defer func() {
		if err != nil {
			_ = reader.Close()
		}
	}()
	if source.IncidentID != table.IncidentID || source.ContentSHA256 != table.SourceContentSHA256 || source.MappingFingerprint != table.MappingFingerprint || source.CreatedByUserID != table.CreatedByUserID || source.HeaderRowRef != 1 || source.DataStartRowRef != 2 || source.LocatorKind != "csv_file" || source.Locator != "file" || sanitizeSourceFilenameDisplay(source.OriginalFilename) != table.SourceFilenameDisplay {
		return nil, fmt.Errorf("committed source metadata: %w", errPortableNetworkFlow)
	}
	var envelope struct {
		OwnerMapping json.RawMessage `json:"owner_mapping"`
	}
	if err := json.Unmarshal(source.ApprovedMapping, &envelope); err != nil {
		return nil, err
	}
	limits := p.limits
	limits.MaxRawCellScalarLength = table.SourceRawCellScalarLimit
	mapping, err := decodeApprovedMapping(envelope.OwnerMapping, limits)
	if err != nil {
		return nil, err
	}
	decoder, err := newCSVDecoder(reader, table.SourceContentSHA256, limits, parseModeApply)
	if err != nil {
		return nil, err
	}
	if !sourceColumnsMatch(mapping.SourceColumns, decoder.Columns) {
		return nil, errPortableNetworkFlow
	}
	v := &portableSourceVerifier{table: table, reader: reader, decoder: decoder, mapping: mapping, fields: sourceFieldMappings(mapping), limits: limits, digest: sha256.New(), diagnostics: map[string]rejectedRowDiagnostic{}}
	err = walkPortable(ctx, p.members, "diagnostics.ndjson", portableDiagnosticMembers, "", func(d portableDiagnostic) error {
		if d.TableID == table.TableID {
			v.diagnostics[d.Diagnostic.DiagnosticID] = d.Diagnostic
			v.retainedDiagnostics++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (v *portableSourceVerifier) next(ctx context.Context) (flowRow, error) {
	for {
		if err := ctx.Err(); err != nil {
			return flowRow{}, err
		}
		record, err := v.decoder.Next()
		if err != nil {
			return flowRow{}, err
		}
		row, diagnostics, err := validateCSVRecord(record, v.decoder.Columns, v.mapping, v.table.MappingFingerprint, v.table.SourceContentSHA256, v.limits, v.fields)
		if err != nil {
			return flowRow{}, err
		}
		if len(diagnostics) == 0 {
			v.accepted++
			_, _ = v.digest.Write([]byte(row.SourceRowDigestSHA256 + row.NormalizedRowDigestSHA256))
			identity := rowImportContext{incidentID: v.table.IncidentID, actorUserID: v.table.CreatedByUserID, committedSourceID: v.table.CommittedSourceID, sourceSHA: v.table.SourceContentSHA256, fingerprint: v.table.MappingFingerprint, sourceProfile: v.table.SourceProfileID, parserProfile: v.table.ParserProfileID, createdAt: v.table.CreatedAt}
			return identity.persistedRow(v.table.TableID, row)
		}
		v.rejected++
		for _, diagnostic := range diagnostics {
			v.diagnosticCount++
			raw, err := canonicalJSON(diagnostic)
			if err != nil {
				return flowRow{}, err
			}
			_, _ = v.digest.Write(raw)
			if v.diagnosticCount <= v.retainedDiagnostics {
				retained, ok := v.diagnostics[diagnostic.DiagnosticID]
				if !ok || !canonicalEqual(retained, diagnostic) {
					return flowRow{}, fmt.Errorf("source diagnostic mismatch: %w", errPortableNetworkFlow)
				}
				delete(v.diagnostics, diagnostic.DiagnosticID)
			}
		}
	}
}

func (v *portableSourceVerifier) finish(ctx context.Context) error {
	_, err := v.next(ctx)
	if !errors.Is(err, io.EOF) {
		if err != nil {
			return err
		}
		return fmt.Errorf("missing accepted source observation: %w", errPortableNetworkFlow)
	}
	if len(v.diagnostics) != 0 || v.accepted != v.table.RowCountAccepted || v.rejected != v.table.RowCountRejected || v.table.DiagnosticsTruncated != (v.diagnosticCount > v.retainedDiagnostics) || hex.EncodeToString(v.digest.Sum(nil)) != v.table.SourceInterpretationSHA256 {
		return fmt.Errorf("committed source interpretation mismatch: %w", errPortableNetworkFlow)
	}
	return v.reader.Close()
}

func (p *preparedNetworkFlow) verifyCommittedObservations(ctx context.Context, sources committedPortabilitySources) error {
	if err := p.validateDiagnostics(ctx); err != nil {
		return err
	}
	var verifier *portableSourceVerifier
	defer func() {
		if verifier != nil {
			_ = verifier.reader.Close()
		}
	}()
	seen := map[string]bool{}
	var previous *flowRow
	err := walkPortable(ctx, p.members, "observations.ndjson", portableObservationMembers, portableObservationNullable, func(observation portableObservation) error {
		row := observation.retained()
		table, ok := p.tables[row.NetworkFlowTableID]
		if !ok || (previous != nil && comparePortableRows(*previous, row) >= 0) {
			return errPortableNetworkFlow
		}
		if verifier == nil || verifier.table.TableID != table.TableID {
			if verifier != nil {
				if err := verifier.finish(ctx); err != nil {
					return err
				}
			}
			var err error
			verifier, err = p.openSourceVerifier(ctx, sources, table)
			if err != nil {
				return err
			}
			seen[table.TableID] = true
		}
		if err := p.validateObservation(row, table); err != nil {
			return err
		}
		expected, err := verifier.next(ctx)
		if err != nil {
			return err
		}
		want, err := json.Marshal(portableFlowRow(expected))
		if err != nil {
			return err
		}
		actual, err := json.Marshal(observation)
		if err != nil {
			return err
		}
		if !canonicalEqual(json.RawMessage(want), json.RawMessage(actual)) {
			return fmt.Errorf("committed source observation mismatch: %w", errPortableNetworkFlow)
		}
		p.counts["observations"]++
		previous = &row
		return nil
	})
	if err != nil {
		return err
	}
	if verifier == nil || len(seen) != len(p.tables) {
		return errPortableNetworkFlow
	}
	return verifier.finish(ctx)
}
