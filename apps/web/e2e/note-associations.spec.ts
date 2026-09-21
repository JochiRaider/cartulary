import {
  cartularyDesignPresentation,
  genericEditSubmitTestId,
  genericEditValueTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorPanelTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  notesViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { createViewRow } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { openGenericInspectorForRecord } from "./support/workbook/rowMutations";

test("Notes manage existing sources evidence and directional references with retained exact replay", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("NASSOC"),
    "Note association inspector",
  );
  const source = await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "host.display_name": "Original host source",
  });
  const linked = await publicHttpOperation({
    request: atJsonOrigin(page.request, apiBase),
    headers: await csrfHeaders(page),
    operationID: "createRecordLinkedNote",
    pathParameters: { record_id: source.record_id },
    body: {
      client_txn_id: uniqueTxn("linked"),
      "note.title": "Inspected Note",
    },
  });
  if (!linked.ok) throw new Error(JSON.stringify(linked.payload));
  const note = linked.payload.data.row;
  const other = await createViewRow(page, incident, notesViewSchemaId, {
    client_txn_id: uniqueTxn("other"),
    "note.title": "Related Note",
  });
  const evidence = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("evidence"),
    "evidence.title": "Associated evidence",
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${notesViewSchemaId}`,
  );
  await openGenericInspectorForRecord(page, notesViewSchemaId, note.record_id);
  const sources = page.getByRole("region", {
    name: "Note sources",
    exact: true,
  });
  await expect(sources).toContainText("Original host source");
  await expect(
    sources.getByRole("button", { name: "Remove Original host source" }),
  ).toBeEnabled();
  const related = page.getByRole("region", {
    name: "Note related notes",
    exact: true,
  });
  await related
    .getByRole("button", { name: "Manage related notes", exact: true })
    .click();
  await related
    .getByRole("listbox", { name: "Related notes" })
    .selectOption(other.record_id);
  await related.getByRole("button", { name: "Link selected records" }).click();
  await expect(related).toContainText("Associations saved.");
  await related
    .getByRole("button", { name: "Related Note", exact: true })
    .click();
  const incoming = page.getByRole("region", {
    name: "Referenced by",
    exact: true,
  });
  await expect(
    incoming.getByRole("button", { name: "Inspected Note" }),
  ).toBeVisible();
  await expect(incoming.getByRole("button", { name: /Remove/ })).toHaveCount(0);
  await incoming.getByRole("button", { name: "Inspected Note" }).click();
  await expect(sources).toContainText("Original host source");
  const routePath = `**/api/v1/records/${note.record_id}/note-associations*`;
  const bodies: string[] = [];
  let failReads = false;
  await page.route(routePath, async (route) => {
    if (route.request().method() === "GET") {
      if (failReads)
        return route.fulfill({
          status: 503,
          json: {
            error: {
              code: "internal_error",
              message: "Read unavailable",
              request_id: "read-failure",
              status: 503,
              retryable: true,
            },
          },
        });
      return route.continue();
    }
    bodies.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    if (bodies.length === 1) {
      failReads = true;
      await route.abort("failed");
    } else await route.fulfill({ response });
  });
  const panel = page.getByRole("region", {
    name: "Note evidence",
    exact: true,
  });
  await panel
    .getByRole("button", { name: "Manage evidence", exact: true })
    .click();
  await page
    .getByRole("region", { name: "Note evidence", exact: true })
    .getByRole("listbox", { name: "Evidence" })
    .selectOption(evidence.record_id);
  await panel.getByRole("button", { name: "Link selected records" }).click();
  await expect(panel).toContainText("outcome is unconfirmed");
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(notesViewSchemaId))
    .click();
  await openRecoveryItem(page, /^Note evidence associations ·/);
  const recovery = page.getByRole("region", {
    name: "Recovery navigation",
    exact: true,
  });
  await recovery
    .getByRole("button", { name: "Recover submission", exact: true })
    .click();
  await expect(recovery).toContainText(
    "Associations saved. Reads need refresh.",
  );
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).toBe(bodies[0]);
  await expect(
    recovery.getByRole("button", { name: "Retry refresh", exact: true }),
  ).toBeEnabled();
  failReads = false;
  await recovery
    .getByRole("button", { name: "Retry refresh", exact: true })
    .click();
  await expect(
    recovery.getByRole("button", { name: "Retry refresh", exact: true }),
  ).toHaveCount(0);
  expect(bodies).toHaveLength(2);
  await recovery.getByRole("button", { name: /Close recovery/i }).click();
  await openGenericInspectorForRecord(page, notesViewSchemaId, note.record_id);
  await expect(panel).toContainText("Associated evidence");
  await test.info().attach("note-associations-populated-review", {
    body: await page.screenshot({ animations: "disabled" }),
    contentType: "image/png",
  });
  await page.unroute(routePath);
  await panel
    .getByRole("button", { name: "Remove Associated evidence" })
    .click();
  await expect(panel).toContainText("No evidence associated with this Note.");
});

test("Inspector field rejection retains raw text and caret and cannot invalidate newer authoring", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("IFIELD"),
    "Inspector local feedback",
  );
  const row = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("feedback"),
    "evidence.title": "Evidence title",
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${evidenceViewSchemaId}`,
  );
  await openGenericInspectorForRecord(
    page,
    evidenceViewSchemaId,
    row.record_id,
  );
  await page
    .locator(`[data-inspector-edit-field="${"evidence.title"}"]`)
    .click();
  const input = page.getByTestId(genericEditValueTestId(evidenceViewSchemaId));
  let release: () => void = () => {};
  let reached = false;
  let responseGate = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route(`**/api/v1/records/${row.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    reached = true;
    await responseGate;
    await route.fulfill({
      status: 422,
      headers: { "X-Request-ID": "field-rejected" },
      json: {
        error: {
          code: "invalid_mutation_payload",
          status: 422,
          message: "Field rejected",
          request_id: "field-rejected",
          retryable: false,
          details: { field: "evidence.title", reason_code: "required" },
        },
      },
    });
  });
  await input.fill("  exact draft Ω  ");
  await page.getByTestId(genericEditSubmitTestId(evidenceViewSchemaId)).click();
  await expect.poll(() => reached).toBe(true);
  await input.focus();
  await input.evaluate((element) =>
    (element as HTMLInputElement).setSelectionRange(3, 7),
  );
  release();
  await expect(input).toHaveAttribute("aria-invalid", "true");
  await expect(page.getByRole("alert").filter({ hasText: /\S/ })).toHaveCount(
    1,
  );
  const details = page.getByTestId(
    workbookInspectorPanelTestId(evidenceViewSchemaId, "details"),
  );
  await expect(details.getByRole("alert")).toHaveText("required");
  await expect(input).toHaveValue("  exact draft Ω  ");
  await expect(input).toBeFocused();
  expect(
    await input.evaluate((element) => [
      (element as HTMLInputElement).selectionStart,
      (element as HTMLInputElement).selectionEnd,
    ]),
  ).toEqual([3, 7]);
  const description = await input.getAttribute("aria-describedby");
  expect(await details.getByRole("alert").getAttribute("id")).toBe(description);
  await input.fill("next revision");
  await expect(input).not.toHaveAttribute("aria-invalid", "true");
  reached = false;
  responseGate = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.getByTestId(genericEditSubmitTestId(evidenceViewSchemaId)).click();
  await expect.poll(() => reached).toBe(true);
  await input.fill("newer revision while waiting");
  release();
  await expect(
    page.getByTestId(genericEditSubmitTestId(evidenceViewSchemaId)),
  ).toBeEnabled();
  await expect(input).not.toHaveAttribute("aria-invalid", "true");
  await expect(input).toHaveValue("newer revision while waiting");
});

test("Inspector collection and reference rejections retain authored values and selected identities", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ICOLL"),
    "Inspector collection feedback",
  );
  const note = await createViewRow(page, incident, notesViewSchemaId, {
    client_txn_id: uniqueTxn("note"),
    "note.title": "Retained reference Note",
  });
  const task = await createViewRow(page, incident, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("task"),
    "task.title": "Reference authoring",
    "task.task_kind": "follow_up",
  });
  for (const [schema, record, field] of [
    [notesViewSchemaId, note.record_id, "note.tags"],
    [taskRequestsViewSchemaId, task.record_id, "task.linked_record_ids"],
  ] as const) {
    await page.goto(`/?incident_id=${incident}&view_schema_id=${schema}`);
    await openGenericInspectorForRecord(page, schema, record);
    await page.locator(`[data-inspector-edit-field="${field}"]`).click();
    const input = page.getByTestId(genericEditValueTestId(schema));
    const details = page.getByTestId(
      workbookInspectorPanelTestId(schema, "details"),
    );
    if (field === "note.tags") await input.fill("  retained-tag  ");
    else {
      await details
        .getByRole("button", { name: "Choose linked records", exact: true })
        .click();
      const picker = page.getByRole("dialog", {
        name: "Choose linked records",
      });
      await picker
        .getByRole("combobox", { name: "Reference surface" })
        .selectOption(notesViewSchemaId);
      const options = picker.getByRole("listbox", {
        name: "Linked Records candidates",
      });
      await expect(
        options.getByRole("option", { name: /Retained reference Note/ }),
      ).toHaveCount(1);
      await options.selectOption({
        label: `Retained reference Note (${note.record_id})`,
      });
      await picker.getByRole("button", { name: "Use selection" }).click();
    }
    const authored = await input.inputValue();
    await page.route(`**/api/v1/records/${record}`, async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      await route.fulfill({
        status: 422,
        json: {
          error: {
            code: "invalid_mutation_payload",
            status: 422,
            message: "Field rejected",
            request_id: "collection-rejected",
            retryable: false,
            details: { field, reason_code: "invalid_value" },
          },
        },
      });
    });
    await page.getByTestId(genericEditSubmitTestId(schema)).click();
    await expect(input).toHaveAttribute("aria-invalid", "true");
    await expect(input).toHaveValue(authored);
    await expect(details.getByRole("alert")).toHaveText("invalid_value");
    expect(await input.getAttribute("aria-describedby")).toBe(
      await details.getByRole("alert").getAttribute("id"),
    );
    if (field === "task.linked_record_ids") {
      await details
        .getByRole("button", { name: "Choose linked records", exact: true })
        .click();
      const picker = page.getByRole("dialog", {
        name: "Choose linked records",
      });
      await expect(
        picker.getByRole("button", {
          name: "Remove selected Retained reference Note",
          exact: true,
        }),
      ).toHaveCount(1);
    }
  }
});

test("Inspector header remains reachable while the body scrolls at supported narrow zoom and spacing", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("IHEAD"),
    "Inspector header geometry",
  );
  const label = `Long evidence record ${"context ".repeat(35)}`;
  const row = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("header"),
    "evidence.title": label,
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${evidenceViewSchemaId}`,
  );
  await openGenericInspectorForRecord(
    page,
    evidenceViewSchemaId,
    row.record_id,
  );
  const shell = page.locator(
      `aside[data-view-schema-id="${evidenceViewSchemaId}"]`,
    ),
    body = shell.locator("[data-inspector-scroll-body]");
  const close = page.getByTestId(
    workbookInspectorCloseButtonTestId(evidenceViewSchemaId),
  );
  await shell.getByRole("button", { name: "Edit Title", exact: true }).click();
  await page
    .getByTestId(genericEditValueTestId(evidenceViewSchemaId))
    .fill("Unfinished title");
  await shell
    .getByRole("button", { name: "Close editor", exact: true })
    .click();
  await expect(
    shell.getByRole("button", { name: "Unfinished work (1)", exact: true }),
  ).toBeVisible();
  for (const [width, height, zoom, spacing] of [
    [1280, 720, 1, false],
    [1024, 720, 1, false],
    [768, 640, 1, false],
    [320, 640, 1, false],
    [1280, 720, 2, false],
    [768, 480, 1, true],
    [320, 640, 1, true],
  ] as const) {
    await page.setViewportSize({ width, height });
    await page.evaluate(
      ({ zoom, spacing }) => {
        document.documentElement.style.zoom = String(zoom);
        document.getElementById("inspector-text-spacing-test")?.remove();
        if (spacing) {
          const sheet = document.createElement("style");
          sheet.id = "inspector-text-spacing-test";
          sheet.textContent = `[data-inspector-state] * { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } [data-inspector-state] p { margin-block-end: 2em !important; }`;
          document.head.append(sheet);
        }
      },
      { zoom, spacing },
    );
    await body.evaluate((element) => {
      element.scrollTop = element.scrollHeight;
    });
    await expect(close).toBeInViewport();
    await expect(shell.locator("header h2")).toBeInViewport();
    const bounds = await shell.evaluate((element) => {
      const header = element.querySelector("header"),
        body = element.querySelector("[data-inspector-scroll-body]");
      if (!header || !body) throw new Error("Missing layout");
      return {
        headerHeight: header.getBoundingClientRect().height,
        inspectorWidth: element.getBoundingClientRect().width,
        horizontalOverflow: body.scrollWidth - body.clientWidth,
        documentOverflow:
          document.documentElement.scrollWidth -
          document.documentElement.clientWidth,
        headerBottom: header.getBoundingClientRect().bottom,
        bodyTop: body.getBoundingClientRect().top,
        outerScroll: element.scrollTop,
        bodyScroll: body.scrollTop,
      };
    });
    expect(bounds.horizontalOverflow).toBeLessThanOrEqual(1);
    expect(bounds.documentOverflow).toBeLessThanOrEqual(1);
    if (width === 1280 && zoom === 1) {
      expect(bounds.inspectorWidth).toBe(
        cartularyDesignPresentation.inspector.referenceViewport
          .inspector_width_px,
      );
      expect(bounds.headerHeight).toBeLessThanOrEqual(
        cartularyDesignPresentation.inspector.persistentRegionMaxHeightPx,
      );
      await expect(
        shell.getByRole("button", { name: "Sections", exact: true }),
      ).toHaveCount(0);
    }
    expect(bounds.outerScroll).toBe(0);
    expect(bounds.bodyTop).toBeGreaterThanOrEqual(bounds.headerBottom - 1);
    expect(bounds.bodyScroll).toBeGreaterThan(0);
    const action = shell.getByTestId(
      workbookInspectorFeatureActionTestId(
        evidenceViewSchemaId,
        "create_related.note",
      ),
    );
    await close.focus();
    await action.focus();
    await expect(action).toBeInViewport();
    await expect(close).toBeInViewport();
    await test
      .info()
      .attach(`inspector-header-${width}-${height}-${zoom}-${spacing}`, {
        body: await page.screenshot({ animations: "disabled" }),
        contentType: "image/png",
      });
  }
  await page.evaluate(() => {
    document.documentElement.style.zoom = "1";
  });
  await close.click();
  await expect(shell).toHaveCount(0);
  const opener = page.getByTestId(
    workbookInspectorToggleTestId(evidenceViewSchemaId),
  );
  await expect(opener).toBeInViewport();
  await opener.click();
  await expect(close).toBeInViewport();
  await expect(
    page.getByRole("region", { name: "Primary grid", includeHidden: true }),
  ).toHaveAttribute("inert", "");
  await close.click();
});
