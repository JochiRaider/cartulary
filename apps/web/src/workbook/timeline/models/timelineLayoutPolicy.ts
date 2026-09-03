import { stringifyGridValue } from "../../utils/workbookValueFormat";
import { readTimelineCellValue, type WorkbookRow } from "./timelineRowModel";

export function timelineColumnWidth(fieldKey: string): number {
  switch (fieldKey) {
    case "timeline.date_entered_text":
    case "timeline.activity_utc_text":
    case "timeline.activity_local_text":
      return 180;
    case "timeline.edited_at":
      return 248;
    case "timeline.raw_activity_text":
      return 320;
    case "timeline.activity_synopsis_text":
      return 300;
    case "timeline.device_object_text":
      return 240;
    case "timeline.ip_address_text":
      return 160;
  }
  return 224;
}

const timelineColumnExpansionWeights: Readonly<Record<string, number>> = {
  "timeline.raw_activity_text": 3,
  "timeline.activity_synopsis_text": 2,
  "timeline.device_object_text": 1,
  "timeline.data_source_text": 1,
};

export function buildExpandedTimelineColumnWidths({
  actionsColumnWidth,
  fieldKeys,
  gridShellWidth,
  rowGutterWidth,
}: {
  readonly actionsColumnWidth: number;
  readonly fieldKeys: readonly string[];
  readonly gridShellWidth: number;
  readonly rowGutterWidth: number;
}): Record<string, number> {
  const baseWidths = Object.fromEntries(
    fieldKeys.map((fieldKey) => [fieldKey, timelineColumnWidth(fieldKey)]),
  );
  const availableDataWidth =
    Math.floor(gridShellWidth) - rowGutterWidth - actionsColumnWidth;
  const baseDataWidth = fieldKeys.reduce(
    (sum, fieldKey) => sum + (baseWidths[fieldKey] ?? 0),
    0,
  );
  const extraWidth = Math.max(0, availableDataWidth - baseDataWidth);
  if (extraWidth < 1) return baseWidths;

  const expandableFields = fieldKeys
    .map((fieldKey) => ({
      fieldKey,
      weight: timelineColumnExpansionWeights[fieldKey] ?? 0,
    }))
    .filter((entry) => entry.weight > 0);
  const totalWeight = expandableFields.reduce(
    (sum, entry) => sum + entry.weight,
    0,
  );
  if (totalWeight < 1) return baseWidths;

  const expandedWidths = { ...baseWidths };
  let assignedWidth = 0;
  expandableFields.forEach((entry, index) => {
    const addedWidth =
      index === expandableFields.length - 1
        ? extraWidth - assignedWidth
        : Math.floor((extraWidth * entry.weight) / totalWeight);
    assignedWidth += addedWidth;
    expandedWidths[entry.fieldKey] =
      (expandedWidths[entry.fieldKey] ?? timelineColumnWidth(entry.fieldKey)) +
      addedWidth;
  });
  return expandedWidths;
}

export function timelineGroupLabel(row: WorkbookRow, fieldKey: string): string {
  const value = stringifyGridValue(
    readTimelineCellValue(row.rawRow, fieldKey),
  ).trim();
  return value === "" ? "Unassigned" : value;
}
