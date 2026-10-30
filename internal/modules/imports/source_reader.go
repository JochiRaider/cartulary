package imports

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

const sourceReadChunkBytes = 64 * 1024

// sourceReader retains at most one database slice. Every query is completed
// before Read returns, so a borrowed connection can also insert accepted rows.
type sourceReader struct {
	mu         sync.Mutex
	ctx        context.Context
	tx         pgx.Tx
	capability ImportSourceCapability
	owned      bool
	stop       func() bool
	terminal   error
	buffer     []byte
	offset     int64
	digest     hash.Hash
}

func newSourceReader(ctx context.Context, tx pgx.Tx, capability ImportSourceCapability, owned bool) *sourceReader {
	r := &sourceReader{ctx: ctx, tx: tx, capability: capability, owned: owned, digest: sha256.New()}
	r.mu.Lock()
	r.stop = context.AfterFunc(ctx, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.finish(ctx.Err())
	})
	r.mu.Unlock()
	return r
}

func (r *sourceReader) Read(dst []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.terminal != nil {
		return 0, r.terminal
	}
	if err := r.ctx.Err(); err != nil {
		return 0, r.finish(err)
	}
	if len(dst) == 0 {
		return 0, nil
	}
	if len(r.buffer) == 0 {
		remaining := r.capability.SourceByteSize - r.offset
		if remaining < 0 {
			return 0, r.finish(importApplyBlockedError("source_changed"))
		}
		length := int64(sourceReadChunkBytes)
		if remaining < length {
			length = remaining + 1
		}
		var data []byte
		err := r.tx.QueryRow(r.ctx, `SELECT substring(source_bytes FROM $2 FOR $3)
FROM import_source_streams WHERE source_stream_ref=$1`, r.capability.SourceStreamRef, r.offset+1, length).Scan(&data)
		if err != nil {
			return 0, r.finish(err)
		}
		want := min(remaining, int64(sourceReadChunkBytes))
		if int64(len(data)) != want {
			return 0, r.finish(importApplyBlockedError("source_changed"))
		}
		if remaining == 0 {
			if hex.EncodeToString(r.digest.Sum(nil)) != r.capability.SourceContentSHA256 {
				return 0, r.finish(importApplyBlockedError("source_changed"))
			}
			return 0, r.finish(io.EOF)
		}
		r.offset += int64(len(data))
		_, _ = r.digest.Write(data)
		r.buffer = data
	}
	n := copy(dst, r.buffer)
	r.buffer = r.buffer[n:]
	return n, nil
}

func (r *sourceReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.terminal != nil {
		return nil
	}
	err := r.finish(io.ErrClosedPipe)
	if err == io.ErrClosedPipe {
		return nil
	}
	return err
}

// finish is called with the mutex held. EOF is only recorded after integrity
// checks; explicit close and cancellation cannot certify the source.
func (r *sourceReader) finish(err error) error {
	if r.terminal != nil {
		return r.terminal
	}
	if r.stop != nil {
		r.stop()
	}
	r.buffer = nil
	if r.owned {
		if closeErr := closeSourceSnapshot(r.tx); closeErr != nil {
			if err == io.EOF {
				err = closeErr
			} else {
				err = errors.Join(err, closeErr)
			}
		}
	}
	r.terminal = err
	return err
}

func closeSourceSnapshot(tx pgx.Tx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := tx.Rollback(ctx)
	if errors.Is(err, pgx.ErrTxClosed) {
		return nil
	}
	return err
}
