import { workbookInspectorFeatureActionTestId } from "@cartulary/ui-contracts";
import type {
  InspectorConfig,
  InspectorDisabledCondition,
  InspectorFeatureGroup,
} from "@cartulary/view-contracts";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import {
  inspectorFeatureDisabledTokens,
  resolveSemanticInspectorFeature,
} from "../semanticInspectorDispatcher";

export const workbookInspectorNoRowMessage =
  "Select a saved row to inspect its details.";

export type WorkbookInspectorSubjectPresentation = {
  readonly label: string;
  readonly recordId: string;
  readonly rowVersion: number | null;
  readonly stateLabel?: string | undefined;
  readonly surfaceLabel: string;
};

export type WorkbookInspectorTechnicalField = {
  readonly label: string;
  readonly value: string;
};

export type WorkbookInspectorActionBinding = {
  readonly featureGroup: InspectorFeatureGroup;
  readonly semanticKey: string;
  readonly testId: string;
  readonly viewSchemaId: string;
};

export type WorkbookHistoryEventPresentation = {
  readonly actorLabel?: string | undefined;
  readonly committedAt: string;
  readonly key: string;
  readonly operation: string;
  readonly summary: string;
  readonly technicalFields: readonly WorkbookInspectorTechnicalField[];
};

export function bindWorkbookInspectorAction(
  config: InspectorConfig,
  featureGroupKey: string,
): WorkbookInspectorActionBinding | null {
  const resolution = resolveSemanticInspectorFeature(config, featureGroupKey);
  if (
    resolution.kind === "unsupported" ||
    resolution.disposition === "panel_read"
  ) {
    return null;
  }
  return {
    featureGroup: resolution.featureGroup,
    semanticKey: resolution.semanticKey,
    testId: workbookInspectorFeatureActionTestId(
      config.viewSchemaId,
      resolution.featureGroup.featureGroupKey,
    ),
    viewSchemaId: config.viewSchemaId,
  };
}

export function workbookInspectorDisabledReason({
  currentIncidentRole,
  featureGroup,
  stateTokens,
}: {
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly featureGroup: InspectorFeatureGroup;
  readonly stateTokens: ReadonlySet<InspectorDisabledCondition>;
}): string | null {
  const activeTokens = inspectorFeatureDisabledTokens({
    currentIncidentRole,
    featureGroup,
    stateTokens,
  });
  const token = featureGroup.disabledWhen.find((candidate) =>
    activeTokens.has(candidate),
  );
  if (token === undefined) return null;
  if (
    token === "authorization_lost" &&
    featureGroup.minimumIncidentRole !== null &&
    currentIncidentRole !== null
  ) {
    return `Requires the ${featureGroup.minimumIncidentRole} incident role.`;
  }
  return disabledReasonByToken[token];
}

const disabledReasonByToken = {
  no_row_selected: "Select a saved row to use this action.",
  incident_closed: "This incident is closed and read-only.",
  authorization_lost: "You no longer have access to this action.",
  row_version_changed: "This row changed; refresh it before retrying.",
  record_deleted: "This action is unavailable for a deleted record.",
  record_merged: "This record was merged and can no longer be changed.",
  evidence_preview_unavailable: "Preview is unavailable for this evidence.",
  merge_target_unavailable: "Select a valid merge target.",
  record_not_deleted: "This action is available only for deleted records.",
  rollback_target_unavailable:
    "Select an available history change to roll back.",
  party_text_unavailable: "No party reference text is available to link.",
  pivot_target_unavailable: "No matching destination is available.",
} as const satisfies Record<InspectorDisabledCondition, string>;
