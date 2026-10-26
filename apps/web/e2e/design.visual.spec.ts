import { applyFilterChip } from "@cartulary/test-utils/grid";
import {
  cartularyDefaultThemeId,
  gridRowTestId,
  gridShellTestId,
  workbookInlineDraftRowTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page, Route } from "@playwright/test";
import { expect } from "./fixtures";
import { gridSavedRows } from "./pages/workbookInspector";
import { createIncident } from "./support/incidents/fixtures";
import { closeIncidentForVisual } from "./support/incidents/visualState";
import {
  navigateVisualApplication,
  reloadVisualApplication,
} from "./support/runtime/applicationReadiness";
import { responseBarrier } from "./support/runtime/cleanup";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  awaitResponseBarrier,
  holdBrowserRequest as holdBrowserApiRequest,
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import { assertVisualRegression } from "./support/visual/capture";
import { injectDesignFixture, test } from "./support/visual/fixtures";
import {
  createRichTimelineIncident,
  createVisualTimelineRow,
} from "./support/visual/timelineVisualFixture";
import { assertWorkbookGridVisualRegression } from "./support/visual/workbookCaptures";

const exposedThemeCssVars = [
  "--ct-colors-accent",
  "--ct-colors-canvas",
  "--ct-colors-surface-1",
  "--ct-colors-surface-2",
  "--ct-colors-ink",
  "--ct-colors-ink-muted",
  "--ct-colors-semantic-success",
  "--ct-colors-semantic-caution",
  "--ct-colors-semantic-conflict",
  "--ct-colors-semantic-destructive",
  "--ct-density-default-rowHeight",
  "--ct-density-default-cellPadding",
  "--ct-component-button-primary-backgroundColor",
  "--ct-component-button-primary-textColor",
  "--ct-component-button-secondary-backgroundColor",
  "--ct-component-button-secondary-border",
  "--ct-component-button-danger-textColor",
  "--ct-component-text-input-backgroundColor",
  "--ct-component-chip-backgroundColor",
  "--ct-component-grid-cell-padding",
  "--ct-component-focus-ring-border",
] as const;

async function failWorkbookVisualQuery(route: Route) {
  await route.abort("failed");
}

async function injectExposedThemeVisualFixture(page: Page) {
  await injectDesignFixture(page, {
    ariaLabel: "Exposed theme token state fixture",
    fixtureName: "components",
    missingMainMessage: "Expected workbook shell main before theme fixture",
    styleText: `
      [data-design-fixture='components'] {
        position: fixed;
        inset: var(--ct-spacing-xl);
        box-sizing: border-box;
        display: grid;
        grid-template-columns: 1.1fr 0.9fr;
        gap: var(--ct-spacing-md);
        overflow: hidden;
        background: var(--ct-colors-canvas);
        color: var(--ct-colors-ink);
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-lg);
        padding: var(--ct-spacing-lg);
        box-shadow: var(--ct-elevation-panel);
        font-family: var(--ct-typography-ui-fontFamily);
        font-size: var(--ct-typography-ui-fontSize);
        font-weight: var(--ct-typography-ui-fontWeight);
        letter-spacing: var(--ct-typography-ui-letterSpacing);
        line-height: var(--ct-typography-ui-lineHeight);
        z-index: 1000;
      }

      [data-design-fixture='components'] * {
        box-sizing: border-box;
      }

      .theme-fixture-panel {
        display: grid;
        gap: var(--ct-spacing-sm);
        align-content: start;
        min-width: 0;
        background: var(--ct-colors-surface-1);
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        padding: var(--ct-spacing-md);
      }

      .theme-fixture-title {
        margin: 0;
        color: var(--ct-colors-ink);
        font-family: var(--ct-typography-surface-title-fontFamily);
        font-size: var(--ct-typography-surface-title-fontSize);
        font-weight: var(--ct-typography-surface-title-fontWeight);
        letter-spacing: var(--ct-typography-surface-title-letterSpacing);
        line-height: var(--ct-typography-surface-title-lineHeight);
      }

      .theme-fixture-note {
        margin: 0;
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
        font-weight: var(--ct-typography-metadata-fontWeight);
        letter-spacing: var(--ct-typography-metadata-letterSpacing);
        line-height: var(--ct-typography-metadata-lineHeight);
      }

      .theme-fixture-swatches,
      .theme-fixture-components,
      .theme-fixture-states {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: var(--ct-spacing-xs);
      }

      .theme-swatch,
      .theme-state {
        min-height: var(--ct-density-default-rowHeight);
        display: flex;
        align-items: center;
        gap: var(--ct-spacing-xs);
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-sm);
        padding: var(--ct-density-default-cellPadding);
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink);
      }

      .theme-swatch::before,
      .theme-state::before {
        content: "";
        inline-size: var(--ct-component-icon-inline-size);
        block-size: var(--ct-component-icon-inline-size);
        border-radius: var(--ct-rounded-pill);
        border: var(--ct-border-hairline);
        flex: 0 0 auto;
      }

      .theme-swatch[data-token='accent']::before {
        background: var(--ct-colors-accent);
      }

      .theme-swatch[data-token='surface']::before {
        background: var(--ct-colors-surface-3);
      }

      .theme-swatch[data-token='ink']::before {
        background: var(--ct-colors-ink);
      }

      .theme-swatch[data-token='hairline']::before {
        background: var(--ct-colors-hairline-strong);
      }

      .theme-state[data-state='success']::before {
        background: var(--ct-colors-semantic-success);
      }

      .theme-state[data-state='caution']::before {
        background: var(--ct-colors-semantic-caution);
      }

      .theme-state[data-state='conflict']::before {
        background: var(--ct-colors-semantic-conflict);
      }

      .theme-state[data-state='destructive']::before {
        background: var(--ct-colors-semantic-destructive);
      }

      .theme-button {
        min-height: var(--ct-density-default-rowHeight);
        border: 0;
        border-radius: var(--ct-component-button-primary-rounded);
        padding: var(--ct-component-button-primary-padding);
        font-family: var(--ct-typography-button-fontFamily);
        font-size: var(--ct-typography-button-fontSize);
        font-weight: var(--ct-typography-button-fontWeight);
        letter-spacing: var(--ct-typography-button-letterSpacing);
        line-height: var(--ct-typography-button-lineHeight);
      }

      .theme-button-primary {
        background: var(--ct-component-button-primary-backgroundColor);
        color: var(--ct-component-button-primary-textColor);
      }

      .theme-button-secondary {
        background: var(--ct-component-button-secondary-backgroundColor);
        color: var(--ct-component-button-secondary-textColor);
        border: var(--ct-component-button-secondary-border);
      }

      .theme-button-danger {
        background: var(--ct-component-button-danger-backgroundColor);
        color: var(--ct-component-button-danger-textColor);
        border: var(--ct-border-hairline);
      }

      .theme-input,
      .theme-grid-cell {
        min-height: var(--ct-density-default-rowHeight);
        width: 100%;
        border: var(--ct-component-text-input-border);
        border-radius: var(--ct-component-text-input-rounded);
        background: var(--ct-component-text-input-backgroundColor);
        color: var(--ct-component-text-input-textColor);
        padding: var(--ct-component-text-input-padding);
        font: inherit;
      }

      .theme-chip {
        display: inline-flex;
        align-items: center;
        width: max-content;
        border: var(--ct-component-chip-border);
        border-radius: var(--ct-component-chip-rounded);
        background: var(--ct-component-chip-backgroundColor);
        color: var(--ct-component-chip-textColor);
        padding: var(--ct-component-chip-padding);
      }

      .theme-grid-cell {
        display: flex;
        align-items: center;
        background: var(--ct-component-grid-cell-backgroundColor);
        color: var(--ct-component-grid-cell-textColor);
        padding: var(--ct-component-grid-cell-padding);
        font-family: var(--ct-typography-grid-cell-fontFamily);
        font-size: var(--ct-typography-grid-cell-fontSize);
        font-weight: var(--ct-typography-grid-cell-fontWeight);
        letter-spacing: var(--ct-typography-grid-cell-letterSpacing);
        line-height: var(--ct-typography-grid-cell-lineHeight);
      }

      .theme-focus-sample {
        outline: var(--ct-component-focus-ring-border);
        outline-offset: var(--ct-component-focus-ring-offset);
      }
    `,
    html: `
      <div class="theme-fixture-panel">
        <h2 class="theme-fixture-title">dark_graphite token states</h2>
        <p class="theme-fixture-note">Generated CSS variables applied through the workbook runtime.</p>
        <div class="theme-fixture-swatches" aria-label="Color token samples">
          <div class="theme-swatch" data-token="accent">Accent</div>
          <div class="theme-swatch" data-token="surface">Surface</div>
          <div class="theme-swatch" data-token="ink">Ink</div>
          <div class="theme-swatch" data-token="hairline">Hairline</div>
        </div>
        <div class="theme-fixture-states" aria-label="Semantic state samples">
          <div class="theme-state" data-state="success">Success state</div>
          <div class="theme-state" data-state="caution">Caution state</div>
          <div class="theme-state" data-state="conflict">Conflict state</div>
          <div class="theme-state" data-state="destructive">Destructive state</div>
        </div>
      </div>
      <div class="theme-fixture-panel">
        <h2 class="theme-fixture-title">Component and density states</h2>
        <p class="theme-fixture-note">Default density, buttons, input, chip, focus, and grid-cell tokens.</p>
        <div class="theme-fixture-components">
          <button class="theme-button theme-button-primary" type="button">Primary</button>
          <button class="theme-button theme-button-secondary theme-focus-sample" type="button">Secondary focus</button>
          <button class="theme-button theme-button-danger" type="button">Danger</button>
          <span class="theme-chip">Evidence chip</span>
        </div>
        <input class="theme-input" value="Readonly token input" readonly />
        <div class="theme-grid-cell">Grid cell typography and default density</div>
      </div>
    `,
  });
}

async function assertExposedThemeCssVariables(page: Page) {
  const missingVars = await page.evaluate(
    (varNames) => {
      const styles = window.getComputedStyle(document.documentElement);
      return varNames.filter((name) => !styles.getPropertyValue(name).trim());
    },
    [...exposedThemeCssVars],
  );
  expect(missingVars).toEqual([]);
}

test("Capture exposed dark_graphite token and theme states with deterministic density, color, component, focus, and semantic-state samples.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALDESIGNREADINESS"),
    "browser.design-readiness exposed theme visual fixture",
  );
  await createVisualTimelineRow(page, incidentId, "visualdesignreadiness-row", {
    client_txn_id: uniqueTxn("VISUALDESIGNREADINESS-ROW"),
    "timeline.activity_utc_text": "2026-05-31T11:00:00Z",
    "timeline.activity_synopsis_text":
      "browser.design-readiness exposed theme fixture row",
  });

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(page.locator("main.cartulary-shell").first()).toHaveAttribute(
    "data-cartulary-theme",
    cartularyDefaultThemeId,
  );
  await assertExposedThemeCssVariables(page);
  await injectExposedThemeVisualFixture(page);

  const fixture = page.locator("[data-design-fixture='components']");
  await expect(fixture).toBeVisible();
  await assertVisualRegression(page, "design-exposed-theme-states", fixture);

  const loadingIncidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALDESIGNLOADING"),
    "browser.design-readiness production grid loading states",
  );
  const heldInitialQuery = await holdBrowserApiRequest(page, {
    method: "POST",
    path: `/api/v1/incidents/${loadingIncidentId}/views/${timelineViewSchemaId}/query`,
  });
  try {
    await navigateVisualApplication(page, `/?incident_id=${loadingIncidentId}`);
    await heldInitialQuery.waitForHit;
    const loadingGrid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
    await expect(loadingGrid).toBeVisible();
    await loadingGrid.evaluate((element) => {
      element.setAttribute("data-design-fixture", "delayed-loading");
    });
    const initialLoading = loadingGrid.locator(
      '[data-grid-data-state="initial_loading"]',
    );
    await expect(initialLoading).toContainText("Loading Timeline…");
    await expect(initialLoading.getByRole("button")).toHaveCount(0);
    await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(0);
    await expect(
      page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
    ).toHaveCount(0);
    await assertWorkbookGridVisualRegression(
      page,
      "design-immediate-initial-loading",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    await expect(initialLoading).toContainText("Still loading this surface");
    await expect(initialLoading.getByRole("button")).toHaveCount(0);
    await assertWorkbookGridVisualRegression(
      page,
      "design-delayed-initial-loading",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    heldInitialQuery.release();

    await heldInitialQuery.dispose();
  } finally {
    heldInitialQuery.cancel();
  }

  const staleIncidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALDESIGNSTALE"),
    "browser.design-readiness production stale grid state",
  );
  const staleRow = await createVisualTimelineRow(
    page,
    staleIncidentId,
    "visualdesignstale-row",
    {
      client_txn_id: uniqueTxn("VISUALDESIGNSTALE-ROW"),
      "timeline.activity_utc_text": "2026-05-31T12:00:00Z",
      "timeline.activity_synopsis_text":
        "Previously authorized row retained during stale refresh",
    },
  );
  await navigateVisualApplication(page, `/?incident_id=${staleIncidentId}`);
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, staleRow.record_id)),
  ).toBeVisible();
  const staleQueryPattern = `**/api/v1/incidents/${staleIncidentId}/views/${timelineViewSchemaId}/query`;
  let resolveStaleQueryHit: () => void = () => undefined;
  let releaseStaleQuery: () => void = () => undefined;
  const staleQueryHitBarrier = responseBarrier(page);
  resolveStaleQueryHit = staleQueryHitBarrier.release;
  const staleQueryHit = staleQueryHitBarrier.promise;
  const staleQueryReleaseBarrier = responseBarrier(page);
  releaseStaleQuery = staleQueryReleaseBarrier.release;
  const staleQueryRelease = staleQueryReleaseBarrier.promise;
  const staleQueryFailure = async (route: Route) => {
    if (route.request().method().toUpperCase() !== "POST") {
      await route.fallback();
      return;
    }
    resolveStaleQueryHit();
    if (!(await awaitResponseBarrier(route, staleQueryRelease))) return;
    await failWorkbookVisualQuery(route);
  };
  await installOwnedRoute(page, staleQueryPattern, staleQueryFailure);
  try {
    await applyFilterChip(
      page,
      timelineViewSchemaId,
      "timeline.capture_state",
      "reviewed",
    );
    await staleQueryHit;
    const refreshingState = page.locator('[data-grid-data-state="refreshing"]');
    await expect(refreshingState).toBeVisible();
    await expect(refreshingState).toContainText("Refreshing Timeline…");
    await expect(refreshingState.getByRole("button")).toHaveCount(0);
    await expect(
      page.getByTestId(gridRowTestId(timelineViewSchemaId, staleRow.record_id)),
    ).toBeVisible();
    await expect(
      page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await page
      .getByTestId(gridShellTestId(timelineViewSchemaId))
      .evaluate((element) => {
        element.setAttribute("data-design-fixture", "error-presentation");
      });
    await assertWorkbookGridVisualRegression(
      page,
      "design-grid-background-refresh",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    releaseStaleQuery();
    const staleState = page.locator('[data-grid-data-state="stale_error"]');
    await expect(staleState).toBeVisible();
    await expect(staleState).toContainText("Workbook view load failed.");
    await expect(staleState).toContainText(
      "Previously loaded rows may be stale.",
    );
    await expect(
      staleState.getByRole("button", { name: "Retry" }),
    ).toBeVisible();
    await expect(
      page.getByTestId(gridRowTestId(timelineViewSchemaId, staleRow.record_id)),
    ).toBeVisible();
    await expect(
      page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await page
      .getByTestId(gridShellTestId(timelineViewSchemaId))
      .evaluate((element) => {
        element.setAttribute("data-design-fixture", "error-presentation");
      });
    await assertWorkbookGridVisualRegression(
      page,
      "design-grid-stale-refresh",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    releaseStaleQuery();

    await removeOwnedRoute(page, staleQueryPattern, staleQueryFailure);
  } finally {
    staleQueryReleaseBarrier.cancel();
  }
  await closeIncidentForVisual(page, staleIncidentId);
  await test.step("reload visual application", () =>
    reloadVisualApplication(page));
  await expect(
    page.getByRole("status", { name: "Incident lifecycle", exact: true }),
  ).toHaveText("Closed, read-only");
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, staleRow.record_id)),
  ).toBeVisible();
  await expect(
    page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
  ).toHaveCount(0);
  await assertWorkbookGridVisualRegression(
    page,
    "design-grid-closed-read-only-rows",
    timelineViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );

  const unavailableIncidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALDESIGNUNAVAILABLE"),
    "browser.design-readiness production unavailable grid state",
  );
  const unavailableQueryPattern = `**/api/v1/incidents/${unavailableIncidentId}/views/${timelineViewSchemaId}/query`;
  const unavailableQueryFailure = async (route: Route) => {
    if (route.request().method().toUpperCase() !== "POST") {
      await route.fallback();
      return;
    }
    await failWorkbookVisualQuery(route);
  };
  await installOwnedRoute(
    page,
    unavailableQueryPattern,
    unavailableQueryFailure,
  );
  {
    await navigateVisualApplication(
      page,
      `/?incident_id=${unavailableIncidentId}`,
    );
    const unavailableGrid = page.getByTestId(
      gridShellTestId(timelineViewSchemaId),
    );
    const unavailableState = unavailableGrid.locator(
      '[data-grid-data-state="unavailable"]',
    );
    await expect(unavailableState).toBeVisible();
    await expect(unavailableState).toContainText("Workbook view load failed.");
    await expect(
      unavailableState.getByRole("button", { name: "Retry" }),
    ).toBeVisible();
    await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(0);
    await expect(
      page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
    ).toHaveCount(0);
    await unavailableGrid.evaluate((element) => {
      element.setAttribute("data-design-fixture", "error-presentation");
    });
    await assertWorkbookGridVisualRegression(
      page,
      "design-grid-unavailable-initial-load",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    await removeOwnedRoute(
      page,
      unavailableQueryPattern,
      unavailableQueryFailure,
    );
  }
});
