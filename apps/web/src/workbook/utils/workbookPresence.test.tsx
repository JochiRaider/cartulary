import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import {
  presenceForCell,
  presenceForRow,
  projectWorkbookPresence,
  samePresencePresentation,
} from "../collaboration/workbookPresencePresentation";
import {
  applyWorkbookPresenceDelta,
  replaceWorkbookPresenceSnapshot,
} from "../collaboration/workbookPresenceProjection";
import {
  WorkbookCellPresenceMarker,
  WorkbookPresenceCellLayout,
} from "../components/WorkbookPresenceMarkers";
import { WorkbookPresenceSummary } from "../components/WorkbookStatusStrip";
import type { PresenceRecord } from "./workbookPresence";

afterEach(cleanup);
const sheet = {
  kind: "view_schema",
  id: "cartulary.view.timeline.v2",
} as const;
const record = (connection: string, user = connection): PresenceRecord => ({
  connection_id: connection,
  user_id: user,
  display_name: connection,
  mode: "viewing",
  sheet_ref: sheet,
  observed_at: "1970-01-01T00:00:00Z",
  expires_at: "1970-01-01T00:00:45Z",
});
const project = (records: PresenceRecord[], nowMs = 0) =>
  projectWorkbookPresence({
    records,
    activeSheetRef: sheet,
    connectionId: "self",
    nowMs,
  });
const editing = (
  connection: string,
  user = connection,
  row = "row",
  field = "summary",
): PresenceRecord => ({
  ...record(connection, user),
  mode: "editing",
  record_id: row,
  field_key: field,
});

describe("Workbook presence", () => {
  it("counts unique users in compact displays", () => {
    const { presentation } = project([
      record("a", "user"),
      record("b", "user"),
    ]);
    expect(presentation.header.users).toHaveLength(1);
    expect(presentation.header.overflow).toBe(0);
  });
  it("orders activity before display names", () => {
    const { presentation } = project([
      record("Alpha"),
      editing("Zulu"),
      { ...record("Bravo"), record_id: "row" },
      { ...record("Aardvark"), mode: "idle" },
    ]);
    expect(presentation.header.shown.map((p) => p.connection_id)).toEqual([
      "Zulu",
      "Bravo",
      "Alpha",
      "Aardvark",
    ]);
  });
  it("excludes expired records", () => {
    expect(
      project([record("expired")], 45_000).presentation.header.users,
    ).toEqual([]);
    expect(project([record("active")], 44_999).nextExpiryAtMs).toBe(45_000);
    expect(project([record("expired")], 45_000).nextExpiryAtMs).toBeNull();
  });
  it("selects users independently after row and cell matching", () => {
    const { presentation } = project([
      editing("one", "user", "row-a", "summary"),
      editing("two", "user", "row-b", "details"),
      record("three", "user"),
    ]);
    expect(presentation.header.users).toHaveLength(1);
    expect(presenceForRow(presentation, "row-a").users[0]?.connection_id).toBe(
      "one",
    );
    expect(
      presenceForCell(presentation, "row-b", "details").users[0]?.connection_id,
    ).toBe("two");
    expect(presenceForCell(presentation, "row-a", "details").users).toEqual([]);
    expect(presenceForRow(presentation, null).users).toEqual([]);
  });
  it("counts overflow at every capacity after duplicate state selection", () => {
    const records = Array.from({ length: 6 }, (_, index) =>
      editing(String(index)),
    );
    records.push(editing("duplicate-a", "0"), editing("duplicate-b", "1"));
    const { presentation } = project(records);
    expect(presentation.header.shown).toHaveLength(5);
    expect(presentation.header.overflow).toBe(1);
    expect(presenceForRow(presentation, "row").shown).toHaveLength(3);
    expect(presenceForRow(presentation, "row").overflow).toBe(3);
    expect(presenceForCell(presentation, "row", "summary").shown).toHaveLength(
      2,
    );
    expect(presenceForCell(presentation, "row", "summary").overflow).toBe(4);
  });
  it("uses UTC precision NFC and code points with input independent ties", () => {
    const records = [
      { ...record("z", "same"), display_name: "E\u0301" },
      { ...record("a", "same"), display_name: "É" },
      { ...record("nanos"), observed_at: "1970-01-01T00:00:00.000000001Z" },
      { ...record("private"), display_name: "\ue000" },
      { ...record("astral"), display_name: "\u{10000}" },
      { ...record("older", "latest"), observed_at: "1969-12-31T23:59:59Z" },
      {
        ...record("newer", "latest"),
        observed_at: "1970-01-01T00:00:00.000000002Z",
      },
    ];
    const expected = ["newer", "nanos", "a", "private", "astral"];
    for (let offset = 0; offset < records.length; offset += 1) {
      const rotated = [...records.slice(offset), ...records.slice(0, offset)];
      for (const input of [rotated, [...rotated].reverse()]) {
        expect(
          project(input).presentation.header.users.map((p) => p.connection_id),
        ).toEqual(expected);
      }
    }
    expect(records[0]?.display_name).toBe("E\u0301");
    expect(
      project([
        {
          ...record("utc", "same"),
          observed_at: "1970-01-01T00:00:00.000000001Z",
        },
        {
          ...record("offset", "same"),
          observed_at: "1970-01-01T01:00:00.000000002+01:00",
          expires_at: "1970-01-01T00:00:45+00:00",
        },
      ]).presentation.header.users.map((user) => user.connection_id),
    ).toEqual(["offset"]);
    expect(
      project([
        { ...record("fraction"), expires_at: "1970-01-01T00:00:00.000000001Z" },
      ]).nextExpiryAtMs,
    ).toBe(1);
  });
  it("rejects malformed times modes and anchors without accepting focused wire mode", () => {
    const invalid = [
      { ...record("invalid-date"), observed_at: "2026-02-30T00:00:00Z" },
      { ...record("invalid-expiry"), expires_at: "later" },
      { ...record("wrong-mode"), mode: "focused" },
      { ...record("wrong-field"), field_key: "summary" },
      { ...record("empty-row"), record_id: "" },
    ] as PresenceRecord[];
    expect(project(invalid).presentation.header.users).toEqual([]);
  });
  it("preserves exact sheet boundaries connection self exclusion and extension restrictions", () => {
    const extension = {
      kind: "extension_workspace",
      extension_profile_id: "network_flow_activity",
      workspace_key: "network_analysis",
    } as const;
    const records: PresenceRecord[] = [
      record("self", "own-account"),
      record("other-tab", "own-account"),
      { ...record("saved"), sheet_ref: { kind: "saved_view", id: sheet.id } },
      { ...record("extension"), sheet_ref: extension },
      { ...record("bad-extension"), sheet_ref: extension, record_id: "row" },
    ];
    expect(
      project(records).presentation.header.users.map((p) => p.connection_id),
    ).toEqual(["other-tab"]);
    const result = projectWorkbookPresence({
      records,
      activeSheetRef: extension,
      connectionId: null,
      nowMs: 0,
    });
    expect(
      result.presentation.header.users.map((p) => p.connection_id),
    ).toEqual(["extension"]);
    expect(result.presentation.rows.size).toBe(0);
    expect(result.presentation.cells.size).toBe(0);
  });
  it("keeps canonical connections and display identity stable across renewal", () => {
    const one = record("one", "user"),
      two = record("two", "user");
    const canonical = replaceWorkbookPresenceSnapshot(new Map(), {
      presences: [two, one],
    });
    expect(canonical.size).toBe(2);
    const before = project(Array.from(canonical.values()));
    const renewed = applyWorkbookPresenceDelta(canonical, {
      delta_kind: "upsert",
      presence: { ...one, expires_at: "1970-01-01T00:01:00Z" },
    });
    const after = project(Array.from(renewed.values()));
    expect(
      samePresencePresentation(before.presentation, after.presentation),
    ).toBe(true);
    expect(renewed.size).toBe(2);
    expect(
      replaceWorkbookPresenceSnapshot(canonical, { presences: [one, one] }),
    ).toBe(canonical);
  });
  it("keeps presence outside live announcements", () => {
    render(
      <WorkbookPresenceSummary
        records={project([record("Analyst")]).presentation.header}
      />,
    );
    expect(screen.queryByRole("status")).toBeNull();
    const renderEditor = (records: PresenceRecord[]) => (
      <WorkbookPresenceCellLayout
        editing
        marker={
          <WorkbookCellPresenceMarker
            fieldKey="summary"
            fieldLabel="Summary"
            recordId="row"
            presences={presenceForCell(
              project(records).presentation,
              "row",
              "summary",
            )}
          />
        }
      >
        <input aria-label="Local draft" defaultValue="retained" />
      </WorkbookPresenceCellLayout>
    );
    const view = render(renderEditor([]));
    const input = screen.getByRole("textbox");
    input.focus();
    view.rerender(
      renderEditor([
        {
          ...record("remote"),
          mode: "editing",
          record_id: "row",
          field_key: "summary",
        },
      ]),
    );
    expect(screen.getByRole("textbox")).toBe(input);
    expect(document.activeElement).toBe(input);
    view.rerender(renderEditor([]));
    expect(screen.getByRole("textbox")).toBe(input);
    expect(document.activeElement).toBe(input);
  });
});
