import { createHash } from "node:crypto";
import type {
  GetCurrentAccountPreferencesResponse,
  PutCurrentAccountPreferencesRequest,
} from "@cartulary/protocol-ts/http";
import {
  incidentControlsScrollportTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
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
import { publishIndependentFrontendBuild } from "./support/runtime/frontendBuild";
import {
  settleVisualGeometry,
  verifyVisualGeometry,
} from "./support/visual/capture";

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
