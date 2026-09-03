export type WorkbookGridEntryFocusRequest =
  | { readonly kind: "idle" }
  | {
      readonly kind: "pending";
      readonly viewSchemaId: string;
      readonly generation: number;
    };

export type WorkbookGridEntryFocusAcknowledgement = {
  readonly viewSchemaId: string;
  readonly generation: number;
};

export type WorkbookGridEntryFocusOwner = {
  readonly request: WorkbookGridEntryFocusRequest;
  readonly acknowledge: (
    acknowledgement: WorkbookGridEntryFocusAcknowledgement,
  ) => void;
};
