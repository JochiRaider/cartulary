import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";
import { apiBase } from "../runtime/configuration";
import { createEnvironmentTestControlClient } from "../transport/testControlEnvironment";

export async function fetchTimelineRecordSubstrate(
  page: Page,
  recordId: string,
) {
  const response = await createEnvironmentTestControlClient(page.request, {
    endpointOrigin: apiBase,
  }).request({
    method: "GET",
    path: `/api/v1/test/timeline/records/${recordId}/substrate`,
  });
  expect(response.ok).toBeTruthy();
  return (
    response.body as {
      data: {
        record_id: string;
        row_version: number;
        capture_state: string;
        replacement_record_id: string | null;
        record_revision_count: number;
      };
    }
  ).data;
}

export async function fetchTimelineRecordChangeCount(
  page: Page,
  recordId: string,
) {
  const response = await createEnvironmentTestControlClient(page.request, {
    endpointOrigin: apiBase,
  }).request({
    method: "GET",
    path: "/api/v1/test/timeline/record-changes",
  });
  expect(response.ok).toBeTruthy();
  const body = response.body as {
    data: { record_changes: Array<{ record_id: string }> };
  };
  return body.data.record_changes.filter(
    (change) => change.record_id === recordId,
  ).length;
}
