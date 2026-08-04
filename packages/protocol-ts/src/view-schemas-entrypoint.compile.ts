import {
  listViewSchemaRegistryEntries,
  type ViewInspectorPanelID,
  type ViewSchemaRegistryEntry,
  viewSchemaRegistry,
} from "@cartulary/protocol-ts/view-schemas";

declare const firstViewSchemaEntry: ViewSchemaRegistryEntry;
declare const panelID: ViewInspectorPanelID;

export const viewSchemaCompileSurface = {
  entries: listViewSchemaRegistryEntries(),
  firstViewSchemaEntry,
  panelID,
  registry: viewSchemaRegistry,
};
