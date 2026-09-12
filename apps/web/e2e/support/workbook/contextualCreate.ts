import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  workbookInspectorFeatureActionTestId,
} from "@cartulary/ui-contracts";
import {
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow } from "./query";
import { openGenericInspectorForRecord } from "./rowMutations";

/** Fixture-only data through the same public creation routes used by analysts. */
export async function openContextualCreationFixture(
  page: Page,
  target: "task_request" | "decision",
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("CTD-PRESENTATION"),
    "Contextual creation review",
  );
  const source = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "evidence.title": "Reviewed Evidence source",
  });
  await navigate(
    `/?incident_id=${incident}&view_schema_id=${evidenceViewSchemaId}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
  ).toBeVisible();
  await openGenericInspectorForRecord(
    page,
    evidenceViewSchemaId,
    source.record_id,
  );
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        evidenceViewSchemaId,
        `create_related.${target}`,
      ),
    )
    .click();
  const view =
    target === "task_request"
      ? taskRequestsViewSchemaId
      : decisionsViewSchemaId;
  await expect(page.getByTestId(genericCreateSubmitTestId(view))).toBeVisible();
  if (target === "task_request") {
    await page
      .getByTestId(genericCreateFieldTestId("task.title"))
      .fill("Review related investigation");
    await page
      .getByTestId(genericCreateFieldTestId("task.task_kind"))
      .selectOption("follow_up");
  } else {
    await page
      .getByTestId(genericCreateFieldTestId("decision.summary"))
      .fill("Retain the reviewed context");
    await page
      .getByTestId(genericCreateFieldTestId("decision.decision_type"))
      .selectOption("containment");
    await page
      .getByTestId(genericCreateFieldTestId("decision.rationale"))
      .fill("The selected Evidence supports this decision.");
  }
  return { incident, source, view };
}
