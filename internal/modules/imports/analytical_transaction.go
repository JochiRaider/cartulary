package imports

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// AnalyticalUnitCommit carries Core-owned process-local authority. Its fields
// cannot be supplied by an extension. Only the application storage adapter sees
// the transaction, while callbacks receive the logical check/publication ports.
type AnalyticalUnitCommit struct {
	service         *service
	execution       jobs.Execution
	actor           authn.UserRecord
	start           applyStartResult
	unit            applyUnitData
	target          importTarget
	request         ExtensionImportApplyRequest
	requestHash     []byte
	participantID   string
	stopObservation func()
	coreFailure     error
	replayed        *unitApplyOutcome
	outcome         unitApplyOutcome
	validated       bool
	published       bool
}

func (s *service) applyAnalyticalUnit(ctx context.Context, execution jobs.Execution, actor authn.UserRecord, start applyStartResult, unitID uuid.UUID, stopObservation func()) (unitApplyOutcome, error) {
	if s.analyticalTransactions == nil {
		return unitApplyOutcome{}, crossownertransaction.ErrUnavailable
	}
	unit, err := s.store.readApplyUnit(ctx, s.store.db, start, unitID, false)
	if err != nil {
		if existing, findErr := s.store.getUnitOutcome(ctx, start.ImportSessionID, unitID); findErr == nil {
			return existing, nil
		}
		return unitApplyOutcome{}, err
	}
	target, ok := lookupApprovedImportTarget(unit.approvedMapping)
	if !ok || target.TargetKind == ImportTargetKindViewSchema || !target.importable(s.extensionProfileClaimed) {
		return unitApplyOutcome{}, importApplyBlockedError("target_kind_not_importable")
	}
	facade := s.extensionImportFacades[analyticalImportTargetKey{TargetKind: target.TargetKind, ExtensionProfileID: target.ExtensionProfileID}]
	if facade == nil || !target.ownerApplyFacadeAvailable() {
		return unitApplyOutcome{}, importApplyBlockedError("owner_apply_contract_unavailable")
	}
	request := extensionApplyRequest(actor, start, unit, target)
	participant, err := facade.ImportTransaction(request)
	if err != nil {
		return unitApplyOutcome{}, &translatedImportUnitError{failure: translateExtensionOwnerFailure(target, facade, err), cause: err}
	}
	if participant == nil {
		return unitApplyOutcome{}, crossownertransaction.ErrParticipantSet
	}
	key := analyticalApplyIdempotencyKey(actor, start)
	record, err := lookupRouteIdempotency(ctx, s.store.db, key)
	if err != nil {
		return unitApplyOutcome{}, err
	}
	if len(record.RequestHash) != 32 {
		return unitApplyOutcome{}, importApplyBlockedError("source_changed")
	}
	commit := &AnalyticalUnitCommit{service: s, execution: execution, actor: actor, start: start,
		unit: unit, target: target, request: request, requestHash: append([]byte(nil), record.RequestHash...),
		participantID: participant.ID(), stopObservation: stopObservation}
	_, err = s.analyticalTransactions.Execute(ctx, crossownertransaction.Operation{
		OperationID: unitCommitID(start.ImportSessionID, unitID), NormalizedRequestSHA256: hex.EncodeToString(record.RequestHash),
		Participants: []crossownertransaction.Participant{participant}, Preconditions: commit, Finalizer: commit,
	})
	if err != nil {
		if crossownertransaction.IsFatalIntegrity(err) {
			return unitApplyOutcome{}, err
		}
		if commit.replayed != nil {
			return *commit.replayed, nil
		}
		if commit.coreFailure != nil && errors.Is(err, commit.coreFailure) {
			return unitApplyOutcome{}, err
		}
		return unitApplyOutcome{}, &translatedImportUnitError{failure: translateExtensionOwnerFailure(target, facade, err), cause: err}
	}
	return commit.outcome, nil
}

func analyticalApplyIdempotencyKey(actor authn.UserRecord, start applyStartResult) authn.RouteIdempotencyKey {
	return authn.RouteIdempotencyKey{RouteKey: "imports.sessions.apply", ActorUserID: actor.ID, ScopeKey: start.ImportSessionID.String(), ClientTxnID: start.ClientTxnID}
}

func (c *AnalyticalUnitCommit) Check(ctx context.Context, access crossownertransaction.PreconditionCapability) error {
	port, ok := access.(interface {
		ValidateImportUnit(context.Context, *AnalyticalUnitCommit) error
	})
	if !ok {
		return crossownertransaction.ErrUnavailable
	}
	c.coreFailure = port.ValidateImportUnit(ctx, c)
	return c.coreFailure
}

// ValidateAnalyticalUnitTx is the physical Core adapter; it does not expose its
// borrowed transaction back to the logical callback.
func ValidateAnalyticalUnitTx(ctx context.Context, tx pgx.Tx, c *AnalyticalUnitCommit) error {
	if c == nil || c.service == nil || tx == nil || c.validated {
		return crossownertransaction.ErrUnavailable
	}
	unit, target, actor, err := c.service.validateUnitTx(ctx, tx, c.actor, c.start, c.unit.UnitID)
	if err != nil {
		if existing, findErr := findUnitOutcome(ctx, tx, c.start.ImportSessionID, c.unit.UnitID); findErr == nil {
			c.replayed = &existing
		}
		return err
	}
	if !reflect.DeepEqual(extensionApplyRequest(actor, c.start, unit, target), c.request) {
		return importApplyBlockedError("source_changed")
	}
	record, err := lookupRouteIdempotency(ctx, tx, analyticalApplyIdempotencyKey(c.actor, c.start))
	if err != nil {
		return err
	}
	if !bytes.Equal(record.RequestHash, c.requestHash) {
		return importApplyBlockedError("source_changed")
	}
	c.unit, c.target, c.validated = unit, target, true
	return nil
}

func (c *AnalyticalUnitCommit) Publish(ctx context.Context, access crossownertransaction.FinalizationCapability, values map[string]any) error {
	port, ok := access.(interface {
		PublishImportUnit(context.Context, *AnalyticalUnitCommit, map[string]any) error
	})
	if !ok {
		return crossownertransaction.ErrUnavailable
	}
	c.coreFailure = port.PublishImportUnit(ctx, c, values)
	return c.coreFailure
}

// PublishAnalyticalUnitTx runs the same short barrier and publication as a Core
// view-schema unit. Only the coordinator may subsequently commit the transaction.
func PublishAnalyticalUnitTx(ctx context.Context, tx pgx.Tx, c *AnalyticalUnitCommit, values map[string]any) error {
	if c == nil || c.service == nil || tx == nil || !c.validated || c.published || len(values) != 1 {
		return crossownertransaction.ErrWrite
	}
	result, ok := values[c.participantID].(ExtensionImportApplyResult)
	if !ok {
		return crossownertransaction.ErrWrite
	}
	outcome, err := c.service.publishUnitTx(ctx, tx, c.execution, c.actor, c.start, c.unit, c.target,
		appliedUnitCommit{OwnerResult: result.OwnerResponse, ResourceRefs: append([]jobs.ResourceRef(nil), result.ResourceRefs...)}, c.stopObservation)
	if err != nil {
		return err
	}
	c.outcome, c.published = outcome, true
	return nil
}

// ProveAnalyticalUnitCommit reads the immutable common unit outcome after an
// ambiguous driver reply. A missing or mismatched proof remains indeterminate.
func ProveAnalyticalUnitCommit(ctx context.Context, c *AnalyticalUnitCommit) bool {
	if c == nil || !c.published {
		return false
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	got, err := c.service.store.getUnitOutcome(ctx, c.start.ImportSessionID, c.unit.UnitID)
	want := c.outcome
	return err == nil && got.Status == "applied" && got.UnitCommitID == want.UnitCommitID && got.ApplyJobID == want.ApplyJobID &&
		got.ActorUserID == want.ActorUserID && got.OwnerBindingID == want.OwnerBindingID && got.TargetKind == want.TargetKind &&
		got.SourceDigest == want.SourceDigest && got.MappingFingerprint == want.MappingFingerprint
}
