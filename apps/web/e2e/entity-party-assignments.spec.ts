import {
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridShellTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  partiesViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { selectReferenceCandidates } from "./support/workbook/references";
import { openGenericInspectorForRecord } from "./support/workbook/rowMutations";

test("Host and Identity Party assignments use ordinary reference selection and preserve add and remove drafts", async ({
  page,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("EPA"),
    "Entity Party assignments",
  );
  const parties = [];
  for (const name of ["Assigned organization", "Additional organization"]) {
    parties.push(
      await createViewRow(page, incident, partiesViewSchemaId, {
        client_txn_id: uniqueTxn("assignment-party"),
        "party.display_name": name,
        "party.party_kind": "organization",
      }),
    );
  }
  const assigned = parties[0];
  const additional = parties[1];
  if (!assigned || !additional) throw new Error("Missing assignment fixtures");
  for (const [view, family] of [
    [hostsViewSchemaId, "host"],
    [identitiesViewSchemaId, "identity"],
  ] as const) {
    const row = await createViewRow(page, incident, view, {
      client_txn_id: uniqueTxn("assignment-entity"),
      [`${family}.display_name`]: `Assigned ${family}`,
    });
    await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
    await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
    await openGenericInspectorForRecord(page, view, row.record_id);
    const panel = page.getByTestId(
      workbookInspectorPanelTestId(view, "details"),
    );
    const input = page.getByTestId(genericEditValueTestId(view));
    const submit = page.getByTestId(genericEditSubmitTestId(view));
    const current = async () => {
      const rows = await queryViewRows(page, incident, view);
      const entity = rows.find(
        (candidate) => candidate.record_id === row.record_id,
      );
      if (!entity) throw new Error("Assigned entity missing from query");
      return entity;
    };
    let version = 1;
    for (const kind of ["subject", "responsible", "observer"] as const) {
      const field = `${family}.${kind}_party_ids`;
      await panel.locator(`[data-inspector-edit-field="${field}"]`).click();
      await selectReferenceCandidates(
        page,
        genericEditValueTestId(view),
        assigned.record_id,
      );
      await expect(input).toBeFocused();
      await expect(input).toHaveValue(assigned.record_id);
      await submit.click();
      version++;
      await expect
        .poll(async () => (await current()).row_version)
        .toBe(version);
      expect((await current()).cells[field]?.value).toMatchObject({
        kind: "collection_value_v1",
        ordered: false,
        items: [
          {
            party_id: assigned.record_id,
            item_ref: `party_ref:${assigned.record_id}`,
            display_text: "Assigned organization",
          },
        ],
      });
    }
    const field = `${family}.observer_party_ids`;
    await selectReferenceCandidates(
      page,
      genericEditValueTestId(view),
      additional.record_id,
    );
    await panel
      .getByRole("button", { name: "Remove parties", exact: true })
      .click();
    expect((await current()).row_version).toBe(version);
    await input.selectOption(`party_ref:${assigned.record_id}`);
    await panel
      .getByRole("button", { name: "Add parties", exact: true })
      .click();
    await expect(input).toHaveValue(additional.record_id);
    await panel
      .getByRole("button", { name: "Remove parties", exact: true })
      .click();
    await expect(input).toHaveValue(`party_ref:${assigned.record_id}`);
    await expect(submit).toBeDisabled();
    await panel
      .getByRole("button", { name: "Resume draft", exact: true })
      .click();
    await expect(input).toBeFocused();
    await info.attach(`${family}-assignment-remove`, {
      body: await panel
        .getByRole("group", {
          name: "Unsaved change: Observer Parties",
          exact: true,
        })
        .screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
    await submit.click();
    version++;
    await expect.poll(async () => (await current()).row_version).toBe(version);
    expect((await current()).cells[field]?.value).toMatchObject({ items: [] });
    await panel
      .getByRole("button", { name: "Add parties", exact: true })
      .click();
    await expect(input).toHaveValue(additional.record_id);
    await expect(submit).toBeDisabled();
    await panel
      .getByRole("button", {
        name: "Review retained actions for Observer Parties",
        exact: true,
      })
      .click();
    await expect(input).toBeFocused();
    await submit.click();
    version++;
    await expect.poll(async () => (await current()).row_version).toBe(version);
    await page.reload();
    await openGenericInspectorForRecord(page, view, row.record_id);
    await expect(panel).toContainText("Additional organization");
    expect((await current()).cells[field]?.value).toMatchObject({
      items: [{ party_id: additional.record_id }],
    });
  }
});
