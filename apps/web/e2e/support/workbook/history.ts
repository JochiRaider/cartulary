import type {
  GetRecordHistoryResponse,
  RecordHistoryData,
} from "@cartulary/protocol-ts/http";
import { rowHistoryItemTestId } from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";

import { apiBase } from "../runtime/configuration";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

type RecordHistoryOptions = {
  readonly cursorToken?: string;
  readonly limit?: number;
};

export async function fetchRecordHistoryPage(
  page: Page,
  recordId: string,
  options: RecordHistoryOptions = {},
): Promise<GetRecordHistoryResponse> {
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
  return response.payload satisfies GetRecordHistoryResponse;
}

export async function fetchRecordHistory(
  page: Page,
  recordId: string,
  options: RecordHistoryOptions = {},
): Promise<RecordHistoryData> {
  return (await fetchRecordHistoryPage(page, recordId, options)).data;
}

export async function fetchFullRecordHistory(
  page: Page,
  recordId: string,
): Promise<RecordHistoryData> {
  let envelope = await fetchRecordHistoryPage(page, recordId);
  let data = envelope.data;
  const items = new Map(
    data.items.map((item) => [item.history_item_ref, item]),
  );
  const cursors = new Set<string>();
  while (envelope.meta.paging.has_more) {
    const cursorToken = envelope.meta.paging.next_cursor;
    if (!cursorToken || cursors.has(cursorToken))
      throw new Error("History cursor did not advance");
    cursors.add(cursorToken);
    envelope = await fetchRecordHistoryPage(page, recordId, { cursorToken });
    if (
      envelope.data.representation_generation !== data.representation_generation
    )
      throw new Error("History generation changed during fixture read");
    for (const item of envelope.data.items)
      items.set(item.history_item_ref, item);
    if (envelope.data.row_version >= data.row_version) data = envelope.data;
  }
  return { ...data, items: [...items.values()] };
}

export async function fetchRecordHistoryCount(page: Page, recordId: string) {
  return (await fetchFullRecordHistory(page, recordId)).items.length;
}

/** Open the source event before inspecting detail or choosing a reversal. */
export async function openHistoryEventDetails(
  root: Page | Locator,
  historyItemRef: string,
) {
  const disclosure = root
    .getByTestId(rowHistoryItemTestId({ historyItemRef }))
    .locator(":scope > details");
  if ((await disclosure.getAttribute("open")) === null)
    await disclosure.locator(":scope > summary").click();
}
