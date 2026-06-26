package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestAccessorsNoSDKRegisteredScopes(t *testing.T) {
	for _, scope := range []Scope{
		ScopeHTTPAPI,
		ScopeWorkbook,
		ScopeCollaboration,
		ScopeJobs,
		ScopePostgres,
		ScopeObjectStore,
		ScopeTelemetry,
	} {
		if !RegisteredScope(scope) {
			t.Fatalf("scope %q is not registered", scope)
		}
		ctx, span := Tracer(scope, "").Start(context.Background(), "test")
		span.End()
		meter := Meter(scope, "")
		histogram, err := meter.Float64Histogram("cartulary.test.duration")
		if err != nil {
			t.Fatalf("create no-SDK histogram for %s: %v", scope, err)
		}
		histogram.Record(ctx, 0.001)

		logger := Logger(scope, "")
		if logger.ScopeName() != string(scope) {
			t.Fatalf("logger scope mismatch: got %q want %q", logger.ScopeName(), scope)
		}
		if logger.InstrumentationVersion() != VersionUnknown {
			t.Fatalf("logger version mismatch: got %q", logger.InstrumentationVersion())
		}
		if logger.SchemaURL() != "" || len(logger.ScopeAttributes()) != 0 {
			t.Fatalf("logger scope must use null schema URL and empty attributes")
		}
		if logger.Enabled() {
			t.Fatalf("no-SDK logger for %s must be disabled", scope)
		}
		logger.Emit(ctx, LogRecord{Body: "safe local diagnostic", Severity: "info"})
	}
}

func TestAccessorsUnknownScopeFallsBackToNoopTelemetryScope(t *testing.T) {
	unknown := Scope("cartulary.unregistered")
	if RegisteredScope(unknown) {
		t.Fatal("test scope should remain unregistered")
	}

	ctx, span := Tracer(unknown, "  ").Start(context.Background(), "unknown")
	if span.SpanContext().IsValid() {
		t.Fatal("unknown scope should use a no-op tracer without creating a valid span context")
	}
	span.End()

	counter, err := Meter(unknown, "").Int64Counter("cartulary.test.unknown")
	if err != nil {
		t.Fatalf("unknown scope should return no-op/API-safe meter: %v", err)
	}
	counter.Add(ctx, 1)

	logger := Logger(unknown, "")
	if logger.ScopeName() != string(ScopeTelemetry) {
		t.Fatalf("unknown scope should fall back to telemetry logger scope, got %q", logger.ScopeName())
	}
	if logger.InstrumentationVersion() != VersionUnknown || logger.SchemaURL() != "" || len(logger.ScopeAttributes()) != 0 {
		t.Fatalf("unknown logger fallback must retain no-SDK scope defaults")
	}
	if logger.Enabled() {
		t.Fatal("unknown logger fallback must be disabled in no-SDK mode")
	}
	logger.Emit(ctx, LogRecord{Body: "ignored"})
}

func TestAccessorsConcurrentNoSDKUse(t *testing.T) {
	scopes := []Scope{
		ScopeHTTPAPI,
		ScopeWorkbook,
		ScopeCollaboration,
		ScopeJobs,
		ScopePostgres,
		ScopeObjectStore,
		ScopeTelemetry,
		Scope("cartulary.unregistered"),
	}

	var wg sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 64; i++ {
				for _, scope := range scopes {
					ctx, span := Tracer(scope, "1.2.3").Start(context.Background(), "concurrent")
					span.End()
					counter, err := Meter(scope, "1.2.3").Int64Counter("cartulary.test.concurrent")
					if err != nil {
						t.Errorf("create counter for scope %s: %v", scope, err)
						return
					}
					counter.Add(ctx, 1)
					logger := Logger(scope, "1.2.3")
					if logger.InstrumentationVersion() != "1.2.3" {
						t.Errorf("logger version for scope %s: got %q", scope, logger.InstrumentationVersion())
						return
					}
					logger.Emit(ctx, LogRecord{Body: "concurrent", Severity: "debug"})
				}
			}
		}()
	}
	wg.Wait()
}

func TestHTTPMiddlewareNoSDK(t *testing.T) {
	handler := HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/incidents/10000000-0000-0000-0000-000000000001/views/cartulary.view.timeline.v2/query" {
			t.Fatalf("handler saw unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}), VersionUnknown)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/10000000-0000-0000-0000-000000000001/views/cartulary.view.timeline.v2/query?secret=value", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("middleware changed status: got %d", rec.Code)
	}

	routeTemplate, routeFamily := safeHTTPRoute(req.URL.Path)
	if strings.Contains(routeTemplate, "10000000") || strings.Contains(routeTemplate, "cartulary.view.timeline") {
		t.Fatalf("safe route leaked concrete identifiers: %q", routeTemplate)
	}
	if routeFamily != "incidents" {
		t.Fatalf("unexpected route family: got %q", routeFamily)
	}
}
