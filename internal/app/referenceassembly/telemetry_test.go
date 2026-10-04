package referenceassembly_test

import (
	"context"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry/testsupport"
)

func TestReferencePackObserverEmitsOnlyClosedTokens_Unit(t *testing.T) {
	ctx := context.Background()
	capture := testsupport.StartCapture()
	defer capture.Close(ctx)
	observer := referenceassembly.NewOperationObserver("1.2.3")
	names := reference_data.ReferenceOperationNames()
	for _, name := range names {
		_, end := observer.BeginReferenceOperation(ctx, name)
		end("success")
		end("failed")
	}
	_, end := observer.BeginReferenceOperation(ctx, "reference_pack.lookup/password=private")
	end("success")
	_, end = observer.BeginReferenceOperation(ctx, "reference_pack.lookup")
	end("https://user:secret@example.invalid/payload")
	spans := capture.EndedSpans()
	if len(spans) != len(names)+1 {
		t.Fatal("missing, duplicate or unregistered spans", len(spans))
	}
	for _, span := range spans {
		if !slices.Contains(names, span.Name) || len(span.Attributes) != 2 || span.Attributes["cartulary.operation"] != span.Name {
			t.Fatal("unsafe span shape", span)
		}
		if result := span.Attributes["cartulary.result"]; result != "success" && result != "failed" {
			t.Fatal("unsafe result", result)
		}
	}
	points, err := capture.MetricPoints(ctx)
	if err != nil || len(points) != len(names)+1 {
		t.Fatal("missing metrics", len(points), err)
	}
	for _, point := range points {
		if point.Name != "cartulary.reference_pack.operation.duration" || len(point.Attributes) != 2 || !slices.Contains(names, point.Attributes["cartulary.operation"]) || !point.IsFloat || point.FloatValue < 0 {
			t.Fatal("unsafe metric shape", point)
		}
	}
}
