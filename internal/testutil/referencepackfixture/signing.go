// Package referencepackfixture provides test-only signing for canonical pack
// artifacts. It never acquires or interprets source datasets. Private keys stay
// in memory and cannot authorize a production repository.
package referencepackfixture

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"

	"github.com/JochiRaider/cartulary/referencepacks/format/publication"
)

type SigningContext struct {
	scope publication.Scope
	keys  map[string]ed25519.PrivateKey
}

func NewSigningContext(repository, packKey string) (*SigningContext, error) {
	scope, err := publication.FixtureScope(repository, packKey)
	if err != nil {
		return nil, err
	}
	s := &SigningContext{scope: scope, keys: map[string]ed25519.PrivateKey{}}
	for _, role := range []string{"root-1", "root-2", "root-3", "targets", "snapshot", "timestamp"} {
		_, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		s.keys[role] = key
	}
	return s, nil
}

func (s *SigningContext) Scope() publication.Scope { return s.scope }
func (s *SigningContext) Repository() string       { return s.scope.Repository() }
func (s *SigningContext) PublicKeys(ctx context.Context) (map[string]ed25519.PublicKey, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	keys := map[string]ed25519.PublicKey{}
	for role, key := range s.keys {
		keys[role] = append(ed25519.PublicKey(nil), key.Public().(ed25519.PublicKey)...)
	}
	return keys, nil
}
func (s *SigningContext) Sign(ctx context.Context, role string, message []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	key, ok := s.keys[role]
	if !ok {
		return nil, errors.New("unknown fixture signing role")
	}
	return ed25519.Sign(key, message), nil
}

func Envelope(ctx context.Context, members map[string][]byte, signer *SigningContext, rootExpires, onlineExpires string, version int64) ([]byte, []byte, error) {
	if signer == nil {
		return nil, nil, errors.New("fixture signer required")
	}
	return publication.BuildFixtureEnvelope(ctx, members, signer, rootExpires, onlineExpires, version)
}
