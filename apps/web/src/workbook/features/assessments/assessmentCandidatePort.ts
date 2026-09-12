import type {
  AssessmentSubjectType,
  AssessmentSupportCandidate,
} from "../../models/assessmentWorkbookModel";
import type { WorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";

export type AssessmentCandidatePage = Readonly<{
  candidates: readonly AssessmentSupportCandidate[];
  hasMore: boolean;
  nextCursor: string | null;
}>;
export type AssessmentCandidateQuery = Readonly<{
  queryState: WorkbookQueryState;
  cursor: string | null;
  signal: AbortSignal;
}>;
export interface AssessmentCandidateReadPort {
  subjects(
    type: AssessmentSubjectType,
    input: AssessmentCandidateQuery,
  ): Promise<WorkbookPortResult<AssessmentCandidatePage>>;
  support(
    input: AssessmentCandidateQuery,
  ): Promise<WorkbookPortResult<AssessmentCandidatePage>>;
}
