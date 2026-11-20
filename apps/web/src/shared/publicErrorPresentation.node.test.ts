import { describe, expect, it } from "vitest";
import { resolvePublicErrorPresentation } from "./publicErrorPresentation";

describe("public error presentation", () => {
  it("uses structured code and operation context for every current locus", () => {
    const cases = [
      {
        expectedFamily: "local_validation",
        input: {
          code: "invalid_mutation_payload",
          hasAuthorizedMaterialization: true,
          operationFamily: "field_mutation",
          status: 422,
        },
      },
      {
        expectedFamily: "same_field_conflict",
        input: {
          code: "same_field_conflict",
          hasAuthorizedMaterialization: true,
          operationFamily: "field_mutation",
          status: 409,
        },
      },
      {
        expectedFamily: "client_txn_conflict",
        input: {
          code: "client_txn_conflict",
          hasAuthorizedMaterialization: true,
          operationFamily: "field_mutation",
          status: 409,
        },
      },
      {
        expectedFamily: "queue_overflow",
        input: {
          code: "pending_queue_capacity_exceeded",
          hasAuthorizedMaterialization: true,
          operationFamily: "field_mutation",
          status: 409,
        },
      },
      {
        expectedFamily: "stale_refresh",
        input: {
          code: "transport_unavailable",
          hasAuthorizedMaterialization: true,
          operationFamily: "surface_load",
          status: 503,
        },
      },
      {
        expectedFamily: "initial_load_failure",
        input: {
          code: "transport_unavailable",
          hasAuthorizedMaterialization: false,
          operationFamily: "surface_load",
          status: 503,
        },
      },
      {
        expectedFamily: "authentication_required",
        input: {
          code: "session_required",
          hasAuthorizedMaterialization: true,
          operationFamily: "authentication",
          status: 401,
        },
      },
      // Either trigger alone presents authentication_required.
      {
        expectedFamily: "authentication_required",
        input: {
          code: "invalid_credentials",
          hasAuthorizedMaterialization: true,
          operationFamily: "surface_refresh",
          status: 401,
        },
      },
      {
        expectedFamily: "authentication_required",
        input: {
          code: "session_required",
          hasAuthorizedMaterialization: true,
          operationFamily: "surface_refresh",
          status: 400,
        },
      },
      ...(["authorization_denied", "incident_not_found"] as const).map(
        (code) => ({
          expectedFamily: "permission_or_incident_access_loss" as const,
          input: {
            code,
            hasAuthorizedMaterialization: true,
            operationFamily: "surface_refresh" as const,
            status: 403,
          },
        }),
      ),
      {
        expectedFamily: "extension_unavailable",
        input: {
          code: "internal_error",
          hasAuthorizedMaterialization: true,
          operationFamily: "extension",
          status: 503,
        },
      },
      {
        expectedFamily: "evidence_preview_blocked",
        input: {
          code: "evidence_access_unavailable",
          reasonCode: "unsupported_preview",
          hasAuthorizedMaterialization: true,
          operationFamily: "evidence_preview",
          status: 409,
        },
      },
      {
        expectedFamily: "unknown_future_error",
        input: {
          code: "internal_error",
          hasAuthorizedMaterialization: true,
          operationFamily: "field_mutation",
          status: 500,
        },
      },
    ] as const;

    for (const operationFamily of [
      "evidence_preview",
      "evidence_download",
    ] as const) {
      expect(
        resolvePublicErrorPresentation({
          code: "object_store_unavailable",
          hasAuthorizedMaterialization: true,
          operationFamily,
          status: 503,
        }),
      ).toMatchObject({
        family: "unknown_future_error",
        actions: [],
        live: "assertive",
      });
    }
    for (const testCase of cases) {
      expect(resolvePublicErrorPresentation(testCase.input).family).toBe(
        testCase.expectedFamily,
      );
    }
  });

  it("selects a retry family for service_unavailable in every operation family", () => {
    const families = [
      "authentication",
      "evidence_preview",
      "evidence_download",
      "extension",
      "field_mutation",
      "surface_load",
      "surface_refresh",
    ] as const;
    for (const operationFamily of families)
      for (const hasAuthorizedMaterialization of [false, true]) {
        const presentation = resolvePublicErrorPresentation({
          code: "service_unavailable",
          hasAuthorizedMaterialization,
          operationFamily,
          reasonCode: "admission_paused",
          status: 503,
        });
        expect(presentation.family).toBe(
          operationFamily === "surface_load" && !hasAuthorizedMaterialization
            ? "initial_load_failure"
            : "stale_refresh",
        );
        expect(presentation.actions).toEqual([
          presentation.family === "initial_load_failure"
            ? "retry_after_failure"
            : "retry",
        ]);
      }
  });

  it("falls back without accepting or inspecting human message text", () => {
    expect(
      resolvePublicErrorPresentation({
        code: "internal_error",
        hasAuthorizedMaterialization: false,
        operationFamily: "field_mutation",
        status: 500,
      }),
    ).toMatchObject({
      actions: [],
      family: "unknown_future_error",
      retention: "retain_only_verified_authorized_data",
    });
  });
});
