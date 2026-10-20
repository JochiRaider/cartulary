import { describe, expect, it, vi } from "vitest";
import {
  type WorkbookRecoveryItem,
  WorkbookRecoveryNavigation,
  workbookConflictRecoveryKey,
  workbookRecoveryKey,
} from "./workbookRecoveryNavigation";

const item = (
  id: string,
  attention: WorkbookRecoveryItem["attention"] = "attention",
  order = 0,
): WorkbookRecoveryItem => ({
  id,
  attention,
  order,
  label: "Retained action",
  summary: "Review required",
  origin: "Timeline",
  sheetRef: { kind: "view_schema", id: "timeline" },
});

describe("Recovery navigation", () => {
  it("orders safe contributions and keeps selection by identity when other work changes", () => {
    const navigation = new WorkbookRecoveryNavigation();
    const activate = vi.fn(),
      detach = vi.fn();
    const a = navigation.register("creation", { activate, detach });
    const b = navigation.register("batch", {});
    a.update([item("draft", "draft"), item("later", "attention", 8)]);
    b.update([item("batch", "attention", 1)]);
    navigation.activate(workbookRecoveryKey("creation", "later"));
    b.update([item("batch", "completed", 1), item("new", "progress", 2)]);
    expect(navigation.getSnapshot().selected).toBe(
      workbookRecoveryKey("creation", "later"),
    );
    expect(navigation.getSnapshot().entries.map((entry) => entry.id)).toEqual([
      "later",
      "new",
      "draft",
      "batch",
    ]);
    expect(navigation.getSnapshot().count).toBe(3);
    expect(activate).toHaveBeenCalledExactlyOnceWith("later");
    expect(detach).not.toHaveBeenCalled();
  });
  it("detaches exactly the previous owner when switching closing or withdrawing a selection", () => {
    const navigation = new WorkbookRecoveryNavigation();
    const detachA = vi.fn(),
      detachB = vi.fn();
    const a = navigation.register("a", { detach: detachA });
    const b = navigation.register("b", { detach: detachB });
    a.update([item("one")]);
    b.update([item("two")]);
    navigation.activate(workbookRecoveryKey("a", "one"));
    navigation.activate(workbookRecoveryKey("b", "two"));
    expect(detachA).toHaveBeenCalledOnce();
    navigation.close();
    navigation.close();
    expect(detachB).toHaveBeenCalledOnce();
    navigation.activate(workbookRecoveryKey("a", "one"));
    a.update([]);
    expect(detachA).toHaveBeenCalledTimes(2);
    expect(navigation.getSnapshot()).toMatchObject({
      open: true,
      selected: null,
      count: 1,
    });
    expect(navigation.getSnapshot().message).toBeTruthy();
  });
  it("ignores replaced registrations and permanently retires every presentation reference", () => {
    const navigation = new WorkbookRecoveryNavigation();
    const old = navigation.register("source", {});
    old.update([item("old")]);
    const current = navigation.register("source", {});
    current.update([item("new")]);
    old.update([item("leak")]);
    old.unregister();
    expect(navigation.getSnapshot().entries.map((entry) => entry.id)).toEqual([
      "new",
    ]);
    navigation.dispose();
    current.update([item("late")]);
    navigation.activate(workbookRecoveryKey("source", "new"));
    expect(navigation.getSnapshot()).toMatchObject({
      entries: [],
      count: 0,
      open: false,
      selected: null,
    });
  });
  it("keeps completed notice dismissal separate from work and rejects stale activation", () => {
    const navigation = new WorkbookRecoveryNavigation();
    const activate = vi.fn();
    const source = navigation.register("source", { activate });
    source.update([item("working"), item("saved", "completed")]);
    navigation.dismissCompleted(workbookRecoveryKey("source", "working"));
    navigation.dismissCompleted(workbookRecoveryKey("source", "saved"));
    expect(navigation.getSnapshot().entries.map((entry) => entry.id)).toEqual([
      "working",
    ]);
    navigation.activate(workbookRecoveryKey("source", "saved"));
    expect(activate).not.toHaveBeenCalled();
    expect(navigation.getSnapshot()).toMatchObject({
      open: true,
      selected: null,
      count: 1,
    });
  });
  it("coordinates secondary panels without reopening them after recovery closes", () => {
    const navigation = new WorkbookRecoveryNavigation();
    const detachInspector = vi.fn(),
      detachDialog = vi.fn();
    navigation.attachExternal(Symbol("inspector"), detachInspector);
    navigation.openList();
    expect(detachInspector).toHaveBeenCalledOnce();
    navigation.attachExternal(Symbol("dialog"), detachDialog);
    expect(navigation.getSnapshot().open).toBe(false);
    navigation.openList();
    navigation.close();
    expect(detachDialog).toHaveBeenCalledOnce();
    expect(detachInspector).toHaveBeenCalledOnce();
  });
  it("rejects duplicate source identities without admitting an ambiguous navigation target", () => {
    const source = new WorkbookRecoveryNavigation().register("source", {});
    expect(() => source.update([item("same"), item("same")])).toThrow(
      "Duplicate recovery work identity",
    );
  });
  it("counts compound conflicts and acknowledged refresh once while preserving independent debt", () => {
    const navigation = new WorkbookRecoveryNavigation();
    const core = navigation.register("core", {});
    const work = navigation.register("owner", {});
    const debt = navigation.register("refresh", {});
    core.update([
      { ...item("conflict"), conflictOperationId: "captured-link" },
      { ...item("same-field"), conflictKey: "row:field" },
    ]);
    debt.update([
      { ...item("view-a"), refreshOnlyView: "view-a" },
      { ...item("view-b"), refreshOnlyView: "view-b" },
    ]);
    work.update([
      {
        ...item("operation"),
        operationIds: ["captured-link"],
        conflictKeys: ["row:field"],
        refreshViews: ["view-a"],
      },
    ]);
    expect(
      workbookConflictRecoveryKey(navigation.getSnapshot().entries, {
        key: "row:field",
      }),
    ).toBe(workbookRecoveryKey("owner", "operation"));
    expect(navigation.getSnapshot().count).toBe(2);
    expect(navigation.getSnapshot().entries.map((entry) => entry.id)).toEqual([
      "operation",
      "view-b",
    ]);
    navigation.activate(workbookRecoveryKey("owner", "operation"));
    work.update([
      {
        ...item("operation", "completed"),
        operationIds: ["captured-link"],
        conflictKeys: ["row:field"],
      },
    ]);
    expect(navigation.getSnapshot().count).toBe(2);
    expect(navigation.getSnapshot().selected).toBe(
      workbookRecoveryKey("owner", "operation"),
    );
    work.unregister();
    expect(navigation.getSnapshot().count).toBe(4);
    expect(navigation.getSnapshot().selected).toBeNull();
  });
});
