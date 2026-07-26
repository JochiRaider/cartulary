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

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

var incidentBundleSHA256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type incidentBundleRootStorage struct {
	temporary *rootedfs.Root
	exports   *rootedfs.Root
}

func newIncidentBundleRootStorage(temporaryRoot string, exportRoot string) (*incidentBundleRootStorage, error) {
	temporary, err := rootedfs.OpenOrCreate(temporaryRoot)
	if err != nil {
		return nil, fmt.Errorf("open incident bundle temporary capability: %w", err)
	}
	exports, err := rootedfs.OpenOrCreate(exportRoot)
	if err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("open incident bundle export capability: %w", err)
	}
	return &incidentBundleRootStorage{temporary: temporary, exports: exports}, nil
}

func (storage *incidentBundleRootStorage) Close() {
	if storage == nil {
		return
	}
	_ = storage.exports.Close()
	_ = storage.temporary.Close()
}

func (storage *incidentBundleRootStorage) Stage(ctx context.Context, fileSHA string, data []byte) (incidentbundles.BundleStagingRef, error) {
	if storage == nil || storage.temporary == nil {
		return incidentbundles.BundleStagingRef{}, errors.New("incident bundle temporary storage is unavailable")
	}
	namePrefix := fileSHA
	if !incidentBundleSHA256Pattern.MatchString(namePrefix) {
		namePrefix = uuid.NewString()
	}
	reference, err := incidentbundles.ParseBundleStagingRef(
		"incident-bundles/imports/" + namePrefix + "-" + uuid.NewString() + ".bundle",
	)
	if err != nil {
		return incidentbundles.BundleStagingRef{}, err
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return incidentbundles.BundleStagingRef{}, err
	}
	if err := storage.temporary.MakePrivateDir(rootedfs.MustParseReference("incident-bundles/imports")); err != nil {
		return incidentbundles.BundleStagingRef{}, err
	}
	if err := storage.temporary.CreateExclusive(ctx, rootReference, bytesWriter(data)); err != nil {
		return incidentbundles.BundleStagingRef{}, err
	}
	return reference, nil
}

func (storage *incidentBundleRootStorage) Publish(ctx context.Context, bundleID string, data []byte) (incidentbundles.BundleStorageRef, error) {
	if storage == nil || storage.exports == nil {
		return incidentbundles.BundleStorageRef{}, errors.New("incident bundle export storage is unavailable")
	}
	parsedBundleID, err := uuid.Parse(bundleID)
	if err != nil {
		return incidentbundles.BundleStorageRef{}, errors.New("incident bundle export identifier is invalid")
	}
	reference, err := incidentbundles.ParseBundleStorageRef("incident-bundles/" + parsedBundleID.String() + ".zip")
	if err != nil {
		return incidentbundles.BundleStorageRef{}, err
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return incidentbundles.BundleStorageRef{}, err
	}
	if err := storage.exports.MakePrivateDir(rootedfs.MustParseReference("incident-bundles")); err != nil {
		return incidentbundles.BundleStorageRef{}, err
	}
	if err := storage.exports.CreateExclusive(ctx, rootReference, bytesWriter(data)); err != nil {
		return incidentbundles.BundleStorageRef{}, err
	}
	return reference, nil
}

func (storage *incidentBundleRootStorage) ReadStaged(reference incidentbundles.BundleStagingRef, maxBytes int64) ([]byte, error) {
	if storage == nil || storage.temporary == nil {
		return nil, errors.New("incident bundle temporary storage is unavailable")
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return nil, err
	}
	data, _, err := storage.temporary.ReadRegular(rootReference, maxBytes)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (storage *incidentBundleRootStorage) RemoveStaged(reference incidentbundles.BundleStagingRef) error {
	if storage == nil || storage.temporary == nil {
		return errors.New("incident bundle temporary storage is unavailable")
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return err
	}
	return removeIncidentBundleRegular(storage.temporary, rootReference)
}

func (storage *incidentBundleRootStorage) RemovePublished(reference incidentbundles.BundleStorageRef) error {
	if storage == nil || storage.exports == nil {
		return errors.New("incident bundle export storage is unavailable")
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return err
	}
	return removeIncidentBundleRegular(storage.exports, rootReference)
}

func removeIncidentBundleRegular(root *rootedfs.Root, reference rootedfs.Reference) error {
	err := root.RemoveRegular(reference)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func bytesWriter(data []byte) rootedfs.WriteFunc {
	immutable := bytes.Clone(data)
	return func(destination io.Writer) error {
		_, err := io.Copy(destination, bytes.NewReader(immutable))
		return err
	}
}

type sharedIncidentBundleStorage struct {
	bytes sharedPublicationBytes
}

func newSharedIncidentBundleStorage(store objectstore.Store) incidentbundles.BundleStorage {
	return &sharedIncidentBundleStorage{bytes: sharedPublicationBytes{store: store}}
}

func (storage *sharedIncidentBundleStorage) Stage(
	ctx context.Context,
	fileSHA string,
	data []byte,
) (incidentbundles.BundleStagingRef, error) {
	namePrefix := fileSHA
	if !incidentBundleSHA256Pattern.MatchString(namePrefix) {
		namePrefix = uuid.NewString()
	}
	reference, err := incidentbundles.ParseBundleStagingRef(
		"incident-bundles/imports/" + namePrefix + "-" + uuid.NewString() + ".bundle",
	)
	if err != nil {
		return incidentbundles.BundleStagingRef{}, err
	}
	if err := storage.bytes.put(ctx, reference.String(), data, "application/octet-stream"); err != nil {
		return incidentbundles.BundleStagingRef{}, err
	}
	return reference, nil
}

func (storage *sharedIncidentBundleStorage) Publish(
	ctx context.Context,
	bundleID string,
	data []byte,
) (incidentbundles.BundleStorageRef, error) {
	parsedBundleID, err := uuid.Parse(bundleID)
	if err != nil {
		return incidentbundles.BundleStorageRef{}, errors.New("incident bundle export identifier is invalid")
	}
	reference, err := incidentbundles.ParseBundleStorageRef("incident-bundles/" + parsedBundleID.String() + ".zip")
	if err != nil {
		return incidentbundles.BundleStorageRef{}, err
	}
	if err := storage.bytes.put(ctx, reference.String(), data, "application/zip"); err != nil {
		return incidentbundles.BundleStorageRef{}, err
	}
	return reference, nil
}

func (storage *sharedIncidentBundleStorage) ReadStaged(
	reference incidentbundles.BundleStagingRef,
	maxBytes int64,
) ([]byte, error) {
	return storage.bytes.read(reference.String(), maxBytes)
}

func (storage *sharedIncidentBundleStorage) RemoveStaged(reference incidentbundles.BundleStagingRef) error {
	return storage.bytes.remove(reference.String())
}

func (storage *sharedIncidentBundleStorage) RemovePublished(reference incidentbundles.BundleStorageRef) error {
	return storage.bytes.remove(reference.String())
}
