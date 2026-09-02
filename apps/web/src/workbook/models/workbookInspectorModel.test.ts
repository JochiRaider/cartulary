import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  buildWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
} from "../inspector/workbookInspectorSubject";
import {
  initialWorkbookInspectorState,
  workbookInspectorReducer,
  workbookInspectorStateIsOpen,
} from "./workbookInspectorModel";

const lifecycleKey = "timeline:base";
const timeline = requireViewContract("cartulary.view.timeline.v2");

const timelineSubject = (
  recordId = "row-1",
  rowVersion = 1,
  kind: WorkbookInspectorSubject["kind"] = "live",
): WorkbookInspectorSubject => {
  const subject = buildWorkbookInspectorSubject({
    config: timeline.inspectorConfig,
    kind,
    label: "Timeline row",
    recordId,
    rowVersion,
    surfaceLabel: timeline.title,
  });
  if (subject === null) throw new Error("Expected a valid Timeline subject");
  return subject;
};

const initialState = () => initialWorkbookInspectorState({ lifecycleKey });

describe("workbookInspectorModel", () => {
  it("selects immutable schema configuration, starts closed, and exposes the declared no-row state", () => {
    const state = initialState();
    expect(timeline.inspectorConfig.viewSchemaId).toBe(timeline.viewSchemaId);
    expect(state).toEqual({
      invalidationCause: null,
      invalidationGeneration: 0,
      lifecycleKey,
      phase: "closed",
      subject: null,
    });
    expect(workbookInspectorStateIsOpen(state)).toBe(false);
    expect(timeline.inspectorConfig.noRowState).toBe("no_row_selected");
  });

  it("keeps saved-view inheritance tied to the base view_schema_id", () => {
    const savedView = {
      saved_view_id: "saved-1",
      view_schema_id: timeline.viewSchemaId,
    };
    expect(timeline.inspectorConfig.viewSchemaId).toBe(
      savedView.view_schema_id,
    );
  });

  it("opens explicitly and represents no selection without inventing a row identity", () => {
    const opened = workbookInspectorReducer(initialState(), {
      lifecycleKey,
      subject: null,
      type: "open",
    });
    const closed = workbookInspectorReducer(opened, {
      lifecycleKey,
      type: "close",
    });
    expect(opened.phase).toBe("open_no_subject");
    expect(workbookInspectorStateIsOpen(opened)).toBe(true);
    expect(closed.phase).toBe("closed");
    expect(
      workbookInspectorReducer(closed, { lifecycleKey, type: "close" }),
    ).toBe(closed);
  });

  it("retargets synchronously by stable schema, record, and row-version identity", () => {
    const subject = timelineSubject();
    const opened = workbookInspectorReducer(initialState(), {
      lifecycleKey,
      subject,
      type: "open",
    });
    const equal = workbookInspectorReducer(opened, {
      lifecycleKey,
      subject: { ...subject, label: "Presentation-only rename" },
      type: "retarget",
    });
    const versionChanged = workbookInspectorReducer(opened, {
      lifecycleKey,
      subject: timelineSubject("row-1", 2),
      type: "retarget",
    });
    const deleted = workbookInspectorReducer(versionChanged, {
      lifecycleKey,
      subject: timelineSubject("row-1", 2, "deleted"),
      type: "retarget",
    });
    expect(opened.phase).toBe("open_ready");
    expect(equal).toBe(opened);
    expect(versionChanged.invalidationGeneration).toBe(1);
    expect(versionChanged.subject?.rowVersion).toBe(2);
    expect(deleted.invalidationGeneration).toBe(2);
    expect(deleted.subject?.kind).toBe("deleted");
  });

  it("Verify active view_schema_id selects inspector_config_v1, saved views inherit immutable view_schema_id config, no-row state is no_row_selected, and stale row-bound inspector state invalidates across row, row-version, authorization, incident-lifecycle, delete, merge, hard refresh, and surface changes.", () => {
    const subject = timelineSubject();
    const ready = workbookInspectorReducer(initialState(), {
      lifecycleKey,
      subject,
      type: "open",
    });
    for (const reason of [
      "authorization_lost",
      "hard_refresh",
      "incident_closed",
      "record_deleted",
      "record_merged",
      "surface_changed",
    ] as const) {
      const invalidated = workbookInspectorReducer(ready, {
        lifecycleKey,
        reason,
        type: "invalidate",
      });
      expect(invalidated).toMatchObject({
        invalidationCause: reason,
        invalidationGeneration: 1,
        phase: "closed",
        subject: null,
      });
    }
    const completed = workbookInspectorReducer(ready, {
      lifecycleKey,
      reason: "action_completed",
      type: "invalidate",
    });
    expect(completed).toMatchObject({
      invalidationCause: "action_completed",
      invalidationGeneration: 1,
      phase: "open_ready",
      subject,
    });
  });

  it("resets closed when the active surface schema changes", () => {
    const subject = timelineSubject();
    const ready = workbookInspectorReducer(initialState(), {
      lifecycleKey,
      subject,
      type: "open",
    });
    const nextLifecycleKey = "timeline:saved-view";
    const changed = workbookInspectorReducer(ready, {
      lifecycleKey: nextLifecycleKey,
      type: "lifecycle_changed",
    });
    expect(changed).toMatchObject({
      invalidationCause: "surface_changed",
      invalidationGeneration: 1,
      lifecycleKey: nextLifecycleKey,
      phase: "closed",
      subject: null,
    });
    expect(
      workbookInspectorReducer(changed, {
        lifecycleKey,
        subject,
        type: "open",
      }),
    ).toBe(changed);
    expect(
      workbookInspectorReducer(changed, {
        lifecycleKey,
        reason: "hard_refresh",
        type: "invalidate",
      }),
    ).toBe(changed);
    expect(
      workbookInspectorReducer(changed, {
        lifecycleKey,
        subject,
        type: "retarget",
      }),
    ).toBe(changed);
    expect(
      workbookInspectorReducer(changed, { lifecycleKey, type: "close" }),
    ).toBe(changed);
    const reopened = workbookInspectorReducer(changed, {
      lifecycleKey: nextLifecycleKey,
      subject,
      type: "open",
    });
    expect(reopened).toMatchObject({ phase: "open_ready", subject });
  });

  it("filters panels and feature groups by declared semantic config", () => {
    const config = requireViewContract(
      "cartulary.view.hosts.v1",
    ).inspectorConfig;
    expect(
      config.featureGroups
        .filter((group) => group.panelId === "relationships")
        .map((group) => group.featureGroupKey),
    ).toContain("entity.relationships.manage");
    expect(
      config.featureGroups
        .filter((group) => group.panelId === "workflow")
        .map((group) => group.featureGroupKey),
    ).toContain("create_related.note");
  });
});
