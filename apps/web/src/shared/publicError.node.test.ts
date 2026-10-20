import { describe, expect, it } from "vitest";
import { publicErrorView } from "./publicError";

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
});
