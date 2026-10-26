import { Buffer } from "node:buffer";
import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import {
  expectNoFocusTrap,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectAndRecordContrast } from "./support/accessibility/records";
import {
  importNetworkFlowCSV,
  networkFlowMinimalCSV,
  openClaimedNetworkAnalysis,
} from "./support/extensions/network_flow_activity/workspace";
import { semanticGridCell } from "./support/workbook/keyboardGrid";
import { expectNetworkAnalysisChromeGeometry } from "./support/workbook/presentationGeometry";
import { expectIncidentHeadingAndHeaders } from "./support/workbook/shellAccessibility";

test("a11y.network-analysis.row-01 Verify claimed Network Analysis incident heading, content-backed headers, tabs, queries, grids, inspector, graph, mapping, focus return, and ARIA evidence.", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NETWORKFLOWA11Y");
  await expectIncidentHeadingAndHeaders(page);
  await page
    .getByTestId(networkAnalysisTestId("import-input"))
    .setInputFiles(networkFlowMinimalCSV);
  const mappingDialog = page.getByTestId(
    networkAnalysisTestId("mapping-dialog"),
  );
  await expect(mappingDialog).toBeVisible({ timeout: 30_000 });
  const mappingProfile = page.getByTestId(
    networkAnalysisTestId("mapping-profile"),
  );
  await expect(
    page
      .getByRole("region", { name: "Recovery navigation", exact: true })
      .locator(":scope > h2"),
  ).toBeFocused();
  // Inspect the retained import surface itself; it lives above workspace unmount.
  for (const viewport of [
    { width: 390, height: 480 },
    { width: 768, height: 360 },
    { width: 1440, height: 450 },
  ]) {
    await page.setViewportSize(viewport);
    const geometry = await mappingDialog.evaluate((node) => ({
      width: node.clientWidth,
      scroll: node.scrollWidth,
      left: node.getBoundingClientRect().left,
      right: node.getBoundingClientRect().right,
    }));
    expect(geometry.scroll).toBeLessThanOrEqual(geometry.width + 1);
    expect(geometry.left).toBeGreaterThanOrEqual(0);
    expect(geometry.right).toBeLessThanOrEqual(viewport.width + 1);
    const previewAction = mappingDialog.getByTestId(
      networkAnalysisTestId("mapping-preview"),
    );
    await previewAction.focus();
    await previewAction.scrollIntoViewIfNeeded();
    await expect(previewAction).toBeInViewport();
    await previewAction.press("Tab");
    await expect(
      mappingDialog.getByRole("button", { name: "Close", exact: true }),
    ).not.toBeFocused();
    await expect
      .poll(() =>
        mappingDialog.evaluate(
          (node) => !node.contains(document.activeElement),
        ),
      )
      .toBe(true);
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    expect(
      await mappingDialog.evaluate(
        (node) => node.scrollWidth <= node.clientWidth + 1,
      ),
    ).toBe(true);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        ".network-flow-dialog * { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; } .network-flow-dialog p { margin-block-end: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 480 });
      expect(
        await mappingDialog.evaluate(
          (node) => node.scrollWidth <= node.clientWidth + 1,
        ),
      ).toBe(true);
      await test.info().attach("network-flow-mapping-text-spacing", {
        body: await page.screenshot(),
        contentType: "image/png",
      });
    },
  );
  await page.setViewportSize({ width: 1440, height: 900 });
  await mappingProfile.focus();
  await expectAllInteractiveControlsNamed(page);
  await expectAndRecordContrast(page, [
    networkAnalysisTestId("mapping-dialog"),
    networkAnalysisTestId("mapping-profile"),
    networkAnalysisTestId("mapping-preview"),
  ]);
  await page
    .getByTestId(networkAnalysisTestId("mapping-display-name"))
    .fill("accessible-flow");
  await page.getByTestId(networkAnalysisTestId("mapping-preview")).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("mapping-preview-summary")),
  ).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("mapping-apply")).click();
  await expect(mappingDialog).toHaveCount(0);

  const workspace = page.getByTestId(networkAnalysisTestId("workspace"));
  await expect(workspace).toHaveAttribute(
    "data-extension-profile-id",
    "network_flow_activity",
  );
  await expect(
    page.getByRole("tab", { name: /accessible-flow/ }),
  ).toHaveAttribute("aria-selected", "true");
  await expect(
    page.getByRole("region", { name: "Network Flow filters" }),
  ).toBeVisible();
  await expect(
    page.getByRole("grid", { name: "Accepted Network Flow rows" }),
  ).toBeVisible();
  const semanticCell = page
    .getByRole("grid", { name: "Accepted Network Flow rows" })
    .locator(
      '[role="gridcell"] [data-grid-field-key]:not([data-grid-field-key^="__"])',
    )
    .first();
  await expectIncidentHeadingAndHeaders(page, "Source row");
  const focusedGridCell = semanticGridCell(semanticCell);
  await semanticCell.click();
  await expect(
    page.getByRole("complementary", { name: "Network Flow cell inspector" }),
  ).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("inspector-close")).click();
  await expect(focusedGridCell).toBeFocused();

  const graphMode = page.getByTestId(networkAnalysisTestId("mode-graph"));
  await graphMode.focus();
  await expectVisibleFocus(graphMode);
  await graphMode.press("Enter");
  const edgeSelect = page
    .getByTestId(/^network-flow-edge-/)
    .first()
    .getByRole("button", { name: /^Select edge/u });
  await expect(edgeSelect).toBeVisible();
  await edgeSelect.click();
  const drawer = page.getByTestId(networkAnalysisTestId("contributor-drawer"));
  await expect(drawer).toBeVisible();
  await expect(drawer.getByRole("button", { name: /close/i })).toBeVisible();
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  await expectAndRecordContrast(page, [
    networkAnalysisTestId("workspace"),
    networkAnalysisTestId("mode-graph"),
    networkAnalysisTestId("contributor-close"),
  ]);
  await page.getByTestId(networkAnalysisTestId("contributor-close")).click();
  await expect(edgeSelect).toBeFocused();

  await page.getByRole("button", { name: "Saved graphs" }).click();
  const savedPanel = page.getByTestId(networkAnalysisTestId("saved-graphs"));
  const saveTrigger = page.getByTestId(
    networkAnalysisTestId("saved-graph-create"),
  );
  await saveTrigger.focus();
  await expectVisibleFocus(saveTrigger);
  await saveTrigger.press("Enter");
  const savedName = page.getByTestId(networkAnalysisTestId("saved-graph-name"));
  await expect(savedName).toBeFocused();
  await savedName.press("Shift+Tab");
  await expect(
    page.getByRole("button", { name: "Close", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(savedName).toBeFocused();
  await savedName.fill("é".repeat(33));
  await expect(savedName).toHaveAttribute("aria-invalid", "true");
  await expect(savedName).toHaveAccessibleDescription(/exceeds 64 bytes/);
  await expect(
    page.getByRole("button", { name: "Save graph", exact: true }),
  ).toBeDisabled();
  await expect(
    page.getByText("This name exceeds 64 bytes.", { exact: false }),
  ).toBeVisible();
  await savedName.fill("é".repeat(32));
  const savedDialog = page.getByTestId(
    networkAnalysisTestId("saved-graph-dialog"),
  );
  for (const viewport of [
    { width: 390, height: 480 },
    { width: 768, height: 360 },
  ]) {
    await page.setViewportSize(viewport);
    expect(
      await savedDialog.evaluate(
        (node) => node.scrollWidth <= node.clientWidth + 1,
      ),
    ).toBe(true);
    const submit = savedDialog.getByRole("button", {
      name: "Save graph",
      exact: true,
    });
    await submit.focus();
    await submit.scrollIntoViewIfNeeded();
    await expect(submit).toBeInViewport();
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    expect(
      await savedDialog.evaluate(
        (node) => node.scrollWidth <= node.clientWidth + 1,
      ),
    ).toBe(true);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        ".network-flow-dialog * { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; }",
    },
    async () => {
      expect(
        await savedDialog.evaluate(
          (node) => node.scrollWidth <= node.clientWidth + 1,
        ),
      ).toBe(true);
      await test.info().attach("saved-graph-name-spacing", {
        body: await page.screenshot(),
        contentType: "image/png",
      });
    },
  );
  await page.keyboard.press("Escape");
  await expect(saveTrigger).toBeFocused();
  await saveTrigger.press("Enter");
  await expect(savedName).toHaveValue("é".repeat(32));
  await savedName.fill("Accessible saved graph");
  const captureQuery = savedDialog.getByRole("button", {
    name: "Use current query",
  });
  await captureQuery.focus();
  await expectVisibleFocus(captureQuery);
  await captureQuery.press("Enter");
  await expect(savedName).toHaveValue("Accessible saved graph");
  await page.getByRole("button", { name: "Save graph" }).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("saved-graph-dialog")),
  ).toHaveCount(0);
  await expect(saveTrigger).toBeFocused();
  await expect(
    savedPanel.getByText("Materialization succeeded.", { exact: true }),
  ).toBeVisible({ timeout: 15_000 });

  const savedOperation = savedPanel.getByRole("region", {
    name: "Saved graph operation",
  });
  const reloadOperation = savedOperation.getByRole("button", {
    name: "Reload operation graph",
  });
  await page.setViewportSize({ width: 768, height: 540 });
  await reloadOperation.focus();
  await expectVisibleFocus(reloadOperation);
  await reloadOperation.press("Enter");
  await expect(reloadOperation).toBeFocused();
  expect(
    await savedOperation.evaluate(
      (node) => node.scrollWidth <= node.clientWidth + 1,
    ),
  ).toBe(true);
  await page.setViewportSize({ width: 1440, height: 900 });

  const savedVertex = page
    .getByTestId(/^network-flow-saved-graph-vertex-/u)
    .first()
    .getByRole("button");
  await savedVertex.click();
  const savedContributors = page.getByTestId(
    networkAnalysisTestId("saved-graph-contributors"),
  );
  await expect(savedContributors).toBeVisible();
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  await expectAndRecordContrast(page, [
    networkAnalysisTestId("saved-graphs"),
    networkAnalysisTestId("saved-graph-result"),
    networkAnalysisTestId("saved-graph-contributors"),
  ]);
  const retainedVertex = await savedVertex.elementHandle();
  const renameSaved = savedPanel.getByRole("button", {
    name: "Rename",
    exact: true,
  });
  await renameSaved.focus();
  await renameSaved.press("Enter");
  await savedName.fill("Renamed accessible graph");
  await page
    .getByRole("button", { name: "Rename graph", exact: true })
    .press("Enter");
  await expect(renameSaved).toBeFocused();
  await expect(savedContributors).toContainText("Renamed accessible graph");
  expect(
    await savedVertex.evaluate(
      (node, retained) => node === retained,
      retainedVertex,
    ),
  ).toBe(true);
  await savedContributors
    .getByRole("button", { name: "Close", exact: true })
    .focus();
  await page.keyboard.press("Escape");
  await expect(savedVertex).toBeFocused();
  await retainedVertex?.dispose();
  const refreshSaved = savedPanel.getByRole("button", {
    name: "Refresh",
    exact: true,
  });
  await refreshSaved.focus();
  await refreshSaved.press("Enter");
  await expect(
    page.getByRole("dialog", { name: "Refresh saved graph" }),
  ).toContainText("Renamed accessible graph");
  await page.keyboard.press("Escape");
  await expect(refreshSaved).toBeFocused();
  const retireSaved = savedPanel.getByRole("button", {
    name: "Retire",
    exact: true,
  });
  await retireSaved.focus();
  await retireSaved.press("Enter");
  await expect(
    page.getByRole("dialog", { name: "Retire saved graph" }),
  ).toContainText("Leased immutable results remain protected");
  await page.keyboard.press("Escape");
  await expect(retireSaved).toBeFocused();

  await page
    .getByTestId(networkAnalysisTestId("graph-surface-explore"))
    .click();
  await page.getByTestId(networkAnalysisTestId("mode-rows")).click();
  await page
    .getByLabel("Endpoint IP value")
    .fill(`2001:db8:${"longsegment".repeat(12)}`);
  await page
    .getByTestId(networkAnalysisTestId("advanced-filters"))
    .locator("summary")
    .click();
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 720 },
    { width: 768, height: 640 },
  ]) {
    await page.setViewportSize(viewport);
    await expectNetworkAnalysisChromeGeometry(page);
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectNetworkAnalysisChromeGeometry(page, {
      enforceDocumentBlockExtent: false,
    });
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await withPresentation(
    page,
    {
      textSpacing: `
        .network-flow-chrome * {
          letter-spacing: 0.12em !important;
          line-height: 1.5 !important;
          word-spacing: 0.16em !important;
        }
      `,
    },
    async () => {
      await expectNetworkAnalysisChromeGeometry(page);
    },
  );
  await page.setViewportSize({ width: 1440, height: 900 });
});

test("a11y.network-analysis.indicator-link keyboard target selection local errors and responsive recovery", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFLINKA11Y");
  const value = "2001:db8:1234:5678:9abc:def0:1234:5678";
  await importNetworkFlowCSV(page, {
    displayName: "IPv6 link source",
    file: {
      name: "ipv6-link.csv",
      mimeType: "text/csv",
      buffer: Buffer.from(
        `Start Time,End Time,Source IP,Destination IP,Source Port,Destination Port,Protocol,Bytes,Packets\n2026-07-10T12:00:00Z,2026-07-10T12:00:05Z,${value},2001:db8::20,443,51515,6,1200,12\n`,
      ),
    },
  });
  await page
    .getByRole("gridcell", { name: /Source IP:/u })
    .first()
    .click();
  const trigger = page.getByRole("button", {
    name: "Link 1 selected row",
    exact: true,
  });
  await trigger.focus();
  await trigger.press("Enter");
  const dialog = page.getByTestId(
    networkAnalysisTestId("indicator-link-dialog"),
  );
  const confirmation = page.getByTestId(
    networkAnalysisTestId("indicator-link-confirmation"),
  );
  await expect(
    page
      .getByRole("region", { name: "Recovery navigation", exact: true })
      .locator(":scope > h2"),
  ).toBeFocused();
  await expect(dialog).not.toHaveAttribute("aria-modal");
  await confirmation.fill(value.toUpperCase());
  await confirmation.press("Enter");
  await expect(confirmation).toHaveAttribute("aria-invalid", "true");
  const errorId = await confirmation.getAttribute("aria-describedby");
  expect(errorId).toContain("network-flow-indicator-confirmation-error");
  await expect(
    dialog.locator("#network-flow-indicator-confirmation-error"),
  ).toContainText("exactly");
  for (const viewport of [
    { width: 390, height: 720 },
    { width: 768, height: 320 },
  ]) {
    await page.setViewportSize(viewport);
    expect(
      await dialog.evaluate((node) => node.scrollWidth <= node.clientWidth + 1),
    ).toBe(true);
    await dialog
      .getByRole("button", { name: "Link Indicator", exact: true })
      .focus();
    await page.keyboard.press("Tab");
    await expect(
      dialog.getByLabel("Create or reuse indicator", { exact: true }),
    ).not.toBeFocused();
    await expect
      .poll(() =>
        dialog.evaluate((node) => !node.contains(document.activeElement)),
      )
      .toBe(true);
    await page.keyboard.press("Shift+Tab");
    await expect(
      dialog.getByRole("button", { name: "Link Indicator", exact: true }),
    ).toBeFocused();
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    expect(
      await dialog.evaluate((node) => node.scrollWidth <= node.clientWidth + 1),
    ).toBe(true);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        ".network-flow-link-dialog * { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; } .network-flow-link-dialog p { margin-block-end: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 480 });
      expect(
        await dialog.evaluate(
          (node) => node.scrollWidth <= node.clientWidth + 1,
        ),
      ).toBe(true);
      await confirmation.focus();
      await expectAllInteractiveControlsNamed(page);
      await test.info().attach("indicator-link-ipv6-spacing", {
        body: await page.screenshot(),
        contentType: "image/png",
      });
      await test.info().attach("indicator-link-accessibility-tree", {
        body: await dialog.ariaSnapshot(),
        contentType: "text/plain",
      });
    },
  );
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);
  await expect(trigger).toBeFocused();
  await trigger.press("Enter");
  const create = dialog.getByLabel("Create or reuse indicator", {
    exact: true,
  });
  await create.focus();
  await page.keyboard.press("ArrowDown");
  await expect(
    dialog.getByLabel("Existing indicator", { exact: true }),
  ).toBeChecked();
  const knownId = dialog.getByLabel("Known indicator ID (alternative)");
  await knownId.fill("invalid-id");
  await confirmation.fill(value);
  await confirmation.press("Enter");
  await expect(knownId).toHaveAttribute("aria-invalid", "true");
  await expect(knownId).toBeFocused();
  await create.focus();
  await page.keyboard.press("Space");
  await expect(confirmation).toHaveValue("");
  await confirmation.fill(value);
  await confirmation.press("Enter");
  await expect(dialog.getByRole("status")).toHaveText(
    "Indicator binding created.",
  );
  await dialog.getByRole("button", { name: "Done", exact: true }).focus();
  await page.keyboard.press("Enter");
  await expect(dialog).toHaveCount(0);
  await expect(trigger).toBeFocused();
});
