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
  type WorkbookSurface,
} from "@cartulary/test-utils";
import {
  requireViewContract,
  resolveHeaderSortFieldKey,
} from "@cartulary/view-contracts";
import {
  type ChangeEvent,
  type FormEvent,
  type KeyboardEvent,
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
const timelineContract = requireViewContract(timelineViewSchemaId);
const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const csrfCookieName = "cartulary_csrf";
const csrfHeaderName = "X-CSRF-Token";

type SaveState = "Syncing" | "Saved" | "Conflict";
type SurfaceKind = "timeline" | "hosts" | "identities";
type EditableField =
  | "timeline.occurred_at"
  | "timeline.summary"
  | "timeline.details"
  | "timeline.source_text";
type RelationshipFieldKey = "timeline.host_refs" | "timeline.identity_refs";
type RelationshipDraftKey = "hostRefs" | "identityRefs";
type FocusFieldKey = keyof RowValues | RelationshipDraftKey;
type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";

type RowValues = {
  occurredAt: string;
  summary: string;
  details: string;
  sourceText: string;
};

type RelationshipDrafts = {
  hostRefs: string;
  identityRefs: string;
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

type WorkbookRow = {
  key: string;
  recordId: string | null;
  rowVersion: number | null;
  captureState: string;
  values: RowValues;
  committedValues: RowValues;
  collectionValues: Record<RelationshipDraftKey, CollectionItem[]>;
  relationshipDrafts: RelationshipDrafts;
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

type TimelineQueryEnvelope = {
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

type ViewportContinuityFollowup = "none" | "entity-refresh";

type ViewportContinuityState = {
  token: number;
  attemptVersion: number;
  focusRowId: string;
  preservedScroll: { top: number; left: number } | null;
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
};

type MergePlanLine = {
  label: string;
  outcome: string;
};

const fieldOrder: Array<{
  fieldKey: EditableField;
  key: keyof RowValues;
  label: string;
  multiline?: boolean;
}> = [
  {
    fieldKey: "timeline.occurred_at",
    key: "occurredAt",
    label: "Time (RFC3339)",
  },
  {
    fieldKey: "timeline.summary",
    key: "summary",
    label: "Summary",
  },
  {
    fieldKey: "timeline.details",
    key: "details",
    label: "Details",
    multiline: true,
  },
  {
    fieldKey: "timeline.source_text",
    key: "sourceText",
    label: "Source Text",
    multiline: true,
  },
];

const relationshipFields: Array<{
  fieldKey: RelationshipFieldKey;
  draftKey: RelationshipDraftKey;
  entityType: "host" | "identity";
  label: string;
}> = [
  {
    fieldKey: "timeline.host_refs",
    draftKey: "hostRefs",
    entityType: "host",
    label: "Hosts",
  },
  {
    fieldKey: "timeline.identity_refs",
    draftKey: "identityRefs",
    entityType: "identity",
    label: "Identities",
  },
];

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

function emptyRelationshipDrafts(): RelationshipDrafts {
  return {
    hostRefs: "",
    identityRefs: "",
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
    },
    relationshipDrafts: emptyRelationshipDrafts(),
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
    },
    relationshipDrafts: emptyRelationshipDrafts(),
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

export function buildCreatePayload(row: WorkbookRow, clientTxnId: string) {
  const payload: Record<string, unknown> = {
    client_txn_id: clientTxnId,
  };

  for (const field of fieldOrder) {
    const normalized = normalizeValue(row.values[field.key]);
    if (normalized !== "") {
      payload[field.fieldKey] = normalized;
    }
  }

  for (const relationshipField of relationshipFields) {
    const actions = buildCollectionActions(
      row.relationshipDrafts[relationshipField.draftKey],
    );
    if (actions !== null) {
      payload[relationshipField.fieldKey] = actions;
    }
  }

  if (Object.keys(payload).length < 2) {
    return null;
  }
  return payload;
}

function buildScalarPatchPayload(row: WorkbookRow, clientTxnId: string) {
  const changes = fieldOrder
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
  fieldKey: RelationshipFieldKey,
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

function buildMutationSignature(payload: unknown): string {
  return JSON.stringify(payload);
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
  const [pendingFocusKey, setPendingFocusKey] = useState<string | null>(null);
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
  const rowInputRefs = useRef(
    new Map<string, HTMLInputElement | HTMLTextAreaElement>(),
  );
  const gridShellRef = useRef<HTMLDivElement | null>(null);
  const pendingScrollRef = useRef<{ top: number; left: number } | null>(null);
  const viewportContinuityTokenRef = useRef(1);
  const [viewportContinuity, setViewportContinuity] =
    useState<ViewportContinuityState | null>(null);

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
    () =>
      websocketPath(
        apiBase,
        `/ws/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/changes`,
      ),
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

  const captureGridScroll = useCallback(() => {
    pendingScrollRef.current = currentGridScrollSnapshot();
  }, [currentGridScrollSnapshot]);

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
    (preservedScroll: { top: number; left: number } | null) => {
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

  const beginViewportContinuity = useCallback(
    (
      focusRowId: string,
      options: { followup?: ViewportContinuityFollowup } = {},
    ) => {
      const token = viewportContinuityTokenRef.current;
      viewportContinuityTokenRef.current += 1;
      const followup = options.followup ?? "none";
      setViewportContinuity({
        token,
        attemptVersion: 0,
        focusRowId,
        preservedScroll: currentGridScrollSnapshot(),
        followup,
        followupSettled: followup === "none",
        baselineHostEntities: hostEntities,
        baselineIdentityEntities: identityEntities,
      });
      return token;
    },
    [currentGridScrollSnapshot, hostEntities, identityEntities],
  );

  const settleViewportContinuityFollowup = useCallback((token: number) => {
    setViewportContinuity((current) => {
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
    setViewportContinuity((current) =>
      current?.token === token ? null : current,
    );
  }, []);

  const tryRestoreViewportContinuity = useCallback(
    (continuity: ViewportContinuityState) => {
      const element = document.querySelector<HTMLButtonElement>(
        `[data-testid="row-${continuity.focusRowId}-inspect"]`,
      );
      if (element === null) {
        return false;
      }
      window.focus();
      element.focus({ preventScroll: true });
      restoreGridScroll(continuity.preservedScroll);
      return document.activeElement === element;
    },
    [restoreGridScroll],
  );

  const shouldHoldViewportContinuity = useCallback(
    (continuity: ViewportContinuityState) => {
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

  const restoreInputFocus = useCallback(
    (focusKey: string) => {
      let cancelled = false;
      const focusTarget = (attempt: number) => {
        if (cancelled) {
          return;
        }

        const selectorTestId = focusTestIdForKey(focusKey);
        const selector =
          selectorTestId === null
            ? null
            : document.querySelector<HTMLInputElement | HTMLTextAreaElement>(
                `[data-testid="${selectorTestId}"]`,
              );
        const element = selector ?? rowInputRefs.current.get(focusKey);
        if (element) {
          const gridShell = gridShellRef.current;
          const preservedScroll =
            gridShell === null
              ? null
              : {
                  top: gridShell.scrollTop,
                  left: gridShell.scrollLeft,
                };
          window.focus();
          element.focus({ preventScroll: true });
          restoreGridScroll(preservedScroll);
          if (document.activeElement === element) {
            setPendingFocusKey((current) =>
              current === focusKey ? null : current,
            );
            return;
          }
        }

        if (attempt < 60) {
          window.setTimeout(() => focusTarget(attempt + 1), 50);
        }
      };

      focusTarget(0);
      return () => {
        cancelled = true;
      };
    },
    [restoreGridScroll],
  );

  const restoreRowFocus = useCallback(
    (
      recordId: string,
      preservedScrollOverride: { top: number; left: number } | null = null,
    ) => {
      let cancelled = false;
      const focusTarget = (attempt: number) => {
        if (cancelled) {
          return;
        }

        const element = document.querySelector<HTMLButtonElement>(
          `[data-testid="row-${recordId}-inspect"]`,
        );
        if (element) {
          const preservedScroll =
            preservedScrollOverride ??
            (gridShellRef.current === null
              ? null
              : {
                  top: gridShellRef.current.scrollTop,
                  left: gridShellRef.current.scrollLeft,
                });
          window.focus();
          element.focus({ preventScroll: true });
          restoreGridScroll(preservedScroll);
          if (document.activeElement === element) {
            return;
          }
        }

        if (attempt < 60) {
          window.setTimeout(() => focusTarget(attempt + 1), 50);
        }
      };

      focusTarget(0);
      return () => {
        cancelled = true;
      };
    },
    [restoreGridScroll],
  );

  const focusMountedInput = useCallback(
    (element: HTMLInputElement | HTMLTextAreaElement, focusKey: string) => {
      if (pendingFocusKey !== focusKey) {
        return;
      }
      window.setTimeout(() => {
        const gridShell = gridShellRef.current;
        const preservedScroll =
          gridShell === null
            ? null
            : {
                top: gridShell.scrollTop,
                left: gridShell.scrollLeft,
              };
        element.focus({ preventScroll: true });
        restoreGridScroll(preservedScroll);
        if (document.activeElement === element) {
          setPendingFocusKey((current) =>
            current === focusKey ? null : current,
          );
        }
      }, 0);
    },
    [pendingFocusKey, restoreGridScroll],
  );

  const applyRowMutation = useCallback(
    (
      rowKey: string,
      envelope: TimelineMutationEnvelope,
      options: {
        focusField?: FocusFieldKey | null;
        continueOnFreshDraft?: boolean;
        detectAutoResolution?: boolean;
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
      if (options.continueOnFreshDraft && hydrated.draftSummaryKey) {
        setPendingFocusKey(hydrated.draftSummaryKey);
      } else if (options.focusField && committed.recordId !== null) {
        const focusKey = `${committed.key}:${options.focusField}`;
        setPendingFocusKey(focusKey);
      }
      setViewportContinuity((current) =>
        current === null
          ? null
          : {
              ...current,
              attemptVersion: current.attemptVersion + 1,
            },
      );
    },
    [nextDraftIndex, selectedRowId],
  );

  const loadRows = useCallback(
    async (options: LoadRowsOptions) => {
      const requestSequence = loadSequenceRef.current + 1;
      loadSequenceRef.current = requestSequence;

      if (options.showLoading) {
        setIsLoading(true);
      }
      setLoadError(null);

      const result = await fetchJSON<TimelineQueryEnvelope>(queryPath, {
        method: "POST",
        body: queryBody,
      });

      if (requestSequence !== loadSequenceRef.current) {
        return;
      }

      if (!result.ok) {
        setLoadError("Timeline projection load failed.");
        setIsLoading(false);
        return;
      }

      const envelope = readEnvelope<TimelineQueryEnvelope>(result.payload);
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
    [nextDraftIndex, queryBody, queryPath],
  );

  loadRowsRef.current = loadRows;

  useEffect(() => {
    void loadRows({ showLoading: true });
  }, [loadRows]);

  useLayoutEffect(() => {
    void rows;
    if (pendingFocusKey === null) {
      return;
    }
    return restoreInputFocus(pendingFocusKey);
  }, [pendingFocusKey, restoreInputFocus, rows]);

  useLayoutEffect(() => {
    if (viewportContinuity === null || viewportContinuity.attemptVersion < 1) {
      return;
    }
    if (!tryRestoreViewportContinuity(viewportContinuity)) {
      return;
    }
    if (shouldHoldViewportContinuity(viewportContinuity)) {
      return;
    }
    clearViewportContinuity(viewportContinuity.token);
  }, [
    clearViewportContinuity,
    shouldHoldViewportContinuity,
    tryRestoreViewportContinuity,
    viewportContinuity,
  ]);

  useEffect(() => {
    void rows;
    if (pendingScrollRef.current === null || gridShellRef.current === null) {
      return;
    }
    gridShellRef.current.scrollTop = pendingScrollRef.current.top;
    gridShellRef.current.scrollLeft = pendingScrollRef.current.left;
    pendingScrollRef.current = null;
  }, [rows]);

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

    const socket = new WebSocket(changeSocketURL);
    socket.onmessage = (event) => {
      const message = JSON.parse(event.data) as unknown;
      if (
        shouldIgnoreSelfOriginatedRecordChange(message, resolvePendingSocketTxn)
      ) {
        return;
      }
      if (!isRecordChangedMessage(message)) {
        return;
      }
      captureGridScroll();
      void loadRowsRef.current({ showLoading: false });
    };
    return () => {
      socket.close();
    };
  }, [captureGridScroll, changeSocketURL, incidentId, resolvePendingSocketTxn]);

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

  function nextClientTxnId() {
    const value = clientTxnRef.current;
    clientTxnRef.current += 1;
    return `timeline-client-${value}`;
  }

  function beginSave() {
    pendingOpsRef.current += 1;
    setSaveState("Syncing");
  }

  function finishSave(nextState: SaveState) {
    pendingOpsRef.current = Math.max(0, pendingOpsRef.current - 1);
    if (pendingOpsRef.current > 0 && nextState !== "Conflict") {
      setSaveState("Syncing");
      return;
    }
    setSaveState(nextState);
  }

  function setRowValue(rowKey: string, field: keyof RowValues, value: string) {
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
  }

  function setRelationshipDraft(
    rowKey: string,
    field: RelationshipDraftKey,
    value: string,
  ) {
    setRows((current) => {
      const nextRows = current.map((row) =>
        row.key === rowKey
          ? {
              ...row,
              relationshipDrafts: {
                ...row.relationshipDrafts,
                [field]: value,
              },
            }
          : row,
      );
      rowsRef.current = nextRows;
      return nextRows;
    });
  }

  function registerInput(
    rowKey: string,
    field: FocusFieldKey,
    element: HTMLInputElement | HTMLTextAreaElement | null,
  ) {
    const key = `${rowKey}:${field}`;
    if (element === null) {
      rowInputRefs.current.delete(key);
      return;
    }
    rowInputRefs.current.set(key, element);
  }

  function queueScalarSave(rowKey: string, focusContinuation: boolean) {
    const snapshot = rowsRef.current.find(
      (candidate) => candidate.key === rowKey,
    );
    if (!snapshot) {
      return;
    }

    const clientTxnId = nextClientTxnId();
    const payload =
      snapshot.recordId === null
        ? buildCreatePayload(snapshot, clientTxnId)
        : buildScalarPatchPayload(snapshot, clientTxnId);
    if (payload === null) {
      return;
    }

    const mutationSignature = buildMutationSignature(payload);
    if (pendingSignaturesRef.current.get(rowKey) === mutationSignature) {
      return;
    }
    pendingSignaturesRef.current.set(rowKey, mutationSignature);
    beginSave();
    captureGridScroll();

    setRows((current) => {
      const nextRows = current.map((row) =>
        row.key === rowKey
          ? { ...row, pendingSignature: mutationSignature }
          : row,
      );
      rowsRef.current = nextRows;
      return nextRows;
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
          finishSave("Conflict");
          return;
        }

        pendingSignaturesRef.current.delete(rowKey);
        const envelope = readEnvelope<TimelineMutationEnvelope>(result.payload);
        applyRowMutation(rowKey, envelope, {
          continueOnFreshDraft: focusContinuation && snapshot.recordId === null,
          detectAutoResolution: false,
        });
        finishSave("Saved");
      });
  }

  function queueRelationshipSave(
    rowKey: string,
    fieldKey: RelationshipFieldKey,
    focusField: RelationshipDraftKey,
  ) {
    const snapshot = rowsRef.current.find(
      (candidate) => candidate.key === rowKey,
    );
    if (!snapshot) {
      return;
    }
    const draftValue = snapshot.relationshipDrafts[focusField];
    const clientTxnId = nextClientTxnId();
    const payload =
      snapshot.recordId === null
        ? buildCreatePayload(snapshot, clientTxnId)
        : buildCollectionPatchPayload(
            snapshot,
            fieldKey,
            draftValue,
            clientTxnId,
          );
    if (payload === null) {
      return;
    }

    beginSave();
    captureGridScroll();
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
          setInspectorMessage(parseErrorMessage(result.payload));
          finishSave("Conflict");
          return;
        }

        const envelope = readEnvelope<TimelineMutationEnvelope>(result.payload);
        const preservedScroll =
          pendingScrollRef.current === null
            ? null
            : { ...pendingScrollRef.current };
        applyRowMutation(rowKey, envelope);
        finishSave("Saved");
        if (envelope.data.row.record_id) {
          window.setTimeout(() => {
            restoreRowFocus(envelope.data.row.record_id, preservedScroll);
          }, 0);
        }
      });
  }

  function queueAction(rowKey: string, action: "mark-reviewed" | "supersede") {
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
    beginSave();
    captureGridScroll();
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
          finishSave("Conflict");
          return;
        }

        await loadRowsRef.current({ showLoading: false });
        finishSave("Saved");
      });
  }

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

    const viewportContinuityToken = beginViewportContinuity(snapshot.recordId, {
      followup:
        action === "resolve_item" && resolvedRecordId === undefined
          ? "entity-refresh"
          : "none",
    });
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

  function handleBlur(rowKey: string) {
    queueScalarSave(rowKey, false);
  }

  function handleKeyDown(
    event: KeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
    rowKey: string,
  ) {
    if (event.key === "Enter" || event.key === "Tab") {
      event.preventDefault();
      queueScalarSave(rowKey, true);
    }
  }

  function handleRelationshipKeyDown(
    event: KeyboardEvent<HTMLInputElement>,
    rowKey: string,
    fieldKey: RelationshipFieldKey,
    draftKey: RelationshipDraftKey,
  ) {
    if (event.key === "Enter" || event.key === "Tab") {
      event.preventDefault();
      queueRelationshipSave(rowKey, fieldKey, draftKey);
    }
  }

  function handlePaste(rowKey: string) {
    window.setTimeout(() => {
      queueScalarSave(rowKey, false);
    }, 0);
  }

  function handleSelectRow(recordId: string) {
    setSelectedRowId(recordId);
    setInspectorMessage(null);
  }

  function handleSelectMention(rowRecordId: string, itemRef: string) {
    setSelectedRowId(rowRecordId);
    setSelectedMentionRef(itemRef);
    setInspectorMessage(null);
  }

  const timelineColumns: readonly GridColumn<WorkbookRow>[] = [
    {
      fieldKey: "timeline.capture_state",
      headerTestId: gridSortHeaderTestId("timeline", "timeline.capture_state"),
      label: "State",
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
    ...fieldOrder.map(
      (field): GridColumn<WorkbookRow> => ({
        fieldKey: field.fieldKey,
        headerTestId: gridSortHeaderTestId("timeline", field.fieldKey),
        label: timelineContract.fieldMap[field.fieldKey]?.label ?? field.label,
        renderCell: (row) =>
          field.multiline ? (
            <textarea
              aria-label={`${field.label} ${row.recordId ?? "draft row"}`}
              data-testid={
                row.recordId === null
                  ? draftCellTestId(field.key)
                  : rowCellTestId(row.recordId, field.key)
              }
              ref={(element) => {
                registerInput(row.key, field.key, element);
                if (element) {
                  focusMountedInput(element, `${row.key}:${field.key}`);
                }
              }}
              rows={3}
              style={textareaStyle}
              value={row.values[field.key]}
              onBlur={() => {
                handleBlur(row.key);
              }}
              onChange={(event: ChangeEvent<HTMLTextAreaElement>) => {
                setRowValue(row.key, field.key, event.target.value);
              }}
              onFocus={() => {
                if (row.recordId) {
                  handleSelectRow(row.recordId);
                }
              }}
              onInput={(event: FormEvent<HTMLTextAreaElement>) => {
                setRowValue(row.key, field.key, event.currentTarget.value);
              }}
              onKeyDown={(event) => {
                handleKeyDown(event, row.key);
              }}
              onPaste={() => {
                handlePaste(row.key);
              }}
            />
          ) : (
            <input
              aria-label={`${field.label} ${row.recordId ?? "draft row"}`}
              data-testid={
                row.recordId === null
                  ? draftCellTestId(field.key)
                  : rowCellTestId(row.recordId, field.key)
              }
              ref={(element) => {
                registerInput(row.key, field.key, element);
                if (element) {
                  focusMountedInput(element, `${row.key}:${field.key}`);
                }
              }}
              style={inputStyle}
              type="text"
              value={row.values[field.key]}
              onBlur={() => {
                handleBlur(row.key);
              }}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                setRowValue(row.key, field.key, event.target.value);
              }}
              onFocus={() => {
                if (row.recordId) {
                  handleSelectRow(row.recordId);
                }
              }}
              onInput={(event: FormEvent<HTMLInputElement>) => {
                setRowValue(row.key, field.key, event.currentTarget.value);
              }}
              onKeyDown={(event) => {
                handleKeyDown(event, row.key);
              }}
              onPaste={() => {
                handlePaste(row.key);
              }}
            />
          ),
        sortableFieldKey: resolveHeaderSortFieldKey(
          timelineContract,
          field.fieldKey,
        ),
      }),
    ),
    ...relationshipFields.map(
      (field): GridColumn<WorkbookRow> => ({
        fieldKey: field.fieldKey,
        label: field.label,
        renderCell: (row) => (
          <>
            <div
              data-testid={
                row.recordId === null
                  ? draftCellTestId(`${field.draftKey}-items`)
                  : relationshipItemsTestId(row.recordId, field.draftKey)
              }
              style={relationshipItemsWrapStyle}
            >
              {row.collectionValues[field.draftKey].length > 0 ? (
                row.collectionValues[field.draftKey].map((item) => (
                  <RelationshipChip
                    key={item.itemRef}
                    entityIndex={entityIndex}
                    item={item}
                    onSelect={() => {
                      if (row.recordId) {
                        handleSelectMention(row.recordId, item.itemRef);
                      }
                    }}
                  />
                ))
              ) : (
                <span style={emptyRelationshipStyle}>No items</span>
              )}
            </div>
            <input
              aria-label={`${field.label} ${row.recordId ?? "draft row"}`}
              data-testid={
                row.recordId === null
                  ? draftCellTestId(`${field.draftKey}-input`)
                  : rowCellTestId(row.recordId, `${field.draftKey}-input`)
              }
              key={`${row.key}:${field.draftKey}:${row.rowVersion ?? "draft"}`}
              ref={(element) => {
                registerInput(row.key, field.draftKey, element);
                if (element) {
                  focusMountedInput(element, `${row.key}:${field.draftKey}`);
                }
              }}
              style={inputStyle}
              type="text"
              value={row.relationshipDrafts[field.draftKey]}
              onBlur={() => {
                queueRelationshipSave(row.key, field.fieldKey, field.draftKey);
              }}
              onChange={(event) => {
                setRelationshipDraft(
                  row.key,
                  field.draftKey,
                  event.target.value,
                );
              }}
              onFocus={() => {
                if (row.recordId) {
                  handleSelectRow(row.recordId);
                }
              }}
              onKeyDown={(event) => {
                handleRelationshipKeyDown(
                  event,
                  row.key,
                  field.fieldKey,
                  field.draftKey,
                );
              }}
              placeholder={`Add ${field.label.toLowerCase()} token`}
            />
          </>
        ),
      }),
    ),
  ];

  const timelineActionsColumn: GridActionsColumn<WorkbookRow> = {
    label: "Actions",
    renderCell: ({ data: row }) =>
      row.recordId === null ? (
        <span style={bodyStyle}>Draft row</span>
      ) : (
        <div style={actionStackStyle}>
          <button
            data-testid={`row-${row.recordId}-inspect`}
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
  };

  const timelineGridRows: readonly GridRow<WorkbookRow>[] = rows.map((row) => ({
    key: row.key,
    recordId: row.recordId,
    data: row,
    onSelect: () => {
      if (row.recordId) {
        handleSelectRow(row.recordId);
      }
    },
    selected: row.recordId !== null && row.recordId === selectedRowId,
    testId: row.recordId === null ? undefined : `timeline-row-${row.recordId}`,
    variant: row.recordId === null ? "draft" : "default",
  }));

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
              getGroupLabel={(row, fieldKey) =>
                timelineGroupLabel(row, fieldKey)
              }
              getGroupRowTestId={(fieldKey, value) =>
                gridGroupRowTestId("timeline", fieldKey, value)
              }
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
                : "Select a saved row"}
            </h2>
            <p style={bodyStyle}>
              Routine mention review stays on the workbook surface.
            </p>
          </div>
          {selectedRow?.recordId ? (
            <>
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
            <p style={bodyStyle}>
              Pick a saved row to inspect unresolved, resolved, and dismissed
              mentions.
            </p>
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
      renderCell: (row) => row.state,
      sortableFieldKey:
        entityType === "host" ? "host.host_state" : "identity.identity_state",
    },
    {
      fieldKey: "row_version",
      label: "Version",
      renderCell: (row) => row.rowVersion,
    },
  ];
  const entityActionsColumn: GridActionsColumn<EntityRow> = {
    label: "Actions",
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
      const result = await fetchJSON<TimelineQueryEnvelope>(
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
      const envelope = readEnvelope<TimelineQueryEnvelope>(result.payload);
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

export function WorkbookShell({
  incidentId,
  apiBase,
  onIncidentSnapshot,
  onIncidentAccessLost,
}: WorkbookShellProps) {
  const params = useMemo(() => new URLSearchParams(window.location.search), []);
  const initialSurface = (params.get("surface") ?? "timeline") as SurfaceKind;
  const [surface, setSurface] = useState<SurfaceKind>(
    initialSurface === "hosts" || initialSurface === "identities"
      ? initialSurface
      : "timeline",
  );
  const [hostRows, setHostRows] = useState<EntityRow[]>([]);
  const [identityRows, setIdentityRows] = useState<EntityRow[]>([]);
  const [entityLoadError, setEntityLoadError] = useState<string | null>(null);
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

  useEffect(() => {
    void Promise.all([loadEntities(), loadSessionRole()]);
  }, [loadEntities, loadSessionRole]);

  useEffect(() => {
    const next = new URLSearchParams(window.location.search);
    next.set("incident_id", incidentId);
    next.set("surface", surface);
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
          <button
            data-testid="surface-tab-timeline"
            style={{
              ...surfaceTabStyle,
              ...(surface === "timeline" ? surfaceTabActiveStyle : null),
            }}
            type="button"
            onClick={() => {
              setSurface("timeline");
            }}
          >
            Timeline
          </button>
          <button
            data-testid="surface-tab-hosts"
            style={{
              ...surfaceTabStyle,
              ...(surface === "hosts" ? surfaceTabActiveStyle : null),
            }}
            type="button"
            onClick={() => {
              setSurface("hosts");
            }}
          >
            Hosts
          </button>
          <button
            data-testid="surface-tab-identities"
            style={{
              ...surfaceTabStyle,
              ...(surface === "identities" ? surfaceTabActiveStyle : null),
            }}
            type="button"
            onClick={() => {
              setSurface("identities");
            }}
          >
            Identities
          </button>
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

      {surface === "timeline" ? (
        <TimelineWorkbook
          apiBase={apiBase}
          currentIncidentRole={currentIncidentRole}
          entityIndex={entityIndex}
          hostEntities={hostRows}
          identityEntities={identityRows}
          incidentId={incidentId}
          onRefreshEntities={loadEntities}
        />
      ) : (
        <EntityWorkbookSurface
          apiBase={apiBase}
          currentIncidentRole={currentIncidentRole}
          entityIndex={entityIndex}
          entityType={surface === "hosts" ? "host" : "identity"}
          filterDraft={
            surface === "hosts" ? hostFilterDraft : identityFilterDraft
          }
          incidentId={incidentId}
          onApplyFilter={
            surface === "hosts" ? applyHostFilter : applyIdentityFilter
          }
          onFilterDraftChange={
            surface === "hosts" ? setHostFilterDraft : setIdentityFilterDraft
          }
          onGroupByChange={(groupBy) => {
            if (surface === "hosts") {
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
            if (surface === "hosts") {
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
            if (surface === "hosts") {
              setHostQueryState((current) =>
                toggleSortField(hostsContract, current, fieldKey),
              );
              return;
            }
            setIdentityQueryState((current) =>
              toggleSortField(identitiesContract, current, fieldKey),
            );
          }}
          queryState={surface === "hosts" ? hostQueryState : identityQueryState}
          rows={surface === "hosts" ? hostRows : identityRows}
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

const inputStyle = {
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
};

const relationshipChipStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  borderRadius: "999px",
  padding: "0.35rem 0.65rem",
  font: "inherit",
  lineHeight: 1.2,
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
