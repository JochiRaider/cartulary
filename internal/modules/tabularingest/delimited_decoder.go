package tabularingest

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrInvalidDelimited distinguishes lexical failures from source I/O failures.
var ErrInvalidDelimited = errors.New("invalid delimited input")

type delimitedSyntaxError string

func (e delimitedSyntaxError) Error() string { return string(e) }
func (e delimitedSyntaxError) Unwrap() error { return ErrInvalidDelimited }

// DelimitedRecord describes one logical record. Fields is nil when retention
// was disabled; FieldCount still accounts for every explicit field.
type DelimitedRecord struct {
	Fields     []string
	FieldCount int
}

// DelimitedDecoder is the shared CSV/TSV lexer. Geometry, retention, limits on
// rows, and missing-field meaning belong to its lifecycle adapter.
type DelimitedDecoder struct {
	ctx        context.Context
	reader     *bufio.Reader
	delimiter  byte
	maxColumns int
	processed  uint64
	terminal   error
}

func NewDelimitedDecoder(ctx context.Context, source io.Reader, format string, maxColumns int) (*DelimitedDecoder, error) {
	normalized, err := normalizedSourceFormat("", format)
	if err != nil {
		return nil, err
	}
	delimiter := byte('\t')
	if normalized == SourceFormatCSV {
		delimiter = ','
	}
	return &DelimitedDecoder{ctx: ctx, reader: bufio.NewReaderSize(source, 64*1024), delimiter: delimiter, maxColumns: maxColumns}, nil
}

// Next validates a complete record, including quoted line endings. Disabling
// retention discards decoded text while preserving lexical and width checks.
// The caller must read through EOF to certify the entire source.
func (d *DelimitedDecoder) Next(retain bool) (DelimitedRecord, error) {
	if d.terminal != nil {
		return DelimitedRecord{}, d.terminal
	}
	record, err := d.next(retain)
	if err != nil {
		d.terminal = err
		return DelimitedRecord{}, err
	}
	return record, nil
}

func (d *DelimitedDecoder) next(retain bool) (DelimitedRecord, error) {
	if err := d.ctx.Err(); err != nil {
		return DelimitedRecord{}, err
	}
	if _, err := d.peek(); err != nil {
		return DelimitedRecord{}, err
	}
	var record DelimitedRecord
	for {
		var value strings.Builder
		write := func(b byte) {
			if retain {
				value.WriteByte(b)
			}
		}
		first, err := d.peek()
		if err != nil && err != io.EOF {
			return DelimitedRecord{}, err
		}
		if err == nil && first == '"' {
			if _, err := d.read(); err != nil {
				return DelimitedRecord{}, err
			}
			for {
				b, err := d.read()
				if err == io.EOF {
					return DelimitedRecord{}, delimitedSyntaxError("malformed_quotes")
				}
				if err != nil {
					return DelimitedRecord{}, err
				}
				if b != '"' {
					write(b)
					continue
				}
				next, err := d.peek()
				if err != nil && err != io.EOF {
					return DelimitedRecord{}, err
				}
				if err == nil && next == '"' {
					if _, err := d.read(); err != nil {
						return DelimitedRecord{}, err
					}
					write('"')
					continue
				}
				break
			}
		} else {
			for {
				b, err := d.peek()
				if err == io.EOF {
					break
				}
				if err != nil {
					return DelimitedRecord{}, err
				}
				if b == d.delimiter || b == '\r' || b == '\n' {
					break
				}
				if b == '"' {
					return DelimitedRecord{}, delimitedSyntaxError("malformed_quotes")
				}
				if _, err := d.read(); err != nil {
					return DelimitedRecord{}, err
				}
				write(b)
			}
		}
		record.FieldCount++
		if d.maxColumns > 0 && record.FieldCount > d.maxColumns {
			return DelimitedRecord{}, fmt.Errorf("too_many_columns: tabular column count exceeded")
		}
		if retain {
			record.Fields = append(record.Fields, value.String())
		}
		b, err := d.peek()
		if err == io.EOF {
			return record, nil
		}
		if err != nil {
			return DelimitedRecord{}, err
		}
		if b != d.delimiter && b != '\r' && b != '\n' {
			return DelimitedRecord{}, delimitedSyntaxError("malformed_quotes")
		}
		if _, err := d.read(); err != nil {
			return DelimitedRecord{}, err
		}
		if b == d.delimiter {
			continue
		}
		if b == '\r' {
			next, err := d.peek()
			if err != nil && err != io.EOF {
				return DelimitedRecord{}, err
			}
			if err == nil && next == '\n' {
				if _, err := d.read(); err != nil {
					return DelimitedRecord{}, err
				}
			}
		}
		return record, nil
	}
}

func (d *DelimitedDecoder) peek() (byte, error) {
	data, err := d.reader.Peek(1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

func (d *DelimitedDecoder) read() (byte, error) {
	if d.processed%4096 == 0 {
		if err := d.ctx.Err(); err != nil {
			return 0, err
		}
	}
	d.processed++
	return d.reader.ReadByte()
}
