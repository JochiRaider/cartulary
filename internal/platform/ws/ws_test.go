package ws

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHubSessionRevocationSubscribers(t *testing.T) {
	t.Run("revoke notifies each registered listener exactly once", func(t *testing.T) {
		hub := NewHub()
		sessionID := uuid.New()

		first, unregisterFirst := hub.RegisterSession(sessionID)
		defer unregisterFirst()
		second, unregisterSecond := hub.RegisterSession(sessionID)
		defer unregisterSecond()

		hub.RevokeSession(sessionID, "session_revoked")

		requireRevocationReason(t, first, "session_revoked")
		requireRevocationReason(t, second, "session_revoked")
		requireNoRevocationReason(t, first)
		requireNoRevocationReason(t, second)

		hub.RevokeSession(sessionID, "ignored")
		requireNoRevocationReason(t, first)
		requireNoRevocationReason(t, second)
	})

	t.Run("unregister prevents later delivery", func(t *testing.T) {
		hub := NewHub()
		sessionID := uuid.New()

		revocations, unregister := hub.RegisterSession(sessionID)
		unregister()

		other, unregisterOther := hub.RegisterSession(sessionID)
		defer unregisterOther()

		hub.RevokeSession(sessionID, "session_revoked")

		requireNoRevocationReason(t, revocations)
		requireRevocationReason(t, other, "session_revoked")
	})
}

func requireRevocationReason(t testing.TB, revocations <-chan string, want string) {
	t.Helper()

	select {
	case got := <-revocations:
		if got != want {
			t.Fatalf("unexpected revocation reason: got %q want %q", got, want)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for revocation reason %q", want)
	}
}

func requireNoRevocationReason(t testing.TB, revocations <-chan string) {
	t.Helper()

	select {
	case got := <-revocations:
		t.Fatalf("unexpected revocation reason: got %q", got)
	default:
	}
}
