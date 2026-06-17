import {
  normalizeViewRowPatchV1,
  normalizeViewRowV1,
  requireViewContract,
  visibleFields,
} from "@cartulary/view-contracts";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { PendingReplayUnitState } from "../../utils/workbookPendingQueue";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import {
  type CollectionItem,
  type RelationshipFieldKey,
  readCollectionItems,
} from "./workbookMentionChips";

export { clipboardTextLooksTabular } from "../../utils/workbookClipboard";
export type {
  WorkbookRecordFreshnessDecision,
  WorkbookVersionedRecord,
} from "./timelineRowsModel";
export { decideWorkbookRecordFreshness } from "./timelineRowsModel";

export type EditableField =
  | "timeline.occurred_at"
  | "timeline.summary"
  | "timeline.details"
  | "timeline.source_text";

export type RelationshipDraftKey = "hostRefs" | "identityRefs";
export type CollectionFieldKey = RelationshipFieldKey | "timeline.tags";
export type CollectionDraftKey = RelationshipDraftKey | "tags";
export type FocusFieldKey = keyof RowValues | CollectionDraftKey;
export type AssessmentSubjectType = "host" | "identity";
export type AssessmentConfidenceBand = "unset" | "low" | "medium" | "high";

export type RowValues = {
  occurredAt: string;
  summary: string;
  details: string;
  sourceText: string;
};

export type CollectionDrafts = {
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

export type ViewApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
  group_values?: Record<string, unknown>;
};

export type TimelineApiRow = ViewApiRow & {
  view_schema_id: string;
};

export type EntityApiRow = ViewApiRow;

export type WorkbookRow = {
  key: string;
  recordId: string | null;
  rowVersion: number | null;
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

export type AssessmentCreateDraft = {
  assessedAt: string;
  assessmentState: string;
  confidenceBand: AssessmentConfidenceBand;
  rationale: string;
  subjectRecordId: string;
  subjectType: AssessmentSubjectType;
  supportRecordIds: string[];
};

export type SameFieldConflictPayload = {
  conflict_token: string;
  record_id: string;
  field_key: string;
  conflict_resolution_class: string;
  base_row_version: number;
  current_row_version: number;
  client_value: unknown;
  server_value: unknown;
  base_value?: unknown;
  server_updated_by?: string;
  server_updated_at?: string;
  suggested_merged_value?: unknown;
};

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

export type TimelineConflictResolution =
  | "keep_saved"
  | "use_unsaved"
  | "merged_value";

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

export type TimelineReadonlyBinding = {
  kind: "readonly";
  fieldKey: string;
};

export type TimelineFieldBinding =
  | TimelineScalarBinding
  | TimelineCollectionBinding
  | TimelineReadonlyBinding;

export type TimelineScalarEditorSurface = "grid" | "inspector";

export const timelineScalarEditorSurfaces: readonly TimelineScalarEditorSurface[] =
  ["grid", "inspector"];

const timelineScalarBindingIndex: Record<EditableField, TimelineScalarBinding> =
  {
    "timeline.occurred_at": {
      kind: "scalar",
      fieldKey: "timeline.occurred_at",
      key: "occurredAt",
    },
    "timeline.summary": {
      kind: "scalar",
      fieldKey: "timeline.summary",
      key: "summary",
    },
    "timeline.details": {
      kind: "scalar",
      fieldKey: "timeline.details",
      key: "details",
      multiline: true,
    },
    "timeline.source_text": {
      kind: "scalar",
      fieldKey: "timeline.source_text",
      key: "sourceText",
      multiline: true,
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

const timelineInspectorEditableFields: readonly EditableField[] = [
  "timeline.details",
  "timeline.source_text",
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
  visibleFields(timelineContract).map((field) =>
    timelineFieldBinding(field.fieldKey),
  );

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

export function timelineFocusFieldForFieldKey(
  fieldKey: string,
): FocusFieldKey | null {
  const binding = timelineFieldBinding(fieldKey);
  if (binding.kind === "scalar") {
    return binding.key;
  }
  if (binding.kind === "collection") {
    return binding.draftKey;
  }
  return null;
}

export function timelineColumnWidth(fieldKey: string): number {
  switch (fieldKey) {
    case "timeline.occurred_at":
      return 180;
    case "timeline.edited_at":
      return 248;
    case "timeline.summary":
      return 320;
    case "timeline.host_refs":
      return 300;
    case "timeline.identity_refs":
      return 320;
    case "timeline.evidence_count":
      return 112;
    case "timeline.tags":
      return 240;
    default:
      return 224;
  }
}

const timelineColumnExpansionWeights: Record<string, number> = {
  "timeline.summary": 3,
  "timeline.host_refs": 2,
  "timeline.identity_refs": 2,
  "timeline.tags": 1,
  "timeline.edited_at": 2,
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
    occurredAt: "",
    summary: "",
    details: "",
    sourceText: "",
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

export function ensureDraftRow(
  rows: WorkbookRow[],
  nextDraftIndex: number,
): WorkbookRow[] {
  if (rows.some((row) => row.recordId === null)) {
    return rows;
  }
  return [...rows, createDraftRow(nextDraftIndex)];
}

function normalizeValue(value: string): string {
  return value.trim();
}

export function readTimelineStringCell(
  row: TimelineApiRow | EntityApiRow,
  fieldKey: string,
): string {
  const raw = row.cells[fieldKey]?.value;
  return typeof raw === "string" ? raw : "";
}

export function readTimelineCellValue(
  row: TimelineApiRow | EntityApiRow | null,
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
    occurredAt: readTimelineStringCell(row, "timeline.occurred_at"),
    summary: readTimelineStringCell(row, "timeline.summary"),
    details: readTimelineStringCell(row, "timeline.details"),
    sourceText: readTimelineStringCell(row, "timeline.source_text"),
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
      const current = normalizeValue(row.values[field.key]);
      const committed = normalizeValue(row.committedValues[field.key]);
      if (current === committed) {
        return null;
      }
      return {
        field_key: field.fieldKey,
        value: current === "" ? null : current,
      };
    })
    .filter(
      (change): change is { field_key: EditableField; value: string | null } =>
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

export function buildAttachedEvidenceCreatePayload(
  evidenceRecordId: string,
  clientTxnId: string,
) {
  return {
    client_txn_id: clientTxnId,
    "timeline.attached_evidence_ids": {
      kind: "collection_actions_v1",
      actions: [
        {
          op: "add_record_ref",
          linked_record_id: evidenceRecordId,
        },
      ],
    },
  };
}

export function buildAttachedEvidencePatchPayload(
  row: WorkbookRow,
  evidenceRecordId: string,
  clientTxnId: string,
) {
  if (row.rowVersion === null) {
    return null;
  }
  return {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.rowVersion,
    client_txn_id: clientTxnId,
    changes: [
      {
        field_key: "timeline.attached_evidence_ids",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "add_record_ref",
              linked_record_id: evidenceRecordId,
            },
          ],
        },
      },
    ],
  };
}

export function materializePendingReplayPayload(
  unit: PendingReplayUnitState,
  currentRow: WorkbookRow | undefined,
) {
  if (unit.kind === "create") {
    return unit.payloadIntent;
  }
  if (currentRow?.rowVersion === null || currentRow?.rowVersion === undefined) {
    return null;
  }
  return {
    ...unit.payloadIntent,
    base_row_version: currentRow.rowVersion,
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
    const normalized = normalizeValue(row.values[field.key]);
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

export function confidenceScoreFromBand(
  band: AssessmentConfidenceBand,
): number | null {
  switch (band) {
    case "low":
      return 25;
    case "medium":
      return 55;
    case "high":
      return 85;
    default:
      return null;
  }
}

export function buildAssessmentCreatePayload(
  draft: AssessmentCreateDraft,
  clientTxnId: string,
): Record<string, unknown> | null {
  const subjectRecordId = normalizeValue(draft.subjectRecordId);
  const assessmentState = normalizeValue(draft.assessmentState);
  const rationale = normalizeValue(draft.rationale);
  if (subjectRecordId === "" || assessmentState === "" || rationale === "") {
    return null;
  }

  const payload: Record<string, unknown> = {
    client_txn_id: clientTxnId,
    "assessment.subject_ref": subjectRecordId,
    "assessment.subject_type": draft.subjectType,
    "assessment.assessment_state": assessmentState,
    "assessment.confidence_score": confidenceScoreFromBand(
      draft.confidenceBand,
    ),
    "assessment.rationale": rationale,
  };

  const assessedAt = normalizeValue(draft.assessedAt);
  if (assessedAt !== "") {
    payload["assessment.assessed_at"] = assessedAt;
  }

  const supportRecordIds = Array.from(
    new Set(
      draft.supportRecordIds
        .map((recordId) => normalizeValue(recordId))
        .filter((recordId) => recordId !== ""),
    ),
  );
  if (supportRecordIds.length > 0) {
    payload["assessment.support_refs"] = {
      kind: "collection_actions_v1",
      actions: supportRecordIds.map((recordId) => ({
        op: "add_record_ref",
        linked_record_id: recordId,
      })),
    };
  }

  return payload;
}
