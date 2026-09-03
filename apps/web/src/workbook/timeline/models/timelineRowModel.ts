import {
  normalizeViewRowPatchV1,
  normalizeViewRowV1,
  requireViewContract,
} from "@cartulary/view-contracts";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { RowValues } from "./timelineFieldRegistry";
import {
  type CollectionItem,
  readCollectionItems,
} from "./workbookMentionChips";
import type { WorkbookVersionedRecord } from "./workbookRecordFreshness";

export type CollectionDrafts = {
  readonly hostRefs: string;
  readonly identityRefs: string;
  readonly tags: string;
};

export type TagCollectionItem = {
  readonly itemRef: string;
  readonly itemKind: string;
  readonly displayText: string;
  readonly rawText: string;
};

export type TimelineApiRow = WorkbookQueryRow & {
  readonly view_schema_id: string;
};

export type WorkbookRow = WorkbookVersionedRecord & {
  readonly key: string;
  readonly viewSchemaId: string | null;
  readonly captureState: string;
  readonly values: RowValues;
  readonly committedValues: RowValues;
  readonly collectionValues: {
    readonly hostRefs: CollectionItem[];
    readonly identityRefs: CollectionItem[];
    readonly tags: TagCollectionItem[];
  };
  readonly collectionDrafts: CollectionDrafts;
  readonly pendingSignature: string | null;
  readonly rawRow: TimelineApiRow | null;
};

export type TimelinePatchCells = {
  readonly record_id: string;
  readonly row_version: number;
  readonly cells: Record<string, { value: unknown }>;
  readonly group_values?: Record<string, unknown>;
};

const timelineContract = requireViewContract(timelineViewSchemaId);

function emptyValues(): RowValues {
  return {
    dateEnteredText: "",
    analystText: "",
    mitreStageText: "",
    deviceObjectText: "",
    ipAddressText: "",
    activityUTCText: "",
    activityLocalText: "",
    rawActivityText: "",
    activitySynopsisText: "",
    dataSourceText: "",
  };
}

function emptyCollectionDrafts(): CollectionDrafts {
  return { hostRefs: "", identityRefs: "", tags: "" };
}

export function createDraftRow(index: number): WorkbookRow {
  return {
    key: `draft-${index}`,
    recordId: null,
    rowVersion: null,
    viewSchemaId: timelineViewSchemaId,
    captureState: "rough",
    values: emptyValues(),
    committedValues: emptyValues(),
    collectionValues: { hostRefs: [], identityRefs: [], tags: [] },
    collectionDrafts: emptyCollectionDrafts(),
    pendingSignature: null,
    rawRow: null,
  };
}

export function createDraftRowForKey(rowKey: string): WorkbookRow | null {
  return rowKey.startsWith("draft-")
    ? { ...createDraftRow(0), key: rowKey }
    : null;
}

function readTimelineStringCell(
  row: TimelineApiRow | WorkbookQueryRow,
  fieldKey: string,
): string {
  const raw = row.cells[fieldKey]?.value;
  return typeof raw === "string" ? raw : "";
}

export function readTimelineCellValue(
  row: TimelineApiRow | WorkbookQueryRow | null,
  fieldKey: string,
): unknown {
  return row?.cells[fieldKey]?.value ?? null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

export function readTimelineTagItems(row: TimelineApiRow): TagCollectionItem[] {
  const raw = row.cells["timeline.tags"]?.value;
  const value = isRecord(raw) && Array.isArray(raw.items) ? raw.items : [];
  return value.flatMap((item, index) => {
    if (!isRecord(item)) return [];
    const rawText =
      typeof item.raw_text === "string"
        ? item.raw_text
        : typeof item.display_text === "string"
          ? item.display_text
          : "";
    if (rawText === "") return [];
    return [
      {
        itemRef:
          typeof item.item_ref === "string"
            ? item.item_ref
            : `tag-item-${index}:${rawText}`,
        itemKind: typeof item.item_kind === "string" ? item.item_kind : "tag",
        displayText:
          typeof item.display_text === "string" ? item.display_text : rawText,
        rawText,
      },
    ];
  });
}

export function validateTimelineViewSchemaId(value: unknown, source: string) {
  if (value !== timelineViewSchemaId) {
    throw new Error(
      `Timeline view row envelope failed: ${source} view_schema_id must be ${timelineViewSchemaId}.`,
    );
  }
}

function materializeTimelineCells(
  cells: Readonly<Record<string, { readonly value: unknown }>>,
): Record<string, { value: unknown }> {
  return Object.fromEntries(
    Object.entries(cells).map(([fieldKey, cell]) => [
      fieldKey,
      { value: cell.value },
    ]),
  );
}

export function normalizeTimelineFullRow(
  row: unknown,
  source: string,
): TimelineApiRow {
  const normalized = normalizeViewRowV1(timelineContract, row, source);
  validateTimelineViewSchemaId(normalized.viewSchemaId, source);
  return {
    view_schema_id: normalized.viewSchemaId,
    record_id: normalized.recordId,
    row_version: normalized.rowVersion,
    cells: materializeTimelineCells(normalized.cells),
    ...(normalized.groupValues === undefined
      ? {}
      : { group_values: { ...normalized.groupValues } }),
  };
}

export function normalizeTimelinePatchCells(
  patch: unknown,
  source: string,
): TimelinePatchCells {
  const normalized = normalizeViewRowPatchV1(timelineContract, patch, source);
  return {
    record_id: normalized.recordId,
    row_version: normalized.rowVersion,
    cells: materializeTimelineCells(normalized.cells),
    ...(normalized.groupValues === undefined
      ? {}
      : { group_values: { ...normalized.groupValues } }),
  };
}

export function rowFromApi(row: TimelineApiRow): WorkbookRow {
  const values: RowValues = {
    dateEnteredText: readTimelineStringCell(row, "timeline.date_entered_text"),
    analystText: readTimelineStringCell(row, "timeline.analyst_text"),
    mitreStageText: readTimelineStringCell(row, "timeline.mitre_stage_text"),
    deviceObjectText: readTimelineStringCell(
      row,
      "timeline.device_object_text",
    ),
    ipAddressText: readTimelineStringCell(row, "timeline.ip_address_text"),
    activityUTCText: readTimelineStringCell(row, "timeline.activity_utc_text"),
    activityLocalText: readTimelineStringCell(
      row,
      "timeline.activity_local_text",
    ),
    rawActivityText: readTimelineStringCell(row, "timeline.raw_activity_text"),
    activitySynopsisText: readTimelineStringCell(
      row,
      "timeline.activity_synopsis_text",
    ),
    dataSourceText: readTimelineStringCell(row, "timeline.data_source_text"),
  };
  return {
    key: row.record_id,
    recordId: row.record_id,
    rowVersion: row.row_version,
    viewSchemaId: row.view_schema_id,
    captureState: readTimelineStringCell(row, "timeline.capture_state"),
    values,
    committedValues: values,
    collectionValues: {
      hostRefs: readCollectionItems(row, "timeline.host_refs"),
      identityRefs: readCollectionItems(row, "timeline.identity_refs"),
      tags: readTimelineTagItems(row),
    },
    collectionDrafts: emptyCollectionDrafts(),
    pendingSignature: null,
    rawRow: row,
  };
}

export function applyViewRowPatch(
  row: TimelineApiRow,
  patch: TimelinePatchCells,
): TimelineApiRow {
  return {
    ...row,
    row_version: patch.row_version,
    cells: { ...row.cells, ...patch.cells },
    ...(patch.group_values
      ? {
          group_values: {
            ...(row.group_values ?? {}),
            ...patch.group_values,
          },
        }
      : {}),
  };
}
