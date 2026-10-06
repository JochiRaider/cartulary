package authn

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// Independently computed with Python hashlib.pbkdf2_hmac, SHA-256, 600000
// iterations, salt bytes 00..0f and a 32-byte result.
const passwordVector = "pbkdf2-sha256$v=1$i=600000$AAECAwQFBgcICQoLDA0ODw$atIkBsu02T8+eKABFCd1KJL+GwfpPONGm0YcdNJCFfA"

func TestPasswordRecordVectorsAndRejection(t *testing.T) {
	ctx := context.Background()
	if ok, err := VerifyPasswordHash(ctx, passwordVector, "Password vector 1!"); err != nil || !ok {
		t.Fatalf("independent vector %t %v", ok, err)
	}
	if ok, err := VerifyPasswordHash(ctx, passwordVector, "Password vector 2!"); err != nil || ok {
		t.Fatalf("wrong password %t %v", ok, err)
	}
	for _, record := range []string{"", strings.Repeat("x", 129), strings.Replace(passwordVector, "v=1", "v=2", 1), strings.Replace(passwordVector, "600000", "600001", 1), strings.Replace(passwordVector, "pbkdf2-sha256", "argon2id", 1), strings.Replace(passwordVector, "0ODw$", "0ODw=$", 1), passwordVector + "=", strings.Replace(passwordVector, "0ODw$", "0ODx$", 1), passwordVector + "$", strings.Replace(passwordVector, "AAEC", "\nAEC", 1)} {
		// A canceled context would reject capacity admission. Record rejection must
		// still win, proving untrusted cost/framing never reaches admission or KDF.
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := VerifyPasswordHash(canceled, record, "Password vector 1!"); !errors.Is(err, ErrPasswordRecord) {
			t.Fatalf("record rejection %v", err)
		}
	}
	if ok, err := VerifyPasswordHash(ctx, passwordVector, strings.Repeat("a", 1025)); err != nil || ok {
		t.Fatalf("password input bound %t %v", ok, err)
	}
	first, err := HashPassword(ctx, "Fresh password 1!")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword(ctx, "Fresh password 1!")
	if err != nil || first == second {
		t.Fatal("password salt reused")
	}
	if ok, err := VerifyPasswordHash(ctx, first, "Fresh password 1!"); err != nil || !ok {
		t.Fatalf("record roundtrip %t %v", ok, err)
	}
}

func TestPasswordWorkflowCapacityCancellationAndLifetime(t *testing.T) {
	budget := newPasswordBudget()
	ctx := context.Background()
	first, err := budget.begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := budget.begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	waitCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan error, PasswordPendingWorkflows)
	for range PasswordPendingWorkflows {
		go func() {
			workflow, err := budget.begin(waitCtx)
			if workflow != nil {
				workflow.Close()
			}
			results <- err
		}()
	}
	deadline := time.After(time.Second)
	for len(budget.total) != PasswordActiveWorkflows+PasswordPendingWorkflows {
		select {
		case <-deadline:
			t.Fatal("waiters did not fill bounded queue")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	start := time.Now()
	if _, err := budget.begin(ctx); !errors.Is(err, ErrAuthenticationCapacity) {
		t.Fatalf("overflow %v", err)
	}
	if time.Since(start) > 250*time.Millisecond {
		t.Fatal("excess work was queued")
	}
	cancel()
	for range PasswordPendingWorkflows {
		if err := <-results; !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation %v", err)
		}
	}
	if len(budget.total) != PasswordActiveWorkflows {
		t.Fatal("canceled waiters leaked capacity")
	}
	// Sequential verify + replace must work while the other slot remains held.
	if ok, err := first.Verify(passwordVector, "Password vector 1!"); err != nil || !ok {
		t.Fatalf("held workflow verify %t %v", ok, err)
	}
	if _, err := first.Hash("Replacement password 1!"); err != nil {
		t.Fatalf("workflow reacquired budget: %v", err)
	}
	first.Close()
	first.Close()
	if _, err := first.Hash("Replacement password 1!"); !errors.Is(err, ErrPasswordWorkflow) {
		t.Fatalf("closed workflow %v", err)
	}
	third, err := budget.begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer third.Close()
	// Closing an in-flight capability cannot free its slot until its critical
	// section ends, including when request cancellation races with cleanup.
	third.mu.Lock()
	closed := make(chan struct{})
	go func() { third.Close(); close(closed) }()
	select {
	case <-closed:
		t.Fatal("closed active derivation")
	case <-time.After(10 * time.Millisecond):
	}
	if len(budget.total) != PasswordActiveWorkflows {
		t.Fatal("released active capacity early")
	}
	third.mu.Unlock()
	<-closed
}

func TestPasswordWorkflowQueueDeadline(t *testing.T) {
	budget := newPasswordBudget()
	first, err := budget.begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := budget.begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	start := time.Now()
	if _, err := budget.begin(context.Background()); !errors.Is(err, ErrAuthenticationCapacity) {
		t.Fatalf("queue deadline %v", err)
	}
	if elapsed := time.Since(start); elapsed < PasswordQueueLimit || elapsed > PasswordQueueLimit+time.Second {
		t.Fatalf("queue lifetime %s", elapsed)
	}
	if len(budget.total) != PasswordActiveWorkflows {
		t.Fatal("timed-out waiter leaked capacity")
	}
}

func TestTOTPSHA256VectorsAndSecretAdmission(t *testing.T) {
	// RFC 6238 Appendix B SHA-256 at t=59 is 46119246 with eight digits.
	// The current six-digit profile is its final six digits: 119246.
	secret := EncodeSecretBase32([]byte("12345678901234567890123456789012"))
	for _, at := range []int64{29, 59, 89} {
		if !ValidateTOTPCode(secret, "119246", time.Unix(at, 0)) {
			t.Fatalf("valid window at %d", at)
		}
	}
	if ValidateTOTPCode(secret, "119246", time.Unix(119, 0)) {
		t.Fatal("accepted outside window")
	}
	for _, code := range []string{"46119246", "287082", "11924", "11924x"} {
		if ValidateTOTPCode(secret, code, time.Unix(59, 0)) {
			t.Fatalf("invalid code %s", code)
		}
	}
	for _, invalid := range []string{strings.ToLower(secret), secret + "=", EncodeSecretBase32([]byte("12345678901234567890")), ""} {
		if ValidateTOTPCode(invalid, "119246", time.Unix(59, 0)) {
			t.Fatal("unsupported secret")
		}
	}
	raw, encoded, err := GenerateTOTPSecret()
	if err != nil || len(raw) != 32 || len(encoded) != 52 {
		t.Fatalf("generated secret shape %v", err)
	}
}
