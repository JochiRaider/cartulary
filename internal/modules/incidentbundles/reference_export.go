package incidentbundles

import (
	"context"
	"errors"
	"io"
)

// Size declarations are checked before reading and at EOF. The member store
// accounts all source owners together, including generated transport catalogs.
func boundedContainerSink(members *memberStore) func(context.Context, string, int64, io.Reader) error {
	return func(ctx context.Context, path string, size int64, source io.Reader) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if size < 0 || size > members.maximumBytes()-members.total {
			return &verificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
		}
		if _, exists := members.members[path]; exists {
			return errors.New("incident bundle: duplicate owner container")
		}
		if err := checkMemberCount(len(members.members)+3, members.limits); err != nil {
			return err
		}
		_, err := members.WriteMember(ctx, path, &exactContainerReader{source: source, remaining: size})
		return err
	}
}

type exactContainerReader struct {
	source    io.Reader
	remaining int64
}

func (r *exactContainerReader) Read(p []byte) (int, error) {
	if r.remaining == 0 {
		var probe [1]byte
		n, err := r.source.Read(probe[:])
		if n != 0 {
			return 0, errors.New("incident bundle: incomplete owner container")
		}
		return 0, err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.source.Read(p)
	r.remaining -= int64(n)
	if errors.Is(err, io.EOF) && r.remaining != 0 {
		return n, io.ErrUnexpectedEOF
	}
	return n, err
}
