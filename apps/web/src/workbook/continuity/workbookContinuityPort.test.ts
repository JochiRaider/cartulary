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
  it("captures and restores semantic identity through an opaque one-shot token", () => {
    const focus = vi.fn(() => true);
    const restore = vi.fn(() => true);
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

    expect(token).toMatch(/^workbook-continuity-\d+$/);
    expect(port.restore(supersededToken)).toBe(false);
    expect(port.snapshot().anchor).toEqual(anchor());
    expect(port.restore(token)).toBe(true);
    expect(restore).toHaveBeenCalledWith(
      anchor(),
      expect.objectContaining({ scroll: "private-driver-state" }),
    );
    expect(port.restore(token)).toBe(false);
  });

  it("focuses and selects only stable schema, record, and field identities", () => {
    const focus = vi.fn(() => true);
    const select = vi.fn();
    const port = createWorkbookContinuityPort({
      capture: () => null,
      focus,
      restore: () => false,
      select,
    });

    expect(port.focus(anchor("record-2", "timeline.tags"))).toBe(true);
    port.select(anchor("record-2", "timeline.tags"));

    expect(focus).toHaveBeenCalledWith(anchor("record-2", "timeline.tags"));
    expect(select).toHaveBeenCalledWith(anchor("record-2", "timeline.tags"));
  });

  it("clears captures and selection and disposes idempotently", () => {
    const select = vi.fn();
    const port = createWorkbookContinuityPort({
      capture: () => null,
      focus: () => true,
      restore: () => true,
      select,
    });
    port.select(anchor());
    const token = port.capture();

    port.clear();
    expect(port.snapshot().anchor).toBeNull();
    expect(port.restore(token)).toBe(false);

    port.dispose();
    port.dispose();
    expect(port.focus(anchor())).toBe(false);
    expect(() => port.capture()).toThrow("disposed");
  });
});
