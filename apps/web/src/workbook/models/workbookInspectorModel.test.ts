import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  initialWorkbookInspectorState,
  inspectorFeatureGroupsForPanel,
  inspectorNoRowState,
  inspectorPanelIsDeclared,
  selectInspectorConfig,
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
    const config = selectInspectorConfig(timeline);
    const state = initialWorkbookInspectorState(config);

    expect(config.viewSchemaId).toBe(timeline.viewSchemaId);
    expect(state).toMatchObject({
      configViewSchemaId: "cartulary.view.timeline.v2",
      invalidationGeneration: 0,
      isOpen: false,
      status: "closed",
      subject: null,
    });
    expect(inspectorNoRowState(config)).toBe("no_row_selected");
  });

  it("keeps saved-view inheritance tied to the base view_schema_id", () => {
    const baseContract = requireViewContract("cartulary.view.timeline.v2");
    const savedView = {
      saved_view_id: "saved-1",
      view_schema_id: baseContract.viewSchemaId,
    };

    expect(selectInspectorConfig(baseContract).viewSchemaId).toBe(
      savedView.view_schema_id,
    );
  });

  it("opens explicitly and represents no selection without inventing a row identity", () => {
    const config = selectInspectorConfig(
      requireViewContract("cartulary.view.timeline.v2"),
    );
    const opened = workbookInspectorReducer(
      initialWorkbookInspectorState(config),
      { type: "open" },
    );

    expect(opened.isOpen).toBe(true);
    expect(opened.status).toBe("no_row_selected");
    expect(opened.subject).toBeNull();
  });

  it("retargets synchronously by stable schema, record, and row-version identity", () => {
    const config = selectInspectorConfig(
      requireViewContract("cartulary.view.timeline.v2"),
    );
    const opened = workbookInspectorReducer(
      initialWorkbookInspectorState(config),
      { type: "open" },
    );
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
    const config = selectInspectorConfig(
      requireViewContract("cartulary.view.timeline.v2"),
    );
    const ready = workbookInspectorReducer(
      workbookInspectorReducer(initialWorkbookInspectorState(config), {
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
    const timelineConfig = selectInspectorConfig(
      requireViewContract("cartulary.view.timeline.v2"),
    );
    const hostsConfig = selectInspectorConfig(
      requireViewContract("cartulary.view.hosts.v1"),
    );
    const ready = workbookInspectorReducer(
      workbookInspectorReducer(initialWorkbookInspectorState(timelineConfig), {
        type: "retarget",
        subject: timelineSubject(),
      }),
      { type: "open", panelId: "workflow" },
    );
    const switched = workbookInspectorReducer(ready, {
      type: "reset_config",
      config: hostsConfig,
    });

    expect(switched).toMatchObject({
      activePanelId: "details",
      configViewSchemaId: "cartulary.view.hosts.v1",
      invalidationGeneration: ready.invalidationGeneration + 1,
      isOpen: false,
      status: "closed",
      subject: null,
    });
  });

  it("filters panels and feature groups by declared semantic config", () => {
    const config = selectInspectorConfig(
      requireViewContract("cartulary.view.hosts.v1"),
    );

    expect(inspectorPanelIsDeclared(config, "relationships")).toBe(true);
    expect(inspectorPanelIsDeclared(config, "workflow")).toBe(true);
    expect(
      inspectorFeatureGroupsForPanel(config, "relationships").map(
        (group) => group.featureGroupKey,
      ),
    ).toContain("entity.relationships.manage");
    expect(
      inspectorFeatureGroupsForPanel(config, "workflow").map(
        (group) => group.featureGroupKey,
      ),
    ).toContain("create_related.note");
  });
});
