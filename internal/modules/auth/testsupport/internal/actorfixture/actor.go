// Package actorfixture prepares test-only actor inputs. It is internal to Auth's
// test support so production consumers cannot use it as a credential workflow.
package actorfixture

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

// Attributes describes an independent actor, with an active ordinary user as
// the default. Authentication tests must use the password-taking fixture APIs.
type Attributes struct {
	// ID optionally fixes identity for ordering assertions; zero allocates a fresh UUID.
	ID              uuid.UUID
	Email           string
	DisplayName     string
	MFARequired     bool
	DeploymentAdmin bool
	Disabled        bool
}

var preparedCredential = sync.OnceValues(func() (string, error) {
	var entropy [32]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		return "", err
	}
	return authn.HashPassword(context.Background(), base64.RawURLEncoding.EncodeToString(entropy[:]))
})

// Credential is immutable, lazy and shared only within this test process. The
// random password is never returned, retained in diagnostics or used to log in.
func Credential() (string, error) { return preparedCredential() }
