import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
export type MentionCandidate = Readonly<{
  recordId: string;
  rowVersion: number;
  displayText: string;
  entityType: "host" | "identity";
}>;
export interface TimelineMentionCandidatePort {
  page(
    entityType: MentionCandidate["entityType"],
    cursor: string | null,
    signal: AbortSignal,
  ): Promise<
    | WorkbookOperationOutcome<
        Readonly<{
          candidates: readonly MentionCandidate[];
          nextCursor: string | null;
          hasMore: boolean;
        }>
      >
    | { readonly kind: "aborted" }
  >;
}
