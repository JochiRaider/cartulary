package referencetest

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"testing"
	"time"
)

// Lifetime supplies an explicit policy to real-storage compositions.
func Lifetime(t testing.TB) *reference_data.VerificationLifetime {
	t.Helper()
	lifetime, err := reference_data.NewVerificationLifetime(time.Second, func(err error) { t.Errorf("unexpected verification fatal: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	return lifetime
}
