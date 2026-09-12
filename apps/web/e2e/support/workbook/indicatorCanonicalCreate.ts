import { indicatorCreateTestId } from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";
import {
  createObservationFixture,
  openObservationEditor,
} from "./indicatorObservations";

export async function createCanonicalObservationFixture(page: Page) {
  const fixture = await createObservationFixture(page);
  const result = await publicHttpOperation({
    operationID: "createManualIndicatorObservation",
    pathParameters: { source_record_id: fixture.source.record_id },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
    body: {
      client_txn_id: uniqueTxn("canonical-origin"),
      base_row_version: fixture.source.row_version,
      source_field_key: "timeline.raw_activity_text",
      span_start_byte: 0,
      span_end_byte: 13,
    },
  });
  if (!result.ok)
    throw new Error(`Observation fixture failed: ${result.status}`);
  return { ...fixture, observation: result.payload.data.observation };
}
export async function openCanonicalProposal(
  page: Page,
  sourceId: string,
  value = "NEW[.]EXAMPLE",
) {
  const editor = await openObservationEditor(page, sourceId);
  const item = editor.getByRole("article", {
    name: "Observation: alpha.example",
    exact: true,
  });
  await item
    .getByRole("button", { name: "Create canonical Indicator…", exact: true })
    .click();
  const form = page.getByTestId(indicatorCreateTestId("editor"));
  await form
    .getByLabel("Indicator type", { exact: true })
    .selectOption("domain_name");
  await form.getByLabel("Value kind", { exact: true }).selectOption("atomic");
  await form.getByLabel("Canonical value", { exact: true }).fill(value);
  await expect(form.getByText("Original observed text:")).toContainText(
    "alpha.example",
  );
  return { editor, item, form };
}
export async function submitCanonicalProposal(form: Locator) {
  await form
    .getByRole("button", { name: "Create canonical Indicator", exact: true })
    .click();
}
