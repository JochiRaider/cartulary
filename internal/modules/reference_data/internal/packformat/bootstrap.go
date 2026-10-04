package packformat

import (
	"bytes"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// DecodeBootstrap admits the complete deployment file before any repository is
// installed. The caller retains these roots only for previously unseen IDs;
// an existing trust repository can advance only through signed root rotation.
func DecodeBootstrap(data []byte) (map[string]TrustSnapshot, error) {
	invalid := errors.New("reference pack: invalid trust bootstrap configuration")
	if len(data) == 0 || len(data) > 8388608 {
		return nil, invalid
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return nil, invalid
	}
	m, ok := exact(value, "schema_id", "repositories")
	if !ok || m["schema_id"] != "cartulary.reference_pack_trust_bootstrap.v1" {
		return nil, invalid
	}
	canonical, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(data, canonical) {
		return nil, invalid
	}
	items, ok := m["repositories"].([]any)
	if !ok || len(items) < 1 || len(items) > 64 {
		return nil, invalid
	}
	out := make(map[string]TrustSnapshot, len(items))
	previous := ""
	for _, item := range items {
		repository, ok := exact(item, "repository_id", "trusted_root", "trusted_root_sha256")
		if !ok {
			return nil, invalid
		}
		id, ok := repository["repository_id"].(string)
		if !ok || len(id) > 128 || !keyPattern.MatchString(id) || id <= previous {
			return nil, invalid
		}
		rootBytes, err := canonicaljson.Marshal(repository["trusted_root"])
		if err != nil || !equalDigest(rootBytes, repository["trusted_root_sha256"]) {
			return nil, invalid
		}
		root, err := decodeRoot(rootBytes)
		if err != nil || root.signed["cartulary"].(map[string]any)["trust_repository_id"] != id {
			return nil, invalid
		}
		snapshot, err := AdmitBootstrap(rootBytes)
		if err != nil {
			return nil, invalid
		}
		out[id] = snapshot
		previous = id
	}
	return out, nil
}

func RootVersion(data []byte) (int64, error) {
	root, err := decodeRoot(data)
	return root.version, err
}
