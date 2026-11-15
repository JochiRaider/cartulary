package runtime

import (
	"reflect"
	"testing"

	telemetrytest "github.com/JochiRaider/cartulary/internal/platform/telemetry/testsupport"
)

func TestProjectionTelemetrySafeVocabulary(t *testing.T) {
	capture := telemetrytest.StartCapture()
	defer capture.Close(t.Context())
	store := &Store{}
	_, finish := store.startProjectionSpan(t.Context())
	finish(nil)
	spans := capture.EndedSpans()
	want := map[string]string{
		"cartulary.view_schema_id": "unknown",
		"cartulary.operation":      "rebuild",
		"cartulary.result":         "success",
	}
	if len(spans) != 1 || spans[0].Name != "cartulary.workbook.projection" || !reflect.DeepEqual(spans[0].Attributes, want) {
		t.Fatalf("projection span = %#v, want one cartulary.workbook.projection span with %#v", spans, want)
	}
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
