import type {
  CollectionItem,
  RelationshipFieldKey,
} from "./workbookMentionChips";
import { timelineViewSchemaId } from "./workbookSurfaceRegistry";

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

type TimelineScalarBinding = {
  kind: "scalar";
  fieldKey: EditableField;
  key: keyof RowValues;
  multiline?: boolean;
};

type TimelineCollectionBinding = {
  kind: "collection";
  fieldKey: CollectionFieldKey;
  draftKey: CollectionDraftKey;
  collectionKind: "relationship" | "tag";
  entityType?: "host" | "identity";
};

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

export type WorkbookVersionedRecord = {
  readonly recordId: string | null;
  readonly rowVersion: number | null;
};

export type WorkbookRecordFreshnessDecision = {
  readonly comparable: boolean;
  readonly stale: boolean;
};

export function decideWorkbookRecordFreshness(
  incoming: WorkbookVersionedRecord,
  knownRowVersion: number | null | undefined,
): WorkbookRecordFreshnessDecision {
  if (
    incoming.recordId === null ||
    incoming.rowVersion === null ||
    knownRowVersion === null ||
    knownRowVersion === undefined
  ) {
    return {
      comparable: false,
      stale: false,
    };
  }
  return {
    comparable: true,
    stale: incoming.rowVersion < knownRowVersion,
  };
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

export function clipboardTextLooksTabular(text: string): boolean {
  return text.includes("\n") || text.includes("\r") || text.includes("\t");
}
