package publication

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"os"
	"path/filepath"
)

var signingRoles = []string{"root-1", "root-2", "root-3", "targets", "snapshot", "timestamp"}

// Signer exposes public keys and signatures, never private key material. The
// binding is checked before every envelope. Remote/HSM adapters can implement
// this port without entering the consumer dependency closure.
type Signer interface {
	Scope() Scope
	PublicKeys(context.Context) (map[string]ed25519.PublicKey, error)
	Sign(context.Context, string, []byte) ([]byte, error)
}
type privateContext struct {
	SchemaID   string            `json:"schema_id"`
	Repository string            `json:"trust_repository_id"`
	PackKey    string            `json:"pack_key"`
	Class      AuthorityClass    `json:"authority_class"`
	Seeds      map[string]string `json:"private_seeds"`
}
type fileSigner struct {
	scope Scope
	keys  map[string]ed25519.PrivateKey
}

func (s *fileSigner) Scope() Scope { return s.scope }
func (s *fileSigner) PublicKeys(ctx context.Context) (map[string]ed25519.PublicKey, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	keys := map[string]ed25519.PublicKey{}
	for role, key := range s.keys {
		keys[role] = append(ed25519.PublicKey(nil), key.Public().(ed25519.PublicKey)...)
	}
	return keys, nil
}
func (s *fileSigner) Sign(ctx context.Context, role string, message []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	key, ok := s.keys[role]
	if !ok {
		return nil, errors.New("unknown signing role")
	}
	return ed25519.Sign(key, message), nil
}

// LoadFileSigner confines a private signing context outside every supplied
// public root. Creation is explicit and uses immutable no-replace retention.
func LoadFileSigner(name string, create bool, scope Scope, publicRoots ...string) (Signer, error) {
	if !scope.valid() || len(publicRoots) == 0 {
		return nil, errors.New("explicit signer scope and public roots required")
	}
	for _, root := range publicRoots {
		if root == "" {
			return nil, errors.New("empty public root")
		}
		if err := signingOutside(name, root); err != nil {
			return nil, err
		}
	}
	raw, err := ReadRegular(name, 65536)
	if errors.Is(err, os.ErrNotExist) && create {
		private := privateContext{"cartulary.reference_pack_signing_context.v1", scope.repository, scope.key, scope.class, map[string]string{}}
		for _, role := range signingRoles {
			seed := make([]byte, ed25519.SeedSize)
			if _, err = rand.Read(seed); err != nil {
				return nil, err
			}
			private.Seeds[role] = hex.EncodeToString(seed)
		}
		raw, err = canonicaljson.Marshal(private)
		if err != nil {
			return nil, err
		}
		if err = os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			return nil, err
		}
		root, err := os.OpenRoot(filepath.Dir(name))
		if err != nil {
			return nil, err
		}
		err = Retain(root, filepath.Base(name), raw, 0600)
		closeErr := root.Close()
		if err = errors.Join(err, closeErr); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("private signing context requires confined permissions")
	}
	var private privateContext
	if err = DecodeClosed(raw, &private); err != nil {
		return nil, err
	}
	if private.SchemaID != "cartulary.reference_pack_signing_context.v1" || private.Repository != scope.repository || private.PackKey != scope.key || private.Class != scope.class || len(private.Seeds) != len(signingRoles) {
		return nil, errors.New("private signing authority does not match publisher")
	}
	signer := &fileSigner{scope: scope, keys: map[string]ed25519.PrivateKey{}}
	seen := map[string]bool{}
	for _, role := range signingRoles {
		seed, e := hex.DecodeString(private.Seeds[role])
		if e != nil || len(seed) != ed25519.SeedSize {
			return nil, errors.New("invalid private signing key")
		}
		key := ed25519.NewKeyFromSeed(seed)
		identity := hex.EncodeToString(key.Public().(ed25519.PublicKey))
		if seen[identity] {
			return nil, errors.New("distinct signing keys required")
		}
		seen[identity] = true
		signer.keys[role] = key
	}
	return signer, nil
}
