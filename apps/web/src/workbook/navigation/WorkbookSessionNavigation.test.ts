import { describe, expect, it } from "vitest";
import {
  type WorkbookReturnOrigin,
  WorkbookSessionNavigation,
} from "./WorkbookSessionNavigation";

const origin: WorkbookReturnOrigin = {
  incidentId: "incident",
  sheetRef: { kind: "view_schema", id: "timeline" },
  recordId: "original",
  invoker: "grid",
};
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
    const pending = owner.navigate("a", origin, (signal) => {
      captured = signal;
      return new Promise((resolve) => {
        finish = resolve;
      });
    });
    expect(owner.navigate("a", origin, async () => "failed")).toBe(pending);
    await Promise.resolve();
    await owner.navigate("b", origin, async () => "same");
    expect(captured.aborted).toBe(true);
    finish("changed");
    expect(await pending).toBe(false);
    expect(owner.getSnapshot().trail).toEqual([]);
    await owner.navigate("c", origin, async () => "failed");
    expect(owner.getSnapshot().trail).toEqual([]);
    await owner.navigate("d", origin, async () => "changed");
    expect(owner.getSnapshot().trail).toEqual([origin]);
    let acceptedSignal!: AbortSignal;
    await owner.navigate("accepted", origin, async (signal) => {
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
        String(i),
        { ...origin, recordId: String(i) },
        async () => "changed",
      );
    expect(owner.getSnapshot().trail).toHaveLength(32);
    expect(owner.getSnapshot().trail[0]?.recordId).toBe("3");
    await owner.navigate("return", origin, async () => "failed", true);
    expect(owner.getSnapshot().trail).toHaveLength(32);
    await owner.navigate("return", origin, async () => "same", true);
    expect(owner.getSnapshot().trail).toHaveLength(31);
    owner.setReadable(false);
    expect(owner.getSnapshot().trail).toEqual([]);
  });
});
