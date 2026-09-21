import { workbookInspectorFeatureActionTestId } from "@cartulary/ui-contracts";
import type {
  InspectorConfig,
  InspectorDisabledCondition,
  InspectorFeatureGroup,
} from "@cartulary/view-contracts";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { InspectorContextualCapability } from "../inspectorCapabilityResolver";

export const workbookInspectorNoRowMessage =
  "Select a saved row to inspect its details.";

export type WorkbookInspectorTechnicalField = {
  readonly label: string;
  readonly value: string;
};

export type WorkbookInspectorActionBinding = {
  readonly capability: InspectorContextualCapability;
  readonly featureGroup: InspectorFeatureGroup;
  readonly semanticKey: string;
  readonly testId: string;
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
  capability: InspectorContextualCapability,
): WorkbookInspectorActionBinding {
  return {
    capability,
    featureGroup: capability.featureGroup,
    semanticKey: capability.semanticKey,
    testId: workbookInspectorFeatureActionTestId(
      config.viewSchemaId,
      capability.featureGroup.featureGroupKey,
    ),
  };
}

export type WorkbookInspectorDisabledReason =
  | {
      readonly kind: "condition";
      readonly condition: InspectorDisabledCondition;
    }
  | {
      readonly kind: "minimum_role";
      readonly role: Exclude<WorkbookIncidentRole, "">;
    }
  | {
      readonly kind: "owner";
      readonly owner: string;
      readonly code: string;
      readonly parameters: Readonly<Record<string, string | number | boolean>>;
      readonly message: string;
    };

// Each source owns its closed cause vocabulary. Shared presentation groups only
// identity and parameters, never message text or feature implementation details.
export function ownerInspectorDisabledReason<Code extends string>(
  owner: string,
  code: Code,
  message: string,
  parameters: Readonly<Record<string, string | number | boolean>> = {},
): WorkbookInspectorDisabledReason {
  return { kind: "owner", owner, code, parameters, message };
}
export function workbookInspectorDisabledReasonKey(
  reason: WorkbookInspectorDisabledReason,
): string {
  switch (reason.kind) {
    case "condition":
      return JSON.stringify([reason.kind, reason.condition]);
    case "minimum_role":
      return JSON.stringify([reason.kind, reason.role]);
    case "owner":
      return JSON.stringify([
        reason.kind,
        reason.owner,
        reason.code,
        Object.entries(reason.parameters).sort(([a], [b]) =>
          a.localeCompare(b),
        ),
      ]);
  }
}
export function workbookInspectorDisabledReasonText(
  reason: WorkbookInspectorDisabledReason,
): string {
  switch (reason.kind) {
    case "condition":
      return disabledReasonByToken[reason.condition];
    case "minimum_role":
      return `Requires the ${reason.role} incident role.`;
    case "owner":
      return reason.message;
  }
}

export function workbookInspectorDisabledReason({
  currentIncidentRole,
  featureGroup,
  stateTokens,
}: {
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly featureGroup: InspectorFeatureGroup;
  readonly stateTokens: ReadonlySet<InspectorDisabledCondition>;
}): WorkbookInspectorDisabledReason | null {
  const activeTokens = new Set(stateTokens);
  const minimumRole = featureGroup.minimumIncidentRole;
  const currentRole = currentIncidentRole || null;
  if (
    minimumRole !== null &&
    (currentRole === null || roleRank[currentRole] < roleRank[minimumRole])
  ) {
    activeTokens.add("authorization_lost");
  }
  const token = featureGroup.disabledWhen.find((candidate) =>
    activeTokens.has(candidate),
  );
  if (token === undefined) return null;
  if (
    token === "authorization_lost" &&
    featureGroup.minimumIncidentRole !== null &&
    currentIncidentRole !== null
  ) {
    return { kind: "minimum_role", role: featureGroup.minimumIncidentRole };
  }
  return { kind: "condition", condition: token };
}

const roleRank: Readonly<Record<Exclude<WorkbookIncidentRole, "">, number>> = {
  viewer: 0,
  editor: 1,
  reviewer: 2,
  admin: 3,
};

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
