import { useEffect, useState } from "react";
import { apiPath } from "../../services/browserApi";
import { fetchWorkbookJSON, readEnvelope } from "../../services/workbookApi";
import {
  type AssessmentSupportCandidate,
  assessmentSupportCandidate,
} from "../models/assessmentWorkbookModel";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  normalizeTimelineFullRow,
  validateTimelineViewSchemaId,
} from "../timeline/models/workbookTimelineModel";

type WorkbookQueryEnvelope = {
  data: {
    view_schema_id: string;
    rows: unknown[];
  };
};

export function useAssessmentSupportCandidates({
  apiBase,
  incidentId,
}: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}) {
  const [supportCandidates, setSupportCandidates] = useState<
    AssessmentSupportCandidate[]
  >([]);

  useEffect(() => {
    let isCurrent = true;
    async function loadSupportRows() {
      const result = await fetchWorkbookJSON<WorkbookQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
        ),
        {
          method: "POST",
          body: JSON.stringify({}),
        },
      );
      if (!isCurrent) {
        return;
      }
      if (!result.ok) {
        setSupportCandidates([]);
        return;
      }
      const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
      try {
        validateTimelineViewSchemaId(
          envelope.data.view_schema_id,
          "assessment support query response",
        );
        setSupportCandidates(
          envelope.data.rows.map((row, index) => {
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
      isCurrent = false;
    };
  }, [apiBase, incidentId]);

  return supportCandidates;
}
