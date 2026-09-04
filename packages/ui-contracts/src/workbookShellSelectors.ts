import type { StableTestId, WorkbookInspectorPanelId } from "./selectorCore";
import {
  requireClosedToken,
  requireFeatureGroupKey,
  requireFieldKey,
  requireRecordId,
  stableTestId,
} from "./selectorCore";
import { viewFirstTestId, viewScopedTestId } from "./viewSchemaSelectors";

export const workbookInspectorPanelIds = [
  "details",
  "relationships",
  "evidence",
  "history",
  "workflow",
] as const satisfies readonly WorkbookInspectorPanelId[];

export type WorkbookShellSlot =
  | "inspector"
  | "primary-grid"
  | "status-strip"
  | "top-bar"
  | "view-bar";

export const workbookShellSlots = [
  "top-bar",
  "view-bar",
  "primary-grid",
  "inspector",
  "status-strip",
] as const satisfies readonly WorkbookShellSlot[];

export const workbookShellSlotLabels = {
  inspector: "Inspector",
  "primary-grid": "Primary grid",
  "status-strip": "Status strip",
  "top-bar": "Workbook top bar",
  "view-bar": "View controls",
} as const satisfies Record<WorkbookShellSlot, string>;

export function workbookIncidentIdentityTestId(): StableTestId {
  return stableTestId("workbook-incident-identity");
}

export function workbookResponsiveBandTestId(): StableTestId {
  return stableTestId("workbook-responsive-band");
}

export function workbookSurfacesMenuTriggerTestId(): StableTestId {
  return stableTestId("workbook-surfaces-menu-trigger");
}

export function workbookSurfacesMenuTestId(): StableTestId {
  return stableTestId("workbook-surfaces-menu");
}

export function workbookSurfacesMenuOptionTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(
    viewScopedTestId("workbook-surfaces-menu-option", viewSchemaId),
  );
}

export function workbookViewBarQueryControlsTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "view-bar-query"));
}

export function workbookSortMenuTriggerTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "sort-menu-trigger"));
}

export function workbookSortMenuTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "sort-menu"));
}

export function workbookSortOptionTestId(
  viewSchemaId: string,
  fieldKey: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(viewSchemaId, `sort-option-${requireFieldKey(fieldKey)}`),
  );
}

export type WorkbookQueryEntryKind = "filter" | "group" | "sort";

export function workbookQueryEntryTestId(
  viewSchemaId: string,
  kind: WorkbookQueryEntryKind,
  fieldKey: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `query-entry-${requireClosedToken(
        ["filter", "group", "sort"] as const,
        kind,
        "workbook query entry kind",
      )}-${requireFieldKey(fieldKey)}`,
    ),
  );
}

export function workbookSortAppliedEntryTestId(
  viewSchemaId: string,
  fieldKey: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(viewSchemaId, `sort-applied-${requireFieldKey(fieldKey)}`),
  );
}

export function workbookGroupMenuTriggerTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "group-menu-trigger"));
}

export function workbookGroupMenuTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "group-menu"));
}

export function workbookColumnsMenuTriggerTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "columns-menu-trigger"));
}

export function workbookColumnsMenuTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "columns-menu"));
}

export function workbookFilterOperatorTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "filter-operator"));
}

export function workbookFilterClearButtonTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "filter-clear"));
}

export function workbookQueryOverflowEntryTestId(
  viewSchemaId: string,
  kind: WorkbookQueryEntryKind,
  fieldKey: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `query-overflow-${requireClosedToken(
        ["filter", "group", "sort"] as const,
        kind,
        "workbook query entry kind",
      )}-${requireFieldKey(fieldKey)}`,
    ),
  );
}

export function workbookFilterPopoverTriggerTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "filter-popover-trigger"));
}

export function workbookFilterPopoverTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "filter-popover"));
}

export function workbookShellReadyTestId(): string {
  return "workbook-shell-ready";
}

export function workbookShellSlotTestId(slot: WorkbookShellSlot): StableTestId {
  return stableTestId(`workbook-shell-slot-${requireWorkbookShellSlot(slot)}`);
}

export function workbookShellSlotLabel(slot: WorkbookShellSlot): string {
  return workbookShellSlotLabels[requireWorkbookShellSlot(slot)];
}

export function workbookAddRowButtonTestId(viewSchemaId: string): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "add-row"));
}

export function workbookInspectorToggleTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "inspector-toggle"));
}

export function workbookInspectorCloseButtonTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "inspector-close"));
}

export function workbookInspectorPanelTestId(
  viewSchemaId: string,
  panelId: WorkbookInspectorPanelId,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `inspector-panel-${requireWorkbookInspectorPanelId(panelId)}`,
    ),
  );
}

export function workbookInspectorFeatureActionTestId(
  viewSchemaId: string,
  featureGroupKey: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `inspector-feature-action-${requireFeatureGroupKey(featureGroupKey)}`,
    ),
  );
}

export function workbookInlineDraftRowTestId(
  viewSchemaId: string,
): StableTestId {
  return stableTestId(viewFirstTestId(viewSchemaId, "inline-draft-row"));
}

export function workbookRowActionMenuButtonTestId(
  viewSchemaId: string,
  recordId: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `row-action-menu-${requireRecordId(recordId)}`,
    ),
  );
}

export function workbookRowContextMenuTestId(
  viewSchemaId: string,
  recordId: string,
): StableTestId {
  return stableTestId(
    viewFirstTestId(
      viewSchemaId,
      `row-context-menu-${requireRecordId(recordId)}`,
    ),
  );
}

function requireWorkbookShellSlot(slot: WorkbookShellSlot): WorkbookShellSlot {
  return requireClosedToken(workbookShellSlots, slot, "workbook shell slot");
}

function requireWorkbookInspectorPanelId(
  panelId: WorkbookInspectorPanelId,
): WorkbookInspectorPanelId {
  return requireClosedToken(
    workbookInspectorPanelIds,
    panelId,
    "workbook inspector panel",
  );
}
