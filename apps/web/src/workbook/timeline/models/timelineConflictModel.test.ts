import { describe, expect, it } from "vitest";
import {
  parseSameFieldConflict,
  parseSameFieldConflictFields,
} from "./timelineConflictModel";

const validConflict = {
  base_row_version: 3,
  base_value: "base",
  client_value: "local",
  conflict_resolution_class: "text_compare_merge",
  conflict_token: "token-1",
  current_row_version: 4,
  field_key: "timeline.activity_synopsis_text",
  record_id: "row-1",
  server_updated_at: "2026-04-24T12:00:00Z",
  server_updated_by: "analyst@example.com",
  server_value: "saved",
  suggested_merged_value: "merged",
};

describe("timelineConflictModel", () => {
  it("parses same-field conflict envelopes without optional metadata loss", () => {
    expect(
      parseSameFieldConflict({
        error: {
          code: "same_field_conflict",
          conflict: validConflict,
        },
      }),
    ).toEqual(validConflict);
  });

  it("rejects missing, malformed, and incomplete conflict envelopes", () => {
    expect(parseSameFieldConflict(null)).toBeNull();
    expect(
      parseSameFieldConflict({ error: { code: "validation_failed" } }),
    ).toBeNull();
    for (const key of [
      "conflict_token",
      "record_id",
      "field_key",
      "conflict_resolution_class",
      "base_row_version",
      "current_row_version",
    ] as const) {
      const conflict = { ...validConflict };
      delete conflict[key];
      expect(
        parseSameFieldConflict({
          error: { code: "same_field_conflict", conflict },
        }),
      ).toBeNull();
    }
  });

  it("rejects empty string identity fields for pending queue admission", () => {
    expect(
      parseSameFieldConflictFields({
        ...validConflict,
        record_id: "",
      }),
    ).toBeNull();
    expect(parseSameFieldConflictFields(validConflict)).toMatchObject({
      conflict_token: "token-1",
      record_id: "row-1",
    });
  });
});
