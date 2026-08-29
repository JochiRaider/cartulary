package stream

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestPostgresStreamAndDispatcherRejectIncompleteConstruction_Unit(t *testing.T) {
	if stream, err := NewPostgresStream(nil); err == nil || stream != nil {
		t.Fatalf("PostgresStream = %#v, error = %v; want missing database rejection", stream, err)
	}
	db := struct{ postgres.DB }{}
	stream, err := NewPostgresStream(db)
	if err != nil {
		t.Fatal(err)
	}
	broadcaster := replayBroadcasterFunc(func(protocol.Message) error { return nil })
	tests := []struct {
		name      string
		store     *PostgresStream
		broadcast replayBroadcaster
		now       func() time.Time
		version   string
		onLoss    func()
	}{
		{name: "stream", broadcast: broadcaster, now: time.Now, version: "test", onLoss: func() {}},
		{name: "broadcaster", store: stream, now: time.Now, version: "test", onLoss: func() {}},
		{name: "clock", store: stream, broadcast: broadcaster, version: "test", onLoss: func() {}},
		{name: "service version", store: stream, broadcast: broadcaster, now: time.Now, onLoss: func() {}},
		{name: "loss callback", store: stream, broadcast: broadcaster, now: time.Now, version: "test"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dispatcher, err := NewDispatcher(test.store, test.broadcast, test.now, test.version, test.onLoss)
			if err == nil || dispatcher != nil {
				t.Fatalf("Dispatcher = %#v, error = %v; want incomplete construction rejection", dispatcher, err)
			}
		})
	}
	dispatcher, err := NewDispatcher(stream, broadcaster, time.Now, "1.2.3-test", func() {})
	if err != nil {
		t.Fatal(err)
	}
	if dispatcher.serviceVersion != "1.2.3-test" {
		t.Fatalf("dispatcher service version = %q", dispatcher.serviceVersion)
	}
}

type replayBroadcasterFunc func(protocol.Message) error

func (broadcast replayBroadcasterFunc) DeliverReplayable(message protocol.Message) error {
	return broadcast(message)
}

func TestDispatcherSupervisorClassifiesGracefulReturnUnexpectedReturnAndPanic_Unit(t *testing.T) {
	newDispatcher := func(losses *atomic.Int32) *Dispatcher {
		return &Dispatcher{onUnexpectedLoss: func() { losses.Add(1) }}
	}

	t.Run("graceful cancellation", func(t *testing.T) {
		var losses atomic.Int32
		dispatcher := newDispatcher(&losses)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		done := make(chan struct{})
		dispatcher.supervise(ctx, done, func(context.Context) {})
		if losses.Load() != 0 || dispatcher.terminal {
			t.Fatalf("graceful return reported loss=%d terminal=%t", losses.Load(), dispatcher.terminal)
		}
	})

	t.Run("unexpected return", func(t *testing.T) {
		var losses atomic.Int32
		dispatcher := newDispatcher(&losses)
		done := make(chan struct{})
		dispatcher.supervise(context.Background(), done, func(context.Context) {})
		dispatcher.signalUnexpectedLoss()
		if losses.Load() != 1 || !dispatcher.terminal {
			t.Fatalf("unexpected return reported loss=%d terminal=%t", losses.Load(), dispatcher.terminal)
		}
	})

	t.Run("panic", func(t *testing.T) {
		var losses atomic.Int32
		dispatcher := newDispatcher(&losses)
		done := make(chan struct{})
		dispatcher.supervise(context.Background(), done, func(context.Context) { panic(errors.New("boom")) })
		dispatcher.signalUnexpectedLoss()
		if losses.Load() != 1 || !dispatcher.terminal {
			t.Fatalf("panic reported loss=%d terminal=%t", losses.Load(), dispatcher.terminal)
		}
	})
}
