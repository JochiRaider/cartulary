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
      collectionActions: { "host.aliases": "alias" },
    }),
  },
  {
    viewSchemaId: identitiesViewSchemaId,
    ownerId: "entities_observations",
    renderer: "entity_identities",
    policy: defineWorkbookSurfacePolicy({
      collectionActions: { "identity.aliases": "alias" },
    }),
  },
  {
    viewSchemaId: indicatorsViewSchemaId,
    ownerId: "entities_observations",
    renderer: "contract",
    policy: defineWorkbookSurfacePolicy(),
  },
] as const satisfies readonly WorkbookSurfacePolicyDefinition[];
