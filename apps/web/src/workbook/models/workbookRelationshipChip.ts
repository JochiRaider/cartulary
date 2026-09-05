import type { EntityType } from "@cartulary/ui-contracts";

export type WorkbookRelationshipChipState =
  | "unresolved"
  | "resolved"
  | "auto_resolved"
  | "dismissed";

export type WorkbookRelationshipChipPresentation = {
  readonly entityType: EntityType;
  readonly source: {
    readonly kind: "entity_mention";
    readonly recordId: string | null;
    readonly fieldKey: string | null;
    readonly itemRef: string;
  };
  readonly rawText: string;
  readonly targetRecordId: string | null;
  readonly previousTarget: {
    readonly recordId: string;
    readonly label: string;
  } | null;
  readonly resolution: {
    readonly method: "manual" | "auto" | "import" | "system" | null;
    readonly sourceMethod: string | null;
    readonly provenance: string | null;
    readonly confidence: number | null;
    readonly matchedAliasText: string | null;
  };
  readonly label: string;
  readonly onSelect?: (() => void) | undefined;
  readonly selected: boolean;
  readonly selectorIdentity: string;
  readonly state: WorkbookRelationshipChipState;
};

export function relationshipChipAccessibleName(
  chip: WorkbookRelationshipChipPresentation,
): string {
  switch (chip.state) {
    case "unresolved":
      return `Unresolved ${chip.entityType} mention: ${chip.rawText}`;
    case "dismissed":
      return `Dismissed mention: ${chip.rawText}`;
    case "auto_resolved":
      return `Auto-resolved ${chip.entityType}: ${chip.label}${chip.resolution.matchedAliasText === null ? "" : `; matched ${chip.resolution.matchedAliasText}`}`;
    case "resolved":
      return `Resolved ${chip.entityType}: ${chip.label}`;
  }
}
