import type {
  CreateViewRowRequest,
  CreateViewRowResponse,
  HTTPOperationResponse,
  PatchRecordRequest,
  PatchRecordResponse,
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
} from "@cartulary/protocol-ts/http";
import type { Page, Response } from "@playwright/test";

import { expect } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";
import { apiBase } from "../runtime/configuration";
import {
  publicHttpOperation,
  readHttpOperationResponse,
} from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

export type ViewApiRow = CreateViewRowResponse["data"]["row"];
export type ViewApiCell = ViewApiRow["cells"][string];
export type WorkbookMutationOperationID = "createViewRow" | "patchRecord";

export type ViewRowPollOptions = {
  readonly diagnosticContext?: string;
  readonly timeout?: number;
};

export async function createViewRow(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  payload: Record<string, unknown>,
): Promise<ViewApiRow> {
  const response = await publicHttpOperation({
    body: payload as unknown as CreateViewRowRequest,
    headers: await csrfHeaders(page),
    operationID: "createViewRow",
    pathParameters: {
      incident_id: incidentId,
      view_schema_id: viewSchemaId,
    },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `createViewRow failed with HTTP ${response.status}: ${JSON.stringify(response.payload)}`,
    );
  }
  return response.payload.data.row;
}

export async function queryViewRows(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  requestBody: QueryWorkbookViewRequest = {},
): Promise<ViewApiRow[]> {
  const response = await publicHttpOperation({
    body: requestBody,
    operationID: "queryWorkbookView",
    pathParameters: {
      incident_id: incidentId,
      view_schema_id: viewSchemaId,
    },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `queryWorkbookView failed with HTTP ${response.status}: ${JSON.stringify(response.payload)}`,
    );
  }
  return (response.payload satisfies QueryWorkbookViewResponse).data.rows;
}

export async function waitForViewRow(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  recordId: string,
  options: ViewRowPollOptions = {},
): Promise<ViewApiRow> {
  return waitForMatchingViewRow({
    diagnostic: `${viewSchemaId} default query should include created row ${recordId}`,
    incidentId,
    matcher: (row) => row.record_id === recordId,
    options,
    page,
    viewSchemaId,
  });
}

export async function waitForViewRowByCell(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  fieldKey: string,
  value: unknown,
  options: ViewRowPollOptions = {},
): Promise<ViewApiRow> {
  return waitForMatchingViewRow({
    diagnostic: `${viewSchemaId} default query should include row where ${fieldKey}=${JSON.stringify(value)}`,
    incidentId,
    matcher: (row) => row.cells[fieldKey]?.value === value,
    options,
    page,
    viewSchemaId,
  });
}

async function waitForMatchingViewRow({
  diagnostic,
  incidentId,
  matcher,
  options,
  page,
  viewSchemaId,
}: {
  readonly diagnostic: string;
  readonly incidentId: string;
  readonly matcher: (row: ViewApiRow) => boolean;
  readonly options: ViewRowPollOptions;
  readonly page: Page;
  readonly viewSchemaId: string;
}): Promise<ViewApiRow> {
  let lastRows: ViewApiRow[] = [];
  let matched: ViewApiRow | undefined;
  try {
    await expect
      .poll(
        async () => {
          lastRows = await queryViewRows(page, incidentId, viewSchemaId);
          matched = lastRows.find(matcher);
          return matched !== undefined;
        },
        {
          message: pollDiagnostic(
            diagnostic,
            incidentId,
            viewSchemaId,
            options.diagnosticContext,
          ),
          timeout: options.timeout ?? 10_000,
        },
      )
      .toBe(true);
  } catch (error) {
    throw new Error(
      [
        pollDiagnostic(
          diagnostic,
          incidentId,
          viewSchemaId,
          options.diagnosticContext,
        ),
        `last_rows=${JSON.stringify(
          lastRows.map((row) => ({
            record_id: row.record_id,
            row_version: row.row_version,
          })),
        )}`,
      ].join("\n"),
      { cause: error },
    );
  }
  if (matched === undefined) {
    throw new Error(
      `${diagnostic}\nlast_row_ids=${JSON.stringify(lastRows.map((row) => row.record_id))}`,
    );
  }
  return matched;
}

function pollDiagnostic(
  diagnostic: string,
  incidentId: string,
  viewSchemaId: string,
  context: string | undefined,
) {
  return [
    diagnostic,
    `incident_id=${incidentId}`,
    `view_schema_id=${viewSchemaId}`,
    ...(context === undefined ? [] : [`context=${context}`]),
  ].join("\n");
}

export async function readWorkbookMutation<
  OperationID extends WorkbookMutationOperationID,
>(
  response: Response,
  operationID: OperationID,
): Promise<HTTPOperationResponse<OperationID>> {
  return readHttpOperationResponse(response, operationID);
}

export async function patchRecord(
  page: Page,
  recordId: string,
  payload: Record<string, unknown>,
) {
  const response = await publicHttpOperation({
    body: payload as unknown as PatchRecordRequest,
    headers: await csrfHeaders(page),
    operationID: "patchRecord",
    pathParameters: { record_id: recordId },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `patchRecord failed with HTTP ${response.status}: ${JSON.stringify(response.payload)}`,
    );
  }
  return (response.payload satisfies PatchRecordResponse).data.row;
}
