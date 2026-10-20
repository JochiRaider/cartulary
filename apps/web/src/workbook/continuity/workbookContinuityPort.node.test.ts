import { describe, expect, it, vi } from "vitest";
import {
  createWorkbookContinuityPort,
  type WorkbookContinuityAnchor,
} from "./workbookContinuityPort";

const anchor = (
  recordId = "record-1",
  fieldKey = "timeline.activity_synopsis_text",
): WorkbookContinuityAnchor => ({
  fieldKey,
  recordId,
  viewSchemaId: "cartulary.view.timeline.v2",
});

describe("WorkbookContinuityPort", () => {
  it("captures and restores semantic identity through an opaque one-shot token", async () => {
    const focus = vi.fn(async () => true);
    const restore = vi.fn(async () => true);
    const select = vi.fn();
    const port = createWorkbookContinuityPort({
      capture: (subject) => ({ subject, scroll: "private-driver-state" }),
      focus,
      restore,
      select,
    });

    port.select(anchor());
    const supersededToken = port.capture();
    const token = port.capture();

    expect(token).not.toBe(supersededToken);
    expect(await port.restore(supersededToken)).toBe(false);
    expect(port.snapshot().anchor).toEqual(anchor());
    expect(await port.restore(token)).toBe(true);
    expect(restore).toHaveBeenCalledWith(
      anchor(),
      expect.objectContaining({ scroll: "private-driver-state" }),
      expect.any(AbortSignal),
    );
    expect(await port.restore(token)).toBe(false);
  });

  it("focuses and selects only stable schema, record, and field identities", async () => {
    const focus = vi.fn(async () => true);
    const select = vi.fn();
    const port = createWorkbookContinuityPort({
      capture: () => null,
      focus,
      restore: async () => false,
      select,
    });

    expect(await port.focus(anchor("record-2", "timeline.tags"))).toBe(true);
    port.select(anchor("record-2", "timeline.tags"));

    expect(focus).toHaveBeenCalledWith(
      anchor("record-2", "timeline.tags"),
      expect.any(AbortSignal),
    );
    expect(select).toHaveBeenCalledWith(anchor("record-2", "timeline.tags"));
  });

  it("clears captures and selection and disposes idempotently", async () => {
    const select = vi.fn();
    const port = createWorkbookContinuityPort({
      capture: () => null,
      focus: async () => true,
      restore: async () => true,
      select,
    });
    port.select(anchor());
    const token = port.capture();

    port.clear();
    expect(port.snapshot().anchor).toBeNull();
    expect(await port.restore(token)).toBe(false);

    port.dispose();
    port.dispose();
    expect(await port.focus(anchor())).toBe(false);
    expect(() => port.capture()).toThrow("disposed");
  });
});
