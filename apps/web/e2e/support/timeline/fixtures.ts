import type { Page } from "@playwright/test";

import { timelineViewSchemaId } from "../contracts/workbookSurfaces";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow } from "../workbook/query";

export const timelineFixtureBaseOccurredAt = "2026-04-10T10:00:00.000Z";

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
    const payload: Record<string, unknown> = {
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
