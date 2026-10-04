package packformat

import (
	"bytes"
	"encoding/base64"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var preparedShape = compileProjection("prepared.v1.schema.json")

func admittedPreparedSize(size int) bool { return size > 0 && size <= 71303168 }

// ValidatePrepared admits a complete durable preparation, not a successful
// public version. Publication must separately prove object ownership, index
// completion, the execution lease and every captured revision.
func ValidatePrepared(data []byte) error {
	invalid := errors.New("reference pack: invalid prepared result")
	// 64 MiB envelope, base64 of a 1 MiB manifest, and at most 68 bounded
	// object descriptors fit below 68 MiB without changing any pack limit.
	if !admittedPreparedSize(len(data)) {
		return invalid
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil || !preparedShape.matches(value) {
		return invalid
	}
	canonical, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(canonical, data) {
		return invalid
	}
	prepared := value.(map[string]any)
	envelope := prepared["envelope"].(map[string]any)
	encodedEnvelope, err := canonicaljson.Marshal(envelope)
	if err != nil || envelope["distribution_kind"] != "operator_imported" || ValidateSuccessfulEnvelope(encodedEnvelope) != nil {
		return invalid
	}
	encodedManifest := prepared["manifest"].(string)
	manifestBytes, err := base64.StdEncoding.Strict().DecodeString(encodedManifest)
	if err != nil || len(manifestBytes) > 1048576 || base64.StdEncoding.EncodeToString(manifestBytes) != encodedManifest {
		return invalid
	}
	manifest, err := DecodeManifest(manifestBytes, true)
	if err != nil {
		return invalid
	}
	payload, err := PayloadDigest(manifest.Files)
	manifestDigest := Digest(manifestBytes)
	if err != nil || manifest.Key != envelope["pack_key"] || manifest.Version != envelope["pack_version"] ||
		manifestDigest != envelope["manifest_sha256"] || payload != envelope["payload_sha256"] {
		return invalid
	}
	wanted := map[string]Member{"manifest.json": {Size: int64(len(manifestBytes)), SHA256: manifestDigest}}
	for _, file := range manifest.Files {
		wanted[file.Path] = Member{Size: file.Size, SHA256: file.SHA256}
	}
	objects := prepared["objects"].([]any)
	if len(objects) != len(wanted)+1 {
		return invalid
	}
	type objectBinding struct {
		digest, reference string
		size              int64
	}
	byID := map[string]objectBinding{}
	byReference := map[string]string{}
	previous := ""
	container := false
	for _, value := range objects {
		object := value.(map[string]any)
		id, path := object["id"].(string), object["path"].(string)
		reference := object["reference"].(string)
		size, ok := integer(object["size"], 0, 9007199254740991)
		digest := object["sha256"].(string)
		if !ok || path <= previous || len(reference) > 4096 {
			return invalid
		}
		previous = path
		binding := objectBinding{digest, reference, size}
		if prior, exists := byID[id]; exists && prior != binding {
			return invalid
		}
		if prior, exists := byReference[reference]; exists && prior != id {
			return invalid
		}
		byID[id], byReference[reference] = binding, id
		if path == "container" {
			if size == 0 || digest != envelope["container_sha256"] || reference != envelope["container_ref"] {
				return invalid
			}
			container = true
			continue
		}
		member, exists := wanted[path]
		if !exists || member.Size != size || member.SHA256 != digest {
			return invalid
		}
		delete(wanted, path)
	}
	if !container || len(wanted) != 0 {
		return invalid
	}
	return nil
}
