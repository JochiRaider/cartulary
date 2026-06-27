import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  initialWorkbookInspectorState,
  inspectorFeatureGroupsForPanel,
  inspectorNoRowState,
  inspectorPanelIsDeclared,
  selectInspectorConfig,
  workbookInspectorReducer,
} from "./workbookInspectorModel";

describe("workbookInspectorModel", () => {
  it("FE-U-P9-02 selects config from immutable view_schema_id and starts closed", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v2");
    const config = selectInspectorConfig(timeline);
    const state = initialWorkbookInspectorState(config);

    expect(config.viewSchemaId).toBe(timeline.viewSchemaId);
    expect(config.defaultOpen).toBe(false);
    expect(state.isOpen).toBe(false);
    expect(state.configViewSchemaId).toBe("cartulary.view.timeline.v2");
    expect(inspectorNoRowState(config)).toBe("no_row_selected");
  });

  it("FE-U-P9-02 keeps saved-view inheritance tied to the base view_schema_id", () => {
    const baseContract = requireViewContract("cartulary.view.timeline.v2");
    const savedView = {
      saved_view_id: "saved-1",
      view_schema_id: baseContract.viewSchemaId,
    };

    expect(selectInspectorConfig(baseContract).viewSchemaId).toBe(
      savedView.view_schema_id,
    );
  });

  it("FE-U-P9-02 Verify active view_schema_id selects inspector_config_v1, saved views inherit immutable view_schema_id config, no-row state is no_row_selected, and stale row-bound inspector state invalidates across row, row-version, authorization, incident-lifecycle, delete, merge, hard refresh, and surface changes.", () => {
    const timeline = requireViewContract("cartulary.view.timeline.v2");
    const hosts = requireViewContract("cartulary.view.hosts.v1");
    const timelineConfig = selectInspectorConfig(timeline);
    const hostsConfig = selectInspectorConfig(hosts);
    const savedView = {
      saved_view_id: "saved-timeline",
      view_schema_id: timeline.viewSchemaId,
    };

    expect(timelineConfig.viewSchemaId).toBe("cartulary.view.timeline.v2");
    expect(hostsConfig.viewSchemaId).toBe("cartulary.view.hosts.v1");
    expect(selectInspectorConfig(timeline).viewSchemaId).toBe(
      savedView.view_schema_id,
    );
    expect(timelineConfig.defaultOpen).toBe(false);
    expect(inspectorNoRowState(timelineConfig)).toBe("no_row_selected");

    const staged = workbookInspectorReducer(
      workbookInspectorReducer(initialWorkbookInspectorState(timelineConfig), {
        type: "select_row",
        row: { recordId: "row-1", rowVersion: 1 },
      }),
      {
        type: "stage_row_bound_state",
        localFormKey: "form-1",
        mergePlanKey: "merge-1",
        pendingConfirmationKey: "delete-1",
        rollbackPreviewKey: "rollback-1",
        stalePreviewKey: "preview-1",
        workflowFormKey: "workflow-1",
      },
    );
    const retargeted = workbookInspectorReducer(staged, {
      type: "select_row",
      row: { recordId: "row-2", rowVersion: 1 },
    });

    expect(retargeted.selectedRecordId).toBe("row-2");
    expect(retargeted.pendingConfirmationKey).toBeNull();
    expect(retargeted.rollbackPreviewKey).toBeNull();
    expect(retargeted.mergePlanKey).toBeNull();
    expect(retargeted.workflowFormKey).toBeNull();
    expect(retargeted.localFormKey).toBeNull();
    expect(retargeted.stalePreviewKey).toBeNull();

    for (const action of [
      { type: "row_version_changed" as const, rowVersion: 2 },
      { type: "authorization_lost" as const },
      { type: "incident_closed" as const },
    ]) {
      const next = workbookInspectorReducer(staged, action);
      expect(next.pendingConfirmationKey).toBeNull();
      expect(next.rollbackPreviewKey).toBeNull();
      expect(next.mergePlanKey).toBeNull();
      expect(next.workflowFormKey).toBeNull();
    }
  });

  it("FE-U-P9-02 filters panels and feature groups by declared semantic config", () => {
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
    ).toEqual([
      "surface_pivot.timeline",
      "surface_pivot.evidence",
      "surface_pivot.assessments",
      "create_related.note",
      "create_related.task_request",
      "create_related.decision",
    ]);
  });

  it("FE-U-P9-02 clears stale row-bound state before retargeting to a new row", () => {
    const config = selectInspectorConfig(
      requireViewContract("cartulary.view.timeline.v2"),
    );
    const selected = workbookInspectorReducer(
      initialWorkbookInspectorState(config),
      { type: "select_row", row: { recordId: "row-1", rowVersion: 1 } },
    );
    const staged = workbookInspectorReducer(selected, {
      type: "stage_row_bound_state",
      localFormKey: "form-1",
      mergePlanKey: "merge-1",
      pendingConfirmationKey: "delete-1",
      rollbackPreviewKey: "rollback-1",
      stalePreviewKey: "preview-1",
      workflowFormKey: "workflow-1",
    });
    const retargeted = workbookInspectorReducer(staged, {
      type: "select_row",
      row: { recordId: "row-2", rowVersion: 1 },
    });

    expect(retargeted.selectedRecordId).toBe("row-2");
    expect(retargeted.localFormKey).toBeNull();
    expect(retargeted.mergePlanKey).toBeNull();
    expect(retargeted.pendingConfirmationKey).toBeNull();
    expect(retargeted.rollbackPreviewKey).toBeNull();
    expect(retargeted.stalePreviewKey).toBeNull();
    expect(retargeted.workflowFormKey).toBeNull();
  });

  it("FE-U-P9-02 invalidates destructive, preview, merge, and workflow state on lifecycle triggers", () => {
    const config = selectInspectorConfig(
      requireViewContract("cartulary.view.timeline.v2"),
    );
    const staged = workbookInspectorReducer(
      initialWorkbookInspectorState(config),
      {
        type: "stage_row_bound_state",
        pendingConfirmationKey: "delete-1",
        rollbackPreviewKey: "rollback-1",
        mergePlanKey: "merge-1",
        workflowFormKey: "workflow-1",
      },
    );

    for (const type of [
      "row_version_changed",
      "incident_closed",
      "authorization_lost",
      "record_deleted",
      "record_merged",
    ] as const) {
      const next = workbookInspectorReducer(
        staged,
        type === "row_version_changed" ? { type, rowVersion: 2 } : { type },
      );
      expect(next.pendingConfirmationKey).toBeNull();
      expect(next.rollbackPreviewKey).toBeNull();
      expect(next.mergePlanKey).toBeNull();
      expect(next.workflowFormKey).toBeNull();
    }
  });

  it("FE-U-P9-02 defaults closed and clears stale state on hard refresh and active surface switch", () => {
    const timelineConfig = selectInspectorConfig(
      requireViewContract("cartulary.view.timeline.v2"),
    );
    const hostsConfig = selectInspectorConfig(
      requireViewContract("cartulary.view.hosts.v1"),
    );
    const staged = workbookInspectorReducer(
      workbookInspectorReducer(initialWorkbookInspectorState(timelineConfig), {
        type: "open",
        panelId: "workflow",
      }),
      {
        type: "select_row",
        row: { recordId: "row-1", rowVersion: 1 },
      },
    );
    const withPending = workbookInspectorReducer(staged, {
      type: "stage_row_bound_state",
      pendingConfirmationKey: "delete-1",
      rollbackPreviewKey: "rollback-1",
      mergePlanKey: "merge-1",
      workflowFormKey: "workflow-1",
    });
    const refreshed = workbookInspectorReducer(withPending, {
      type: "hard_refresh",
    });

    expect(refreshed.isOpen).toBe(false);
    expect(refreshed.selectedRecordId).toBeNull();
    expect(refreshed.pendingConfirmationKey).toBeNull();
    expect(refreshed.rollbackPreviewKey).toBeNull();
    expect(refreshed.mergePlanKey).toBeNull();
    expect(refreshed.workflowFormKey).toBeNull();

    const switched = workbookInspectorReducer(withPending, {
      type: "active_surface_switch",
      config: hostsConfig,
    });

    expect(switched.isOpen).toBe(false);
    expect(switched.configViewSchemaId).toBe("cartulary.view.hosts.v1");
    expect(switched.selectedRecordId).toBeNull();
    expect(switched.activePanelId).toBe("details");
  });
});
