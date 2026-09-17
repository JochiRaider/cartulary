import { describe, expect, it, vi } from "vitest";
import type { GridCellAnchor, GridPresentationSnapshot } from "./core";
import type { ActiveEditorSession } from "./editorSessionPolicy";
import { createSemanticCellNavigation } from "./semanticCellNavigation";
import { createGridPresentationPort } from "./semanticPresentationPort";

const surface = { kind: "view_schema", viewSchemaId: "test" } as const;
const anchor: GridCellAnchor = {
  surface,
  fieldKey: "field",
  rowIdentity: { kind: "core_record", recordId: "record" },
};
const model = (): GridPresentationSnapshot => ({
  surface,
  fieldKeys: ["field"],
  rowIdentities: [anchor.rowIdentity],
  revision: 1,
});
function fixture() {
  let settle: (value: boolean) => void = () => {};
  const session: ActiveEditorSession = {
    target: anchor,
    detach: vi.fn(),
    cancel: vi.fn(),
    focus: vi.fn(),
    requestCommit: vi.fn(
      () =>
        new Promise<boolean>((resolve) => {
          settle = resolve;
        }),
    ),
  };
  const state = {
    available: true,
    authority: "editable",
    presentation: model(),
    editor: session as ActiveEditorSession | null,
  };
  const focus = vi.fn(async () => "focused" as const);
  const driver = { read: () => state, focus, cancelInteraction: vi.fn() };
  return {
    state,
    session,
    focus,
    owner: createSemanticCellNavigation(driver),
    settle: (value: boolean) => settle(value),
  };
}

describe("semantic cell navigation", () => {
  it("waits for the existing editor gate and passes the abort signal to virtualized focus", async () => {
    const { owner, session, focus, settle } = fixture();
    const beforeFocus = vi.fn();
    const result = owner.navigate(anchor, { beforeFocus });
    expect(focus).not.toHaveBeenCalled();
    expect(session.focus).toHaveBeenCalledOnce();
    settle(true);
    expect(await result).toBe("focused");
    expect(session.requestCommit).toHaveBeenCalledOnce();
    expect(beforeFocus).toHaveBeenCalledOnce();
    expect(focus).toHaveBeenCalledWith(
      { kind: "cell", anchor },
      { signal: expect.any(AbortSignal) },
    );
  });
  it("retains the original editor on rejection and never reveals the destination", async () => {
    const { owner, session, focus, settle } = fixture();
    const result = owner.navigate(anchor);
    settle(false);
    expect(await result).toBe("rejected");
    expect(focus).not.toHaveBeenCalled();
    expect(session.focus).toHaveBeenCalledTimes(2);
    expect(session.cancel).not.toHaveBeenCalled();
  });
  it("cancels intentions without cancelling the editor write settlement", async () => {
    const { owner, session, focus, settle } = fixture();
    const abort = new AbortController();
    const result = owner.navigate(anchor, { signal: abort.signal });
    abort.abort();
    expect(await result).toBe("cancelled");
    settle(true);
    await Promise.resolve();
    expect(focus).not.toHaveBeenCalled();
    expect(session.cancel).not.toHaveBeenCalled();
    expect(session.detach).not.toHaveBeenCalled();
  });
  it("rejects membership authority and source-text changes after delayed acceptance", async () => {
    for (const change of ["membership", "authority", "source"] as const) {
      const { owner, state, focus, settle } = fixture();
      let valid = true;
      const result = owner.navigate(anchor, { isCurrent: () => valid });
      if (change === "membership")
        state.presentation = { ...model(), fieldKeys: [] };
      if (change === "authority") state.authority = "read_only";
      if (change === "source") valid = false;
      settle(true);
      expect(await result).toBe("unavailable");
      expect(focus).not.toHaveBeenCalled();
    }
  });
  it("allows read-only navigation without an editor or mutation capability", async () => {
    const { owner, state, session } = fixture();
    state.editor = null;
    state.authority = "read_only";
    expect(await owner.navigate(anchor)).toBe("focused");
    expect(session.requestCommit).not.toHaveBeenCalled();
  });
});

describe("semantic presentation port", () => {
  it("publishes ordered identity-only membership without changing revision for value-only refresh", () => {
    const port = createGridPresentationPort();
    const listener = vi.fn();
    const unsubscribe = port.subscribe(listener);
    port.publish(model());
    const first = port.getSnapshot();
    port.publish(model());
    expect(port.getSnapshot()).toBe(first);
    expect(listener).toHaveBeenCalledOnce();
    port.publish({ ...model(), fieldKeys: ["other", "field"] });
    expect(port.getSnapshot()?.fieldKeys).toEqual(["other", "field"]);
    expect(port.getSnapshot()?.revision).toBe(2);
    port.publish({ ...model(), rowIdentities: [] });
    expect(port.getSnapshot()?.rowIdentities).toEqual([]);
    port.retire();
    expect(port.getSnapshot()).toBeNull();
    unsubscribe();
  });
});
