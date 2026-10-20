import {
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import type { RecordChangedPayload } from "../workbook/collaboration/workbookCollaborationMessages";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
  workbookQueryStateFromSavedViewQueryJson,
} from "../workbook/models/workbookQuery";
import { timelineViewSchemaId } from "../workbook/models/workbookSurfaceRegistry";
import type { WorkbookViewQueryAccepted } from "../workbook/query/WorkbookViewQueryPort";

type TimelineRowOptions = {
  recordId: string;
  rowVersion: number;
  dateEnteredText?: string;
  analystText?: string;
  mitreStageText?: string;
  deviceObjectText?: string;
  ipAddressText?: string;
  activityUTCText?: string;
  activityLocalText?: string;
  rawActivityText?: string;
  activitySynopsisText?: string;
  dataSourceText?: string;
  occurredAt?: string;
  summary?: string;
  details?: string;
  sourceText?: string;
  captureState: string;
  evidenceCount?: number;
  tags?: string[];
  editedAt?: string;
  hasEvidence?: boolean;
  hostRefs?: Array<Record<string, unknown>>;
  identityRefs?: Array<Record<string, unknown>>;
};

export type WorkbookViewApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
  group_values?: Record<string, unknown>;
  view_schema_id?: string;
};

export type WorkbookViewRowOverrides = Record<string, unknown>;

type RecordChangedPayloadOptions = {
  recordId: string;
  rowVersion: number;
  clientTxnId: string;
  changeSetId?: string;
  actorUserId?: string;
  changedFieldKeys?: string[];
  affectedViews?: RecordChangedPayload["affected_views"];
};

export function successEnvelope(data: unknown, status = 200) {
  const isWorkbookQuery =
    data !== null &&
    typeof data === "object" &&
    "incident_id" in data &&
    "view_schema_id" in data &&
    "rows" in data;
  const isSavedViewList =
    data !== null &&
    typeof data === "object" &&
    "saved_views" in data &&
    Array.isArray(data.saved_views);
  return new Response(
    JSON.stringify({
      data,
      meta: {
        request_id: `req-${status}`,
        ...(isWorkbookQuery
          ? {
              ...workbookQueryMeta(String(data.view_schema_id)),
            }
          : {}),
        ...(isSavedViewList
          ? {
              paging: { has_more: false, limit: 100, next_cursor: null },
            }
          : {}),
      },
    }),
    {
      status,
      headers: { "Content-Type": "application/json" },
    },
  );
}

export function workbookCollectionValue(
  ordered: boolean,
  items: readonly Record<string, unknown>[] = [],
) {
  return {
    kind: "collection_value_v1",
    ordered,
    items: [...items],
  };
}

function defaultWorkbookViewCellValue(field: ViewFieldContract): unknown {
  if (field.readKind === "collection") {
    return workbookCollectionValue(false);
  }
  if (field.readKind === "number") {
    return 0;
  }
  if (field.readKind === "boolean") {
    return false;
  }
  return null;
}

export function fullWorkbookViewRow(
  contract: ViewContract,
  recordId: string,
  rowVersion: number,
  overrides: WorkbookViewRowOverrides,
): WorkbookViewApiRow {
  const cells = Object.fromEntries(
    contract.fields.map((field) => [
      field.fieldKey,
      {
        value:
          field.fieldKey in overrides
            ? overrides[field.fieldKey]
            : defaultWorkbookViewCellValue(field),
      },
    ]),
  );
  return {
    record_id: recordId,
    row_version: rowVersion,
    view_schema_id: contract.viewSchemaId,
    cells,
    ...(contract.groupingFields.length === 0
      ? {}
      : {
          group_values: Object.fromEntries(
            contract.groupingFields.map((fieldKey) => [
              fieldKey,
              cells[fieldKey]?.value,
            ]),
          ),
        }),
  };
}

function viewRowsEnvelope(
  viewSchemaId: string,
  rows: readonly WorkbookViewApiRow[],
  incidentId = "incident-1",
) {
  return successEnvelope({
    incident_id: incidentId,
    view_schema_id: viewSchemaId,
    rows: rows.map(({ view_schema_id: rowViewSchemaId, ...row }) => {
      void rowViewSchemaId;
      return row;
    }),
  });
}

export function viewRowsEnvelopeForView(
  viewSchemaId: string,
  rowsByView: Record<string, readonly WorkbookViewApiRow[]>,
  incidentId = "incident-1",
) {
  return viewRowsEnvelope(
    viewSchemaId,
    rowsByView[viewSchemaId] ?? [],
    incidentId,
  );
}

export function timelineRowsEnvelope(
  rows: readonly ReturnType<typeof timelineRow>[],
  incidentId = "10000000-0000-4000-8000-000000000001",
) {
  return successEnvelope({
    incident_id: incidentId,
    view_schema_id: timelineViewSchemaId,
    rows: [...rows],
  });
}

export function timelineMutationEnvelope(
  row: ReturnType<typeof timelineRow>,
  changeSetId: string,
) {
  return successEnvelope({
    view_schema_id: timelineViewSchemaId,
    change_set_id: changeSetId,
    row,
  });
}

export function errorEnvelope(
  code: string,
  status: number,
  conflict?: unknown,
) {
  return new Response(
    JSON.stringify({
      error: {
        status,
        code,
        message: code,
        request_id: "req-error",
        retryable: false,
        details: {},
        ...(conflict === undefined ? {} : { conflict }),
      },
    }),
    {
      status,
      headers: { "Content-Type": "application/json" },
    },
  );
}

export function buildRecordChangedPayload({
  recordId,
  rowVersion,
  clientTxnId,
  changeSetId = `change-set-${rowVersion}`,
  actorUserId = "user-1",
  changedFieldKeys = ["timeline.host_refs"],
  affectedViews = [
    {
      view_schema_id: timelineViewSchemaId,
      change_kind: "invalidate",
    },
  ],
}: RecordChangedPayloadOptions): RecordChangedPayload {
  return {
    record_id: recordId,
    row_version: rowVersion,
    change_set_id: changeSetId,
    client_txn_id: clientTxnId,
    actor_user_id: actorUserId,
    changed_field_keys: changedFieldKeys,
    affected_views: affectedViews,
  };
}

export function timelineRow({
  recordId,
  rowVersion,
  dateEnteredText = "",
  analystText = "",
  mitreStageText = "",
  deviceObjectText = "",
  ipAddressText = "",
  activityUTCText,
  activityLocalText = "",
  rawActivityText,
  activitySynopsisText,
  dataSourceText = "",
  occurredAt = "",
  summary = "",
  details = "",
  sourceText = "",
  captureState,
  evidenceCount = 0,
  tags = [],
  editedAt = "",
  hasEvidence = false,
  hostRefs = [],
  identityRefs = [],
}: TimelineRowOptions) {
  const utcText = activityUTCText ?? occurredAt;
  const rawText = rawActivityText ?? (sourceText !== "" ? sourceText : details);
  const synopsisText = activitySynopsisText ?? summary;
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "timeline.date_entered_text": { value: dateEnteredText },
      "timeline.analyst_text": { value: analystText },
      "timeline.mitre_stage_text": { value: mitreStageText },
      "timeline.device_object_text": { value: deviceObjectText },
      "timeline.ip_address_text": { value: ipAddressText },
      "timeline.activity_utc_text": { value: utcText },
      "timeline.activity_local_text": { value: activityLocalText },
      "timeline.raw_activity_text": { value: rawText },
      "timeline.activity_synopsis_text": { value: synopsisText },
      "timeline.data_source_text": { value: dataSourceText },
      "timeline.host_refs": { value: collectionValue(true, hostRefs) },
      "timeline.identity_refs": { value: collectionValue(true, identityRefs) },
      "timeline.evidence_count": { value: evidenceCount },
      "timeline.tags": { value: collectionValue(false, tags.map(tagItem)) },
      "timeline.attached_evidence_ids": { value: collectionValue(false, []) },
      "timeline.edited_at": { value: editedAt },
      "timeline.recorded_at": { value: "" },
      "timeline.activity_sort_ts": { value: utcText },
      "timeline.date_entered_sort_day": {
        value: dateEnteredText === "" ? null : dateEnteredText.slice(0, 10),
      },
      "timeline.activity_time_pair_state": { value: "disabled" },
      "timeline.capture_state": { value: captureState },
      "timeline.replacement_record_id": { value: null },
      "timeline.has_evidence": { value: hasEvidence },
      "timeline.has_unresolved_mentions": { value: false },
    },
    group_values: {
      "timeline.date_entered_sort_day":
        dateEnteredText === "" ? null : dateEnteredText.slice(0, 10),
      "timeline.activity_time_pair_state": "disabled",
      "timeline.capture_state": captureState,
      "timeline.has_evidence": hasEvidence,
      "timeline.has_unresolved_mentions": false,
    },
  };
}

function collectionValue(
  ordered: boolean,
  items: Array<Record<string, unknown>>,
) {
  return workbookCollectionValue(ordered, items);
}

function tagItem(value: string) {
  return {
    item_ref: `record_tag:${value}`,
    item_kind: "tag",
    display_text: value,
    raw_text: value,
  };
}

/** Complete first-page metadata for fixtures; malformed-envelope tests supply raw metadata. */
export function acceptedQueryMetadata(
  viewSchemaId: string,
  queryState: WorkbookQueryState = emptyWorkbookQueryState(),
): Pick<
  WorkbookViewQueryAccepted,
  "canonicalQuery" | "paging" | "producingRequest"
> {
  const sort = [...queryState.sort];
  for (const entry of requireViewContract(viewSchemaId).defaultSort) {
    if (!sort.some((candidate) => candidate.fieldKey === entry.fieldKey))
      sort.push(entry);
  }
  return {
    canonicalQuery: {
      filters: queryState.filters,
      sort,
      ...(queryState.groupBy ? { groupBy: queryState.groupBy } : {}),
    },
    paging: { limit: 100, hasMore: false, nextCursor: null },
    producingRequest: { queryState, limit: 100 },
  };
}

export function workbookQueryMeta(
  viewSchemaId: string,
  queryState: WorkbookQueryState = emptyWorkbookQueryState(),
) {
  const { canonicalQuery } = acceptedQueryMetadata(viewSchemaId, queryState);
  return {
    request_id: "req-query",
    paging: { limit: 100, has_more: false, next_cursor: null },
    query: {
      filters: canonicalQuery.filters.map((filter) => ({
        field_key: filter.fieldKey,
        op: filter.op,
        arg: filter.arg,
      })),
      sort: canonicalQuery.sort.map((sort) => ({
        field_key: sort.fieldKey,
        direction: sort.direction,
      })),
      ...(canonicalQuery.groupBy ? { group_by: canonicalQuery.groupBy } : {}),
    },
  };
}

/** Transport for interaction fixtures that author rows separately from query controls.
 * Captures the producing body before a deferred reply; raw adapter-contract tests
 * deliberately use their own transport instead.
 */
export function withWorkbookQueryFixtureMetadata(fetchFixture: unknown) {
  if (typeof fetchFixture !== "function")
    throw new Error("Expected callable interaction fixture");
  return async (
    input: RequestInfo | URL,
    init?: RequestInit,
  ): Promise<Response> => {
    const isQuery = /\/views\/[^/]+\/query(?:\?.*)?$/.test(String(input));
    const body = isQuery ? JSON.parse(String(init?.body ?? "{}")) : null;
    const response = await fetchFixture(input, init);
    if (!(response instanceof Response))
      throw new Error("Interaction fixture must return a Response");
    if (!isQuery || !response.ok) return response;
    const envelope = await response.clone().json();
    const schema = envelope.data?.view_schema_id;
    if (typeof schema !== "string" || !Array.isArray(envelope.data?.rows))
      return response;
    return new Response(
      JSON.stringify({
        ...envelope,
        meta: workbookQueryMeta(
          schema,
          workbookQueryStateFromSavedViewQueryJson(
            requireViewContract(schema),
            body,
          ),
        ),
      }),
      { status: response.status, headers: response.headers },
    );
  };
}
