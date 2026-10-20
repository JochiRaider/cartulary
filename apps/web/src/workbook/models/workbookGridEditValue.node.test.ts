import {
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { workbookGridEditChange } from "./workbookGridEditValue";

const field = (view: string, key: string) => {
  const result = requireViewContract(`cartulary.view.${view}.v1`).fieldMap[key];
  if (!result) throw new Error("Missing field fixture");
  return result;
};
describe("Grid field intent", () => {
  it("covers all 97 declared direct-value fields across 15 surfaces and seven editor families", () => {
    const surfaces = new Set<string>(),
      families = new Set<string>();
    let editable = 0;
    for (const view of listViewContracts())
      for (const f of view.fields) {
        if (!f.gridEditable || f.writeKind !== "direct_value") {
          expect(
            workbookGridEditChange(f, "local"),
            `${view.viewSchemaId}:${f.fieldKey}`,
          ).toBeNull();
          continue;
        }
        editable++;
        surfaces.add(view.viewSchemaId);
        const family =
          f.directScalarContractId === "timestamp_instant_v1"
            ? "timestamp"
            : f.directReferenceContractId
              ? "reference"
              : f.enumValues?.length
                ? "enum"
                : f.readKind === "number"
                  ? "number"
                  : f.readKind === "boolean"
                    ? "boolean"
                    : f.stringContractId === "multiline_body_v1" ||
                        f.stringContractId === "reason_note_v1"
                      ? "multiline"
                      : "text";
        families.add(family);
        const raw =
          family === "timestamp"
            ? "2026-06-01T08:04:05.123456789+02:00"
            : family === "reference"
              ? "20000000-0000-4000-8000-000000000001"
              : family === "enum"
                ? (f.enumValues?.[0] ?? "missing-enum-contract")
                : family === "number"
                  ? "42"
                  : family === "boolean"
                    ? "true"
                    : f.stringContractId === "email_address_v1"
                      ? "analyst@example.test"
                      : f.stringContractId === "timezone_name_v1"
                        ? "UTC"
                        : family === "multiline"
                          ? "line one\nline two"
                          : "Value";
        expect(
          workbookGridEditChange(f, raw)?.field_key,
          `${view.viewSchemaId}:${f.fieldKey}`,
        ).toBe(f.fieldKey);
        expect(
          workbookGridEditChange(f, null) !== null,
          `${f.fieldKey} explicit null`,
        ).toBe(f.clearable);
      }
    expect(editable).toBe(97);
    expect(surfaces.size).toBe(15);
    expect([...families].sort()).toEqual([
      "boolean",
      "enum",
      "multiline",
      "number",
      "reference",
      "text",
      "timestamp",
    ]);
  });
  it("distinguishes unfinished timestamp and reference text from explicit null", () => {
    for (const f of [
      field("evidence", "evidence.requested_at"),
      field("evidence", "evidence.source_party_id"),
    ]) {
      expect(workbookGridEditChange(f, "")).toBeNull();
      expect(workbookGridEditChange(f, "  ")).toBeNull();
      expect(workbookGridEditChange(f, null)).toEqual({
        field_key: f.fieldKey,
        value: null,
      });
    }
    const timestamp = field("evidence", "evidence.requested_at");
    for (const raw of [
      "2026-02-30T12:00:00Z",
      "2026-01-01",
      "2026-01-01T12:00:00",
      " 2026-01-01T12:00:00Z ",
    ])
      expect(workbookGridEditChange(timestamp, raw)).toBeNull();
    expect(
      workbookGridEditChange(timestamp, "2026-01-01T12:00:00+02:00")?.value,
    ).toBe("2026-01-01T12:00:00+02:00");
    const reference = field("evidence", "evidence.source_party_id");
    expect(
      workbookGridEditChange(
        reference,
        " 20000000-0000-4000-8000-000000000001 ",
      ),
    ).toBeNull();
  });
  it("validates each scalar family and never introduces a specialized editor", () => {
    expect(
      workbookGridEditChange(
        field("findings", "finding.confidence_score"),
        "1e2",
      ),
    ).toBeNull();
    expect(
      workbookGridEditChange(
        field("findings", "finding.confidence_score"),
        "101",
      ),
    ).toBeNull();
    expect(
      workbookGridEditChange(field("findings", "finding.confidence_score"), "0")
        ?.value,
    ).toBe(0);
    expect(
      workbookGridEditChange(
        field("forensic_keywords", "forensic_keyword.case_sensitive"),
        "false",
      )?.value,
    ).toBe(false);
    expect(
      workbookGridEditChange(
        field("decisions", "decision.status"),
        "not-a-status",
      ),
    ).toBeNull();
    expect(
      workbookGridEditChange(field("notes", "note.body"), " a\r\nb ")?.value,
    ).toBe("a\nb");
    expect(
      workbookGridEditChange(field("notes", "note.title"), null),
    ).toBeNull();
    for (const view of ["assessments", "indicators"])
      for (const f of requireViewContract(`cartulary.view.${view}.v1`).fields)
        expect(workbookGridEditChange(f, "text")).toBeNull();
  });
});
