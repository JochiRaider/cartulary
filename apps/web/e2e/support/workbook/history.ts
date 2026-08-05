import type {
  GetRecordHistoryResponse,
  RecordHistoryData,
} from "@cartulary/protocol-ts/http";
import type { Page } from "@playwright/test";

import { apiBase } from "../runtime/configuration";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

type RecordHistoryOptions = {
  readonly cursorToken?: string;
  readonly limit?: number;
};

export async function fetchRecordHistory(
  page: Page,
  recordId: string,
  options: RecordHistoryOptions = {},
): Promise<RecordHistoryData> {
  const response = await publicHttpOperation({
    operationID: "getRecordHistory",
    pathParameters: { record_id: recordId },
    query: {
      ...(options.cursorToken === undefined
        ? {}
        : { cursor_token: options.cursorToken }),
      ...(options.limit === undefined ? {} : { limit: options.limit }),
    },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `getRecordHistory failed with HTTP ${response.status}: ${JSON.stringify(response.payload)}`,
    );
  }
  return (response.payload satisfies GetRecordHistoryResponse).data;
}

export async function fetchRecordHistoryCount(page: Page, recordId: string) {
  return (await fetchRecordHistory(page, recordId, { limit: 100 })).items
    .length;
}
