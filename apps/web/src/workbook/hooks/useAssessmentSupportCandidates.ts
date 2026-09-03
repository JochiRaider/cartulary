import { requireViewContract } from "@cartulary/view-contracts";
import { useEffect, useState } from "react";
import {
  type AssessmentSupportCandidate,
  assessmentSupportCandidate,
} from "../models/assessmentWorkbookModel";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import { normalizeTimelineFullRow } from "../timeline/models/timelineRowModel";

const timelineContract = requireViewContract(timelineViewSchemaId);

export function useAssessmentSupportCandidates({
  viewQuery,
}: {
  readonly viewQuery: WorkbookViewQueryPort;
}) {
  const [supportCandidates, setSupportCandidates] = useState<
    AssessmentSupportCandidate[]
  >([]);

  useEffect(() => {
    const controller = new AbortController();
    async function loadSupportRows() {
      const result = await viewQuery.query({
        contract: timelineContract,
        queryState: emptyWorkbookQueryState(),
        signal: controller.signal,
      });
      if (controller.signal.aborted || result.kind === "aborted") {
        return;
      }
      if (result.kind === "rejected") {
        setSupportCandidates([]);
        return;
      }
      try {
        setSupportCandidates(
          result.value.rows.map((row, index) => {
            const timelineRow = normalizeTimelineFullRow(
              row,
              `assessment support query rows[${index}]`,
            );
            return assessmentSupportCandidate(
              timelineRow.record_id,
              timelineRow.cells["timeline.activity_synopsis_text"]?.value,
            );
          }),
        );
      } catch {
        setSupportCandidates([]);
      }
    }
    void loadSupportRows();
    return () => {
      controller.abort();
    };
  }, [viewQuery]);

  return supportCandidates;
}
