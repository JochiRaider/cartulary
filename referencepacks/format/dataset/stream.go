package dataset

import (
	"bufio"
	"context"
	"errors"
	"io"
)

type contextReader struct {
	ctx context.Context
	io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.Reader.Read(p)
}
func operational(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrStorage
}

func descriptorBytes(ctx context.Context, source Source) (raw []byte, err error) {
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	stream, err := source.Open(ctx, DescriptorPath)
	if err != nil {
		return nil, operational(ctx)
	}
	defer func() {
		if stream.Close() != nil {
			err = operational(ctx)
		}
	}()
	buffer := make([]byte, MaxDescriptorBytes)
	reader := contextReader{ctx, stream}
	n, err := io.ReadFull(reader, buffer)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return buffer[:n], nil
	}
	if err != nil {
		return nil, operational(ctx)
	}
	var probe [1]byte
	_, err = io.ReadFull(reader, probe[:])
	if errors.Is(err, io.EOF) {
		return buffer, nil
	}
	if err != nil {
		return nil, operational(ctx)
	}
	return nil, excess("max_dataset_descriptor_bytes", MaxDescriptorBytes, MaxDescriptorBytes+1, true)
}

func rows(ctx context.Context, source Source, path string, expected, maximum int64, visit func([]byte) error) error {
	return scanRows(ctx, source, path, expected, maximum, func(_ int64, raw []byte, terminated bool) error {
		if !terminated {
			return invalid()
		}
		return visit(raw)
	}, func(_ int64, err error) error { return err })
}

// scanRows stops at the first resource guard, but lets diagnostic callers
// continue across malformed bounded rows. Every opened member is closed even
// when the sink, context or source fails.
func scanRows(ctx context.Context, source Source, path string, expected, maximum int64, visit func(int64, []byte, bool) error, report func(int64, error) error) (err error) {
	if err = ctx.Err(); err != nil {
		return err
	}
	stream, err := source.Open(ctx, path)
	if err != nil {
		return operational(ctx)
	}
	defer func() {
		if stream.Close() != nil {
			err = operational(ctx)
		}
	}()
	reader := bufio.NewReaderSize(contextReader{ctx, stream}, 65536)
	var count int64
	var row []byte
	for {
		if err = ctx.Err(); err != nil {
			return err
		}
		fragment, readErr := reader.ReadSlice('\n')
		if len(fragment) > MaxRowBytes-len(row) {
			return report(count, excess("max_line_bytes", MaxRowBytes, uint64(len(row)+len(fragment)), true))
		}
		if len(fragment) > cap(row)-len(row) {
			capacity := min(MaxRowBytes, max(len(row)+len(fragment), cap(row)*2))
			next := make([]byte, len(row), capacity)
			copy(next, row)
			row = next
		}
		row = append(row, fragment...)
		if errors.Is(readErr, bufio.ErrBufferFull) {
			continue
		}
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return operational(ctx)
		}
		if len(row) == 0 && errors.Is(readErr, io.EOF) {
			if count != expected {
				return report(-1, invalid())
			}
			return nil
		}
		if count >= maximum || count >= expected {
			return report(-1, invalid())
		}
		terminated := readErr == nil
		data := row
		if terminated {
			data = row[:len(row)-1]
		}
		if err := visit(count, data, terminated); err != nil {
			return err
		}
		count++
		if !terminated {
			if count != expected {
				return report(-1, invalid())
			}
			return nil
		}
		row = row[:0]
	}
}
