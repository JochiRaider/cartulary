package projections

import "testing"

func TestProjectionTelemetrySafeVocabulary(t *testing.T) {
	if got := safeProjectionViewSchemaID(timelineViewSchemaID); got != timelineViewSchemaID {
		t.Fatalf("timeline view schema = %q", got)
	}
	for viewSchemaID := range querySurfacesForTest() {
		if got := safeProjectionViewSchemaID(viewSchemaID); got != viewSchemaID {
			t.Fatalf("generic projection view schema %s sanitized to %q", viewSchemaID, got)
		}
	}
	if got := safeProjectionViewSchemaID("cartulary.view.timeline.v2/10000000-0000-4000-8000-000000000001"); got != "unknown" {
		t.Fatalf("unsafe projection view schema = %q", got)
	}
}

func TestProjectionTelemetryNoSDK(t *testing.T) {
	store := NewStore(nil, nil)
	_, finish := store.startProjectionSpan(t.Context(), timelineViewSchemaID)
	finish(nil)
	_, finish = store.startProjectionSpan(t.Context(), "incident/10000000")
	finish(assertionError("projection rebuild failed"))
}

type assertionError string

func (err assertionError) Error() string {
	return string(err)
}
