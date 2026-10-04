package reference_data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"slices"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/jackc/pgx/v5"
)

// A retained container is an input envelope, not permission to silently repair
// missing or altered logical content. Reverify reports loss of any member.
type retainedContentQuery interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (c *Coordinator) checkRetainedMembers(ctx context.Context, m frozenMember) error {
	return checkRetainedVersion(ctx, c.pool, c.storage, m.Key, m.Version)
}
func checkRetainedVersion(ctx context.Context, db retainedContentQuery, storage VerificationStorage, key, version string) error {
	for _, check := range []retainedCheck{retainedPresence, retainedLengths, retainedHashes} {
		if err := inspectRetainedVersion(ctx, db, storage, key, version, check); err != nil {
			return err
		}
	}
	return nil
}

type retainedCheck uint8

const (
	retainedPresence retainedCheck = iota
	retainedLengths
	retainedHashes
)

// Presence, length and digest are separate passes at their registry ranks.
// Earlier passes never read later-check content. Consumers use the same order
// as live verification, so traversal order cannot select the verdict.
func inspectRetainedVersion(ctx context.Context, db retainedContentQuery, storage VerificationStorage, key, version string, check retainedCheck) error {
	if check != retainedPresence && check != retainedLengths && check != retainedHashes {
		return errors.New("reference pack: unknown retained check")
	}
	var manifest []byte
	var distribution string
	if err := db.QueryRow(ctx, `SELECT manifest_bytes,distribution_kind FROM reference_pack_versions JOIN reference_pack_candidates USING(pack_key,pack_version) WHERE pack_key=$1 AND pack_version=$2`, key, version).Scan(&manifest, &distribution); err != nil {
		return err
	}
	decoded, err := packformat.DecodeManifest(manifest, distribution == "operator_imported")
	if err != nil {
		return err
	}
	expected := map[string]packformat.Member{"manifest.json": {SHA256: packformat.Digest(manifest), Size: int64(len(manifest))}}
	paths := []string{"manifest.json"}
	for _, file := range decoded.Files {
		expected[file.Path] = packformat.Member{SHA256: file.SHA256, Size: file.Size}
		paths = append(paths, file.Path)
	}
	slices.Sort(paths)
	type retainedObject struct {
		reference string
		digest    string
		size      int64
		available bool
	}
	objects := make(map[string]retainedObject, len(paths))
	rows, err := db.Query(ctx, `SELECT f.logical_path,o.storage_ref,o.sha256,o.size_bytes,o.available FROM reference_pack_object_refs f JOIN reference_pack_objects o USING(object_id) WHERE f.owner_kind='version' AND f.owner_id=$1 ORDER BY f.logical_path COLLATE "C"`, versionObjectID(key, version))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var path string
		var object retainedObject
		if err := rows.Scan(&path, &object.reference, &object.digest, &object.size, &object.available); err != nil {
			return err
		}
		if _, ok := expected[path]; !ok {
			return errors.New("reference pack: unexpected retained logical binding")
		}
		if _, duplicate := objects[path]; duplicate {
			return errors.New("reference pack: duplicate retained logical binding")
		}
		objects[path] = object
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()
	// The admitted manifest bounds this inventory to 67 logical objects. Keep
	// all findings at this boundary, with ordinals over trusted logical paths.
	// A storage execution failure still aborts even after a content finding.
	findings := map[string][]packformat.Finding{}
	report := func(id string, index int) {
		findings[id] = append(findings[id], packformat.Finding{Path: "$.objects[" + strconv.Itoa(index) + "]"})
	}
	for index, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		wanted := expected[path]
		object, exists := objects[path]
		if !exists || !object.available {
			report("retained_payload", index)
			continue
		}
		ref, err := ParseStorageRef(object.reference)
		if err != nil {
			return err
		}
		reader, actual, err := storage.OpenPublished(ctx, ref)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if errors.Is(err, fs.ErrNotExist) || errors.Is(err, ErrArtifactUnavailable) {
				report("retained_payload", index)
				continue
			}
			return err
		}
		if check != retainedHashes {
			if err := reader.Close(); err != nil {
				return err
			}
			if check == retainedLengths && (actual != wanted.Size || object.size != wanted.Size) {
				report("member_length", index)
			}
			continue
		}
		hash := sha256.New()
		_, readErr := io.CopyBuffer(hash, &attemptReader{ctx: ctx, source: io.NewSectionReader(reader, 0, wanted.Size)}, make([]byte, 32768))
		closeErr := reader.Close()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if readErr != nil || closeErr != nil {
			return errors.Join(readErr, closeErr)
		}
		if actual != object.size || object.size != wanted.Size || object.digest != wanted.SHA256 || hex.EncodeToString(hash.Sum(nil)) != wanted.SHA256 {
			report("member_hash", index)
		}
	}
	for _, id := range []string{"retained_payload", "member_length", "member_hash"} {
		if len(findings[id]) == 0 {
			continue
		}
		summary, err := packformat.CheckSummary(ctx, id, nil, func(_ context.Context, emit packformat.FindingSink) error {
			for _, finding := range findings[id] {
				if err := emit(finding); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		return &ContentRejection{Code: summary.Issues[0].Code, CheckID: id, CandidateKey: key, CandidateVersion: version, Summary: summary}
	}
	return nil
}
