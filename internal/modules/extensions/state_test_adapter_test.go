package extensions

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
)

const testStateFamilyID = "test_profile.members"

type stateTestFaults struct {
	mu          sync.Mutex
	commitModes []string
	acquired    int
	released    int
	active      int
	maxActive   int
	onAttempt   func()
	onAcquire   func()
	onRelease   func()
	onSnapshot  func(int)
	snapshots   int
}

func (f *stateTestFaults) nextCommitMode() string {
	if f == nil {
		return ""
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.commitModes) == 0 {
		return ""
	}
	mode := f.commitModes[0]
	f.commitModes = append([]string(nil), f.commitModes[1:]...)
	return mode
}

func (f *stateTestFaults) sessionAcquired() {
	if f == nil {
		return
	}
	f.mu.Lock()
	f.acquired++
	f.active++
	if f.active > f.maxActive {
		f.maxActive = f.active
	}
	hook := f.onAcquire
	f.mu.Unlock()
	if hook != nil {
		hook()
	}
}

func (f *stateTestFaults) sessionReleased() {
	if f == nil {
		return
	}
	f.mu.Lock()
	f.released++
	f.active--
	hook := f.onRelease
	f.mu.Unlock()
	if hook != nil {
		hook()
	}
}

func (f *stateTestFaults) lifecycle() (acquired, released, maxActive int) {
	if f == nil {
		return 0, 0, 0
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.acquired, f.released, f.maxActive
}

type liveStateTestStore struct {
	store    *extensionstore.Store
	faults   *stateTestFaults
	sequence atomic.Uint64
}

func newLiveStateTestStore(store *extensionstore.Store, faults *stateTestFaults) *liveStateTestStore {
	return &liveStateTestStore{store: store, faults: faults}
}

func (s *liveStateTestStore) WithProfileSession(ctx context.Context, profileID string, lockTimeout time.Duration, operation func(ProfileSession) error) error {
	if s.faults != nil && s.faults.onAttempt != nil {
		s.faults.onAttempt()
	}
	err := s.store.WithProfileLock(ctx, profileID, lockTimeout, func(session *extensionstore.Session) error {
		s.faults.sessionAcquired()
		defer s.faults.sessionReleased()
		return operation(&liveStateTestSession{
			owner:     s,
			session:   session,
			profileID: profileID,
		})
	})
	if errors.Is(err, extensionstore.ErrProfileLockTimeout) {
		return ErrStateMigrationLockTimeout
	}
	return err
}

type liveStateTestSession struct {
	owner     *liveStateTestStore
	session   *extensionstore.Session
	profileID string
}

func (s *liveStateTestSession) Snapshot(ctx context.Context, profileID, lineageID string, familyIDs []string) (StateSnapshot, error) {
	transaction, err := s.Begin(ctx, profileID, familyIDs)
	if err != nil {
		return StateSnapshot{}, err
	}
	metadata, err := transaction.StateMetadata(ctx, profileID)
	if err != nil {
		_, _ = transaction.Rollback(context.WithoutCancel(ctx))
		return StateSnapshot{}, err
	}
	ledger, err := transaction.MigrationLedger(ctx, profileID, lineageID)
	if err != nil {
		_, _ = transaction.Rollback(context.WithoutCancel(ctx))
		return StateSnapshot{}, err
	}
	counts, err := transaction.FamilyCounts(ctx, familyIDs)
	if err != nil {
		_, _ = transaction.Rollback(context.WithoutCancel(ctx))
		return StateSnapshot{}, err
	}
	outcome, rollbackErr := transaction.Rollback(context.WithoutCancel(ctx))
	if outcome == CommitIndeterminate {
		return StateSnapshot{}, fmt.Errorf("snapshot rollback indeterminate: %w", rollbackErr)
	}
	if s.owner.faults != nil {
		s.owner.faults.mu.Lock()
		s.owner.faults.snapshots++
		ordinal := s.owner.faults.snapshots
		hook := s.owner.faults.onSnapshot
		s.owner.faults.mu.Unlock()
		if hook != nil {
			hook(ordinal)
		}
	}
	return StateSnapshot{Metadata: metadata, FamilyCounts: counts, Ledger: ledger}, rollbackErr
}

func (s *liveStateTestSession) Begin(ctx context.Context, profileID string, familyIDs []string) (StateTransaction, error) {
	if profileID != s.profileID {
		return nil, errors.New("test state profile scope violation")
	}
	tx, err := s.session.Begin(ctx)
	if err != nil {
		return nil, err
	}
	families := append([]string(nil), familyIDs...)
	sort.Strings(families)
	return &liveStateTestTransaction{
		tx:           tx,
		profileID:    profileID,
		familyIDs:    families,
		capabilityID: fmt.Sprintf("test-state:%s:%d", profileID, s.owner.sequence.Add(1)),
		faults:       s.owner.faults,
	}, nil
}

type liveStateTestTransaction struct {
	tx           *extensionstore.Tx
	profileID    string
	familyIDs    []string
	capabilityID string
	faults       *stateTestFaults
}

func (t *liveStateTestTransaction) CapabilityID() string {
	return t.capabilityID
}

func (t *liveStateTestTransaction) StateFamilyIDs() []string {
	return append([]string(nil), t.familyIDs...)
}

func (t *liveStateTestTransaction) IsStateWriteCapability() bool {
	return true
}

func (t *liveStateTestTransaction) FamilyCounts(ctx context.Context, familyIDs []string) (map[string]int64, error) {
	families := append([]string(nil), familyIDs...)
	sort.Strings(families)
	if !sameTestStrings(families, t.familyIDs) {
		return nil, errors.New("test state family scope violation")
	}
	return t.tx.FamilyCounts(ctx, families)
}

func (t *liveStateTestTransaction) StateMetadata(ctx context.Context, profileID string) (*StateMetadata, error) {
	if profileID != t.profileID {
		return nil, errors.New("test state metadata scope violation")
	}
	metadata, err := t.tx.StateMetadata(ctx, profileID)
	if err != nil || metadata == nil {
		return nil, err
	}
	return &StateMetadata{
		ProfileID:          metadata.ProfileID,
		MigrationLineageID: metadata.MigrationLineageID,
		StateVersion:       metadata.StateVersion,
		LastMigrationID:    cloneTestString(metadata.LastMigrationID),
		MetadataVersion:    metadata.MetadataVersion,
		CreatedAt:          metadata.CreatedAt,
		UpdatedAt:          metadata.UpdatedAt,
	}, nil
}

func (t *liveStateTestTransaction) MigrationLedger(ctx context.Context, profileID, lineageID string) ([]MigrationLedgerEntry, error) {
	if profileID != t.profileID {
		return nil, errors.New("test state ledger scope violation")
	}
	entries, err := t.tx.MigrationLedger(ctx, profileID, lineageID)
	if err != nil {
		return nil, err
	}
	result := make([]MigrationLedgerEntry, len(entries))
	for index, entry := range entries {
		result[index] = MigrationLedgerEntry{
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

func (t *liveStateTestTransaction) InsertStateMetadata(ctx context.Context, metadata StateMetadata) error {
	return t.tx.InsertStateMetadata(ctx, extensionstore.StateMetadata{
		ProfileID:          metadata.ProfileID,
		MigrationLineageID: metadata.MigrationLineageID,
		StateVersion:       metadata.StateVersion,
		LastMigrationID:    cloneTestString(metadata.LastMigrationID),
		MetadataVersion:    metadata.MetadataVersion,
		CreatedAt:          metadata.CreatedAt,
		UpdatedAt:          metadata.UpdatedAt,
	})
}

func (t *liveStateTestTransaction) UpdateStateMetadata(ctx context.Context, before StateMetadata, stateVersion int, migrationID string, now time.Time) error {
	return t.tx.UpdateStateMetadata(ctx, extensionstore.StateMetadata{
		ProfileID:          before.ProfileID,
		MigrationLineageID: before.MigrationLineageID,
		StateVersion:       before.StateVersion,
		LastMigrationID:    cloneTestString(before.LastMigrationID),
		MetadataVersion:    before.MetadataVersion,
		CreatedAt:          before.CreatedAt,
		UpdatedAt:          before.UpdatedAt,
	}, stateVersion, migrationID, now)
}

func (t *liveStateTestTransaction) InsertMigrationLedger(ctx context.Context, entry MigrationLedgerEntry) error {
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

func (t *liveStateTestTransaction) Commit(ctx context.Context) (CommitOutcome, error) {
	switch t.faults.nextCommitMode() {
	case "committed_with_error":
		outcome, err := t.tx.Commit(ctx)
		if outcome != extensionstore.CommitProven || err != nil {
			return CommitIndeterminate, err
		}
		return CommitCommitted, errors.New("lost commit acknowledgment")
	case "absent":
		_, _ = t.tx.Rollback(context.WithoutCancel(ctx))
		return CommitAbsent, errors.New("commit was proven absent")
	case "indeterminate":
		_, _ = t.tx.Commit(ctx)
		return CommitIndeterminate, errors.New("commit acknowledgment and proof unavailable")
	default:
		outcome, err := t.tx.Commit(ctx)
		return mapTestCommitOutcome(outcome), err
	}
}

func (t *liveStateTestTransaction) Rollback(ctx context.Context) (CommitOutcome, error) {
	outcome, err := t.tx.Rollback(ctx)
	return mapTestCommitOutcome(outcome), err
}

// PutTestMember is an owner-local typed mutation used only by the synthetic
// profile in state-admission tests. Extensions receives only the base scoped
// capability and has no SQL or physical-table access.
func (t *liveStateTestTransaction) PutTestMember(ctx context.Context, memberID string) error {
	if t.profileID != "test_profile" || !sameTestStrings(t.familyIDs, []string{testStateFamilyID}) {
		return errors.New("test state write capability scope violation")
	}
	return t.tx.Exec(ctx, `
INSERT INTO extension_state_test_members (member_id)
VALUES ($1)
ON CONFLICT (member_id) DO NOTHING
`, memberID)
}

func mapTestCommitOutcome(outcome extensionstore.CommitOutcome) CommitOutcome {
	switch outcome {
	case extensionstore.CommitProven:
		return CommitCommitted
	case extensionstore.CommitAbsent:
		return CommitAbsent
	default:
		return CommitIndeterminate
	}
}

func sameTestStrings(left, right []string) bool {
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

func cloneTestString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
