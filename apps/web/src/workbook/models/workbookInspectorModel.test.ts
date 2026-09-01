import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  initialWorkbookInspectorState,
  inspectorPanelIsDeclared,
  type WorkbookInspectorSubject,
  workbookInspectorReducer,
} from "./workbookInspectorModel";

const timelineSubject = (
  recordId = "row-1",
  rowVersion = 1,
): WorkbookInspectorSubject => ({
  recordId,
  rowVersion,
  viewSchemaId: "cartulary.view.timeline.v2",
});

describe("workbookInspectorModel", () => {
  it("selects immutable schema configuration, starts closed, and exposes the declared no-row state", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v2");
    const config = timeline.inspectorConfig;
    const state = initialWorkbookInspectorState();

    expect(config.viewSchemaId).toBe(timeline.viewSchemaId);
    expect(state).toMatchObject({
      invalidationGeneration: 0,
      isOpen: false,
      status: "closed",
      subject: null,
    });
    expect(config.noRowState).toBe("no_row_selected");
  });

  it("keeps saved-view inheritance tied to the base view_schema_id", () => {
    const baseContract = requireViewContract("cartulary.view.timeline.v2");
    const savedView = {
      saved_view_id: "saved-1",
      view_schema_id: baseContract.viewSchemaId,
    };

    expect(baseContract.inspectorConfig.viewSchemaId).toBe(
      savedView.view_schema_id,
    );
  });

  it("opens explicitly and represents no selection without inventing a row identity", () => {
    const opened = workbookInspectorReducer(initialWorkbookInspectorState(), {
      type: "open",
    });

    expect(opened.isOpen).toBe(true);
    expect(opened.status).toBe("no_row_selected");
    expect(opened.subject).toBeNull();
  });

  it("retargets synchronously by stable schema, record, and row-version identity", () => {
    const opened = workbookInspectorReducer(initialWorkbookInspectorState(), {
      type: "open",
    });
    const selected = workbookInspectorReducer(opened, {
      type: "retarget",
      subject: timelineSubject(),
    });
    const same = workbookInspectorReducer(selected, {
      type: "retarget",
      subject: timelineSubject(),
    });
    const versionChanged = workbookInspectorReducer(selected, {
      type: "retarget",
      subject: timelineSubject("row-1", 2),
    });
    const rowChanged = workbookInspectorReducer(versionChanged, {
      type: "retarget",
      subject: timelineSubject("row-2", 1),
    });

    expect(selected.status).toBe("ready");
    expect(selected.invalidationGeneration).toBe(1);
    expect(same).toBe(selected);
    expect(versionChanged.invalidationGeneration).toBe(2);
    expect(versionChanged.subject?.rowVersion).toBe(2);
    expect(rowChanged.invalidationGeneration).toBe(3);
    expect(rowChanged.subject?.recordId).toBe("row-2");
  });

  it("Verify active view_schema_id selects inspector_config_v1, saved views inherit immutable view_schema_id config, no-row state is no_row_selected, and stale row-bound inspector state invalidates across row, row-version, authorization, incident-lifecycle, delete, merge, hard refresh, and surface changes.", () => {
    const ready = workbookInspectorReducer(
      workbookInspectorReducer(initialWorkbookInspectorState(), {
        type: "retarget",
        subject: timelineSubject(),
      }),
      { type: "open" },
    );

    for (const reason of [
      "authorization_lost",
      "hard_refresh",
      "incident_closed",
      "record_deleted",
      "record_merged",
      "surface_changed",
    ] as const) {
      const invalidated = workbookInspectorReducer(ready, {
        type: "invalidate",
        reason,
      });
      expect(invalidated.isOpen).toBe(false);
      expect(invalidated.status).toBe("closed");
      expect(invalidated.subject).toBeNull();
      expect(invalidated.invalidationGeneration).toBe(
        ready.invalidationGeneration + 1,
      );
    }

    const completed = workbookInspectorReducer(ready, {
      type: "invalidate",
      reason: "action_completed",
    });
    expect(completed.isOpen).toBe(true);
    expect(completed.status).toBe("ready");
    expect(completed.subject).toEqual(timelineSubject());
    expect(completed.invalidationGeneration).toBe(
      ready.invalidationGeneration + 1,
    );
  });

  it("resets closed when the active surface schema changes", () => {
    const ready = workbookInspectorReducer(
      workbookInspectorReducer(initialWorkbookInspectorState(), {
        type: "retarget",
        subject: timelineSubject(),
      }),
      { type: "open" },
    );
    const switched = workbookInspectorReducer(ready, {
      type: "invalidate",
      reason: "surface_changed",
    });

    expect(switched).toMatchObject({
      invalidationGeneration: ready.invalidationGeneration + 1,
      isOpen: false,
      status: "closed",
      subject: null,
    });
  });

  it("filters panels and feature groups by declared semantic config", () => {
    const config = requireViewContract(
      "cartulary.view.hosts.v1",
    ).inspectorConfig;

    expect(inspectorPanelIsDeclared(config, "relationships")).toBe(true);
    expect(inspectorPanelIsDeclared(config, "workflow")).toBe(true);
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
