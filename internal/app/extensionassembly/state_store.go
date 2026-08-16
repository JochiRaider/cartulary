package extensionassembly

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"sync/atomic"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
)

// StateStore adapts the physical PostgreSQL extensionstore to the logical,
// capability-limited state-coordination port. It contains no SQL or profile
// table knowledge.
type StateStore struct {
	store    *extensionstore.Store
	sequence atomic.Uint64
}

func NewStateStore(store *extensionstore.Store) (*StateStore, error) {
	if store == nil {
		return nil, errors.New("extension state store adapter requires a physical store")
	}
	return &StateStore{store: store}, nil
}

func (s *StateStore) WithProfileSession(ctx context.Context, profileID string, lockTimeout time.Duration, operation func(extensions.ProfileSession) error) error {
	if s == nil || s.store == nil || operation == nil {
		return errors.New("extension state store adapter is unavailable")
	}
	err := s.store.WithProfileLock(ctx, profileID, lockTimeout, func(session *extensionstore.Session) error {
		return operation(&stateSession{
			owner:     s,
			session:   session,
			profileID: profileID,
		})
	})
	if errors.Is(err, extensionstore.ErrProfileLockTimeout) {
		return extensions.ErrStateMigrationLockTimeout
	}
	return err
}

type stateSession struct {
	owner     *StateStore
	session   *extensionstore.Session
	profileID string
}

func (s *stateSession) Snapshot(ctx context.Context, profileID, lineageID string, familyIDs []string) (extensions.StateSnapshot, error) {
	tx, err := s.Begin(ctx, profileID, familyIDs)
	if err != nil {
		return extensions.StateSnapshot{}, err
	}
	metadata, err := tx.StateMetadata(ctx, profileID)
	if err != nil {
		_, _ = tx.Rollback(context.WithoutCancel(ctx))
		return extensions.StateSnapshot{}, err
	}
	ledger, err := tx.MigrationLedger(ctx, profileID, lineageID)
	if err != nil {
		_, _ = tx.Rollback(context.WithoutCancel(ctx))
		return extensions.StateSnapshot{}, err
	}
	counts, err := tx.FamilyCounts(ctx, familyIDs)
	if err != nil {
		_, _ = tx.Rollback(context.WithoutCancel(ctx))
		return extensions.StateSnapshot{}, err
	}
	outcome, rollbackErr := tx.Rollback(context.WithoutCancel(ctx))
	if outcome == extensions.CommitIndeterminate {
		return extensions.StateSnapshot{}, fmt.Errorf("extension state snapshot rollback indeterminate: %w", rollbackErr)
	}
	if rollbackErr != nil {
		return extensions.StateSnapshot{}, rollbackErr
	}
	return extensions.StateSnapshot{
		Metadata:     metadata,
		FamilyCounts: counts,
		Ledger:       ledger,
	}, nil
}

func (s *stateSession) Begin(ctx context.Context, profileID string, familyIDs []string) (extensions.StateTransaction, error) {
	if s == nil || s.session == nil || profileID == "" || profileID != s.profileID {
		return nil, errors.New("extension state profile session mismatch")
	}
	normalized, err := normalizeStateFamilies(familyIDs)
	if err != nil {
		return nil, err
	}
	tx, err := s.session.Begin(ctx)
	if err != nil {
		return nil, err
	}
	sequence := s.owner.sequence.Add(1)
	return &stateTransaction{
		tx:           tx,
		profileID:    profileID,
		familyIDs:    normalized,
		capabilityID: fmt.Sprintf("extension-state:%s:%d", profileID, sequence),
	}, nil
}

type stateTransaction struct {
	tx           *extensionstore.Tx
	profileID    string
	familyIDs    []string
	capabilityID string
}

func (t *stateTransaction) CapabilityID() string {
	if t == nil {
		return ""
	}
	return t.capabilityID
}

func (t *stateTransaction) StateFamilyIDs() []string {
	if t == nil {
		return nil
	}
	return append([]string(nil), t.familyIDs...)
}

func (t *stateTransaction) IsStateWriteCapability() bool {
	return t != nil && t.tx != nil
}

func (t *stateTransaction) FamilyCounts(ctx context.Context, familyIDs []string) (map[string]int64, error) {
	if t == nil || t.tx == nil {
		return nil, errors.New("extension state capability is unavailable")
	}
	normalized, err := normalizeStateFamilies(familyIDs)
	if err != nil {
		return nil, err
	}
	if !equalStateFamilies(normalized, t.familyIDs) {
		return nil, errors.New("extension state capability family scope violation")
	}
	return t.tx.FamilyCounts(ctx, normalized)
}

func (t *stateTransaction) ValidateFamilyState(ctx context.Context, familyID string) error {
	if t == nil || t.tx == nil {
		return errors.New("extension state capability is unavailable")
	}
	if !slices.Contains(t.familyIDs, familyID) {
		return errors.New("extension state capability family scope violation")
	}
	return t.tx.ValidateFamilyState(ctx, familyID)
}

func (t *stateTransaction) StateMetadata(ctx context.Context, profileID string) (*extensions.StateMetadata, error) {
	if t == nil || t.tx == nil || profileID != t.profileID {
		return nil, errors.New("extension state metadata scope violation")
	}
	metadata, err := t.tx.StateMetadata(ctx, profileID)
	if err != nil || metadata == nil {
		return nil, err
	}
	return &extensions.StateMetadata{
		ProfileID:          metadata.ProfileID,
		MigrationLineageID: metadata.MigrationLineageID,
		StateVersion:       metadata.StateVersion,
		LastMigrationID:    cloneString(metadata.LastMigrationID),
		MetadataVersion:    metadata.MetadataVersion,
		CreatedAt:          metadata.CreatedAt,
		UpdatedAt:          metadata.UpdatedAt,
	}, nil
}

func (t *stateTransaction) MigrationLedger(ctx context.Context, profileID, lineageID string) ([]extensions.MigrationLedgerEntry, error) {
	if t == nil || t.tx == nil || profileID != t.profileID {
		return nil, errors.New("extension migration ledger scope violation")
	}
	entries, err := t.tx.MigrationLedger(ctx, profileID, lineageID)
	if err != nil {
		return nil, err
	}
	result := make([]extensions.MigrationLedgerEntry, len(entries))
	for index, entry := range entries {
		result[index] = extensions.MigrationLedgerEntry{
			ProfileID:                 entry.ProfileID,
			MigrationLineageID:        entry.MigrationLineageID,
			MigrationID:               entry.MigrationID,
			FromStateVersion:          entry.FromStateVersion,
			ToStateVersion:            entry.ToStateVersion,
			MigrationDefinitionSHA256: entry.MigrationDefinitionSHA256,
			CommittedAt:               entry.CommittedAt,
			ResultingStateVersion:     entry.ResultingStateVersion,
		}
	}
	return result, nil
}

func (t *stateTransaction) InsertStateMetadata(ctx context.Context, metadata extensions.StateMetadata) error {
	if t == nil || t.tx == nil || metadata.ProfileID != t.profileID {
		return errors.New("extension state metadata write scope violation")
	}
	return t.tx.InsertStateMetadata(ctx, extensionstore.StateMetadata{
		ProfileID:          metadata.ProfileID,
		MigrationLineageID: metadata.MigrationLineageID,
		StateVersion:       metadata.StateVersion,
		LastMigrationID:    cloneString(metadata.LastMigrationID),
		MetadataVersion:    metadata.MetadataVersion,
		CreatedAt:          metadata.CreatedAt,
		UpdatedAt:          metadata.UpdatedAt,
	})
}

func (t *stateTransaction) UpdateStateMetadata(ctx context.Context, before extensions.StateMetadata, stateVersion int, migrationID string, now time.Time) error {
	if t == nil || t.tx == nil || before.ProfileID != t.profileID {
		return errors.New("extension state metadata update scope violation")
	}
	return t.tx.UpdateStateMetadata(ctx, extensionstore.StateMetadata{
		ProfileID:          before.ProfileID,
		MigrationLineageID: before.MigrationLineageID,
		StateVersion:       before.StateVersion,
		LastMigrationID:    cloneString(before.LastMigrationID),
		MetadataVersion:    before.MetadataVersion,
		CreatedAt:          before.CreatedAt,
		UpdatedAt:          before.UpdatedAt,
	}, stateVersion, migrationID, now)
}

func (t *stateTransaction) InsertMigrationLedger(ctx context.Context, entry extensions.MigrationLedgerEntry) error {
	if t == nil || t.tx == nil || entry.ProfileID != t.profileID {
		return errors.New("extension migration ledger write scope violation")
	}
	return t.tx.InsertMigrationLedger(ctx, extensionstore.MigrationLedgerEntry{
		ProfileID:                 entry.ProfileID,
		MigrationLineageID:        entry.MigrationLineageID,
		MigrationID:               entry.MigrationID,
		FromStateVersion:          entry.FromStateVersion,
		ToStateVersion:            entry.ToStateVersion,
		MigrationDefinitionSHA256: entry.MigrationDefinitionSHA256,
		CommittedAt:               entry.CommittedAt,
		ResultingStateVersion:     entry.ResultingStateVersion,
	})
}

func (t *stateTransaction) Commit(ctx context.Context) (extensions.CommitOutcome, error) {
	outcome, err := t.tx.Commit(ctx)
	return mapCommitOutcome(outcome), err
}

func (t *stateTransaction) Rollback(ctx context.Context) (extensions.CommitOutcome, error) {
	outcome, err := t.tx.Rollback(ctx)
	return mapCommitOutcome(outcome), err
}

func mapCommitOutcome(outcome extensionstore.CommitOutcome) extensions.CommitOutcome {
	switch outcome {
	case extensionstore.CommitProven:
		return extensions.CommitCommitted
	case extensionstore.CommitAbsent:
		return extensions.CommitAbsent
	default:
		return extensions.CommitIndeterminate
	}
}

func normalizeStateFamilies(familyIDs []string) ([]string, error) {
	result := append([]string(nil), familyIDs...)
	sort.Strings(result)
	for index, familyID := range result {
		if familyID == "" || (index > 0 && result[index-1] == familyID) {
			return nil, errors.New("extension state families must be nonempty and unique")
		}
	}
	return result, nil
}

func equalStateFamilies(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
