import type {
  CreateViewRowRequest,
  HTTPOperationResponse,
  PatchRecordRequest,
  PatchRecordResponse,
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
  ViewRow,
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

type WorkbookMutationOperationID = "createViewRow" | "patchRecord";

type ViewRowPollOptions = {
  readonly diagnosticContext?: string;
  readonly mode?: "poll" | "single_attempt";
  readonly timeout?: number;
};

export async function createViewRow(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  payload: CreateViewRowRequest,
): Promise<ViewRow> {
  const response = await publicHttpOperation({
    body: payload,
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
): Promise<ViewRow[]> {
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
): Promise<ViewRow> {
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
): Promise<ViewRow> {
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
  readonly matcher: (row: ViewRow) => boolean;
  readonly options: ViewRowPollOptions;
  readonly page: Page;
  readonly viewSchemaId: string;
}): Promise<ViewRow> {
  let lastRows: ViewRow[] = [];
  let matched: ViewRow | undefined;
  const queryForMatch = async () => {
    lastRows = await queryViewRows(page, incidentId, viewSchemaId);
    matched = lastRows.find(matcher);
    return matched !== undefined;
  };
  try {
    if (options.mode === "single_attempt") {
      if (!(await queryForMatch())) {
        throw new Error("single workbook row query did not match");
      }
    } else {
      await expect
        .poll(queryForMatch, {
          message: pollDiagnostic(
            diagnostic,
            incidentId,
            viewSchemaId,
            options.diagnosticContext,
          ),
          timeout: options.timeout ?? 10_000,
        })
        .toBe(true);
    }
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
  payload: PatchRecordRequest,
) {
  const response = await publicHttpOperation({
    body: payload,
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
