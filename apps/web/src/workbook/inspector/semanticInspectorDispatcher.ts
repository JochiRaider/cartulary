import type {
  InspectorConfig,
  InspectorDisabledCondition,
  InspectorFeatureGroup,
} from "@cartulary/view-contracts";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";

export type SemanticInspectorDisposition =
  | "panel_read"
  | "contextual_workflow_or_pivot"
  | "direct_history_action"
  | "existing_owner_control";

export type SemanticInspectorFeatureResolution =
  | {
      readonly kind: "supported";
      readonly disposition: SemanticInspectorDisposition;
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | { readonly kind: "unsupported" };

// Each value pins the ordered set of complete stable semantic tuples for one
// current view schema: view schema, feature key, panel, route kind, route
// owner, and action key. A changed or additive tuple fails closed until this
// authored registration is deliberately updated and reviewed.
const registeredInspectorConfigFingerprints: Readonly<Record<string, string>> =
  Object.freeze({
    "cartulary.view.timeline.v2": "27:f988d895",
    "cartulary.view.hosts.v1": "16:4a395ac9",
    "cartulary.view.identities.v1": "16:1c096141",
    "cartulary.view.evidence.v1": "19:f9f930e7",
    "cartulary.view.notes.v1": "14:67c3c511",
    "cartulary.view.indicators.v1": "12:a93fc881",
    "cartulary.view.assessments.v1": "11:275a8346",
    "cartulary.view.task_requests.v1": "16:a22c9b79",
    "cartulary.view.decisions.v1": "14:11b62214",
    "cartulary.view.parties.v1": "13:00c29d32",
    "cartulary.view.comm_log.v1": "13:2b546c22",
    "cartulary.view.handoff.v1": "13:0244a3e3",
    "cartulary.view.status_review.v1": "14:a27703d9",
    "cartulary.view.lesson.v1": "12:0b28e2af",
    "cartulary.view.findings.v1": "14:610cd279",
    "cartulary.view.investigative_queries.v1": "12:85a5d5ff",
    "cartulary.view.forensic_keywords.v1": "11:a8202f53",
  });

const roleRank: Readonly<Record<Exclude<WorkbookIncidentRole, "">, number>> = {
  viewer: 0,
  editor: 1,
  reviewer: 2,
  admin: 3,
};

export function semanticInspectorFeatureKey(
  viewSchemaId: string,
  featureGroup: InspectorFeatureGroup,
): string {
  return [
    viewSchemaId,
    featureGroup.featureGroupKey,
    featureGroup.panelId,
    featureGroup.routeBinding.kind,
    featureGroup.routeBinding.owner,
    featureGroup.routeBinding.actionKey ?? "",
  ].join("|");
}

export function resolveSemanticInspectorFeature(
  config: InspectorConfig,
  featureGroupKey: string,
): SemanticInspectorFeatureResolution {
  if (!inspectorConfigIsExplicitlyRegistered(config)) {
    return { kind: "unsupported" };
  }
  const featureGroup = config.featureGroups.find(
    (candidate) => candidate.featureGroupKey === featureGroupKey,
  );
  if (featureGroup === undefined) {
    return { kind: "unsupported" };
  }
  return {
    kind: "supported",
    disposition: semanticInspectorDisposition(featureGroup),
    featureGroup,
    semanticKey: semanticInspectorFeatureKey(config.viewSchemaId, featureGroup),
  };
}

export function inspectorFeatureDisabledTokens({
  currentIncidentRole,
  featureGroup,
  stateTokens,
}: {
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly featureGroup: InspectorFeatureGroup;
  readonly stateTokens: ReadonlySet<InspectorDisabledCondition>;
}): ReadonlySet<InspectorDisabledCondition> {
  const tokens = new Set(stateTokens);
  const minimumRole = featureGroup.minimumIncidentRole;
  const currentRole = currentIncidentRole || null;
  if (
    minimumRole !== null &&
    (currentRole === null || roleRank[currentRole] < roleRank[minimumRole])
  ) {
    tokens.add("authorization_lost");
  }
  return tokens;
}

function semanticInspectorDisposition(
  featureGroup: InspectorFeatureGroup,
): SemanticInspectorDisposition {
  if (featureGroup.routeBinding.kind === "panel_read") {
    return "panel_read";
  }
  if (
    featureGroup.routeBinding.kind === "view_row_create" ||
    featureGroup.routeBinding.kind === "surface_pivot"
  ) {
    return "contextual_workflow_or_pivot";
  }
  if (
    featureGroup.panelId === "history" &&
    (featureGroup.featureGroupKey === "record.delete" ||
      featureGroup.featureGroupKey === "record.restore" ||
      featureGroup.featureGroupKey === "history.rollback")
  ) {
    return "direct_history_action";
  }
  return "existing_owner_control";
}

function inspectorConfigIsExplicitlyRegistered(
  config: InspectorConfig,
): boolean {
  const expected = registeredInspectorConfigFingerprints[config.viewSchemaId];
  return (
    expected !== undefined && expected === inspectorConfigFingerprint(config)
  );
}

function inspectorConfigFingerprint(config: InspectorConfig): string {
  const corpus = config.featureGroups
    .map((featureGroup) =>
      semanticInspectorFeatureKey(config.viewSchemaId, featureGroup),
    )
    .join("\n");
  let hash = 2_166_136_261;
  for (let index = 0; index < corpus.length; index += 1) {
    hash ^= corpus.charCodeAt(index);
    hash = Math.imul(hash, 16_777_619) >>> 0;
  }
  return `${config.featureGroups.length}:${hash.toString(16).padStart(8, "0")}`;
}
