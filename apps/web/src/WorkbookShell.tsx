import {
  type GridActionsColumn,
  type GridColumn,
  type GridRow,
  GridTable,
  GridViewport,
  reconcileRecordRows,
} from "@cartulary/grid-adapter";
import {
  draftCellTestId,
  gridGroupRowTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  type WorkbookSurface,
} from "@cartulary/test-utils";
import {
  listViewContracts,
  requireViewContract,
  resolveHeaderSortFieldKey,
  type ViewContract,
  type ViewFieldContract,
  visibleFields,
} from "@cartulary/view-contracts";
import {
  type ChangeEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  startTransition,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { WorkbookGridControls } from "./WorkbookGridControls";
import {
  captureViewportAnchor,
  computeRestoredViewportScroll,
  isRectFullyVisibleWithinContainer,
  type ScrollPosition,
  type ViewportSnapshot,
} from "./workbookContinuity";
import {
  applyFilterDraft,
  buildQueryRequest,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  removeFilterField,
  toggleSortField,
  updateGroupBy,
  type WorkbookQueryState,
} from "./workbookQuery";
import {
  buildAutoResolutionNotices,
  buildInspectorMentions,
  buildMentionPatchPayload,
  isRecordChangedMessage,
  readCollectionItems,
  shouldIgnoreSelfOriginatedRecordChange,
} from "./workbookShellPhase4";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const hostsViewSchemaId = "cartulary.view.hosts.v1";
const identitiesViewSchemaId = "cartulary.view.identities.v1";
const evidenceViewSchemaId = "cartulary.view.evidence.v1";
const notesViewSchemaId = "cartulary.view.notes.v1";
const assessmentsViewSchemaId = "cartulary.view.assessments.v1";
const timelineContract = requireViewContract(timelineViewSchemaId);
const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const assessmentsContract = requireViewContract(assessmentsViewSchemaId);
const allWorkbookContracts = listViewContracts();
const builtInSurfaceIDs = [
  timelineViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  evidenceViewSchemaId,
  notesViewSchemaId,
] as const;
const legacySurfaceParamToViewSchemaID: Readonly<Record<string, string>> = {
  hosts: hostsViewSchemaId,
  identities: identitiesViewSchemaId,
  timeline: timelineViewSchemaId,
};
const csrfCookieName = "cartulary_csrf";
const csrfHeaderName = "X-CSRF-Token";

type SaveState = "Syncing" | "Saved" | "Conflict";
type SurfaceKind = string;
type EditableField =
  | "timeline.occurred_at"
  | "timeline.summary"
  | "timeline.details"
  | "timeline.source_text";
type RelationshipFieldKey = "timeline.host_refs" | "timeline.identity_refs";
type RelationshipDraftKey = "hostRefs" | "identityRefs";
type CollectionFieldKey = RelationshipFieldKey | "timeline.tags";
type CollectionDraftKey = RelationshipDraftKey | "tags";
type FocusFieldKey = keyof RowValues | CollectionDraftKey;
type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";
type AssessmentSubjectType = "host" | "identity";
type AssessmentConfidenceBand = "unset" | "low" | "medium" | "high";

type RowValues = {
  occurredAt: string;
  summary: string;
  details: string;
  sourceText: string;
};

type CollectionDrafts = {
  hostRefs: string;
  identityRefs: string;
  tags: string;
};

type CollectionItem = {
  itemRef: string;
  entityType: "host" | "identity";
  itemKind: "resolved_ref" | "unresolved_mention" | string;
  displayText: string;
  rawText: string;
  resolvedRecordId: string | null;
  resolutionMethod: string | null;
  autoResolved: boolean;
  provenance: string | null;
  confidence: number | null;
  matchedAliasText: string | null;
};

type TagCollectionItem = {
  itemRef: string;
  itemKind: "tag" | string;
  displayText: string;
  rawText: string;
};

type WorkbookRow = {
  key: string;
  recordId: string | null;
  rowVersion: number | null;
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

type TimelineWorkbookProps = {
  incidentId: string;
  apiBase?: string | undefined;
  hostEntities?: EntityRow[];
  identityEntities?: EntityRow[];
  entityIndex?: Record<string, EntityRow>;
  currentIncidentRole?: IncidentRole | null;
  onRefreshEntities?: () => Promise<void> | void;
};

type WorkbookShellProps = {
  incidentId: string;
  apiBase?: string | undefined;
  onIncidentSnapshot?:
    | ((incident: {
        incident_id: string;
        incident_key: string;
        title: string;
        description: string | null;
        severity: string | null;
        tlp: string | null;
        current_phase: string | null;
        primary_external_case_ref: string | null;
        incident_version: number;
      }) => void)
    | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
};

type WorkbookQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: TimelineApiRow[];
  };
};

type TimelineMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    row: TimelineApiRow;
  };
};

type TimelineActionEnvelope = {
  data: {
    record_id: string;
    incident_id: string;
    row_version: number;
    capture_state: string;
    change_set_id: string;
    reason: string | null;
    replacement_record_id: string | null;
  };
};

type SessionEnvelope = {
  data: {
    memberships: Array<{
      incident_id: string;
      role: IncidentRole;
    }>;
  };
};

type ViewQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: EntityApiRow[];
  };
};

type ViewMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id: string;
    row: EntityApiRow;
  };
};

type MergeEnvelope = {
  data: {
    incident_id: string;
    survivor_record_id: string;
    loser_record_id: string;
    survivor_row_version: number;
    loser_row_version: number;
    change_set_id: string;
    merged_into_record_id: string;
    merge_summary: {
      record_type: string;
      repointed_mention_resolution_count: number;
      repointed_link_count: number;
      deduped_link_count: number;
      repointed_tag_count: number;
      deduped_tag_count: number;
      repointed_assessment_count: number;
      exact_match_classes: Array<{
        identifier_class: string;
        promoted_count: number;
        carried_count: number;
        duplicate_noop_count: number;
        blocked_conflict_count: number;
        provenance_only_count: number;
        suggestion_only_count: number;
      }>;
    };
  };
};

type EvidenceHandleEnvelope = {
  data: {
    href: string;
    method: "GET";
    filename: string;
    preview_kind?: string | null;
    content_type?: string | null;
  };
};

type EvidencePreviewState = {
  href: string;
  recordId: string;
  title: string;
  previewKind: string | null;
};

type TimelineApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
};

type EntityApiRow = TimelineApiRow;

type EntityRow = {
  entityType: "host" | "identity";
  recordId: string;
  rowVersion: number;
  label: string;
  secondaryText: string;
  state: string;
  aliasTexts: string[];
  linkedEventCount: number;
  rawRow: EntityApiRow;
  identifiers: Array<{
    key: string;
    label: string;
    value: string;
  }>;
};

type AssessmentCreateDraft = {
  assessedAt: string;
  assessmentState: string;
  confidenceBand: AssessmentConfidenceBand;
  rationale: string;
  subjectRecordId: string;
  subjectType: AssessmentSubjectType;
  supportRecordIds: string[];
};

type ViewportContinuityFollowup = "none" | "entity-refresh";

type ViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

type ViewportContinuityRequest = {
  token: number;
  attemptVersion: number;
  target: ViewportContinuityTarget;
  preservedViewport: ViewportSnapshot | null;
  followup: ViewportContinuityFollowup;
  followupSettled: boolean;
  baselineHostEntities: EntityRow[];
  baselineIdentityEntities: EntityRow[];
};

type DismissedMention = {
  rowRecordId: string;
  fieldKey: RelationshipFieldKey;
  entityType: "host" | "identity";
  itemRef: string;
  rawText: string;
  resolvedRecordId: string | null;
  resolutionMethod: string | null;
  autoResolved: boolean;
};

type InspectorMention = DismissedMention & {
  status: "unresolved" | "resolved" | "dismissed";
  displayText: string;
  provenance: string | null;
  confidence: number | null;
  matchedAliasText: string | null;
};

type AutoResolutionNotice = {
  itemRef: string;
  rowRecordId: string;
  fieldKey: RelationshipFieldKey;
  entityType: "host" | "identity";
  rawText: string;
  resolvedRecordId: string;
  matchedAliasText: string | null;
};

type LoadRowsOptions = {
  showLoading: boolean;
  viewportContinuityToken?: number;
};

type MergePlanLine = {
  label: string;
  outcome: string;
};

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

type TimelineReadonlyBinding = {
  kind: "readonly";
  fieldKey: string;
};

type TimelineFieldBinding =
  | TimelineScalarBinding
  | TimelineCollectionBinding
  | TimelineReadonlyBinding;
type TimelineScalarEditorSurface = "grid" | "inspector";

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

const timelineInspectorEditableFields: readonly EditableField[] = [
  "timeline.details",
  "timeline.source_text",
];

function timelineFieldBinding(fieldKey: string): TimelineFieldBinding {
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

const timelineVisibleBindings: readonly TimelineFieldBinding[] = visibleFields(
  timelineContract,
).map((field) => timelineFieldBinding(field.fieldKey));
const timelineInspectorBindings: readonly TimelineScalarBinding[] =
  timelineInspectorEditableFields.map(
    (fieldKey) => timelineFieldBinding(fieldKey) as TimelineScalarBinding,
  );

function timelineColumnWidth(fieldKey: string): number {
  switch (fieldKey) {
    case "timeline.occurred_at":
    case "timeline.edited_at":
      return 180;
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

const mergeIdentifierFields: Record<
  EntityRow["entityType"],
  Array<{ key: string; label: string }>
> = {
  host: [
    { key: "host.aad_device_id", label: "AAD Device ID" },
    { key: "host.fqdn", label: "FQDN" },
    { key: "host.hostname", label: "Hostname" },
  ],
  identity: [
    { key: "identity.aad_object_id", label: "AAD Object ID" },
    { key: "identity.sid", label: "SID" },
    { key: "identity.upn", label: "UPN" },
    { key: "identity.email", label: "Email" },
    { key: "identity.sam_account_name", label: "SAM Account Name" },
  ],
};

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

function normalizeValue(value: string): string {
  return value.trim();
}

function readStringCell(
  row: TimelineApiRow | EntityApiRow,
  fieldKey: string,
): string {
  const raw = row.cells[fieldKey]?.value;
  return typeof raw === "string" ? raw : "";
}

function readNumberCell(row: EntityApiRow, fieldKey: string): number {
  const raw = row.cells[fieldKey]?.value;
  return typeof raw === "number" ? raw : 0;
}

function readCellValue(
  row: TimelineApiRow | EntityApiRow | null,
  fieldKey: string,
): unknown {
  return row?.cells[fieldKey]?.value ?? null;
}

function stringifyGridValue(value: unknown): string {
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "boolean") {
    return value ? "true" : "false";
  }
  if (typeof value === "number") {
    return String(value);
  }
  if (
    value &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    "items" in value &&
    Array.isArray(value.items)
  ) {
    return value.items
      .map((item) => {
        if (!item || typeof item !== "object") {
          return null;
        }
        const object = item as Record<string, unknown>;
        return typeof object.display_text === "string"
          ? object.display_text
          : typeof object.raw_text === "string"
            ? object.raw_text
            : null;
      })
      .filter((item): item is string => item !== null)
      .join(", ");
  }
  return "";
}

function readTagItems(row: TimelineApiRow): TagCollectionItem[] {
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

function timelineGroupLabel(row: WorkbookRow, fieldKey: string) {
  const value = stringifyGridValue(readCellValue(row.rawRow, fieldKey)).trim();
  return value === "" ? "Unassigned" : value;
}

function entityGroupLabel(row: EntityRow, fieldKey: string) {
  const value = stringifyGridValue(readCellValue(row.rawRow, fieldKey)).trim();
  return value === "" ? "Unassigned" : value;
}

function rowFromApi(row: TimelineApiRow): WorkbookRow {
  const values: RowValues = {
    occurredAt: readStringCell(row, "timeline.occurred_at"),
    summary: readStringCell(row, "timeline.summary"),
    details: readStringCell(row, "timeline.details"),
    sourceText: readStringCell(row, "timeline.source_text"),
  };

  return {
    key: row.record_id,
    recordId: row.record_id,
    rowVersion: row.row_version,
    captureState: readStringCell(row, "timeline.capture_state"),
    values,
    committedValues: values,
    collectionValues: {
      hostRefs: readCollectionItems(row, "timeline.host_refs"),
      identityRefs: readCollectionItems(row, "timeline.identity_refs"),
      tags: readTagItems(row),
    },
    collectionDrafts: emptyCollectionDrafts(),
    pendingSignature: null,
    rawRow: row,
  };
}

function entityRowFromApi(
  row: EntityApiRow,
  entityType: EntityRow["entityType"],
): EntityRow {
  const labelField =
    entityType === "host" ? "host.display_name" : "identity.display_name";
  const secondaryCandidates =
    entityType === "host"
      ? ["host.hostname", "host.fqdn"]
      : ["identity.email", "identity.upn", "identity.sam_account_name"];
  const stateField =
    entityType === "host" ? "host.host_state" : "identity.identity_state";
  const aliasesField =
    entityType === "host" ? "host.aliases" : "identity.aliases";
  const linkedEventField =
    entityType === "host"
      ? "host.linked_event_count"
      : "identity.linked_event_count";
  const identifiers = mergeIdentifierFields[entityType]
    .map((field) => {
      const value = readStringCell(row, field.key);
      if (value === "") {
        return null;
      }
      return {
        key: field.key,
        label: field.label,
        value,
      };
    })
    .filter(
      (
        value,
      ): value is {
        key: string;
        label: string;
        value: string;
      } => value !== null,
    );
  const aliasItems = (() => {
    const raw = row.cells[aliasesField]?.value;
    if (
      !raw ||
      typeof raw !== "object" ||
      Array.isArray(raw) ||
      !("items" in raw) ||
      !Array.isArray(raw.items)
    ) {
      return [] as string[];
    }
    return raw.items
      .map((item) => {
        if (!item || typeof item !== "object") {
          return null;
        }
        const object = item as Record<string, unknown>;
        return typeof object.raw_text === "string" ? object.raw_text : null;
      })
      .filter((value): value is string => value !== null);
  })();
  const secondaryText =
    secondaryCandidates
      .map((field) => readStringCell(row, field))
      .find((value) => value !== "") ?? "";
  const label =
    readStringCell(row, labelField) || secondaryText || row.record_id;

  return {
    entityType,
    recordId: row.record_id,
    rowVersion: row.row_version,
    label,
    secondaryText,
    state: readStringCell(row, stateField),
    aliasTexts: aliasItems,
    linkedEventCount: readNumberCell(row, linkedEventField),
    rawRow: row,
    identifiers,
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

function ensureDraftRowWithFreshIndex(
  rows: WorkbookRow[],
  nextDraftIndex: () => number,
): {
  rows: WorkbookRow[];
  draftSummaryKey: string | null;
} {
  if (rows.some((row) => row.recordId === null)) {
    return {
      rows,
      draftSummaryKey: null,
    };
  }

  const draftIndex = nextDraftIndex();
  return {
    rows: [...rows, createDraftRow(draftIndex)],
    draftSummaryKey: `draft-${draftIndex}:summary`,
  };
}

function buildCollectionActions(rawInput: string) {
  const actions = rawInput
    .split(/\r?\n/u)
    .filter((segment) => segment.trim() !== "")
    .map((rawText) => ({
      op: "add_token",
      raw_text: rawText,
    }));
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

function buildScalarPatchPayload(row: WorkbookRow, clientTxnId: string) {
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

  if (changes.length < 1 || row.rowVersion === null) {
    return null;
  }

  return {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.rowVersion,
    client_txn_id: clientTxnId,
    changes,
  };
}

function buildCollectionPatchPayload(
  row: WorkbookRow,
  fieldKey: CollectionFieldKey,
  draftValue: string,
  clientTxnId: string,
) {
  const actionPayload = buildCollectionActions(draftValue);
  if (row.rowVersion === null || actionPayload === null) {
    return null;
  }

  return {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.rowVersion,
    client_txn_id: clientTxnId,
    changes: [
      {
        field_key: fieldKey,
        action_payload: actionPayload,
      },
    ],
  };
}

// Dedup queued autosaves by the logical mutation payload, not the per-request txn id.
function buildStableMutationSignature(payload: unknown): string {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return JSON.stringify(payload);
  }

  const { client_txn_id: _clientTxnID, ...stablePayload } = payload as Record<
    string,
    unknown
  >;
  return JSON.stringify(stablePayload);
}

function readEnvelope<T>(payload: unknown): T {
  return payload as T;
}

function readCookie(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }

  const prefix = `${name}=`;
  for (const segment of document.cookie.split(";")) {
    const trimmed = segment.trim();
    if (trimmed.startsWith(prefix)) {
      return decodeURIComponent(trimmed.slice(prefix.length));
    }
  }
  return null;
}

async function fetchJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<{
  ok: boolean;
  status: number;
  payload:
    | T
    | { error?: { code?: string; message?: string; details?: unknown } };
}> {
  const method = (init?.method ?? "GET").toUpperCase();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init?.headers as Record<string, string> | undefined),
  };
  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    const csrfToken = readCookie(csrfCookieName);
    if (csrfToken !== null && csrfToken !== "") {
      headers[csrfHeaderName] = csrfToken;
    }
  }

  const response = await fetch(input, {
    credentials: "include",
    ...init,
    headers,
  });
  const payload = (await response.json()) as
    | T
    | { error?: { code?: string; message?: string; details?: unknown } };
  return { ok: response.ok, status: response.status, payload };
}

function apiPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    return path;
  }
  return `${trimmedBase.replace(/\/$/, "")}${path}`;
}

function websocketPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${window.location.host}${path}`;
  }

  const target = new URL(trimmedBase);
  target.protocol = target.protocol === "https:" ? "wss:" : "ws:";
  target.pathname = path;
  target.search = "";
  target.hash = "";
  return target.toString();
}

function tabClientInstanceId(): string {
  const key = "cartulary.client_instance_id";
  try {
    const existing = window.sessionStorage.getItem(key);
    if (existing) {
      return existing;
    }
    const created =
      window.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`;
    window.sessionStorage.setItem(key, created);
    return created;
  } catch {
    return `${Date.now()}-${Math.random()}`;
  }
}

function workbookPresence() {
  return {
    sheet_ref: {
      kind: "view_schema",
      id: timelineViewSchemaId,
    },
    mode: "viewing",
  };
}

function sanitizeTestId(value: string) {
  return value.replace(/[^a-zA-Z0-9_-]+/gu, "-");
}

function focusTestIdForKey(focusKey: string) {
  const separatorIndex = focusKey.indexOf(":");
  if (separatorIndex < 0) {
    return null;
  }
  const rowKey = focusKey.slice(0, separatorIndex);
  const fieldKey = focusKey.slice(separatorIndex + 1);
  const fieldSuffix =
    fieldKey === "hostRefs" || fieldKey === "identityRefs"
      ? `${fieldKey}-input`
      : fieldKey;
  if (rowKey.startsWith("draft-")) {
    return `draft-row-${fieldSuffix}`;
  }
  return `row-${rowKey}-${fieldSuffix}`;
}

function relationshipItemLabel(
  item: CollectionItem | InspectorMention,
  entityIndex: Record<string, EntityRow>,
) {
  if ("status" in item && item.status === "dismissed") {
    return item.displayText || item.rawText;
  }
  if (item.resolvedRecordId) {
    const resolvedEntity = entityIndex[item.resolvedRecordId];
    if (resolvedEntity) {
      return resolvedEntity.label;
    }
  }
  return item.displayText || item.rawText;
}

function timelineRelationshipLabel(fieldKey: RelationshipFieldKey) {
  return fieldKey === "timeline.identity_refs" ? "Identities" : "Hosts";
}

function pruneDismissedMentions(
  dismissedMentionsByRow: Record<string, DismissedMention[]>,
  row: WorkbookRow,
) {
  if (row.recordId === null) {
    return dismissedMentionsByRow;
  }

  const activeRefs = new Set(
    [
      ...row.collectionValues.hostRefs,
      ...row.collectionValues.identityRefs,
    ].map((item) => item.itemRef),
  );
  const next = { ...dismissedMentionsByRow };
  const current = next[row.recordId] ?? [];
  const remaining = current.filter((item) => !activeRefs.has(item.itemRef));
  if (remaining.length < 1) {
    delete next[row.recordId];
    return next;
  }
  next[row.recordId] = remaining;
  return next;
}

function compareValue(value: string) {
  return value.trim().toLowerCase();
}

function buildMergePlan(survivor: EntityRow, loser: EntityRow) {
  const identifierLines: MergePlanLine[] = mergeIdentifierFields[
    survivor.entityType
  ].flatMap((field) => {
    const survivorValue =
      survivor.identifiers.find((identifier) => identifier.key === field.key)
        ?.value ?? "";
    const loserValue =
      loser.identifiers.find((identifier) => identifier.key === field.key)
        ?.value ?? "";
    if (survivorValue === "" && loserValue === "") {
      return [];
    }
    if (survivorValue === "" && loserValue !== "") {
      return [{ label: field.label, outcome: `Promote ${loserValue}` }];
    }
    if (
      survivorValue !== "" &&
      loserValue !== "" &&
      compareValue(survivorValue) === compareValue(loserValue)
    ) {
      return [{ label: field.label, outcome: `Duplicate no-op ${loserValue}` }];
    }
    if (survivorValue !== "" && loserValue !== "") {
      return [
        {
          label: field.label,
          outcome: `Conflict survivor=${survivorValue} loser=${loserValue}`,
        },
      ];
    }
    return [{ label: field.label, outcome: `Survivor keeps ${survivorValue}` }];
  });

  const survivorAliases = new Set(survivor.aliasTexts.map(compareValue));
  const aliasesToCopy = loser.aliasTexts.filter(
    (value) => !survivorAliases.has(compareValue(value)),
  );
  const duplicateAliases = loser.aliasTexts.filter((value) =>
    survivorAliases.has(compareValue(value)),
  );

  return {
    identifierLines,
    aliasesToCopy,
    duplicateAliases,
    provenanceOnlySummary: "Not exposed on this surface.",
    dependencySummary:
      survivor.linkedEventCount > 0 || loser.linkedEventCount > 0
        ? `Linked events visible on surface: survivor=${survivor.linkedEventCount}, loser=${loser.linkedEventCount}.`
        : "Dependency counts are not exposed on this surface.",
  };
}

function parseErrorMessage(payload: unknown) {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return "Request failed.";
  }
  const error = payload.error;
  if (!error || typeof error !== "object") {
    return "Request failed.";
  }
  if ("code" in error && typeof error.code === "string") {
    return error.code;
  }
  if ("message" in error && typeof error.message === "string") {
    return error.message;
  }
  return "Request failed.";
}

function RelationshipChip({
  item,
  entityIndex,
  onSelect,
  selected = false,
}: {
  item: CollectionItem | InspectorMention;
  entityIndex: Record<string, EntityRow>;
  onSelect?: () => void;
  selected?: boolean;
}) {
  const label = relationshipItemLabel(item, entityIndex);
  const isInspectorItem = "status" in item;
  const isResolved = isInspectorItem
    ? item.status === "resolved"
    : item.itemKind === "resolved_ref";
  const isDismissed = isInspectorItem ? item.status === "dismissed" : false;
  const isAutoResolved = item.autoResolved;
  const chipStyle = {
    ...relationshipChipStyle,
    ...(isDismissed
      ? dismissedChipStyle
      : isResolved
        ? isAutoResolved
          ? autoResolvedChipStyle
          : resolvedChipStyle
        : unresolvedChipStyle),
    ...(selected ? selectedChipStyle : null),
  };
  const labelPrefix = isDismissed
    ? "Dismissed"
    : isResolved
      ? isAutoResolved
        ? "Auto-resolved"
        : "Resolved"
      : "Unresolved";

  return onSelect ? (
    <button
      aria-label={`${labelPrefix} ${label}`}
      data-testid={`chip-${sanitizeTestId(item.itemRef)}`}
      style={chipStyle}
      type="button"
      onClick={onSelect}
    >
      <span>{label}</span>
      {isAutoResolved ? <span style={chipMetaStyle}>Auto</span> : null}
      {!isResolved && !isDismissed ? (
        <span style={chipMetaStyle}>Mention</span>
      ) : null}
    </button>
  ) : (
    <span
      aria-label={`${labelPrefix} ${label}`}
      data-testid={`chip-${sanitizeTestId(item.itemRef)}`}
      role="note"
      style={chipStyle}
    >
      <span>{label}</span>
      {isAutoResolved ? <span style={chipMetaStyle}>Auto</span> : null}
    </span>
  );
}

function DraftRowCreateButton({
  onCreate,
  row,
}: {
  readonly onCreate: (row: WorkbookRow) => void;
  readonly row: WorkbookRow;
}) {
  const createBlankRow = (
    event:
      | ReactKeyboardEvent<HTMLButtonElement>
      | ReactMouseEvent<HTMLButtonElement>,
  ) => {
    if (event.currentTarget.disabled) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    onCreate(row);
  };

  return (
    <button
      data-testid="draft-row-create"
      disabled={row.pendingSignature !== null}
      style={actionButtonStyle}
      type="button"
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          createBlankRow(event);
        }
      }}
      onMouseDown={createBlankRow}
    >
      Create blank row
    </button>
  );
}

export function TimelineWorkbook({
  incidentId,
  apiBase,
  hostEntities = [],
  identityEntities = [],
  entityIndex = {},
  currentIncidentRole = "",
  onRefreshEntities,
}: TimelineWorkbookProps) {
  const [rows, setRows] = useState<WorkbookRow[]>(() => [createDraftRow(1)]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [saveState, setSaveState] = useState<SaveState>("Saved");
  const [replacementDrafts, setReplacementDrafts] = useState<
    Record<string, string>
  >({});
  const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
  const [selectedMentionRef, setSelectedMentionRef] = useState<string | null>(
    null,
  );
  const [selectedResolveTargetId, setSelectedResolveTargetId] = useState("");
  const [inspectorMessage, setInspectorMessage] = useState<string | null>(null);
  const [dismissedMentionsByRow, setDismissedMentionsByRow] = useState<
    Record<string, DismissedMention[]>
  >({});
  const [autoResolutionNotices, setAutoResolutionNotices] = useState<
    AutoResolutionNotice[]
  >([]);
  const [queryState, setQueryState] = useState<WorkbookQueryState>(() =>
    emptyWorkbookQueryState(),
  );
  const [filterDraft, setFilterDraft] = useState<FilterDraft>(() =>
    defaultFilterDraft(timelineContract),
  );
  const draftCounterRef = useRef(2);
  const clientTxnRef = useRef(1);
  const pendingOpsRef = useRef(0);
  const pendingSignaturesRef = useRef(new Map<string, string>());
  const pendingSocketTxnTimeoutsRef = useRef(new Map<string, number>());
  const saveQueueRef = useRef(Promise.resolve());
  const rowsRef = useRef(rows);
  const loadSequenceRef = useRef(0);
  const loadRowsRef = useRef<(options: LoadRowsOptions) => Promise<void>>(
    async () => undefined,
  );
  const socketResumeTokenRef = useRef<string | null>(null);
  const socketLastSeenStreamSeqRef = useRef(0);
  const socketClientInstanceIdRef = useRef<string | null>(null);
  const rowInputRefs = useRef(
    new Map<string, HTMLInputElement | HTMLTextAreaElement>(),
  );
  const gridShellRef = useRef<HTMLDivElement | null>(null);
  const viewportContinuityTokenRef = useRef(1);
  const [viewportContinuityRequest, setViewportContinuityRequest] =
    useState<ViewportContinuityRequest | null>(null);

  const queryPath = useMemo(
    () =>
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
      ),
    [apiBase, incidentId],
  );
  const queryBody = useMemo(
    () => JSON.stringify(buildQueryRequest(timelineContract, queryState)),
    [queryState],
  );
  const changeSocketURL = useMemo(
    () => websocketPath(apiBase, `/ws/v1/incidents/${incidentId}`),
    [apiBase, incidentId],
  );
  const nextDraftIndex = useCallback(() => {
    const value = draftCounterRef.current;
    draftCounterRef.current += 1;
    return value;
  }, []);

  const selectedRow = useMemo(
    () =>
      rows.find(
        (row) => row.recordId !== null && row.recordId === selectedRowId,
      ) ?? null,
    [rows, selectedRowId],
  );
  const draftRow = useMemo(
    () => rows.find((row) => row.recordId === null) ?? null,
    [rows],
  );
  const dismissedForSelectedRow = selectedRow?.recordId
    ? (dismissedMentionsByRow[selectedRow.recordId] ?? [])
    : [];
  const inspectorMentions = useMemo(
    () =>
      buildInspectorMentions(selectedRow ?? undefined, dismissedForSelectedRow),
    [dismissedForSelectedRow, selectedRow],
  );
  const selectedMention =
    inspectorMentions.find((item) => item.itemRef === selectedMentionRef) ??
    inspectorMentions[0] ??
    null;
  const canManageMentions =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";

  const applyQueryFilter = useCallback(() => {
    setQueryState((current) => applyFilterDraft(current, filterDraft));
    setFilterDraft((current) => ({
      ...current,
      booleanValue: "",
      value: "",
    }));
  }, [filterDraft]);

  const handleQueryGroupByChange = useCallback((groupBy: string | null) => {
    setQueryState((current) =>
      updateGroupBy(timelineContract, current, groupBy),
    );
  }, []);

  const handleQuerySortToggle = useCallback((fieldKey: string) => {
    setQueryState((current) =>
      toggleSortField(timelineContract, current, fieldKey),
    );
  }, []);

  const currentGridScrollSnapshot = useCallback(() => {
    const element = gridShellRef.current;
    if (!element) {
      return null;
    }
    return {
      top: element.scrollTop,
      left: element.scrollLeft,
    };
  }, []);

  const currentGridViewportSnapshot = useCallback(
    (target: HTMLElement | null = null): ViewportSnapshot | null => {
      const gridShell = gridShellRef.current;
      const scroll = currentGridScrollSnapshot();
      if (gridShell === null || scroll === null) {
        return null;
      }
      return {
        scroll,
        anchor:
          target === null
            ? null
            : captureViewportAnchor(
                gridShell.getBoundingClientRect(),
                target.getBoundingClientRect(),
              ),
      };
    },
    [currentGridScrollSnapshot],
  );

  const trackPendingSocketTxn = useCallback((clientTxnId: string) => {
    const existingTimeout =
      pendingSocketTxnTimeoutsRef.current.get(clientTxnId);
    if (existingTimeout !== undefined) {
      window.clearTimeout(existingTimeout);
    }
    const timeoutId = window.setTimeout(() => {
      pendingSocketTxnTimeoutsRef.current.delete(clientTxnId);
    }, 30_000);
    pendingSocketTxnTimeoutsRef.current.set(clientTxnId, timeoutId);
  }, []);

  const resolvePendingSocketTxn = useCallback(
    (clientTxnId: string | null | undefined) => {
      if (!clientTxnId) {
        return false;
      }

      const timeoutId = pendingSocketTxnTimeoutsRef.current.get(clientTxnId);
      if (timeoutId === undefined) {
        return false;
      }

      window.clearTimeout(timeoutId);
      pendingSocketTxnTimeoutsRef.current.delete(clientTxnId);
      return true;
    },
    [],
  );

  const restoreGridScroll = useCallback(
    (preservedScroll: ScrollPosition | null) => {
      const gridShell = gridShellRef.current;
      if (gridShell === null || preservedScroll === null) {
        return;
      }
      gridShell.scrollTop = preservedScroll.top;
      gridShell.scrollLeft = preservedScroll.left;
      window.requestAnimationFrame(() => {
        const currentGridShell = gridShellRef.current;
        if (currentGridShell === null) {
          return;
        }
        currentGridShell.scrollTop = preservedScroll.top;
        currentGridShell.scrollLeft = preservedScroll.left;
      });
    },
    [],
  );

  const restoreGridViewportForElement = useCallback(
    (
      resolveElement: () =>
        | HTMLButtonElement
        | HTMLInputElement
        | HTMLTextAreaElement
        | null,
      preservedViewport: ViewportSnapshot | null,
    ) => {
      const element = resolveElement();
      if (element === null) {
        return false;
      }
      // Continuity restores the previous scroll position first, then applies
      // only the extra delta needed to keep the target fully visible.
      const currentViewport =
        preservedViewport ??
        ({
          scroll: currentGridScrollSnapshot(),
          anchor: null,
        } satisfies ViewportSnapshot);
      const preservedScroll = currentViewport.scroll;
      window.focus();
      element.focus({ preventScroll: true });
      restoreGridScroll(preservedScroll);
      const restoreViewportGeometryNow = () => {
        const currentGridShell = gridShellRef.current;
        const currentElement = resolveElement();
        if (
          currentGridShell === null ||
          preservedScroll === null ||
          currentElement === null ||
          !currentElement.isConnected
        ) {
          return false;
        }
        restoreGridScroll(
          computeRestoredViewportScroll({
            preservedScroll,
            preservedAnchor: currentViewport.anchor,
            containerRect: currentGridShell.getBoundingClientRect(),
            elementRect: currentElement.getBoundingClientRect(),
          }),
        );
        return isRectFullyVisibleWithinContainer(
          currentGridShell.getBoundingClientRect(),
          currentElement.getBoundingClientRect(),
        );
      };
      const restoredNow = restoreViewportGeometryNow();
      const restoreViewportGeometry = (attempt: number) => {
        window.requestAnimationFrame(() => {
          if (restoreViewportGeometryNow()) {
            return;
          }
          if (attempt < 6) {
            restoreViewportGeometry(attempt + 1);
          }
        });
      };
      restoreViewportGeometry(0);
      return document.activeElement === element && restoredNow;
    },
    [currentGridScrollSnapshot, restoreGridScroll],
  );

  const resolveInputElement = useCallback((focusKey: string) => {
    const selectorTestId = focusTestIdForKey(focusKey);
    const selector =
      selectorTestId === null
        ? null
        : document.querySelector<HTMLInputElement | HTMLTextAreaElement>(
            `[data-testid="${selectorTestId}"]`,
          );
    return selector ?? rowInputRefs.current.get(focusKey) ?? null;
  }, []);

  const resolveViewportContinuityElement = useCallback(
    (target: ViewportContinuityTarget) => {
      switch (target.kind) {
        case "row-inspect":
          return document.querySelector<HTMLButtonElement>(
            `[data-testid="${rowInspectButtonTestId(target.recordId)}"]`,
          );
        case "input":
          return resolveInputElement(target.focusKey);
        case "scroll-only":
          return null;
      }
    },
    [resolveInputElement],
  );

  const beginViewportContinuity = useCallback(
    (
      target: ViewportContinuityTarget,
      options: { followup?: ViewportContinuityFollowup } = {},
    ) => {
      const token = viewportContinuityTokenRef.current;
      viewportContinuityTokenRef.current += 1;
      const followup = options.followup ?? "none";
      setViewportContinuityRequest({
        token,
        attemptVersion: 0,
        target,
        preservedViewport: currentGridViewportSnapshot(
          resolveViewportContinuityElement(target),
        ),
        followup,
        followupSettled: followup === "none",
        baselineHostEntities: hostEntities,
        baselineIdentityEntities: identityEntities,
      });
      return token;
    },
    [
      currentGridViewportSnapshot,
      hostEntities,
      identityEntities,
      resolveViewportContinuityElement,
    ],
  );

  const settleViewportContinuityFollowup = useCallback((token: number) => {
    setViewportContinuityRequest((current) => {
      if (!current || current.token !== token) {
        return current;
      }
      return {
        ...current,
        followupSettled: true,
        attemptVersion: current.attemptVersion + 1,
      };
    });
  }, []);

  const clearViewportContinuity = useCallback((token: number) => {
    setViewportContinuityRequest((current) =>
      current?.token === token ? null : current,
    );
  }, []);

  const advanceViewportContinuity = useCallback(
    (
      token: number | undefined,
      options: {
        target?: ViewportContinuityTarget | null;
      } = {},
    ) => {
      if (token === undefined) {
        return;
      }
      setViewportContinuityRequest((current) => {
        if (current === null || current.token !== token) {
          return current;
        }
        return {
          ...current,
          attemptVersion: current.attemptVersion + 1,
          target: options.target ?? current.target,
        };
      });
    },
    [],
  );

  const tryRestoreViewportContinuity = useCallback(
    (continuity: ViewportContinuityRequest) => {
      if (continuity.target.kind === "scroll-only") {
        restoreGridScroll(continuity.preservedViewport?.scroll ?? null);
        return true;
      }
      return restoreGridViewportForElement(
        () => resolveViewportContinuityElement(continuity.target),
        continuity.preservedViewport,
      );
    },
    [
      resolveViewportContinuityElement,
      restoreGridScroll,
      restoreGridViewportForElement,
    ],
  );

  const shouldHoldViewportContinuity = useCallback(
    (continuity: ViewportContinuityRequest) => {
      if (continuity.followup !== "entity-refresh") {
        return false;
      }
      if (!continuity.followupSettled) {
        return true;
      }
      return (
        continuity.baselineHostEntities === hostEntities &&
        continuity.baselineIdentityEntities === identityEntities
      );
    },
    [hostEntities, identityEntities],
  );

  const applyRowMutation = useCallback(
    (
      rowKey: string,
      envelope: TimelineMutationEnvelope,
      options: {
        continueOnFreshDraft?: boolean;
        detectAutoResolution?: boolean;
        promoteToCommittedRowInspect?: boolean;
        viewportContinuityToken?: number;
      } = {},
    ) => {
      const previousRow = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      const committed = rowFromApi(envelope.data.row);
      const nextRows = rowsRef.current.map((row) =>
        row.key === rowKey ? committed : row,
      );
      const hydrated = ensureDraftRowWithFreshIndex(nextRows, nextDraftIndex);
      const hydratedRows = hydrated.rows;

      rowsRef.current = hydratedRows;
      setRows(hydratedRows);

      if (committed.recordId !== null) {
        setDismissedMentionsByRow((current) =>
          pruneDismissedMentions(current, committed),
        );
      }
      if (options.detectAutoResolution !== false) {
        const notices = buildAutoResolutionNotices(previousRow, committed);
        if (notices.length > 0) {
          setAutoResolutionNotices((current) => {
            const knownRefs = new Set(current.map((notice) => notice.itemRef));
            return [
              ...current,
              ...notices.filter((notice) => !knownRefs.has(notice.itemRef)),
            ];
          });
        }
      }
      if (
        selectedRowId !== null &&
        previousRow?.recordId !== null &&
        previousRow !== undefined &&
        previousRow.recordId === selectedRowId
      ) {
        setSelectedRowId(committed.recordId);
      }
      const nextViewportTarget =
        options.continueOnFreshDraft && hydrated.draftSummaryKey
          ? ({
              kind: "input",
              focusKey: hydrated.draftSummaryKey,
            } satisfies ViewportContinuityTarget)
          : options.promoteToCommittedRowInspect && committed.recordId !== null
            ? ({
                kind: "row-inspect",
                recordId: committed.recordId,
              } satisfies ViewportContinuityTarget)
            : null;
      advanceViewportContinuity(options.viewportContinuityToken, {
        target: nextViewportTarget,
      });
    },
    [advanceViewportContinuity, nextDraftIndex, selectedRowId],
  );

  const loadRows = useCallback(
    async (options: LoadRowsOptions) => {
      const requestSequence = loadSequenceRef.current + 1;
      loadSequenceRef.current = requestSequence;

      if (options.showLoading) {
        setIsLoading(true);
      }
      setLoadError(null);

      const result = await fetchJSON<WorkbookQueryEnvelope>(queryPath, {
        method: "POST",
        body: queryBody,
      });

      if (requestSequence !== loadSequenceRef.current) {
        return;
      }

      if (!result.ok) {
        if (options.viewportContinuityToken !== undefined) {
          clearViewportContinuity(options.viewportContinuityToken);
        }
        setLoadError("Timeline projection load failed.");
        setIsLoading(false);
        return;
      }

      const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
      const projectedRows = [
        ...reconcileRecordRows(
          rowsRef.current.filter((row) => row.recordId !== null),
          envelope.data.rows.map(rowFromApi),
        ),
      ];
      const hydratedRows = ensureDraftRowWithFreshIndex(
        projectedRows,
        nextDraftIndex,
      ).rows;
      startTransition(() => {
        rowsRef.current = hydratedRows;
        setRows(hydratedRows);
      });
      advanceViewportContinuity(options.viewportContinuityToken);
      setDismissedMentionsByRow((current) => {
        const next = { ...current };
        for (const row of projectedRows) {
          if (row.recordId === null) {
            continue;
          }
          Object.assign(next, pruneDismissedMentions(next, row));
        }
        return next;
      });
      setSaveState("Saved");
      setIsLoading(false);
    },
    [
      advanceViewportContinuity,
      clearViewportContinuity,
      nextDraftIndex,
      queryBody,
      queryPath,
    ],
  );

  loadRowsRef.current = loadRows;

  useEffect(() => {
    void loadRows({ showLoading: true });
  }, [loadRows]);

  useLayoutEffect(() => {
    if (
      viewportContinuityRequest === null ||
      viewportContinuityRequest.attemptVersion < 1
    ) {
      return;
    }
    let cancelled = false;
    const restoreTarget = (attempt: number) => {
      if (cancelled) {
        return;
      }
      if (!tryRestoreViewportContinuity(viewportContinuityRequest)) {
        if (attempt < 60) {
          window.setTimeout(() => {
            restoreTarget(attempt + 1);
          }, 50);
        }
        return;
      }
      if (shouldHoldViewportContinuity(viewportContinuityRequest)) {
        return;
      }
      clearViewportContinuity(viewportContinuityRequest.token);
    };
    restoreTarget(0);
    return () => {
      cancelled = true;
    };
  }, [
    clearViewportContinuity,
    shouldHoldViewportContinuity,
    tryRestoreViewportContinuity,
    viewportContinuityRequest,
  ]);

  useEffect(() => {
    return () => {
      for (const timeoutId of pendingSocketTxnTimeoutsRef.current.values()) {
        window.clearTimeout(timeoutId);
      }
      pendingSocketTxnTimeoutsRef.current.clear();
    };
  }, []);

  useEffect(() => {
    if (incidentId.trim() === "") {
      return;
    }

    let closed = false;
    let socket: WebSocket | null = null;
    let reconnectTimer: number | null = null;
    const clientInstanceId =
      socketClientInstanceIdRef.current ?? tabClientInstanceId();
    socketClientInstanceIdRef.current = clientInstanceId;

    const scheduleReconnect = () => {
      if (closed || reconnectTimer !== null) {
        return;
      }
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, 1000);
    };

    const sendSessionEstablishment = (target: WebSocket) => {
      const resumeToken = socketResumeTokenRef.current;
      if (resumeToken) {
        target.send(
          JSON.stringify({
            type: "resume",
            payload: {
              client_instance_id: clientInstanceId,
              resume_token: resumeToken,
              last_seen_stream_seq: socketLastSeenStreamSeqRef.current,
              presence: workbookPresence(),
            },
          }),
        );
        return;
      }
      target.send(
        JSON.stringify({
          type: "hello",
          payload: {
            client_instance_id: clientInstanceId,
            presence: workbookPresence(),
          },
        }),
      );
    };

    const handleMessage = (target: WebSocket, raw: unknown) => {
      if (!raw || typeof raw !== "object") {
        return;
      }
      const message = raw as {
        type?: string;
        stream_seq?: number;
        payload?: Record<string, unknown>;
      };
      if (message.type === "ping") {
        target.send(JSON.stringify({ type: "pong", payload: {} }));
        return;
      }
      if (message.type === "hello_ack" || message.type === "resume_ack") {
        const resumeToken = message.payload?.resume_token;
        if (typeof resumeToken === "string") {
          socketResumeTokenRef.current = resumeToken;
        }
        if (
          message.type === "resume_ack" &&
          message.payload?.status === "reset_required"
        ) {
          void loadRowsRef.current({ showLoading: false });
        }
        return;
      }
      if (message.type === "session_revoked") {
        socketResumeTokenRef.current = null;
        target.close();
        return;
      }
      if (
        shouldIgnoreSelfOriginatedRecordChange(raw, resolvePendingSocketTxn)
      ) {
        return;
      }
      if (!isRecordChangedMessage(raw)) {
        return;
      }
      if (typeof message.stream_seq === "number") {
        socketLastSeenStreamSeqRef.current = Math.max(
          socketLastSeenStreamSeqRef.current,
          message.stream_seq,
        );
      }
      const viewportContinuityToken = beginViewportContinuity({
        kind: "scroll-only",
      });
      void loadRowsRef.current({
        showLoading: false,
        viewportContinuityToken,
      });
    };

    const connect = () => {
      if (closed) {
        return;
      }
      socket = new WebSocket(changeSocketURL);
      socket.onopen = () => {
        if (socket) {
          sendSessionEstablishment(socket);
        }
      };
      socket.onmessage = (event) => {
        if (!socket) {
          return;
        }
        handleMessage(socket, JSON.parse(event.data) as unknown);
      };
      socket.onclose = scheduleReconnect;
      socket.onerror = () => {
        socket?.close();
      };
    };

    connect();
    return () => {
      closed = true;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
      }
      socket?.close();
    };
  }, [
    beginViewportContinuity,
    changeSocketURL,
    incidentId,
    resolvePendingSocketTxn,
  ]);

  useEffect(() => {
    if (selectedRowId === null) {
      return;
    }
    if (!rows.some((row) => row.recordId === selectedRowId)) {
      setSelectedRowId(null);
    }
  }, [rows, selectedRowId]);

  useEffect(() => {
    if (inspectorMentions.length < 1) {
      if (selectedMentionRef !== null) {
        setSelectedMentionRef(null);
      }
      setSelectedResolveTargetId("");
      return;
    }
    if (
      selectedMentionRef !== null &&
      inspectorMentions.some((item) => item.itemRef === selectedMentionRef)
    ) {
      return;
    }
    const [firstMention] = inspectorMentions;
    if (firstMention) {
      setSelectedMentionRef(firstMention.itemRef);
    }
    setSelectedResolveTargetId("");
  }, [inspectorMentions, selectedMentionRef]);

  const nextClientTxnId = useCallback(() => {
    const value = clientTxnRef.current;
    clientTxnRef.current += 1;
    return `timeline-client-${value}`;
  }, []);

  const beginSave = useCallback(() => {
    pendingOpsRef.current += 1;
    setSaveState("Syncing");
  }, []);

  const finishSave = useCallback((nextState: SaveState) => {
    pendingOpsRef.current = Math.max(0, pendingOpsRef.current - 1);
    if (pendingOpsRef.current > 0 && nextState !== "Conflict") {
      setSaveState("Syncing");
      return;
    }
    setSaveState(nextState);
  }, []);

  const setRowValue = useCallback(
    (rowKey: string, field: keyof RowValues, value: string) => {
      setRows((current) => {
        const nextRows = current.map((row) =>
          row.key === rowKey
            ? {
                ...row,
                values: {
                  ...row.values,
                  [field]: value,
                },
              }
            : row,
        );
        rowsRef.current = nextRows;
        return nextRows;
      });
    },
    [],
  );

  const registerInput = useCallback(
    (
      rowKey: string,
      field: FocusFieldKey,
      element: HTMLInputElement | HTMLTextAreaElement | null,
    ) => {
      const key = `${rowKey}:${field}`;
      if (element === null) {
        rowInputRefs.current.delete(key);
        return;
      }
      rowInputRefs.current.set(key, element);
    },
    [],
  );

  const queueScalarSave = useCallback(
    (
      rowKey: string,
      focusField: keyof RowValues,
      options: {
        allowZeroFieldCreate?: boolean;
        continueOnFreshDraft: boolean;
        preserveInputFocus: boolean;
        snapshotOverride?: WorkbookRow;
      },
    ) => {
      const snapshot =
        options.snapshotOverride ??
        rowsRef.current.find((candidate) => candidate.key === rowKey);
      if (!snapshot) {
        return;
      }

      const clientTxnId = nextClientTxnId();
      const payload =
        snapshot.recordId === null
          ? buildCreatePayload(snapshot, clientTxnId, {
              allowZeroFieldCreate: options.allowZeroFieldCreate === true,
            })
          : buildScalarPatchPayload(snapshot, clientTxnId);
      if (payload === null) {
        return;
      }

      const mutationSignature = buildStableMutationSignature(payload);
      if (pendingSignaturesRef.current.get(rowKey) === mutationSignature) {
        return;
      }
      const viewportContinuityToken = beginViewportContinuity(
        options.preserveInputFocus
          ? {
              kind: "input",
              focusKey: `${rowKey}:${focusField}`,
            }
          : {
              kind: "scroll-only",
            },
      );
      pendingSignaturesRef.current.set(rowKey, mutationSignature);
      beginSave();

      startTransition(() => {
        setRows((current) => {
          const nextRows = current.map((row) =>
            row.key === rowKey
              ? { ...row, pendingSignature: mutationSignature }
              : row,
          );
          rowsRef.current = nextRows;
          return nextRows;
        });
      });

      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          trackPendingSocketTxn(clientTxnId);
          const targetPath =
            snapshot.recordId === null
              ? apiPath(
                  apiBase,
                  `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
                )
              : apiPath(apiBase, `/api/v1/records/${snapshot.recordId}`);
          const method = snapshot.recordId === null ? "POST" : "PATCH";
          const result = await fetchJSON<TimelineMutationEnvelope>(targetPath, {
            method,
            body: JSON.stringify(payload),
          });

          if (!result.ok) {
            resolvePendingSocketTxn(clientTxnId);
            pendingSignaturesRef.current.delete(rowKey);
            setRows((current) => {
              const nextRows = current.map((row) =>
                row.key === rowKey ? { ...row, pendingSignature: null } : row,
              );
              rowsRef.current = nextRows;
              return nextRows;
            });
            clearViewportContinuity(viewportContinuityToken);
            finishSave("Conflict");
            return;
          }

          pendingSignaturesRef.current.delete(rowKey);
          const envelope = readEnvelope<TimelineMutationEnvelope>(
            result.payload,
          );
          applyRowMutation(rowKey, envelope, {
            continueOnFreshDraft:
              options.continueOnFreshDraft && snapshot.recordId === null,
            detectAutoResolution: false,
            viewportContinuityToken,
          });
          finishSave("Saved");
        });
    },
    [
      apiBase,
      applyRowMutation,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      finishSave,
      incidentId,
      nextClientTxnId,
      resolvePendingSocketTxn,
      trackPendingSocketTxn,
    ],
  );

  const queueCollectionSave = useCallback(
    (
      rowKey: string,
      fieldKey: CollectionFieldKey,
      focusField: CollectionDraftKey,
      draftValueOverride?: string,
    ) => {
      const snapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      if (!snapshot) {
        return;
      }
      const draftValue =
        draftValueOverride ?? snapshot.collectionDrafts[focusField];
      const effectiveSnapshot =
        draftValueOverride === undefined
          ? snapshot
          : {
              ...snapshot,
              collectionDrafts: {
                ...snapshot.collectionDrafts,
                [focusField]: draftValue,
              },
            };
      const clientTxnId = nextClientTxnId();
      const payload =
        snapshot.recordId === null
          ? buildCreatePayload(effectiveSnapshot, clientTxnId)
          : buildCollectionPatchPayload(
              effectiveSnapshot,
              fieldKey,
              draftValue,
              clientTxnId,
            );
      if (payload === null) {
        return;
      }

      const viewportContinuityToken = beginViewportContinuity(
        snapshot.recordId === null
          ? {
              kind: "scroll-only",
            }
          : {
              kind: "row-inspect",
              recordId: snapshot.recordId,
            },
      );
      beginSave();
      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          trackPendingSocketTxn(clientTxnId);
          const targetPath =
            snapshot.recordId === null
              ? apiPath(
                  apiBase,
                  `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
                )
              : apiPath(apiBase, `/api/v1/records/${snapshot.recordId}`);
          const method = snapshot.recordId === null ? "POST" : "PATCH";
          const result = await fetchJSON<TimelineMutationEnvelope>(targetPath, {
            method,
            body: JSON.stringify(payload),
          });
          if (!result.ok) {
            resolvePendingSocketTxn(clientTxnId);
            clearViewportContinuity(viewportContinuityToken);
            setInspectorMessage(parseErrorMessage(result.payload));
            finishSave("Conflict");
            return;
          }

          const envelope = readEnvelope<TimelineMutationEnvelope>(
            result.payload,
          );
          applyRowMutation(rowKey, envelope, {
            promoteToCommittedRowInspect: snapshot.recordId === null,
            viewportContinuityToken,
          });
          finishSave("Saved");
        });
    },
    [
      apiBase,
      applyRowMutation,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      finishSave,
      incidentId,
      nextClientTxnId,
      resolvePendingSocketTxn,
      trackPendingSocketTxn,
    ],
  );

  const queueAction = useCallback(
    (rowKey: string, action: "mark-reviewed" | "supersede") => {
      const snapshot = rowsRef.current.find(
        (candidate) => candidate.key === rowKey,
      );
      if (
        !snapshot ||
        snapshot.recordId === null ||
        snapshot.rowVersion === null ||
        (action === "supersede" &&
          normalizeValue(replacementDrafts[rowKey] ?? "") === "")
      ) {
        return;
      }

      const clientTxnId = nextClientTxnId();
      const viewportContinuityToken = beginViewportContinuity({
        kind: "row-inspect",
        recordId: snapshot.recordId,
      });
      beginSave();
      saveQueueRef.current = saveQueueRef.current
        .catch(() => undefined)
        .then(async () => {
          const body =
            action === "mark-reviewed"
              ? {
                  base_row_version: snapshot.rowVersion,
                  client_txn_id: clientTxnId,
                  reason: "Reviewed from workbook",
                }
              : {
                  base_row_version: snapshot.rowVersion,
                  client_txn_id: clientTxnId,
                  reason: "Superseded from workbook",
                  replacement_record_id: normalizeValue(
                    replacementDrafts[rowKey] ?? "",
                  ),
                };
          trackPendingSocketTxn(clientTxnId);
          const result = await fetchJSON<TimelineActionEnvelope>(
            apiPath(apiBase, `/api/v1/records/${snapshot.recordId}/${action}`),
            {
              method: "POST",
              body: JSON.stringify(body),
            },
          );
          if (!result.ok) {
            resolvePendingSocketTxn(clientTxnId);
            clearViewportContinuity(viewportContinuityToken);
            finishSave("Conflict");
            return;
          }

          await loadRowsRef.current({
            showLoading: false,
            viewportContinuityToken,
          });
          finishSave("Saved");
        });
    },
    [
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      finishSave,
      nextClientTxnId,
      replacementDrafts,
      resolvePendingSocketTxn,
      trackPendingSocketTxn,
    ],
  );

  function submitMentionAction(
    mention: InspectorMention,
    action: "resolve_item" | "dismiss_item" | "revert_to_unresolved",
    resolvedRecordId?: string,
  ) {
    const snapshot = rowsRef.current.find(
      (candidate) => candidate.recordId === mention.rowRecordId,
    );
    if (!snapshot || snapshot.recordId === null) {
      return;
    }

    const clientTxnId = nextClientTxnId();
    const payload = buildMentionPatchPayload(
      snapshot,
      mention,
      action,
      clientTxnId,
      resolvedRecordId,
    );
    if (payload === null) {
      return;
    }

    const viewportContinuityToken = beginViewportContinuity(
      {
        kind: "row-inspect",
        recordId: snapshot.recordId,
      },
      {
        followup:
          action === "resolve_item" && resolvedRecordId === undefined
            ? "entity-refresh"
            : "none",
      },
    );
    beginSave();
    setInspectorMessage(null);
    saveQueueRef.current = saveQueueRef.current
      .catch(() => undefined)
      .then(async () => {
        trackPendingSocketTxn(clientTxnId);
        const result = await fetchJSON<TimelineMutationEnvelope>(
          apiPath(apiBase, `/api/v1/records/${snapshot.recordId}`),
          {
            method: "PATCH",
            body: JSON.stringify(payload),
          },
        );
        if (!result.ok) {
          resolvePendingSocketTxn(clientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(parseErrorMessage(result.payload));
          finishSave("Conflict");
          return;
        }

        if (action === "dismiss_item") {
          setDismissedMentionsByRow((current) => {
            const rowMentions = current[snapshot.recordId ?? ""] ?? [];
            return {
              ...current,
              [snapshot.recordId ?? ""]: [
                ...rowMentions.filter(
                  (item) => item.itemRef !== mention.itemRef,
                ),
                {
                  rowRecordId: snapshot.recordId ?? "",
                  fieldKey: mention.fieldKey,
                  entityType: mention.entityType,
                  itemRef: mention.itemRef,
                  rawText: mention.rawText,
                  resolvedRecordId: mention.resolvedRecordId,
                  resolutionMethod: mention.resolutionMethod,
                  autoResolved: mention.autoResolved,
                },
              ],
            };
          });
        }
        if (action === "revert_to_unresolved") {
          setDismissedMentionsByRow((current) => {
            if (!snapshot.recordId) {
              return current;
            }
            const rowMentions = (current[snapshot.recordId] ?? []).filter(
              (item) => item.itemRef !== mention.itemRef,
            );
            if (rowMentions.length < 1) {
              const next = { ...current };
              delete next[snapshot.recordId];
              return next;
            }
            return {
              ...current,
              [snapshot.recordId]: rowMentions,
            };
          });
        }

        const envelope = readEnvelope<TimelineMutationEnvelope>(result.payload);
        applyRowMutation(snapshot.key, envelope, {
          detectAutoResolution: false,
          viewportContinuityToken,
        });
        finishSave("Saved");
        if (action === "resolve_item" && resolvedRecordId === undefined) {
          try {
            await onRefreshEntities?.();
          } finally {
            settleViewportContinuityFollowup(viewportContinuityToken);
          }
        }
      });
  }

  const handleBlur = useCallback(
    (rowKey: string, focusField: keyof RowValues) => {
      queueScalarSave(rowKey, focusField, {
        continueOnFreshDraft: false,
        preserveInputFocus: false,
      });
    },
    [queueScalarSave],
  );

  const handleKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
      rowKey: string,
      focusField: keyof RowValues,
    ) => {
      if (event.key === "Enter" || event.key === "Tab") {
        event.preventDefault();
        queueScalarSave(rowKey, focusField, {
          continueOnFreshDraft: true,
          preserveInputFocus: true,
        });
      }
    },
    [queueScalarSave],
  );

  const handleCollectionKeyDown = useCallback(
    (
      event: ReactKeyboardEvent<HTMLInputElement>,
      rowKey: string,
      fieldKey: CollectionFieldKey,
      draftKey: CollectionDraftKey,
    ) => {
      if (event.key === "Enter" || event.key === "Tab") {
        event.preventDefault();
        queueCollectionSave(
          rowKey,
          fieldKey,
          draftKey,
          event.currentTarget.value,
        );
      }
    },
    [queueCollectionSave],
  );

  const handlePaste = useCallback(
    (rowKey: string, focusField: keyof RowValues) => {
      window.setTimeout(() => {
        queueScalarSave(rowKey, focusField, {
          continueOnFreshDraft: false,
          preserveInputFocus: true,
        });
      }, 0);
    },
    [queueScalarSave],
  );

  const handleSelectRow = useCallback((recordId: string) => {
    setSelectedRowId(recordId);
    setInspectorMessage(null);
  }, []);

  const handleSelectMention = useCallback(
    (rowRecordId: string, itemRef: string) => {
      setSelectedRowId(rowRecordId);
      setSelectedMentionRef(itemRef);
      setInspectorMessage(null);
    },
    [],
  );

  const handleCreateBlankDraftRow = useCallback(
    (row: WorkbookRow) => {
      const activeRow =
        rowsRef.current.find((candidate) => candidate.key === row.key) ??
        rowsRef.current.find((candidate) => candidate.recordId === null) ??
        row;
      queueScalarSave(activeRow.key, "summary", {
        allowZeroFieldCreate: true,
        continueOnFreshDraft: true,
        preserveInputFocus: false,
        snapshotOverride: activeRow,
      });
    },
    [queueScalarSave],
  );

  const timelineBindingLabel = useCallback((fieldKey: string) => {
    return timelineContract.fieldMap[fieldKey]?.label ?? fieldKey;
  }, []);

  const timelineScalarControlId = useCallback(
    (
      row: WorkbookRow,
      binding: TimelineScalarBinding,
      surface: TimelineScalarEditorSurface,
    ) => {
      return ["timeline-editor", surface, row.key, binding.fieldKey]
        .map((value) => value.replace(/[^a-zA-Z0-9_-]+/g, "-"))
        .join("-");
    },
    [],
  );

  const renderTimelineScalarControl = useCallback(
    (
      row: WorkbookRow,
      binding: TimelineScalarBinding,
      surface: TimelineScalarEditorSurface,
      controlId: string,
    ) => {
      const label = timelineBindingLabel(binding.fieldKey);
      const gridAccessibleLabel =
        surface === "grid"
          ? `${label} ${row.recordId ?? "draft row"}`
          : undefined;
      const dataTestId =
        row.recordId === null
          ? draftCellTestId(binding.key)
          : rowCellTestId(row.recordId, binding.key);
      if (binding.multiline) {
        return (
          <textarea
            aria-label={gridAccessibleLabel}
            data-testid={dataTestId}
            id={controlId}
            ref={(element) => {
              registerInput(row.key, binding.key, element);
            }}
            rows={3}
            style={textareaStyle}
            value={row.values[binding.key]}
            onBlur={() => {
              handleBlur(row.key, binding.key);
            }}
            onChange={(event: ChangeEvent<HTMLTextAreaElement>) => {
              setRowValue(row.key, binding.key, event.target.value);
            }}
            onFocus={() => {
              if (row.recordId) {
                handleSelectRow(row.recordId);
              }
            }}
            onKeyDown={(event) => {
              handleKeyDown(event, row.key, binding.key);
            }}
            onPaste={() => {
              handlePaste(row.key, binding.key);
            }}
          />
        );
      }
      return (
        <input
          aria-label={gridAccessibleLabel}
          data-testid={dataTestId}
          id={controlId}
          ref={(element) => {
            registerInput(row.key, binding.key, element);
          }}
          style={inputStyle}
          type="text"
          value={row.values[binding.key]}
          onBlur={() => {
            handleBlur(row.key, binding.key);
          }}
          onChange={(event: ChangeEvent<HTMLInputElement>) => {
            setRowValue(row.key, binding.key, event.target.value);
          }}
          onFocus={() => {
            if (row.recordId) {
              handleSelectRow(row.recordId);
            }
          }}
          onKeyDown={(event) => {
            handleKeyDown(event, row.key, binding.key);
          }}
          onPaste={() => {
            handlePaste(row.key, binding.key);
          }}
        />
      );
    },
    [
      handleBlur,
      handleKeyDown,
      handlePaste,
      handleSelectRow,
      registerInput,
      setRowValue,
      timelineBindingLabel,
    ],
  );

  const renderTimelineGridEditor = useCallback(
    (row: WorkbookRow, binding: TimelineScalarBinding) => {
      return renderTimelineScalarControl(
        row,
        binding,
        "grid",
        timelineScalarControlId(row, binding, "grid"),
      );
    },
    [renderTimelineScalarControl, timelineScalarControlId],
  );

  const renderTimelineInspectorEditor = useCallback(
    (row: WorkbookRow, binding: TimelineScalarBinding) => {
      const controlId = timelineScalarControlId(row, binding, "inspector");
      return (
        <div key={binding.fieldKey} style={labelStyle}>
          <label htmlFor={controlId}>
            {timelineBindingLabel(binding.fieldKey)}
          </label>
          {renderTimelineScalarControl(row, binding, "inspector", controlId)}
        </div>
      );
    },
    [
      renderTimelineScalarControl,
      timelineBindingLabel,
      timelineScalarControlId,
    ],
  );

  const renderTimelineCollectionInput = useCallback(
    (row: WorkbookRow, binding: TimelineCollectionBinding) => {
      const label = timelineBindingLabel(binding.fieldKey);
      const items = row.collectionValues[binding.draftKey];
      return (
        <>
          <div
            data-testid={
              row.recordId === null
                ? draftCellTestId(`${binding.draftKey}-items`)
                : relationshipItemsTestId(row.recordId, binding.draftKey)
            }
            style={relationshipItemsWrapStyle}
          >
            {items.length > 0 ? (
              binding.collectionKind === "relationship" ? (
                items.map((item) => (
                  <RelationshipChip
                    key={item.itemRef}
                    entityIndex={entityIndex}
                    item={item as CollectionItem}
                    onSelect={() => {
                      if (row.recordId) {
                        handleSelectMention(row.recordId, item.itemRef);
                      }
                    }}
                  />
                ))
              ) : (
                items.map((item) => (
                  <span key={item.itemRef} style={tagChipStyle}>
                    {(item as TagCollectionItem).displayText}
                  </span>
                ))
              )
            ) : (
              <span style={emptyRelationshipStyle}>No items</span>
            )}
          </div>
          <input
            aria-label={`${label} ${row.recordId ?? "draft row"}`}
            data-testid={
              row.recordId === null
                ? draftCellTestId(`${binding.draftKey}-input`)
                : rowCellTestId(row.recordId, `${binding.draftKey}-input`)
            }
            key={`${row.key}:${binding.draftKey}:${row.rowVersion ?? "draft"}`}
            ref={(element) => {
              registerInput(row.key, binding.draftKey, element);
            }}
            style={inputStyle}
            type="text"
            defaultValue={row.collectionDrafts[binding.draftKey]}
            onBlur={(event) => {
              queueCollectionSave(
                row.key,
                binding.fieldKey,
                binding.draftKey,
                event.currentTarget.value,
              );
            }}
            onFocus={() => {
              if (row.recordId) {
                handleSelectRow(row.recordId);
              }
            }}
            onKeyDown={(event) => {
              handleCollectionKeyDown(
                event,
                row.key,
                binding.fieldKey,
                binding.draftKey,
              );
            }}
            placeholder={`Add ${label.toLowerCase()} token`}
          />
        </>
      );
    },
    [
      entityIndex,
      handleCollectionKeyDown,
      handleSelectRow,
      handleSelectMention,
      registerInput,
      timelineBindingLabel,
      queueCollectionSave,
    ],
  );

  const timelineColumns = useMemo<readonly GridColumn<WorkbookRow>[]>(
    () => [
      {
        fieldKey: "timeline.capture_state",
        headerTestId: gridSortHeaderTestId(
          "timeline",
          "timeline.capture_state",
        ),
        label: "State",
        width: 136,
        renderCell: (row) => (
          <span
            data-testid={
              row.recordId === null
                ? "draft-row-capture-state"
                : rowCellTestId(row.recordId, "capture-state")
            }
          >
            {row.captureState}
          </span>
        ),
        sortableFieldKey: "timeline.capture_state",
      },
      {
        fieldKey: "row_version",
        label: "Version",
        width: 96,
        renderCell: (row) => (
          <span
            data-testid={
              row.recordId === null
                ? "draft-row-version"
                : rowCellTestId(row.recordId, "row-version")
            }
          >
            {row.rowVersion ?? "new"}
          </span>
        ),
      },
      ...timelineVisibleBindings.map(
        (binding): GridColumn<WorkbookRow> => ({
          fieldKey: binding.fieldKey,
          headerTestId: gridSortHeaderTestId("timeline", binding.fieldKey),
          label: timelineBindingLabel(binding.fieldKey),
          width: timelineColumnWidth(binding.fieldKey),
          renderCell: (row) => {
            if (binding.kind === "scalar") {
              return renderTimelineGridEditor(row, binding);
            }
            if (binding.kind === "collection") {
              return renderTimelineCollectionInput(row, binding);
            }
            const text = stringifyGridValue(
              readCellValue(row.rawRow, binding.fieldKey),
            );
            return <span style={bodyStyle}>{text === "" ? "—" : text}</span>;
          },
          sortableFieldKey: resolveHeaderSortFieldKey(
            timelineContract,
            binding.fieldKey,
          ),
        }),
      ),
    ],
    [
      renderTimelineCollectionInput,
      renderTimelineGridEditor,
      timelineBindingLabel,
    ],
  );

  const timelineActionsColumn = useMemo<GridActionsColumn<WorkbookRow>>(
    () => ({
      label: "Actions",
      width: 176,
      renderCell: ({ data: row }) =>
        row.recordId === null ? (
          <div style={actionStackStyle}>
            <DraftRowCreateButton
              onCreate={handleCreateBlankDraftRow}
              row={row}
            />
          </div>
        ) : (
          <div style={actionStackStyle}>
            <button
              data-testid={rowInspectButtonTestId(row.recordId)}
              style={actionButtonStyle}
              type="button"
              onClick={() => {
                handleSelectRow(row.recordId ?? "");
              }}
            >
              Inspect
            </button>
            <button
              data-testid={`row-${row.recordId}-mark-reviewed`}
              disabled={
                row.captureState === "reviewed" ||
                row.captureState === "superseded"
              }
              style={actionButtonStyle}
              type="button"
              onClick={() => {
                queueAction(row.key, "mark-reviewed");
              }}
            >
              Mark reviewed
            </button>
            <input
              data-testid={`row-${row.recordId}-replacement-id`}
              placeholder="Replacement record id"
              style={replacementInputStyle}
              type="text"
              value={replacementDrafts[row.key] ?? ""}
              onChange={(event) => {
                const value = event.target.value;
                setReplacementDrafts((current) => ({
                  ...current,
                  [row.key]: value,
                }));
              }}
            />
            <button
              data-testid={`row-${row.recordId}-supersede`}
              disabled={
                row.captureState === "superseded" ||
                normalizeValue(replacementDrafts[row.key] ?? "") === ""
              }
              style={actionButtonStyle}
              type="button"
              onClick={() => {
                queueAction(row.key, "supersede");
              }}
            >
              Supersede
            </button>
          </div>
        ),
    }),
    [
      handleCreateBlankDraftRow,
      handleSelectRow,
      queueAction,
      replacementDrafts,
    ],
  );

  const timelineGridRows = useMemo<readonly GridRow<WorkbookRow>[]>(
    () =>
      rows.map((row) => ({
        key: row.key,
        recordId: row.recordId,
        data: row,
        onSelect: () => {
          if (row.recordId) {
            handleSelectRow(row.recordId);
          }
        },
        selected: row.recordId !== null && row.recordId === selectedRowId,
        testId:
          row.recordId === null ? undefined : `timeline-row-${row.recordId}`,
        variant: row.recordId === null ? "draft" : "default",
      })),
    [handleSelectRow, rows, selectedRowId],
  );

  const getTimelineGroupLabel = useCallback(
    (row: WorkbookRow, fieldKey: string) => timelineGroupLabel(row, fieldKey),
    [],
  );
  const getTimelineGroupRowTestId = useCallback(
    (fieldKey: string, value: string) =>
      gridGroupRowTestId("timeline", fieldKey, value),
    [],
  );

  function renderInspectorFieldEditors(row: WorkbookRow) {
    return (
      <section style={inspectorSectionStyle}>
        <h3 style={sectionTitleStyle}>Details</h3>
        <div style={inspectorActionStackStyle}>
          {timelineInspectorBindings.map((binding) =>
            renderTimelineInspectorEditor(row, binding),
          )}
        </div>
      </section>
    );
  }

  if (isLoading) {
    return (
      <section style={panelStyle}>
        <p style={eyebrowStyle}>Timeline</p>
        <h1 style={headlineStyle}>Loading projection-backed rows.</h1>
      </section>
    );
  }

  if (loadError !== null) {
    return (
      <section style={panelStyle}>
        <p style={eyebrowStyle}>Timeline</p>
        <h1 style={headlineStyle}>Timeline load failed.</h1>
        <p style={bodyStyle}>{loadError}</p>
      </section>
    );
  }

  return (
    <section style={workbookStyle}>
      <header style={headerStyle}>
        <button
          aria-label="Blur timeline inputs"
          data-testid="timeline-blur-surface"
          tabIndex={-1}
          type="button"
          style={blurSurfaceButtonStyle}
          onMouseDown={(event) => {
            event.currentTarget.focus();
          }}
        />
        <div>
          <p style={eyebrowStyle}>Phase 3 Workbook</p>
          <h1 style={headlineStyle}>Timeline mutation substrate</h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
        <div style={statusClusterStyle}>
          <span style={statusLabelStyle}>Save State</span>
          <strong
            aria-live="polite"
            data-testid="save-state"
            style={{
              ...statusValueStyle,
              color:
                saveState === "Conflict"
                  ? "rgb(145 30 30)"
                  : saveState === "Syncing"
                    ? "rgb(146 64 14)"
                    : "rgb(21 128 61)",
            }}
          >
            {saveState}
          </strong>
        </div>
      </header>

      {autoResolutionNotices.length > 0 ? (
        <aside style={noticeStackStyle}>
          {autoResolutionNotices.map((notice) => (
            <div
              key={notice.itemRef}
              data-testid={`auto-resolution-notice-${sanitizeTestId(notice.itemRef)}`}
              style={noticeCardStyle}
            >
              <p style={noticeTitleStyle}>Auto-resolved mention</p>
              <p style={bodyStyle}>
                Raw token <strong>{notice.rawText}</strong> matched{" "}
                <strong>
                  {entityIndex[notice.resolvedRecordId]?.label ??
                    notice.rawText}
                </strong>
                {notice.matchedAliasText ? (
                  <>
                    {" "}
                    via alias <strong>{notice.matchedAliasText}</strong>
                  </>
                ) : null}
                .
              </p>
              <div style={inlineButtonRowStyle}>
                <button
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    const mention = buildInspectorMentions(
                      rows.find((row) => row.recordId === notice.rowRecordId) ??
                        undefined,
                      dismissedMentionsByRow[notice.rowRecordId] ?? [],
                    ).find((item) => item.itemRef === notice.itemRef);
                    if (mention) {
                      submitMentionAction(mention, "revert_to_unresolved");
                    }
                    setAutoResolutionNotices((current) =>
                      current.filter((item) => item.itemRef !== notice.itemRef),
                    );
                  }}
                >
                  Undo
                </button>
                <button
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    handleSelectMention(notice.rowRecordId, notice.itemRef);
                    setAutoResolutionNotices((current) =>
                      current.filter((item) => item.itemRef !== notice.itemRef),
                    );
                  }}
                >
                  Review
                </button>
              </div>
            </div>
          ))}
        </aside>
      ) : null}

      <div style={splitShellStyle}>
        <div>
          <WorkbookGridControls
            contract={timelineContract}
            filterDraft={filterDraft}
            onApplyFilter={applyQueryFilter}
            onFilterDraftChange={setFilterDraft}
            onGroupByChange={handleQueryGroupByChange}
            onRemoveFilter={(fieldKey) => {
              setQueryState((current) => removeFilterField(current, fieldKey));
            }}
            queryState={queryState}
            surface="timeline"
          />
          <GridViewport
            ref={gridShellRef}
            style={gridShellStyle}
            testId={gridShellTestId("timeline")}
          >
            <GridTable
              actionsColumn={timelineActionsColumn}
              columns={timelineColumns}
              getGroupLabel={getTimelineGroupLabel}
              getGroupRowTestId={getTimelineGroupRowTestId}
              groupBy={queryState.groupBy}
              onToggleSort={handleQuerySortToggle}
              rows={timelineGridRows}
              sort={queryState.sort}
            />
          </GridViewport>
        </div>

        <aside data-testid="timeline-inspector" style={inspectorShellStyle}>
          <div style={inspectorHeaderStyle}>
            <p style={eyebrowStyle}>Inspector</p>
            <h2 style={inspectorTitleStyle}>
              {selectedRow?.recordId
                ? `Timeline row ${selectedRow.recordId}`
                : draftRow
                  ? "Draft timeline row"
                  : "Select a saved row"}
            </h2>
            <p style={bodyStyle}>
              Routine mention review and hidden-field editing stay on the
              workbook surface.
            </p>
          </div>
          {selectedRow?.recordId ? (
            <>
              {renderInspectorFieldEditors(selectedRow)}
              <section style={inspectorSectionStyle}>
                <h3 style={sectionTitleStyle}>Mentions</h3>
                <div style={mentionGroupStyle}>
                  {["unresolved", "resolved", "dismissed"].map((status) => {
                    const group = inspectorMentions.filter(
                      (item) => item.status === status,
                    );
                    return (
                      <div key={status} style={mentionGroupColumnStyle}>
                        <p style={groupLabelStyle}>
                          {status === "dismissed"
                            ? "Dismissed"
                            : status === "resolved"
                              ? "Resolved"
                              : "Unresolved"}
                        </p>
                        {group.length > 0 ? (
                          group.map((item) => (
                            <button
                              key={item.itemRef}
                              data-testid={`mention-${sanitizeTestId(item.itemRef)}`}
                              style={{
                                ...mentionListButtonStyle,
                                ...(selectedMention?.itemRef === item.itemRef
                                  ? mentionListButtonSelectedStyle
                                  : null),
                              }}
                              type="button"
                              onClick={() => {
                                handleSelectMention(
                                  item.rowRecordId,
                                  item.itemRef,
                                );
                              }}
                            >
                              <RelationshipChip
                                entityIndex={entityIndex}
                                item={item}
                                selected={
                                  selectedMention?.itemRef === item.itemRef
                                }
                              />
                            </button>
                          ))
                        ) : (
                          <span style={emptyRelationshipStyle}>None</span>
                        )}
                      </div>
                    );
                  })}
                </div>
              </section>

              {selectedMention ? (
                <section style={inspectorSectionStyle}>
                  <h3 style={sectionTitleStyle}>Selected mention</h3>
                  <dl style={detailListStyle}>
                    <div>
                      <dt style={detailTermStyle}>Raw token</dt>
                      <dd style={detailValueStyle}>
                        {selectedMention.rawText}
                      </dd>
                    </div>
                    <div>
                      <dt style={detailTermStyle}>Field</dt>
                      <dd style={detailValueStyle}>
                        {timelineRelationshipLabel(selectedMention.fieldKey)}
                      </dd>
                    </div>
                    <div>
                      <dt style={detailTermStyle}>Status</dt>
                      <dd style={detailValueStyle}>{selectedMention.status}</dd>
                    </div>
                    <div>
                      <dt style={detailTermStyle}>Target</dt>
                      <dd style={detailValueStyle}>
                        {selectedMention.resolvedRecordId
                          ? relationshipItemLabel(selectedMention, entityIndex)
                          : "None"}
                      </dd>
                    </div>
                  </dl>

                  {selectedMention.status === "unresolved" ? (
                    <div style={inspectorActionStackStyle}>
                      <label style={labelStyle}>
                        Resolve to existing
                        <select
                          data-testid="inspector-resolve-target"
                          style={selectStyle}
                          value={selectedResolveTargetId}
                          onChange={(event) => {
                            const value = event.target.value;
                            setSelectedResolveTargetId(value);
                            if (value !== "") {
                              setInspectorMessage(`Selected ${value}`);
                            }
                          }}
                        >
                          <option value="">Select target</option>
                          {(selectedMention.entityType === "host"
                            ? hostEntities
                            : identityEntities
                          ).map((entity) => (
                            <option
                              key={entity.recordId}
                              value={entity.recordId}
                            >
                              {entity.label}
                            </option>
                          ))}
                        </select>
                      </label>
                      <div style={inlineButtonRowStyle}>
                        <button
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            if (selectedResolveTargetId === "") {
                              setInspectorMessage("Select a target first.");
                              return;
                            }
                            submitMentionAction(
                              selectedMention,
                              "resolve_item",
                              selectedResolveTargetId,
                            );
                          }}
                        >
                          Resolve to existing
                        </button>
                        <button
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            submitMentionAction(
                              selectedMention,
                              "resolve_item",
                            );
                          }}
                        >
                          {selectedMention.entityType === "host"
                            ? "Create host"
                            : "Create identity"}
                        </button>
                      </div>
                    </div>
                  ) : null}

                  {selectedMention.status === "resolved" ? (
                    <div style={inlineButtonRowStyle}>
                      {canManageMentions ? (
                        <button
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            submitMentionAction(
                              selectedMention,
                              "dismiss_item",
                            );
                          }}
                        >
                          Dismiss
                        </button>
                      ) : null}
                      {canManageMentions ? (
                        <button
                          style={secondaryActionButtonStyle}
                          type="button"
                          onClick={() => {
                            submitMentionAction(
                              selectedMention,
                              "revert_to_unresolved",
                            );
                          }}
                        >
                          Revert to unresolved
                        </button>
                      ) : null}
                    </div>
                  ) : null}

                  {selectedMention.status === "dismissed" ? (
                    <div style={inlineButtonRowStyle}>
                      <button
                        style={secondaryActionButtonStyle}
                        type="button"
                        onClick={() => {
                          submitMentionAction(
                            selectedMention,
                            "revert_to_unresolved",
                          );
                        }}
                      >
                        Restore to unresolved
                      </button>
                    </div>
                  ) : null}
                </section>
              ) : null}
              {inspectorMessage ? (
                <p data-testid="timeline-inspector-message" style={bodyStyle}>
                  {inspectorMessage}
                </p>
              ) : null}
            </>
          ) : (
            <>
              {draftRow ? renderInspectorFieldEditors(draftRow) : null}
              <p style={bodyStyle}>
                Pick a saved row to inspect unresolved, resolved, and dismissed
                mentions.
              </p>
            </>
          )}
        </aside>
      </div>
    </section>
  );
}

function EntityWorkbookSurface({
  incidentId,
  apiBase,
  entityType,
  filterDraft,
  onApplyFilter,
  onFilterDraftChange,
  onGroupByChange,
  onRemoveFilter,
  rows,
  onToggleSort,
  queryState,
  currentIncidentRole,
  entityIndex,
  onRefreshEntities,
}: {
  incidentId: string;
  apiBase?: string | undefined;
  entityType: EntityRow["entityType"];
  filterDraft: FilterDraft;
  onApplyFilter: () => void;
  onFilterDraftChange: (draft: FilterDraft) => void;
  onGroupByChange: (groupBy: string | null) => void;
  onRemoveFilter: (fieldKey: string) => void;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
  rows: EntityRow[];
  currentIncidentRole: IncidentRole | null;
  entityIndex: Record<string, EntityRow>;
  onRefreshEntities: () => Promise<void>;
}) {
  const [selectedRecordId, setSelectedRecordId] = useState<string | null>(null);
  const [mergeCandidateId, setMergeCandidateId] = useState<string>("");
  const [mergeReason, setMergeReason] = useState("Merge duplicate entity");
  const [mergeMessage, setMergeMessage] = useState<string | null>(null);
  const [timelinePreviewRows, setTimelinePreviewRows] = useState<WorkbookRow[]>(
    [],
  );

  const selectedEntity =
    rows.find((row) => row.recordId === selectedRecordId) ?? rows[0] ?? null;
  const canMerge =
    currentIncidentRole === "reviewer" || currentIncidentRole === "admin";
  const survivorLabel = selectedEntity?.label ?? "Select a record";
  const surface: WorkbookSurface =
    entityType === "host" ? "hosts" : "identities";
  const contract = entityType === "host" ? hostsContract : identitiesContract;
  const loserEntity =
    rows.find((row) => row.recordId === mergeCandidateId) ?? null;
  const mergePlan =
    selectedEntity && loserEntity
      ? buildMergePlan(selectedEntity, loserEntity)
      : null;
  const entityColumns: readonly GridColumn<EntityRow>[] = [
    {
      fieldKey:
        entityType === "host" ? "host.display_name" : "identity.display_name",
      headerTestId: gridSortHeaderTestId(
        surface,
        entityType === "host" ? "host.display_name" : "identity.display_name",
      ),
      label:
        contract.fieldMap[
          entityType === "host" ? "host.display_name" : "identity.display_name"
        ]?.label ?? "Name",
      width: 240,
      renderCell: (row) => row.label,
      sortableFieldKey:
        entityType === "host" ? "host.display_name" : "identity.display_name",
    },
    {
      fieldKey: entityType === "host" ? "host.hostname" : "identity.upn",
      headerTestId: gridSortHeaderTestId(
        surface,
        entityType === "host" ? "host.hostname" : "identity.upn",
      ),
      label:
        contract.fieldMap[
          entityType === "host" ? "host.hostname" : "identity.upn"
        ]?.label ?? "Primary",
      width: 260,
      renderCell: (row) => row.secondaryText || "None",
      sortableFieldKey:
        entityType === "host" ? "host.hostname" : "identity.upn",
    },
    {
      fieldKey: entityType === "host" ? "host.aliases" : "identity.aliases",
      label:
        contract.fieldMap[
          entityType === "host" ? "host.aliases" : "identity.aliases"
        ]?.label ?? "Aliases",
      width: 320,
      renderCell: (row) => (
        <div style={relationshipItemsWrapStyle}>
          {row.aliasTexts.length > 0 ? (
            row.aliasTexts.map((alias) => (
              <span key={alias} style={aliasChipStyle}>
                {alias}
              </span>
            ))
          ) : (
            <span style={emptyRelationshipStyle}>No aliases</span>
          )}
        </div>
      ),
    },
    {
      fieldKey:
        entityType === "host" ? "host.host_state" : "identity.identity_state",
      headerTestId: gridSortHeaderTestId(
        surface,
        entityType === "host" ? "host.host_state" : "identity.identity_state",
      ),
      label:
        contract.fieldMap[
          entityType === "host" ? "host.host_state" : "identity.identity_state"
        ]?.label ?? "State",
      width: 140,
      renderCell: (row) => row.state,
      sortableFieldKey:
        entityType === "host" ? "host.host_state" : "identity.identity_state",
    },
    {
      fieldKey: "row_version",
      label: "Version",
      width: 96,
      renderCell: (row) => row.rowVersion,
    },
  ];
  const entityActionsColumn: GridActionsColumn<EntityRow> = {
    label: "Actions",
    width: 176,
    renderCell: ({ data: row }) => (
      <button
        data-testid={`inspect-${entityType}-${row.recordId}`}
        style={actionButtonStyle}
        type="button"
        onClick={() => {
          setSelectedRecordId(row.recordId);
          setMergeMessage(null);
        }}
      >
        Inspect
      </button>
    ),
  };
  const entityGridRows: readonly GridRow<EntityRow>[] = rows.map((row) => ({
    key: row.recordId,
    recordId: row.recordId,
    data: row,
    selected: row.recordId === selectedEntity?.recordId,
    testId: `${entityType}-row-${row.recordId}`,
  }));

  useEffect(() => {
    if (selectedEntity) {
      setSelectedRecordId(selectedEntity.recordId);
      return;
    }
    setSelectedRecordId(null);
  }, [selectedEntity]);

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
      const draftKey = entityType === "host" ? "hostRefs" : "identityRefs";
      const previewRows = envelope.data.rows
        .map(rowFromApi)
        .filter((row) =>
          row.collectionValues[draftKey].some(
            (item) => item.resolvedRecordId === recordId,
          ),
        );
      setTimelinePreviewRows(previewRows);
    },
    [apiBase, entityType, incidentId],
  );

  async function confirmMerge() {
    if (!selectedEntity || !loserEntity) {
      return;
    }
    setMergeMessage(null);
    const result = await fetchJSON<MergeEnvelope>(
      apiPath(apiBase, `/api/v1/records/${selectedEntity.recordId}/merge`),
      {
        method: "POST",
        body: JSON.stringify({
          loser_record_id: loserEntity.recordId,
          survivor_base_row_version: selectedEntity.rowVersion,
          loser_base_row_version: loserEntity.rowVersion,
          client_txn_id: `merge-${Date.now()}`,
          reason: mergeReason,
        }),
      },
    );
    if (!result.ok) {
      setMergeMessage(parseErrorMessage(result.payload));
      return;
    }

    const envelope = readEnvelope<MergeEnvelope>(result.payload);
    setMergeMessage(
      `Merged ${loserEntity.label} into ${selectedEntity.label} (${envelope.data.merge_summary.record_type}).`,
    );
    await onRefreshEntities();
    await loadTimelinePreview(selectedEntity.recordId);
    setSelectedRecordId(selectedEntity.recordId);
    setMergeCandidateId("");
  }

  return (
    <section style={workbookStyle}>
      <header style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>
            {entityType === "host" ? "Hosts" : "Identities"}
          </p>
          <h1 style={headlineStyle}>
            {entityType === "host" ? "Hosts surface" : "Identities surface"}
          </h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
      </header>

      <div style={splitShellStyle}>
        <div>
          <WorkbookGridControls
            contract={contract}
            filterDraft={filterDraft}
            onApplyFilter={onApplyFilter}
            onFilterDraftChange={onFilterDraftChange}
            onGroupByChange={onGroupByChange}
            onRemoveFilter={onRemoveFilter}
            queryState={queryState}
            surface={surface}
          />
          <GridViewport
            style={gridShellStyle}
            testId={gridShellTestId(surface)}
          >
            <GridTable
              actionsColumn={entityActionsColumn}
              columns={entityColumns}
              getGroupLabel={(row, fieldKey) => entityGroupLabel(row, fieldKey)}
              getGroupRowTestId={(fieldKey, value) =>
                gridGroupRowTestId(surface, fieldKey, value)
              }
              groupBy={queryState.groupBy}
              onToggleSort={onToggleSort}
              rows={entityGridRows}
              sort={queryState.sort}
            />
          </GridViewport>
        </div>

        <aside
          data-testid={`${entityType}-inspector`}
          style={inspectorShellStyle}
        >
          <div style={inspectorHeaderStyle}>
            <p style={eyebrowStyle}>Inspector</p>
            <h2 style={inspectorTitleStyle}>{survivorLabel}</h2>
            <p style={bodyStyle}>
              Merge review stays inside the workbook shell.
            </p>
          </div>
          {selectedEntity ? (
            <>
              <section style={inspectorSectionStyle}>
                <h3 style={sectionTitleStyle}>Identifiers</h3>
                <ul style={flatListStyle}>
                  {selectedEntity.identifiers.length > 0 ? (
                    selectedEntity.identifiers.map((identifier) => (
                      <li key={identifier.key}>
                        {identifier.label}: {identifier.value}
                      </li>
                    ))
                  ) : (
                    <li>No exact-match identifiers visible.</li>
                  )}
                </ul>
              </section>

              {canMerge ? (
                <section style={inspectorSectionStyle}>
                  <h3 style={sectionTitleStyle}>Merge</h3>
                  <label style={labelStyle}>
                    Merge loser
                    <select
                      data-testid="merge-loser-record"
                      style={selectStyle}
                      value={mergeCandidateId}
                      onChange={(event) => {
                        setMergeCandidateId(event.target.value);
                        setMergeMessage(null);
                      }}
                    >
                      <option value="">Select duplicate</option>
                      {rows
                        .filter(
                          (row) => row.recordId !== selectedEntity.recordId,
                        )
                        .map((row) => (
                          <option key={row.recordId} value={row.recordId}>
                            {row.label}
                          </option>
                        ))}
                    </select>
                  </label>
                  <label style={labelStyle}>
                    Merge reason
                    <input
                      data-testid="merge-reason"
                      style={inputStyle}
                      type="text"
                      value={mergeReason}
                      onChange={(event) => {
                        setMergeReason(event.target.value);
                      }}
                    />
                  </label>
                  {loserEntity && mergePlan ? (
                    <div data-testid="merge-plan" style={mergePlanStyle}>
                      <p style={noticeTitleStyle}>
                        Survivor {selectedEntity.label} absorbs loser{" "}
                        {loserEntity.label}
                      </p>
                      <p style={bodyStyle}>
                        Survivor record {selectedEntity.recordId}
                        <br />
                        Loser record {loserEntity.recordId}
                      </p>
                      <ul style={flatListStyle}>
                        {mergePlan.identifierLines.map((line) => (
                          <li key={`${line.label}:${line.outcome}`}>
                            {line.label}: {line.outcome}
                          </li>
                        ))}
                        <li>
                          Aliases to copy:{" "}
                          {mergePlan.aliasesToCopy.length > 0
                            ? mergePlan.aliasesToCopy.join(", ")
                            : "none"}
                        </li>
                        <li>
                          Alias duplicate no-op:{" "}
                          {mergePlan.duplicateAliases.length > 0
                            ? mergePlan.duplicateAliases.join(", ")
                            : "none"}
                        </li>
                        <li>
                          Provenance-only values:{" "}
                          {mergePlan.provenanceOnlySummary}
                        </li>
                        <li>{mergePlan.dependencySummary}</li>
                      </ul>
                      <button
                        data-testid="merge-confirm"
                        style={secondaryActionButtonStyle}
                        type="button"
                        onClick={() => {
                          void confirmMerge();
                        }}
                      >
                        Confirm merge
                      </button>
                    </div>
                  ) : (
                    <button
                      data-testid="merge-start"
                      style={secondaryActionButtonStyle}
                      type="button"
                      onClick={() => {
                        setMergeMessage(
                          "Select a loser to review the merge plan.",
                        );
                      }}
                    >
                      Start merge
                    </button>
                  )}
                </section>
              ) : (
                <section style={inspectorSectionStyle}>
                  <h3 style={sectionTitleStyle}>Merge</h3>
                  <p style={bodyStyle}>
                    Merge is available to reviewer or admin roles.
                  </p>
                </section>
              )}

              {timelinePreviewRows.length > 0 ? (
                <section style={inspectorSectionStyle}>
                  <h3 style={sectionTitleStyle}>Dependent Timeline</h3>
                  <div style={timelinePreviewStackStyle}>
                    {timelinePreviewRows.map((row) => (
                      <article
                        key={row.recordId ?? row.key}
                        data-testid={`timeline-preview-row-${row.recordId ?? row.key}`}
                        style={timelinePreviewCardStyle}
                      >
                        <p style={noticeTitleStyle}>
                          {row.values.summary || "Untitled row"}
                        </p>
                        <div style={relationshipItemsWrapStyle}>
                          {row.collectionValues[
                            entityType === "host" ? "hostRefs" : "identityRefs"
                          ].map((item) => (
                            <RelationshipChip
                              key={item.itemRef}
                              entityIndex={entityIndex}
                              item={item}
                            />
                          ))}
                        </div>
                      </article>
                    ))}
                  </div>
                </section>
              ) : null}

              {mergeMessage ? (
                <p data-testid="merge-message" style={bodyStyle}>
                  {mergeMessage}
                </p>
              ) : null}
            </>
          ) : (
            <p style={bodyStyle}>No active records on this surface.</p>
          )}
        </aside>
      </div>
    </section>
  );
}

function AssessmentWorkbookSurface({
  apiBase,
  assessmentRows,
  currentIncidentRole,
  filterDraft,
  hostRows,
  identityRows,
  incidentId,
  loadError,
  onApplyFilter,
  onFilterDraftChange,
  onGroupByChange,
  onRefreshAssessmentRows,
  onRemoveFilter,
  onToggleSort,
  queryState,
}: {
  apiBase?: string | undefined;
  assessmentRows: EntityApiRow[];
  currentIncidentRole: IncidentRole | null;
  filterDraft: FilterDraft;
  hostRows: EntityRow[];
  identityRows: EntityRow[];
  incidentId: string;
  loadError: string | null;
  onApplyFilter: () => void;
  onFilterDraftChange: (draft: FilterDraft) => void;
  onGroupByChange: (groupBy: string | null) => void;
  onRefreshAssessmentRows: () => Promise<void>;
  onRemoveFilter: (fieldKey: string) => void;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
}) {
  const [draft, setDraft] = useState<AssessmentCreateDraft>(() =>
    initialAssessmentDraft(),
  );
  const [supportRows, setSupportRows] = useState<TimelineApiRow[]>([]);
  const [message, setMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const subjectRows = draft.subjectType === "host" ? hostRows : identityRows;
  const canCreate =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const stateOptions = enumValuesFor(
    assessmentsContract,
    "assessment.assessment_state",
    ["unknown", "suspected", "confirmed", "disproven", "cleared"],
  );
  const confidenceBandOptions = enumValuesFor(
    assessmentsContract,
    "assessment.confidence_band",
    ["unset", "low", "medium", "high"],
  ).filter(isAssessmentConfidenceBand);
  const columns: readonly GridColumn<EntityApiRow>[] = visibleFields(
    assessmentsContract,
  ).map((field) => ({
    fieldKey: field.fieldKey,
    headerTestId: gridSortHeaderTestId(assessmentsViewSchemaId, field.fieldKey),
    label: field.label,
    width: assessmentColumnWidth(field.fieldKey),
    renderCell: (row) => (
      <span data-testid={rowCellTestId(row.record_id, field.fieldKey)}>
        {genericCellLabel(row.cells[field.fieldKey]?.value)}
      </span>
    ),
    sortableFieldKey: resolveHeaderSortFieldKey(
      assessmentsContract,
      field.fieldKey,
    ),
  }));
  const gridRows: readonly GridRow<EntityApiRow>[] = assessmentRows.map(
    (row) => ({
      key: row.record_id,
      recordId: row.record_id,
      data: row,
      testId: `assessment-row-${row.record_id}`,
    }),
  );

  useEffect(() => {
    setDraft((current) => {
      if (
        current.subjectRecordId !== "" &&
        subjectRows.some((row) => row.recordId === current.subjectRecordId)
      ) {
        return current;
      }
      return {
        ...current,
        subjectRecordId: subjectRows[0]?.recordId ?? "",
      };
    });
  }, [subjectRows]);

  useEffect(() => {
    let isCurrent = true;
    async function loadSupportRows() {
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
      if (!isCurrent) {
        return;
      }
      if (!result.ok) {
        setSupportRows([]);
        return;
      }
      const envelope = readEnvelope<WorkbookQueryEnvelope>(result.payload);
      setSupportRows(envelope.data.rows);
    }
    void loadSupportRows();
    return () => {
      isCurrent = false;
    };
  }, [apiBase, incidentId]);

  async function submitAssessment() {
    const payload = buildAssessmentCreatePayload(
      draft,
      `assessment-${Date.now()}`,
    );
    if (payload === null) {
      setMessage("Subject, state, and rationale are required.");
      return;
    }

    setIsSubmitting(true);
    setMessage(null);
    try {
      const result = await fetchJSON<TimelineMutationEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${assessmentsViewSchemaId}/rows`,
        ),
        {
          method: "POST",
          body: JSON.stringify(payload),
        },
      );
      if (!result.ok) {
        setMessage(parseErrorMessage(result.payload));
        return;
      }
      await onRefreshAssessmentRows();
      setDraft((current) => ({
        ...initialAssessmentDraft(),
        subjectType: current.subjectType,
        subjectRecordId: current.subjectRecordId,
      }));
      setMessage("Assessment created.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section style={workbookStyle}>
      <header style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>System view</p>
          <h1 style={headlineStyle}>{assessmentsContract.title}</h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
      </header>

      <div style={splitShellStyle}>
        <div>
          <WorkbookGridControls
            contract={assessmentsContract}
            filterDraft={filterDraft}
            onApplyFilter={onApplyFilter}
            onFilterDraftChange={onFilterDraftChange}
            onGroupByChange={onGroupByChange}
            onRemoveFilter={onRemoveFilter}
            queryState={queryState}
            surface={assessmentsViewSchemaId}
          />
          {loadError ? (
            <p data-testid="assessment-surface-load-error" style={bodyStyle}>
              {loadError}
            </p>
          ) : null}
          <GridViewport
            style={gridShellStyle}
            testId={gridShellTestId(assessmentsViewSchemaId)}
          >
            <GridTable
              columns={columns}
              getGroupLabel={(row, fieldKey) =>
                genericCellLabel(row.cells[fieldKey]?.value)
              }
              getGroupRowTestId={(fieldKey, value) =>
                gridGroupRowTestId(assessmentsViewSchemaId, fieldKey, value)
              }
              groupBy={queryState.groupBy}
              onToggleSort={onToggleSort}
              rows={gridRows}
              sort={queryState.sort}
            />
          </GridViewport>
        </div>

        <aside
          data-testid="assessment-create-panel"
          style={inspectorShellStyle}
        >
          <div style={inspectorHeaderStyle}>
            <p style={eyebrowStyle}>Create</p>
            <h2 style={inspectorTitleStyle}>Append assessment</h2>
          </div>
          <div style={inspectorSectionStyle}>
            <label style={labelStyle}>
              Subject type
              <select
                data-testid="assessment-create-subject-type"
                style={selectStyle}
                value={draft.subjectType}
                onChange={(event) => {
                  const subjectType =
                    event.target.value === "identity" ? "identity" : "host";
                  const nextRows =
                    subjectType === "host" ? hostRows : identityRows;
                  setDraft((current) => ({
                    ...current,
                    subjectType,
                    subjectRecordId: nextRows[0]?.recordId ?? "",
                  }));
                }}
              >
                {enumValuesFor(assessmentsContract, "assessment.subject_type", [
                  "host",
                  "identity",
                ]).map((value) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </select>
            </label>

            <label style={labelStyle}>
              Subject
              <select
                data-testid="assessment-create-subject"
                style={selectStyle}
                value={draft.subjectRecordId}
                onChange={(event) => {
                  setDraft((current) => ({
                    ...current,
                    subjectRecordId: event.target.value,
                  }));
                }}
              >
                <option value="">Select subject</option>
                {subjectRows.map((row) => (
                  <option key={row.recordId} value={row.recordId}>
                    {row.label}
                  </option>
                ))}
              </select>
            </label>

            <label style={labelStyle}>
              State
              <select
                data-testid="assessment-create-state"
                style={selectStyle}
                value={draft.assessmentState}
                onChange={(event) => {
                  setDraft((current) => ({
                    ...current,
                    assessmentState: event.target.value,
                  }));
                }}
              >
                {stateOptions.map((value) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </select>
            </label>

            <label style={labelStyle}>
              Confidence
              <select
                data-testid="assessment-create-confidence-band"
                style={selectStyle}
                value={draft.confidenceBand}
                onChange={(event) => {
                  const confidenceBand = isAssessmentConfidenceBand(
                    event.target.value,
                  )
                    ? event.target.value
                    : "unset";
                  setDraft((current) => ({
                    ...current,
                    confidenceBand,
                  }));
                }}
              >
                {confidenceBandOptions.map((value) => (
                  <option key={value} value={value}>
                    {value}
                  </option>
                ))}
              </select>
            </label>

            <label style={labelStyle}>
              Rationale
              <textarea
                data-testid="assessment-create-rationale"
                rows={4}
                style={textareaStyle}
                value={draft.rationale}
                onChange={(event) => {
                  setDraft((current) => ({
                    ...current,
                    rationale: event.target.value,
                  }));
                }}
              />
            </label>

            <label style={labelStyle}>
              Assessed
              <input
                data-testid="assessment-create-assessed-at"
                placeholder="RFC3339 timestamp"
                style={inputStyle}
                type="text"
                value={draft.assessedAt}
                onChange={(event) => {
                  setDraft((current) => ({
                    ...current,
                    assessedAt: event.target.value,
                  }));
                }}
              />
            </label>

            <label style={labelStyle}>
              Support refs
              <select
                data-testid="assessment-create-support-refs"
                multiple
                size={Math.min(Math.max(supportRows.length, 2), 5)}
                style={selectStyle}
                value={draft.supportRecordIds}
                onChange={(event) => {
                  const supportRecordIds = Array.from(
                    event.currentTarget.selectedOptions,
                  ).map((option) => option.value);
                  setDraft((current) => ({
                    ...current,
                    supportRecordIds,
                  }));
                }}
              >
                {supportRows.map((row) => (
                  <option key={row.record_id} value={row.record_id}>
                    {supportRowLabel(row)}
                  </option>
                ))}
              </select>
            </label>

            <button
              data-testid="assessment-create-submit"
              disabled={!canCreate || isSubmitting}
              style={secondaryActionButtonStyle}
              type="button"
              onClick={() => {
                void submitAssessment();
              }}
            >
              Create assessment
            </button>
            {message ? (
              <p data-testid="assessment-create-message" style={bodyStyle}>
                {message}
              </p>
            ) : null}
          </div>
        </aside>
      </div>
    </section>
  );
}

function GenericWorkbookSurface({
  apiBase,
  contract,
  filterDraft,
  incidentId,
  loadError,
  onApplyFilter,
  onFilterDraftChange,
  onGroupByChange,
  onRemoveFilter,
  onRefresh,
  onToggleSort,
  queryState,
  rows,
}: {
  apiBase?: string | undefined;
  contract: ViewContract;
  filterDraft: FilterDraft;
  incidentId: string;
  loadError: string | null;
  onApplyFilter: () => void;
  onFilterDraftChange: (draft: FilterDraft) => void;
  onGroupByChange: (groupBy: string | null) => void;
  onRemoveFilter: (fieldKey: string) => void;
  onRefresh: () => Promise<void> | void;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
  rows: EntityApiRow[];
}) {
  const surface = contract.viewSchemaId as WorkbookSurface;
  const writableFields = useMemo(
    () => contract.fields.filter((field) => field.writeKind !== "read_only"),
    [contract],
  );
  const [createDraft, setCreateDraft] = useState<Record<string, string>>({});
  const [editRecordId, setEditRecordId] = useState("");
  const [editFieldKey, setEditFieldKey] = useState("");
  const [editValue, setEditValue] = useState("");
  const [mutationError, setMutationError] = useState<string | null>(null);
  const [mutationState, setMutationState] = useState<SaveState>("Saved");
  const [evidenceMessageByRecordID, setEvidenceMessageByRecordID] = useState<
    Record<string, string>
  >({});
  const [evidencePreview, setEvidencePreview] =
    useState<EvidencePreviewState | null>(null);
  const isEvidenceSurface = contract.viewSchemaId === evidenceViewSchemaId;

  const setEvidenceMessage = useCallback(
    (recordId: string, message: string | null) => {
      setEvidenceMessageByRecordID((current) => {
        const next = { ...current };
        if (message === null) {
          delete next[recordId];
        } else {
          next[recordId] = message;
        }
        return next;
      });
    },
    [],
  );

  const issueEvidenceHandle = useCallback(
    async (row: EntityApiRow, kind: "preview" | "download") => {
      setEvidenceMessage(row.record_id, null);
      const result = await fetchJSON<EvidenceHandleEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/evidence-records/${row.record_id}/${kind}-handle`,
        ),
        { method: "POST", body: JSON.stringify({}) },
      );
      if (!result.ok) {
        setEvidenceMessage(row.record_id, parseErrorMessage(result.payload));
        return;
      }
      const envelope = readEnvelope<EvidenceHandleEnvelope>(result.payload);
      const href =
        envelope.data.href.startsWith("/") && apiBase
          ? apiPath(apiBase, envelope.data.href)
          : envelope.data.href;
      if (kind === "preview") {
        setEvidencePreview({
          href,
          recordId: row.record_id,
          title:
            stringifyGridValue(row.cells["evidence.title"]?.value).trim() ||
            row.record_id,
          previewKind: envelope.data.preview_kind ?? null,
        });
        setEvidenceMessage(row.record_id, "Preview loaded inline.");
        return;
      }

      const anchor = document.createElement("a");
      anchor.href = href;
      anchor.download = envelope.data.filename || "evidence";
      anchor.rel = "noopener";
      document.body.append(anchor);
      anchor.click();
      anchor.remove();
      setEvidenceMessage(row.record_id, "Download handle issued.");
    },
    [apiBase, setEvidenceMessage],
  );

  const columns: readonly GridColumn<EntityApiRow>[] = visibleFields(
    contract,
  ).map((field) => ({
    fieldKey: field.fieldKey,
    headerTestId: gridSortHeaderTestId(surface, field.fieldKey),
    label: field.label,
    width: field.defaultHidden ? 160 : 220,
    renderCell: (row) => (
      <span data-testid={rowCellTestId(row.record_id, field.fieldKey)}>
        {genericCellLabel(row.cells[field.fieldKey]?.value)}
      </span>
    ),
    sortableFieldKey: resolveHeaderSortFieldKey(contract, field.fieldKey),
  }));
  const evidenceActionsColumn = useMemo<
    GridActionsColumn<EntityApiRow> | undefined
  >(() => {
    if (!isEvidenceSurface) {
      return undefined;
    }
    return {
      label: "Access",
      width: 208,
      renderCell: ({ data: row }) => {
        const uploadState = stringifyGridValue(
          row.cells["evidence.upload_state"]?.value,
        );
        const lifecycleState = stringifyGridValue(
          row.cells["evidence.lifecycle_state"]?.value,
        );
        const canAccess =
          uploadState === "available" &&
          (lifecycleState === "available" || lifecycleState === "released");
        const message =
          evidenceMessageByRecordID[row.record_id] ??
          (canAccess ? null : `Blocked: ${uploadState || "no blob"}`);
        return (
          <div style={actionStackStyle}>
            <div style={inlineButtonRowStyle}>
              <button
                data-testid={`evidence-preview-${row.record_id}`}
                disabled={!canAccess}
                style={actionButtonStyle}
                type="button"
                onClick={() => {
                  void issueEvidenceHandle(row, "preview");
                }}
              >
                Preview
              </button>
              <button
                data-testid={`evidence-download-${row.record_id}`}
                disabled={!canAccess}
                style={actionButtonStyle}
                type="button"
                onClick={() => {
                  void issueEvidenceHandle(row, "download");
                }}
              >
                Download
              </button>
            </div>
            {message ? (
              <span
                data-testid={`evidence-access-message-${row.record_id}`}
                style={evidenceAccessMessageStyle}
              >
                {message}
              </span>
            ) : null}
          </div>
        );
      },
    };
  }, [evidenceMessageByRecordID, isEvidenceSurface, issueEvidenceHandle]);
  const gridRows: readonly GridRow<EntityApiRow>[] = rows.map((row) => ({
    key: row.record_id,
    recordId: row.record_id,
    data: row,
    testId: `generic-row-${contract.viewSchemaId}-${row.record_id}`,
  }));
  const selectedEditRow =
    rows.find((row) => row.record_id === editRecordId) ?? null;
  const selectedEditField =
    writableFields.find((field) => field.fieldKey === editFieldKey) ??
    writableFields[0] ??
    null;

  const submitCreate = async () => {
    const payload = buildGenericCreatePayload(
      writableFields,
      createDraft,
      `generic-create-${contract.viewSchemaId}-${Date.now()}`,
    );
    if (payload === null) {
      setMutationError("invalid_mutation_payload");
      return;
    }
    setMutationState("Syncing");
    setMutationError(null);
    const result = await fetchJSON<ViewMutationEnvelope>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${contract.viewSchemaId}/rows`,
      ),
      { method: "POST", body: JSON.stringify(payload) },
    );
    if (!result.ok) {
      setMutationState("Conflict");
      setMutationError(parseErrorMessage(result.payload));
      return;
    }
    setCreateDraft({});
    setMutationState("Saved");
    await onRefresh();
  };

  const submitEdit = async () => {
    if (selectedEditRow === null || selectedEditField === null) {
      setMutationError("invalid_mutation_payload");
      return;
    }
    const change = buildGenericPatchChange(selectedEditField, editValue);
    if (change === null) {
      setMutationError("invalid_mutation_payload");
      return;
    }
    setMutationState("Syncing");
    setMutationError(null);
    const result = await fetchJSON<ViewMutationEnvelope>(
      apiPath(apiBase, `/api/v1/records/${selectedEditRow.record_id}`),
      {
        method: "PATCH",
        body: JSON.stringify({
          view_schema_id: contract.viewSchemaId,
          base_row_version: selectedEditRow.row_version,
          client_txn_id: `generic-patch-${contract.viewSchemaId}-${Date.now()}`,
          changes: [change],
        }),
      },
    );
    if (!result.ok) {
      setMutationState("Conflict");
      setMutationError(parseErrorMessage(result.payload));
      return;
    }
    setEditValue("");
    setMutationState("Saved");
    await onRefresh();
  };

  return (
    <section style={workbookStyle}>
      <header style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>
            {contract.surfaceKind === "built_in_sheet"
              ? "Built-in sheet"
              : "System view"}
          </p>
          <h1 style={headlineStyle}>{contract.title}</h1>
          <p style={bodyStyle}>Incident {incidentId}</p>
        </div>
        <div style={roleBadgeStyle} data-testid="generic-mutation-state">
          {mutationState}
        </div>
      </header>

      {writableFields.length > 0 ? (
        <section style={genericMutationPanelStyle}>
          <div style={genericFormGridStyle}>
            {writableFields.map((field) => (
              <label
                key={field.fieldKey}
                htmlFor={`generic-create-input-${field.fieldKey}`}
                style={labelStyle}
              >
                {field.label}
                {field.writeKind === "action_payload" ? (
                  <textarea
                    data-testid={`generic-create-field-${field.fieldKey}`}
                    id={`generic-create-input-${field.fieldKey}`}
                    style={textareaStyle}
                    value={createDraft[field.fieldKey] ?? ""}
                    onChange={(event) => {
                      setCreateDraft((current) => ({
                        ...current,
                        [field.fieldKey]: event.target.value,
                      }));
                    }}
                  />
                ) : (
                  <input
                    data-testid={`generic-create-field-${field.fieldKey}`}
                    id={`generic-create-input-${field.fieldKey}`}
                    style={inputStyle}
                    value={createDraft[field.fieldKey] ?? ""}
                    onChange={(event) => {
                      setCreateDraft((current) => ({
                        ...current,
                        [field.fieldKey]: event.target.value,
                      }));
                    }}
                  />
                )}
              </label>
            ))}
          </div>
          <button
            data-testid={`generic-create-submit-${contract.viewSchemaId}`}
            style={actionButtonStyle}
            type="button"
            onClick={() => {
              void submitCreate();
            }}
          >
            Create
          </button>

          {rows.length > 0 && selectedEditField !== null ? (
            <div style={genericEditRowStyle}>
              <select
                data-testid={`generic-edit-record-${contract.viewSchemaId}`}
                style={selectStyle}
                value={editRecordId}
                onChange={(event) => {
                  setEditRecordId(event.target.value);
                }}
              >
                <option value="">Row</option>
                {rows.map((row) => (
                  <option key={row.record_id} value={row.record_id}>
                    {row.record_id}
                  </option>
                ))}
              </select>
              <select
                data-testid={`generic-edit-field-${contract.viewSchemaId}`}
                style={selectStyle}
                value={editFieldKey}
                onChange={(event) => {
                  setEditFieldKey(event.target.value);
                }}
              >
                <option value="">Field</option>
                {writableFields.map((field) => (
                  <option key={field.fieldKey} value={field.fieldKey}>
                    {field.label}
                  </option>
                ))}
              </select>
              {selectedEditField.writeKind === "action_payload" ? (
                <textarea
                  data-testid={`generic-edit-value-${contract.viewSchemaId}`}
                  style={textareaStyle}
                  value={editValue}
                  onChange={(event) => {
                    setEditValue(event.target.value);
                  }}
                />
              ) : (
                <input
                  data-testid={`generic-edit-value-${contract.viewSchemaId}`}
                  style={inputStyle}
                  value={editValue}
                  onChange={(event) => {
                    setEditValue(event.target.value);
                  }}
                />
              )}
              <button
                data-testid={`generic-edit-submit-${contract.viewSchemaId}`}
                style={actionButtonStyle}
                type="button"
                onClick={() => {
                  void submitEdit();
                }}
              >
                Update
              </button>
            </div>
          ) : null}

          {mutationError ? (
            <p
              data-testid="generic-mutation-error"
              style={genericErrorTextStyle}
            >
              {mutationError}
            </p>
          ) : null}
        </section>
      ) : null}

      <WorkbookGridControls
        contract={contract}
        filterDraft={filterDraft}
        onApplyFilter={onApplyFilter}
        onFilterDraftChange={onFilterDraftChange}
        onGroupByChange={onGroupByChange}
        onRemoveFilter={onRemoveFilter}
        queryState={queryState}
        surface={surface}
      />

      {loadError ? (
        <p data-testid="generic-surface-load-error" style={bodyStyle}>
          {loadError}
        </p>
      ) : null}

      {isEvidenceSurface && evidencePreview ? (
        <section
          data-testid="evidence-preview-panel"
          style={evidencePreviewPanelStyle}
        >
          <div style={evidencePreviewHeaderStyle}>
            <div>
              <p style={eyebrowStyle}>Preview</p>
              <h2 style={sectionTitleStyle}>{evidencePreview.title}</h2>
            </div>
            <button
              style={secondaryActionButtonStyle}
              type="button"
              onClick={() => {
                setEvidencePreview(null);
              }}
            >
              Close
            </button>
          </div>
          <iframe
            data-testid={`evidence-preview-frame-${evidencePreview.recordId}`}
            src={evidencePreview.href}
            style={evidencePreviewFrameStyle}
            title={`Evidence preview ${evidencePreview.title}`}
          />
          {evidencePreview.previewKind ? (
            <p style={evidenceAccessMessageStyle}>
              {evidencePreview.previewKind}
            </p>
          ) : null}
        </section>
      ) : null}

      <GridViewport style={gridShellStyle} testId={gridShellTestId(surface)}>
        <GridTable
          actionsColumn={evidenceActionsColumn}
          columns={columns}
          getGroupLabel={(row, fieldKey) =>
            genericCellLabel(row.cells[fieldKey]?.value)
          }
          getGroupRowTestId={(fieldKey, value) =>
            gridGroupRowTestId(surface, fieldKey, value)
          }
          groupBy={queryState.groupBy}
          onToggleSort={onToggleSort}
          rows={gridRows}
          sort={queryState.sort}
        />
      </GridViewport>
    </section>
  );
}

function genericCellLabel(value: unknown): string {
  if (value === null || value === undefined || value === "") {
    return "None";
  }
  if (typeof value === "string" || typeof value === "number") {
    return String(value);
  }
  if (typeof value === "boolean") {
    return value ? "Yes" : "No";
  }
  if (typeof value === "object" && value !== null && "items" in value) {
    const items = (value as { items?: unknown }).items;
    if (Array.isArray(items)) {
      return `${items.length} item${items.length === 1 ? "" : "s"}`;
    }
  }
  return JSON.stringify(value);
}

function buildGenericCreatePayload(
  fields: readonly ViewFieldContract[],
  draft: Record<string, string>,
  clientTxnId: string,
): Record<string, unknown> | null {
  const payload: Record<string, unknown> = { client_txn_id: clientTxnId };
  for (const field of fields) {
    const value = normalizeValue(draft[field.fieldKey] ?? "");
    if (value === "") {
      continue;
    }
    if (field.writeKind === "action_payload") {
      const actionPayload = buildGenericCollectionActions(field, value);
      if (actionPayload !== null) {
        payload[field.fieldKey] = actionPayload;
      }
      continue;
    }
    payload[field.fieldKey] = value;
  }
  return Object.keys(payload).length > 1 ? payload : null;
}

function buildGenericPatchChange(
  field: ViewFieldContract,
  rawValue: string,
): Record<string, unknown> | null {
  const value = normalizeValue(rawValue);
  if (value === "" && !field.clearable) {
    return null;
  }
  if (field.writeKind === "action_payload") {
    const actionPayload = buildGenericCollectionActions(field, value);
    return actionPayload === null
      ? null
      : { field_key: field.fieldKey, action_payload: actionPayload };
  }
  return {
    field_key: field.fieldKey,
    value: value === "" && field.clearable ? null : value,
  };
}

function buildGenericCollectionActions(
  field: ViewFieldContract,
  rawValue: string,
): Record<string, unknown> | null {
  const tokens = rawValue
    .split(/\r?\n/u)
    .map((value) => normalizeValue(value))
    .filter((value) => value !== "");
  if (tokens.length === 0) {
    return null;
  }
  const actions = tokens.map((value) =>
    field.fieldKey === "note.tags"
      ? { op: "add_token", raw_text: value }
      : { op: "add_record_ref", linked_record_id: value },
  );
  return { kind: "collection_actions_v1", actions };
}

function enumValuesFor(
  contract: ViewContract,
  fieldKey: string,
  fallback: readonly string[],
): readonly string[] {
  return contract.fieldMap[fieldKey]?.enumValues ?? fallback;
}

function isAssessmentConfidenceBand(
  value: string,
): value is AssessmentConfidenceBand {
  return (
    value === "unset" ||
    value === "low" ||
    value === "medium" ||
    value === "high"
  );
}

function initialAssessmentDraft(): AssessmentCreateDraft {
  const [assessmentState = "unknown"] = enumValuesFor(
    assessmentsContract,
    "assessment.assessment_state",
    ["unknown", "suspected", "confirmed", "disproven", "cleared"],
  );
  const confidenceBand = enumValuesFor(
    assessmentsContract,
    "assessment.confidence_band",
    ["unset", "low", "medium", "high"],
  ).find(isAssessmentConfidenceBand);
  return {
    assessedAt: "",
    assessmentState,
    confidenceBand: confidenceBand ?? "unset",
    rationale: "",
    subjectRecordId: "",
    subjectType: "host",
    supportRecordIds: [],
  };
}

function assessmentColumnWidth(fieldKey: string): number {
  switch (fieldKey) {
    case "assessment.subject_ref":
      return 300;
    case "assessment.rationale":
      return 360;
    case "assessment.assessed_at":
      return 210;
    case "assessment.assessor":
      return 300;
    default:
      return 180;
  }
}

function supportRowLabel(row: TimelineApiRow): string {
  const summary = readStringCell(row, "timeline.summary");
  if (summary !== "") {
    return summary;
  }
  return row.record_id;
}

function surfaceSlug(viewSchemaID: string): string {
  return (
    Object.entries(legacySurfaceParamToViewSchemaID).find(
      ([, mappedViewSchemaID]) => mappedViewSchemaID === viewSchemaID,
    )?.[0] ??
    viewSchemaID.replace(/^cartulary\.view\./, "").replace(/\.v1$/, "")
  );
}

export function WorkbookShell({
  incidentId,
  apiBase,
  onIncidentSnapshot,
  onIncidentAccessLost,
}: WorkbookShellProps) {
  const params = useMemo(() => new URLSearchParams(window.location.search), []);
  const initialViewSchemaID = useMemo(() => {
    const explicit = params.get("view_schema_id");
    if (
      explicit &&
      allWorkbookContracts.some(
        (contract) => contract.viewSchemaId === explicit,
      )
    ) {
      return explicit;
    }
    const legacySurface = params.get("surface") ?? "timeline";
    return (
      legacySurfaceParamToViewSchemaID[legacySurface] ?? timelineViewSchemaId
    );
  }, [params]);
  const [surface, setSurface] = useState<SurfaceKind>(initialViewSchemaID);
  const [hostRows, setHostRows] = useState<EntityRow[]>([]);
  const [identityRows, setIdentityRows] = useState<EntityRow[]>([]);
  const [entityLoadError, setEntityLoadError] = useState<string | null>(null);
  const [genericRows, setGenericRows] = useState<EntityApiRow[]>([]);
  const [genericLoadError, setGenericLoadError] = useState<string | null>(null);
  const [assessmentRows, setAssessmentRows] = useState<EntityApiRow[]>([]);
  const [assessmentLoadError, setAssessmentLoadError] = useState<string | null>(
    null,
  );
  const [currentIncidentRole, setCurrentIncidentRole] =
    useState<IncidentRole | null>(null);
  const [hostQueryState, setHostQueryState] = useState<WorkbookQueryState>(() =>
    emptyWorkbookQueryState(),
  );
  const [identityQueryState, setIdentityQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [hostFilterDraft, setHostFilterDraft] = useState<FilterDraft>(() =>
    defaultFilterDraft(hostsContract),
  );
  const [identityFilterDraft, setIdentityFilterDraft] = useState<FilterDraft>(
    () => defaultFilterDraft(identitiesContract),
  );
  const [assessmentQueryState, setAssessmentQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [assessmentFilterDraft, setAssessmentFilterDraft] =
    useState<FilterDraft>(() => defaultFilterDraft(assessmentsContract));
  const activeContract = useMemo(
    () =>
      allWorkbookContracts.find(
        (contract) => contract.viewSchemaId === surface,
      ) ?? timelineContract,
    [surface],
  );
  const [genericQueryState, setGenericQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [genericFilterDraft, setGenericFilterDraft] = useState<FilterDraft>(
    () => defaultFilterDraft(activeContract),
  );

  const entityIndex = useMemo(() => {
    const index: Record<string, EntityRow> = {};
    for (const row of [...hostRows, ...identityRows]) {
      index[row.recordId] = row;
    }
    return index;
  }, [hostRows, identityRows]);

  const loadSessionRole = useCallback(async () => {
    const result = await fetchJSON<SessionEnvelope>(
      apiPath(apiBase, "/api/v1/auth/session"),
    );
    if (!result.ok) {
      setCurrentIncidentRole("");
      onIncidentAccessLost?.();
      return;
    }
    const envelope = readEnvelope<SessionEnvelope>(result.payload);
    const membership =
      envelope.data.memberships.find(
        (entry) => entry.incident_id === incidentId,
      ) ?? null;
    if (membership === null) {
      onIncidentAccessLost?.();
    }
    setCurrentIncidentRole(membership?.role ?? "");
  }, [apiBase, incidentId, onIncidentAccessLost]);

  const queryEntityView = useCallback(
    async (
      viewSchemaId: string,
      entityType: EntityRow["entityType"],
      queryState: WorkbookQueryState,
    ) => {
      const contract =
        viewSchemaId === hostsViewSchemaId ? hostsContract : identitiesContract;
      const result = await fetchJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
        ),
        {
          method: "POST",
          body: JSON.stringify(buildQueryRequest(contract, queryState)),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      return envelope.data.rows.map((row) => entityRowFromApi(row, entityType));
    },
    [apiBase, incidentId],
  );

  const loadEntities = useCallback(async () => {
    setEntityLoadError(null);
    try {
      const [nextHosts, nextIdentities] = await Promise.all([
        queryEntityView(hostsViewSchemaId, "host", hostQueryState),
        queryEntityView(identitiesViewSchemaId, "identity", identityQueryState),
      ]);
      setHostRows((current) => [...reconcileRecordRows(current, nextHosts)]);
      setIdentityRows((current) => [
        ...reconcileRecordRows(current, nextIdentities),
      ]);
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Entity load failed.";
      if (
        typeof message === "string" &&
        (message.includes("incident_not_found") ||
          message.includes("authorization_denied"))
      ) {
        onIncidentAccessLost?.();
      }
      setEntityLoadError(message);
    }
  }, [
    hostQueryState,
    identityQueryState,
    onIncidentAccessLost,
    queryEntityView,
  ]);

  const applyHostFilter = useCallback(() => {
    setHostQueryState((current) => applyFilterDraft(current, hostFilterDraft));
    setHostFilterDraft((current) => ({
      ...current,
      booleanValue: "",
      value: "",
    }));
  }, [hostFilterDraft]);

  const applyIdentityFilter = useCallback(() => {
    setIdentityQueryState((current) =>
      applyFilterDraft(current, identityFilterDraft),
    );
    setIdentityFilterDraft((current) => ({
      ...current,
      booleanValue: "",
      value: "",
    }));
  }, [identityFilterDraft]);

  const isSpecializedSurface =
    surface === timelineViewSchemaId ||
    surface === hostsViewSchemaId ||
    surface === identitiesViewSchemaId ||
    surface === assessmentsViewSchemaId;

  const loadAssessmentSurface = useCallback(async () => {
    if (surface !== assessmentsViewSchemaId) {
      return;
    }
    setAssessmentLoadError(null);
    try {
      const result = await fetchJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${assessmentsViewSchemaId}/query`,
        ),
        {
          method: "POST",
          body: JSON.stringify(
            buildQueryRequest(assessmentsContract, assessmentQueryState),
          ),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      setAssessmentRows(envelope.data.rows);
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Assessment load failed.";
      if (
        typeof message === "string" &&
        (message.includes("incident_not_found") ||
          message.includes("authorization_denied"))
      ) {
        onIncidentAccessLost?.();
      }
      setAssessmentLoadError(message);
      setAssessmentRows([]);
    }
  }, [
    apiBase,
    assessmentQueryState,
    incidentId,
    onIncidentAccessLost,
    surface,
  ]);

  const loadGenericSurface = useCallback(async () => {
    if (isSpecializedSurface) {
      return;
    }
    setGenericLoadError(null);
    try {
      const result = await fetchJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${surface}/query`,
        ),
        {
          method: "POST",
          body: JSON.stringify(
            buildQueryRequest(activeContract, genericQueryState),
          ),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      setGenericRows(envelope.data.rows);
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Surface load failed.";
      if (
        typeof message === "string" &&
        (message.includes("incident_not_found") ||
          message.includes("authorization_denied"))
      ) {
        onIncidentAccessLost?.();
      }
      setGenericLoadError(message);
      setGenericRows([]);
    }
  }, [
    activeContract,
    apiBase,
    genericQueryState,
    incidentId,
    isSpecializedSurface,
    onIncidentAccessLost,
    surface,
  ]);

  useEffect(() => {
    void Promise.all([loadEntities(), loadSessionRole()]);
  }, [loadEntities, loadSessionRole]);

  useEffect(() => {
    setGenericQueryState(emptyWorkbookQueryState());
    setGenericFilterDraft(defaultFilterDraft(activeContract));
    setGenericRows([]);
    setGenericLoadError(null);
  }, [activeContract]);

  useEffect(() => {
    void loadGenericSurface();
  }, [loadGenericSurface]);

  useEffect(() => {
    void loadAssessmentSurface();
  }, [loadAssessmentSurface]);

  useEffect(() => {
    const next = new URLSearchParams(window.location.search);
    next.set("incident_id", incidentId);
    next.set("view_schema_id", surface);
    const legacySurface = Object.entries(legacySurfaceParamToViewSchemaID).find(
      ([, viewSchemaID]) => viewSchemaID === surface,
    )?.[0];
    if (legacySurface) {
      next.set("surface", legacySurface);
    } else {
      next.delete("surface");
    }
    window.history.replaceState({}, "", `/?${next.toString()}`);
  }, [incidentId, surface]);

  return (
    <section style={panelStyle}>
      <div style={heroStyle}>
        <p style={eyebrowStyle}>Cartulary</p>
        <h1 style={headlineStyle}>Timeline workbook shell</h1>
        <p style={bodyStyle}>
          Phase 3 workbook behavior stays intact while Phase 4 adds entity
          mention review, stub creation, merge review, and same-surface
          auto-resolution disclosure.
        </p>
      </div>

      <div style={toolbarStyle}>
        <div style={tabStripStyle}>
          {builtInSurfaceIDs.map((viewSchemaID) => {
            const contract = requireViewContract(viewSchemaID);
            return (
              <button
                key={viewSchemaID}
                data-testid={`surface-tab-${surfaceSlug(viewSchemaID)}`}
                style={{
                  ...surfaceTabStyle,
                  ...(surface === viewSchemaID ? surfaceTabActiveStyle : null),
                }}
                type="button"
                onClick={() => {
                  setSurface(viewSchemaID);
                }}
              >
                {contract.title}
              </button>
            );
          })}
          <label style={systemViewSelectLabelStyle}>
            System view
            <select
              data-testid="system-view-selector"
              style={selectStyle}
              value={
                builtInSurfaceIDs.includes(
                  surface as (typeof builtInSurfaceIDs)[number],
                )
                  ? ""
                  : surface
              }
              onChange={(event) => {
                if (event.target.value) {
                  setSurface(event.target.value);
                }
              }}
            >
              <option value="">Select view</option>
              {allWorkbookContracts
                .filter(
                  (contract) =>
                    !builtInSurfaceIDs.includes(
                      contract.viewSchemaId as (typeof builtInSurfaceIDs)[number],
                    ),
                )
                .map((contract) => (
                  <option
                    key={contract.viewSchemaId}
                    value={contract.viewSchemaId}
                  >
                    {contract.title}
                  </option>
                ))}
            </select>
          </label>
        </div>
        <div style={roleBadgeStyle}>
          Current incident role: {currentIncidentRole || "viewer"}
        </div>
      </div>

      <IncidentAdminPanel
        apiBase={apiBase}
        currentIncidentRole={currentIncidentRole}
        incidentId={incidentId}
        onIncidentAccessLost={onIncidentAccessLost}
        onIncidentSnapshot={onIncidentSnapshot}
        onSessionRoleChange={loadSessionRole}
      />

      {entityLoadError ? (
        <p data-testid="entity-load-error" style={bodyStyle}>
          {entityLoadError}
        </p>
      ) : null}

      {surface === timelineViewSchemaId ? (
        <TimelineWorkbook
          apiBase={apiBase}
          currentIncidentRole={currentIncidentRole}
          entityIndex={entityIndex}
          hostEntities={hostRows}
          identityEntities={identityRows}
          incidentId={incidentId}
          onRefreshEntities={loadEntities}
        />
      ) : surface === hostsViewSchemaId ||
        surface === identitiesViewSchemaId ? (
        <EntityWorkbookSurface
          apiBase={apiBase}
          currentIncidentRole={currentIncidentRole}
          entityIndex={entityIndex}
          entityType={surface === hostsViewSchemaId ? "host" : "identity"}
          filterDraft={
            surface === hostsViewSchemaId
              ? hostFilterDraft
              : identityFilterDraft
          }
          incidentId={incidentId}
          onApplyFilter={
            surface === hostsViewSchemaId
              ? applyHostFilter
              : applyIdentityFilter
          }
          onFilterDraftChange={
            surface === hostsViewSchemaId
              ? setHostFilterDraft
              : setIdentityFilterDraft
          }
          onGroupByChange={(groupBy) => {
            if (surface === hostsViewSchemaId) {
              setHostQueryState((current) =>
                updateGroupBy(hostsContract, current, groupBy),
              );
              return;
            }
            setIdentityQueryState((current) =>
              updateGroupBy(identitiesContract, current, groupBy),
            );
          }}
          onRemoveFilter={(fieldKey) => {
            if (surface === hostsViewSchemaId) {
              setHostQueryState((current) =>
                removeFilterField(current, fieldKey),
              );
              return;
            }
            setIdentityQueryState((current) =>
              removeFilterField(current, fieldKey),
            );
          }}
          onRefreshEntities={loadEntities}
          onToggleSort={(fieldKey) => {
            if (surface === hostsViewSchemaId) {
              setHostQueryState((current) =>
                toggleSortField(hostsContract, current, fieldKey),
              );
              return;
            }
            setIdentityQueryState((current) =>
              toggleSortField(identitiesContract, current, fieldKey),
            );
          }}
          queryState={
            surface === hostsViewSchemaId ? hostQueryState : identityQueryState
          }
          rows={surface === hostsViewSchemaId ? hostRows : identityRows}
        />
      ) : surface === assessmentsViewSchemaId ? (
        <AssessmentWorkbookSurface
          apiBase={apiBase}
          assessmentRows={assessmentRows}
          currentIncidentRole={currentIncidentRole}
          filterDraft={assessmentFilterDraft}
          hostRows={hostRows}
          identityRows={identityRows}
          incidentId={incidentId}
          loadError={assessmentLoadError}
          onApplyFilter={() => {
            setAssessmentQueryState((current) =>
              applyFilterDraft(current, assessmentFilterDraft),
            );
            setAssessmentFilterDraft((current) => ({
              ...current,
              booleanValue: "",
              value: "",
            }));
          }}
          onFilterDraftChange={setAssessmentFilterDraft}
          onGroupByChange={(groupBy) => {
            setAssessmentQueryState((current) =>
              updateGroupBy(assessmentsContract, current, groupBy),
            );
          }}
          onRefreshAssessmentRows={loadAssessmentSurface}
          onRemoveFilter={(fieldKey) => {
            setAssessmentQueryState((current) =>
              removeFilterField(current, fieldKey),
            );
          }}
          onToggleSort={(fieldKey) => {
            setAssessmentQueryState((current) =>
              toggleSortField(assessmentsContract, current, fieldKey),
            );
          }}
          queryState={assessmentQueryState}
        />
      ) : (
        <GenericWorkbookSurface
          key={activeContract.viewSchemaId}
          apiBase={apiBase}
          contract={activeContract}
          filterDraft={genericFilterDraft}
          incidentId={incidentId}
          loadError={genericLoadError}
          onApplyFilter={() => {
            setGenericQueryState((current) =>
              applyFilterDraft(current, genericFilterDraft),
            );
            setGenericFilterDraft((current) => ({
              ...current,
              booleanValue: "",
              value: "",
            }));
          }}
          onFilterDraftChange={setGenericFilterDraft}
          onGroupByChange={(groupBy) => {
            setGenericQueryState((current) =>
              updateGroupBy(activeContract, current, groupBy),
            );
          }}
          onRemoveFilter={(fieldKey) => {
            setGenericQueryState((current) =>
              removeFilterField(current, fieldKey),
            );
          }}
          onRefresh={loadGenericSurface}
          onToggleSort={(fieldKey) => {
            setGenericQueryState((current) =>
              toggleSortField(activeContract, current, fieldKey),
            );
          }}
          queryState={genericQueryState}
          rows={genericRows}
        />
      )}
    </section>
  );
}

const panelStyle = {
  width: "min(96rem, 100%)",
  margin: "0 auto",
  padding: "2rem",
  borderRadius: "1.5rem",
  background: "rgb(255 252 247 / 0.92)",
  boxShadow: "0 24px 80px rgb(29 78 70 / 0.12)",
  border: "1px solid rgb(185 204 196 / 0.8)",
};

const heroStyle = {
  marginBottom: "1.5rem",
};

const toolbarStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "center",
  flexWrap: "wrap" as const,
  marginBottom: "1rem",
};

const workbookStyle = {
  marginTop: "1.5rem",
};

const headerStyle = {
  position: "relative" as const,
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "flex-start",
  marginBottom: "1rem",
};

const blurSurfaceButtonStyle = {
  position: "absolute" as const,
  inset: 0,
  zIndex: 1,
  border: 0,
  padding: 0,
  margin: 0,
  background: "transparent",
  color: "transparent",
  cursor: "default",
};

const statusClusterStyle = {
  display: "grid",
  gap: "0.25rem",
  minWidth: "8rem",
  textAlign: "right" as const,
};

const statusLabelStyle = {
  fontSize: "0.75rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  color: "rgb(81 110 103)",
};

const statusValueStyle = {
  fontSize: "1rem",
};

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  color: "rgb(87 112 104)",
};

const headlineStyle = {
  margin: "0.35rem 0 0.5rem",
  fontSize: "2rem",
  lineHeight: 1.1,
};

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "rgb(53 79 72)",
};

const splitShellStyle = {
  display: "grid",
  gap: "1rem",
  alignItems: "start",
  gridTemplateColumns: "minmax(0, 3fr) minmax(20rem, 1.25fr)",
};

const gridShellStyle = {
  overflow: "auto",
  overflowAnchor: "none" as const,
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(255 255 255 / 0.82)",
  maxHeight: "70vh",
};

const actionStackStyle = {
  display: "grid",
  gap: "0.5rem",
};

const genericMutationPanelStyle = {
  display: "grid",
  gap: "0.75rem",
  padding: "1rem",
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(247 250 248)",
};

const genericFormGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.75rem",
};

const genericEditRowStyle = {
  display: "grid",
  gridTemplateColumns:
    "minmax(10rem, 1fr) minmax(10rem, 1fr) minmax(14rem, 2fr) auto",
  gap: "0.75rem",
  alignItems: "end",
};

const inputStyle = {
  boxSizing: "border-box" as const,
  minWidth: 0,
  width: "100%",
  borderRadius: "0.75rem",
  border: "1px solid rgb(192 205 198)",
  background: "rgb(255 255 255)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "inherit",
};

const textareaStyle = {
  ...inputStyle,
  resize: "vertical" as const,
};

const replacementInputStyle = {
  ...inputStyle,
  fontSize: "0.9rem",
};

const actionButtonStyle = {
  borderRadius: "999px",
  border: "1px solid rgb(129 165 154)",
  background: "rgb(234 244 239)",
  color: "rgb(34 74 63)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "rgb(247 249 247)",
};

const genericErrorTextStyle = {
  margin: 0,
  color: "rgb(153 27 27)",
  fontWeight: 700,
};

const evidenceAccessMessageStyle = {
  margin: 0,
  fontSize: "0.85rem",
  color: "rgb(87 109 103)",
};

const evidencePreviewPanelStyle = {
  display: "grid",
  gap: "0.75rem",
  margin: "1rem 0",
  padding: "1rem",
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(255 255 255 / 0.86)",
};

const evidencePreviewHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "start",
};

const evidencePreviewFrameStyle = {
  width: "100%",
  minHeight: "28rem",
  border: "1px solid rgb(199 214 207)",
  borderRadius: "0.75rem",
  background: "rgb(255 255 255)",
};

const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "rgb(52 79 72)",
};

const inspectorShellStyle = {
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(248 251 249)",
  padding: "1rem",
  position: "sticky" as const,
  top: "1rem",
};

const inspectorHeaderStyle = {
  display: "grid",
  gap: "0.35rem",
  marginBottom: "1rem",
};

const inspectorTitleStyle = {
  margin: 0,
  fontSize: "1.25rem",
};

const inspectorSectionStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
};

const sectionTitleStyle = {
  margin: 0,
  fontSize: "1rem",
};

const relationshipItemsWrapStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "0.4rem",
  marginBottom: "0.55rem",
  maxWidth: "100%",
  minWidth: 0,
};

const relationshipChipStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  borderRadius: "999px",
  padding: "0.35rem 0.65rem",
  font: "inherit",
  lineHeight: 1.2,
  maxWidth: "100%",
  minWidth: 0,
  overflowWrap: "anywhere" as const,
};

const unresolvedChipStyle = {
  border: "1px dashed rgb(188 118 23)",
  background: "rgb(255 246 219)",
  color: "rgb(124 71 10)",
};

const resolvedChipStyle = {
  border: "1px solid rgb(73 143 108)",
  background: "rgb(225 245 233)",
  color: "rgb(29 90 59)",
};

const autoResolvedChipStyle = {
  border: "1px solid rgb(44 100 138)",
  background: "rgb(223 241 255)",
  color: "rgb(20 67 97)",
};

const dismissedChipStyle = {
  border: "1px solid rgb(128 128 128)",
  background: "rgb(240 240 240)",
  color: "rgb(76 76 76)",
};

const selectedChipStyle = {
  boxShadow: "0 0 0 2px rgb(31 94 70 / 0.22)",
};

const chipMetaStyle = {
  fontSize: "0.72rem",
  textTransform: "uppercase" as const,
  letterSpacing: "0.04em",
};

const aliasChipStyle = {
  ...relationshipChipStyle,
  border: "1px solid rgb(188 205 198)",
  background: "rgb(240 247 243)",
  color: "rgb(45 74 66)",
};

const tagChipStyle = {
  ...relationshipChipStyle,
  border: "1px solid rgb(178 191 184)",
  background: "rgb(242 246 243)",
  color: "rgb(54 74 67)",
};

const emptyRelationshipStyle = {
  color: "rgb(117 136 130)",
  fontSize: "0.9rem",
};

const mentionGroupStyle = {
  display: "grid",
  gap: "0.75rem",
};

const mentionGroupColumnStyle = {
  display: "grid",
  gap: "0.5rem",
};

const groupLabelStyle = {
  margin: 0,
  fontSize: "0.8rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  color: "rgb(89 112 105)",
};

const mentionListButtonStyle = {
  border: "none",
  background: "transparent",
  padding: 0,
  textAlign: "left" as const,
  cursor: "pointer",
};

const mentionListButtonSelectedStyle = {
  outline: "none",
};

const detailListStyle = {
  display: "grid",
  gap: "0.75rem",
  margin: 0,
};

const detailTermStyle = {
  fontSize: "0.75rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  color: "rgb(87 109 103)",
};

const detailValueStyle = {
  margin: "0.2rem 0 0",
};

const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap" as const,
};

const inspectorActionStackStyle = {
  display: "grid",
  gap: "0.75rem",
};

const noticeStackStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
};

const noticeCardStyle = {
  borderRadius: "1rem",
  border: "1px solid rgb(158 194 214)",
  background: "rgb(240 248 255)",
  padding: "0.85rem 1rem",
  display: "grid",
  gap: "0.5rem",
};

const noticeTitleStyle = {
  margin: 0,
  fontSize: "0.95rem",
  fontWeight: 600,
};

const tabStripStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap" as const,
};

const surfaceTabStyle = {
  borderRadius: "999px",
  border: "1px solid rgb(176 194 187)",
  background: "rgb(248 251 249)",
  color: "rgb(42 77 67)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

const surfaceTabActiveStyle = {
  background: "rgb(34 74 63)",
  color: "rgb(250 252 251)",
  borderColor: "rgb(34 74 63)",
};

const systemViewSelectLabelStyle = {
  ...labelStyle,
  minWidth: "16rem",
};

const roleBadgeStyle = {
  borderRadius: "999px",
  background: "rgb(236 244 239)",
  color: "rgb(44 76 66)",
  padding: "0.45rem 0.8rem",
  fontSize: "0.9rem",
};

const selectStyle = {
  ...inputStyle,
  appearance: "auto" as const,
};

const mergePlanStyle = {
  borderRadius: "1rem",
  border: "1px solid rgb(205 219 213)",
  background: "rgb(255 255 255)",
  padding: "0.9rem",
  display: "grid",
  gap: "0.65rem",
};

const flatListStyle = {
  margin: 0,
  paddingLeft: "1.2rem",
  display: "grid",
  gap: "0.35rem",
};

const timelinePreviewStackStyle = {
  display: "grid",
  gap: "0.75rem",
};

const timelinePreviewCardStyle = {
  borderRadius: "0.9rem",
  border: "1px solid rgb(210 222 216)",
  background: "rgb(255 255 255)",
  padding: "0.85rem",
  display: "grid",
  gap: "0.55rem",
};
