package authn

import (
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestSupportPhase1_ConcurrencyLimitRevokesLRUNonCurrent(t *testing.T) {
	base := time.Date(2026, time.April, 17, 12, 0, 0, 0, time.UTC)

	currentID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	oldestID := uuid.MustParse("10000000-0000-0000-0000-000000000002")
	active := []SessionSummary{
		{SessionID: currentID, LastQualifyingActivityAt: base.Add(-10 * time.Minute), AuthenticatedAt: base.Add(-10 * time.Minute)},
		{SessionID: oldestID, LastQualifyingActivityAt: base.Add(-59 * time.Minute), AuthenticatedAt: base.Add(-59 * time.Minute)},
		{SessionID: uuid.MustParse("10000000-0000-0000-0000-000000000003"), LastQualifyingActivityAt: base.Add(-20 * time.Minute), AuthenticatedAt: base.Add(-20 * time.Minute)},
		{SessionID: uuid.MustParse("10000000-0000-0000-0000-000000000004"), LastQualifyingActivityAt: base.Add(-30 * time.Minute), AuthenticatedAt: base.Add(-30 * time.Minute)},
		{SessionID: uuid.MustParse("10000000-0000-0000-0000-000000000005"), LastQualifyingActivityAt: base.Add(-40 * time.Minute), AuthenticatedAt: base.Add(-40 * time.Minute)},
		{SessionID: uuid.MustParse("10000000-0000-0000-0000-000000000006"), LastQualifyingActivityAt: base.Add(-50 * time.Minute), AuthenticatedAt: base.Add(-50 * time.Minute)},
	}

	victim, ok := SelectSessionForConcurrencyLimit(active, currentID)
	if !ok {
		t.Fatal("expected one non-current victim session")
	}
	if victim.SessionID != oldestID {
		t.Fatalf("unexpected concurrency victim: got %s want %s", victim.SessionID, oldestID)
	}
	if ConcurrencyLimitReasonCode != "concurrency_limit" {
		t.Fatalf("unexpected concurrency reason code constant: %q", ConcurrencyLimitReasonCode)
	}
}

func TestSupportPhase1_RevocationScopes(t *testing.T) {
	if scope := RevocationScopeForAction(RevocationActionLogout); scope != RevokeCurrentSessionOnly {
		t.Fatalf("logout must revoke only the current session, got %v", scope)
	}

	for _, action := range []RevocationAction{
		RevocationActionPasswordChange,
		RevocationActionTOTPReplacement,
		RevocationActionAdminPasswordReset,
		RevocationActionAdminTOTPReset,
		RevocationActionAccountDisablement,
		RevocationActionExplicitRevokeAll,
	} {
		if scope := RevocationScopeForAction(action); scope != RevokeAllUserSessions {
			t.Fatalf("expected %v to revoke all user sessions, got %v", action, scope)
		}
	}
}
