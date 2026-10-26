import { assertActiveFilterChipVisible } from "@cartulary/test-utils/grid";
import {
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridSortHeaderTestId,
  savedViewActionMenuTriggerTestId,
  savedViewCreateButtonTestId,
  savedViewModifiedTestId,
  savedViewNameInputTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewStartupToggleTestId,
  savedViewStatusTestId,
  timelineRowMarkReviewedButtonTestId,
  workbookAddRowButtonTestId,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
  workbookInspectorToggleTestId,
  workbookPreferenceTestId,
  workbookQueryEntryTestId,
  workbookQueryOverflowEntryTestId,
  workbookResponsiveBandTestId,
  workbookShellReadyTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { openIncidentControls } from "./pages/deploymentAdministration";
import { openIncidentFromLanding } from "./pages/incidentDirectory";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import {
  expectNoFocusTrap,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import {
  expectAndRecordContrast,
  expectLocatorContrast,
} from "./support/accessibility/records";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installOwnedRoute } from "./support/transport/requestInterception";
import {
  mountedGridCell,
  mountedGridTarget,
} from "./support/workbook/keyboardGrid";
import { expectTextSpacingViewBarResilience } from "./support/workbook/presentationGeometry";
import { createViewRow } from "./support/workbook/query";
import { openFilterPopover } from "./support/workbook/queryControls";
import { clickTimelineRowAction } from "./support/workbook/rowMutations";

test("a11y.saved query controls and startup actions preserve keyboard state", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const longSavedViewName =
    "a11y.saved-view-query workbook layout resilience with a deliberately long selected view name";
  const longTagToken =
    "unbroken-tag-token-0123456789-abcdefghijklmnopqrstuvwxyz";
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YSAVEDVIEWQUERY"),
    "a11y.saved-view-query query controls",
  );
  const reviewedRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.saved-view-query-reviewed"),
      "timeline.activity_synopsis_text": "a11y.saved-view-query reviewed row",
      "timeline.tags": {
        actions: [{ op: "add_tag", tag_name: longTagToken }],
        kind: "collection_actions_v1",
      },
    },
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.saved-view-query-rough"),
    "timeline.activity_synopsis_text": "a11y.saved-view-query rough row",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const summarySortHeader = await mountedGridTarget(
    page,
    timelineViewSchemaId,
    gridSortHeaderTestId(
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    ),
  );
  await expectVisibleFocus(summarySortHeader);
  await summarySortHeader.press("Enter");
  await expect(summarySortHeader).toContainText("Asc");
  await clickTimelineRowAction(
    page,
    reviewedRow.record_id,
    timelineRowMarkReviewedButtonTestId(reviewedRow.record_id),
  );
  await expect(
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      reviewedRow.record_id,
      "timeline.capture_state",
    ),
  ).toHaveText("reviewed");
  await openFilterPopover(page, timelineViewSchemaId);
  const filterField = page.getByTestId(
    gridFilterFieldTestId(timelineViewSchemaId),
  );
  await expectVisibleFocus(filterField);
  await filterField.selectOption("timeline.capture_state");
  const filterValue = page.getByTestId(
    gridFilterValueTestId(timelineViewSchemaId),
  );
  await expectVisibleFocus(filterValue);
  try {
    await filterValue.selectOption("reviewed");
  } catch {
    await filterValue.fill("reviewed");
  }
  const filterApply = page.getByTestId(
    gridFilterApplyTestId(timelineViewSchemaId),
  );
  await expectVisibleFocus(filterApply);
  await filterApply.press("Enter");
  await assertActiveFilterChipVisible(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
  );
  const filterChip = page.getByTestId(
    workbookQueryEntryTestId(
      timelineViewSchemaId,
      "filter",
      "timeline.capture_state",
    ),
  );
  await expectVisibleFocus(filterChip);
  const sortMenuTrigger = page.getByTestId(
    workbookSortMenuTriggerTestId(timelineViewSchemaId),
  );
  await expectVisibleFocus(sortMenuTrigger);
  await sortMenuTrigger.press("Enter");
  const sortMenu = page.getByTestId(
    workbookSortMenuTestId(timelineViewSchemaId),
  );
  await expect(sortMenu).toBeVisible();
  for (const fieldKey of [
    "timeline.activity_sort_ts",
    "timeline.date_entered_sort_day",
    "timeline.analyst_text",
    "timeline.mitre_stage_text",
    "timeline.device_object_text",
    "timeline.has_evidence",
    "timeline.capture_state",
  ]) {
    const option = page.getByTestId(
      workbookSortOptionTestId(timelineViewSchemaId, fieldKey),
    );
    await option.click();
    await expect(option).toHaveAttribute("aria-checked", "true");
  }
  await sortMenu.press("Escape");
  await expect(sortMenu).toHaveCount(0);
  await expect(sortMenuTrigger).toBeFocused();
  await expect(sortMenuTrigger).toHaveAttribute(
    "aria-label",
    "Sort, 8 applied",
  );
  await expect(sortMenuTrigger).not.toHaveAttribute("title");
  await openFilterPopover(page, timelineViewSchemaId);
  await filterField.selectOption("timeline.tags");
  await filterValue.fill(longTagToken);
  await filterApply.press("Enter");
  const groupingSelect = page.getByTestId(
    gridGroupingSelectTestId(timelineViewSchemaId),
  );
  await expectVisibleFocus(groupingSelect);
  await groupingSelect.selectOption("timeline.capture_state");
  await expect(groupingSelect).toHaveAttribute("title", "Capture State");
  const reviewedGroup = page.getByTestId(
    gridGroupRowTestId(
      timelineViewSchemaId,
      "timeline.capture_state",
      "reviewed",
    ),
  );
  await expect(reviewedGroup).toBeVisible();
  await expectVisibleFocus(reviewedGroup);
  await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");
  await reviewedGroup.press("Enter");
  await expect(reviewedGroup).toHaveAttribute("aria-expanded", "false");
  await reviewedGroup.press("Enter");
  await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");
  const savedViewSelector = page.getByTestId(
    savedViewSelectorTestId(timelineViewSchemaId),
  );
  await expectVisibleFocus(savedViewSelector);
  await expect(savedViewSelector).toHaveAttribute(
    "data-selected-sheet-ref-kind",
    "view_schema",
  );
  await openSavedViewActionMenu(page, timelineViewSchemaId);
  const savedViewNameInput = page.getByTestId(
    savedViewNameInputTestId(timelineViewSchemaId),
  );
  await expectVisibleFocus(savedViewNameInput);
  await savedViewNameInput.fill(longSavedViewName);
  const createSavedViewButton = page.getByTestId(
    savedViewCreateButtonTestId(timelineViewSchemaId),
  );
  await expectVisibleFocus(createSavedViewButton);
  await createSavedViewButton.press("Enter");
  const savedViewStatus = page.getByTestId(
    savedViewStatusTestId(timelineViewSchemaId),
  );
  await expect(savedViewStatus).toHaveAttribute("aria-live", "polite");
  await expect(savedViewStatus).toHaveText("Saved view created.");
  await expect(savedViewSelector).toHaveAttribute(
    "data-selected-sheet-ref-kind",
    "saved_view",
  );
  await expect(savedViewSelector).toHaveAttribute("title", longSavedViewName);
  await openSavedViewActionMenu(page, timelineViewSchemaId);
  const homeButton = page.getByTestId(
    savedViewSetHomeButtonTestId(timelineViewSchemaId),
  );
  const startup = page.getByTestId(
    savedViewStartupToggleTestId(timelineViewSchemaId),
  );
  await expect(startup).toHaveAttribute("aria-expanded", "false");
  await expectVisibleFocus(startup);
  await startup.press("Enter");
  await expectVisibleFocus(homeButton);
  await homeButton.press("Enter");
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "shortcut-outcome")),
  ).toHaveText("Home update confirmed.");
  await openSavedViewActionMenu(page, timelineViewSchemaId);
  const defaultButton = page.getByTestId(
    savedViewSetDefaultButtonTestId(timelineViewSchemaId),
  );
  await expectVisibleFocus(defaultButton);
  await defaultButton.press("Enter");
  await expect(
    page.getByTestId(workbookPreferenceTestId("default", "shortcut-outcome")),
  ).toHaveText("Incident default update confirmed.");
  await expect(
    page.getByTestId(workbookPreferenceTestId("default", "shortcut-outcome")),
  ).toBeVisible();
  await expect(defaultButton).toBeFocused();
  await startup.click();
  await expect(defaultButton).toHaveCount(0);
  await expect(
    page.getByTestId(workbookPreferenceTestId("default", "shortcut-outcome")),
  ).toHaveText("Incident default update confirmed.");
  await page.keyboard.press("Escape");
  await expect(
    page.getByTestId(savedViewActionMenuTriggerTestId(timelineViewSchemaId)),
  ).toBeFocused();
});

test("a11y.query overflow remains reachable through density enlargement reflow and text spacing", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const longSavedViewName =
    "a11y.saved-view-query workbook layout resilience with a deliberately long selected view name";
  const longTagToken =
    "unbroken-tag-token-0123456789-abcdefghijklmnopqrstuvwxyz";
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YSAVEDVIEWQUERY"),
    "a11y.saved-view-query query controls",
  );
  const reviewedRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.saved-view-query-reviewed"),
      "timeline.activity_synopsis_text": "a11y.saved-view-query reviewed row",
      "timeline.tags": {
        actions: [{ op: "add_tag", tag_name: longTagToken }],
        kind: "collection_actions_v1",
      },
    },
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.saved-view-query-rough"),
    "timeline.activity_synopsis_text": "a11y.saved-view-query rough row",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await clickTimelineRowAction(
    page,
    reviewedRow.record_id,
    timelineRowMarkReviewedButtonTestId(reviewedRow.record_id),
  );
  const saved = await createSavedView(page, incidentId, {
    display_name: longSavedViewName,
    view_schema_id: timelineViewSchemaId,
    query_json: {
      group_by: "timeline.capture_state",
      filters: [
        {
          field_key: "timeline.capture_state",
          op: "eq",
          arg: { value: "reviewed" },
        },
        {
          field_key: "timeline.tags",
          op: "contains_any",
          arg: { values: [longTagToken] },
        },
      ],
      sort: [
        "timeline.activity_synopsis_text",
        "timeline.activity_sort_ts",
        "timeline.date_entered_sort_day",
        "timeline.analyst_text",
        "timeline.mitre_stage_text",
        "timeline.device_object_text",
        "timeline.has_evidence",
        "timeline.capture_state",
      ].map((field_key) => ({ direction: "asc", field_key })),
    },
  });
  await page.reload();
  await selectSavedView(page, timelineViewSchemaId, saved.saved_view_id);
  const groupingSelect = page.getByTestId(
    gridGroupingSelectTestId(timelineViewSchemaId),
  );
  const reviewedGroup = page.getByTestId(
    gridGroupRowTestId(
      timelineViewSchemaId,
      "timeline.capture_state",
      "reviewed",
    ),
  );
  await groupingSelect.selectOption("timeline.has_evidence");
  await expect(
    page.getByTestId(savedViewModifiedTestId(timelineViewSchemaId)),
  ).toHaveText("Modified");
  await expect(page.locator('[data-grid-data-state="refreshing"]')).toHaveCount(
    0,
  );
  await expect(
    page.locator(
      '[data-grid-data-state="stale_error"], [data-grid-data-state="unavailable"]',
    ),
  ).toHaveCount(0);
  const queryControls = page.getByTestId(
    workbookViewBarQueryControlsTestId(timelineViewSchemaId),
  );
  await expect(queryControls).toHaveAttribute(
    "data-hidden-query-chip-count",
    "8",
  );
  const filterTrigger = page.getByTestId(
    workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
  );
  await expect(filterTrigger).toHaveAttribute(
    "aria-label",
    "Filters, 2 active filters, 8 hidden query entries",
  );
  await expectVisibleFocus(filterTrigger);
  await filterTrigger.press("Enter");
  const filterPopover = page.getByTestId(
    workbookFilterPopoverTestId(timelineViewSchemaId),
  );
  await expect(filterPopover).toBeVisible();
  await expect(filterTrigger).toHaveAttribute(
    "aria-controls",
    workbookFilterPopoverTestId(timelineViewSchemaId),
  );
  await expect(
    filterPopover.getByTestId(
      workbookQueryOverflowEntryTestId(
        timelineViewSchemaId,
        "filter",
        "timeline.tags",
      ),
    ),
  ).toContainText(longTagToken);
  await filterPopover.press("Escape");
  await expect(filterPopover).toHaveCount(0);
  await expect(filterTrigger).toBeFocused();
  await withPresentation(
    page,
    {
      textSpacing: `
        * {
          line-height: 1.5 !important;
          letter-spacing: 0.12em !important;
          word-spacing: 0.16em !important;
        }
        p {
          margin-block-end: 2em !important;
        }
      `,
    },
    async () => {
      await expectTextSpacingViewBarResilience(page, {
        capacity: 3,
        height: 900,
        hiddenCount: 8,
        width: 1440,
      });
      await expectTextSpacingViewBarResilience(page, {
        capacity: 2,
        height: 720,
        hiddenCount: 9,
        width: 1024,
      });
      await page
        .getByRole("toolbar", { name: "Active query chips" })
        .getByRole("button")
        .first()
        .focus();
      await expectTextSpacingViewBarResilience(page, {
        capacity: 0,
        height: 640,
        hiddenCount: 11,
        width: 768,
      });
      await expect(filterTrigger).toBeFocused();
      await expectVisibleFocus(
        page.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
      );
      await expectVisibleFocus(
        page.getByTestId(workbookAddRowButtonTestId(timelineViewSchemaId)),
      );
    },
  );
  await page.setViewportSize({ width: 1440, height: 900 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expect(
      page.getByTestId(workbookResponsiveBandTestId()),
    ).toHaveAttribute(
      "data-workbook-responsive-band",
      "below_supported_minimum",
    );
    await expect(
      page.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(workbookAddRowButtonTestId(timelineViewSchemaId)),
    ).toBeVisible();
    const zoomedDocumentGeometry = await page.evaluate(() => ({
      clientWidth: document.documentElement.clientWidth,
      scrollWidth: document.documentElement.scrollWidth,
    }));
    expect(zoomedDocumentGeometry.scrollWidth).toBeLessThanOrEqual(
      zoomedDocumentGeometry.clientWidth + 1,
    );
  });
  await groupingSelect.selectOption("timeline.capture_state");
  await expect(reviewedGroup).toBeVisible();
  await expectLocatorContrast(
    await mountedGridTarget(
      page,
      timelineViewSchemaId,
      gridSortHeaderTestId(
        timelineViewSchemaId,
        "timeline.activity_synopsis_text",
      ),
    ),
    "query-sort-header",
  );
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  await openFilterPopover(page, timelineViewSchemaId);
  await expectAndRecordContrast(
    page,
    [
      gridFilterFieldTestId(timelineViewSchemaId),
      gridFilterValueTestId(timelineViewSchemaId),
      gridFilterApplyTestId(timelineViewSchemaId),
    ],
    "filter-popover",
  );
  await openSavedViewActionMenu(page, timelineViewSchemaId);
  await page
    .getByTestId(savedViewStartupToggleTestId(timelineViewSchemaId))
    .click();
  await expectAndRecordContrast(page, [
    workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
    gridGroupingSelectTestId(timelineViewSchemaId),
    gridGroupRowTestId(
      timelineViewSchemaId,
      "timeline.capture_state",
      "reviewed",
    ),
    savedViewSelectorTestId(timelineViewSchemaId),
    savedViewNameInputTestId(timelineViewSchemaId),
    savedViewCreateButtonTestId(timelineViewSchemaId),
    savedViewSetHomeButtonTestId(timelineViewSchemaId),
    savedViewSetDefaultButtonTestId(timelineViewSchemaId),
    savedViewStatusTestId(timelineViewSchemaId),
  ]);
});

test("a11y.saved-view uncertain creation recovery preserves keyboard focus and draft", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YSAVEDVIEWQUERY"),
    "a11y.saved-view-query query controls",
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openSavedViewActionMenu(page, timelineViewSchemaId);
  const savedViewNameInput = page.getByTestId(
    savedViewNameInputTestId(timelineViewSchemaId),
  );
  const createSavedViewButton = page.getByTestId(
    savedViewCreateButtonTestId(timelineViewSchemaId),
  );
  await installOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/saved-views`,
    async (route) => {
      if (route.request().method() !== "POST") {
        await route.continue();
        return;
      }
      const response = await route.fetch();
      expect(response.status()).toBe(201);
      await route.abort("failed");
    },
  );
  await savedViewNameInput.fill("Accessible recovery draft");
  await createSavedViewButton.press("Enter");
  const recovery = page.getByRole("region", {
    name: "Saved-view operation",
    exact: true,
  });
  const finishRecovery = recovery.getByRole("button", {
    name: "End recovery without another write",
    exact: true,
  });
  await expect(finishRecovery).toBeEnabled();
  await savedViewNameInput.press("Escape");
  await expect(
    page.getByTestId(savedViewActionMenuTriggerTestId(timelineViewSchemaId)),
  ).toBeFocused();
  await page.setViewportSize({ width: 768, height: 640 });
  await openSavedViewActionMenu(page, timelineViewSchemaId);
  const recoveryName = page.getByLabel("Name for reviewed saved-view request", {
    exact: true,
  });
  await expectVisibleFocus(recoveryName);
  await recoveryName.fill("Accessible revised name");
  await recoveryName.press("Tab");
  const riskConfirmation = recovery.getByRole("checkbox", {
    name: /may create a duplicate/,
  });
  await expect(riskConfirmation).toBeFocused();
  await riskConfirmation.press("Space");
  await riskConfirmation.press("Tab");
  await expect(
    recovery.getByRole("button", {
      name: "Make a new create attempt",
      exact: true,
    }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(finishRecovery).toBeFocused();
  await finishRecovery.press("Escape");
  await expect(
    page.getByTestId(savedViewActionMenuTriggerTestId(timelineViewSchemaId)),
  ).toBeFocused();
  await openSavedViewActionMenu(page, timelineViewSchemaId);
  await expect(recoveryName).toHaveValue("Accessible revised name");
  await expectAllInteractiveControlsNamed(page);
  await expectVisibleFocus(finishRecovery);
  await finishRecovery.press("Enter");
  await expect(recovery).toContainText("Recovery ended by your choice");
});

test("a11y.preferences independent clear and recovery controls preserve keyboard focus and responsive reachability", async ({
  testActorPage: page,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WP-A11Y"),
    "Preference accessibility",
  );
  await openIncidentFromLanding(page, incidentId);
  await openIncidentControls(page);
  const panel = page.getByRole("region", {
    name: "Workbook startup preferences",
    exact: true,
  });
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const kind of ["home", "default"] as const)
      for (const action of ["set", "clear", "refresh"] as const) {
        const control = page.getByTestId(
          workbookPreferenceTestId(kind, action),
        );
        await control.scrollIntoViewIfNeeded();
        await expectVisibleFocus(control);
        expect(
          await panel.evaluate(
            (node) => node.scrollWidth <= node.clientWidth + 1,
          ),
        ).toBe(true);
      }
  }
  const home = page.getByTestId(workbookPreferenceTestId("home", "clear"));
  await home.focus();
  await page.keyboard.press("Enter");
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "outcome")),
  ).toHaveText("Home clear confirmed.");
  await expect(home).toBeFocused();
  await installOwnedRoute(
    page,
    `**/api/v1/incidents/${incidentId}/workbook-preferences/me`,
    async (route) => {
      if (route.request().method() === "PUT") {
        const response = await route.fetch();
        expect(response.status()).toBe(200);
        await route.abort("failed");
      } else await route.continue();
    },
  );
  await home.press("Enter");
  const keep = page.getByTestId(
    workbookPreferenceTestId("home", "keep-observed"),
  );
  await expect(keep).toHaveAttribute("aria-disabled", "false");
  await expect(home).toBeFocused();
  await expectVisibleFocus(keep);
  await page.keyboard.press("Enter");
  await expect(
    page.getByTestId(workbookPreferenceTestId("home", "refresh")),
  ).toBeFocused();
  await expect(
    page.getByRole("status", {
      name: "Workbook preference updates",
      exact: true,
    }),
  ).toHaveCount(1);
  await expect(panel.getByRole("status")).toHaveCount(0);
  await expectAllInteractiveControlsNamed(page);
  await testInfo.attach("preferences-accessibility-tree", {
    body: await panel.ariaSnapshot(),
    contentType: "text/plain",
  });
  await page.keyboard.press("Escape");
  await expect(
    page.getByRole("button", { name: "Account and application navigation" }),
  ).toBeFocused();
  await page.setViewportSize({ width: 1280, height: 720 });
  await openSavedViewActionMenu(page, timelineViewSchemaId);
  await page
    .getByTestId(savedViewStartupToggleTestId(timelineViewSchemaId))
    .click();
  await page
    .getByRole("button", {
      name: "Inspect and recover workbook preferences…",
      exact: true,
    })
    .click();
  await expect(panel).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Close incident controls", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(
    page.getByTestId(savedViewActionMenuTriggerTestId(timelineViewSchemaId)),
  ).toBeFocused();
});

test("a11y.saved-view-discovery keyboard activation dismissal scope and bounded layout survive density zoom and text spacing", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("SVDA11Y"),
    "Saved-view discovery accessibility",
  );
  const longName = "A long saved-view name for investigation continuity "
    .repeat(4)
    .trim();
  const created = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/saved-views`,
    {
      headers: await csrfHeaders(page),
      data: {
        display_name: longName,
        scope: "shared",
        view_schema_id: timelineViewSchemaId,
        query_json: {
          sort: [],
          filters: [],
          group_by: "timeline.capture_state",
        },
        layout_json: {},
      },
    },
  );
  expect(created.status()).toBe(201);
  const saved = (await created.json()).data;
  const before = (
    await (
      await page.request.get(`${apiBase}/api/v1/account/preferences`)
    ).json()
  ).data;
  try {
    for (const density of ["compact", "default", "comfortable"]) {
      const current = (
        await (
          await page.request.get(`${apiBase}/api/v1/account/preferences`)
        ).json()
      ).data;
      expect(
        (
          await page.request.put(`${apiBase}/api/v1/account/preferences`, {
            headers: await csrfHeaders(page),
            data: {
              base_preferences_version: current.preferences_version,
              client_txn_id: uniqueTxn("svd-density"),
              density_mode: density,
            },
          })
        ).ok(),
      ).toBe(true);
      await page.goto(`/?incident_id=${incidentId}`);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      for (const width of [1440, 768]) {
        await page.setViewportSize({ width, height: 900 });
        const trigger = page.getByTestId(
          savedViewSelectorTestId(timelineViewSchemaId),
        );
        await trigger.focus();
        await trigger.press("Enter");
        const browser = page.getByRole("dialog", {
          name: "Saved views",
          exact: true,
        });
        const candidate = browser.getByRole("option", {
          name: /A long saved-view name/,
        });
        await expect(candidate).toBeVisible();
        await expect(candidate).toContainText("shared");
        const firstPage = browser.getByRole("button", {
          name: "First",
          exact: true,
        });
        // Cached options can remain visible while discovery refresh disables paging.
        // Start the keyboard sequence only after that read has been admitted.
        await expect(firstPage).toBeEnabled();
        await browser
          .getByRole("option", { name: "Unsaved view", exact: true })
          .press("End");
        await expect(candidate).toBeFocused();
        await candidate.press("Tab");
        await expect(firstPage).toBeFocused();
        await browser.press("Escape");
        await expect(trigger).toBeFocused();
        await expect(trigger).toHaveAttribute(
          "data-selected-sheet-ref-kind",
          "view_schema",
        );
        await trigger.press("Enter");
        await expect(candidate).toBeVisible();
        await expect(firstPage).toBeEnabled();
        await candidate.focus();
        await candidate.press("Space");
        await expect(trigger).toHaveAttribute(
          "data-selected-saved-view-id",
          saved.saved_view_id,
        );
        await expect(
          page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
        ).toHaveValue("timeline.capture_state");
        if (width === 1440) {
          await page
            .getByTestId(gridGroupingSelectTestId(timelineViewSchemaId))
            .selectOption("");
          const modified = page.getByTestId(
            savedViewModifiedTestId(timelineViewSchemaId),
          );
          await expect(modified).toBeVisible();
          const triggerBounds = await trigger.boundingBox();
          const modifiedBounds = await modified.boundingBox();
          if (!triggerBounds || !modifiedBounds)
            throw new Error("Saved-view controls are not rendered");
          expect(triggerBounds.x + triggerBounds.width).toBeLessThanOrEqual(
            modifiedBounds.x,
          );
          await expect(trigger).toContainText("shared");
        }
        await trigger.press("Enter");
        await expect(browser).toBeVisible();
        const bounds = await browser.boundingBox();
        expect(bounds).not.toBeNull();
        if (!bounds) throw new Error("Saved-view browser is not rendered");
        expect(bounds.x).toBeGreaterThanOrEqual(0);
        expect(bounds.x + bounds.width).toBeLessThanOrEqual(width + 1);
        await expectAllInteractiveControlsNamed(page);
        await expectAndRecordContrast(page, [
          savedViewSelectorTestId(timelineViewSchemaId),
        ]);
        await test.info().attach(`saved-view-browser-${density}-${width}`, {
          body: await page.screenshot({
            animations: "disabled",
            caret: "hide",
          }),
          contentType: "image/png",
        });
        await browser
          .getByRole("option", { name: "Unsaved view", exact: true })
          .click();
      }
    }
    await page.setViewportSize({ width: 1440, height: 900 });
    await withPresentation(
      page,
      {
        textSpacing:
          "* {line-height:1.5!important;letter-spacing:0.12em!important;word-spacing:0.16em!important} p {margin-block-end:2em!important}",
      },
      async () => {
        await withPresentation(page, { zoom: "200%" }, async () => {
          // At 720 effective CSS pixels the existing shell hides configuration
          // controls. Exercise the browser at a supported zoomed width instead.
          await expect(
            page.getByTestId(workbookResponsiveBandTestId()),
          ).toHaveAttribute(
            "data-workbook-responsive-band",
            "below_supported_minimum",
          );
          await page.setViewportSize({ width: 1920, height: 1200 });
          const trigger = page.getByTestId(
            savedViewSelectorTestId(timelineViewSchemaId),
          );
          await trigger.click();
          const browser = page.getByRole("dialog", {
            name: "Saved views",
            exact: true,
          });
          await expect(
            browser.getByRole("option", { name: /A long saved-view name/ }),
          ).toBeVisible();
          const bounds = await browser.boundingBox();
          const anchor = await trigger.boundingBox();
          if (!bounds || !anchor)
            throw new Error("Zoomed saved-view controls are not rendered");
          expect(bounds.x + bounds.width).toBeLessThanOrEqual(1921);
          expect(bounds.y + bounds.height).toBeLessThanOrEqual(1201);
          expect(Math.abs(bounds.y - anchor.y - anchor.height)).toBeLessThan(2);
          await test.info().attach("saved-view-browser-zoom-spacing", {
            body: await page.screenshot({
              animations: "disabled",
              caret: "hide",
            }),
            contentType: "image/png",
          });
          await browser.press("Escape");
          await expect(trigger).toBeFocused();
          // Outside dismissal keeps the newly chosen focus destination.
          await trigger.click();
          const outside = page.getByRole("button", {
            name: "Query controls",
            exact: true,
          });
          await outside.click();
          await expect(browser).toHaveCount(0);
          await expect(outside).toBeFocused();
        });
      },
    );
  } finally {
    const current = (
      await (
        await page.request.get(`${apiBase}/api/v1/account/preferences`)
      ).json()
    ).data;
    await page.request.put(`${apiBase}/api/v1/account/preferences`, {
      headers: await csrfHeaders(page),
      data: {
        base_preferences_version: current.preferences_version,
        client_txn_id: uniqueTxn("svd-density-restore"),
        density_mode: before.density_mode,
      },
    });
  }
});

import {
  createSavedView,
  openSavedViewActionMenu,
  selectSavedView,
} from "./support/workbook/savedViews";
