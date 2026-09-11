import type { ListIndicatorObservationsResponse } from "@cartulary/protocol-ts/http";
import {
  indicatorObservationTestId,
  workbookInspectorFeatureActionTestId,
} from "@cartulary/ui-contracts";
import {
  indicatorsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { apiBase } from "../runtime/configuration";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";
import { fetchFullRecordHistory } from "./history";
import { createViewRow, queryViewRows } from "./query";
import { openTimelineInspector } from "./rowMutations";

export const observationPrefix = "alpha.example\r\n雪😀 e\u0301 ";
export const observationRawText = `${observationPrefix}alpha.example\rend\n`;
export async function createObservationFixture(page: Page) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("OBSERVATIONS"),
    "Source-bound Indicator observations",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "timeline.raw_activity_text": observationRawText,
    "timeline.activity_synopsis_text": "Repeated source observation",
  });
  const oldTarget = await createViewRow(
    page,
    incidentId,
    indicatorsViewSchemaId,
    {
      client_txn_id: uniqueTxn("indicator"),
      "indicator.indicator_type": "domain_name",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "alpha.example",
    },
  );
  const newTarget = await createViewRow(
    page,
    incidentId,
    indicatorsViewSchemaId,
    {
      client_txn_id: uniqueTxn("indicator"),
      "indicator.indicator_type": "domain_name",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "beta.example",
    },
  );
  return { incidentId, source, oldTarget, newTarget };
}
export async function openObservationEditor(page: Page, sourceId: string) {
  await openTimelineInspector(page, sourceId);
  const action = page.getByTestId(
    workbookInspectorFeatureActionTestId(
      timelineViewSchemaId,
      "indicator.observations.manage",
    ),
  );
  await action.focus();
  await action.press("Enter");
  const editor = page.getByTestId(indicatorObservationTestId("editor"));
  await expect(editor).toBeVisible();
  await editor
    .getByLabel("Source field", { exact: true })
    .selectOption("timeline.raw_activity_text");
  return editor;
}
export async function selectRepeatedObservation(page: Page) {
  await expect(
    page.getByRole("button", { name: "Use selected text", exact: true }),
  ).toBeEnabled();
  const source = page.getByTestId(indicatorObservationTestId("source"));
  await source.evaluate((node, start) => {
    const control = node as HTMLTextAreaElement;
    control.focus();
    control.setSelectionRange(start, start + 13);
  }, observationPrefix.replaceAll("\r\n", "\n").length);
  await page
    .getByRole("button", { name: "Use selected text", exact: true })
    .click();
  await expect(
    page.getByTestId(indicatorObservationTestId("preview")),
  ).toContainText("alpha.example");
}
export async function listSourceObservations(page: Page, sourceId: string) {
  const response = await publicHttpOperation({
    operationID: "listSourceRecordIndicatorObservations",
    pathParameters: { source_record_id: sourceId },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok)
    throw new Error(`Source observations failed: ${response.status}`);
  return response.payload satisfies ListIndicatorObservationsResponse;
}
async function listTargetObservations(page: Page, targetId: string) {
  const response = await publicHttpOperation({
    operationID: "listIndicatorObservations",
    pathParameters: { indicator_id: targetId },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok)
    throw new Error(`Indicator observations failed: ${response.status}`);
  return response.payload;
}
export async function observationServiceSnapshot(
  page: Page,
  fixture: Awaited<ReturnType<typeof createObservationFixture>>,
) {
  const { incidentId, source, oldTarget, newTarget } = fixture;
  return {
    source: (await queryViewRows(page, incidentId, timelineViewSchemaId)).find(
      (row) => row.record_id === source.record_id,
    ),
    indicators: await queryViewRows(page, incidentId, indicatorsViewSchemaId),
    observations: (await listSourceObservations(page, source.record_id)).data
      .observations,
    oldTarget: (await listTargetObservations(page, oldTarget.record_id)).data
      .observations,
    newTarget: (await listTargetObservations(page, newTarget.record_id)).data
      .observations,
    history: await Promise.all(
      [source, oldTarget, newTarget].map((row) =>
        fetchFullRecordHistory(page, row.record_id),
      ),
    ),
  };
}
