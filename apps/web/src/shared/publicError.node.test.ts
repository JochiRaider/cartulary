import { errorRegistry } from "@cartulary/protocol-ts/errors";
import { networkFlowErrorRegistry } from "@cartulary/protocol-ts/network-flow";
import { describe, expect, it } from "vitest";
import {
  clientErrorCodes,
  isTemporaryUnavailability,
  isUnappliedRefusal,
  notAppliedText,
  publicErrorView,
  stillUnconfirmedText,
} from "./publicError";

describe("publicError normalization", () => {
  it("retains structured codes, safe messages, and allowlisted details", () => {
    expect(
      publicErrorView({
        code: "authorization_denied",
        details: {
          reason_code: "incident_not_found",
          internal_context: "must not escape",
        },
        message: "Access denied.",
        status: 403,
      }),
    ).toEqual({
      code: "authorization_denied",
      details: [
        {
          key: "reason_code",
          label: "Reason",
          value: "incident_not_found",
        },
      ],
      status: 403,
      statusText: "Access denied.",
    });
  });

  it("replaces unsafe messages and details with status-safe presentation", () => {
    expect(
      publicErrorView({
        code: "import_failed",
        details: { reason_code: "stack trace at /home/service/import.go" },
        message: "SELECT secret FROM credentials at /home/service/db.go",
        status: 500,
      }),
    ).toEqual({
      code: "import_failed",
      details: [],
      status: 500,
      statusText: "Request failed.",
    });
  });

  it("presents a declared service_unavailable as temporary unavailability without echoing its reason", () => {
    for (const reason_code of [
      "extension_publication_pending",
      "admission_paused",
      "shutdown_in_progress",
      "extension_integrity_failure",
    ]) {
      const error = {
        code: "service_unavailable",
        details: { reason_code },
        message: "service unavailable",
        retryable: true,
        status: 503,
      };
      expect(isTemporaryUnavailability(error)).toBe(true);
      expect(publicErrorView(error)).toEqual({
        code: "service_unavailable",
        details: [],
        status: 503,
        statusText: "Cartulary is temporarily unavailable. Try again shortly.",
      });
      expect(JSON.stringify(publicErrorView(error))).not.toContain(reason_code);
    }
    // Neither another 503 code nor a contract failure is temporary unavailability.
    expect(
      isTemporaryUnavailability({
        code: "authentication_capacity_exhausted",
        status: 503,
      }),
    ).toBe(false);
    expect(
      isTemporaryUnavailability({
        code: "invalid_public_contract_response",
        status: 502,
      }),
    ).toBe(false);
    expect(
      isTemporaryUnavailability({ code: "service_unavailable", status: 502 }),
    ).toBe(false);
  });
});

describe("unapplied refusal", () => {
  it("is exactly a decoded 503 service_unavailable and has not-applied and still-unconfirmed copy", () => {
    // Core 01 §3.3.6.1–3.3.6.2: every service_unavailable reason refused
    // admission or proved the commit absent (Core 03 REQ-03-086a).
    const refused = { code: "service_unavailable", status: 503 };
    expect(isUnappliedRefusal(503, refused)).toBe(true);
    expect(isUnappliedRefusal(503, { code: "service_unavailable" })).toBe(true);
    for (const [status, error] of [
      [500, { code: "internal_error", status: 500 }],
      [503, { code: "authentication_capacity_exhausted", status: 503 }],
      [502, { code: "service_unavailable", status: 502 }],
      [502, refused],
      [null, refused],
      [503, null],
    ] as const)
      expect(isUnappliedRefusal(status, error)).toBe(false);
    expect(notAppliedText()).toBe(
      "Cartulary is temporarily unavailable. This action was not applied; try again shortly.",
    );
    expect(notAppliedText("Retry")).toBe(
      "Cartulary is temporarily unavailable. This action was not applied; choose Retry to try again.",
    );
    expect(stillUnconfirmedText()).toBe(
      "Cartulary is temporarily unavailable. The earlier outcome is still unconfirmed; try recovery again shortly.",
    );
  });
});

describe("client-local public error codes", () => {
  it("form one closed sorted set disjoint from every registered server error code", () => {
    const registeredCodes = new Set<string>([
      ...errorRegistry.errors.map((entry) => entry.code),
      ...networkFlowErrorRegistry.errors.map((entry) => entry.code),
    ]);
    expect(registeredCodes.has("internal_error")).toBe(true);
    expect(clientErrorCodes.length).toBeGreaterThan(0);
    expect([...new Set(clientErrorCodes)].sort()).toEqual([
      ...clientErrorCodes,
    ]);
    expect(
      clientErrorCodes.every((code) => /^[a-z][a-z0-9_]*$/u.test(code)),
    ).toBe(true);
    expect(
      clientErrorCodes.filter((code) => registeredCodes.has(code)),
    ).toEqual([]);
  });
});
