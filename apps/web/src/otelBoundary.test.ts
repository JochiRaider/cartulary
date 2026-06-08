import { afterEach, describe, expect, it, vi } from "vitest";

describe("OpenTelemetry browser boundary", () => {
  afterEach(() => {
    localStorage.clear();
    sessionStorage.clear();
    document.documentElement.removeAttribute(
      "data-telemetry-exporter-endpoint",
    );
    window.history.replaceState({}, "", "/");
    vi.unstubAllGlobals();
  });

  it("keeps browser state from creating telemetry exporters or remote telemetry requests", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    localStorage.setItem(
      "telemetry.exporter.endpoint",
      "https://collector.example.test:4318/v1/traces",
    );
    sessionStorage.setItem(
      "OTEL_EXPORTER_OTLP_ENDPOINT",
      "http://localhost:4318",
    );
    document.documentElement.setAttribute(
      "data-telemetry-exporter-endpoint",
      "https://collector.example.test:4318/v1/metrics",
    );
    window.history.replaceState(
      {},
      "",
      "/?telemetry.exporter.endpoint=https%3A%2F%2Fcollector.example.test%3A4318",
    );

    await Promise.resolve();

    expect(fetchMock).not.toHaveBeenCalled();
    for (const globalName of forbiddenTelemetryGlobals) {
      expect(
        (globalThis as Record<string, unknown>)[globalName],
      ).toBeUndefined();
    }
  });
});

const forbiddenTelemetryGlobals = [
  "__OTEL_EXPORTER__",
  "__OTEL_TRACER_PROVIDER__",
  "__OTEL_METER_PROVIDER__",
  "__OTEL_LOGGER_PROVIDER__",
  "__CARTULARY_TELEMETRY_CONFIG__",
  "OpenTelemetry",
];
