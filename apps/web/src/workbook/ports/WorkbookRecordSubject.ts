/** Accepted record identity and display context shared by History and inspector attachment. */
type WorkbookRecordSubjectIdentity = {
  readonly kind: "live" | "deleted";
  readonly recordId: string;
  readonly rowVersion: number;
  readonly viewSchemaId: string;
};

type WorkbookRecordSubjectContext = WorkbookRecordSubjectIdentity & {
  readonly label: string;
  readonly surfaceLabel: string;
};

export type WorkbookRecordSubject =
  | (WorkbookRecordSubjectContext & {
      readonly kind: "live";
      readonly stateLabel?: string | undefined;
    })
  | (WorkbookRecordSubjectContext & {
      readonly kind: "deleted";
      readonly stateLabel: string;
    });
