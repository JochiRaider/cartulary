import { useCallback, useState } from "react";
import { apiPath } from "./browserApi";
import { fetchJSON, readEnvelope } from "./workbookApi";
import { timelineViewSchemaId } from "./workbookSurfaceRegistry";
import {
  normalizeTimelineFullRow,
  rowFromApi,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "./workbookTimelineModel";

type WorkbookQueryEnvelope = {
  data: {
    view_schema_id: string;
    rows: unknown[];
  };
};

export function useEntityTimelinePreview({
  apiBase,
  entityType,
  incidentId,
}: {
  readonly apiBase: string | undefined;
  readonly entityType: "host" | "identity";
  readonly incidentId: string;
}) {
  const [timelinePreviewRows, setTimelinePreviewRows] = useState<WorkbookRow[]>(
    [],
  );

  const loadTimelinePreview = useCallback(
    async (recordId: string) => {
      const result = await fetchJSON<WorkbookQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
        ),
        {
          method: "POST",
          body: JSON.stringify({}),
        },
      );
      if (!result.ok) {
        setTimelinePreviewRows([]);
        return;
      }
      const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
      try {
        validateTimelineViewSchemaId(
          envelope.data.view_schema_id,
          "timeline preview query response",
        );
      } catch {
        setTimelinePreviewRows([]);
        return;
      }
      const draftKey = entityType === "host" ? "hostRefs" : "identityRefs";
      let previewRows: WorkbookRow[];
      try {
        previewRows = envelope.data.rows
          .map((row, index) =>
            rowFromApi(
              normalizeTimelineFullRow(
                row,
                `timeline preview query rows[${index}]`,
              ),
            ),
          )
          .filter((row) =>
            row.collectionValues[draftKey].some(
              (item) => item.resolvedRecordId === recordId,
            ),
          );
      } catch {
        setTimelinePreviewRows([]);
        return;
      }
      setTimelinePreviewRows(previewRows);
    },
    [apiBase, entityType, incidentId],
  );

  return {
    loadTimelinePreview,
    timelinePreviewRows,
  };
}
