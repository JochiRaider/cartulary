import type {
  TimelineCreateRequest,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { authHeadersForStorageState } from "../auth/storageState";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";
import { createViewRow } from "../workbook/query";

const timelineFixtureBaseOccurredAt = "2026-04-10T10:00:00.000Z";

export function timelineFixtureOccurredAt(
  offsetMinutes: number,
  baseOccurredAt = timelineFixtureBaseOccurredAt,
) {
  const baseMs = Date.parse(baseOccurredAt);
  if (!Number.isFinite(baseMs)) {
    throw new Error(
      `invalid timeline fixture base timestamp: ${baseOccurredAt}`,
    );
  }
  return new Date(baseMs + offsetMinutes * 60_000).toISOString();
}

export async function createTimelineFillers(
  page: Page,
  incidentId: string,
  prefix: string,
  count: number,
  options: {
    occurredAtStart?: string;
    occurredAtStepMinutes?: number;
  } = {},
) {
  const occurredAtStartMs =
    options.occurredAtStart === undefined
      ? null
      : Date.parse(options.occurredAtStart);
  if (occurredAtStartMs !== null && !Number.isFinite(occurredAtStartMs)) {
    throw new Error(
      `invalid timeline filler start timestamp: ${options.occurredAtStart}`,
    );
  }
  const occurredAtStepMs = (options.occurredAtStepMinutes ?? 1) * 60_000;
  for (let index = 1; index <= count; index += 1) {
    const payload: TimelineCreateRequest = {
      client_txn_id: uniqueTxn(`${prefix}-${index}`),
      "timeline.activity_synopsis_text": `${prefix} ${index}`,
    };
    if (occurredAtStartMs !== null) {
      payload["timeline.activity_utc_text"] = new Date(
        occurredAtStartMs + (index - 1) * occurredAtStepMs,
      ).toISOString();
    }
    await createViewRow(page, incidentId, timelineViewSchemaId, payload);
  }
}

// One authentication snapshot belongs to this serial fixture operation only.
// Explicit timestamps make returned creation identities match presentation order.
export async function createTimelineRangeRows(
  page: Page,
  incidentId: string,
  count: number,
) {
  const headers = authHeadersForStorageState(
    await page.context().storageState(),
  );
  const request = atJsonOrigin(page.request, apiBase);
  const rows: ViewRow[] = [];
  for (let index = 0; index < count; index++) {
    const response = await publicHttpOperation({
      operationID: "createViewRow",
      request,
      headers,
      pathParameters: {
        incident_id: incidentId,
        view_schema_id: timelineViewSchemaId,
      },
      body: {
        client_txn_id: uniqueTxn("range-seed"),
        "timeline.activity_utc_text": timelineFixtureOccurredAt(index),
        "timeline.activity_synopsis_text": `Range fact ${index}`,
        "timeline.data_source_text": `Range source ${index}`,
      },
    });
    if (!response.ok)
      throw new Error(
        `Timeline range fixture create failed with HTTP ${response.status}`,
      );
    rows.push(response.payload.data.row);
  }
  return rows;
}
