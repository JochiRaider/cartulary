package incidentportability

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

// Member is bounded metadata for an immutable logical archive member. It never
// contains source bytes, a host path, a storage key or an executable callback.
type Member struct {
	Path     string `json:"path"`
	ByteSize int64  `json:"byte_size"`
	SHA256   string `json:"sha256"`
}

type MemberReader interface {
	Members() []Member
	OpenMember(context.Context, string) (io.ReadCloser, Member, error)
}

type MemberWriter interface {
	WriteMember(context.Context, string, io.Reader) (Member, error)
}

// ExportNDJSONTo holds at most one canonical source record while the admitted
// member writer consumes it. The caller's read snapshot remains authoritative.
func ExportNDJSONTo(ctx context.Context, query Queryer, incidentID uuid.UUID, path string, sql sqlshape.Fragment, destination MemberWriter) error {
	rows, err := query.Query(ctx, sqlshape.Text(sql), incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	_, err = destination.WriteMember(ctx, path, &ndjsonRowsReader{ctx: ctx, rows: rows})
	return err
}

type ndjsonRowsReader struct {
	ctx     context.Context
	rows    pgx.Rows
	pending []byte
}

func (r *ndjsonRowsReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}
	if len(r.pending) == 0 {
		if !r.rows.Next() {
			if err := r.rows.Err(); err != nil {
				return 0, err
			}
			return 0, io.EOF
		}
		var raw []byte
		if err := r.rows.Scan(&raw); err != nil {
			return 0, err
		}
		if len(raw) > MaxNDJSONLogicalLineBytes {
			return 0, strictNDJSONError("", 0, "line_too_long")
		}
		encoded, err := CanonicalRawJSON(raw)
		if err != nil {
			return 0, err
		}
		if len(encoded)-1 > MaxNDJSONLogicalLineBytes {
			return 0, strictNDJSONError("", 0, "line_too_long")
		}
		r.pending = encoded
	}
	n := copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

// WriteMemberFrom connects an owner encoder to the member store without a
// whole-member buffer. The producer is cancelled and joined before returning,
// including when storage stops consuming early.
func WriteMemberFrom(ctx context.Context, destination MemberWriter, path string, write func(context.Context, io.Writer) error) error {
	_, err := writeMemberFrom(ctx, destination, path, write)
	return err
}

// writeMemberFrom is WriteMemberFrom returning the sealed member descriptor.
func writeMemberFrom(ctx context.Context, destination MemberWriter, path string, write func(context.Context, io.Writer) error) (Member, error) {
	producerContext, cancel := context.WithCancel(ctx)
	defer cancel()
	reader, writer := io.Pipe()
	done := make(chan error, 1)
	go func() {
		err := write(producerContext, writer)
		_ = writer.CloseWithError(err)
		done <- err
	}()
	member, err := destination.WriteMember(ctx, path, reader)
	cancel()
	_ = reader.Close()
	produced := <-done
	if err != nil {
		return Member{}, err
	}
	if produced != nil {
		return Member{}, produced
	}
	return member, ctx.Err()
}

// writeStagedMemberFrom is writeMemberFrom for preparation streams. The
// member is allocated with the producer's first byte, so a producer that
// writes nothing stages no object and the zero Member is returned (Core 01
// REQ-01-641: a stream with no rows creates no staged object).
func writeStagedMemberFrom(ctx context.Context, destination MemberWriter, path string, write func(context.Context, io.Writer) error) (Member, error) {
	lazy := &lazyMemberWriter{ctx: ctx, destination: destination, path: path}
	return lazy.finish(write(ctx, lazy))
}

// lazyMemberWriter starts the destination write on the first byte. The
// destination consumes the pipe on its own goroutine while the producer
// writes, as writeMemberFrom does with the roles reversed.
type lazyMemberWriter struct {
	ctx         context.Context
	destination MemberWriter
	path        string
	pipe        *io.PipeWriter
	done        chan stagedMemberResult
}

type stagedMemberResult struct {
	member Member
	err    error
}

func (w *lazyMemberWriter) Write(payload []byte) (int, error) {
	if len(payload) == 0 {
		return 0, nil
	}
	if w.pipe == nil {
		reader, writer := io.Pipe()
		w.pipe, w.done = writer, make(chan stagedMemberResult, 1)
		go func() {
			member, err := w.destination.WriteMember(w.ctx, w.path, reader)
			// A destination that stops early unblocks the producer.
			_ = reader.Close()
			w.done <- stagedMemberResult{member: member, err: err}
		}()
	}
	return w.pipe.Write(payload)
}

func (w *lazyMemberWriter) finish(produced error) (Member, error) {
	if w.pipe == nil {
		if produced != nil {
			return Member{}, produced
		}
		return Member{}, w.ctx.Err()
	}
	_ = w.pipe.CloseWithError(produced)
	result := <-w.done
	if result.err != nil {
		return Member{}, result.err
	}
	if produced != nil {
		return Member{}, produced
	}
	return result.member, w.ctx.Err()
}

// WriteNDJSONRecord enforces the logical row ceiling independently of member
// size. Canonical encoding preserves numbers and appends exactly one LF.
func WriteNDJSONRecord(destination io.Writer, row map[string]any) error {
	encoded, err := CanonicalJSONString(row)
	if err != nil {
		return err
	}
	if len(encoded)-1 > MaxNDJSONLogicalLineBytes {
		return strictNDJSONError("", 0, "line_too_long")
	}
	n, err := destination.Write(encoded)
	if err == nil && n != len(encoded) {
		return io.ErrShortWrite
	}
	return err
}
