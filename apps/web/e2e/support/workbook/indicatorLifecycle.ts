import type { AppendIndicatorStateIntervalRequest } from "@cartulary/protocol-ts/http";
import {
  gridShellTestId,
  indicatorLifecycleTestId,
  rowCellTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { indicatorsViewSchemaId } from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";
import { createIncident } from "../incidents/fixtures";
import { apiBase } from "../runtime/configuration";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";
import { createViewRow } from "./query";
import { activateCommittedGridCell } from "./rowMutations";
export async function createLifecycleFixture(page: Page) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("INTERVALS"),
    "Indicator lifecycle intervals",
  );
  const indicator = await createViewRow(
    page,
    incidentId,
    indicatorsViewSchemaId,
    {
      client_txn_id: uniqueTxn("indicator"),
      "indicator.indicator_type": "ipv4_addr",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "198.51.100.24",
    },
  );
  return { incidentId, indicator };
}
export async function openLifecycleEditor(
  page: Page,
  incidentId: string,
  recordId: string,
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
) {
  await navigate(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(indicatorsViewSchemaId)}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(indicatorsViewSchemaId)),
  ).toBeVisible();
  await page
    .getByTestId(workbookInspectorToggleTestId(indicatorsViewSchemaId))
    .click();
  const cell = page
    .getByTestId(rowCellTestId(recordId, "indicator.indicator_type"))
    .locator("xpath=ancestor::*[@role='gridcell'][1]");
  await activateCommittedGridCell(cell);
  const action = page.getByTestId(
    workbookInspectorFeatureActionTestId(
      indicatorsViewSchemaId,
      "indicator.lifecycle.manage",
    ),
  );
  await action.focus();
  await action.press("Enter");
  const editor = page.getByTestId(indicatorLifecycleTestId("editor"));
  await expect(editor).toBeVisible();
  return editor;
}
export async function listLifecycleIntervals(page: Page, recordId: string) {
  const result = await publicHttpOperation({
    operationID: "listIndicatorStateIntervals",
    pathParameters: { indicator_id: recordId },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!result.ok) throw new Error(`Interval list failed: ${result.status}`);
  return result.payload;
}
export async function appendLifecycleInterval(
  page: Page,
  recordId: string,
  body: AppendIndicatorStateIntervalRequest,
) {
  const result = await publicHttpOperation({
    operationID: "appendIndicatorStateInterval",
    pathParameters: { indicator_id: recordId },
    body,
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!result.ok) throw new Error(`Interval append failed: ${result.status}`);
  return result.payload.data;
}
