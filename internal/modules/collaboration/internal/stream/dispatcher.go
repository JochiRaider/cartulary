package stream

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

const (
	dispatchInterval   = 100 * time.Millisecond
	dispatchBackoffMax = 30 * time.Second
)

type replayBroadcaster interface {
	DeliverReplayable(protocol.Message) error
}

type notificationConnectionAcquirer interface {
	Acquire(context.Context) (*pgxpool.Conn, error)
}

// Dispatcher owns bounded process orchestration. All durable reads and writes
// are delegated to PostgresStream.
type Dispatcher struct {
	store       *PostgresStream
	acquirer    notificationConnectionAcquirer
	broadcaster replayBroadcaster
	now         func() time.Time
	interval    time.Duration

	mu              sync.Mutex
	runMu           sync.Mutex
	cancel          context.CancelFunc
	done            chan struct{}
	tailCursor      map[uuid.UUID]int64
	tailInitialized bool
}

func NewDispatcher(store *PostgresStream, broadcaster replayBroadcaster, now func() time.Time) *Dispatcher {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	dispatcher := &Dispatcher{
		store:       store,
		broadcaster: broadcaster,
		now:         now,
		interval:    dispatchInterval,
		tailCursor:  make(map[uuid.UUID]int64),
	}
	if store != nil {
		dispatcher.acquirer, _ = store.db.(notificationConnectionAcquirer)
	}
	return dispatcher
}

func (dispatcher *Dispatcher) Start(parent context.Context) error {
	if dispatcher == nil || dispatcher.store == nil || dispatcher.store.db == nil || dispatcher.broadcaster == nil {
		return errors.New("collaboration dispatcher is not configured")
	}
	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	if dispatcher.cancel != nil {
		return nil
	}
	dispatcher.runMu.Lock()
	cursor, err := dispatcher.store.SeedTailCursor(parent)
	if err != nil {
		dispatcher.runMu.Unlock()
		return err
	}
	dispatcher.tailCursor = cursor
	dispatcher.tailInitialized = true
	dispatcher.runMu.Unlock()
	ctx, cancel := context.WithCancel(parent)
	dispatcher.cancel = cancel
	dispatcher.done = make(chan struct{})
	go dispatcher.run(ctx, dispatcher.done)
	return nil
}

func (dispatcher *Dispatcher) Close(ctx context.Context) error {
	if dispatcher == nil {
		return nil
	}
	dispatcher.mu.Lock()
	cancel := dispatcher.cancel
	done := dispatcher.done
	dispatcher.cancel = nil
	dispatcher.done = nil
	dispatcher.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (dispatcher *Dispatcher) run(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	notifications := make(chan struct{}, 1)
	listenerDone := make(chan struct{})
	if dispatcher.acquirer != nil {
		go dispatcher.listenReplayNotifications(ctx, dispatcher.acquirer, notifications, listenerDone)
	} else {
		close(listenerDone)
	}
	backoff := dispatcher.interval
	for {
		_, err := dispatcher.RunOnce(ctx)
		delay := dispatcher.interval
		if err != nil {
			delay = backoff
			backoff *= 2
			if backoff > dispatchBackoffMax {
				backoff = dispatchBackoffMax
			}
		} else {
			backoff = dispatcher.interval
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			<-listenerDone
			return
		case <-notifications:
			timer.Stop()
		case <-timer.C:
		}
	}
}

func (dispatcher *Dispatcher) listenReplayNotifications(
	ctx context.Context,
	acquirer notificationConnectionAcquirer,
	notifications chan<- struct{},
	done chan<- struct{},
) {
	defer close(done)
	connection, err := acquirer.Acquire(ctx)
	if err != nil {
		return
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, `LISTEN cartulary_collaboration_replay`); err != nil {
		return
	}
	for {
		if _, err := connection.Conn().WaitForNotification(ctx); err != nil {
			return
		}
		select {
		case notifications <- struct{}{}:
		default:
		}
	}
}

func (dispatcher *Dispatcher) RunOnce(ctx context.Context) (processed int, runErr error) {
	if dispatcher == nil || dispatcher.store == nil || dispatcher.store.db == nil || dispatcher.broadcaster == nil {
		return 0, errors.New("collaboration dispatcher is not configured")
	}
	dispatcher.runMu.Lock()
	defer dispatcher.runMu.Unlock()
	defer func() {
		dispatcher.recordDispatcherRun(ctx, processed, runErr)
	}()
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if !dispatcher.tailInitialized {
		dispatcher.tailCursor = make(map[uuid.UUID]int64)
		dispatcher.tailInitialized = true
	}
	now := dispatcher.now().UTC()
	sequenced, err := dispatcher.store.SequencePending(ctx, now)
	if err != nil {
		return 0, err
	}
	messages, err := dispatcher.store.TailReplay(ctx, dispatcher.tailCursor)
	if err != nil {
		return sequenced, err
	}
	tailed := 0
	for _, message := range messages {
		if err := dispatcher.broadcaster.DeliverReplayable(message); err != nil {
			return sequenced + tailed, fmt.Errorf("deliver collaboration replay tail event: %w", err)
		}
		incidentID, _ := uuid.Parse(message.IncidentID)
		dispatcher.tailCursor[incidentID] = *message.StreamSeq
		tailed++
	}
	if err := dispatcher.store.PruneReplay(ctx, now); err != nil {
		return sequenced + tailed, err
	}
	return sequenced + tailed, nil
}
