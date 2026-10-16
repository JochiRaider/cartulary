package performancefixturelifecycle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

const (
	objectSnapshotSchema           = "cartulary.performance_fixture_objects.v1"
	objectSnapshotFile             = "object-snapshot.json"
	objectSnapshotDirectory        = "objects"
	maxSnapshotManifestBytes       = 16 << 20
	maxSnapshotObjectCount         = 100000
	maxSnapshotObjectBytes   int64 = 1 << 30
	maxSnapshotTotalBytes    int64 = 128 << 30
)

// ObjectSnapshotSummary is the retained, non-capability-bearing proof of the
// private object manifest. Storage keys and payloads remain suite-private.
type ObjectSnapshotSummary struct {
	ManifestSHA256 string `json:"manifest_sha256"`
	ObjectCount    int    `json:"object_count"`
	TotalBytes     int64  `json:"total_bytes"`
}

// UnmarshalJSON preserves required zero-valued fields: omission is not an
// assertion that the sealed population was empty.
func (summary *ObjectSnapshotSummary) UnmarshalJSON(raw []byte) error {
	if _, err := canonicaljson.DecodeStrict(raw); err != nil {
		return errors.New("object snapshot proof JSON is ambiguous")
	}
	var wire struct {
		ManifestSHA256 *string `json:"manifest_sha256"`
		ObjectCount    *int    `json:"object_count"`
		TotalBytes     *int64  `json:"total_bytes"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return errors.New("object snapshot proof shape is invalid")
	}
	if wire.ManifestSHA256 == nil || wire.ObjectCount == nil || wire.TotalBytes == nil {
		return errors.New("object snapshot proof is incomplete")
	}
	candidate := ObjectSnapshotSummary{ManifestSHA256: *wire.ManifestSHA256, ObjectCount: *wire.ObjectCount, TotalBytes: *wire.TotalBytes}
	if !validObjectSnapshotSummary(candidate) {
		return errors.New("object snapshot proof exceeds its admitted bounds")
	}
	*summary = candidate
	return nil
}

type snapshotObject struct {
	Key         string `json:"key"`
	ByteSize    int64  `json:"byte_size"`
	ContentType string `json:"content_type"`
	SHA256      string `json:"sha256"`
}

type objectSnapshotManifest struct {
	SchemaID    string           `json:"schema_id"`
	SnapshotKey string           `json:"snapshot_key"`
	Objects     []snapshotObject `json:"objects"`
}

func sealObjectSnapshot(ctx context.Context, privateRoot, key string, source objectstore.Store) (ObjectSnapshotSummary, error) {
	objects, err := inspectSnapshotObjects(ctx, source)
	if err != nil {
		return ObjectSnapshotSummary{}, err
	}
	manifest := objectSnapshotManifest{SchemaID: objectSnapshotSchema, SnapshotKey: key, Objects: objects}
	raw, summary, err := encodeObjectSnapshot(manifest)
	if err != nil {
		return ObjectSnapshotSummary{}, err
	}
	root, err := rootedfs.Open(privateRoot)
	if err != nil {
		return ObjectSnapshotSummary{}, errors.New("object snapshot private root is unavailable or unsafe")
	}
	defer root.Close()
	if err := root.CreateExclusive(ctx, rootedfs.MustParseReference(objectSnapshotFile), func(writer io.Writer) error {
		_, err := writer.Write(raw)
		return err
	}); err != nil {
		return ObjectSnapshotSummary{}, errors.New("object snapshot cannot be sealed exclusively")
	}
	return summary, nil
}

func readObjectSnapshot(privateRoot, key string, expected ObjectSnapshotSummary) (objectSnapshotManifest, error) {
	if !validObjectSnapshotSummary(expected) {
		return objectSnapshotManifest{}, errors.New("object snapshot expected proof is invalid")
	}
	root, err := rootedfs.Open(privateRoot)
	if err != nil {
		return objectSnapshotManifest{}, errors.New("object snapshot private root is unavailable or unsafe")
	}
	defer root.Close()
	raw, _, err := root.ReadRegular(rootedfs.MustParseReference(objectSnapshotFile), maxSnapshotManifestBytes)
	if err != nil {
		return objectSnapshotManifest{}, errors.New("object snapshot manifest is unavailable or unsafe")
	}
	var manifest objectSnapshotManifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return objectSnapshotManifest{}, errors.New("object snapshot manifest is invalid")
	}
	canonical, actual, err := encodeObjectSnapshot(manifest)
	// Canonical bytes also reject duplicate members and trailing JSON rather than
	// silently interpreting a different representation of the admitted manifest.
	if err != nil || manifest.SnapshotKey != key || !bytes.Equal(raw, canonical) || actual != expected {
		return objectSnapshotManifest{}, errors.New("object snapshot manifest does not match its sealed proof")
	}
	return manifest, nil
}

// restoreObjectSnapshot requires an empty destination. A failed transfer may
// leave partial objects; the caller owns bucket cleanup and must not expose the
// clone until this function and database preparation both succeed.
func restoreObjectSnapshot(ctx context.Context, privateRoot, key string, expected ObjectSnapshotSummary, destination objectstore.Store) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if destination == nil {
		return errors.New("object snapshot destination is required")
	}
	manifest, err := readObjectSnapshot(privateRoot, key, expected)
	if err != nil {
		return err
	}
	existing, err := destination.ListObjects(ctx, "")
	if err != nil || len(existing) != 0 {
		return errors.New("object snapshot destination must be empty and available")
	}
	objectRoot := filepath.Join(privateRoot, objectSnapshotDirectory)
	// NewFilesystemStore may create a missing directory. Admission first requires
	// the sealed source to exist and rejects symlinked path components.
	guard, err := rootedfs.Open(objectRoot)
	if err != nil {
		return errors.New("object snapshot payload root is unavailable or unsafe")
	}
	defer guard.Close()
	source, err := objectstore.NewFilesystemStore(objectRoot)
	if err != nil {
		return errors.New("object snapshot payload store is unavailable")
	}
	defer source.Close()
	entries, err := source.ListObjects(ctx, "")
	if err != nil {
		return errors.New("object snapshot payload inventory is unavailable")
	}
	if err := matchSnapshotInventory(entries, manifest.Objects); err != nil {
		return err
	}
	for _, object := range manifest.Objects {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := guard.Check(); err != nil {
			return errors.New("object snapshot payload root changed")
		}
		body, info, err := source.ReadObject(ctx, object.Key, objectstore.ReadOptions{})
		if err != nil || body == nil {
			return errors.New("object snapshot payload is unavailable or unsafe")
		}
		if info.Key != object.Key || info.Size != object.ByteSize || info.ContentType != object.ContentType {
			_ = body.Close()
			return errors.New("object snapshot payload metadata changed")
		}
		hash := sha256.New()
		reader := &snapshotCountingReader{reader: io.TeeReader(io.LimitReader(body, object.ByteSize+1), hash), ctx: ctx}
		putErr := destination.PutObject(ctx, object.Key, reader, object.ByteSize, object.ContentType)
		closeErr := body.Close()
		if putErr != nil || closeErr != nil || reader.count != object.ByteSize || "sha256:"+hex.EncodeToString(hash.Sum(nil)) != object.SHA256 {
			return errors.New("object snapshot payload transfer failed integrity validation")
		}
	}
	// Verify stored bytes, not merely the stream submitted to the adapter.
	actual, err := inspectSnapshotObjects(ctx, destination)
	if err != nil || !slices.Equal(actual, manifest.Objects) {
		return errors.New("object snapshot restored content differs from its sealed source")
	}
	return nil
}

type snapshotCountingReader struct {
	reader io.Reader
	ctx    context.Context
	count  int64
}

func (reader *snapshotCountingReader) Read(buffer []byte) (int, error) {
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := reader.reader.Read(buffer)
	reader.count += int64(n)
	return n, err
}

func inspectSnapshotObjects(ctx context.Context, store objectstore.Store) ([]snapshotObject, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("object snapshot store is required")
	}
	entries, err := store.ListObjects(ctx, "")
	if err != nil {
		return nil, errors.New("object snapshot inventory is unavailable")
	}
	slices.SortFunc(entries, func(a, b objectstore.ObjectInfo) int { return strings.Compare(a.Key, b.Key) })
	objects := make([]snapshotObject, len(entries))
	for i, entry := range entries {
		objects[i] = snapshotObject{Key: entry.Key, ByteSize: entry.Size, ContentType: entry.ContentType, SHA256: "sha256:" + strings.Repeat("0", 64)}
	}
	if _, err := validateSnapshotObjects(objects); err != nil {
		return nil, err
	}
	for i, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		body, info, err := store.ReadObject(ctx, entry.Key, objectstore.ReadOptions{})
		if err != nil || body == nil {
			return nil, errors.New("object snapshot payload is unavailable")
		}
		// Listings establish membership and size; S3 does not include MIME.
		if info.Key != entry.Key || info.Size != entry.Size {
			_ = body.Close()
			return nil, errors.New("object snapshot inventory changed while sealing")
		}
		hash := sha256.New()
		reader := &snapshotCountingReader{reader: io.LimitReader(body, entry.Size+1), ctx: ctx}
		_, copyErr := io.Copy(hash, reader)
		closeErr := body.Close()
		if copyErr != nil || closeErr != nil || reader.count != entry.Size {
			return nil, errors.New("object snapshot payload length or read failed")
		}
		objects[i].ContentType = info.ContentType
		objects[i].SHA256 = "sha256:" + hex.EncodeToString(hash.Sum(nil))
	}
	if _, err := validateSnapshotObjects(objects); err != nil {
		return nil, err
	}
	return objects, nil
}

func matchSnapshotInventory(entries []objectstore.ObjectInfo, expected []snapshotObject) error {
	if len(entries) != len(expected) {
		return errors.New("object snapshot payload membership changed")
	}
	slices.SortFunc(entries, func(a, b objectstore.ObjectInfo) int { return strings.Compare(a.Key, b.Key) })
	for i, entry := range entries {
		if entry.Key != expected[i].Key || entry.Size != expected[i].ByteSize || entry.ContentType != expected[i].ContentType {
			return errors.New("object snapshot payload membership or metadata changed")
		}
	}
	return nil
}

func encodeObjectSnapshot(manifest objectSnapshotManifest) ([]byte, ObjectSnapshotSummary, error) {
	if manifest.SchemaID != objectSnapshotSchema || !lowerHexDigest(manifest.SnapshotKey) {
		return nil, ObjectSnapshotSummary{}, errors.New("object snapshot identity is invalid")
	}
	total, err := validateSnapshotObjects(manifest.Objects)
	if err != nil {
		return nil, ObjectSnapshotSummary{}, err
	}
	raw, err := canonicaljson.Marshal(manifest)
	if err != nil || len(raw)+1 > maxSnapshotManifestBytes {
		return nil, ObjectSnapshotSummary{}, errors.New("object snapshot manifest exceeds its bound")
	}
	raw = append(raw, '\n')
	digest := sha256.Sum256(raw)
	return raw, ObjectSnapshotSummary{ManifestSHA256: "sha256:" + hex.EncodeToString(digest[:]), ObjectCount: len(manifest.Objects), TotalBytes: total}, nil
}

func validObjectSnapshotSummary(summary ObjectSnapshotSummary) bool {
	return strings.HasPrefix(summary.ManifestSHA256, "sha256:") && lowerHexDigest(strings.TrimPrefix(summary.ManifestSHA256, "sha256:")) && summary.ObjectCount >= 0 && summary.ObjectCount <= maxSnapshotObjectCount && summary.TotalBytes >= 0 && summary.TotalBytes <= maxSnapshotTotalBytes && (summary.ObjectCount != 0 || summary.TotalBytes == 0)
}

func validateSnapshotObjects(objects []snapshotObject) (int64, error) {
	if objects == nil || len(objects) > maxSnapshotObjectCount {
		return 0, errors.New("object snapshot population is invalid")
	}
	keys := make([]string, len(objects))
	var total int64
	for i, object := range objects {
		if len(object.Key) > 1024 || strings.HasSuffix(object.Key, ".meta.json") || strings.HasSuffix(object.Key, ".tmp") || object.ByteSize < 0 || object.ByteSize > maxSnapshotObjectBytes || len(object.ContentType) > 255 || !utf8.ValidString(object.ContentType) || strings.ContainsAny(object.ContentType, "\r\n\x00") || !strings.HasPrefix(object.SHA256, "sha256:") || !lowerHexDigest(strings.TrimPrefix(object.SHA256, "sha256:")) || (i > 0 && objects[i-1].Key >= object.Key) {
			return 0, errors.New("object snapshot entry is invalid or unordered")
		}
		total += object.ByteSize
		if total > maxSnapshotTotalBytes {
			return 0, errors.New("object snapshot byte population exceeds its bound")
		}
		keys[i] = object.Key
	}
	if _, err := rootedfs.ValidateReferenceSet(keys); err != nil {
		return 0, errors.New("object snapshot storage references are invalid")
	}
	return total, nil
}
