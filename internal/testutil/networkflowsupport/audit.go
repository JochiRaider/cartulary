package networkflowsupport

import (
	"context"
	"errors"
	"fmt"
	"sync"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
)

type AuditBinding struct {
	Scope                              hc.NetworkFlowAuditScope
	ActorID, IncidentID                uuid.UUID
	ResourceID, ClientTxnID, RequestID string
}

type AuditFixture struct {
	mu       sync.Mutex
	registry *hc.NetworkFlowAuditAssertionRegistry
	db       postgres.DB
	binding  AuditBinding
	baseline int
	verified bool
}

// NewAuditFixture binds scope and observes the committed baseline before the
// operation. A newly allocated resource may be supplied to Verify from its real
// response; its operation must have an empty baseline until that identity exists.
func NewAuditFixture(ctx context.Context, registry *hc.NetworkFlowAuditAssertionRegistry, db postgres.DB, b AuditBinding) (*AuditFixture, error) {
	if registry == nil || db == nil {
		return nil, errors.New("audit fixture dependencies required")
	}
	var present bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1) AND EXISTS(SELECT 1 FROM incidents WHERE id=$2)`, b.ActorID, b.IncidentID).Scan(&present); err != nil || !present {
		return nil, errors.New("unresolved owned audit actor or incident")
	}
	selector := auditSelector(b)
	baseline, err := authstoretest.CountAuditOccurrences(ctx, db, selector)
	if err != nil {
		return nil, err
	}
	if b.ResourceID == "" && baseline != 0 {
		return nil, errors.New("nonempty audit baseline requires a resolved resource")
	}
	identity := b.ActorID.String() + ":" + b.IncidentID.String() + ":" + b.ClientTxnID + ":" + b.RequestID + ":" + b.ResourceID
	if err := registry.BindFixture(b.Scope, identity); err != nil {
		return nil, err
	}
	return &AuditFixture{registry: registry, db: db, binding: b, baseline: baseline}, nil
}
func auditSelector(b AuditBinding) authstoretest.AuditOccurrenceSelector {
	return authstoretest.AuditOccurrenceSelector{ActorID: b.ActorID, IncidentID: b.IncidentID, EventCode: b.Scope.EventCode, ResourceID: b.ResourceID, ClientTxnID: b.ClientTxnID, RequestID: b.RequestID}
}

// Verify runs only after the real HTTP operation or durable job completion.
// Replay is required only for adopted idempotent operations, never graph query.
func (f *AuditFixture) Verify(ctx context.Context, resourceID string, replay func() error) (err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	defer func() {
		if err != nil {
			f.registry.RecordFailure()
		}
	}()
	if f.verified {
		return errors.New("audit fixture already verified")
	}
	f.verified = true
	assertion, ok := f.registry.Consume(f.binding.Scope)
	if !ok {
		return errors.New("required matching audit assertion missing")
	}
	if f.baseline != assertion.BaselineCount {
		return fmt.Errorf("audit baseline: observed %d expected %d", f.baseline, assertion.BaselineCount)
	}
	b := f.binding
	if b.ResourceID != "" && resourceID != "" && b.ResourceID != resourceID {
		return errors.New("audit result resource differs from bound resource")
	}
	if resourceID != "" {
		b.ResourceID = resourceID
	}
	if assertion.ExpectedFinalCount > 0 && b.ResourceID == "" {
		return errors.New("positive audit assertion requires a resolved result resource")
	}
	observed, err := authstoretest.CountAuditOccurrences(ctx, f.db, auditSelector(b))
	if err != nil {
		return err
	}
	if observed != assertion.ExpectedFinalCount {
		return fmt.Errorf("audit final count: observed %d expected %d", observed, assertion.ExpectedFinalCount)
	}
	if assertion.AssertionKind == hc.NetworkFlowAuditAssertionNoAuditReplay {
		if replay == nil || b.Scope.EventCode == hc.NetworkFlowAuditEventGraphQueryExecuted {
			return errors.New("audit replay requires an adopted replay operation")
		}
		if err := replay(); err != nil {
			return err
		}
		after, err := authstoretest.CountAuditOccurrences(ctx, f.db, auditSelector(b))
		if err != nil {
			return err
		}
		if after != observed {
			return fmt.Errorf("audit replay changed count: %d to %d", observed, after)
		}
	} else if replay != nil {
		return errors.New("replay supplied to a non-replay audit assertion")
	}
	return nil
}
