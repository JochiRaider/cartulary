package packformat

import (
	"context"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"slices"
	"time"
)

type fixtureRole struct{ keys map[string]bool }
type fixtureMetadata struct {
	version     int64
	expires     time.Time
	signedBytes []byte
	roles       map[string]fixtureRole
}

func fixtureMetadataHeader(data []byte, role string) (fixtureMetadata, error) {
	header, err := envelope.ReadMetadataHeader(data, role)
	if err != nil {
		return fixtureMetadata{}, formatError(err)
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return fixtureMetadata{}, err
	}
	signed := value.(map[string]any)["signed"].(map[string]any)
	raw, err := canonicaljson.Marshal(signed)
	if err != nil {
		return fixtureMetadata{}, err
	}
	result := fixtureMetadata{version: header.Version, expires: header.Expires, signedBytes: raw, roles: map[string]fixtureRole{}}
	if role == "root" {
		for name, value := range signed["roles"].(map[string]any) {
			keys := map[string]bool{}
			for _, id := range value.(map[string]any)["keyids"].([]any) {
				keys[id.(string)] = true
			}
			result.roles[name] = fixtureRole{keys: keys}
		}
	}
	return result, nil
}
func fixtureMetadataFindings(data []byte, kind string, _ int, path string, emit FindingSink) (map[string]any, error) {
	return nil, formatError(envelope.MetadataFindings(data, kind, path, emit))
}
func sortedTrustKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
func fixtureTrustSchemaFindings(ctx context.Context, files map[string][]byte, emit FindingSink) error {
	p, _ := TrustChecks(files, nil, "", time.Time{}, TrustSnapshot{})
	return p["tuf_schema"](ctx, emit)
}
