import { changeGrouping, gridGroupRowTestId } from "@cartulary/test-utils";
import type { Page, TestInfo } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createViewRow,
  patchTimelineRecord,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

const hostsViewSchemaId = "cartulary.view.hosts.v1";
const timelineViewSchemaId = "cartulary.view.timeline.v1";
const hostRefsFieldKey = "timeline.host_refs";

type ViewRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
};

type CollectionItem = Record<string, unknown>;

test("visual workbook default and grouped states", async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 1440, height: 900 });

  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("V301"),
    "Workbook visual default",
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("v301-alpha"),
    "timeline.summary": "Alpha grouped row",
  });
  const reviewedRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("v301-beta"),
      "timeline.summary": "Beta grouped row",
    },
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await page.getByTestId(`row-${reviewedRow.record_id}-mark-reviewed`).click();
  await expect(
    page.getByTestId(`row-${reviewedRow.record_id}-capture-state`),
  ).toHaveText("reviewed");

  await captureScreenshot(
    page,
    testInfo,
    "workbook-default-viewport",
    page.getByRole("main"),
  );

  await changeGrouping(page, "timeline", "timeline.capture_state");
  await expect(
    page.getByTestId(
      gridGroupRowTestId("timeline", "timeline.capture_state", "reviewed"),
    ),
  ).toBeVisible();
  await captureScreenshot(
    page,
    testInfo,
    "workbook-grouped-grid",
    page.getByTestId("timeline-grid-shell"),
  );
});

test("visual workbook mention and save-state states", async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 1440, height: 900 });

  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("V302"),
    "Workbook visual mentions",
  );
  const resolvedHost = (await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn("v302-host"),
      "host.display_name": "Gateway node",
      "host.hostname": "gateway-node.example.test",
      "host.aliases": collectionActionsPayload(["VPN Gateway"]),
    },
  )) as ViewRow;
  const unresolvedRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("v302-unresolved"),
      "timeline.summary": "Unresolved mention row",
      "timeline.host_refs": collectionActionsPayload(["WS-023?"]),
    },
  )) as ViewRow;
  const resolvedRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("v302-resolved"),
      "timeline.summary": "Resolved mention row",
      "timeline.host_refs": collectionActionsPayload([" vpn   gateway "]),
    },
  )) as ViewRow;
  const resolvedMention = requireItemByRawText(
    collectionItems(resolvedRow, hostRefsFieldKey),
    " vpn   gateway ",
  );
  await patchTimelineRecord(page, resolvedRow.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: resolvedRow.row_version,
    client_txn_id: uniqueTxn("v302-resolve"),
    changes: [
      {
        field_key: hostRefsFieldKey,
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "resolve_item",
              item_ref: resolvedMention.item_ref,
              resolved_record_id: resolvedHost.record_id,
            },
          ],
        },
      },
    ],
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(`row-${unresolvedRow.record_id}-hostRefs-items`),
  ).toContainText("WS-023?");
  await expect(
    page
      .getByTestId(`row-${resolvedRow.record_id}-hostRefs-items`)
      .getByLabel("Resolved Gateway node"),
  ).toBeVisible();
  await captureScreenshot(
    page,
    testInfo,
    "workbook-mention-states",
    page.getByTestId("timeline-grid-shell"),
  );

  const delayedPatchRoute = `${apiBase}/api/v1/records/${resolvedRow.record_id}`;
  await page.route(delayedPatchRoute, async (route) => {
    await page.waitForTimeout(300);
    await route.continue();
  });

  const summaryInput = page.getByTestId(`row-${resolvedRow.record_id}-summary`);
  await summaryInput.fill("Save-state strip visual");
  await summaryInput.press("Enter");
  await expect(page.getByTestId("save-state")).toHaveText("Syncing");
  await captureScreenshot(
    page,
    testInfo,
    "workbook-save-state-strip",
    page.locator('[data-testid="save-state"]').locator(".."),
  );
  await page.unroute(delayedPatchRoute);
});

function collectionActionsPayload(tokens: readonly string[]) {
  return {
    kind: "collection_actions_v1",
    actions: tokens.map((rawText) => ({
      op: "add_token",
      raw_text: rawText,
    })),
  };
}

async function captureScreenshot(
  page: Page,
  testInfo: TestInfo,
  name: string,
  target: {
    screenshot: (options: { path: string }) => Promise<Buffer>;
  },
) {
  const path = testInfo.outputPath(`${name}.png`);
  await target.screenshot({ path });
  await testInfo.attach(name, {
    path,
    contentType: "image/png",
  });
  await expect.soft(page).toBeTruthy();
}

function collectionItems(row: ViewRow, fieldKey: string): CollectionItem[] {
  const value = row.cells[fieldKey]?.value;
  if (
    !value ||
    typeof value !== "object" ||
    Array.isArray(value) ||
    !("items" in value) ||
    !Array.isArray(value.items)
  ) {
    return [];
  }
  return value.items.filter(
    (item): item is CollectionItem => item !== null && typeof item === "object",
  );
}

function requireItemByRawText(
  items: readonly CollectionItem[],
  rawText: string,
): CollectionItem {
  const item = items.find((entry) => entry.raw_text === rawText);
  if (!item) {
    throw new Error(`missing collection item for raw text ${rawText}`);
  }
  return item;
}
