package projections

import "testing"

func TestProjectionTelemetrySafeVocabulary(t *testing.T) {
	if got := safeProjectionViewSchemaID(timelineViewSchemaID); got != timelineViewSchemaID {
		t.Fatalf("timeline view schema = %q", got)
	}
	if got := safeProjectionViewSchemaID("cartulary.view.timeline.v1/10000000-0000-4000-8000-000000000001"); got != "unknown" {
		t.Fatalf("unsafe projection view schema = %q", got)
	}
}

func TestProjectionTelemetryNoSDK(t *testing.T) {
	store := NewStore(nil)
	_, finish := store.startProjectionSpan(t.Context(), timelineViewSchemaID)
	finish(nil)
	_, finish = store.startProjectionSpan(t.Context(), "incident/10000000")
	finish(assertionError("projection rebuild failed"))
}

type assertionError string

func (err assertionError) Error() string {
	return string(err)
}
