import {
  normalizeViewRowPatchV1,
  normalizeViewRowV1,
  requireViewContract,
} from "@cartulary/view-contracts";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { WorkbookSameFieldConflictPayload } from "../../runtime/workbookConflictModel";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import {
  type CollectionItem,
  type RelationshipFieldKey,
  readCollectionItems,
} from "./workbookMentionChips";
import type { WorkbookVersionedRecord } from "./workbookRecordFreshness";

type EditableField =
  | "timeline.date_entered_text"
  | "timeline.analyst_text"
  | "timeline.mitre_stage_text"
  | "timeline.device_object_text"
  | "timeline.ip_address_text"
  | "timeline.activity_utc_text"
  | "timeline.activity_local_text"
  | "timeline.raw_activity_text"
  | "timeline.activity_synopsis_text"
  | "timeline.data_source_text";

type RelationshipDraftKey = "hostRefs" | "identityRefs";
export type CollectionFieldKey = RelationshipFieldKey | "timeline.tags";
export type CollectionDraftKey = RelationshipDraftKey | "tags";
export type FocusFieldKey = keyof RowValues | CollectionDraftKey;

export type RowValues = {
  dateEnteredText: string;
  analystText: string;
  mitreStageText: string;
  deviceObjectText: string;
  ipAddressText: string;
  activityUTCText: string;
  activityLocalText: string;
  rawActivityText: string;
  activitySynopsisText: string;
  dataSourceText: string;
};

type CollectionDrafts = {
  hostRefs: string;
  identityRefs: string;
  tags: string;
};

export type TagCollectionItem = {
  itemRef: string;
  itemKind: "tag" | string;
  displayText: string;
  rawText: string;
};

export type TimelineApiRow = WorkbookQueryRow & {
  view_schema_id: string;
};

export type WorkbookRow = WorkbookVersionedRecord & {
  key: string;
  viewSchemaId: string | null;
  captureState: string;
  values: RowValues;
  committedValues: RowValues;
  collectionValues: {
    hostRefs: CollectionItem[];
    identityRefs: CollectionItem[];
    tags: TagCollectionItem[];
  };
  collectionDrafts: CollectionDrafts;
  pendingSignature: string | null;
  rawRow: TimelineApiRow | null;
};

export type SameFieldConflictPayload = WorkbookSameFieldConflictPayload;

export type LocalConflictState = {
  key: string;
  anchor: {
    record_id: string;
    field_key: string;
    base_row_version: number;
    current_row_version?: number;
  };
  conflict: SameFieldConflictPayload;
  focusKey: string;
  localValue: unknown;
  mergedDraft: string;
};

export type PasteConflictGroupState = {
  keys: string[];
};

const timelineContract = requireViewContract(timelineViewSchemaId);

export type TimelineScalarBinding = {
  kind: "scalar";
  fieldKey: EditableField;
  key: keyof RowValues;
  multiline?: boolean;
};

export type TimelineCollectionBinding = {
  kind: "collection";
  fieldKey: CollectionFieldKey;
  draftKey: CollectionDraftKey;
  collectionKind: "relationship" | "tag";
  entityType?: "host" | "identity";
};

type TimelineReadonlyBinding = {
  kind: "readonly";
  fieldKey: string;
};

type TimelineFieldBinding =
  | TimelineScalarBinding
  | TimelineCollectionBinding
  | TimelineReadonlyBinding;

export type TimelineScalarEditorSurface = "grid" | "inspector";

export const timelineScalarEditorSurfaces: readonly TimelineScalarEditorSurface[] =
  ["grid", "inspector"];

const timelineScalarBindingIndex: Record<EditableField, TimelineScalarBinding> =
  {
    "timeline.date_entered_text": {
      kind: "scalar",
      fieldKey: "timeline.date_entered_text",
      key: "dateEnteredText",
    },
    "timeline.analyst_text": {
      kind: "scalar",
      fieldKey: "timeline.analyst_text",
      key: "analystText",
    },
    "timeline.mitre_stage_text": {
      kind: "scalar",
      fieldKey: "timeline.mitre_stage_text",
      key: "mitreStageText",
    },
    "timeline.device_object_text": {
      kind: "scalar",
      fieldKey: "timeline.device_object_text",
      key: "deviceObjectText",
    },
    "timeline.ip_address_text": {
      kind: "scalar",
      fieldKey: "timeline.ip_address_text",
      key: "ipAddressText",
    },
    "timeline.activity_utc_text": {
      kind: "scalar",
      fieldKey: "timeline.activity_utc_text",
      key: "activityUTCText",
    },
    "timeline.activity_local_text": {
      kind: "scalar",
      fieldKey: "timeline.activity_local_text",
      key: "activityLocalText",
    },
    "timeline.raw_activity_text": {
      kind: "scalar",
      fieldKey: "timeline.raw_activity_text",
      key: "rawActivityText",
      multiline: true,
    },
    "timeline.activity_synopsis_text": {
      kind: "scalar",
      fieldKey: "timeline.activity_synopsis_text",
      key: "activitySynopsisText",
      multiline: true,
    },
    "timeline.data_source_text": {
      kind: "scalar",
      fieldKey: "timeline.data_source_text",
      key: "dataSourceText",
    },
  };

const timelineCollectionBindingIndex: Record<
  CollectionFieldKey,
  TimelineCollectionBinding
> = {
  "timeline.host_refs": {
    kind: "collection",
    fieldKey: "timeline.host_refs",
    draftKey: "hostRefs",
    collectionKind: "relationship",
    entityType: "host",
  },
  "timeline.identity_refs": {
    kind: "collection",
    fieldKey: "timeline.identity_refs",
    draftKey: "identityRefs",
    collectionKind: "relationship",
    entityType: "identity",
  },
  "timeline.tags": {
    kind: "collection",
    fieldKey: "timeline.tags",
    draftKey: "tags",
    collectionKind: "tag",
  },
};

export const timelineScalarBindings: readonly TimelineScalarBinding[] =
  Object.values(timelineScalarBindingIndex);

export const timelineObservationSourceFields = timelineScalarBindings.map(
  (binding) => ({
    fieldKey: binding.fieldKey,
    label:
      timelineContract.fieldMap[binding.fieldKey]?.label ?? binding.fieldKey,
  }),
);

export const timelineCollectionBindings: readonly TimelineCollectionBinding[] =
  Object.values(timelineCollectionBindingIndex);

const timelineInspectorEditableFields: readonly EditableField[] = [
  "timeline.raw_activity_text",
  "timeline.activity_synopsis_text",
];

export function timelineFieldBinding(fieldKey: string): TimelineFieldBinding {
  if (fieldKey in timelineScalarBindingIndex) {
    return timelineScalarBindingIndex[fieldKey as EditableField];
  }
  if (fieldKey in timelineCollectionBindingIndex) {
    return timelineCollectionBindingIndex[fieldKey as CollectionFieldKey];
  }
  return {
    kind: "readonly",
    fieldKey,
  };
}

export const timelineVisibleBindings: readonly TimelineFieldBinding[] =
  timelineContract.fields.map((field) => timelineFieldBinding(field.fieldKey));

export const timelineInspectorBindings: readonly TimelineScalarBinding[] =
  timelineInspectorEditableFields.map(
    (fieldKey) => timelineFieldBinding(fieldKey) as TimelineScalarBinding,
  );

export function timelineScalarBindingForField(
  fieldKey: string,
): TimelineScalarBinding | null {
  const binding = timelineFieldBinding(fieldKey);
  return binding.kind === "scalar" ? binding : null;
}

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
    default:
      return 224;
  }
}

const timelineColumnExpansionWeights: Record<string, number> = {
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
  const baseWidths: Record<string, number> = Object.fromEntries(
    fieldKeys.map((fieldKey) => [fieldKey, timelineColumnWidth(fieldKey)]),
  );
  const availableDataWidth =
    Math.floor(gridShellWidth) - rowGutterWidth - actionsColumnWidth;
  const baseDataWidth = fieldKeys.reduce(
    (sum, fieldKey) => sum + (baseWidths[fieldKey] ?? 0),
    0,
  );
  const extraWidth = Math.max(0, availableDataWidth - baseDataWidth);
  if (extraWidth < 1) {
    return baseWidths;
  }

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
  if (totalWeight < 1) {
    return baseWidths;
  }

  const expandedWidths = { ...baseWidths };
  let assignedWidth = 0;
  expandableFields.forEach((entry, index) => {
    const isLast = index === expandableFields.length - 1;
    const addedWidth = isLast
      ? extraWidth - assignedWidth
      : Math.floor((extraWidth * entry.weight) / totalWeight);
    assignedWidth += addedWidth;
    expandedWidths[entry.fieldKey] =
      (expandedWidths[entry.fieldKey] ?? timelineColumnWidth(entry.fieldKey)) +
      addedWidth;
  });
  return expandedWidths;
}

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
  return {
    hostRefs: "",
    identityRefs: "",
    tags: "",
  };
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
    collectionValues: {
      hostRefs: [],
      identityRefs: [],
      tags: [],
    },
    collectionDrafts: emptyCollectionDrafts(),
    pendingSignature: null,
    rawRow: null,
  };
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

export function readTimelineTagItems(row: TimelineApiRow): TagCollectionItem[] {
  const raw = row.cells["timeline.tags"]?.value;
  const value =
    raw &&
    typeof raw === "object" &&
    !Array.isArray(raw) &&
    "items" in raw &&
    Array.isArray(raw.items)
      ? raw.items
      : [];
  return value
    .map((item, index) => {
      if (!item || typeof item !== "object") {
        return null;
      }
      const object = item as Record<string, unknown>;
      const rawText =
        typeof object.raw_text === "string"
          ? object.raw_text
          : typeof object.display_text === "string"
            ? object.display_text
            : "";
      if (rawText === "") {
        return null;
      }
      return {
        itemRef:
          typeof object.item_ref === "string"
            ? object.item_ref
            : `tag-item-${index}:${rawText}`,
        itemKind:
          typeof object.item_kind === "string" ? object.item_kind : "tag",
        displayText:
          typeof object.display_text === "string"
            ? object.display_text
            : rawText,
        rawText,
      };
    })
    .filter((item): item is TagCollectionItem => item !== null);
}

export function timelineGroupLabel(row: WorkbookRow, fieldKey: string): string {
  const value = stringifyGridValue(
    readTimelineCellValue(row.rawRow, fieldKey),
  ).trim();
  return value === "" ? "Unassigned" : value;
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

export type TimelinePatchCells = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
  group_values?: Record<string, unknown>;
};

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
    cells: {
      ...row.cells,
      ...patch.cells,
    },
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

export function createDraftRowForKey(rowKey: string): WorkbookRow | null {
  if (!rowKey.startsWith("draft-")) {
    return null;
  }
  return { ...createDraftRow(0), key: rowKey };
}

function buildCollectionActions(
  fieldKey: CollectionFieldKey,
  rawInput: string,
) {
  const actions = rawInput
    .split(/\r?\n/u)
    .filter((segment) => segment.trim() !== "")
    .map((rawText) =>
      fieldKey === "timeline.tags"
        ? {
            op: "add_tag",
            tag_name: rawText,
          }
        : {
            op: "add_token",
            raw_text: rawText,
          },
    );
  if (actions.length < 1) {
    return null;
  }
  return {
    kind: "collection_actions_v1",
    actions,
  };
}

export function buildScalarPatchIntent(row: WorkbookRow, clientTxnId: string) {
  const changes = Object.values(timelineScalarBindingIndex)
    .map((field) => {
      const current = row.values[field.key];
      const committed = row.committedValues[field.key];
      if (current === committed) {
        return null;
      }
      return {
        field_key: field.fieldKey,
        value: current,
      };
    })
    .filter(
      (change): change is { field_key: EditableField; value: string } =>
        change !== null,
    )
    .sort((left, right) => left.field_key.localeCompare(right.field_key));

  if (changes.length < 1) {
    return null;
  }

  return {
    view_schema_id: timelineViewSchemaId,
    client_txn_id: clientTxnId,
    changes,
  };
}

export function buildCollectionPatchIntent(
  fieldKey: CollectionFieldKey,
  draftValue: string,
  clientTxnId: string,
) {
  const actionPayload = buildCollectionActions(fieldKey, draftValue);
  if (actionPayload === null) {
    return null;
  }

  return {
    view_schema_id: timelineViewSchemaId,
    client_txn_id: clientTxnId,
    changes: [
      {
        field_key: fieldKey,
        action_payload: actionPayload,
      },
    ],
  };
}

export function inputFocusKey(
  rowKey: string,
  field: FocusFieldKey,
  surface: TimelineScalarEditorSurface = "grid",
) {
  return `${rowKey}:${field}:${surface}`;
}

export function timelineRelationshipLabel(fieldKey: RelationshipFieldKey) {
  return fieldKey === "timeline.identity_refs" ? "Identities" : "Hosts";
}

type BuildCreatePayloadOptions = {
  allowZeroFieldCreate?: boolean;
};

export function buildCreatePayload(
  row: WorkbookRow,
  clientTxnId: string,
  options: BuildCreatePayloadOptions = {},
) {
  const payload: Record<string, unknown> = {
    client_txn_id: clientTxnId,
  };

  for (const field of Object.values(timelineScalarBindingIndex)) {
    const normalized = row.values[field.key];
    if (normalized !== "") {
      payload[field.fieldKey] = normalized;
    }
  }

  for (const field of Object.values(timelineCollectionBindingIndex)) {
    const actions = buildCollectionActions(
      field.fieldKey,
      row.collectionDrafts[field.draftKey],
    );
    if (actions !== null) {
      payload[field.fieldKey] = actions;
    }
  }

  if (Object.keys(payload).length < 2 && !options.allowZeroFieldCreate) {
    return null;
  }
  return payload;
}
