import {
  applyFilterChip,
  removeFilterChip,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  authTestId,
  draftCellTestId,
  draftTimelineCollectionInputTestId,
  gridRowTestId,
  gridRowVersionAttribute,
  gridScrollportSelector,
  gridShellTestId,
  incidentLandingTestId,
  indicatorObservationTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  timelineCollectionInputTestId,
  timelineInspectorSectionTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Locator } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { revokeAllSessions } from "./support/auth/sessions";
import { installVisualPreferences } from "./support/auth/visualPreferences";
import {
  collectionActionsPayload,
  collectionItems,
  findRow,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createTimelineFillers } from "./support/timeline/fixtures";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { showTimelineCollectionColumns } from "./support/workbook/collections";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

const fields = [
  ["timeline.host_refs", "hosts"],
  ["timeline.identity_refs", "identities"],
  ["timeline.tags", "tags"],
] as const;

// This collection-only check follows the painted surface-1 grid/inspector stack.
// Viewer cells add the Grid Adapter's read-only stripe over that surface.
async function expectCollectionCueContrast(cue: Locator) {
  const paint = await cue.evaluate((element) => {
    type Color = { r: number; g: number; b: number; a: number };
    const parse = (value: string): Color => {
      const hex = /^#([0-9a-f]{6})$/i.exec(value);
      if (hex) {
        const bytes = hex[1];
        if (!bytes)
          throw new Error(`Unsupported collection paint color: ${value}`);
        return {
          r: Number.parseInt(bytes.slice(0, 2), 16),
          g: Number.parseInt(bytes.slice(2, 4), 16),
          b: Number.parseInt(bytes.slice(4, 6), 16),
          a: 1,
        };
      }
      const match =
        /^rgba?\((\d+(?:\.\d+)?),\s*(\d+(?:\.\d+)?),\s*(\d+(?:\.\d+)?)(?:,\s*(\d+(?:\.\d+)?))?\)$/.exec(
          value,
        );
      if (!match)
        throw new Error(`Unsupported collection paint color: ${value}`);
      return {
        r: Number(match[1]),
        g: Number(match[2]),
        b: Number(match[3]),
        a: match[4] === undefined ? 1 : Number(match[4]),
      };
    };
    const blend = (top: Color, bottom: Color): Color => ({
      r: top.r * top.a + bottom.r * (1 - top.a),
      g: top.g * top.a + bottom.g * (1 - top.a),
      b: top.b * top.a + bottom.b * (1 - top.a),
      a: 1,
    });
    const layers: { color: string; image: string; readOnly: boolean }[] = [];
    for (let node: Element | null = element; node; node = node.parentElement) {
      const style = getComputedStyle(node);
      if (style.opacity !== "1")
        throw new Error(
          `Unexpected collection paint opacity: ${style.opacity}`,
        );
      layers.push({
        color: style.backgroundColor,
        image: style.backgroundImage,
        readOnly: node.classList.contains(
          "cartulary-grid-cell-state-read-only",
        ),
      });
      if (parse(style.backgroundColor).a === 1) break;
    }
    const base = layers.at(-1);
    if (!base || parse(base.color).a !== 1)
      throw new Error("Collection cue has no opaque paint base");
    const surface1 = parse(
      getComputedStyle(document.documentElement)
        .getPropertyValue("--ct-colors-surface-1")
        .trim(),
    );
    const background = parse(base.color);
    if (
      background.r !== surface1.r ||
      background.g !== surface1.g ||
      background.b !== surface1.b
    )
      throw new Error(`Unexpected collection paint base: ${base.color}`);
    let backgrounds = [background];
    for (const layer of [...layers].reverse()) {
      const color = parse(layer.color);
      backgrounds = backgrounds.map((under) => blend(color, under));
      if (layer.image === "none") continue;
      const stops = [
        ...layer.image.matchAll(
          /color\(srgb ([\d.]+) ([\d.]+) ([\d.]+) \/ ([\d.]+)\)/g,
        ),
      ];
      if (
        !layer.readOnly ||
        !layer.image.startsWith("repeating-linear-gradient") ||
        stops.length !== 2 ||
        stops[0]?.[0] !== stops[1]?.[0]
      )
        throw new Error(`Unexpected collection paint image: ${layer.image}`);
      const stripe: Color = {
        r: Number(stops[0]?.[1]) * 255,
        g: Number(stops[0]?.[2]) * 255,
        b: Number(stops[0]?.[3]) * 255,
        a: Number(stops[0]?.[4]),
      };
      backgrounds = backgrounds.flatMap((under) => [
        under,
        blend(stripe, under),
      ]);
    }
    const foreground = parse(getComputedStyle(element).color);
    if (foreground.a !== 1)
      throw new Error("Unexpected translucent collection cue text");
    const luminance = (color: Color) => {
      const channel = (value: number) => {
        const normalized = value / 255;
        return normalized <= 0.04045
          ? normalized / 12.92
          : ((normalized + 0.055) / 1.055) ** 2.4;
      };
      return (
        0.2126 * channel(color.r) +
        0.7152 * channel(color.g) +
        0.0722 * channel(color.b)
      );
    };
    const ratios = backgrounds.map((under) => {
      const light = Math.max(luminance(foreground), luminance(under));
      const dark = Math.min(luminance(foreground), luminance(under));
      return (light + 0.05) / (dark + 0.05);
    });
    return {
      foreground: getComputedStyle(element).color,
      fontSize: getComputedStyle(element).fontSize,
      base: base.color,
      layers,
      minimumRatio: Math.min(...ratios),
    };
  });
  expect(paint.minimumRatio, JSON.stringify(paint)).toBeGreaterThanOrEqual(4.5);
  return paint;
}

test("Timeline saved tag removal preserves capture and exact collection identity at supported viewports", async ({
  page,
  workerAdmin,
}) => {
  test.setTimeout(240_000);
  const preferences = await installVisualPreferences(page, workerAdmin.user_id);
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 720 },
  ]) {
    preferences.select(viewport.width === 1440 ? "comfortable" : "compact");
    await page.setViewportSize(viewport);
    const incident = await createIncident(
      page,
      uniqueIncidentKey("TSRM"),
      "Saved tag removal",
    );
    const tags = (names: readonly [string, ...string[]]) => {
      const [first, ...rest] = names;
      const actions: [
        { op: "add_tag"; tag_name: string },
        ...{ op: "add_tag"; tag_name: string }[],
      ] = [
        { op: "add_tag", tag_name: first },
        ...rest.map((tag_name) => ({ op: "add_tag" as const, tag_name })),
      ];
      return { kind: "collection_actions_v1" as const, actions };
    };
    const longLabel = `z-${"long".repeat(14)}`;
    const row = await createViewRow(page, incident, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("tsrm-source"),
      "timeline.activity_synopsis_text": "Tag removal source",
      "timeline.tags": tags(["alpha", "beta Ω", "gamma", longLabel]),
    });
    const other = await createViewRow(page, incident, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("tsrm-other"),
      "timeline.activity_synopsis_text": "Other tagged source",
      "timeline.tags": tags(["beta Ω"]),
    });
    const before = findRow(
      await queryViewRows(page, incident, timelineViewSchemaId),
      row.record_id,
    );
    const target = collectionItems(before, "timeline.tags").find(
      (item) => item.display_text === "beta Ω",
    );
    if (!target) throw new Error("Missing returned beta item_ref");
    const patches: unknown[] = [];
    page.on("request", (request) => {
      if (
        request.method() === "PATCH" &&
        request.url().endsWith(`/records/${row.record_id}`)
      )
        patches.push(request.postDataJSON());
    });
    await page.goto(`/?incident_id=${incident}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await showTimelineCollectionColumns(page);
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(
        row.record_id,
        "timeline.tags",
        "grid",
      ),
    });
    await page
      .getByTestId(
        relationshipItemsTestId(row.record_id, "timeline.tags", "grid"),
      )
      .getByRole("button", { name: "Inspect tag: alpha" })
      .click();
    const inspectorInput = page.getByTestId(
      timelineCollectionInputTestId(
        row.record_id,
        "timeline.tags",
        "inspector",
      ),
    );
    await expect(inspectorInput).toBeVisible();
    await inspectorInput.fill("  unsent capture Ω  ");
    const remove = page.getByRole("button", { name: "Remove tag: beta Ω" });
    const gridOffset = await page
      .locator(gridScrollportSelector())
      .evaluate((element) => ({
        left: element.scrollLeft,
        top: element.scrollTop,
      }));
    await remove.focus();
    await page.keyboard.press("Enter");
    await expect(remove).toHaveCount(0);
    await expect(
      page.getByRole("button", { name: "Remove tag: gamma" }),
    ).toBeFocused();
    await expect(inspectorInput).toHaveValue("  unsent capture Ω  ");
    expect(
      await page.locator(gridScrollportSelector()).evaluate((element) => ({
        left: element.scrollLeft,
        top: element.scrollTop,
      })),
    ).toEqual(gridOffset);
    const longAction = page.getByRole("button", {
      name: `Remove tag: ${longLabel}`,
    });
    await longAction.scrollIntoViewIfNeeded();
    expect(
      await longAction.evaluate((element) => {
        const button = element.getBoundingClientRect();
        const chip = element.parentElement?.getBoundingClientRect();
        return (
          chip !== undefined &&
          button.right <= chip.right + 1 &&
          button.left >= chip.left - 1
        );
      }),
    ).toBe(true);
    expect(patches).toHaveLength(1);
    expect(patches[0]).toMatchObject({
      changes: [
        {
          field_key: "timeline.tags",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [{ op: "remove_tag", item_ref: target.item_ref }],
          },
        },
      ],
    });
    const savedRows = await queryViewRows(page, incident, timelineViewSchemaId);
    expect(
      collectionItems(findRow(savedRows, row.record_id), "timeline.tags").map(
        (item) => item.display_text,
      ),
    ).toEqual(["alpha", "gamma", longLabel]);
    expect(
      collectionItems(findRow(savedRows, other.record_id), "timeline.tags").map(
        (item) => item.display_text,
      ),
    ).toEqual(["beta Ω"]);
    await test.info().attach(`saved-tag-removal-${viewport.width}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
  }
});

test("Timeline viewer inspects saved tags without a removal action", async ({
  page,
  sessionTracker,
}) => {
  test.setTimeout(180_000);
  await page.setViewportSize({ width: 1024, height: 720 });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("TSRV"),
    "Viewer saved tags",
  );
  const row = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("tsrv-row"),
    "timeline.activity_synopsis_text": "Viewer tag inspection",
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [
        { op: "add_tag", tag_name: "visible" },
        { op: "add_tag", tag_name: "hidden" },
      ],
    },
  });
  const viewer = await createIncidentMemberUser(page, incident, {
    email: uniqueEmail("tsrv-viewer"),
    display_name: "Tag viewer",
    initial_password: "ViewerTags1!",
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "timeline-collection-input",
    email: viewer.email,
    password: viewer.initial_password,
    purpose: "Saved tag viewer inspection",
    userId: viewer.user_id,
  });
  await page.goto(`/?incident_id=${incident}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page);
  for (const [field, label] of fields) {
    const items = relationshipItemsTestId(row.record_id, field, "grid");
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: items,
    });
    const cell = page
      .getByTestId(items)
      .locator('xpath=ancestor::*[@role="gridcell"][1]');
    await expect(cell).toHaveAttribute("aria-readonly", "true");
    await expect(cell).toHaveAttribute(
      "aria-description",
      `Read-only ${label[0]?.toUpperCase()}${label.slice(1)}`,
    );
    await expect(
      cell.getByRole("button", { name: `Add ${label} token` }),
    ).toHaveCount(0);
  }
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: relationshipItemsTestId(
      row.record_id,
      "timeline.tags",
      "grid",
    ),
  });
  await page
    .getByTestId(
      relationshipOverflowButtonTestId(row.record_id, "timeline.tags"),
    )
    .click();
  const collection = page.getByRole("group", {
    name: "Tags collection editor",
  });
  await expect(
    collection.getByRole("note", { name: "Tag: visible" }),
  ).toBeVisible();
  await expect(
    collection.getByRole("note", { name: "Tag: hidden" }),
  ).toBeVisible();
  await expect(
    collection.getByRole("button", { name: /Remove tag:/u }),
  ).toHaveCount(0);
  await expect(
    page.getByTestId(
      timelineCollectionInputTestId(
        row.record_id,
        "timeline.tags",
        "inspector",
      ),
    ),
  ).toHaveAttribute("readonly");
});

test("Timeline saved tag removal replays uncertainty and refreshes an acknowledged write without resending", async ({
  page,
}) => {
  test.setTimeout(90_000);
  await page.setViewportSize({ width: 1440, height: 900 });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("TSRR"),
    "Saved tag recovery",
  );
  const row = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("tsrr-row"),
    "timeline.activity_synopsis_text": "Tag recovery source",
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [
        { op: "add_tag", tag_name: "retain" },
        { op: "add_tag", tag_name: "mistake" },
      ],
    },
  });
  await page.goto(`/?incident_id=${incident}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page);
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: relationshipItemsTestId(
      row.record_id,
      "timeline.tags",
      "grid",
    ),
  });
  await page
    .getByTestId(
      relationshipItemsTestId(row.record_id, "timeline.tags", "grid"),
    )
    .getByRole("button", { name: "Inspect tag: mistake" })
    .click();
  const inspectorInput = page.getByTestId(
    timelineCollectionInputTestId(row.record_id, "timeline.tags", "inspector"),
  );
  await inspectorInput.fill("unsent recovery text");
  const bodies: string[] = [];
  let failRefresh = false;
  let failedRefreshes = 0;
  const patchPath = `**/api/v1/records/${row.record_id}`;
  const queryPath = `**/views/${timelineViewSchemaId}/query`;
  await page.route(queryPath, async (route) => {
    if (failRefresh && failedRefreshes++ === 0) {
      await route.fulfill({ status: 503, body: "unavailable" });
      return;
    }
    await route.continue();
  });
  await page.route(patchPath, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    bodies.push(route.request().postData() ?? "");
    if (bodies.length === 1) {
      await route.fulfill({ status: 503, body: "unavailable" });
      return;
    }
    const response = await route.fetch();
    await route.fulfill({ response });
    failRefresh = true;
  });
  await page.getByRole("button", { name: "Remove tag: mistake" }).click();
  await expect(
    page.getByRole("button", { name: "Retry original change" }),
  ).toBeVisible();
  await expect(inspectorInput).toHaveValue("unsent recovery text");
  await page.getByRole("button", { name: "Retry original change" }).click();
  await expect(
    page.getByRole("button", { name: "Refresh saved change" }),
  ).toBeVisible();
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).toBe(bodies[0]);
  expect(failedRefreshes).toBeGreaterThan(0);
  await test.info().attach("saved-tag-refresh-recovery", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await page.unroute(queryPath);
  await page.getByRole("button", { name: "Refresh saved change" }).click();
  await expect(
    page.getByRole("button", { name: "Refresh saved change" }),
  ).toHaveCount(0);
  expect(bodies).toHaveLength(2);
  await expect(inspectorInput).toHaveValue("unsent recovery text");
  const saved = findRow(
    await queryViewRows(page, incident, timelineViewSchemaId),
    row.record_id,
  );
  expect(
    collectionItems(saved, "timeline.tags").map((item) => item.display_text),
  ).toEqual(["retain"]);
  await page.unroute(patchPath);
});

test("Timeline unsaved cells keep keyboard focus through draft discard", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("TCKC"),
    "Collection keyboard continuity",
  );
  const row = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("tckc-row"),
    "timeline.activity_utc_text": "2026-04-01T00:00:00Z",
    "timeline.activity_synopsis_text": "Keyboard continuity",
  });
  await page.goto(`/?incident_id=${incident}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page);
  let patches = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${row.record_id}`)
    )
      patches++;
  });
  for (const [field, label] of fields) {
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
    });
    await page
      .getByTestId(relationshipItemsTestId(row.record_id, field, "grid"))
      .locator("xpath=ancestor::fieldset[1]")
      .getByRole("button", { name: `Add ${label} token` })
      .click();
    await page
      .getByTestId(timelineCollectionInputTestId(row.record_id, field, "grid"))
      .fill(`  ${label} Ω 東京  `);
    await page
      .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
      .click();
    await page
      .getByTestId(workbookColumnsMenuTestId(timelineViewSchemaId))
      .getByRole("checkbox", {
        name: label[0]?.toUpperCase() + label.slice(1),
        exact: true,
      })
      .uncheck();
    await page.keyboard.press("Escape");
  }
  const summary = page.getByText("Unsaved cells (3)", { exact: true });
  await summary.press("Enter");
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("textbox", { name: "Retained Hosts" }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("button", { name: "Discard Hosts draft" }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("textbox", { name: "Retained Identities" }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("button", { name: "Discard Identities draft" }),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(
    page.getByText("Unsaved cells (2)", { exact: true }),
  ).toBeVisible();
  await test.info().attach("focus-after-middle-discard", {
    body: JSON.stringify(
      await page.evaluate(() => ({
        tag: document.activeElement?.tagName,
        label: document.activeElement?.getAttribute("aria-label"),
        text: document.activeElement?.textContent?.trim().slice(0, 80),
      })),
    ),
    contentType: "application/json",
  });
  await expect(
    page.getByRole("textbox", { name: "Retained Tags" }),
  ).toBeFocused();
  await expect(
    page.getByRole("textbox", { name: "Retained Hosts" }),
  ).toHaveValue("  hosts Ω 東京  ");
  await page.keyboard.press("Shift+Tab");
  await expect(
    page.getByRole("button", { name: "Discard Hosts draft" }),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(
    page.getByText("Unsaved cells (1)", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("textbox", { name: "Retained Tags" }),
  ).toBeFocused();
  await expect(
    page.getByRole("textbox", { name: "Retained Tags" }),
  ).toHaveValue("  tags Ω 東京  ");
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("button", { name: "Discard Tags draft" }),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(
    page.getByText("Unsaved cells (1)", { exact: true }),
  ).toHaveCount(0);
  await expect(page.getByRole("grid")).toBeFocused();
  await page.keyboard.press("Tab");
  expect(
    await page.evaluate(() => document.activeElement !== document.body),
  ).toBe(true);
  await page.keyboard.press("Shift+Tab");
  expect(
    await page.evaluate(() => {
      const active = document.activeElement;
      return (
        active instanceof HTMLElement &&
        active !== document.body &&
        active.isConnected &&
        active.getClientRects().length > 0
      );
    }),
  ).toBe(true);
  expect(patches).toBe(0);
});

test("Timeline collection identical native replacement survives older settlement", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TCIR"),
    "Collection native revision",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("tcir-row"),
    "timeline.activity_synopsis_text": "Native collection revision",
    "timeline.host_refs": collectionActionsPayload([
      "seed-host?",
      "other-host?",
    ]),
    "timeline.identity_refs": collectionActionsPayload([
      "seed-identity?",
      "other-identity?",
    ]),
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [
        { op: "add_tag", tag_name: "seed" },
        { op: "add_tag", tag_name: "other" },
      ],
    },
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page);
  const close = page.getByTestId(
    workbookInspectorCloseButtonTestId(timelineViewSchemaId),
  );
  const borrowed = page
    .getByRole("group", { name: "Workbook browsing" })
    .getByRole("button", { name: "Refresh", exact: true });
  for (const [field, label] of fields) {
    if (await close.count()) await close.click();
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
    });
    const cell = page
      .getByRole("group", {
        name: `${label[0]?.toUpperCase()}${label.slice(1)} collection cell`,
        exact: true,
      })
      .filter({
        has: page.getByTestId(
          relationshipItemsTestId(row.record_id, field, "grid"),
        ),
      });
    const input = page.getByTestId(
      timelineCollectionInputTestId(row.record_id, field, "grid"),
    );
    const inspector = page.getByTestId(
      timelineCollectionInputTestId(row.record_id, field, "inspector"),
    );
    const gridRow = page.getByTestId(
      gridRowTestId(timelineViewSchemaId, row.record_id),
    );
    const originalVersion = await gridRow.getAttribute(gridRowVersionAttribute);
    let release = () => {};
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    let reachedResponse = () => {};
    const responseHeld = new Promise<void>((resolve) => {
      reachedResponse = resolve;
    });
    let patches = 0;
    const path = `**/api/v1/records/${row.record_id}`;
    await page.route(path, async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      patches++;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      reachedResponse();
      await gate;
      await route.fulfill({ response });
    });
    try {
      await page
        .getByTestId(relationshipOverflowButtonTestId(row.record_id, field))
        .click();
      const inspectorDraft = `Independent ${label} Ω?`;
      await inspector.fill(inspectorDraft);
      await borrowed.focus();
      expect(patches).toBe(0);
      await cell.getByRole("button", { name: `Add ${label} token` }).click();
      const token = `native replacement ${label} Ω?`;
      await input.fill(token);
      await input.press(field === "timeline.identity_refs" ? "Tab" : "Enter");
      await responseHeld;
      // The real blur joins the same departure; the next input is a new edit.
      await input.evaluate((element: HTMLInputElement) => element.blur());
      await input.focus();
      expect(patches).toBe(1);
      await expect(input).toBeFocused();
      await input.evaluate((element: HTMLInputElement) => {
        element.dataset.nativeInputCount = "0";
        element.addEventListener("input", () => {
          element.dataset.nativeInputCount = String(
            Number(element.dataset.nativeInputCount) + 1,
          );
        });
      });
      await input.press("ControlOrMeta+A");
      await page.keyboard.insertText(token);
      await expect(input).toHaveAttribute("data-native-input-count", "1");
      await expect(input).toHaveValue(token);
      await input.evaluate((element: HTMLInputElement) =>
        element.setSelectionRange(2, 7, "backward"),
      );
      release();
      await expect
        .poll(() => gridRow.getAttribute(gridRowVersionAttribute))
        .not.toBe(originalVersion);
      await expect(input).toHaveValue(token);
      await expect(input).toBeFocused();
      await expect
        .poll(() =>
          input.evaluate((element: HTMLInputElement) => [
            element.selectionStart,
            element.selectionEnd,
            element.selectionDirection,
          ]),
        )
        .toEqual([2, 7, "backward"]);
      await expect(inspector).toHaveValue(inspectorDraft);
      expect(patches).toBe(1);
      if (field === "timeline.host_refs") {
        await page
          .context()
          .grantPermissions(["clipboard-read", "clipboard-write"]);
        await page.evaluate(
          (value) => navigator.clipboard.writeText(value),
          `${token}!`,
        );
        await input.press("ControlOrMeta+A");
        await input.press("ControlOrMeta+V");
        await expect(input).toHaveValue(`${token}!`);
        await input.press("ControlOrMeta+Z");
        await expect(input).toHaveValue(token);
        await input.press("ControlOrMeta+Y");
        await expect(input).toHaveValue(`${token}!`);
      }
      await input.press("Escape");
      await expect(input).toHaveCount(0);
    } finally {
      release();
      await page.unroute(path);
    }
  }
});

test("Timeline collection newer native authoring fences an older rejection", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TCIRJ"),
    "Collection rejection revision",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("tcirj-row"),
    "timeline.activity_synopsis_text": "Native collection rejection",
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [
        { op: "add_tag", tag_name: "seed" },
        { op: "add_tag", tag_name: "other" },
      ],
    },
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page);
  const field = "timeline.tags";
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
  });
  const input = page.getByTestId(
    timelineCollectionInputTestId(row.record_id, field, "grid"),
  );
  const overflow = page.getByTestId(
    relationshipOverflowButtonTestId(row.record_id, field),
  );
  let release = () => {};
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  let reachedResponse = () => {};
  const responseHeld = new Promise<void>((resolve) => {
    reachedResponse = resolve;
  });
  let patches = 0;
  let receipts = 0;
  const path = `**/api/v1/records/${row.record_id}`;
  await page.route(path, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    patches++;
    const response = await route.fetch();
    expect(response.status()).toBe(400);
    expect((await response.json()).error.details.reason_code).toBe(
      "no_effective_change",
    );
    reachedResponse();
    await gate;
    await route.fulfill({ response });
    receipts++;
  });
  try {
    await page.getByRole("button", { name: "Add tags token" }).click();
    await input.fill("seed");
    await input.press("Enter");
    await responseHeld;
    await input.press("ControlOrMeta+A");
    await page.keyboard.insertText("seed");
    await input.evaluate((element: HTMLInputElement) =>
      element.setSelectionRange(1, 3, "backward"),
    );
    await overflow.focus();
    expect(patches).toBe(1);
    release();
    await expect.poll(() => receipts).toBe(1);
    await page.evaluate(
      () =>
        new Promise<void>((resolve) =>
          requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
        ),
    );
    await expect(overflow).toBeFocused();
    await expect(input).toHaveValue("seed");
    expect(
      await input.evaluate((element: HTMLInputElement) => [
        element.selectionStart,
        element.selectionEnd,
        element.selectionDirection,
      ]),
    ).toEqual([1, 3, "backward"]);
    expect(patches).toBe(1);
  } finally {
    release();
    await page.unroute(path);
  }
});

test("Timeline collection recordless equal edit survives create promotion", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TCIRP"),
    "Collection recordless revision",
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page);
  const field = "timeline.tags";
  const path =
    "/api/v1/incidents/" +
    incidentId +
    "/views/" +
    timelineViewSchemaId +
    "/rows";
  const held = await holdBrowserRequest(page, { method: "POST", path });
  let patches = 0;
  const count = (request: import("@playwright/test").Request) => {
    if (
      request.method() === "PATCH" &&
      request.url().includes("/api/v1/records/")
    )
      patches++;
  };
  page.on("request", count);
  try {
    const draft = page.getByTestId(draftTimelineCollectionInputTestId(field));
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: draftTimelineCollectionInputTestId(field),
    });
    const token = "recordless native Ω?";
    await draft.fill(token);
    await draft.press("Enter");
    await held.waitForHit;
    await expect(draft).toBeFocused();
    await draft.press("ControlOrMeta+A");
    await page.keyboard.insertText(token);
    await draft.evaluate((element: HTMLInputElement) =>
      element.setSelectionRange(2, 8),
    );
    held.release();
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, incidentId, timelineViewSchemaId)).length,
      )
      .toBe(1);
    const saved = (
      await queryViewRows(page, incidentId, timelineViewSchemaId)
    )[0];
    if (!saved) throw new Error("Missing promoted collection row");
    const promoted = page.getByTestId(
      timelineCollectionInputTestId(saved.record_id, field, "grid"),
    );
    await expect(promoted).toHaveValue(token);
    await expect(promoted).toBeFocused();
    expect(
      await promoted.evaluate((element: HTMLInputElement) => [
        element.selectionStart,
        element.selectionEnd,
      ]),
    ).toEqual([2, 8]);
    expect(collectionItems(saved, field)).toHaveLength(1);
    expect(held.hitCount()).toBe(1);
    expect(patches).toBe(0);
    await expect(draft).toHaveValue("");
  } finally {
    await held.dispose();
    page.off("request", count);
  }
});

test("Timeline collection input production characterization", async ({
  page,
}, testInfo) => {
  test.setTimeout(240_000);
  await page.setViewportSize({ width: 1440, height: 900 });
  // Diagnostic React commit observations only; never used for timing gates.
  await page.addInitScript(() => {
    const counters: Record<string, number> = {
      commits: 0,
      collectionRenders: 0,
      columnPropReplacements: 0,
    };
    type Fiber = {
      type?: unknown;
      flags: number;
      child?: Fiber;
      sibling?: Fiber;
      memoizedProps?: Record<string, unknown> & {
        columns?: unknown;
        binding?: { kind?: string };
      };
      alternate?: Fiber;
    };
    const observed = window as unknown as {
      __tciWork: typeof counters;
      __REACT_DEVTOOLS_GLOBAL_HOOK__: unknown;
    };
    observed.__tciWork = counters;
    let previous = new WeakSet<object>();
    observed.__REACT_DEVTOOLS_GLOBAL_HOOK__ = {
      supportsFiber: true,
      inject: () => 1,
      onCommitFiberUnmount: () => {},
      onCommitFiberRoot: (_renderer: number, root: { current: Fiber }) => {
        counters.commits = (counters.commits ?? 0) + 1;
        const current = new WeakSet<object>();
        const visit = (fiber: Fiber | undefined) => {
          if (!fiber) return;
          current.add(fiber);
          if (
            !previous.has(fiber) &&
            typeof fiber.type === "function" &&
            (fiber.flags & 1) !== 0 &&
            fiber.memoizedProps?.binding?.kind === "collection"
          ) {
            counters.collectionRenders = (counters.collectionRenders ?? 0) + 1;
            for (const [key, value] of Object.entries(
              fiber.memoizedProps ?? {},
            )) {
              if (value !== fiber.alternate?.memoizedProps?.[key])
                counters[`changed:${key}`] =
                  (counters[`changed:${key}`] ?? 0) + 1;
            }
          }
          if (
            !previous.has(fiber) &&
            Array.isArray(fiber.memoizedProps?.columns) &&
            fiber.alternate &&
            fiber.memoizedProps?.columns !==
              fiber.alternate.memoizedProps?.columns
          )
            counters.columnPropReplacements =
              (counters.columnPropReplacements ?? 0) + 1;
          visit(fiber.child);
          visit(fiber.sibling);
        };
        visit(root.current);
        previous = current;
      },
    };
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TCI"),
    "Collection input characterization",
  );
  let row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("tci-row"),
    "timeline.activity_utc_text": "2026-04-01T00:00:00Z",
    "timeline.activity_synopsis_text": "Collection authoring",
    "timeline.raw_activity_text": "Observation source Ω",
    "timeline.host_refs": collectionActionsPayload([
      "first-host?",
      "second-host Ω?",
    ]),
    "timeline.identity_refs": collectionActionsPayload([
      "first-identity?",
      "second-identity 東京?",
    ]),
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [
        { op: "add_tag", tag_name: "triage" },
        { op: "add_tag", tag_name: "second Ω" },
      ],
    },
  });
  const requests: { method: string; path: string }[] = [];
  page.on("request", (request) => {
    const path = new URL(request.url()).pathname;
    if (path.startsWith("/api/"))
      requests.push({ method: request.method(), path });
  });
  const observations: unknown[] = [];
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page);
  const close = page.getByTestId(
    workbookInspectorCloseButtonTestId(timelineViewSchemaId),
  );
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  const cdp = await page.context().newCDPSession(page);
  await cdp.send("Performance.enable");
  const metrics = async () => {
    const result = await cdp.send("Performance.getMetrics");
    return Object.fromEntries(
      result.metrics
        .filter(({ name }) =>
          [
            "ScriptDuration",
            "LayoutCount",
            "RecalcStyleCount",
            "Nodes",
          ].includes(name),
        )
        .map(({ name, value }) => [name, value]),
    );
  };
  const mounted = () =>
    page
      .getByRole("group", { name: /collection (cell|editor)$/ })
      .evaluateAll((groups) => ({
        cells: groups.length,
        inputs: groups.reduce(
          (count, group) => count + group.querySelectorAll("input").length,
          0,
        ),
        invisibleInputs: groups.reduce(
          (count, group) =>
            count +
            [...group.querySelectorAll("input")].filter(
              (input) => getComputedStyle(input).opacity === "0",
            ).length,
          0,
        ),
      }));
  const mutations = () =>
    requests.filter(
      ({ method, path }) =>
        method !== "GET" &&
        /\/api\/v1\/(?:records|entity-mentions|timeline-records)(?:\/|$)/u.test(
          path,
        ),
    ).length;
  const renderWork = () =>
    page.evaluate(() => ({
      ...(window as unknown as { __tciWork: Record<string, number> }).__tciWork,
    }));
  try {
    for (const loaded of [1, 100, 200, 300]) {
      if (loaded === 100) {
        await createTimelineFillers(page, incidentId, "tci-window", 299, {
          occurredAtStart: "2026-04-02T00:00:00Z",
        });
        await browsing
          .getByRole("button", { name: "Refresh", exact: true })
          .click();
      } else if (loaded > 100) {
        await browsing
          .getByRole("button", { name: "Load more", exact: true })
          .click();
      }
      await expect(browsing).toContainText(`${loaded} records loaded`);
      for (const inspectorOpen of [false, true]) {
        if (await close.count()) await close.click();
        if (inspectorOpen) {
          await page
            .getByTestId(
              relationshipOverflowButtonTestId(row.record_id, "timeline.tags"),
            )
            .click();
        }
        if (inspectorOpen && loaded === 300) {
          await page
            .getByTestId(
              workbookInspectorFeatureActionTestId(
                timelineViewSchemaId,
                "indicator.observations.manage",
              ),
            )
            .click();
          await expect(
            page.getByTestId(indicatorObservationTestId("editor")),
          ).toBeVisible();
        }
        for (const [field, label] of fields) {
          await scrollGridTargetIntoView({
            page,
            surface: timelineViewSchemaId,
            targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
          });
          const cell = page
            .getByRole("group", {
              name: `${label[0]?.toUpperCase()}${label.slice(1)} collection cell`,
              exact: true,
            })
            .filter({
              has: page.getByTestId(
                relationshipItemsTestId(row.record_id, field, "grid"),
              ),
            });
          // Profile Add after ordinary semantic selection has settled. Selection
          // itself is measured separately from authoring activation.
          await cell
            .locator("xpath=ancestor::*[@role='gridcell'][1]")
            .click({ position: { x: 1, y: 1 } });
          await page.evaluate(
            () =>
              new Promise<void>((resolve) =>
                requestAnimationFrame(() =>
                  requestAnimationFrame(() => resolve()),
                ),
              ),
          );
          const beforeMounted = await mounted();
          const beforeMetrics = await metrics();
          const beforeRenderWork = await renderWork();
          const beforeRequests = requests.length;
          const beforeMutations = mutations();
          const add = cell.getByRole("button", {
            name: `Add ${label} token`,
            exact: true,
          });
          const activationActions = (await add.count()) ? 1 : 0;
          if (activationActions) await add.click();
          const activatedRenderWork = await renderWork();
          const input = page.getByTestId(
            timelineCollectionInputTestId(row.record_id, field, "grid"),
          );
          await input.fill("raw Ω 東京 middle token?");
          const typedRenderWork = await renderWork();
          await input.evaluate((element: HTMLInputElement) => {
            element.focus();
            element.setSelectionRange(4, 8, "backward");
          });
          const editingMode =
            loaded === 300
              ? "composition"
              : loaded === 100
                ? "caret"
                : "selection";
          if (editingMode === "caret")
            await input.evaluate((element: HTMLInputElement) =>
              element.setSelectionRange(7, 7),
            );
          if (editingMode === "composition")
            await cdp.send("Input.imeSetComposition", {
              text: "仮",
              selectionStart: 1,
              selectionEnd: 1,
            });
          const expectedSelection = await input.evaluate(
            (element: HTMLInputElement) => ({
              text: element.value,
              focused: document.activeElement === element,
              start: element.selectionStart,
              end: element.selectionEnd,
              direction: element.selectionDirection,
            }),
          );
          const original = await input.elementHandle();
          const response = await patchRecord(page, row.record_id, {
            view_schema_id: timelineViewSchemaId,
            base_row_version: row.row_version,
            client_txn_id: uniqueTxn("tci-remote"),
            changes: [
              {
                field_key: "timeline.analyst_text",
                value: `remote-${loaded}-${inspectorOpen}-${field}`,
              },
            ],
          });
          row = response;
          await expect(
            page.getByTestId(
              gridRowTestId(timelineViewSchemaId, row.record_id),
            ),
          ).toHaveAttribute(gridRowVersionAttribute, String(row.row_version));
          const selection = await input.evaluate(
            (element: HTMLInputElement) => ({
              text: element.value,
              focused: document.activeElement === element,
              start: element.selectionStart,
              end: element.selectionEnd,
              direction: element.selectionDirection,
            }),
          );
          const sameInput = await original?.evaluate(
            (element) => element.isConnected,
          );
          const afterMetrics = await metrics();
          observations.push({
            loaded,
            inspectorOpen,
            field,
            editingMode,
            expectedSelection,
            activationActions,
            beforeMounted,
            afterMounted: await mounted(),
            sameInput,
            selection,
            beforeRenderWork,
            activatedRenderWork,
            typedRenderWork,
            browserRequests: requests.length - beforeRequests,
            browserMutations: mutations() - beforeMutations,
            diagnosticWork: Object.fromEntries(
              Object.entries(afterMetrics).map(([name, value]) => [
                name,
                value - (beforeMetrics[name] ?? 0),
              ]),
            ),
          });
          expect(sameInput).toBe(true);
          expect(selection).toEqual(expectedSelection);
          expect((await mounted()).invisibleInputs).toBe(0);
          expect(mutations() - beforeMutations).toBe(0);
          if (editingMode === "composition")
            await cdp.send("Input.imeSetComposition", {
              text: "",
              selectionStart: 0,
              selectionEnd: 0,
            });
          await input.fill("");
          await browsing
            .getByRole("button", { name: "Refresh", exact: true })
            .focus();
        }
      }
    }
  } finally {
    await testInfo.attach("collection-input-observations", {
      body: JSON.stringify(
        { kind: "production-characterization-not-ac043-timing", observations },
        null,
        2,
      ),
      contentType: "application/json",
    });
    await cdp.detach();
  }
});

test("Timeline collection authoring isolates surfaces and settles native departure once", async ({
  page,
}) => {
  test.setTimeout(180_000);
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TCIA"),
    "Collection authoring lifetime",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("tcia-row"),
    "timeline.activity_synopsis_text": "Independent collection authoring",
    "timeline.host_refs": collectionActionsPayload([
      "seed-host?",
      "second-host?",
    ]),
    "timeline.identity_refs": collectionActionsPayload([
      "seed-identity?",
      "second-identity?",
    ]),
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [
        { op: "add_tag", tag_name: "seed" },
        { op: "add_tag", tag_name: "second" },
      ],
    },
  });
  await page.addInitScript(() => {
    const log: unknown[] = [];
    (window as unknown as { tciFocus: unknown[] }).tciFocus = log;
    document.addEventListener(
      "focusin",
      (event) => {
        const target = event.target as HTMLElement;
        log.push({
          tag: target.tagName,
          role: target.getAttribute("role"),
          label: target.getAttribute("aria-label"),
          field: target.dataset.gridFieldKey,
        });
      },
      true,
    );
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page);
  const close = page.getByTestId(
    workbookInspectorCloseButtonTestId(timelineViewSchemaId),
  );
  const borrowed = page
    .getByRole("group", { name: "Workbook browsing" })
    .getByRole("button", { name: "Refresh", exact: true });
  for (const [field, label] of fields) {
    if (await close.count()) await close.click();
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
    });
    const cell = page
      .getByRole("group", {
        name: `${label[0]?.toUpperCase()}${label.slice(1)} collection cell`,
        exact: true,
      })
      .filter({
        has: page.getByTestId(
          relationshipItemsTestId(row.record_id, field, "grid"),
        ),
      });
    const semanticCell = cell.locator('xpath=ancestor::*[@role="gridcell"][1]');
    await expect(semanticCell).toHaveAttribute("aria-readonly", "false");
    await expect(semanticCell).not.toHaveAttribute(
      "aria-description",
      /Read-only/u,
    );
    await expect(
      semanticCell.locator('[data-grid-state-marker="read-only"]'),
    ).toHaveCount(0);
    await expect(semanticCell).not.toHaveClass(
      /cartulary-grid-cell-is-read-only/u,
    );
    const grid = page.getByTestId(
      timelineCollectionInputTestId(row.record_id, field, "grid"),
    );
    const inspector = page.getByTestId(
      timelineCollectionInputTestId(row.record_id, field, "inspector"),
    );
    const path = `**/api/v1/records/${row.record_id}`;
    let release = () => {};
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    let patches = 0;
    let receipts = 0;
    await page.route(path, async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      patches++;
      if (patches === 1) await gate;
      const response = await route.fetch();
      if (field === "timeline.tags" && patches === 3) {
        expect(response.status()).toBe(400);
        expect((await response.json()).error.details.reason_code).toBe(
          "no_effective_change",
        );
      } else expect(response.ok()).toBe(true);
      await route.fulfill({ response });
      receipts++;
    });
    try {
      await expect(grid).toHaveCount(0);
      await cell.getByRole("button", { name: `Add ${label} token` }).click();
      await expect(grid).toBeFocused();
      await grid.fill("grid retained Ω 東京?");
      await page
        .getByTestId(relationshipOverflowButtonTestId(row.record_id, field))
        .click();
      await expect(inspector).toHaveValue("");
      await inspector.fill("Inspector retained Ω?");
      await inspector.press("Escape");
      await expect(inspector).toHaveValue("");
      await expect(inspector.locator("..")).toBeFocused();
      await expect(close).toBeVisible();
      await expect(grid).toHaveValue("grid retained Ω 東京?");
      await inspector.fill("Inspector retained Ω?");
      await borrowed.focus();
      await grid.focus();
      await grid.press("Escape");
      await expect(grid).toHaveCount(0);
      await expect
        .poll(() =>
          page.evaluate(() => ({
            role: document.activeElement?.getAttribute("role"),
            field: document.activeElement
              ?.querySelector("[data-grid-field-key]")
              ?.getAttribute("data-grid-field-key"),
            tag: document.activeElement?.tagName,
            label: document.activeElement?.getAttribute("aria-label"),
          })),
        )
        .toEqual({ role: "gridcell", field, tag: "DIV", label: null });
      await expect(inspector).toHaveValue("Inspector retained Ω?");
      expect(patches).toBe(0);

      await cell.getByRole("button", { name: `Add ${label} token` }).click();
      const token = `captured Ω 東京 ${label}?`;
      await grid.fill(token);
      await grid.press("Enter");
      await expect.poll(() => patches).toBe(1);
      await expect(grid).toBeFocused();
      // Real DOM blur repeats the departure without another keyboard action.
      await grid.evaluate((element: HTMLInputElement) => element.blur());
      await grid.fill("newer unsubmitted Ω?");
      await grid.evaluate((element: HTMLInputElement) =>
        element.setSelectionRange(2, 7, "backward"),
      );
      release();
      await expect.poll(() => receipts).toBe(1);
      await expect(grid).toHaveValue("newer unsubmitted Ω?");
      await expect(grid).toBeFocused();
      await expect
        .poll(() =>
          grid.evaluate((element: HTMLInputElement) => [
            element.selectionStart,
            element.selectionEnd,
            element.selectionDirection,
          ]),
        )
        .toEqual([2, 7, "backward"]);
      await expect(inspector).toHaveValue("Inspector retained Ω?");
      const saved = () =>
        queryViewRows(page, incidentId, timelineViewSchemaId).then((rows) =>
          findRow(rows, row.record_id),
        );
      const tokenCount = async () =>
        collectionItems(await saved(), field).filter(
          (item) =>
            (field === "timeline.tags" ? item.display_text : item.raw_text) ===
            token,
        ).length;
      await expect.poll(tokenCount).toBe(1);
      expect(patches).toBe(1);
      await grid.press("Escape");
      await cell.getByRole("button", { name: `Add ${label} token` }).click();
      await grid.fill("Independent grid draft during Inspector acceptance?");
      await borrowed.focus();
      await inspector.press("Enter");
      await expect.poll(() => receipts).toBe(2);
      await expect(inspector).toHaveValue("");
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
      });
      if (!(await grid.count()))
        await cell.getByRole("button", { name: `Add ${label} token` }).click();
      await expect(grid).toHaveValue(
        "Independent grid draft during Inspector acceptance?",
      );
      await expect
        .poll(async () =>
          collectionItems(await saved(), field).some(
            (item) =>
              (field === "timeline.tags"
                ? item.display_text
                : item.raw_text) === "Inspector retained Ω?",
          ),
        )
        .toBe(true);
      await grid.press("Escape");
      await cell.getByRole("button", { name: `Add ${label} token` }).click();
      await grid.fill(token);
      await grid.press("Tab");
      await expect.poll(() => receipts).toBe(3);
      await expect.poll(tokenCount).toBe(field === "timeline.tags" ? 1 : 2);
      if (field === "timeline.tags") {
        await expect(grid).toHaveValue(token);
        await expect(grid).toBeFocused();
        await grid.press("Escape");
      }
      await expect(grid).toHaveCount(0);
      await expect(inspector).toHaveValue("");
      expect(patches).toBe(3);
    } finally {
      release();
      await test.info().attach("collection-focus", {
        body: JSON.stringify(
          await page.evaluate(
            () => (window as unknown as { tciFocus: unknown[] }).tciFocus,
          ),
        ),
        contentType: "application/json",
      });
      await page.unroute(path);
    }
  }
  const incidentResponse = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}`,
  );
  expect(incidentResponse.ok()).toBeTruthy();
  const incident = (await incidentResponse.json()).data;
  const closed = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/close`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_incident_version: incident.incident_version,
        client_txn_id: uniqueTxn("collection-read-only-close"),
        reason: "Collection authoring presentation verification",
      },
    },
  );
  expect(closed.ok()).toBeTruthy();
  let closedPatches = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${row.record_id}`)
    )
      closedPatches++;
  });
  await page.reload();
  await expect(
    page.getByRole("status", { name: "Incident lifecycle", exact: true }),
  ).toHaveText("Closed, read-only");
  await showTimelineCollectionColumns(page);
  for (const [field, label] of fields) {
    const items = relationshipItemsTestId(row.record_id, field, "grid");
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: items,
    });
    const cell = page
      .getByTestId(items)
      .locator('xpath=ancestor::*[@role="gridcell"][1]');
    await expect(cell).toHaveAttribute("aria-readonly", "true");
    await expect(cell).toHaveAttribute(
      "aria-description",
      `Read-only ${label[0]?.toUpperCase()}${label.slice(1)}`,
    );
    await expect(
      cell.getByRole("button", { name: `Add ${label} token` }),
    ).toHaveCount(0);
  }
  expect(closedPatches).toBe(0);
});

test("Timeline collection authoring follows first-input record promotion without submission", async ({
  page,
}) => {
  test.setTimeout(180_000);
  await page.setViewportSize({ width: 1440, height: 900 });
  for (const [field] of fields) {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("TCIP"),
      "Collection promotion",
    );
    await page.goto("/?incident_id=" + incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await showTimelineCollectionColumns(page);
    const path =
      "/api/v1/incidents/" +
      incidentId +
      "/views/" +
      timelineViewSchemaId +
      "/rows";
    const held = await holdBrowserRequest(page, { method: "POST", path });
    let patches = 0;
    const count = (request: import("@playwright/test").Request) => {
      if (
        request.method() === "PATCH" &&
        request.url().includes("/api/v1/records/")
      )
        patches++;
    };
    page.on("request", count);
    try {
      const summary = page.getByTestId(
        draftCellTestId("timeline.activity_synopsis_text"),
      );
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: draftCellTestId("timeline.activity_synopsis_text"),
      });
      await summary.fill("Capture first");
      await held.waitForHit;
      const draft = page.getByTestId(draftTimelineCollectionInputTestId(field));
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: draftTimelineCollectionInputTestId(field),
      });
      await draft.fill("new retained Ω 東京?");
      await draft.evaluate((element: HTMLInputElement) =>
        element.setSelectionRange(2, 7),
      );
      held.release();
      await expect
        .poll(
          async () =>
            (await queryViewRows(page, incidentId, timelineViewSchemaId))
              .length,
        )
        .toBe(1);
      const saved = (
        await queryViewRows(page, incidentId, timelineViewSchemaId)
      )[0];
      if (!saved) throw new Error("Missing captured row");
      const promoted = page.getByTestId(
        timelineCollectionInputTestId(saved.record_id, field, "grid"),
      );
      await expect(promoted).toHaveValue("new retained Ω 東京?");
      await expect(promoted).toBeFocused();
      expect(
        await promoted.evaluate((element: HTMLInputElement) => [
          element.selectionStart,
          element.selectionEnd,
        ]),
      ).toEqual([2, 7]);
      expect(collectionItems(saved, field)).toHaveLength(0);
      expect(held.hitCount()).toBe(1);
      expect(patches).toBe(0);
      await promoted.press("Escape");
      await expect(promoted).toHaveCount(0);
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: draftTimelineCollectionInputTestId(field),
      });
      await draft.fill("Collection-led capture Ω?");
      await draft.press("Enter");
      await expect
        .poll(
          async () =>
            (await queryViewRows(page, incidentId, timelineViewSchemaId))
              .length,
        )
        .toBe(2);
      await expect(draft).toHaveValue("");
      await expect(draft).toBeFocused();
      const created = (
        await queryViewRows(page, incidentId, timelineViewSchemaId)
      ).find((row) => row.record_id !== saved.record_id);
      if (!created) throw new Error("Missing collection-led capture");
      expect(collectionItems(created, field)).toHaveLength(1);
      expect(held.hitCount()).toBe(2);
      expect(patches).toBe(0);
    } finally {
      await held.dispose();
      page.off("request", count);
    }
  }
});

test("Timeline collection drafts survive detachment and retire with authority", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  test.setTimeout(180_000);
  const contrastObservations: {
    cue: string;
    paint: Awaited<ReturnType<typeof expectCollectionCueContrast>>;
  }[] = [];
  await page.setViewportSize({ width: 1440, height: 900 });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("TCIL"),
    "Collection lifetime",
  );
  const row = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("tcil-row"),
    "timeline.activity_utc_text": "2026-04-01T00:00:00Z",
    "timeline.activity_synopsis_text": "Collection lifetime",
  });
  await createTimelineFillers(page, incident, "tcil-window", 99, {
    occurredAtStart: "2026-04-02T00:00:00Z",
  });
  const members = [];
  for (const name of ["original", "replacement"])
    members.push(
      await createIncidentMemberUser(page, incident, {
        email: uniqueEmail("tcil-" + name),
        display_name: name,
        initial_password: "CollectionLife1!",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      }),
    );
  const original = members[0],
    replacement = members[1];
  if (!original || !replacement)
    throw new Error("Missing collection fixture members");
  const login = (member: typeof original, recovery = false) =>
    sessionTracker.loginTrackedUser(page, {
      recovery,
      createdBy: "timeline-collection-input",
      email: member.email,
      password: member.initial_password,
      purpose: "Collection authority lifetime",
      userId: member.user_id,
    });
  await login(original);
  const sockets = installIncidentSocketMonitor(page, incident);
  await page.goto("/?incident_id=" + incident);
  await sockets.waitForAcceptedSocket();
  await showTimelineCollectionColumns(page);
  const emptyGridCue = (field: string) =>
    page
      .getByTestId(relationshipItemsTestId(row.record_id, field, "grid"))
      .getByText("No items", { exact: true });
  for (const [field] of fields) {
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
    });
    await expect(emptyGridCue(field)).toBeVisible();
    contrastObservations.push({
      cue: `editor empty ${field}`,
      paint: await expectCollectionCueContrast(emptyGridCue(field)),
    });
  }
  await openTimelineInspector(page, row.record_id);
  const emptyInspectorTags = page
    .getByTestId(timelineInspectorSectionTestId("relationships"))
    .getByRole("group", { name: "Tags collection editor" })
    .getByText("No items", { exact: true });
  await expect(emptyInspectorTags).toBeVisible();
  contrastObservations.push({
    cue: "editor empty inspector tags",
    paint: await expectCollectionCueContrast(emptyInspectorTags),
  });
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  const refresh = browsing.getByRole("button", {
    name: "Refresh",
    exact: true,
  });
  const input = (field: string) =>
    page.getByTestId(
      timelineCollectionInputTestId(row.record_id, field, "grid"),
    );
  const activate = async (field: string, label: string) => {
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
    });
    const group = page
      .getByTestId(relationshipItemsTestId(row.record_id, field, "grid"))
      .locator("xpath=ancestor::fieldset[1]");
    if (!(await input(field).count()))
      await group
        .getByRole("button", { name: "Add " + label + " token" })
        .click();
    return input(field);
  };
  const text = (label: string) =>
    "  " + label + " Ω 東京 " + "long raw token? ".repeat(20);
  let patches = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith("/records/" + row.record_id)
    )
      patches++;
  });
  for (const [field, label] of fields) {
    await (await activate(field, label)).fill(text(label));
    await refresh.focus();
  }
  // The viewport releases DOM attachments without submitting retained text.
  const scrollport = page
    .getByTestId(gridShellTestId(timelineViewSchemaId))
    .locator(gridScrollportSelector());
  await scrollport.hover();
  await page.mouse.wheel(0, 100_000);
  await expect(
    page.getByTestId(
      relationshipItemsTestId(row.record_id, "timeline.host_refs", "grid"),
    ),
  ).toHaveCount(0);
  await page.mouse.wheel(0, -100_000);
  for (const [field, label] of fields) {
    await expect(input(field)).toHaveCount(0);
    await expect(await activate(field, label)).toHaveValue(text(label));
    await refresh.focus();
  }
  // Explicit query replacement detaches the row; returning does not reopen it.
  for (const [field, label] of fields) {
    await page
      .getByTestId(workbookColumnsMenuTriggerTestId(timelineViewSchemaId))
      .click();
    const menu = page.getByTestId(
      workbookColumnsMenuTestId(timelineViewSchemaId),
    );
    const visibility = menu.getByRole("checkbox", {
      name: label[0]?.toUpperCase() + label.slice(1),
      exact: true,
    });
    await visibility.uncheck();
    await expect(input(field)).toHaveCount(0);
    await visibility.check();
    await page.keyboard.press("Escape");
    await expect(input(field)).toHaveCount(0);
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
    });
    const retainedCue = page
      .getByTestId(relationshipItemsTestId(row.record_id, field, "grid"))
      .locator("xpath=ancestor::fieldset[1]")
      .getByRole("note", {
        name: `${label[0]?.toUpperCase()}${label.slice(1)} token draft retained`,
      });
    await expect(retainedCue).toHaveText("Draft");
    contrastObservations.push({
      cue: `editor retained ${field}`,
      paint: await expectCollectionCueContrast(retainedCue),
    });
    await expect(await activate(field, label)).toHaveValue(text(label));
    await refresh.focus();
  }
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  for (const [field] of fields) await expect(input(field)).toHaveCount(0);
  await page.getByText("Unsaved cells (3)", { exact: true }).click();
  for (const [, label] of fields)
    await expect(
      page.getByLabel("Retained " + label[0]?.toUpperCase() + label.slice(1), {
        exact: true,
      }),
    ).toHaveValue(text(label));
  await removeFilterChip(page, timelineViewSchemaId, "timeline.capture_state");
  for (const [field, label] of fields) {
    await expect(input(field)).toHaveCount(0);
    await expect(await activate(field, label)).toHaveValue(text(label));
    await refresh.focus();
  }
  expect(patches).toBe(0);
  const queryPath = "**/views/" + timelineViewSchemaId + "/query";
  await page.route(
    queryPath,
    (route) =>
      route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "internal_error",
            message: "Collection refresh fixture failure",
          },
        }),
      }),
    { times: 1 },
  );
  const failedRefresh = page.waitForResponse(
    (response) =>
      response.url().endsWith("/views/" + timelineViewSchemaId + "/query") &&
      response.status() === 503,
  );
  await refresh.click();
  await failedRefresh;
  for (const [field, label] of fields) {
    await expect(await activate(field, label)).toHaveValue(text(label));
    await refresh.focus();
  }
  await refresh.click();
  expect(patches).toBe(0);
  // Let one real write commit, then lose authority before its acknowledgement.
  let release!: () => void;
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  let committed = false;
  const path = "**/api/v1/records/" + row.record_id;
  await page.route(path, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    const response = await route.fetch();
    expect(response.status()).toBe(200);
    committed = true;
    await gate;
    await route.fulfill({ response });
  });
  try {
    const host = await activate("timeline.host_refs", "hosts");
    await host.fill("accepted before suspension?");
    await host.press("Enter");
    await expect.poll(() => committed).toBe(true);
    await host.fill(text("hosts"));
    await revokeAllSessions(
      workerAdminRequest,
      original.user_id,
      "Collection suspension",
    );
    await expect(page.getByTestId(authTestId("shell"))).toBeVisible();
    release();
    for (const [field] of fields) await expect(input(field)).toHaveCount(0);
    await login(original, true);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await showTimelineCollectionColumns(page);
    for (const [field, label] of fields) {
      await expect(input(field)).toHaveCount(0);
      await expect(await activate(field, label)).toHaveValue(text(label));
      await refresh.focus();
    }
    await revokeAllSessions(
      workerAdminRequest,
      original.user_id,
      "Collection account replacement",
    );
    await expect(page.getByTestId(authTestId("shell"))).toBeVisible();
    await login(replacement, true);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await showTimelineCollectionColumns(page);
    for (const [field, label] of fields) {
      await expect(input(field)).toHaveCount(0);
      await expect(await activate(field, label)).toHaveValue("");
      await input(field).press("Escape");
    }
    expect(patches).toBe(1);
    const rows = await queryViewRows(page, incident, timelineViewSchemaId);
    const saved = findRow(rows, row.record_id);
    expect(
      collectionItems(saved, "timeline.host_refs").map((item) => item.raw_text),
    ).toEqual(["accepted before suspension?"]);
    expect(collectionItems(saved, "timeline.identity_refs")).toHaveLength(0);
    expect(collectionItems(saved, "timeline.tags")).toHaveLength(0);
    await page.unroute(path);
    // The new account's authorized draft is retained across read-only state,
    // then concealed entirely when incident membership is removed.
    const tag = await activate("timeline.tags", "tags");
    await tag.fill("role transition raw Ω");
    await refresh.focus();
    const membership =
      "/api/v1/incidents/" + incident + "/memberships/" + replacement.user_id;
    expect(
      (
        await workerAdminRequest.patch(membership, {
          data: { base_membership_version: 1, role: "viewer" },
        })
      ).ok(),
    ).toBe(true);
    await tag.focus();
    await tag.press("Enter");
    await expect(tag).toHaveCount(0);
    await page.getByText("Unsaved cells (1)", { exact: true }).click();
    await expect(page.getByLabel("Retained Tags", { exact: true })).toHaveValue(
      "role transition raw Ω",
    );
    await page.setViewportSize({ width: 1024, height: 720 });
    for (const field of ["timeline.identity_refs", "timeline.tags"]) {
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: relationshipItemsTestId(row.record_id, field, "grid"),
      });
      await expect(emptyGridCue(field)).toBeVisible();
      contrastObservations.push({
        cue: `viewer empty ${field}`,
        paint: await expectCollectionCueContrast(emptyGridCue(field)),
      });
      await expect(
        page
          .getByTestId(relationshipItemsTestId(row.record_id, field, "grid"))
          .locator("xpath=ancestor::fieldset[1]")
          .getByRole("button", { name: /Add .* token/u }),
      ).toHaveCount(0);
    }
    const viewerDraft = page
      .getByTestId(
        relationshipItemsTestId(row.record_id, "timeline.tags", "grid"),
      )
      .locator("xpath=ancestor::fieldset[1]")
      .getByRole("note", { name: "Tags token draft retained" });
    await expect(viewerDraft).toHaveText("Draft");
    contrastObservations.push({
      cue: "viewer retained timeline.tags",
      paint: await expectCollectionCueContrast(viewerDraft),
    });
    await test.info().attach("collection-cue-contrast", {
      body: JSON.stringify(contrastObservations, null, 2),
      contentType: "application/json",
    });
    expect(
      (
        await workerAdminRequest.delete(membership, {
          data: { base_membership_version: 2 },
        })
      ).status(),
    ).toBe(204);
    await expect(
      page.getByTestId(incidentLandingTestId("shell")),
    ).toBeVisible();
    await expect(page.getByLabel("Tags token draft retained")).toHaveCount(0);
    await expect(page.getByLabel("Retained Tags", { exact: true })).toHaveCount(
      0,
    );
    expect(patches).toBe(2);
  } finally {
    release();
    await page.unroute(path);
  }
});
