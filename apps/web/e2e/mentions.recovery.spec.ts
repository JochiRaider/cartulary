import type {
  CreateViewRowResponse,
  ResolveEntityMentionResponse,
} from "@cartulary/protocol-ts/http";
import {
  gridShellTestId,
  mentionItemTestId,
  surfaceTabTestId,
  workbookShellReadyTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  collectionActionsPayload,
  collectionItems,
  hostRefsFieldKey,
  publicEntityMentionId,
  requireItemByRawText,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";
import { openTimelineInspector } from "./support/workbook/rowMutations";

test("Mention creation survives rejected resolution and sheet navigation without another creation", async ({
  page,
}, testInfo) => {
  await runRecovery("rejected", page, testInfo);
});
test("Mention creation survives uncertain resolution and sheet navigation without another creation", async ({
  page,
}, testInfo) => {
  await runRecovery("uncertain", page, testInfo);
});
async function runRecovery(
  outcome: "rejected" | "uncertain",
  page: Page,
  testInfo: TestInfo,
) {
  await page.setViewportSize({
    width: outcome === "uncertain" ? 768 : 1280,
    height: 800,
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("MENTION-RECOVERY"),
    "Mention creation recovery",
  );
  const rawText = "  reviewed@host.example  ";
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "timeline.activity_synopsis_text": "Create and resolve recovery",
    [hostRefsFieldKey]: collectionActionsPayload([rawText]),
  });
  const mention = requireItemByRawText(
    collectionItems(source, hostRefsFieldKey),
    rawText,
  );
  const mentionId = publicEntityMentionId(mention);
  const creates: string[] = [],
    resolves: string[] = [],
    receipts: ResolveEntityMentionResponse[] = [];
  let created: CreateViewRowResponse | null = null;
  let failRefresh = false;
  let failEntityRefresh = false;
  await page.route(`**/views/${hostsViewSchemaId}/query`, async (route) => {
    if (failEntityRefresh) await route.abort("failed");
    else await route.continue();
  });
  await page.route(`**/views/${timelineViewSchemaId}/query`, async (route) => {
    if (failRefresh) await route.abort("failed");
    else await route.continue();
  });
  await page.route(`**/views/${hostsViewSchemaId}/rows`, async (route) => {
    creates.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    created = await response.json();
    failEntityRefresh = outcome === "uncertain";
    await route.fulfill({ response });
  });
  await page.route(`**/entity-mentions/${mentionId}/resolve`, async (route) => {
    resolves.push(route.request().postData() ?? "");
    if (outcome === "rejected" && resolves.length === 1) {
      const original = route.request().postDataJSON();
      const remote = await page.request.post(route.request().url(), {
        headers: await csrfHeaders(page),
        data: {
          ...original,
          client_txn_id: uniqueTxn("remote-resolution"),
        },
      });
      expect(remote.status()).toBe(200);
    }
    const response = await route.fetch();
    if (outcome === "rejected" && resolves.length === 1)
      expect(response.status()).toBe(409);
    else {
      expect(response.status()).toBe(200);
      receipts.push(await response.json());
    }
    if (outcome === "uncertain" && resolves.length === 1)
      await route.abort("failed");
    else {
      if (outcome === "uncertain") failRefresh = true;
      await route.fulfill({ response });
    }
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openTimelineInspector(page, source.record_id);
  await page.getByTestId(mentionItemTestId(String(mention.item_ref))).click();
  await page.getByRole("button", { name: "Create host", exact: true }).click();
  await expect(
    page.getByRole("textbox", { name: "Display Name value", exact: true }),
  ).toHaveValue(rawText);
  await expect(
    page.getByRole("textbox", { name: "Hostname value", exact: true }),
  ).toHaveValue("");
  const correction = page.getByText("Correction and resolution", {
    exact: true,
  });
  await correction.click();
  await expect(
    page.getByText(
      "Unfinished entity creation is retained. Open Correction and resolution to continue.",
      { exact: true },
    ),
  ).toBeVisible();
  await expect(
    page.getByRole("textbox", { name: "Display Name value", exact: true }),
  ).not.toBeVisible();
  expect(creates).toHaveLength(0);
  await correction.click();
  await expect(
    page.getByRole("textbox", { name: "Display Name value", exact: true }),
  ).toHaveValue(rawText);
  const submit = page.getByRole("button", {
    name: "Create host and resolve",
    exact: true,
  });
  await testInfo.attach("reviewed-mention-create-form", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await submit.focus();
  await submit.press("Enter");
  await expect(
    page.getByRole("region", { name: "Entity created from mention" }),
  ).toContainText("Host saved:");
  await expect.poll(() => resolves.length).toBe(1);
  if (outcome === "uncertain") {
    await expect(
      page.getByRole("region", { name: "Mention operation status" }),
    ).toContainText("The outcome is uncertain.");
    await correction.click();
    await expect(
      page.getByRole("button", { name: "Replay mention action", exact: true }),
    ).toBeVisible();
    expect(resolves).toHaveLength(1);
    await page.getByTestId(workbookSurfacesMenuTriggerTestId()).click();
    await page
      .getByTestId(workbookSurfacesMenuOptionTestId(hostsViewSchemaId))
      .click();
  } else await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  if (outcome === "rejected") {
    await expect(
      page.getByTestId(gridShellTestId(hostsViewSchemaId)),
    ).toContainText(rawText.trim());
  }
  const trigger = recoveryEntry(page);
  await openRecoveryItem(page, /^Mention creation and resolution ·/);
  const recovery = page.getByRole("region", {
    name: "Retained mention operations",
  });
  await expect(recovery).toContainText("Entity saved:");
  if (outcome === "uncertain") {
    await expect(recovery).toContainText(
      "Entity saved; its sheet still needs refreshing.",
    );
    failEntityRefresh = false;
    await recovery
      .getByRole("button", { name: "Refresh created entity" })
      .click();
    await expect(
      recovery.getByRole("region", { name: "Created entity refresh" }),
    ).not.toBeVisible();
    await expect(
      page.getByTestId(gridShellTestId(hostsViewSchemaId)),
    ).toContainText(rawText.trim());
    expect(creates).toHaveLength(1);
    expect(resolves).toHaveLength(1);
  }
  const bounds = await recovery.boundingBox();
  expect(bounds?.x).toBeGreaterThanOrEqual(0);
  expect((bounds?.x ?? 0) + (bounds?.width ?? 0)).toBeLessThanOrEqual(
    page.viewportSize()?.width ?? 0,
  );
  await testInfo.attach("retained-mention-recovery", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  if (outcome === "uncertain") {
    await expect(recovery).toContainText("The outcome is uncertain.");
    await recovery
      .getByRole("button", { name: "Replay mention action" })
      .click();
    await expect(recovery).toContainText(
      "Mention action completed; refresh is still required.",
    );
    failRefresh = false;
    await recovery
      .getByRole("button", { name: "Refresh mention result" })
      .click();
    await expect(recovery).toContainText("Mention action completed.");
    expect(resolves).toHaveLength(2);
    expect(resolves[1]).toBe(resolves[0]);
    expect(receipts[1]?.data).toEqual(receipts[0]?.data);
    await recovery.press("Escape");
    await expect(recovery).not.toBeVisible();
    await expect(trigger).toBeFocused();
  } else {
    await openRecoveryItem(page, /^Mention creation and resolution ·/);
    await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
    await openTimelineInspector(page, source.record_id);
    await page.getByTestId(mentionItemTestId(String(mention.item_ref))).click();
    await page
      .getByRole("button", { name: "Resolve to created host", exact: true })
      .click();
    await expect.poll(() => resolves.length).toBe(2);
    await expect.poll(() => receipts.length).toBe(1);
    await expect(
      page.getByRole("region", { name: "Mention operation status" }),
    ).toContainText("Mention action completed.");
    expect(JSON.parse(resolves[1] ?? "{}").base_mention_row_version).toBe(
      Number(mention.mention_row_version) + 1,
    );
  }
  expect(creates).toHaveLength(1);
  const entities = await queryViewRows(page, incidentId, hostsViewSchemaId);
  expect(entities).toHaveLength(1);
  const saved = created as CreateViewRowResponse | null;
  expect(entities[0]?.record_id).toBe(saved?.data.row.record_id);
  expect(entities[0]?.cells["host.hostname"]?.value).toBeNull();
  await testInfo.attach("mention-create-recovery-receipts", {
    body: JSON.stringify({ creates, resolves, created, receipts }),
    contentType: "application/json",
  });
}
