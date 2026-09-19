package networkflowsupport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
)

// AuthorizationBinding resolves symbolic control references only to rows verified
// in this fixture's database. It never accepts row IDs from an armed request.
type AuthorizationBinding struct {
	ActorRef, IncidentRef, ResourceKind, ResourceRef string
	ActorID, IncidentID, SessionID                   uuid.UUID
	TableID                                          string
}

type AuthorizationFixture struct {
	mu         sync.Mutex
	registry   *hc.NetworkFlowAuthTransitionRegistry
	db         postgres.DB
	binding    AuthorizationBinding
	membership json.RawMessage
	now        func() time.Time
}

func NewAuthorizationFixture(ctx context.Context, registry *hc.NetworkFlowAuthTransitionRegistry, db postgres.DB, b AuthorizationBinding, now func() time.Time) (*AuthorizationFixture, error) {
	if registry == nil || db == nil || now == nil || b.ActorID == uuid.Nil || b.IncidentID == uuid.Nil {
		return nil, errors.New("authorization fixture dependencies required")
	}
	var member json.RawMessage
	if err := db.QueryRow(ctx, `SELECT to_jsonb(m) FROM incident_memberships m JOIN users u ON u.id=m.user_id JOIN incidents i ON i.id=m.incident_id WHERE m.incident_id=$1 AND m.user_id=$2`, b.IncidentID, b.ActorID).Scan(&member); err != nil {
		return nil, fmt.Errorf("resolve owned fixture membership: %w", err)
	}
	if b.SessionID != uuid.Nil {
		var present bool
		if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_sessions WHERE id=$1 AND user_id=$2)`, b.SessionID, b.ActorID).Scan(&present); err != nil || !present {
			return nil, errors.New("session is outside fixture actor")
		}
	}
	if b.TableID != "" {
		var present bool
		if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM network_flow_tables WHERE network_flow_table_id=$1 AND incident_id=$2)`, b.TableID, b.IncidentID).Scan(&present); err != nil || !present {
			return nil, errors.New("table is outside fixture incident")
		}
	}
	identity := b.ActorID.String() + ":" + b.IncidentID.String() + ":" + b.SessionID.String() + ":" + b.TableID
	if err := registry.BindFixture(b.ActorRef, b.IncidentRef, b.ResourceKind, b.ResourceRef, identity); err != nil {
		return nil, err
	}
	return &AuthorizationFixture{registry: registry, db: db, binding: b, membership: member, now: now}, nil
}

// Before runs immediately before the real HTTP request. The cursor boundary is
// used only for a continuation request. Product authentication/admission still
// executes normally after the fixture mutation commits.
func (f *AuthorizationFixture) Before(ctx context.Context, boundary, correlation string) (bool, error) {
	if boundary != hc.NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization && boundary != hc.NetworkFlowAuthTransitionBoundaryCursorBeforeAuthorizationRecheck {
		return false, errors.New("unsupported authorization fixture boundary")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	b := f.binding
	transition, ok := f.registry.ConsumeNetworkFlowAuthTransitionFor(boundary, b.ActorRef, b.IncidentRef, b.ResourceKind, b.ResourceRef, correlation)
	if !ok {
		return false, nil
	}
	var err error
	switch transition.TransitionKind {
	case hc.NetworkFlowAuthTransitionKindIncidentMembershipRevoked:
		_, err = f.db.Exec(ctx, `DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, b.IncidentID, b.ActorID)
	case hc.NetworkFlowAuthTransitionKindIncidentMembershipRestored:
		_, err = f.db.Exec(ctx, `INSERT INTO incident_memberships SELECT * FROM jsonb_populate_record(NULL::incident_memberships,$1::jsonb) ON CONFLICT(incident_id,user_id) DO NOTHING`, f.membership)
	case hc.NetworkFlowAuthTransitionKindIncidentDeleted:
		_, err = f.db.Exec(ctx, `DELETE FROM incidents WHERE id=$1`, b.IncidentID)
	case hc.NetworkFlowAuthTransitionKindSessionRevoked:
		if b.SessionID == uuid.Nil {
			return true, errors.New("fixture has no bound session")
		}
		err = authn.NewStore(f.db).RevokeSession(ctx, b.SessionID, "admin_revoked", f.now())
	case hc.NetworkFlowAuthTransitionKindNetworkFlowTableSoftDeleted:
		if b.TableID == "" {
			return true, errors.New("fixture has no bound table")
		}
		_, err = f.db.Exec(ctx, `UPDATE network_flow_tables SET table_status='soft_deleted',table_version=table_version+1,deleted_at=$3,updated_at=$3 WHERE incident_id=$1 AND network_flow_table_id=$2 AND table_status='active'`, b.IncidentID, b.TableID, f.now())
	case hc.NetworkFlowAuthTransitionKindNetworkFlowTableRenamed:
		if b.TableID == "" {
			return true, errors.New("fixture has no bound table")
		}
		_, err = f.db.Exec(ctx, `UPDATE network_flow_tables SET display_name=$3,table_version=table_version+1,updated_at=$4 WHERE incident_id=$1 AND network_flow_table_id=$2 AND table_status='active'`, b.IncidentID, b.TableID, "Fixture "+b.TableID, f.now())
	default:
		return true, errors.New("unsupported authorization fixture transition")
	}
	return true, err
}
