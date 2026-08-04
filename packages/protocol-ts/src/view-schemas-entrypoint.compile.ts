import {
  listViewSchemaRegistryEntries,
  type ViewSchemaRegistryEntry,
  type ViewSchemaSourceDocument,
  viewSchemaArtifacts,
  viewSchemaRegistry,
} from "@cartulary/protocol-ts/view-schemas";

declare const source: ViewSchemaSourceDocument;
declare const firstViewSchemaEntry: ViewSchemaRegistryEntry;

export const viewSchemaCompileSurface = {
  artifacts: viewSchemaArtifacts,
  entries: listViewSchemaRegistryEntries(),
  firstViewSchemaEntry,
  registry: viewSchemaRegistry,
  source,
};
