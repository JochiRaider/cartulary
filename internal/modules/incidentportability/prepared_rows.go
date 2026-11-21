package incidentportability

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
)

// ErrPreparedStream reports a staged write-ready stream that does not match
// its sealed preparation descriptor. Integrity comes from the staged seal; the
// final transaction re-proves no bundle fact (Core 01 REQ-01-641).
var ErrPreparedStream = errors.New("incident portability prepared stream does not match its preparation")

// PreparedRowStream is one sealed, operation-owned preparation member holding
// one owner's write-ready rows, one owner-encoded line per row. It never refers
// to a bundle member.
type PreparedRowStream struct {
	Member Member
	Rows   int64
}

// PreparedRowCodec is an owner's private line encoding for write-ready rows
// and their COPY projection. Decode reads only what Encode produced in the
// same operation, so it performs no admission.
type PreparedRowCodec[T any] interface {
	Encode(T) ([]byte, error)
	Decode([]byte) (T, error)
	Columns() []string
	Values(T) ([]any, error)
}

// StagePreparedRows writes the rows emitted by produce into one preparation
// member without a whole-member buffer. Each line obeys the logical record
// ceiling, and the returned descriptor binds the sealed size, digest and count.
// A stream with no rows creates no staged object; its descriptor is the zero
// member with zero rows (Core 01 REQ-01-641).
func StagePreparedRows[T any](ctx context.Context, destination MemberWriter, path string, codec PreparedRowCodec[T], produce func(emit func(T) error) error) (PreparedRowStream, error) {
	if destination == nil || codec == nil || produce == nil {
		return PreparedRowStream{}, ErrPreparedStream
	}
	var rows int64
	member, err := writeStagedMemberFrom(ctx, destination, path, func(ctx context.Context, out io.Writer) error {
		writer := bufio.NewWriterSize(out, 64<<10)
		err := produce(func(row T) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			line, err := codec.Encode(row)
			if err != nil {
				return err
			}
			if len(line) == 0 || len(line) > MaxNDJSONLogicalLineBytes || bytes.IndexByte(line, '\n') >= 0 {
				return strictNDJSONError(path, int(rows+1), "line_too_long")
			}
			if _, err := writer.Write(line); err != nil {
				return err
			}
			if err := writer.WriteByte('\n'); err != nil {
				return err
			}
			rows++
			return nil
		})
		if err != nil {
			return err
		}
		return writer.Flush()
	})
	if err != nil {
		return PreparedRowStream{}, err
	}
	return PreparedRowStream{Member: member, Rows: rows}, nil
}

// ReadPreparedRows visits one sealed stream in order. Reaching EOF completes
// the seal's size and digest verification; the count must match the stream.
// An empty stream opens nothing.
func ReadPreparedRows[T any](ctx context.Context, source MemberReader, stream PreparedRowStream, codec PreparedRowCodec[T], visit func(T) error) error {
	if stream.Rows == 0 {
		return requireEmptyStream(stream)
	}
	reader, err := openPreparedRows(ctx, source, stream, codec)
	if err != nil {
		return err
	}
	for reader.next() {
		if err := visit(reader.row); err != nil {
			return errors.Join(err, reader.close())
		}
	}
	return errors.Join(reader.finish(), reader.close())
}

// PreparedCopy is one sealed stream opened as a pgx.CopyFromSource. The owner
// issues the COPY into its own literal relation on the final transaction, then
// calls Finish, which requires the database acknowledgement, the rows read and
// the preparation count to agree and completes the seal verification.
type PreparedCopy[T any] struct {
	reader *preparedRowReader[T]
	stream PreparedRowStream
	keep   func(T) bool
	kept   int64
}

// OpenPreparedFilteredCopy opens stream as a COPY source of only the rows keep
// admits. Finish requires the acknowledgement to equal the admitted rows and
// the whole stream to match its preparation.
func OpenPreparedFilteredCopy[T any](ctx context.Context, source MemberReader, stream PreparedRowStream, codec PreparedRowCodec[T], keep func(T) bool) (*PreparedCopy[T], error) {
	if keep == nil {
		return nil, ErrPreparedStream
	}
	copySource, err := OpenPreparedCopy(ctx, source, stream, codec)
	if err != nil {
		return nil, err
	}
	copySource.keep = keep
	return copySource, nil
}

// OpenPreparedCopy opens stream from source, the operation's preparation store.
// An empty stream opens nothing.
func OpenPreparedCopy[T any](ctx context.Context, source MemberReader, stream PreparedRowStream, codec PreparedRowCodec[T]) (*PreparedCopy[T], error) {
	if stream.Rows == 0 {
		return &PreparedCopy[T]{stream: stream}, requireEmptyStream(stream)
	}
	reader, err := openPreparedRows(ctx, source, stream, codec)
	if err != nil {
		return nil, err
	}
	return &PreparedCopy[T]{reader: reader, stream: stream}, nil
}

func (c *PreparedCopy[T]) Next() bool {
	if c.reader == nil {
		return false
	}
	for {
		if !c.reader.next() {
			return false
		}
		if c.keep == nil || c.keep(c.reader.row) {
			break
		}
	}
	c.kept++
	values, err := c.reader.codec.Values(c.reader.row)
	if err != nil {
		c.reader.err = err
		return false
	}
	c.reader.values = values
	return true
}

func (c *PreparedCopy[T]) Values() ([]any, error) { return c.reader.values, nil }

// Err aborts the COPY statement on a read, decode or seal failure, including
// the digest verdict delivered at EOF.
func (c *PreparedCopy[T]) Err() error {
	if c.reader == nil {
		return nil
	}
	if c.reader.err != nil {
		return c.reader.err
	}
	if c.reader.drained && c.reader.count != c.stream.Rows {
		return ErrPreparedStream
	}
	return nil
}

// Finish closes the stream and checks the COPY outcome against preparation.
func (c *PreparedCopy[T]) Finish(written int64, copyErr error) error {
	if c.reader == nil {
		if copyErr != nil {
			return copyErr
		}
		if written != 0 {
			return ErrPreparedStream
		}
		return nil
	}
	if copyErr != nil {
		// The database or the source aborted the statement; keep its cause.
		return errors.Join(copyErr, c.reader.close())
	}
	if err := errors.Join(c.reader.finish(), c.reader.close()); err != nil {
		return err
	}
	expected := c.stream.Rows
	if c.keep != nil {
		expected = c.kept
	}
	if written != expected {
		return ErrPreparedStream
	}
	return nil
}

// requireEmptyStream admits only the empty descriptor: zero rows bind no
// staged object, so a zero-row stream naming a member is not empty.
func requireEmptyStream(stream PreparedRowStream) error {
	if stream.Rows != 0 || stream.Member != (Member{}) {
		return ErrPreparedStream
	}
	return nil
}

type preparedRowReader[T any] struct {
	stream  PreparedRowStream
	codec   PreparedRowCodec[T]
	weight  func(T) int64
	source  io.ReadCloser
	lines   *bufio.Reader
	row     T
	values  []any
	count   int64
	err     error
	drained bool
}

func openPreparedRows[T any](ctx context.Context, source MemberReader, stream PreparedRowStream, codec PreparedRowCodec[T]) (*preparedRowReader[T], error) {
	if source == nil || codec == nil || stream.Rows < 0 || stream.Member.Path == "" {
		return nil, ErrPreparedStream
	}
	reader, actual, err := source.OpenMember(ctx, stream.Member.Path)
	if err != nil {
		return nil, err
	}
	if reader == nil || actual != stream.Member {
		if reader != nil {
			_ = reader.Close()
		}
		return nil, ErrPreparedStream
	}
	return &preparedRowReader[T]{stream: stream, codec: codec, source: reader, lines: bufio.NewReaderSize(reader, 64<<10)}, nil
}

func (r *preparedRowReader[T]) next() bool {
	if r.err != nil || r.drained {
		return false
	}
	var line []byte
	for {
		fragment, err := r.lines.ReadSlice('\n')
		line = append(line, fragment...)
		if len(line) > MaxNDJSONLogicalLineBytes+1 {
			r.err = ErrPreparedStream
			return false
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		if errors.Is(err, io.EOF) {
			r.drained = true
			if len(line) != 0 {
				r.err = ErrPreparedStream
			}
			return false
		}
		if err != nil {
			r.err = err
			return false
		}
		break
	}
	row, err := r.codec.Decode(line[:len(line)-1])
	if err != nil {
		r.err = errors.Join(ErrPreparedStream, err)
		return false
	}
	if r.weight != nil {
		r.count += r.weight(row)
	} else {
		r.count++
	}
	if r.count > r.stream.Rows {
		r.err = ErrPreparedStream
		return false
	}
	r.row = row
	return true
}

// finish reports the first read, decode or seal failure, and a row count
// different from the preparation descriptor.
func (r *preparedRowReader[T]) finish() error {
	if r.err != nil {
		return r.err
	}
	if !r.drained || r.count != r.stream.Rows {
		return ErrPreparedStream
	}
	return nil
}

func (r *preparedRowReader[T]) close() error { return r.source.Close() }

// EncodeJSONRow encodes one staged owner row as a JSON line without HTML
// escaping, so a staged line stays close to the bundle line it was admitted
// from. Owner codecs whose row type round-trips through encoding/json use it.
func EncodeJSONRow(row any) ([]byte, error) {
	var line bytes.Buffer
	encoder := json.NewEncoder(&line)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(row); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(line.Bytes(), []byte("\n")), nil
}

// DecodeJSONRow decodes one staged JSON line into target.
func DecodeJSONRow(line []byte, target any) error {
	if err := json.Unmarshal(line, target); err != nil {
		return ErrPreparedStream
	}
	return nil
}

// RawRowCodec stages one admitted source object as JSON and decodes it with
// exact numbers, as the bundle reader does, for owners whose final writes
// consume the source object itself.
type RawRowCodec struct{}

func (RawRowCodec) Encode(row map[string]any) ([]byte, error) { return EncodeJSONRow(row) }

func (RawRowCodec) Decode(line []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.UseNumber()
	var row map[string]any
	if err := decoder.Decode(&row); err != nil || row == nil || decoder.More() {
		return nil, ErrPreparedStream
	}
	return row, nil
}

func (RawRowCodec) Columns() []string { return nil }

func (RawRowCodec) Values(map[string]any) ([]any, error) { return nil, ErrPreparedStream }
