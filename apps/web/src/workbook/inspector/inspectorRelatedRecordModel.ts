import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import {
  initialGenericCreateDraft,
  workbookCreationAvailable,
} from "../models/genericWorkbookModel";
import { stringifyGridValue } from "../utils/workbookValueFormat";
import type { WorkbookInspectorErrorPresentation } from "./workbookInspectorErrorModel";
import {
  type WorkbookInspectorLiveRowBinding,
  type WorkbookInspectorLiveSubject,
  workbookInspectorSubjectsEqual,
} from "./workbookInspectorSubject";

type InspectorRelatedRecordDraftResult =
  | {
      readonly kind: "ready";
      readonly draft: Record<string, string>;
    }
  | {
      readonly kind: "invalid_target";
      readonly reason: "creation_unavailable" | "semantic_mismatch";
    };

export type InspectorRelatedRecordFormModel = {
  readonly draft: Record<string, string>;
  readonly featureGroup: InspectorFeatureGroup;
  readonly error: WorkbookInspectorErrorPresentation | null;
  readonly targetContract: ViewContract;
};

export type InspectorRelatedRecordWorkflowState =
  InspectorRelatedRecordFormModel & {
    readonly workflowId: symbol;
    readonly subject: WorkbookInspectorLiveSubject;
    readonly phase: "editing" | "submitting";
  };

export type InspectorRelatedRecordWorkflowAction =
  | {
      readonly type: "begin";
      readonly workflowId: symbol;
      readonly subject: WorkbookInspectorLiveSubject;
      readonly draft: Record<string, string>;
      readonly featureGroup: InspectorFeatureGroup;
      readonly targetContract: ViewContract;
    }
  | {
      readonly type: "update";
      readonly workflowId: symbol;
      readonly fieldKey: string;
      readonly value: string;
    }
  | { readonly type: "submit"; readonly workflowId: symbol }
  | {
      readonly type: "reject";
      readonly workflowId: symbol;
      readonly error: WorkbookInspectorErrorPresentation;
    }
  | { readonly type: "complete"; readonly workflowId: symbol }
  | { readonly type: "cancel"; readonly workflowId: symbol }
  | {
      readonly type: "retarget";
      readonly workflowId: symbol;
      readonly subject: WorkbookInspectorLiveSubject | null;
    };

export function inspectorRelatedRecordWorkflowReducer(
  state: InspectorRelatedRecordWorkflowState | null,
  action: InspectorRelatedRecordWorkflowAction,
): InspectorRelatedRecordWorkflowState | null {
  if (action.type === "begin") {
    return {
      draft: action.draft,
      error: null,
      featureGroup: action.featureGroup,
      phase: "editing",
      subject: action.subject,
      targetContract: action.targetContract,
      workflowId: action.workflowId,
    };
  }
  if (state === null || state.workflowId !== action.workflowId) return state;
  switch (action.type) {
    case "update":
      return state.phase === "editing"
        ? {
            ...state,
            draft: { ...state.draft, [action.fieldKey]: action.value },
            error: null,
          }
        : state;
    case "submit":
      return state.phase === "editing"
        ? { ...state, error: null, phase: "submitting" }
        : state;
    case "reject":
      return state.phase === "submitting"
        ? { ...state, error: action.error, phase: "editing" }
        : state;
    case "complete":
      return state.phase === "submitting" ? null : state;
    case "cancel":
      return null;
    case "retarget":
      return workbookInspectorSubjectsEqual(state.subject, action.subject)
        ? state
        : null;
  }
}

export function buildInspectorRelatedRecordDraft({
  currentUserId,
  featureGroup,
  subject,
  targetContract,
}: {
  readonly currentUserId: string | null;
  readonly featureGroup: InspectorFeatureGroup;
  readonly subject: WorkbookInspectorLiveRowBinding;
  readonly targetContract: ViewContract;
}): InspectorRelatedRecordDraftResult {
  const route = featureGroup.routeBinding;
  if (
    route.kind !== "view_row_create" ||
    route.owner !== "view_row_create_route" ||
    route.targetViewSchemaId !== targetContract.viewSchemaId
  ) {
    return { kind: "invalid_target", reason: "semantic_mismatch" };
  }
  if (!workbookCreationAvailable(targetContract)) {
    return { kind: "invalid_target", reason: "creation_unavailable" };
  }

  const draft = initialGenericCreateDraft(targetContract, currentUserId);
  for (const binding of featureGroup.seedBindings) {
    const value = inspectorRelatedRecordSeedValue(binding.source, subject);
    if (value !== null) draft[binding.targetFieldKey] = value;
  }
  return { kind: "ready", draft };
}

function inspectorRelatedRecordSeedValue(
  source: InspectorFeatureGroup["seedBindings"][number]["source"],
  subject: WorkbookInspectorLiveRowBinding,
): string | null {
  if (source.kind === "selected_record_id") return subject.subject.recordId;
  if (source.kind === "selected_field_value") {
    if (source.sourceFieldKey === undefined) return null;
    const text = stringifyGridValue(
      subject.cells[source.sourceFieldKey]?.value,
    ).trim();
    return text === "" ? null : text;
  }
  if (source.value === null || source.value === undefined) return null;
  return typeof source.value === "string"
    ? source.value
    : JSON.stringify(source.value);
}
