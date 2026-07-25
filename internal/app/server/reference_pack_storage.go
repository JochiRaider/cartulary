package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"regexp"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

var referencePackSHA256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type referencePackRootStorage struct {
	temporary *rootedfs.Root
	published *rootedfs.Root
}

func newReferencePackRootStorage(temporaryRoot string, publishedRoot string) (*referencePackRootStorage, error) {
	temporary, err := rootedfs.OpenOrCreate(temporaryRoot)
	if err != nil {
		return nil, fmt.Errorf("open Reference Pack temporary capability: %w", err)
	}
	published, err := rootedfs.OpenOrCreate(publishedRoot)
	if err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("open Reference Pack storage capability: %w", err)
	}
	return &referencePackRootStorage{temporary: temporary, published: published}, nil
}

func (storage *referencePackRootStorage) Close() {
	if storage == nil {
		return
	}
	_ = storage.published.Close()
	_ = storage.temporary.Close()
}

func (storage *referencePackRootStorage) Stage(ctx context.Context, fileSHA string, data []byte) (reference_data.StagingRef, error) {
	if storage == nil || storage.temporary == nil {
		return reference_data.StagingRef{}, errors.New("Reference Pack temporary storage is unavailable")
	}
	namePrefix := fileSHA
	if !referencePackSHA256Pattern.MatchString(namePrefix) {
		namePrefix = uuid.NewString()
	}
	reference, err := reference_data.ParseStagingRef(
		"reference-packs/imports/" + namePrefix + "-" + uuid.NewString() + ".bundle",
	)
	if err != nil {
		return reference_data.StagingRef{}, err
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return reference_data.StagingRef{}, err
	}
	if err := storage.temporary.MakePrivateDir(rootedfs.MustParseReference("reference-packs/imports")); err != nil {
		return reference_data.StagingRef{}, err
	}
	if err := storage.temporary.CreateExclusive(ctx, rootReference, referencePackBytesWriter(data)); err != nil {
		return reference_data.StagingRef{}, err
	}
	return reference, nil
}

func (storage *referencePackRootStorage) Publish(ctx context.Context, bundleSHA string, data []byte) (reference_data.StorageRef, error) {
	if storage == nil || storage.published == nil {
		return reference_data.StorageRef{}, errors.New("Reference Pack published storage is unavailable")
	}
	if !referencePackSHA256Pattern.MatchString(bundleSHA) {
		return reference_data.StorageRef{}, errors.New("Reference Pack bundle digest is invalid")
	}
	reference, err := reference_data.ParseStorageRef(
		"reference-packs/bundles/" + bundleSHA + "-" + uuid.NewString() + ".bundle",
	)
	if err != nil {
		return reference_data.StorageRef{}, err
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return reference_data.StorageRef{}, err
	}
	if err := storage.published.MakePrivateDir(rootedfs.MustParseReference("reference-packs/bundles")); err != nil {
		return reference_data.StorageRef{}, err
	}
	if err := storage.published.CreateExclusive(ctx, rootReference, referencePackBytesWriter(data)); err != nil {
		return reference_data.StorageRef{}, err
	}
	return reference, nil
}

func (storage *referencePackRootStorage) ReadStaged(reference reference_data.StagingRef, maxBytes int64) ([]byte, error) {
	if storage == nil || storage.temporary == nil {
		return nil, errors.New("Reference Pack temporary storage is unavailable")
	}
	return readReferencePackRegular(storage.temporary, reference.String(), maxBytes)
}

func (storage *referencePackRootStorage) ReadPublished(reference reference_data.StorageRef, maxBytes int64) ([]byte, error) {
	if storage == nil || storage.published == nil {
		return nil, errors.New("Reference Pack published storage is unavailable")
	}
	return readReferencePackRegular(storage.published, reference.String(), maxBytes)
}

func (storage *referencePackRootStorage) RemoveStaged(reference reference_data.StagingRef) error {
	if storage == nil || storage.temporary == nil {
		return errors.New("Reference Pack temporary storage is unavailable")
	}
	return removeReferencePackRegular(storage.temporary, reference.String())
}

func (storage *referencePackRootStorage) RemovePublished(reference reference_data.StorageRef) error {
	if storage == nil || storage.published == nil {
		return errors.New("Reference Pack published storage is unavailable")
	}
	return removeReferencePackRegular(storage.published, reference.String())
}

func readReferencePackRegular(root *rootedfs.Root, rawReference string, maxBytes int64) ([]byte, error) {
	reference, err := rootedfs.ParseReference(rawReference)
	if err != nil {
		return nil, err
	}
	data, _, err := root.ReadRegular(reference, maxBytes)
	return data, err
}

func removeReferencePackRegular(root *rootedfs.Root, rawReference string) error {
	reference, err := rootedfs.ParseReference(rawReference)
	if err != nil {
		return err
	}
	err = root.RemoveRegular(reference)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func referencePackBytesWriter(data []byte) rootedfs.WriteFunc {
	immutable := bytes.Clone(data)
	return func(destination io.Writer) error {
		_, err := io.Copy(destination, bytes.NewReader(immutable))
		return err
	}
}
