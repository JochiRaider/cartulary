package publication

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"slices"
)

// BuildFixtureEnvelope signs explicitly scoped fixture bytes in memory. It
// deliberately does not validate their content: consumer tests need authentic
// envelopes around malformed payloads. Production publication must use
// PublishRelease, which checks approval, admission, reservations and retention.
func BuildFixtureEnvelope(ctx context.Context, members map[string][]byte, signer Signer, rootExpires, onlineExpires string, version int64) ([]byte, []byte, error) {
	if signer == nil || signer.Scope().Class() != FixtureAuthority {
		return nil, nil, errors.New("fixture authority required")
	}
	return buildEnvelope(ctx, members, signer, signer.Scope(), rootExpires, onlineExpires, version)
}

func buildEnvelope(ctx context.Context, members map[string][]byte, signer Signer, scope Scope, rootExpires, onlineExpires string, version int64) ([]byte, []byte, error) {
	manifest, err := envelope.DecodeManifest(members["manifest.json"], true)
	if err != nil {
		return nil, nil, err
	}
	if !scope.valid() || signer == nil || signer.Scope() != scope || manifest.Repository == nil || *manifest.Repository != scope.repository || manifest.Key != scope.key {
		return nil, nil, errors.New("matching publisher and pack identity required")
	}
	payload, err := envelope.PayloadDigest(manifest.Files)
	if err != nil {
		return nil, nil, err
	}
	keys := map[string]any{}
	ids := map[string]string{}
	publicKeys, err := signer.PublicKeys(ctx)
	if err != nil {
		return nil, nil, err
	}
	if len(publicKeys) != len(signingRoles) {
		return nil, nil, errors.New("complete signing roles required")
	}
	for _, role := range signingRoles {
		key := publicKeys[role]
		if len(key) != ed25519.PublicKeySize {
			return nil, nil, errors.New("invalid public signing key")
		}
		public := map[string]any{"keytype": "ed25519", "scheme": "ed25519", "keyval": map[string]any{"public": hex.EncodeToString(key)}}
		b, err := canonicaljson.Marshal(public)
		if err != nil {
			return nil, nil, err
		}
		id := Digest(b)
		if _, exists := keys[id]; exists {
			return nil, nil, errors.New("distinct signing keys required")
		}
		keys[id] = public
		ids[role] = id
	}
	roleByID := map[string]string{}
	for role, id := range ids {
		roleByID[id] = role
	}
	sign := func(signed map[string]any, signers []string) ([]byte, error) {
		b, err := canonicaljson.Marshal(signed)
		if err != nil {
			return nil, err
		}
		slices.Sort(signers)
		sigs := []any{}
		for _, id := range signers {
			signature, err := signer.Sign(ctx, roleByID[id], b)
			if err != nil {
				return nil, err
			}
			if len(signature) != ed25519.SignatureSize || !ed25519.Verify(publicKeys[roleByID[id]], b, signature) {
				return nil, errors.New("signer returned invalid signature")
			}
			sigs = append(sigs, map[string]any{"keyid": id, "sig": hex.EncodeToString(signature)})
		}
		return canonicaljson.Marshal(map[string]any{"signed": signed, "signatures": sigs})
	}
	rootIDs := []string{ids["root-1"], ids["root-2"], ids["root-3"]}
	slices.Sort(rootIDs)
	roles := map[string]any{"root": map[string]any{"keyids": rootIDs, "threshold": 2}}
	for _, role := range []string{"targets", "snapshot", "timestamp"} {
		roles[role] = map[string]any{"keyids": []string{ids[role]}, "threshold": 1}
	}
	root, err := sign(map[string]any{"_type": "root", "spec_version": "1.0.35", "version": 1, "expires": rootExpires, "consistent_snapshot": false, "keys": keys, "roles": roles, "cartulary": map[string]any{"schema_id": "cartulary.reference_pack_tuf_root_binding.v1", "trust_repository_id": scope.repository}}, []string{ids["root-1"], ids["root-2"]})
	if err != nil {
		return nil, nil, err
	}
	if len(members) != len(manifest.Files)+1 {
		return nil, nil, errors.New("unexpected logical member")
	}
	files := map[string][]byte{"manifest.json": members["manifest.json"]}
	for _, f := range manifest.Files {
		b, ok := members[f.Path]
		if !ok || int64(len(b)) != f.Size || Digest(b) != f.SHA256 {
			return nil, nil, errors.New("logical member integrity mismatch")
		}
		files[f.Path] = b
	}
	files["bundle.json"], err = canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_bundle_hint.v1", "trust_repository_id": scope.repository})
	if err != nil {
		return nil, nil, err
	}
	descriptor := func(b []byte) map[string]any {
		return map[string]any{"length": len(b), "hashes": map[string]any{"sha256": Digest(b)}}
	}
	targets := map[string]any{}
	for name, b := range files {
		targets[name] = descriptor(b)
	}
	binding := map[string]any{"schema_id": "cartulary.reference_pack_tuf_targets_binding.v1", "trust_repository_id": scope.repository, "pack_key": manifest.Key, "pack_version": manifest.Version, "pack_release_sequence": manifest.Sequence, "manifest_sha256": Digest(files["manifest.json"]), "payload_sha256": payload}
	files["metadata/targets.json"], err = sign(map[string]any{"_type": "targets", "spec_version": "1.0.35", "version": version, "expires": onlineExpires, "targets": targets, "cartulary": binding}, []string{ids["targets"]})
	if err != nil {
		return nil, nil, err
	}
	for _, role := range []string{"snapshot", "timestamp"} {
		child := "targets.json"
		if role == "timestamp" {
			child = "snapshot.json"
		}
		d := descriptor(files["metadata/"+child])
		d["version"] = version
		files["metadata/"+role+".json"], err = sign(map[string]any{"_type": role, "spec_version": "1.0.35", "version": version, "expires": onlineExpires, "meta": map[string]any{child: d}}, []string{ids[role]})
		if err != nil {
			return nil, nil, err
		}
	}
	bootstrap, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_trust_bootstrap.v1", "repositories": []any{map[string]any{"repository_id": scope.repository, "trusted_root": json.RawMessage(root), "trusted_root_sha256": Digest(root)}}})
	if err != nil {
		return nil, nil, err
	}
	if _, err := envelope.DecodeBootstrap(bootstrap); err != nil {
		return nil, nil, err
	}
	var container bytes.Buffer
	if err = envelope.WriteUSTAR(ctx, &container, files); err != nil {
		return nil, nil, err
	}
	return container.Bytes(), bootstrap, nil
}
