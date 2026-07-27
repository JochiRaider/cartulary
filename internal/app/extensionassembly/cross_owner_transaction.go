package extensionassembly

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func CrossOwnerDescriptors(contracts []extensions.ParticipantContract) []crossownertransaction.Descriptor {
	result := make([]crossownertransaction.Descriptor, 0, len(contracts))
	for _, contract := range contracts {
		if contract.ContractKind != "cartulary.extension_transaction_participant_contract.v3" {
			continue
		}
		result = append(result, crossownertransaction.Descriptor{
			ParticipantID: contract.ParticipantID, OwnerProfileID: contract.OwnerProfileID,
			ContractSHA256: contract.ContractSHA256, InputSchemaID: contract.InputSchemaID,
			PrepareAlgorithmID:    contract.PrepareAlgorithmID,
			ValidationAlgorithmID: contract.ValidationAlgorithmID,
			WriteAlgorithmID:      contract.WriteAlgorithmID,
			SerializationKeyKinds: append([]string(nil), contract.SerializationKeyKinds...),
			OwnedStateFamilyIDs:   append([]string(nil), contract.OwnedStateFamilyIDs...),
		})
	}
	result = append(result, incidentbundles.ImportTransactionDescriptor())
	sort.Slice(result, func(i, j int) bool { return result[i].ParticipantID < result[j].ParticipantID })
	return result
}

// TransactionCapabilityProvider is the explicit application-composition edge
// between a physical PostgreSQL transaction and owner-local logical
// capabilities. It is not visible to the shared semantic coordinator.
type TransactionCapabilityProvider interface {
	TransactionCapabilities(string, pgx.Tx) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error)
}

// TransactionCapabilityMux is the explicit app-owned dispatch edge for the
// currently admitted semantic owners. It is closed rather than a callback
// registry, so participant admission cannot be expanded at runtime.
type TransactionCapabilityMux struct {
	NetworkFlow     *networkflow.Module
	IncidentBundles *incidentbundles.ImportTransactionProvider
}

func (m TransactionCapabilityMux) TransactionCapabilities(participantID string, tx pgx.Tx) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error) {
	switch participantID {
	case networkflow.ImportApplyParticipantID, networkflow.IndicatorLinkParticipantID:
		return m.NetworkFlow.TransactionCapabilities(participantID, tx)
	case incidentbundles.ImportTransactionParticipantID:
		return m.IncidentBundles.TransactionCapabilities(participantID, tx)
	default:
		return nil, nil, fmt.Errorf("%w: %s", crossownertransaction.ErrParticipantSet, participantID)
	}
}

type CrossOwnerBackend struct {
	database postgres.DB
	provider TransactionCapabilityProvider
}

func NewCrossOwnerBackend(database postgres.DB, provider TransactionCapabilityProvider) (*CrossOwnerBackend, error) {
	if database == nil || provider == nil {
		return nil, crossownertransaction.ErrUnavailable
	}
	return &CrossOwnerBackend{database: database, provider: provider}, nil
}

func (b *CrossOwnerBackend) Begin(ctx context.Context, descriptors []crossownertransaction.Descriptor) (crossownertransaction.Transaction, error) {
	if b == nil || b.database == nil || b.provider == nil {
		return nil, crossownertransaction.ErrUnavailable
	}
	tx, err := b.database.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &crossOwnerTransaction{
		tx: tx, provider: b.provider, descriptors: append([]crossownertransaction.Descriptor(nil), descriptors...),
	}, nil
}

type crossOwnerTransaction struct {
	tx          pgx.Tx
	provider    TransactionCapabilityProvider
	descriptors []crossownertransaction.Descriptor
	closed      bool
}

func (t *crossOwnerTransaction) AcquireSerializationLock(ctx context.Context, key crossownertransaction.OrderedSerializationKey) error {
	if t == nil || t.tx == nil || t.closed {
		return crossownertransaction.ErrUnavailable
	}
	identity := key.ParticipantID + "\x1f" + key.KeyKind + "\x1f" + key.Key
	_, err := t.tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, identity)
	return err
}

func (t *crossOwnerTransaction) ReadCapability(participantID string) (crossownertransaction.ReadCapability, error) {
	read, _, err := t.capabilities(participantID)
	return read, err
}

func (t *crossOwnerTransaction) WriteCapability(participantID string) (crossownertransaction.WriteCapability, error) {
	_, write, err := t.capabilities(participantID)
	return write, err
}

func (t *crossOwnerTransaction) capabilities(participantID string) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error) {
	if t == nil || t.tx == nil || t.closed {
		return nil, nil, crossownertransaction.ErrUnavailable
	}
	admitted := false
	for _, descriptor := range t.descriptors {
		if descriptor.ParticipantID == participantID {
			admitted = true
			break
		}
	}
	if !admitted {
		return nil, nil, fmt.Errorf("%w: capability %s", crossownertransaction.ErrParticipantSet, participantID)
	}
	return t.provider.TransactionCapabilities(participantID, t.tx)
}

func (t *crossOwnerTransaction) FinalizationCapability() (crossownertransaction.FinalizationCapability, error) {
	if t == nil || t.tx == nil || t.closed {
		return nil, crossownertransaction.ErrUnavailable
	}
	return crossOwnerFinalization{tx: t.tx}, nil
}

func (t *crossOwnerTransaction) Commit(ctx context.Context) (crossownertransaction.CommitOutcome, error) {
	if t == nil || t.tx == nil || t.closed {
		return crossownertransaction.CommitUnknown, crossownertransaction.ErrUnavailable
	}
	t.closed = true
	err := t.tx.Commit(ctx)
	if err == nil {
		return crossownertransaction.CommitProven, nil
	}
	if errors.Is(err, pgx.ErrTxCommitRollback) || provenAbsentSQLState(err) {
		return crossownertransaction.CommitAbsent, err
	}
	return crossownertransaction.CommitUnknown, err
}

func (t *crossOwnerTransaction) Rollback(ctx context.Context) (crossownertransaction.CommitOutcome, error) {
	if t == nil || t.tx == nil {
		return crossownertransaction.CommitUnknown, crossownertransaction.ErrUnavailable
	}
	if t.closed {
		return crossownertransaction.CommitUnknown, pgx.ErrTxClosed
	}
	t.closed = true
	err := t.tx.Rollback(ctx)
	if err == nil {
		return crossownertransaction.CommitAbsent, nil
	}
	return crossownertransaction.CommitUnknown, err
}

type crossOwnerFinalization struct {
	tx pgx.Tx
}

func (crossOwnerFinalization) FinalizationScope() string { return "shared.finalization" }

func provenAbsentSQLState(err error) bool {
	type sqlState interface{ SQLState() string }
	var state sqlState
	if !errors.As(err, &state) {
		return false
	}
	switch state.SQLState() {
	case "40001", "40P01", "23505", "23503", "23514":
		return true
	default:
		return false
	}
}
