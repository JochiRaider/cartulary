import { gridRowTestId } from "@cartulary/ui-contracts";
import {
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow, patchRecord, queryViewRows } from "./query";

export async function literalSetFixture(
  page: Page,
  view: string = timelineViewSchemaId,
) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("LITERAL-SET"),
    "Literal set membership",
  );
  const field =
    view === notesViewSchemaId
      ? "note.title"
      : "timeline.activity_synopsis_text";
  const tags = view === notesViewSchemaId ? "note.tags" : "timeline.tags";
  const rows = [];
  for (const [label, values] of [
    ["A literal", ["review,priority"]],
    ["B component", ["priority"]],
    ["C separate", ["review", "priority"]],
  ] as const) {
    let row = await createViewRow(page, incident, view, {
      client_txn_id: uniqueTxn("literal-row"),
      [field]: label,
    });
    if (label !== "A literal" || view === notesViewSchemaId)
      row = await patchRecord(page, row.record_id, {
        client_txn_id: uniqueTxn("literal-tags"),
        base_row_version: row.row_version,
        view_schema_id: view,
        changes: [
          {
            field_key: tags,
            action_payload: {
              kind: "collection_actions_v1",
              actions: [
                { op: "add_tag", tag_name: values[0] },
                ...values
                  .slice(1)
                  .map((tag_name) => ({ op: "add_tag" as const, tag_name })),
              ],
            },
          },
        ],
      });
    rows.push(row);
  }
  const first = rows[0];
  if (!first) throw new Error("Missing literal fixture record");
  await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
  await expect(
    page.getByTestId(gridRowTestId(view, first.record_id)),
  ).toBeVisible();
  if (view === timelineViewSchemaId) {
    await page
      .getByTestId(gridRowTestId(view, first.record_id))
      .getByRole("checkbox")
      .check();
    await page
      .getByRole("textbox", { name: "Tag for selected Timeline records" })
      .fill("review,priority");
    await page.getByRole("button", { name: "Assign tag", exact: true }).click();
    await expect
      .poll(async () =>
        (
          await queryViewRows(page, incident, view, {
            filters: [
              {
                field_key: tags,
                op: "contains_all",
                arg: { values: ["review,priority"] },
              },
            ],
          })
        ).map((row) => row.record_id),
      )
      .toEqual([first.record_id]);
    await page
      .getByTestId(gridRowTestId(view, first.record_id))
      .getByRole("checkbox")
      .uncheck();
  }
  return { incident, rows, tags, field };
}

export async function authorLiteralMembers(
  scope: Page | Locator,
  values: readonly string[],
  label = "Value",
) {
  await scope
    .getByRole("textbox", { name: `${label} 1`, exact: true })
    .fill(values[0] ?? "");
  for (const [index, value] of values.entries()) {
    if (index === 0) continue;
    await scope.getByRole("button", { name: "Add value", exact: true }).click();
    await scope
      .getByRole("textbox", { name: `${label} ${index + 1}`, exact: true })
      .fill(value);
  }
}

export async function expectLiteralMembership(
  page: Page,
  view: string,
  ids: readonly string[],
  all: readonly string[],
) {
  for (const id of all) {
    const row = page.getByTestId(gridRowTestId(view, id));
    if (ids.includes(id)) await expect(row).toBeVisible();
    else await expect(row).toHaveCount(0);
  }
}
