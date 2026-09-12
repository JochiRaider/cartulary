import { requireViewContract } from "@cartulary/view-contracts";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { admitCanonicalInspectorFeature } from "../../inspector/canonicalInspectorAdmission";
import { workbookInspectorDisabledReason } from "../../inspector/presentation/workbookInspectorPresentationModel";
import { workbookCreationAvailable } from "../../models/genericWorkbookModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { mentionEntityContract } from "./timelineMentionCreationModel";
import type {
  MentionAction,
  MentionAuthority,
} from "./timelineMentionOperationModel";
export function timelineMentionAuthority(
  authority: {
    actorId: string;
    incidentId: string;
    sessionIdentity: string;
    role: WorkbookIncidentRole;
    closed: boolean;
  } | null,
): MentionAuthority | null {
  if (!authority) return null;
  const config = requireViewContract(timelineViewSchemaId).inspectorConfig;
  const admitted = (key: string) => {
    const feature = admitCanonicalInspectorFeature(config, key);
    return (
      feature?.routeBinding.kind === "entity_mention_action" &&
      feature.routeBinding.owner === "entity_mention_resolve_route" &&
      workbookInspectorDisabledReason({
        currentIncidentRole: authority.role,
        featureGroup: feature,
        stateTokens: new Set(authority.closed ? ["incident_closed"] : []),
      }) === null
    );
  };
  const actions: MentionAction["action"][] = [];
  if (admitted("entity_mentions.resolve")) actions.push("resolve_item");
  if (admitted("entity_mentions.dismiss")) actions.push("dismiss_item");
  if (admitted("entity_mentions.restore")) actions.push("revert_to_unresolved");
  return {
    actorId: authority.actorId,
    incidentId: authority.incidentId,
    sessionIdentity: authority.sessionIdentity,
    role: authority.role,
    mutationsAvailable: !authority.closed,
    actions,
    createTypes: (["host", "identity"] as const).filter(
      (type) =>
        admitted(`entity_mentions.create_${type}`) &&
        workbookCreationAvailable(mentionEntityContract(type)),
    ),
  };
}
