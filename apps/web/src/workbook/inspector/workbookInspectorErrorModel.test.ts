import { describe, expect, it } from "vitest";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import {
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
} from "./workbookInspectorErrorModel";

describe("workbook inspector error model", () => {
  it("preserves the safe primary message for every failure kind", () => {
    const failures = [
      { kind: "validation", message: "Validation failed." },
      {
        kind: "same_field_conflict",
        message: "Conflict.",
        conflict: {
          record_id: "record-1",
          field_key: "note.title",
          current_row_version: 2,
          current_value: "server",
        },
      },
      { kind: "client_txn_conflict", message: "Transaction conflict." },
      { kind: "authentication_required", message: "Sign in." },
      { kind: "authorization_lost", message: "Access lost." },
      { kind: "stale_target", message: "Refresh." },
      { kind: "retryable", message: "Retry." },
      { kind: "invalid_contract", message: "Invalid response." },
      { kind: "terminal", message: "Request failed." },
    ] as readonly WorkbookOperationFailure[];
    for (const failure of failures) {
      expect(workbookInspectorErrorPresentation(failure)).toEqual({
        primaryMessage: failure.message,
        technicalFields: [],
      });
    }
  });

  it("uses the decoded row-version code and retains sanitized detail", () => {
    expect(
      workbookInspectorErrorPresentation({
        kind: "stale_target",
        message: "The record is stale.",
        publicCode: "row_version_conflict",
      }),
    ).toEqual({
      primaryMessage: "This row changed; refresh it before retrying.",
      technicalFields: [
        { label: "Public error code", value: "row_version_conflict" },
        { label: "Server message", value: "The record is stale." },
      ],
    });
  });

  it("keeps local messages on an explicit path without inference", () => {
    expect(
      workbookInspectorLocalErrorPresentation(
        "Local text mentions row_version_conflict literally.",
      ),
    ).toEqual({
      primaryMessage: "Local text mentions row_version_conflict literally.",
      technicalFields: [],
    });
  });
});
