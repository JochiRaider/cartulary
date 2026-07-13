package secretpurpose

import (
	"crypto/sha256"
	"errors"
	"sync"
)

var ErrPurposeReuse = errors.New("startup secret purpose reuse")

// Registry records only one-way material digests and purpose labels. It never
// retains or reports resolved secret bytes or their fingerprints.
type Registry struct {
	mu       sync.Mutex
	refs     map[string]string
	material map[[32]byte]string
}

func NewRegistry() *Registry {
	return &Registry{refs: make(map[string]string), material: make(map[[32]byte]string)}
}

func (r *Registry) Register(referenceName string, purpose string, resolvedMaterial []byte) error {
	if r == nil || referenceName == "" || purpose == "" || len(resolvedMaterial) == 0 {
		return errors.New("startup secret registration is incomplete")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if previous, exists := r.refs[referenceName]; exists && previous != purpose {
		return ErrPurposeReuse
	}
	fingerprint := sha256.Sum256(resolvedMaterial)
	if previous, exists := r.material[fingerprint]; exists && previous != purpose {
		return ErrPurposeReuse
	}
	r.refs[referenceName] = purpose
	r.material[fingerprint] = purpose
	return nil
}
