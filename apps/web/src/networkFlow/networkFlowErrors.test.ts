import { describe, expect, it } from "vitest";
import {
  isNetworkFlowAuthorizationLoss,
  isNetworkFlowLifecycleLoss,
  isNetworkFlowProtectedStateLoss,
  networkFlowRequestError,
} from "./networkFlowErrors";

describe("Network Flow structured errors", () => {
  it("preserves status, code, reason, field, retry action, and a safe message", () => {
    const error = networkFlowRequestError(400, {
      error: {
        code: "network_flow_cursor_invalid",
        message: "The cursor is invalid.",
        details: { reason_code: "expired", field: "cursor_token" },
      },
    });

    expect(error).toMatchObject({
      status: 400,
      code: "network_flow_cursor_invalid",
      reasonCode: "expired",
      field: "cursor_token",
      retryAction: "restart_query",
      retryable: false,
      message: "The cursor is invalid.",
    });
  });

  it("uses exact structured authorization signals instead of message substrings", () => {
    const unrelated = networkFlowRequestError(500, {
      error: {
        code: "network_flow_graph_projection_failed",
        message: "authorization_denied appeared in an upstream note",
      },
    });
    const denied = networkFlowRequestError(403, {
      error: { code: "authorization_denied", message: "Access denied." },
    });

    expect(isNetworkFlowAuthorizationLoss(unrelated)).toBe(false);
    expect(isNetworkFlowAuthorizationLoss(denied)).toBe(true);
  });

  it("falls back to status-safe text when the server message is unsafe", () => {
    const error = networkFlowRequestError(500, {
      error: {
        code: "network_flow_graph_projection_failed",
        message: "stack trace at /home/service/query.go",
      },
    });
    expect(error.message).toBe("Request failed.");
  });

  it("classifies lifecycle and hidden-resource failures for fail-closed clearing", () => {
    const softDeleted = networkFlowRequestError(409, {
      error: {
        code: "network_flow_table_not_active",
        message: "The table is not active.",
        details: { reason_code: "soft_deleted" },
      },
    });
    const hidden = networkFlowRequestError(404, {
      error: {
        code: "network_flow_table_not_found",
        message: "The table was not found.",
      },
    });
    const transport = networkFlowRequestError(503, {
      error: { code: "network_flow_unavailable", message: "Unavailable." },
    });

    expect(isNetworkFlowLifecycleLoss(softDeleted)).toBe(true);
    expect(isNetworkFlowProtectedStateLoss(hidden)).toBe(true);
    expect(isNetworkFlowProtectedStateLoss(transport)).toBe(false);
  });
});
