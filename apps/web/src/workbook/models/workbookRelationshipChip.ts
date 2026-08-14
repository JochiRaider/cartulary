export type WorkbookRelationshipChipState =
  | "unresolved"
  | "resolved"
  | "auto_resolved"
  | "dismissed";

export type WorkbookRelationshipChipPresentation = {
  readonly accessibleDetail?: string | undefined;
  readonly label: string;
  readonly onSelect?: (() => void) | undefined;
  readonly selected: boolean;
  readonly selectorIdentity: string;
  readonly state: WorkbookRelationshipChipState;
};
