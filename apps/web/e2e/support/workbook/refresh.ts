import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";
import { apiBase } from "../runtime/configuration";

export async function fetchRecordHistoryCount(page: Page, recordId: string) {
  const response = await page.request.get(
    `${apiBase}/api/v1/records/${recordId}/history?limit=100`,
  );
  expect(response.ok).toBeTruthy();
  const body = (await response.json()) as {
    data: { items: unknown[] };
  };
  return body.data.items.length;
}
