import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type { TimelineCaptureSubject } from "./timelineCaptureActionModel";

export interface TimelineCandidatePort {
  page(
    cursor: string | null,
    signal: AbortSignal,
  ): Promise<
    | WorkbookOperationOutcome<{
        readonly rows: readonly TimelineCaptureSubject[];
        readonly nextCursor: string | null;
        readonly hasMore: boolean;
      }>
    | { readonly kind: "aborted" }
  >;
}
