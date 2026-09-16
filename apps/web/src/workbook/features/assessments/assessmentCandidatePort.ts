import type {
  AssessmentSubjectType,
  AssessmentSupportCandidate,
} from "../../models/assessmentWorkbookModel";
import type {
  WorkbookCandidatePage,
  WorkbookCandidateQuery,
} from "../../ports/WorkbookCandidateReadPort";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";

export type AssessmentCandidatePage =
  WorkbookCandidatePage<AssessmentSupportCandidate>;
export type AssessmentCandidateQuery = WorkbookCandidateQuery;
export interface AssessmentCandidateReadPort {
  subjects(
    type: AssessmentSubjectType,
    input: AssessmentCandidateQuery,
  ): Promise<WorkbookPortResult<AssessmentCandidatePage>>;
  support(
    input: AssessmentCandidateQuery,
  ): Promise<WorkbookPortResult<AssessmentCandidatePage>>;
}
