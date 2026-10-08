import {
  genericCreateSubmitTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import {
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { openContextualCreationFixture } from "./support/workbook/contextualCreate";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openGenericInspectorForRecord } from "./support/workbook/rowMutations";

const duplicateLabel =
  "Candidate with a shared, deliberately long prefix: authentication, process and network observations need comparison before linking. Complete wording remains available before selection, including this final distinguishing context.";

async function candidates(page: Page) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("MULTI-REF"),
    "Explicit references",
  );
  const rows = [];
  for (const label of ["Source S", duplicateLabel, duplicateLabel]) {
    rows.push(
      await createViewRow(page, incident, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("candidate"),
        "timeline.activity_synopsis_text": label,
      }),
    );
  }
  const [source, a, b] = rows;
  if (!source || !a || !b)
    throw new Error("Missing synthetic reference candidates");
  return { incident, source, a, b };
}

// An actual unmodified pointer gesture, addressed by the contract identity.
function choice(picker: Locator, key: string) {
  return picker.locator(`input[type="checkbox"][value="${key}"]`);
}

async function inspectCandidates(
  page: Page,
  picker: Locator,
  keys: readonly string[],
) {
  for (const key of keys) {
    const input = choice(picker, key);
    await expect(input).toHaveAccessibleName(
      `${duplicateLabel} (${key.replace(/^record:/, "")})`,
    );
    const row = input.locator("..");
    await row.scrollIntoViewIfNeeded();
    await expect(row).toContainText(duplicateLabel);
    await expect(row).toContainText(key.replace(/^record:/, ""));
    expect(
      await row.evaluate(
        (element) => element.scrollWidth <= element.clientWidth + 1,
      ),
    ).toBe(true);
  }
  expect(
    await page.evaluate(
      () =>
        document.documentElement.scrollWidth <=
        document.documentElement.clientWidth + 1,
    ),
  ).toBe(true);
  const scroll = picker.locator("[data-workbook-multi-candidates] > div");
  expect(
    await scroll.evaluate((element) => element.clientHeight),
  ).toBeLessThanOrEqual(193);
  for (const action of await picker
    .getByRole("button", {
      name: /^(Use selection|Apply references|Cancel references)$/,
    })
    .all()) {
    await action.scrollIntoViewIfNeeded();
    await expect(action).toBeInViewport();
  }
}

test("Existing Task plain candidate clicks stage additive references until explicit Update", async ({
  page,
}) => {
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 768 },
  ]) {
    await page.setViewportSize(viewport);
    const f = await candidates(page);
    const task = await createViewRow(
      page,
      f.incident,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("task"),
        "task.title": "Collect multiple references",
        "task.task_kind": "follow_up",
      },
    );
    const writes: unknown[] = [];
    const observe = (request: import("@playwright/test").Request) => {
      if (
        request.method() === "PATCH" &&
        request.url().endsWith(`/records/${task.record_id}`)
      )
        writes.push(request.postDataJSON());
    };
    page.on("request", observe);
    try {
      await page.goto(
        `/?incident_id=${f.incident}&view_schema_id=${taskRequestsViewSchemaId}`,
      );
      await openGenericInspectorForRecord(
        page,
        taskRequestsViewSchemaId,
        task.record_id,
      );
      await page
        .getByTestId(
          workbookInspectorPanelTestId(
            taskRequestsViewSchemaId,
            "relationships",
          ),
        )
        .getByRole("button", { name: "Manage Linked Records", exact: true })
        .click();
      const parent = page.getByTestId(
        genericEditValueTestId(taskRequestsViewSchemaId),
      );
      const choose = page
        .getByTestId(
          workbookInspectorPanelTestId(taskRequestsViewSchemaId, "details"),
        )
        .getByRole("button", { name: "Choose linked records", exact: true });
      const picker = page.getByRole("dialog", {
        name: "Choose linked records",
        exact: true,
      });
      await choose.click();
      await picker
        .getByLabel("Reference surface")
        .selectOption(timelineViewSchemaId);
      await inspectCandidates(
        page,
        picker,
        [f.a, f.b].map((row) => `record:${row.record_id}`),
      );
      await choice(picker, `record:${f.a.record_id}`).click();
      await choice(picker, `record:${f.b.record_id}`).click();
      await expect(picker).toContainText("Selected for this edit: 2.");
      await expect(parent).toHaveValue("");
      expect(writes).toHaveLength(0);
      await picker
        .getByRole("button", { name: "Cancel references", exact: true })
        .click();
      await expect(parent).toHaveValue("");
      await choose.click();
      await picker
        .getByLabel("Reference surface")
        .selectOption(timelineViewSchemaId);
      const first = choice(picker, `record:${f.a.record_id}`);
      const second = choice(picker, `record:${f.b.record_id}`);
      await first.focus();
      await page.keyboard.press("Space");
      await expect(first).toBeChecked();
      for (let steps = 0; steps < 30; steps++) {
        await page.keyboard.press("Tab");
        await expect(picker).toContainText("Selected for this edit: 1.");
        if (
          await second.evaluate((element) => element === document.activeElement)
        )
          break;
      }
      await expect(second).toBeFocused();
      await page.keyboard.press("Space");
      await expect(picker).toContainText("Selected for this edit: 2.");
      const accept = picker.getByRole("button", {
        name: "Use selection",
        exact: true,
      });
      for (let steps = 0; steps < 30; steps++) {
        await page.keyboard.press("Tab");
        await expect(picker).toContainText("Selected for this edit: 2.");
        if (
          await accept.evaluate((element) => element === document.activeElement)
        )
          break;
      }
      await expect(accept).toBeFocused();
      await page.keyboard.press("Enter");
      await expect
        .soft(parent)
        .toHaveValue(`${f.a.record_id}\n${f.b.record_id}`);
      expect(writes).toHaveLength(0);
      await page
        .getByTestId(genericEditSubmitTestId(taskRequestsViewSchemaId))
        .press("Enter");
      await expect(
        page.getByRole("region", { name: "Inspector changes" }),
      ).toContainText("Saved, version 2");
      expect(writes).toEqual([
        expect.objectContaining({
          changes: [
            expect.objectContaining({
              field_key: "task.linked_record_ids",
              action_payload: {
                kind: "collection_actions_v1",
                actions: [f.a, f.b].map((row) => ({
                  op: "add_record_ref",
                  linked_record_id: row.record_id,
                })),
              },
            }),
          ],
        }),
      ]);
      const saved = (
        await queryViewRows(page, f.incident, taskRequestsViewSchemaId)
      ).find((row) => row.record_id === task.record_id);
      const links = saved?.cells["task.linked_record_ids"]?.value as {
        items: { item_ref: string }[];
      };
      expect
        .soft(links.items.map((item) => item.item_ref).sort())
        .toEqual([f.a, f.b].map((row) => `record_ref:${row.record_id}`).sort());
    } finally {
      page.off("request", observe);
    }
  }
});

test("Contextual Task plain candidate clicks retain editable source context through Cancel and Apply", async ({
  page,
}) => {
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 768 },
  ]) {
    await page.setViewportSize(viewport);
    const f = await candidates(page);
    await openContextualCreationFixture(
      page,
      "task_request",
      (url) => page.goto(url),
      { viewSchemaId: timelineViewSchemaId, label: "Source S" },
      { incidentId: f.incident, source: f.source },
    );
    const form = page.getByRole("region", {
      name: "Create task request",
      exact: true,
    });
    const choose = form.getByRole("button", {
      name: "Choose Linked Records",
      exact: true,
    });
    const picker = form.getByRole("region", {
      name: "Choose Linked Records",
      exact: true,
    });
    await choose.click();
    await inspectCandidates(page, picker, [f.a.record_id, f.b.record_id]);
    await choice(picker, f.a.record_id).click();
    await choice(picker, f.b.record_id).click();
    await expect(picker).toContainText("3 selected (maximum 64).");
    await choice(picker, f.source.record_id).click();
    await expect(picker).toContainText("2 selected (maximum 64).");
    await expect(choice(picker, f.source.record_id)).not.toBeChecked();
    await picker
      .getByRole("button", { name: "Refresh candidates", exact: true })
      .click();
    await expect(
      picker.getByRole("button", { name: "Refresh candidates", exact: true }),
    ).toBeEnabled();
    await expect(choice(picker, f.source.record_id)).not.toBeChecked();
    await expect(picker).toContainText("2 selected (maximum 64).");
    expect(
      await queryViewRows(page, f.incident, taskRequestsViewSchemaId),
    ).toHaveLength(0);
    await picker
      .getByRole("button", { name: "Cancel references", exact: true })
      .click();
    await expect(
      form.getByRole("button", {
        name: "Remove Linked Records Source S",
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      form.getByRole("button", {
        name: `Remove Linked Records ${duplicateLabel}`,
        exact: true,
      }),
    ).toHaveCount(0);
    await choose.click();
    await choice(picker, f.a.record_id).click();
    await choice(picker, f.b.record_id).click();
    await expect(picker).toContainText("3 selected (maximum 64).");
    // Removing a staged reference remains an explicit local operation.
    await choice(picker, f.a.record_id).click();
    await picker
      .getByRole("button", { name: "Apply references", exact: true })
      .click();
    await form
      .getByTestId(genericCreateSubmitTestId(taskRequestsViewSchemaId))
      .click();
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, f.incident, taskRequestsViewSchemaId))
            .length,
      )
      .toBe(1);
    const saved = (
      await queryViewRows(page, f.incident, taskRequestsViewSchemaId)
    )[0];
    const links = saved?.cells["task.linked_record_ids"]?.value as {
      items: { item_ref: string }[];
    };
    expect
      .soft(links.items.map((item) => item.item_ref).sort())
      .toEqual(
        [f.source, f.b].map((row) => `record_ref:${row.record_id}`).sort(),
      );
  }
});
