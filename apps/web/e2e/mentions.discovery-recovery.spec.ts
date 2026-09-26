import {
  mentionItemTestId,
  mentionResolveTargetSelectTestId,
  timelineInspectorTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import {
  collectionActionsPayload,
  collectionItems,
  hostRefsFieldKey,
  identityRefsFieldKey,
  requireItemByRawText,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

test("Timeline host target retry stays focused while its page is held", async ({
  page,
}) => {
  await runRetryFocus(page, "host");
});
test("Timeline identity target retry stays focused while its page is held", async ({
  page,
}) => {
  await runRetryFocus(page, "identity");
});

test("Timeline invalid continuation restarts without a cursor and keeps stale selection ineligible", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("MENTION-INVALID-CURSOR"),
    "Mention invalid cursor recovery",
  );
  const target = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("invalid-cursor-target"),
    "host.display_name": "Restart target host",
    "host.hostname": "restart-target.example.test",
  });
  const raw = "unresolved-restart-host";
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("invalid-cursor-source"),
    "timeline.activity_synopsis_text": "Restart mention target discovery",
    [hostRefsFieldKey]: collectionActionsPayload([raw]),
  });
  const mention = requireItemByRawText(
    collectionItems(source, hostRefsFieldKey),
    raw,
  );
  const cursors: (string | null)[] = [];
  let restartAttempted = false;
  let failRestart = false;
  let releaseRestart: (() => void) | undefined;
  const heldRestart = new Promise<void>((resolve) => {
    releaseRestart = resolve;
  });
  await page.route(`**/views/${hostsViewSchemaId}/query`, async (route) => {
    const cursor =
      (route.request().postDataJSON() as { cursor_token?: string })
        .cursor_token ?? null;
    cursors.push(cursor);
    if (cursor !== null) {
      const response = await route.fetch();
      expect(response.status()).toBe(400);
      await route.fulfill({ response });
      return;
    }
    if (failRestart) {
      failRestart = false;
      await heldRestart;
      await route.abort("failed");
      return;
    }
    const response = await route.fetch();
    const body = (await response.json()) as Record<string, unknown>;
    await route.fulfill({
      response,
      json: restartAttempted
        ? body
        : {
            ...body,
            meta: {
              ...(body.meta as Record<string, unknown>),
              paging: {
                limit: 100,
                has_more: true,
                next_cursor: "invalid-continuation",
              },
            },
          },
    });
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openTimelineInspector(page, source.record_id);
  await page.getByTestId(mentionItemTestId(String(mention.item_ref))).click();
  const select = page.getByTestId(mentionResolveTargetSelectTestId());
  await select.selectOption(target.record_id);
  const beforeContinuation = cursors.length;
  await page.getByRole("button", { name: "Load more targets" }).click();
  await expect(
    page.getByText(/This target page could not be used/u),
  ).toBeVisible();
  expect(cursors.slice(beforeContinuation)).toEqual(["invalid-continuation"]);
  restartAttempted = true;
  failRestart = true;
  const restart = page.getByRole("button", {
    name: "Restart target discovery",
  });
  await restart.focus();
  try {
    await page.keyboard.press("Enter");
    await expect.poll(() => cursors.length).toBe(beforeContinuation + 2);
    await expect(restart).toBeVisible();
    await expect(restart).toBeFocused();
    await expect(restart).toHaveAttribute("aria-busy", "true");
    await expect(restart).toHaveAttribute("aria-disabled", "true");
    await page.keyboard.press("Enter");
    expect(cursors.length).toBe(beforeContinuation + 2);
  } finally {
    releaseRestart?.();
  }
  await expect(
    page.getByText(/Previously loaded targets await revalidation/u),
  ).toBeVisible();
  await expect(restart).toBeFocused();
  await expect(select).toHaveValue(target.record_id);
  await expect(
    select.locator(`option[value="${target.record_id}"]`),
  ).toHaveAttribute("disabled", "");
  await expect(
    page.getByRole("button", { name: "Resolve to existing", exact: true }),
  ).toBeDisabled();
  const retry = page.getByRole("button", { name: "Retry target read" });
  await retry.focus();
  await page.keyboard.press("Enter");
  await expect(
    select.locator(`option[value="${target.record_id}"]`),
  ).not.toHaveAttribute("disabled", "");
  await expect(select).toBeFocused();
  await expect(
    page.getByRole("button", { name: "Resolve to existing", exact: true }),
  ).toBeEnabled();
  expect(cursors.slice(beforeContinuation)).toEqual([
    "invalid-continuation",
    null,
    null,
  ]);
});

async function runRetryFocus(page: Page, entityType: "host" | "identity") {
  await page.setViewportSize({
    width: entityType === "identity" ? 768 : 1280,
    height: entityType === "identity" ? 640 : 800,
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey(`MENTION-DISCOVERY-${entityType}`),
    `Mention ${entityType} discovery recovery`,
  );
  const targetView =
    entityType === "host" ? hostsViewSchemaId : identitiesViewSchemaId;
  const target = await createViewRow(page, incidentId, targetView, {
    client_txn_id: uniqueTxn(`discovery-${entityType}`),
    ...(entityType === "host"
      ? {
          "host.display_name": "Discovery target host",
          "host.hostname": "discovery-target.example.test",
        }
      : {
          "identity.display_name": "Discovery target identity",
          "identity.upn": "discovery-target@example.test",
        }),
  });
  const raw = `unresolved-${entityType}`;
  const fieldKey =
    entityType === "host" ? hostRefsFieldKey : identityRefsFieldKey;
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn(`discovery-source-${entityType}`),
    "timeline.activity_synopsis_text": `Discover ${entityType} mention targets`,
    [fieldKey]: collectionActionsPayload([raw]),
  });
  const mention = requireItemByRawText(collectionItems(source, fieldKey), raw);
  const mentionMutations: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() !== "GET" &&
      (/\/entity-mentions\/[^/]+\/resolve$/u.test(request.url()) ||
        /\/views\/[^/]+\/rows$/u.test(request.url()))
    )
      mentionMutations.push(request.url());
  });
  let firstPage: Record<string, unknown> | null = null;
  const requests: (string | null)[] = [];
  let releaseFirstContinuation: (() => void) | undefined;
  const firstContinuation = new Promise<void>((resolve) => {
    releaseFirstContinuation = resolve;
  });
  let releaseHeld: (() => void) | undefined;
  const held = new Promise<void>((resolve) => {
    releaseHeld = resolve;
  });
  await page.route(`**/views/${targetView}/query`, async (route) => {
    const cursor =
      (route.request().postDataJSON() as { cursor_token?: string })
        .cursor_token ?? null;
    requests.push(cursor);
    if (cursor === null) {
      const response = await route.fetch();
      const body = (await response.json()) as Record<string, unknown>;
      firstPage = body;
      await route.fulfill({
        response,
        json: {
          ...body,
          meta: {
            ...(body.meta as Record<string, unknown>),
            paging: {
              limit: 100,
              has_more: true,
              next_cursor: "held-continuation",
            },
          },
        },
      });
    } else if (requests.filter((item) => item === cursor).length === 1) {
      await firstContinuation;
      await route.abort("failed");
    } else {
      await held;
      const body = firstPage;
      if (!body) throw new Error("First target page was not captured");
      await route.fulfill({
        json: {
          ...body,
          data: { ...(body.data as Record<string, unknown>), rows: [] },
          meta: {
            ...(body.meta as Record<string, unknown>),
            paging: { limit: 100, has_more: false, next_cursor: null },
          },
        },
      });
    }
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openTimelineInspector(page, source.record_id);
  await page.getByTestId(mentionItemTestId(String(mention.item_ref))).click();
  const select = page.getByTestId(mentionResolveTargetSelectTestId());
  await expect(
    select.locator(`option[value="${target.record_id}"]`),
  ).toHaveCount(1);
  await select.selectOption(target.record_id);
  const loadMore = page.getByRole("button", { name: "Load more targets" });
  try {
    await loadMore.click();
    await expect
      .poll(() => requests.filter((cursor) => cursor !== null).length)
      .toBe(1);
    await expect(loadMore).toBeVisible();
    await expect(loadMore).toBeFocused();
    await expect(loadMore).toHaveAttribute("aria-busy", "true");
    await expect(loadMore).toHaveAttribute("aria-disabled", "true");
    await page.keyboard.press("Enter");
    expect(requests.filter((cursor) => cursor !== null)).toHaveLength(1);
  } finally {
    releaseFirstContinuation?.();
  }
  const retry = page.getByRole("button", { name: "Retry target read" });
  await expect(retry).toBeVisible();
  await retry.focus();
  try {
    await page.keyboard.press("Enter");
    await expect
      .poll(() => requests.filter((cursor) => cursor !== null).length)
      .toBe(2);
    await expect(retry).toBeVisible();
    await expect(retry).toBeFocused();
  } finally {
    releaseHeld?.();
  }
  await expect(select).toBeFocused();
  await expect(select).toHaveValue(target.record_id);
  await expect(
    page.getByRole("button", { name: "Load more targets" }),
  ).toHaveCount(0);
  expect(mentionMutations).toEqual([]);
  expect(await page.evaluate(() => window.scrollY)).toBe(0);
  const inspector = page.getByTestId(timelineInspectorTestId());
  await expect(inspector).toBeVisible();
  if (entityType === "identity") {
    await page.evaluate(() => {
      document.documentElement.style.zoom = "200%";
    });
    const spacing = await page.addStyleTag({
      content:
        "#root * { letter-spacing: 0.12em !important; line-height: 1.5 !important; word-spacing: 0.16em !important; } #root p { margin-bottom: 2em !important; }",
    });
    try {
      const restart = page.getByRole("button", {
        name: "Restart target discovery",
      });
      await restart.scrollIntoViewIfNeeded();
      const bounds = await restart.boundingBox();
      expect(bounds).not.toBeNull();
      expect((bounds?.x ?? -1) >= 0).toBe(true);
      expect((bounds?.x ?? 0) + (bounds?.width ?? 0)).toBeLessThanOrEqual(768);
      await restart.click({ trial: true });
      await select.scrollIntoViewIfNeeded();
      await expect(select).toBeVisible();
    } finally {
      await spacing.evaluate((element) =>
        element.parentNode?.removeChild(element),
      );
      await page.evaluate(() => {
        document.documentElement.style.zoom = "";
      });
    }
  }
  const resolved = page.waitForResponse(
    (response) =>
      /\/entity-mentions\/[^/]+\/resolve$/u.test(response.url()) &&
      response.request().method() === "POST",
  );
  await page
    .getByRole("button", { name: "Resolve to existing", exact: true })
    .click();
  expect((await resolved).ok()).toBe(true);
  expect(mentionMutations).toHaveLength(1);
}
