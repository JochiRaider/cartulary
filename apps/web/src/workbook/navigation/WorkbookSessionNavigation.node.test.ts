import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  buildSavedViewLayoutJson,
  emptyWorkbookQueryState,
} from "../models/workbookQuery";
import type { WorkbookNavigationIntent } from "./WorkbookNavigationIntent";
import {
  type WorkbookReturnOrigin,
  WorkbookSessionNavigation,
} from "./WorkbookSessionNavigation";

const intent = (id: string): WorkbookNavigationIntent => ({
  target: { sheetRef: { kind: "view_schema", id } },
  entry: "open",
  inspect: false,
});
const origin = {
  incidentId: "incident",
  sheetRef: { kind: "view_schema", id: "timeline" },
  viewSchemaId: "timeline",
  query: emptyWorkbookQueryState(),
  layout: buildSavedViewLayoutJson(
    requireViewContract("cartulary.view.timeline.v2"),
  ),
  recordId: "original",
  invoker: "grid",
} satisfies WorkbookReturnOrigin;
const session = () => {
  const owner = new WorkbookSessionNavigation("incident");
  owner.setReadable(true);
  return owner;
};
describe("session working set and return trail", () => {
  it("detects duplicates before capacity, preserves insertion order, and never evicts pins", () => {
    const owner = session();
    for (let i = 0; i < 20; i++)
      owner.pin({ ...origin, recordId: String(i), label: `Record ${i}` });
    owner.pin({
      ...origin,
      recordId: "0",
      fieldKey: "another-field",
      label: "New label",
    });
    expect(owner.getSnapshot().message).toBe("Already in working set");
    owner.pin({ ...origin, recordId: "overflow", label: "Overflow" });
    expect(owner.getSnapshot().message).toContain("full");
    expect(owner.getSnapshot().pins.map((pin) => pin.recordId)).toEqual(
      Array.from({ length: 20 }, (_, i) => String(i)),
    );
    owner.setReadable(false);
    expect(owner.getSnapshot().pins).toEqual([]);
    owner.setReadable(true);
    expect(owner.getSnapshot().pins).toHaveLength(20);
    owner.clear();
    expect(owner.getSnapshot().pins).toEqual([]);
  });
  it("coalesces identical pivots, cancels obsolete reads, and pushes only successful changed origins", async () => {
    const owner = session();
    let finish!: (value: "changed") => void;
    let captured!: AbortSignal;
    const pending = owner.navigate(intent("a"), origin, (signal) => {
      captured = signal;
      return new Promise((resolve) => {
        finish = resolve;
      });
    });
    expect(owner.navigate(intent("a"), origin, async () => "failed")).toBe(
      pending,
    );
    await Promise.resolve();
    await owner.navigate(intent("b"), origin, async () => "same");
    expect(captured.aborted).toBe(true);
    finish("changed");
    expect(await pending).toBe(false);
    expect(owner.getSnapshot().trail).toEqual([]);
    await owner.navigate(intent("c"), origin, async () => "failed");
    expect(owner.getSnapshot().trail).toEqual([]);
    await owner.navigate(intent("d"), origin, async () => "changed");
    expect(owner.getSnapshot().trail).toEqual([origin]);
    let acceptedSignal!: AbortSignal;
    await owner.navigate(intent("accepted"), origin, async (signal) => {
      acceptedSignal = signal;
      return "changed";
    });
    expect(acceptedSignal.aborted).toBe(false);
    owner.cancel();
    expect(acceptedSignal.aborted).toBe(true);
    // Later interaction cancels attachment, not the already accepted trail push.
    expect(owner.getSnapshot().trail).toEqual([origin, origin]);
  });
  it("bounds semantic origins and consumes Return only after acceptance, without a bounce entry", async () => {
    const owner = session();
    for (let i = 0; i < 35; i++)
      await owner.navigate(
        intent(String(i)),
        { ...origin, recordId: String(i) },
        async () => "changed",
      );
    expect(owner.getSnapshot().trail).toHaveLength(32);
    expect(owner.getSnapshot().trail[0]?.recordId).toBe("3");
    await owner.navigate(intent("return"), origin, async () => "failed", true);
    expect(owner.getSnapshot().trail).toHaveLength(32);
    await owner.navigate(intent("return"), origin, async () => "same", true);
    expect(owner.getSnapshot().trail).toHaveLength(31);
    owner.setReadable(false);
    expect(owner.getSnapshot().trail).toEqual([]);
  });
});

it("navigation completion requires current presentation and cancellation fences accepted destinations", async () => {
  const owner = session();
  await owner.navigate(intent("same"), origin, async () => "same");
  const first = owner.getSnapshot().attemptId;
  expect(owner.getSnapshot().outcome).toBe("admitted");
  owner.cancel();
  owner.completePresentation(first);
  expect(owner.getSnapshot().outcome).toBe("cancelled");
  await owner.navigate(intent("same"), origin, async () => "same");
  expect(owner.getSnapshot().attemptId).toBeGreaterThan(first);
  owner.completePresentation(first);
  expect(owner.getSnapshot().outcome).toBe("admitted");
  owner.completePresentation(owner.getSnapshot().attemptId);
  expect(owner.getSnapshot().outcome).toBe("succeeded");
  expect(owner.getSnapshot().trail).toEqual([]);
  await owner.navigate(intent("failure"), origin, async () => "failed");
  owner.completePresentation(owner.getSnapshot().attemptId);
  expect(owner.getSnapshot().outcome).toBe("failed");
});

it("coalesces admitted presentation until terminal completion then admits a fresh observation", async () => {
  const owner = session();
  const request = owner.navigate(intent("same"), origin, async () => "changed");
  expect(await request).toBe(true);
  const attempt = owner.getSnapshot().attemptId;
  expect(owner.getSnapshot().pending).toBe(true);
  expect(owner.navigate(intent("same"), origin, async () => "failed")).toBe(
    request,
  );
  expect(owner.getSnapshot().trail).toEqual([origin]);
  owner.completePresentation(attempt);
  expect(owner.getSnapshot().pending).toBe(false);
  const fresh = owner.navigate(intent("same"), origin, async () => "same");
  expect(fresh).not.toBe(request);
  await fresh;
  expect(owner.getSnapshot().attemptId).toBe(attempt + 1);
  owner.completePresentation(attempt, "failed");
  expect(owner.getSnapshot().outcome).toBe("admitted");
  owner.completePresentation(attempt + 1);
  expect(owner.getSnapshot().trail).toEqual([origin]);
});

it("distinguishes semantic entry anchor inspection and captured Return configuration", async () => {
  const owner = session();
  const returning: WorkbookReturnOrigin = {
    ...origin,
    sheetRef: { kind: "saved_view", id: "saved" },
    savedViewVersion: 1,
  };
  const base: WorkbookNavigationIntent = {
    target: {
      sheetRef: returning.sheetRef,
      recordId: "record",
      fieldKey: "field",
    },
    entry: "return",
    inspect: false,
    returning,
  };
  let dispatches = 0;
  const dispatch = async () => {
    dispatches++;
    return "same" as const;
  };
  await owner.navigate(base, origin, dispatch, true);
  await owner.navigate(
    {
      ...base,
      returning: {
        ...returning,
        query: { groupBy: null, filters: [], sort: [] },
      },
    },
    origin,
    dispatch,
    true,
  );
  expect(dispatches).toBe(1);
  for (const changed of [
    { ...base, inspect: true },
    { ...base, entry: "pin" as const },
    {
      ...base,
      target: {
        ...base.target,
        sheetRef: { kind: "view_schema" as const, id: "other" },
      },
    },
    { ...base, target: { sheetRef: returning.sheetRef, recordId: "other" } },
    {
      ...base,
      target: {
        sheetRef: returning.sheetRef,
        recordId: "record",
        fieldKey: "other",
      },
    },
    { ...base, returning: { ...returning, savedViewVersion: 2 } },
    {
      ...base,
      returning: {
        ...returning,
        layout: { ...returning.layout, hidden_field_keys: ["field"] },
      },
    },
    {
      ...base,
      returning: {
        ...returning,
        query: {
          ...returning.query,
          sort: [{ fieldKey: "field", direction: "desc" as const }],
        },
      },
    },
  ]) {
    const previous = owner.getSnapshot().attemptId;
    await owner.navigate(changed, origin, dispatch);
    expect(owner.getSnapshot().attemptId).toBe(previous + 1);
  }
  const extension: WorkbookReturnOrigin = {
    incidentId: "incident",
    sheetRef: {
      kind: "extension_workspace",
      extension_profile_id: "profile",
      workspace_key: "root",
    },
    invoker: "view",
  };
  await owner.navigate(intent("extension"), extension, async () => "changed");
  expect(owner.getSnapshot().trail.at(-1)).toEqual(extension);
  // @ts-expect-error An extension origin cannot carry a workbook query.
  const invalid: WorkbookReturnOrigin = {
    ...extension,
    query: emptyWorkbookQueryState(),
  };
  void invalid;
});
