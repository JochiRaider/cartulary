import type { WorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookCanonicalQuery } from "../query/WorkbookViewQueryPort";
import type { WorkbookPortResult } from "./WorkbookPortResult";

export type WorkbookCandidate = Readonly<{
  recordId: string;
  displayText: string;
}>;
export type WorkbookCandidateQuery = Readonly<{
  queryState: WorkbookQueryState;
  cursor: string | null;
  signal: AbortSignal;
  expectedCanonicalQuery?: WorkbookCanonicalQuery;
  /** Fences adapter authority callbacks even when a transport ignores cancellation. */
  isCurrent?: () => boolean;
  /** Adapter rechecks delegate through the current observer, once per read. */
  onAuthorityFailure?: (failure: WorkbookOperationFailure) => void;
}>;
export type WorkbookCandidatePage<T extends WorkbookCandidate> = Readonly<{
  candidates: readonly T[];
  hasMore: boolean;
  nextCursor: string | null;
  canonicalQuery?: WorkbookCanonicalQuery;
}>;
export type WorkbookCandidateReader<T extends WorkbookCandidate> = (
  input: WorkbookCandidateQuery,
) => Promise<WorkbookPortResult<WorkbookCandidatePage<T>>>;
