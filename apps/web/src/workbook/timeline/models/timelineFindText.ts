import { buildEvidenceCountDisplayViewModel } from "../../models/evidenceLifecycleViewModel";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import { projectTimelineCollectionPresentation } from "./timelineCollectionPresentation";
import {
  timelineFieldBinding,
  timelineVisibleBindings,
} from "./timelineFieldRegistry";
import {
  readTimelineCellValue,
  rowFromApi,
  type WorkbookRow,
} from "./timelineRowModel";

/** Readable committed value fragments. No draft, DOM, clipboard escaping or transport fallback. */
export function timelineFindText(
  row: WorkbookRow,
  fieldKey: string,
  entityIndex: Readonly<Record<string, { readonly label: string }>>,
): readonly string[] {
  if (row.recordId === null || row.rawRow === null) return [];
  if (!timelineVisibleBindings.some((binding) => binding.fieldKey === fieldKey))
    return [];
  const binding = timelineFieldBinding(fieldKey);
  if (binding.kind === "collection") {
    const presentation = projectTimelineCollectionPresentation({
      binding,
      entityIndex,
      row: rowFromApi(row.rawRow),
    });
    return presentation.kind === "relationship"
      ? presentation.visibleItems.map((item) => item.chip.label)
      : presentation.visibleItems.map((item) => item.displayText);
  }
  if (fieldKey === "timeline.evidence_count") {
    const presentation = buildEvidenceCountDisplayViewModel({
      projectedCount: readTimelineCellValue(row.rawRow, fieldKey),
      projectedHasEvidence: readTimelineCellValue(
        row.rawRow,
        "timeline.has_evidence",
      ),
    });
    return presentation.stateKey === "inconsistent"
      ? []
      : [presentation.displayCount, String(presentation.hasEvidence)];
  }
  const value = readTimelineCellValue(row.rawRow, fieldKey);
  if (
    typeof value !== "string" &&
    typeof value !== "number" &&
    typeof value !== "boolean"
  )
    return [];
  const text = stringifyGridValue(value);
  return text === "" ? [] : [text];
}
