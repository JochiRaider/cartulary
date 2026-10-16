import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  defineWorkbookSurfacePolicy,
  type WorkbookSurfacePolicyDefinition,
} from "./workbookSurfacePolicy";

export const entitiesObservationsSurfacePolicies = [
  {
    viewSchemaId: hostsViewSchemaId,
    ownerId: "entities_observations",
    renderer: "entity_hosts",
    policy: defineWorkbookSurfacePolicy({
      collectionActions: {
        "host.aliases": "alias",
        "host.subject_party_ids": "party",
        "host.responsible_party_ids": "party",
        "host.observer_party_ids": "party",
      },
    }),
  },
  {
    viewSchemaId: identitiesViewSchemaId,
    ownerId: "entities_observations",
    renderer: "entity_identities",
    policy: defineWorkbookSurfacePolicy({
      collectionActions: {
        "identity.aliases": "alias",
        "identity.subject_party_ids": "party",
        "identity.responsible_party_ids": "party",
        "identity.observer_party_ids": "party",
      },
    }),
  },
  {
    viewSchemaId: indicatorsViewSchemaId,
    ownerId: "entities_observations",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy(),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
