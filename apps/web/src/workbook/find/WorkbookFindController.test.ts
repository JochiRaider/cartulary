import { type GridCellAnchor, gridAnchorKey } from "@cartulary/grid-adapter";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import {
  WorkbookFindController,
  type WorkbookFindSource,
  workbookFindStatus,
} from "./WorkbookFindController";

const surface = { kind: "view_schema", viewSchemaId: "timeline" } as const;
const anchor = (recordId: string, fieldKey = "text"): GridCellAnchor => ({
  surface,
  rowIdentity: { kind: "core_record", recordId },
  fieldKey,
});
function source(
  values: readonly (readonly string[])[],
  fields = ["text"],
): WorkbookFindSource {
  return {
    navigationKey: "window",
    lifetimeKey: "incident",
    readable: true,
    stale: false,
    unavailableReason: null,
    presentation: {
      surface,
      revision: 1,
      fieldKeys: fields,
      rowIdentities: values.map((_, index) => ({
        kind: "core_record",
        recordId: String(index),
      })),
    },
    readText: (target) =>
      target.rowIdentity.kind === "core_record"
        ? [
            values[Number(target.rowIdentity.recordId)]?.[
              fields.indexOf(target.fieldKey)
            ] ?? "",
          ]
        : [],
  };
}
const settle = () => vi.runAllTimersAsync();
afterEach(() => vi.useRealTimers());

describe("Workbook Find", () => {
  it("matches literal normalized fragments once per cell in presented order", async () => {
    vi.useFakeTimers();
    const owner = new WorkbookFindController(async () => "focused");
    owner.setSource(
      source(
        [
          ["cafe\u0301 café", "CAFÉ"],
          ["cafe", "café"],
        ],
        ["left", "right"],
      ),
    );
    owner.open(null);
    owner.setTerm("CAFÉ");
    await settle();
    expect(owner.getSnapshot().matches.map(gridAnchorKey)).toEqual(
      [anchor("0", "left"), anchor("0", "right"), anchor("1", "right")].map(
        gridAnchorKey,
      ),
    );
    owner.setMatchCase(true);
    await settle();
    expect(owner.getSnapshot().matches).toEqual([anchor("0", "right")]);
    owner.setTerm("cafe");
    await settle();
    expect(owner.getSnapshot().matches).toEqual([anchor("1", "left")]);
  });
  it("preserves whitespace punctuation multiline text and explicit lowercasing semantics", async () => {
    vi.useFakeTimers();
    const owner = new WorkbookFindController(async () => "focused");
    owner.setSource(source([["one\n Two.* 🙂"], ["Straße"], ["STRASSE"]]));
    owner.open(null);
    for (const [term, records] of [
      [" Two.*", ["0"]],
      ["one\n Two", ["0"]],
      ["🙂", ["0"]],
      ["strasse", ["2"]],
      [".*", ["0"]],
      ["two*", []],
    ] as const) {
      owner.setTerm(term);
      await settle();
      expect(owner.getSnapshot().matches).toEqual(
        records.map((id) => anchor(id)),
      );
    }
    owner.setTerm("");
    await settle();
    expect(owner.getSnapshot().matches).toEqual([]);
    expect(workbookFindStatus(owner.getSnapshot())).toContain("Enter text");
  });
  it("starts inclusively at the origin and navigates both directions with wrap feedback", async () => {
    vi.useFakeTimers();
    const move = vi.fn(async () => "focused" as const);
    const owner = new WorkbookFindController(move);
    owner.setSource(source([["hit"], ["hit"], ["miss"], ["hit"]]));
    owner.open(anchor("1"));
    owner.setTerm("hit");
    await settle();
    expect(move).not.toHaveBeenCalled();
    await owner.navigate(1);
    expect(owner.getSnapshot().current).toEqual(anchor("1"));
    expect(owner.getSnapshot().open).toBe(false);
    owner.open(anchor("1"));
    expect(owner.getSnapshot().status).toBe("ready");
    await settle();
    await owner.navigate(-1);
    expect(owner.getSnapshot().current).toEqual(anchor("0"));
    await owner.navigate(-1);
    expect(owner.getSnapshot().current).toEqual(anchor("3"));
    expect(workbookFindStatus(owner.getSnapshot())).toContain("Wrapped to end");
    await owner.navigate(1);
    expect(workbookFindStatus(owner.getSnapshot())).toContain(
      "Wrapped to beginning",
    );
  });
  it("retains semantic current matches on reorder and clears disappearing matches without moving", async () => {
    vi.useFakeTimers();
    const move = vi.fn(async () => "focused" as const);
    const owner = new WorkbookFindController(move);
    const original = source([["hit"], ["hit"]]);
    owner.setSource(original);
    owner.open(anchor("0"));
    owner.setTerm("hit");
    await settle();
    await owner.navigate(1);
    owner.setSource({
      ...original,
      navigationKey: "reordered",
      presentation: {
        ...original.presentation,
        rowIdentities: [...original.presentation.rowIdentities].reverse(),
      },
    });
    await settle();
    expect(workbookFindStatus(owner.getSnapshot())).toContain("2 of 2");
    owner.setSource(source([["miss"], ["hit"]]));
    await settle();
    expect(owner.getSnapshot().current).toBeNull();
    expect(move).toHaveBeenCalledTimes(1);
    owner.open(anchor("0"));
    await settle();
    await owner.navigate(1, anchor("1"));
    expect(owner.getSnapshot().current).toEqual(anchor("1"));
  });
  it("cancels obsolete scans and releases all search material on retirement", async () => {
    vi.useFakeTimers();
    const owner = new WorkbookFindController(async () => "focused");
    const read = vi.fn(() => ["old new"]);
    owner.setSource({
      ...source(Array.from({ length: 300 }, () => ["old new"])),
      readText: read,
    });
    owner.open(null);
    owner.setTerm("old");
    await vi.advanceTimersToNextTimerAsync();
    expect(read).toHaveBeenCalledTimes(32);
    owner.setTerm("new");
    await settle();
    expect(owner.getSnapshot().matches).toHaveLength(300);
    owner.setTerm("absent");
    owner.close();
    await settle();
    expect(owner.getSnapshot()).toMatchObject({
      term: "",
      matches: [],
      active: false,
    });
    owner.open(null);
    owner.setTerm("old");
    owner.setSource(null);
    await settle();
    expect(owner.getSnapshot()).toMatchObject({
      term: "",
      matches: [],
      active: false,
    });
  });
  it("fences pending destinations by input window membership and close without cancelling settlement", async () => {
    vi.useFakeTimers();
    for (const supersede of ["term", "window", "close"] as const) {
      const pending = deferred<"focused">();
      let options:
        | Parameters<ConstructorParameters<typeof WorkbookFindController>[0]>[1]
        | undefined;
      const owner = new WorkbookFindController((_anchor, next) => {
        options = next;
        return pending.promise;
      });
      owner.setSource(source([["hit"]]));
      owner.open(null);
      owner.setTerm("hit");
      await settle();
      const navigation = owner.navigate(1);
      if (supersede === "term") owner.setTerm("new");
      if (supersede === "window")
        owner.setSource({ ...source([["hit"]]), navigationKey: "replacement" });
      if (supersede === "close") owner.close();
      expect(options?.signal?.aborted).toBe(true);
      pending.resolve("focused");
      expect(await navigation).toBe("cancelled");
      expect(owner.getSnapshot().current).toBeNull();
      owner.dispose();
    }
  });
  it("rechecks committed text after delayed settlement without treating value-only updates as a new window", async () => {
    vi.useFakeTimers();
    const pending = deferred<"focused">();
    let options:
      | Parameters<ConstructorParameters<typeof WorkbookFindController>[0]>[1]
      | undefined;
    const owner = new WorkbookFindController((_anchor, next) => {
      options = next;
      return pending.promise;
    });
    owner.setSource(source([["hit"]]));
    owner.open(null);
    owner.setTerm("hit");
    await settle();
    const navigation = owner.navigate(1);
    owner.setSource(source([["miss"]]));
    await settle();
    expect(options?.signal?.aborted).toBe(true);
    expect(options?.isCurrent?.()).toBe(false);
    pending.resolve("focused");
    expect(await navigation).toBe("cancelled");
    expect(owner.getSnapshot().navigating).toBe(false);
  });
  it("does not match across fragments and distinguishes unavailable from scoped zero results", async () => {
    vi.useFakeTimers();
    const owner = new WorkbookFindController(async () => "focused");
    owner.setSource({
      ...source([["unused"]]),
      readText: () => ["first", "second"],
    });
    owner.open(null);
    owner.setTerm("firstsecond");
    await settle();
    expect(workbookFindStatus(owner.getSnapshot())).toBe(
      "No matches in loaded rows.",
    );
    owner.setSource({ ...source([]), unavailableReason: "No visible fields." });
    expect(workbookFindStatus(owner.getSnapshot())).toBe("No visible fields.");
  });
});
