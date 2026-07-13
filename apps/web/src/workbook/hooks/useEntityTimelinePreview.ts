import { useCallback, useRef, useState } from "react";
import { apiPath } from "../../services/browserApi";
import { fetchWorkbookJSON, readEnvelope } from "../../services/workbookApi";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  normalizeTimelineFullRow,
  rowFromApi,
  validateTimelineViewSchemaId,
  type WorkbookRow,
} from "../timeline/models/workbookTimelineModel";

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
  const previewSequenceRef = useRef(0);

  const clearTimelinePreview = useCallback(() => {
    previewSequenceRef.current += 1;
    setTimelinePreviewRows([]);
  }, []);

  const loadTimelinePreview = useCallback(
    async (recordId: string) => {
      const sequence = previewSequenceRef.current + 1;
      previewSequenceRef.current = sequence;
      setTimelinePreviewRows([]);
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
      if (!result.ok) {
        if (previewSequenceRef.current === sequence) {
          setTimelinePreviewRows([]);
        }
        return;
      }
      const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
      try {
        validateTimelineViewSchemaId(
          envelope.data.view_schema_id,
          "timeline preview query response",
        );
      } catch {
        if (previewSequenceRef.current === sequence) {
          setTimelinePreviewRows([]);
        }
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
        if (previewSequenceRef.current === sequence) {
          setTimelinePreviewRows([]);
        }
        return;
      }
      if (previewSequenceRef.current === sequence) {
        setTimelinePreviewRows(previewRows);
      }
    },
    [apiBase, entityType, incidentId],
  );

  return {
    clearTimelinePreview,
    loadTimelinePreview,
    timelinePreviewRows,
  };
}
