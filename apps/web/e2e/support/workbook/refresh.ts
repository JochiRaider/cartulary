import type { GetRecordHistoryResponse } from "@cartulary/protocol-ts/http";
import type { Page } from "@playwright/test";

import { apiBase } from "../runtime/configuration";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

export async function fetchRecordHistoryCount(page: Page, recordId: string) {
  const response = await publicHttpOperation({
    operationID: "getRecordHistory",
    pathParameters: { record_id: recordId },
    query: { limit: 100 },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `getRecordHistory failed with HTTP ${response.status}: ${JSON.stringify(response.payload)}`,
    );
  }
  return (response.payload satisfies GetRecordHistoryResponse).data.items
    .length;
}
