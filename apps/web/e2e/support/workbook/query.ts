import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";
import { apiBase } from "../runtime/configuration";
import { atJsonOrigin, requestPublicJson } from "../transport/publicJsonClient";

export type ViewApiCell = {
  value: unknown;
  [key: string]: unknown;
};

export type ViewApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, ViewApiCell>;
  [key: string]: unknown;
};

export async function createViewRow(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  payload: Record<string, unknown>,
): Promise<ViewApiRow> {
  const response = await requestPublicJson({
    body: payload,
    headers: await csrfHeaders(page),
    method: "POST",
    path: `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/rows`,
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(response.ok).toBeTruthy();
  return (response.body as { data: { row: ViewApiRow } }).data.row;
}

export async function queryViewRows(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
): Promise<ViewApiRow[]> {
  const response = await requestPublicJson({
    body: {},
    method: "POST",
    path: `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(response.ok).toBeTruthy();
  return (response.body as { data: { rows: ViewApiRow[] } }).data.rows;
}

export async function waitForViewRow(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  recordId: string,
  options: { timeout?: number } = {},
): Promise<ViewApiRow> {
  let lastRows: ViewApiRow[] = [];
  await expect
    .poll(
      async () => {
        lastRows = await queryViewRows(page, incidentId, viewSchemaId);
        return lastRows.some((row) => row.record_id === recordId);
      },
      {
        message: `${viewSchemaId} default query should include created row ${recordId}`,
        timeout: options.timeout ?? 10_000,
      },
    )
    .toBe(true);
  const row = lastRows.find((candidate) => candidate.record_id === recordId);
  expect(
    row,
    `${viewSchemaId} default query rows: ${lastRows
      .map((candidate) => candidate.record_id)
      .join(", ")}`,
  ).toBeTruthy();
  return row as ViewApiRow;
}

export async function patchRecord(
  page: Page,
  recordId: string,
  payload: Record<string, unknown>,
) {
  const response = await requestPublicJson({
    body: payload,
    headers: await csrfHeaders(page),
    method: "PATCH",
    path: `/api/v1/records/${recordId}`,
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(response.ok).toBeTruthy();
  return (response.body as { data: { row: Record<string, unknown> } }).data.row;
}
