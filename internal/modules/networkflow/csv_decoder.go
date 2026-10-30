package networkflow

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"strings"
	"unicode/utf8"
)

type csvCell struct {
	ScalarCount int64
	SHA256      string
	OverLimit   bool
}

// csvDecoder owns only the header, its bounded preview samples and the current
// record. The raw hash covers read-ahead and the uninterpreted preview tail.
type csvDecoder struct {
	input        *bufio.Reader
	hash         hash.Hash
	expectedSHA  string
	limits       EffectiveLimits
	mode         parseMode
	dataCount    int64
	byteOffset   int64
	finished     bool
	Columns      []sourceColumnDescriptor
	SourceSHA256 string
}

func newCSVDecoder(reader io.Reader, expectedSHA string, limits EffectiveLimits, mode parseMode) (*csvDecoder, error) {
	d := &csvDecoder{hash: sha256.New(), expectedSHA: expectedSHA, limits: limits, mode: mode}
	d.input = bufio.NewReaderSize(io.TeeReader(reader, d.hash), 64*1024)
	if bom, _ := d.input.Peek(3); len(bom) == 3 && bom[0] == 0xef && bom[1] == 0xbb && bom[2] == 0xbf {
		_, _ = d.input.Discard(3)
		d.byteOffset = 3
	}
	header, err := d.readRecord(1, int(limits.MaxColumnsPerCSV), limits.MaxHeaderScalarLength)
	if err == io.EOF {
		reason := "zero_bytes"
		if d.byteOffset == 3 {
			reason = "bom_only"
		}
		return nil, d.sourceError("network_flow_csv_empty_file", reason, 1, nil)
	}
	if err != nil {
		return nil, err
	}
	if int64(header.RawFieldCount) > limits.MaxColumnsPerCSV {
		return nil, d.limitError("column_limit_exceeded", "network_flow.max_columns_per_csv", limits.MaxColumnsPerCSV, int64(header.RawFieldCount), 1, nil)
	}
	for index, cell := range header.Cells {
		if cell.OverLimit {
			return nil, d.limitError("header_scalar_limit_exceeded", "network_flow.max_header_scalar_length", limits.MaxHeaderScalarLength, cell.ScalarCount, 1, sourceColumnOrdinalPtr(index+1))
		}
	}
	d.Columns, err = sourceColumnsFromHeader(header.Fields, limits)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *csvDecoder) Next() (csvRecord, error) {
	if d.finished {
		return csvRecord{}, io.EOF
	}
	if d.mode == parseModePreview && d.dataCount == previewRecordLimit {
		return csvRecord{}, d.finish()
	}
	record, err := d.readRecord(d.dataCount+2, len(d.Columns), d.limits.MaxRawCellScalarLength)
	if err == io.EOF {
		if d.dataCount == 0 {
			return csvRecord{}, d.sourceError("network_flow_no_data_rows", "header_only", 1, nil)
		}
		return csvRecord{}, d.finish()
	}
	if err != nil {
		return csvRecord{}, err
	}
	d.dataCount++
	if d.dataCount > d.limits.MaxRowsPerCSV {
		return csvRecord{}, d.limitError("row_limit_exceeded", "network_flow.max_rows_per_csv", d.limits.MaxRowsPerCSV, d.dataCount, record.SourceRowNumber, nil)
	}
	record.FieldCountOK = record.RawFieldCount == len(d.Columns)
	if record.FieldCountOK && d.dataCount <= previewRecordLimit {
		addColumnSamples(d.Columns, record)
	}
	return record, nil
}

func (d *csvDecoder) finish() error {
	// The remaining bytes are deliberately not decoded as CSV or UTF-8.
	if _, err := io.Copy(io.Discard, d.input); err != nil {
		return err
	}
	d.SourceSHA256 = hex.EncodeToString(d.hash.Sum(nil))
	if d.expectedSHA != "" && d.expectedSHA != d.SourceSHA256 {
		return errSourceChanged
	}
	d.finished = true
	return io.EOF
}

func (d *csvDecoder) readRecord(number int64, maxFields int, scalarLimit int64) (csvRecord, error) {
	record := csvRecord{SourceRowNumber: number, Fields: []string{}, Cells: []csvCell{}}
	var value strings.Builder
	cellHash := sha256.New()
	var scalars int64
	started, quoted, closed, fieldStart := false, false, false, true
	appendRune := func(r rune) {
		scalars++
		var encoded [utf8.UTFMax]byte
		n := utf8.EncodeRune(encoded[:], r)
		_, _ = cellHash.Write(encoded[:n])
		if scalars <= scalarLimit && record.RawFieldCount < maxFields {
			value.WriteRune(r)
		}
	}
	endField := func() {
		if record.RawFieldCount < maxFields {
			record.Fields = append(record.Fields, value.String())
			record.Cells = append(record.Cells, csvCell{ScalarCount: scalars, SHA256: hex.EncodeToString(cellHash.Sum(nil)), OverLimit: scalars > scalarLimit})
		}
		record.RawFieldCount++
		value = strings.Builder{}
		cellHash.Reset()
		scalars = 0
		fieldStart = true
		closed = false
	}
	for {
		r, size, err := d.input.ReadRune()
		if err == io.EOF {
			if quoted {
				return csvRecord{}, d.sourceError("network_flow_csv_malformed_quote", "unclosed_quote", number, sourceColumnOrdinalPtr(record.RawFieldCount+1))
			}
			if !started {
				return csvRecord{}, io.EOF
			}
			endField()
			return record, nil
		}
		if err != nil {
			return csvRecord{}, err
		}
		d.byteOffset += int64(size)
		if r == utf8.RuneError && size == 1 {
			return csvRecord{}, d.sourceError("network_flow_invalid_utf8", "invalid_utf8_sequence", number, sourceColumnOrdinalPtr(record.RawFieldCount+1))
		}
		if r == '\ufeff' {
			return csvRecord{}, d.sourceError("network_flow_invalid_utf8", "bom_not_at_offset_zero", number, sourceColumnOrdinalPtr(record.RawFieldCount+1))
		}
		started = true
		if quoted {
			if r == '"' {
				quoted = false
				closed = true
			} else {
				appendRune(r)
			}
			continue
		}
		if closed && r == '"' {
			quoted = true
			closed = false
			appendRune(r)
			continue
		}
		if closed && r != ',' && r != '\r' && r != '\n' {
			return csvRecord{}, d.sourceError("network_flow_csv_malformed_quote", "invalid_after_closing_quote", number, sourceColumnOrdinalPtr(record.RawFieldCount+1))
		}
		if r == ',' {
			endField()
			continue
		}
		if r == '\r' || r == '\n' {
			if r == '\r' {
				if next, _ := d.input.Peek(1); len(next) == 1 && next[0] == '\n' {
					_, _ = d.input.Discard(1)
					d.byteOffset++
				}
			}
			endField()
			return record, nil
		}
		if fieldStart && r == '"' {
			quoted = true
			fieldStart = false
			continue
		}
		fieldStart = false
		appendRune(r)
	}
}

func (d *csvDecoder) sourceError(code, reason string, row int64, column *int64) *sourceValidationError {
	phase := "apply"
	if d.mode == parseModePreview {
		phase = "preview"
	}
	return &sourceValidationError{Code: code, ReasonCode: reason, Phase: phase, SourceRowNumber: row, SourceColumnOrdinal: column}
}

func (d *csvDecoder) limitError(reason, name string, limit, actual, row int64, column *int64) error {
	err := d.sourceError("network_flow_resource_limit_exceeded", reason, row, column)
	err.LimitName, err.LimitValue, err.ActualValue = name, limit, actual
	return err
}
