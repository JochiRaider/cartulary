import { useEffect, useState } from "react";
import { apiPath } from "../../services/browserApi";
import { fetchWorkbookJSON, readEnvelope } from "../../services/workbookApi";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  normalizeTimelineFullRow,
  type TimelineApiRow,
  validateTimelineViewSchemaId,
} from "../timeline/models/workbookTimelineModel";

type WorkbookQueryEnvelope = {
  data: {
    view_schema_id: string;
    rows: unknown[];
  };
};

export function useAssessmentSupportRows({
  apiBase,
  incidentId,
}: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}) {
  const [supportRows, setSupportRows] = useState<TimelineApiRow[]>([]);

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
        setSupportRows([]);
        return;
      }
      const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
      try {
        validateTimelineViewSchemaId(
          envelope.data.view_schema_id,
          "assessment support query response",
        );
        setSupportRows(
          envelope.data.rows.map((row, index) =>
            normalizeTimelineFullRow(
              row,
              `assessment support query rows[${index}]`,
            ),
          ),
        );
      } catch {
        setSupportRows([]);
      }
    }
    void loadSupportRows();
    return () => {
      isCurrent = false;
    };
  }, [apiBase, incidentId]);

  return supportRows;
}
