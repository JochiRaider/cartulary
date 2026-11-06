package incidentportability

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"io"
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
func ExportNDJSONTo(ctx context.Context, query Queryer, incidentID uuid.UUID, path, sql string, destination MemberWriter) error {
	rows, err := query.Query(ctx, sql, incidentID)
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
	producerContext, cancel := context.WithCancel(ctx)
	defer cancel()
	reader, writer := io.Pipe()
	done := make(chan error, 1)
	go func() {
		err := write(producerContext, writer)
		_ = writer.CloseWithError(err)
		done <- err
	}()
	_, err := destination.WriteMember(ctx, path, reader)
	cancel()
	_ = reader.Close()
	produced := <-done
	if err != nil {
		return err
	}
	if produced != nil {
		return produced
	}
	return ctx.Err()
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
