import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
export type MentionCandidate = Readonly<{
  recordId: string;
  rowVersion: number;
  displayText: string;
  entityType: "host" | "identity";
}>;
export type MentionCandidateRequest = Readonly<{
  entityType: MentionCandidate["entityType"];
  search: string;
  cursor: string | null;
}>;
export interface TimelineMentionCandidatePort {
  readonly policy: Readonly<{
    pageSize: number;
    previousCursorLimit: number;
    settledInputMs: number;
  }>;
  page(
    request: MentionCandidateRequest,
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
