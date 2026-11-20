package stream

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

const (
	dispatchInterval   = 100 * time.Millisecond
	dispatchBackoffMax = 30 * time.Second
)

type replayBroadcaster interface {
	DeliverReplayable(protocol.Message) error
}

// ListenerSessionOpener opens the replay listener's dedicated database session
// outside the runtime pool. The listener owns and closes each session it opens.
type ListenerSessionOpener func(context.Context) (*pgx.Conn, error)

// Dispatcher owns bounded process orchestration. All durable reads and writes
// are delegated to PostgresStream.
type Dispatcher struct {
	store               *PostgresStream
	openListenerSession ListenerSessionOpener
	broadcaster         replayBroadcaster
	now                 func() time.Time
	serviceVersion      string
	onUnexpectedLoss    func()
	interval            time.Duration

	mu              sync.Mutex
	runMu           sync.Mutex
	cancel          context.CancelFunc
	done            chan struct{}
	tailCursor      map[uuid.UUID]int64
	tailInitialized bool
	terminal        bool
}

func NewDispatcher(
	store *PostgresStream,
	openListenerSession ListenerSessionOpener,
	broadcaster replayBroadcaster,
	now func() time.Time,
	serviceVersion string,
	onUnexpectedLoss func(),
) (*Dispatcher, error) {
	if store == nil || store.db == nil {
		return nil, errors.New("collaboration dispatcher stream is required")
	}
	if openListenerSession == nil {
		return nil, errors.New("collaboration dispatcher listener session is required")
	}
	if broadcaster == nil {
		return nil, errors.New("collaboration dispatcher broadcaster is required")
	}
	if now == nil {
		return nil, errors.New("collaboration dispatcher clock is required")
	}
	if strings.TrimSpace(serviceVersion) == "" {
		return nil, errors.New("collaboration dispatcher service version is required")
	}
	if onUnexpectedLoss == nil {
		return nil, errors.New("collaboration dispatcher loss callback is required")
	}
	return &Dispatcher{
		store:               store,
		openListenerSession: openListenerSession,
		broadcaster:         broadcaster,
		now:                 now,
		serviceVersion:      serviceVersion,
		onUnexpectedLoss:    onUnexpectedLoss,
		interval:            dispatchInterval,
		tailCursor:          make(map[uuid.UUID]int64),
	}, nil
}

func (dispatcher *Dispatcher) Prepare(parent context.Context, serving <-chan struct{}) error {
	if serving == nil {
		return errors.New("collaboration serving gate is required")
	}
	if dispatcher == nil || dispatcher.store == nil || dispatcher.store.db == nil || dispatcher.broadcaster == nil {
		return errors.New("collaboration dispatcher is not configured")
	}
	dispatcher.mu.Lock()
	if dispatcher.terminal {
		dispatcher.mu.Unlock()
		return errors.New("collaboration dispatcher is terminal")
	}
	if dispatcher.cancel != nil {
		dispatcher.mu.Unlock()
		return nil
	}
	dispatcher.runMu.Lock()
	cursor, err := dispatcher.store.SeedTailCursor(parent)
	if err != nil {
		dispatcher.runMu.Unlock()
		dispatcher.terminal = true
		dispatcher.mu.Unlock()
		return err
	}
	dispatcher.tailCursor = cursor
	dispatcher.tailInitialized = true
	dispatcher.runMu.Unlock()
	ctx, cancel := context.WithCancel(parent)
	dispatcher.cancel = cancel
	dispatcher.done = make(chan struct{})
	done := dispatcher.done
	dispatcher.mu.Unlock()
	go dispatcher.supervise(ctx, done, func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		case <-serving:
		}
		if ctx.Err() == nil {
			dispatcher.run(ctx)
		}
	})
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

func (dispatcher *Dispatcher) supervise(ctx context.Context, done chan<- struct{}, run func(context.Context)) {
	defer close(done)
	defer func() {
		if recovered := recover(); recovered != nil || ctx.Err() == nil {
			dispatcher.signalUnexpectedLoss()
		}
	}()
	run(ctx)
}

func (dispatcher *Dispatcher) signalUnexpectedLoss() {
	dispatcher.mu.Lock()
	if dispatcher.terminal {
		dispatcher.mu.Unlock()
		return
	}
	dispatcher.terminal = true
	cancel := dispatcher.cancel
	dispatcher.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	dispatcher.onUnexpectedLoss()
}

func (dispatcher *Dispatcher) run(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	notifications := make(chan struct{}, 1)
	listenerDone := make(chan struct{})
	go dispatcher.listenReplayNotifications(ctx, notifications, listenerDone)
	defer func() {
		cancel()
		<-listenerDone
	}()
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
			return
		case <-notifications:
			timer.Stop()
		case <-timer.C:
		}
	}
}

// listenReplayNotifications wakes the dispatcher when replay events are
// committed. The LISTEN lasts for the process lifetime, so, like the leases of
// Core 04 REQ-04-145, it runs on its own dedicated session outside the runtime
// pool and never reduces the capacity available to request or job work. If the
// listener fails, the dispatcher keeps polling.
func (dispatcher *Dispatcher) listenReplayNotifications(
	ctx context.Context,
	notifications chan<- struct{},
	done chan<- struct{},
) {
	defer close(done)
	if dispatcher.openListenerSession == nil {
		return
	}
	connection, err := dispatcher.openListenerSession(ctx)
	if err != nil {
		return
	}
	defer closeListenerSession(connection)
	if _, err := connection.Exec(ctx, `LISTEN cartulary_collaboration_replay`); err != nil {
		return
	}
	for {
		if _, err := connection.WaitForNotification(ctx); err != nil {
			return
		}
		select {
		case notifications <- struct{}{}:
		default:
		}
	}
}

// closeListenerSession closes the listener's session with one bounded
// terminate write, even after the dispatcher's context has ended.
func closeListenerSession(connection *pgx.Conn) {
	closeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = connection.Close(closeContext)
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
