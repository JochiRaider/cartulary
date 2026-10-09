import { createHash } from "node:crypto";
import type {
  GetCurrentAccountPreferencesResponse,
  PutCurrentAccountPreferencesRequest,
} from "@cartulary/protocol-ts/http";
import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  gridScrollportSelector,
  incidentControlsScrollportTestId,
  rowCellTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { openIncidentFromLanding } from "./pages/incidentDirectory";
import { installVisualPreferences } from "./support/auth/visualPreferences";
import { createIncident } from "./support/incidents/fixtures";
import {
  expectApplicationReady,
  installApplicationAssetMonitor,
  reloadVisualApplication,
} from "./support/runtime/applicationReadiness";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { requestTLS12, withTLSProbe } from "./support/runtime/fixtureTLS";
import { publishIndependentFrontendBuild } from "./support/runtime/frontendBuild";
import { waitForLoadedVendoredFonts } from "./support/runtime/visualRenderer";
import { seedVisualTimelineInvestigation } from "./support/timeline/timelineInvestigation";
import {
  settleVisualGeometry,
  verifyVisualGeometry,
} from "./support/visual/capture";
import {
  normalizeMetadataDocument,
  restoreMetadataDocument,
} from "./support/visual/metadataNormalization";
import {
  focusTimelineShellOrigin,
  observeTimelineShellOrigin,
  verifyTimelineShellOrigin,
} from "./support/visual/timelineShellAnchor";
import {
  registerTimelineVisualFixture,
  verifyTimelineCaptureData,
  verifyTimelineVisualCores,
} from "./support/visual/timelineVisualFixture";
import {
  fetchRecordHistory,
  openHistoryEventDetails,
} from "./support/workbook/history";
import {
  activateCommittedGridCell,
  openTimelineInspector,
} from "./support/workbook/rowMutations";

test("browser fixture trust accepts its authority and rejects foreign authorities and hostnames", async ({
  page,
}) => {
  await withTLSProbe("trusted", async (origin) => {
    const response = await page.goto(origin);
    expect(response?.status()).toBe(200);
    expect((await response?.securityDetails())?.protocol).toBe("TLS 1.3");
  });
  for (const [kind, reason] of [
    ["untrusted", "ERR_CERT_AUTHORITY_INVALID"],
    ["wrong-name", "ERR_CERT_COMMON_NAME_INVALID"],
  ] as const) {
    await withTLSProbe(kind, async (origin) => {
      await expect(page.goto(origin)).rejects.toThrow(reason);
    });
  }
  for (const endpoint of ["frontend", "backend"] as const) {
    await expect(requestTLS12(endpoint)).rejects.toThrow(
      /alert protocol version/i,
    );
  }
});

test("rich Timeline recipe preserves owner states and source text during capture preparation", async ({
  workerAdminPage: page,
  workerAdmin,
}, testInfo) => {
  let exampleRequests = 0;
  await page.route(/^https?:\/\/[^/]*\.example\.test(?:[/:]|$)/, (route) => {
    exampleRequests++;
    return route.abort();
  });
  await installVisualPreferences(page, workerAdmin.user_id);
  const investigation = registerTimelineVisualFixture(
    await seedVisualTimelineInvestigation(page, {
      continuationCount: 36,
    }),
  );
  await openIncidentFromLanding(page, investigation.incidentId);
  await page.evaluate(waitForLoadedVendoredFonts);
  const receipt = await verifyTimelineCaptureData(
    page,
    "incident-directory-default-timeline-workbook-shell",
  );
  expect(receipt).toMatchObject({
    kind: "rich",
    core_rows: 12,
    seeded_rows: 48,
  });
  await expect(
    verifyTimelineCaptureData(page, "undeclared-timeline-capture"),
  ).rejects.toThrow("Undeclared Timeline capture");
  const first = investigation.recordId("authentication-anomaly");
  const before = await investigation.verify();
  const sourceBefore = await investigation.readRows();
  await activateCommittedGridCell(
    page
      .getByTestId(rowCellTestId(first, "timeline.date_entered_text"))
      .locator("xpath=ancestor::*[@role='gridcell'][1]"),
  );
  await focusTimelineShellOrigin(page, first);
  await verifyTimelineShellOrigin(page, first);
  const scrollport = page.locator(gridScrollportSelector());
  await scrollport.evaluate((node) => {
    node.scrollLeft = 24;
  });
  expect((await observeTimelineShellOrigin(page, first)).ready).toBe(false);
  await focusTimelineShellOrigin(page, first);
  await verifyTimelineShellOrigin(page, first);
  expect((await observeTimelineShellOrigin(page, "undeclared-row")).ready).toBe(
    false,
  );
  for (const state of ["loading", "refreshing", "stale_error", "unavailable"]) {
    await page.evaluate((state) => {
      const marker = document.createElement("div");
      marker.id = "readiness-regression";
      marker.setAttribute("data-grid-data-state", state);
      document.body.append(marker);
    }, state);
    expect((await observeTimelineShellOrigin(page, first)).ready).toBe(false);
    await page
      .locator("#readiness-regression")
      .evaluate((node) => node.remove());
  }
  const cell = page
    .getByTestId(rowCellTestId(first, "timeline.date_entered_text"))
    .locator("xpath=ancestor::*[@role='gridcell'][1]");
  await cell.evaluate((node) => (node as HTMLElement).blur());
  expect((await observeTimelineShellOrigin(page, first)).ready).toBe(false);
  await focusTimelineShellOrigin(page, first);
  await cell.evaluate((node) => {
    node.animate(
      [{ transform: "translateX(0px)" }, { transform: "translateX(8px)" }],
      { duration: 500, iterations: Infinity },
    );
  });
  try {
    expect((await observeTimelineShellOrigin(page, first)).ready).toBe(false);
  } finally {
    await cell.evaluate((node) => {
      for (const animation of node.getAnimations()) animation.cancel();
    });
  }
  const fonts = page.locator('link[href="/assets/fonts/fonts.css"]');
  await fonts.evaluate((node) => {
    (node as HTMLLinkElement).disabled = true;
  });
  try {
    expect((await observeTimelineShellOrigin(page, first)).ready).toBe(false);
  } finally {
    await fonts.evaluate((node) => {
      (node as HTMLLinkElement).disabled = false;
    });
    await expect
      .poll(() =>
        page.evaluate(() =>
          Array.from(document.fonts).some((face) => face.family === "Inter"),
        ),
      )
      .toBe(true);
    await page.evaluate(waitForLoadedVendoredFonts);
  }
  await focusTimelineShellOrigin(page, first);
  await verifyTimelineShellOrigin(page, first);
  const authored = "2026-04-18T14:12:34Z 123e4567-e89b-42d3-a456-426614174000";
  await page.evaluate((authored) => {
    const fixture = document.createElement("section");
    fixture.id = "metadata-regression";
    const source = document.createElement("span");
    source.setAttribute("data-source-value", "true");
    source.textContent = authored;
    const metadata = document.createElement("time");
    metadata.setAttribute("data-generated-metadata", "regression-time");
    metadata.textContent = "2026-09-30T20:00:00Z";
    fixture.append(source, metadata);
    document.body.append(fixture);
  }, authored);
  const rule = {
    id: "regression-time",
    target: '[data-generated-metadata="regression-time"]',
    expected_count: 1,
    replacement: "2025-01-01T00:00:00Z",
  };
  try {
    await page.evaluate(normalizeMetadataDocument, [rule]);
    await expect(
      page.locator("#metadata-regression [data-source-value]"),
    ).toHaveText(authored);
    await expect(page.locator(rule.target)).toHaveText(rule.replacement);
    for (const key of ["authentication-anomaly", "unexpected-token-use"]) {
      const recordId = investigation.recordId(key);
      await scrollGridCellIntoView({
        page,
        surface: timelineViewSchemaId,
        recordId,
        cellKey: "timeline.activity_utc_text",
      });
      const row = sourceBefore.find((entry) => entry.record_id === recordId);
      await expect(
        page.getByTestId(rowCellTestId(recordId, "timeline.activity_utc_text")),
      ).toHaveText(String(row?.cells["timeline.activity_utc_text"]?.value));
    }
    await scrollGridCellIntoView({
      page,
      surface: timelineViewSchemaId,
      recordId: first,
      cellKey: "timeline.raw_activity_text",
    });
    await expect(
      page.getByTestId(rowCellTestId(first, "timeline.raw_activity_text")),
    ).toContainText("123e4567-e89b-42d3-a456-426614174000");
    await expect(
      page.getByTestId(rowCellTestId(first, "timeline.raw_activity_text")),
    ).toContainText("2026-04-18T14:12:34Z");
    const scriptRecord = investigation.recordId("script-execution");
    await scrollGridCellIntoView({
      page,
      surface: timelineViewSchemaId,
      recordId: scriptRecord,
      cellKey: "timeline.raw_activity_text",
    });
    const raw = page.getByTestId(
      rowCellTestId(scriptRecord, "timeline.raw_activity_text"),
    );
    await expect(raw).toContainText("<script>example only</script>");
    await expect(raw.locator("script")).toHaveCount(0);
  } finally {
    await page.evaluate(restoreMetadataDocument);
  }
  await expect(page.locator(rule.target)).toHaveText("2026-09-30T20:00:00Z");
  for (const rules of [
    [{ ...rule, expected_count: 2 }],
    [{ ...rule, target: '[data-generated-metadata="absent"]' }],
    [rule, { ...rule, id: "overlap" }],
  ]) {
    await expect(
      page.evaluate(normalizeMetadataDocument, rules),
    ).rejects.toThrow();
    await expect(page.locator(rule.target)).toHaveText("2026-09-30T20:00:00Z");
  }
  await page
    .locator(rule.target)
    .evaluate((node) => node.setAttribute("data-source-value", "true"));
  await expect(
    page.evaluate(normalizeMetadataDocument, [rule]),
  ).rejects.toThrow("source overlap");
  await page.locator("#metadata-regression").evaluate((node) => node.remove());
  const candidateRule = {
    id: "candidate-reference-id",
    target: '[data-generated-metadata="candidate-reference-id"]',
    expected_count: 1,
    replacement: "00000000-0000-0000-0000-000000000001",
  };
  for (const mode of [
    "single",
    "multi",
    "committed",
    "empty-committed",
    "outside-chooser",
    "source-value",
    "input",
  ]) {
    await page.evaluate(
      ({ mode, authored }) => {
        const fixture = document.createElement("section");
        fixture.id = "candidate-metadata-regression";
        fixture.innerHTML = `<div data-cartulary-grid-draft-row="true">
          <span data-grid-field-key="handoff.incoming_owner_user_id">
            <fieldset data-workbook-single-candidates>
              <label><input type="radio" checked>
                <span data-source-value></span>
                <span data-generated-metadata="candidate-reference-id">generated-identity</span>
              </label>
            </fieldset>
          </span>
        </div>`;
        const chooser = fixture.querySelector("fieldset");
        const identity = fixture.querySelector("[data-generated-metadata]");
        const input = fixture.querySelector("input");
        const source = fixture.querySelector("[data-source-value]");
        const cell = fixture.querySelector("[data-grid-field-key]");
        const row = fixture.firstElementChild;
        if (!chooser || !identity || !input || !source || !cell || !row)
          throw new Error("Missing candidate normalization fixture");
        input.value = authored;
        source.textContent = authored;
        if (mode === "multi") {
          chooser.removeAttribute("data-workbook-single-candidates");
          chooser.setAttribute("data-workbook-multi-candidates", "");
          input.type = "checkbox";
        }
        if (mode === "committed" || mode === "empty-committed")
          row.removeAttribute("data-cartulary-grid-draft-row");
        if (mode === "empty-committed")
          cell.setAttribute("data-grid-field-key", "");
        if (mode === "outside-chooser")
          chooser.removeAttribute("data-workbook-single-candidates");
        if (mode === "source-value")
          identity.setAttribute("data-source-value", "");
        if (mode === "input") {
          identity.removeAttribute("data-generated-metadata");
          input.setAttribute(
            "data-generated-metadata",
            "candidate-reference-id",
          );
        }
        document.body.append(fixture);
      },
      { mode, authored },
    );
    const fixture = page.locator("#candidate-metadata-regression");
    try {
      if (mode === "single" || mode === "multi") {
        await page.evaluate(normalizeMetadataDocument, [candidateRule]);
        await expect(fixture.locator(candidateRule.target)).toHaveText(
          candidateRule.replacement,
        );
        await expect(fixture.locator("[data-source-value]")).toHaveText(
          authored,
        );
        await expect(fixture.locator("input")).toHaveValue(authored);
        await expect(fixture.locator("input")).toBeChecked();
        await page.evaluate(restoreMetadataDocument);
        await expect(fixture.locator(candidateRule.target)).toHaveText(
          "generated-identity",
        );
      } else {
        await expect(
          page.evaluate(normalizeMetadataDocument, [candidateRule]),
        ).rejects.toThrow("source overlap");
        await expect(fixture.locator("input")).toHaveValue(authored);
      }
    } finally {
      await page.evaluate(restoreMetadataDocument);
      await fixture.evaluate((node) => node.remove());
    }
  }
  await openTimelineInspector(page, first);
  const analyst = page.locator(
    '[data-inspector-saved-field="timeline.analyst_text"] [data-inspector-field-value] > div[id]',
  );
  await analyst.evaluate((node) =>
    node.setAttribute("data-generated-metadata", "regression-time"),
  );
  try {
    await expect(
      page.evaluate(normalizeMetadataDocument, [rule]),
    ).rejects.toThrow("source overlap");
  } finally {
    await analyst.evaluate((node) =>
      node.removeAttribute("data-generated-metadata"),
    );
  }
  const systemRules = ["timeline-recorded-at", "timeline-edited-at"].map(
    (surface) => ({
      id: surface,
      target: `[data-generated-metadata="${surface}"]`,
      replacement: "2025-01-01T00:00:00.000000Z",
      expected_count: 1,
    }),
  );
  const systemBefore = await Promise.all(
    systemRules.map(({ target }) => page.locator(target).textContent()),
  );
  try {
    await page.evaluate(normalizeMetadataDocument, systemRules);
    for (const { target, replacement } of systemRules)
      await expect(page.locator(target)).toHaveText(replacement);
    expect(await investigation.readRows()).toEqual(sourceBefore);
  } finally {
    await page.evaluate(restoreMetadataDocument);
  }
  expect(
    await Promise.all(
      systemRules.map(({ target }) => page.locator(target).textContent()),
    ),
  ).toEqual(systemBefore);
  await page.getByRole("button", { name: "Open history", exact: true }).click();
  const history = await fetchRecordHistory(page, first);
  for (const item of history.items)
    await openHistoryEventDetails(page, item.history_item_ref);
  const timestamps = page.locator('[data-generated-metadata="history-time"]');
  expect(history.items).toHaveLength(3);
  await expect(timestamps).toHaveCount(3);
  const originalTimes = await timestamps.allTextContents();
  try {
    await page.evaluate(normalizeMetadataDocument, [
      {
        id: "actual-history-time",
        target: '[data-generated-metadata="history-time"]',
        expected_count: 3,
        replacement: "2025-01-01 00:00:00 UTC +00:00",
      },
    ]);
    await expect(timestamps).toHaveText([
      "2025-01-01 00:00:00 UTC +00:00",
      "2025-01-01 00:00:00 UTC +00:00",
      "2025-01-01 00:00:00 UTC +00:00",
    ]);
    await expect(
      page
        .locator("[data-history-value]")
        .filter({ hasText: "123e4567-e89b-42d3-a456-426614174000" }),
    ).toContainText("2026-04-18T14:12:34Z");
  } finally {
    await page.evaluate(restoreMetadataDocument);
  }
  expect(await timestamps.allTextContents()).toEqual(originalTimes);
  expect(await investigation.verify()).toEqual(before);
  expect(await investigation.readRows()).toEqual(sourceBefore);
  expect(exampleRequests).toBe(0);
  await testInfo.attach("investigation-receipt", {
    body: JSON.stringify(before),
    contentType: "application/json",
  });
  await verifyTimelineVisualCores();
});

test("visual harness retains loaded assets across an independent frontend publication", async ({
  workerAdminPage: page,
  workerAdmin,
}, testInfo) => {
  await installVisualPreferences(page, workerAdmin.user_id);
  const failures = installApplicationAssetMonitor(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ASSETLIFETIME"),
    "Frontend artifact lifetime",
  );
  await openIncidentFromLanding(page, incidentId);
  await expectApplicationReady(page);
  const assetPaths = await page.evaluate(() =>
    [
      ...new Set(
        performance
          .getEntriesByType("resource")
          .map((entry) => new URL(entry.name).pathname)
          .filter(
            (asset) => asset.startsWith("/assets/") && asset.endsWith(".js"),
          ),
      ),
    ].sort(),
  );
  expect(assetPaths.length).toBeGreaterThan(1);
  const digest = async (asset: string) => {
    const response = await page.request.get(asset);
    expect(response.status()).toBe(200);
    return createHash("sha256")
      .update(await response.body())
      .digest("hex");
  };
  const before = await Promise.all(assetPaths.map(digest));
  let completed = false;
  const build = Promise.all([
    publishIndependentFrontendBuild(false),
    publishIndependentFrontendBuild(true),
  ]).finally(() => {
    completed = true;
  });
  let reloads = 0;
  try {
    do {
      await reloadVisualApplication(page);
      expect(await Promise.all(assetPaths.map(digest))).toEqual(before);
      reloads++;
    } while (!completed || reloads < 5);
    const runRoot = await build;
    expect(failures).toEqual([]);
    await testInfo.attach("frontend-overlap-proof", {
      body: JSON.stringify({
        run_root: runRoot,
        reloads,
        asset_digests: before,
      }),
      contentType: "application/json",
    });
  } finally {
    await build;
  }
});

test("visual harness preferences remain local across density order and page closure", async ({
  workerAdminPage: page,
  workerAdminRequest,
  workerAdmin,
}, testInfo) => {
  const readPersisted = async () => {
    const response = await workerAdminRequest.get(
      "/api/v1/account/preferences",
    );
    expect(response.status()).toBe(200);
    return ((await response.json()) as GetCurrentAccountPreferencesResponse)
      .data;
  };
  // Explicit contamination injection belongs only to this harness regression.
  // Visual scenario variants below never send persistence requests.
  const original = await readPersisted();
  const writeSeed = async (
    density: PutCurrentAccountPreferencesRequest["density_mode"],
  ) => {
    const latest = await readPersisted();
    const data: PutCurrentAccountPreferencesRequest = {
      base_preferences_version: latest.preferences_version,
      client_txn_id: uniqueTxn("visual-isolation-seed"),
      density_mode: density,
    };
    const response = await workerAdminRequest.put(
      "/api/v1/account/preferences",
      { data },
    );
    expect(response.status()).toBe(200);
  };
  let primary: unknown;
  let cleanupFailure: unknown;
  try {
    await writeSeed("comfortable");
    const before = await readPersisted();
    expect(before.density_mode).toBe("comfortable");
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALISOLATION"),
      "Visual preference isolation",
    );
    let current = page;
    for (const order of [
      ["compact", "comfortable"],
      ["comfortable", "compact"],
    ] as const) {
      const preferences = await installVisualPreferences(
        current,
        workerAdmin.user_id,
      );
      expect(preferences.read().density_mode).toBeNull();
      installApplicationAssetMonitor(current);
      await openIncidentFromLanding(current, incidentId);
      await expect(
        current.getByTestId(workbookShellReadyTestId()),
      ).toHaveAttribute("data-cartulary-density", "compact");
      for (const density of order) {
        preferences.select(density);
        await reloadVisualApplication(current);
        await expect(
          current.getByTestId(workbookShellReadyTestId()),
        ).toHaveAttribute("data-cartulary-density", density);
      }
      if (order[0] === "compact") {
        await expect(
          Promise.reject(new Error("injected assertion failure")),
        ).rejects.toThrow("injected assertion failure");
      } else {
        await expect(
          expect.poll(() => false, { timeout: 20 }).toBe(true),
        ).rejects.toThrow();
      }
      await current.close();
      current = await page.context().newPage();
    }
    const fresh = await installVisualPreferences(current, workerAdmin.user_id);
    expect(fresh.read().density_mode).toBeNull();
    expect(await readPersisted()).toEqual(before);
    await current.close();
    await testInfo.attach("visual-preference-isolation", {
      body: JSON.stringify({
        persisted_value_and_version_unchanged: true,
        pages: 3,
        variant_orders: 2,
      }),
      contentType: "application/json",
    });
  } catch (error) {
    primary = error;
    throw error;
  } finally {
    // This API context is owned independently of every page closed above.
    try {
      await writeSeed(original.density_mode);
    } catch (error) {
      cleanupFailure = error;
      await testInfo.attach("secondary-seed-cleanup-failure", {
        body: JSON.stringify({
          stage: "contamination_seed_cleanup",
          status: "fail",
        }),
        contentType: "application/json",
      });
    }
  }
  if (cleanupFailure && !primary) throw cleanupFailure;
});

test("visual harness anchors converge after viewport focus scroll and delayed layout changes", async ({
  page,
}) => {
  await page.setContent(
    `<div style="height:120px"></div><div data-testid="${incidentControlsScrollportTestId()}" style="height:900px;overflow:auto;overflow-anchor:none"><div style="height:600px"></div><button id="anchor">Review change</button><div style="height:1200px"></div></div>`,
  );
  const anchor = page.locator("#anchor");
  for (const width of [1280, 390, 768]) {
    for (const zoom of ["100%", "200%"]) {
      await page.setViewportSize({ width, height: 720 });
      await page.evaluate((zoom) => {
        document.documentElement.style.zoom = zoom;
      }, zoom);
      const expected = await settleVisualGeometry(page, {
        locator: anchor,
        align: "center",
        focus: true,
      });
      await page.evaluate(() => {
        window.scrollTo(0, 100);
        document.querySelector("button")?.blur();
        document
          .querySelector("button")
          ?.previousElementSibling?.setAttribute("style", "height:600px");
      });
      expect(
        await settleVisualGeometry(page, {
          locator: anchor,
          align: "center",
          focus: true,
        }),
      ).toEqual(expected);
    }
  }
  await anchor.evaluate((element) => {
    requestAnimationFrame(() =>
      element.previousElementSibling?.setAttribute("style", "height:650px"),
    );
  });
  await settleVisualGeometry(page, { locator: anchor, align: "start" });
  await anchor.evaluate((element) =>
    element.previousElementSibling?.setAttribute("style", "height:750px"),
  );
  await expect(
    verifyVisualGeometry(page, { locator: anchor, align: "start" }),
  ).rejects.toThrow(/visual geometry/);
  await settleVisualGeometry(page, {
    locator: anchor,
    align: "center",
    focus: true,
  });
  await anchor.evaluate((element) => (element as HTMLElement).blur());
  await expect(
    verifyVisualGeometry(page, {
      locator: anchor,
      align: "center",
      focus: true,
    }),
  ).rejects.toThrow(/visual geometry/);
  await page.setContent(
    `<div style="height:120px"></div><div data-testid="${incidentControlsScrollportTestId()}" style="height:900px;overflow:auto"><div style="height:900px"></div><button id="clamped">Confirm change</button></div>`,
  );
  const clamped = page.locator("#clamped");
  await expect(
    settleVisualGeometry(page, { locator: clamped, align: "center" }),
  ).rejects.toThrow(/visual geometry/);
  await settleVisualGeometry(page, {
    locator: clamped,
    align: "center",
    outerScroll: "drawer_end",
  });
  await clamped.evaluate((element) => document.body.append(element));
  await expect(
    verifyVisualGeometry(page, { locator: clamped, align: "center" }),
  ).rejects.toThrow(/visual geometry/);
});

test("visual harness identifies a missing asset before a secondary page closure", async ({
  page,
}) => {
  installApplicationAssetMonitor(page);
  await page.route("**/assets/*.js", (route) =>
    route.fulfill({ status: 404, body: "" }),
  );
  await page.goto("/");
  await expect(expectApplicationReady(page)).rejects.toMatchObject({
    name: "CartularyFrontendArtifactError",
  });
  await page.close();
});
