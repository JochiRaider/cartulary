package authn

import (
	"testing"
	"time"
)

func TestSupportPhase1_LoginNormalizationAndPasswordExactness(t *testing.T) {
	normalized, comparison, ok := NormalizeEmailAddress(" \u00a0Analyst@Example.Test\t")
	if !ok {
		t.Fatal("expected email normalization to succeed")
	}
	if normalized != "Analyst@Example.Test" {
		t.Fatalf("unexpected normalized email: got %q", normalized)
	}
	if comparison != "analyst@example.test" {
		t.Fatalf("unexpected email comparison key: got %q", comparison)
	}

	password := "  parol\u00e9 secret  "
	accepted, err := ValidatePasswordProvision(password)
	if err != nil {
		t.Fatalf("validate password: %v", err)
	}
	if accepted != password {
		t.Fatalf("password should remain exact after validation: got %q want %q", accepted, password)
	}

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if ok, err := VerifyPasswordHash(hash, password); err != nil || !ok {
		t.Fatalf("expected exact password match to verify, ok=%v err=%v", ok, err)
	}
	if ok, err := VerifyPasswordHash(hash, "parol\u00e9 secret"); err != nil {
		t.Fatalf("verify trimmed password: %v", err)
	} else if ok {
		t.Fatal("expected trimmed password to fail exact verification")
	}
}

func TestSupportPhase1_SessionTiming(t *testing.T) {
	now := time.Date(2026, time.April, 17, 12, 0, 0, 0, time.UTC)

	timing := NewSessionTiming(now)
	if !timing.AuthenticatedAt.Equal(now) {
		t.Fatalf("unexpected authenticated_at: got %s want %s", timing.AuthenticatedAt, now)
	}
	if !timing.LastQualifyingActivityAt.Equal(now) {
		t.Fatalf("unexpected last_qualifying_activity_at: got %s want %s", timing.LastQualifyingActivityAt, now)
	}
	if !timing.IdleExpiresAt.Equal(now.Add(30 * time.Minute)) {
		t.Fatalf("unexpected idle_expires_at: got %s", timing.IdleExpiresAt)
	}
	if !timing.AbsoluteExpiresAt.Equal(now.Add(12 * time.Hour)) {
		t.Fatalf("unexpected absolute_expires_at: got %s", timing.AbsoluteExpiresAt)
	}
	if !timing.SessionExpiresAt.Equal(now.Add(30 * time.Minute)) {
		t.Fatalf("unexpected session_expires_at: got %s", timing.SessionExpiresAt)
	}

	sliding := timing.Slide(now.Add(11*time.Hour + 45*time.Minute))
	if !sliding.IdleExpiresAt.Equal(now.Add(12*time.Hour + 15*time.Minute)) {
		t.Fatalf("unexpected slid idle_expires_at: got %s", sliding.IdleExpiresAt)
	}
	if !sliding.SessionExpiresAt.Equal(timing.AbsoluteExpiresAt) {
		t.Fatalf("session_expires_at must clamp to absolute expiry: got %s want %s", sliding.SessionExpiresAt, timing.AbsoluteExpiresAt)
	}
}
