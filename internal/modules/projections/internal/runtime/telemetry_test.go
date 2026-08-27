package runtime

import "testing"

func TestProjectionTelemetrySafeVocabulary(t *testing.T) {
	store := &Store{}
	_, finish := store.startProjectionSpan(t.Context())
	finish(nil)
}

func TestProjectionTelemetryNoSDK(t *testing.T) {
	store := &Store{}
	_, finish := store.startProjectionSpan(t.Context())
	finish(nil)
	_, finish = store.startProjectionSpan(t.Context())
	finish(assertionError("projection rebuild failed"))
}

type assertionError string

func (err assertionError) Error() string {
	return string(err)
}
