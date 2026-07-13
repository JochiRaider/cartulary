import type {
  InspectorConfig,
  InspectorFeatureGroup,
  InspectorPanelId,
  ViewContract,
} from "@cartulary/view-contracts";

export type WorkbookInspectorRowSubject = {
  readonly recordId: string | null;
  readonly rowVersion: number | null;
};

export type WorkbookInspectorState = {
  readonly activePanelId: InspectorPanelId | null;
  readonly configViewSchemaId: string;
  readonly isOpen: boolean;
  readonly localFormKey: string | null;
  readonly mergePlanKey: string | null;
  readonly pendingConfirmationKey: string | null;
  readonly rollbackPreviewKey: string | null;
  readonly selectedRecordId: string | null;
  readonly selectedRowVersion: number | null;
  readonly stalePreviewKey: string | null;
  readonly workflowFormKey: string | null;
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
      readonly type: "select_row";
      readonly row: WorkbookInspectorRowSubject;
    }
  | { readonly type: "row_version_changed"; readonly rowVersion: number | null }
  | {
      readonly type:
        | "incident_closed"
        | "authorization_lost"
        | "record_deleted"
        | "record_merged"
        | "hard_refresh";
    }
  | {
      readonly type: "active_surface_switch";
      readonly config: InspectorConfig;
    }
  | {
      readonly type: "stage_row_bound_state";
      readonly localFormKey?: string | null | undefined;
      readonly mergePlanKey?: string | null | undefined;
      readonly pendingConfirmationKey?: string | null | undefined;
      readonly rollbackPreviewKey?: string | null | undefined;
      readonly stalePreviewKey?: string | null | undefined;
      readonly workflowFormKey?: string | null | undefined;
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
    isOpen: config.defaultOpen,
    localFormKey: null,
    mergePlanKey: null,
    pendingConfirmationKey: null,
    rollbackPreviewKey: null,
    selectedRecordId: null,
    selectedRowVersion: null,
    stalePreviewKey: null,
    workflowFormKey: null,
  };
}

export function workbookInspectorReducer(
  state: WorkbookInspectorState,
  action: WorkbookInspectorAction,
): WorkbookInspectorState {
  switch (action.type) {
    case "reset_config":
      return initialWorkbookInspectorState(action.config);
    case "open":
      return {
        ...state,
        activePanelId: action.panelId ?? state.activePanelId,
        isOpen: true,
      };
    case "close":
      return {
        ...clearRowBoundInspectorState(state),
        isOpen: false,
      };
    case "select_row":
      if (
        state.selectedRecordId === action.row.recordId &&
        state.selectedRowVersion === action.row.rowVersion
      ) {
        return state;
      }
      return {
        ...clearRowBoundInspectorState(state),
        selectedRecordId: action.row.recordId,
        selectedRowVersion: action.row.rowVersion,
      };
    case "row_version_changed":
      if (state.selectedRowVersion === action.rowVersion) {
        return state;
      }
      return {
        ...clearRowBoundInspectorState(state),
        selectedRowVersion: action.rowVersion,
      };
    case "incident_closed":
    case "authorization_lost":
    case "record_deleted":
    case "record_merged":
      return clearRowBoundInspectorState(state);
    case "hard_refresh":
      return {
        ...clearRowBoundInspectorState(state),
        activePanelId: state.activePanelId,
        isOpen: false,
        selectedRecordId: null,
        selectedRowVersion: null,
      };
    case "active_surface_switch":
      return initialWorkbookInspectorState(action.config);
    case "stage_row_bound_state":
      return {
        ...state,
        localFormKey: action.localFormKey ?? state.localFormKey,
        mergePlanKey: action.mergePlanKey ?? state.mergePlanKey,
        pendingConfirmationKey:
          action.pendingConfirmationKey ?? state.pendingConfirmationKey,
        rollbackPreviewKey:
          action.rollbackPreviewKey ?? state.rollbackPreviewKey,
        stalePreviewKey: action.stalePreviewKey ?? state.stalePreviewKey,
        workflowFormKey: action.workflowFormKey ?? state.workflowFormKey,
      };
  }
}

function clearRowBoundInspectorState(
  state: WorkbookInspectorState,
): WorkbookInspectorState {
  return {
    ...state,
    localFormKey: null,
    mergePlanKey: null,
    pendingConfirmationKey: null,
    rollbackPreviewKey: null,
    stalePreviewKey: null,
    workflowFormKey: null,
  };
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
