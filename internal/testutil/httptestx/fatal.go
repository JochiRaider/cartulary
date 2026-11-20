package httptestx

import (
	"sync"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/runevidence"
)

// fatalSource is the read-only lifecycle fatal view of an in-process runtime.
type fatalSource interface {
	FatalDone() <-chan struct{}
	FatalReason() string
}

// watchServerFatal records an in-process test server's lifecycle fatal as one
// server_fatal contract evidence record (TH-HARNESS-REQ-863) carrying the
// reason code, the in-process kind, the work unit when known and the process
// ID. The run gate fails the run as server_fatal; a test that already failed
// keeps its own failure, and the record names the cause of any 503 the server
// sent afterwards. The returned stop ends the watch, is idempotent, and
// records a fatal that happened before it exactly once.
func watchServerFatal(source fatalSource, lookup func(string) (string, bool)) (stop func()) {
	done := source.FatalDone()
	if done == nil {
		return func() {}
	}
	record := func() {
		// Best effort, as for the contract guards: a run without an explicit
		// identity has no stream.
		_ = runevidence.RecordTo(lookup, runevidence.Evidence{Kind: runevidence.ServerFatal, ReasonCode: source.FatalReason(), Process: runevidence.InProcess})
	}
	quit, finished := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(finished)
		select {
		case <-done:
			record()
		case <-quit:
			// A fatal racing the stop is still recorded.
			select {
			case <-done:
				record()
			default:
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(quit)
			<-finished
		})
	}
}
