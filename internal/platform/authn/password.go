package authn

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/sync/semaphore"
)

const (
	PasswordIterations       = 600000
	PasswordActiveWorkflows  = 2
	PasswordPendingWorkflows = 8
	PasswordQueueLimit       = 2 * time.Second
	passwordSaltBytes        = 16
	passwordHashBytes        = 32
	passwordRecordPrefix     = "pbkdf2-sha256$v=1$i=600000$"
)

var (
	ErrAuthenticationCapacity = errors.New("authentication capacity exhausted")
	ErrPasswordRecord         = errors.New("unsupported password record")
	ErrPasswordWorkflow       = errors.New("password workflow is closed")
	passwordAdmission         = newPasswordBudget()
)

type passwordBudget struct {
	active *semaphore.Weighted
	total  chan struct{}
}

func newPasswordBudget() *passwordBudget {
	return &passwordBudget{active: semaphore.NewWeighted(PasswordActiveWorkflows), total: make(chan struct{}, PasswordActiveWorkflows+PasswordPendingWorkflows)}
}

// PasswordWorkflow retains one process-wide slot for sequential credential
// derivations. Close is idempotent and waits for any running derivation to finish.
// Cancellation never releases a slot while PBKDF2 is still consuming CPU.
type PasswordWorkflow struct {
	mu     sync.Mutex
	ctx    context.Context
	budget *passwordBudget
	closed bool
}

func BeginPasswordWorkflow(ctx context.Context) (*PasswordWorkflow, error) {
	return passwordAdmission.begin(ctx)
}

func (budget *passwordBudget) begin(ctx context.Context) (*PasswordWorkflow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case budget.total <- struct{}{}:
	default:
		return nil, ErrAuthenticationCapacity
	}
	wait, cancel := context.WithTimeout(ctx, PasswordQueueLimit)
	defer cancel()
	if err := budget.active.Acquire(wait, 1); err != nil {
		<-budget.total
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrAuthenticationCapacity
	}
	return &PasswordWorkflow{ctx: ctx, budget: budget}, nil
}

func (workflow *PasswordWorkflow) Close() {
	workflow.mu.Lock()
	defer workflow.mu.Unlock()
	if !workflow.closed && workflow.budget != nil {
		workflow.closed = true
		workflow.budget.active.Release(1)
		<-workflow.budget.total
	}
}

func (workflow *PasswordWorkflow) derive(password string, salt []byte) ([]byte, error) {
	if workflow.closed || workflow.budget == nil || workflow.ctx == nil {
		return nil, ErrPasswordWorkflow
	}
	if err := workflow.ctx.Err(); err != nil {
		return nil, err
	}
	derived, err := pbkdf2.Key(sha256.New, password, salt, PasswordIterations, passwordHashBytes)
	if err != nil {
		return nil, err
	}
	if err := workflow.ctx.Err(); err != nil {
		clear(derived)
		return nil, err
	}
	return derived, nil
}

func (workflow *PasswordWorkflow) Hash(password string) (string, error) {
	accepted, err := ValidatePasswordProvision(password)
	if err != nil {
		return "", err
	}
	workflow.mu.Lock()
	defer workflow.mu.Unlock()
	var salt [passwordSaltBytes]byte
	rand.Read(salt[:])
	derived, err := workflow.derive(accepted, salt[:])
	if err != nil {
		return "", err
	}
	defer clear(derived)
	return passwordRecordPrefix + base64.RawStdEncoding.EncodeToString(salt[:]) + "$" + base64.RawStdEncoding.EncodeToString(derived), nil
}

func (workflow *PasswordWorkflow) Verify(record, password string) (bool, error) {
	salt, want, err := decodePasswordRecord(record)
	if err != nil {
		return false, err
	}
	if !validPasswordVerificationInput(password) {
		return false, nil
	}
	workflow.mu.Lock()
	defer workflow.mu.Unlock()
	got, err := workflow.derive(password, salt)
	if err != nil {
		return false, err
	}
	defer clear(got)
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

func HashPassword(ctx context.Context, password string) (string, error) {
	if _, err := ValidatePasswordProvision(password); err != nil {
		return "", err
	}
	workflow, err := BeginPasswordWorkflow(ctx)
	if err != nil {
		return "", err
	}
	defer workflow.Close()
	return workflow.Hash(password)
}

func VerifyPasswordHash(ctx context.Context, record, password string) (bool, error) {
	if _, _, err := decodePasswordRecord(record); err != nil {
		return false, err
	}
	if !validPasswordVerificationInput(password) {
		return false, nil
	}
	workflow, err := BeginPasswordWorkflow(ctx)
	if err != nil {
		return false, err
	}
	defer workflow.Close()
	return workflow.Verify(record, password)
}

func validPasswordVerificationInput(password string) bool {
	return len(password) <= maxPasswordScalars*utf8.UTFMax && utf8.ValidString(password) && utf8.RuneCountInString(password) <= maxPasswordScalars
}

func decodePasswordRecord(record string) ([]byte, []byte, error) {
	if len(record) > 128 || !strings.HasPrefix(record, passwordRecordPrefix) {
		return nil, nil, ErrPasswordRecord
	}
	parts := strings.Split(strings.TrimPrefix(record, passwordRecordPrefix), "$")
	if len(parts) != 2 || len(parts[0]) != 22 || len(parts[1]) != 43 {
		return nil, nil, ErrPasswordRecord
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[0])
	if err != nil || len(salt) != passwordSaltBytes || base64.RawStdEncoding.EncodeToString(salt) != parts[0] {
		return nil, nil, ErrPasswordRecord
	}
	hash, err := base64.RawStdEncoding.Strict().DecodeString(parts[1])
	if err != nil || len(hash) != passwordHashBytes || base64.RawStdEncoding.EncodeToString(hash) != parts[1] {
		return nil, nil, ErrPasswordRecord
	}
	return salt, hash, nil
}
