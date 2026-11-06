package incidentportabilityassembly

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

// RootStorage realizes the Incident Bundles storage port with admitted,
// descriptor-anchored filesystem roots.
type RootStorage struct {
	temporary *rootedfs.Root
	exports   *rootedfs.Root
}

var _ incidentbundles.BundleStorage = (*RootStorage)(nil)

// NewRootStorage opens the temporary and published-bundle root capabilities.
// The caller owns the returned storage and must close it.
func NewRootStorage(temporaryRoot string, exportRoot string) (*RootStorage, error) {
	temporary, err := rootedfs.OpenOrCreate(temporaryRoot)
	if err != nil {
		return nil, fmt.Errorf("open incident bundle temporary capability: %w", err)
	}
	exports, err := rootedfs.OpenOrCreate(exportRoot)
	if err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("open incident bundle export capability: %w", err)
	}
	return &RootStorage{temporary: temporary, exports: exports}, nil
}

func (storage *RootStorage) Close() {
	if storage == nil {
		return
	}
	_ = storage.exports.Close()
	_ = storage.temporary.Close()
}

func (storage *RootStorage) Stage(ctx context.Context, source io.Reader, maximum int64) (incidentbundles.StagedBundle, error) {
	if storage == nil || storage.temporary == nil || source == nil {
		return incidentbundles.StagedBundle{}, errors.New("incident bundle temporary storage is unavailable")
	}
	reference, err := incidentbundles.ParseBundleStagingRef("incident-bundles/imports/" + uuid.NewString() + ".bundle")
	if err != nil {
		return incidentbundles.StagedBundle{}, err
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return incidentbundles.StagedBundle{}, err
	}
	if err := storage.temporary.MakePrivateDir(rootedfs.MustParseReference("incident-bundles/imports")); err != nil {
		return incidentbundles.StagedBundle{}, err
	}
	content, err := writeArchive(ctx, storage.temporary, rootReference, maximum, func(destination io.Writer) error {
		// LimitReader prevents an untrusted WriterTo implementation from bypassing
		// the fixed copy buffer and supplies one byte to distinguish exact EOF.
		_, err := io.CopyBuffer(destination, io.LimitReader(source, maximum+1), make([]byte, 32<<10))
		return err
	})
	if err != nil {
		return incidentbundles.StagedBundle{}, err
	}
	return incidentbundles.StagedBundle{BundleStagingRef: reference, ArchiveContent: content}, nil
}

func (storage *RootStorage) Publish(ctx context.Context, bundleID string, maximum int64, write func(io.Writer) error) (incidentbundles.PublishedBundle, error) {
	if storage == nil || storage.exports == nil {
		return incidentbundles.PublishedBundle{}, errors.New("incident bundle export storage is unavailable")
	}
	parsed, err := uuid.Parse(bundleID)
	if err != nil {
		return incidentbundles.PublishedBundle{}, errors.New("incident bundle export identifier is invalid")
	}
	reference, err := incidentbundles.ParseBundleStorageRef("incident-bundles/" + parsed.String() + ".zip")
	if err != nil {
		return incidentbundles.PublishedBundle{}, err
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return incidentbundles.PublishedBundle{}, err
	}
	if err := storage.exports.MakePrivateDir(rootedfs.MustParseReference("incident-bundles")); err != nil {
		return incidentbundles.PublishedBundle{}, err
	}
	content, err := writeArchive(ctx, storage.exports, rootReference, maximum, write)
	if err != nil {
		return incidentbundles.PublishedBundle{}, err
	}
	return incidentbundles.PublishedBundle{BundleStorageRef: reference, ArchiveContent: content}, nil
}

func (storage *RootStorage) OpenStaged(reference incidentbundles.BundleStagingRef, maximum int64) (incidentbundles.ArchiveSource, int64, error) {
	if storage == nil || storage.temporary == nil {
		return nil, 0, errors.New("incident bundle temporary storage is unavailable")
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return nil, 0, err
	}
	source, metadata, err := storage.temporary.OpenRegular(rootReference)
	if err != nil {
		return nil, 0, err
	}
	if maximum < 1 || metadata.Size > maximum {
		_ = source.Close()
		return nil, 0, incidentbundles.ErrArchiveBytesExceeded
	}
	return source, metadata.Size, nil
}

// The admitted root owns creation, fsync, exclusive publication and removal on
// any writer error. Hashes and counts describe acknowledged bytes only.
func writeArchive(ctx context.Context, root *rootedfs.Root, reference rootedfs.Reference, maximum int64, write func(io.Writer) error) (incidentbundles.ArchiveContent, error) {
	if maximum < 1 || maximum == int64(1<<63-1) || write == nil {
		return incidentbundles.ArchiveContent{}, errors.New("incident bundle archive write policy is invalid")
	}
	digest := sha256.New()
	var content incidentbundles.ArchiveContent
	err := root.CreateExclusive(ctx, reference, func(destination io.Writer) error {
		bounded := &archiveWriter{ctx: ctx, destination: io.MultiWriter(destination, digest), maximum: maximum}
		if err := write(bounded); err != nil {
			return err
		}
		if bounded.failure != nil {
			return bounded.failure
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		content.ByteSize = bounded.written
		content.SHA256 = hex.EncodeToString(digest.Sum(nil))
		return nil
	})
	if err != nil {
		return incidentbundles.ArchiveContent{}, err
	}
	return content, nil
}

type archiveWriter struct {
	ctx         context.Context
	destination io.Writer
	maximum     int64
	written     int64
	failure     error
}

func (w *archiveWriter) Write(p []byte) (int, error) {
	if w.failure != nil {
		return 0, w.failure
	}
	if err := w.ctx.Err(); err != nil {
		w.failure = err
		return 0, err
	}
	if int64(len(p)) > w.maximum-w.written {
		w.failure = incidentbundles.ErrArchiveBytesExceeded
		return 0, w.failure
	}
	n, err := w.destination.Write(p)
	w.written += int64(n)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	w.failure = err
	return n, err
}

func (storage *RootStorage) RemoveStaged(reference incidentbundles.BundleStagingRef) error {
	if storage == nil || storage.temporary == nil {
		return errors.New("incident bundle temporary storage is unavailable")
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return err
	}
	return removeIncidentBundleRegular(storage.temporary, rootReference)
}

func (storage *RootStorage) RemovePublished(reference incidentbundles.BundleStorageRef) error {
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

// OpenPublished verifies the complete immutable export before exposing a public
// stream. The same confined file descriptor is rewound; no host path escapes.
func (storage *RootStorage) OpenPublished(ctx context.Context, reference incidentbundles.BundleStorageRef, expected incidentbundles.ArchiveContent) (io.ReadCloser, error) {
	if storage == nil || storage.exports == nil || expected.ByteSize < 1 || !validArchiveDigest(expected.SHA256) {
		return nil, errors.New("incident bundle export binding is invalid")
	}
	rootReference, err := rootedfs.ParseReference(reference.String())
	if err != nil {
		return nil, err
	}
	source, metadata, err := storage.exports.OpenRegular(rootReference)
	if err != nil {
		return nil, err
	}
	retained := false
	defer func() {
		if !retained {
			_ = source.Close()
		}
	}()
	if metadata.Size != expected.ByteSize {
		return nil, errors.New("incident bundle export size changed")
	}
	digest := sha256.New()
	reader := &publishedArchiveRead{ctx: ctx, source: source}
	if _, err := io.CopyBuffer(digest, reader, make([]byte, 32<<10)); err != nil {
		return nil, err
	}
	if hex.EncodeToString(digest.Sum(nil)) != expected.SHA256 {
		return nil, errors.New("incident bundle export digest changed")
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	retained = true
	return reader, nil
}

type publishedArchiveRead struct {
	ctx    context.Context
	source rootedfs.ReadSeekCloser
}

func (r *publishedArchiveRead) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.source.Read(p)
}
func (r *publishedArchiveRead) Close() error { return r.source.Close() }

func validArchiveDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && hex.EncodeToString(decoded) == value
}
