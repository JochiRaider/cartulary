import type {
  InspectorConfig,
  InspectorFeatureGroup,
  InspectorPanelId,
  ViewContract,
} from "@cartulary/view-contracts";

export type WorkbookInspectorSubject = {
  readonly viewSchemaId: string;
  readonly recordId: string;
  readonly rowVersion: number;
};

export type WorkbookInspectorStatus = "closed" | "no_row_selected" | "ready";

export type WorkbookInspectorInvalidationReason =
  | "action_completed"
  | "authorization_lost"
  | "hard_refresh"
  | "incident_closed"
  | "record_deleted"
  | "record_merged"
  | "surface_changed";

export type WorkbookInspectorState = {
  readonly activePanelId: InspectorPanelId | null;
  readonly configViewSchemaId: string;
  readonly invalidationGeneration: number;
  readonly invalidationCause:
    | WorkbookInspectorInvalidationReason
    | "close"
    | "retarget"
    | null;
  readonly isOpen: boolean;
  readonly status: WorkbookInspectorStatus;
  readonly subject: WorkbookInspectorSubject | null;
};

export type WorkbookInspectorAction =
  | {
      readonly type: "reset_config";
      readonly config: InspectorConfig;
    }
  | {
      readonly type: "open";
      readonly panelId?: InspectorPanelId | undefined;
    }
  | { readonly type: "close" }
  | {
      readonly type: "retarget";
      readonly subject: WorkbookInspectorSubject | null;
    }
  | {
      readonly type: "select_panel";
      readonly panelId: InspectorPanelId;
    }
  | {
      readonly type: "invalidate";
      readonly reason: WorkbookInspectorInvalidationReason;
    };

export function selectInspectorConfig(contract: ViewContract): InspectorConfig {
  return contract.inspectorConfig;
}

export function initialWorkbookInspectorState(
  config: InspectorConfig,
): WorkbookInspectorState {
  return {
    activePanelId: firstPanelId(config),
    configViewSchemaId: config.viewSchemaId,
    invalidationGeneration: 0,
    invalidationCause: null,
    isOpen: false,
    status: "closed",
    subject: null,
  };
}

export function workbookInspectorReducer(
  state: WorkbookInspectorState,
  action: WorkbookInspectorAction,
): WorkbookInspectorState {
  switch (action.type) {
    case "reset_config":
      return {
        ...initialWorkbookInspectorState(action.config),
        invalidationCause: "surface_changed",
        invalidationGeneration: state.invalidationGeneration + 1,
      };
    case "open":
      return {
        ...state,
        activePanelId: action.panelId ?? state.activePanelId,
        isOpen: true,
        status: state.subject === null ? "no_row_selected" : "ready",
      };
    case "close":
      return {
        ...state,
        invalidationCause: "close",
        invalidationGeneration: state.invalidationGeneration + 1,
        isOpen: false,
        status: "closed",
      };
    case "retarget":
      if (workbookInspectorSubjectsEqual(state.subject, action.subject)) {
        return state;
      }
      return {
        ...state,
        invalidationCause: "retarget",
        invalidationGeneration: state.invalidationGeneration + 1,
        status: state.isOpen
          ? action.subject === null
            ? "no_row_selected"
            : "ready"
          : "closed",
        subject: action.subject,
      };
    case "select_panel":
      return {
        ...state,
        activePanelId: action.panelId,
      };
    case "invalidate":
      return {
        ...state,
        invalidationCause: action.reason,
        invalidationGeneration: state.invalidationGeneration + 1,
        isOpen: action.reason === "action_completed" ? state.isOpen : false,
        status:
          action.reason === "action_completed"
            ? state.isOpen
              ? state.subject === null
                ? "no_row_selected"
                : "ready"
              : "closed"
            : "closed",
        subject: action.reason === "action_completed" ? state.subject : null,
      };
  }
}

export function workbookInspectorSubjectsEqual(
  left: WorkbookInspectorSubject | null,
  right: WorkbookInspectorSubject | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      left.viewSchemaId === right.viewSchemaId &&
      left.recordId === right.recordId &&
      left.rowVersion === right.rowVersion)
  );
}

export function inspectorPanelIsDeclared(
  config: InspectorConfig,
  panelId: InspectorPanelId,
): boolean {
  return config.panels.some((panel) => panel.panelId === panelId);
}

export function inspectorFeatureGroupsForPanel(
  config: InspectorConfig,
  panelId: InspectorPanelId,
): readonly InspectorFeatureGroup[] {
  if (!inspectorPanelIsDeclared(config, panelId)) {
    return [];
  }
  return config.featureGroups.filter((group) => group.panelId === panelId);
}

export function inspectorNoRowState(config: InspectorConfig): string {
  return config.noRowState;
}

function firstPanelId(config: InspectorConfig): InspectorPanelId | null {
  return config.panels[0]?.panelId ?? null;
}
