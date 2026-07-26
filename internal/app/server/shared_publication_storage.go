package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const sharedPublicationOperationTimeout = 30 * time.Second

type sharedPublicationBytes struct {
	store objectstore.Store
}

func (storage sharedPublicationBytes) put(ctx context.Context, key string, data []byte, contentType string) error {
	if storage.store == nil {
		return errors.New("shared publication object storage is unavailable")
	}
	return storage.store.PutObject(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
}

func (storage sharedPublicationBytes) read(key string, maxBytes int64) ([]byte, error) {
	if storage.store == nil || maxBytes < 1 {
		return nil, errors.New("shared publication object storage is unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), sharedPublicationOperationTimeout)
	defer cancel()
	reader, info, err := storage.store.ReadObject(ctx, key, objectstore.ReadOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	if info.Size < 0 || info.Size > maxBytes {
		return nil, fmt.Errorf("shared publication object exceeds the admitted read limit")
	}
	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes || int64(len(data)) != info.Size {
		return nil, errors.New("shared publication object size changed while reading")
	}
	return data, nil
}

func (storage sharedPublicationBytes) remove(key string) error {
	if storage.store == nil {
		return errors.New("shared publication object storage is unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), sharedPublicationOperationTimeout)
	defer cancel()
	return storage.store.DeleteObject(ctx, key)
}
