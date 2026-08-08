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
      {
        expectedFamily: "permission_or_incident_access_loss",
        input: {
          code: "incident_access_lost",
          hasAuthorizedMaterialization: true,
          operationFamily: "surface_refresh",
          status: 403,
        },
      },
      {
        expectedFamily: "extension_unavailable",
        input: {
          code: "future_extension_failure",
          hasAuthorizedMaterialization: true,
          operationFamily: "extension",
          status: 503,
        },
      },
      {
        expectedFamily: "evidence_preview_blocked",
        input: {
          code: "future_preview_failure",
          hasAuthorizedMaterialization: true,
          operationFamily: "evidence_preview",
          status: 503,
        },
      },
      {
        expectedFamily: "unknown_future_error",
        input: {
          code: "future_operation_failure",
          hasAuthorizedMaterialization: true,
          operationFamily: "field_mutation",
          status: 500,
        },
      },
    ] as const;

    for (const testCase of cases) {
      expect(resolvePublicErrorPresentation(testCase.input).family).toBe(
        testCase.expectedFamily,
      );
    }
  });

  it("falls back without accepting or inspecting human message text", () => {
    expect(
      resolvePublicErrorPresentation({
        code: "future_error",
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
