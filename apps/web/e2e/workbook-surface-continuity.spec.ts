import {
  dataTestIdSelector,
  gridGroupingSelectTestId,
  gridShellTestId,
  savedViewFamilySelector,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  saveStateTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  workbookFilterPopoverTriggerTestId,
  workbookShellReadyTestId,
  workbookShellSlotTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  handoffViewSchemaId,
  lessonViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";

import { createIncident } from "./support/incidents/fixtures";

import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";

import { createSavedView } from "./support/workbook/savedViews";

const workbookShellSurfaceCases = [
  {
    groupToken: "coordination",
    label: "Task Requests",
    viewSchemaId: taskRequestsViewSchemaId,
  },
  {
    groupToken: "coordination",
    label: "Decisions",
    viewSchemaId: decisionsViewSchemaId,
  },
  {
    groupToken: "coordination",
    label: "Parties",
    viewSchemaId: partiesViewSchemaId,
  },
  {
    groupToken: "coordination",
    label: "Communications Log",
    viewSchemaId: commLogViewSchemaId,
  },
  {
    groupToken: "coordination",
    label: "Handoff",
    viewSchemaId: handoffViewSchemaId,
  },
  {
    groupToken: "coordination",
    label: "Status Review",
    viewSchemaId: statusReviewViewSchemaId,
  },
  {
    groupToken: "coordination",
    label: "Lesson",
    viewSchemaId: lessonViewSchemaId,
  },
] as const;

test("Verify Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson open inside the same workbook shell and retain view controls.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOKREGISTRY"),
    "browser.coordination-review.row-01 workbook shell coordination surfaces",
  );
  const savedViewIdsBySurface = new Map<string, string>();
  for (const surface of workbookShellSurfaceCases) {
    const savedView = await createSavedView(page, incidentId, {
      display_name: `browser.coordination-review.row-01 ${surface.label} saved view`,
      scope: "shared",
      view_schema_id: surface.viewSchemaId,
    });
    savedViewIdsBySurface.set(surface.viewSchemaId, savedView.saved_view_id);
  }

  const firstSurface = workbookShellSurfaceCases[0];
  if (firstSurface === undefined) {
    throw new Error(
      "browser.coordination-review workbook shell surface list must not be empty",
    );
  }
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      firstSurface.viewSchemaId,
    )}`,
  );

  const shell = page.getByTestId(workbookShellReadyTestId());
  await expect(shell).toBeVisible();

  const topBar = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("top-bar")),
  );
  const viewBar = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("view-bar")),
  );
  const statusStrip = shell.locator(
    dataTestIdSelector(workbookShellSlotTestId("status-strip")),
  );
  await expect(topBar).toBeVisible();
  await expect(viewBar).toBeVisible();
  await expect(statusStrip).toBeVisible();
  await expect(
    statusStrip.getByTestId(systemViewSwitcherTriggerTestId()),
  ).toBeVisible();
  await expect(topBar.locator(savedViewFamilySelector())).toHaveCount(0);
  await expect(statusStrip.getByTestId(saveStateTestId())).toBeVisible();

  await statusStrip.getByTestId(systemViewSwitcherTriggerTestId()).click();
  const menu = page.getByTestId(systemViewSwitcherMenuTestId());
  await expect(menu).toBeVisible();
  for (const surface of workbookShellSurfaceCases) {
    await expect(
      menu.getByTestId(
        systemViewSwitcherOptionTestId(
          surface.groupToken,
          surface.viewSchemaId,
        ),
      ),
    ).toHaveAttribute("data-view-schema-id", surface.viewSchemaId);
  }
  await page.keyboard.press("Escape");
  await expect(menu).toHaveCount(0);

  for (const [index, surface] of workbookShellSurfaceCases.entries()) {
    if (index > 0) {
      const surfaceQuery = page.waitForRequest((request) => {
        return (
          request.method() === "POST" &&
          request
            .url()
            .endsWith(
              `/api/v1/incidents/${incidentId}/views/${surface.viewSchemaId}/query`,
            )
        );
      });
      await statusStrip.getByTestId(systemViewSwitcherTriggerTestId()).click();
      const option = page.getByTestId(
        systemViewSwitcherOptionTestId(
          surface.groupToken,
          surface.viewSchemaId,
        ),
      );
      await expect(option).toHaveAttribute(
        "data-view-schema-id",
        surface.viewSchemaId,
      );
      await option.click();
      await surfaceQuery;
    }

    await expect(shell).toHaveAttribute(
      "data-active-view-schema-id",
      surface.viewSchemaId,
    );
    await expect(page).toHaveURL(
      new RegExp(`view_schema_id=${encodeURIComponent(surface.viewSchemaId)}`),
    );
    await expect(
      page.getByTestId(gridShellTestId(surface.viewSchemaId)),
    ).toBeVisible();
    await expect(
      viewBar.getByTestId(savedViewSelectorTestId(surface.viewSchemaId)),
    ).toHaveCount(1);
    await expect(
      viewBar.getByTestId(
        workbookViewBarQueryControlsTestId(surface.viewSchemaId),
      ),
    ).toBeVisible();
    await expect(
      viewBar.getByTestId(
        workbookFilterPopoverTriggerTestId(surface.viewSchemaId),
      ),
    ).toBeVisible();
    await expect(
      viewBar.getByTestId(gridGroupingSelectTestId(surface.viewSchemaId)),
    ).toBeVisible();

    const activeSelector = viewBar.getByTestId(
      savedViewSelectorTestId(surface.viewSchemaId),
    );
    const savedViewId = savedViewIdsBySurface.get(surface.viewSchemaId);
    if (savedViewId === undefined) {
      throw new Error(`missing saved view for ${surface.viewSchemaId}`);
    }
    await activeSelector.click();
    const savedBrowser = page.getByRole("dialog", {
      name: "Saved views",
      exact: true,
    });
    await expect(
      savedBrowser.getByTestId(
        savedViewOptionTestId(surface.viewSchemaId, savedViewId),
      ),
    ).toHaveAttribute("data-view-schema-id", surface.viewSchemaId);
    for (const [
      otherViewSchemaId,
      otherSavedViewId,
    ] of savedViewIdsBySurface.entries()) {
      if (otherViewSchemaId === surface.viewSchemaId) {
        continue;
      }
      await expect(
        savedBrowser.getByTestId(
          savedViewOptionTestId(surface.viewSchemaId, otherSavedViewId),
        ),
      ).toHaveCount(0);
    }
    await savedBrowser
      .getByTestId(savedViewOptionTestId(surface.viewSchemaId, savedViewId))
      .click();
    await expect(activeSelector).toHaveAttribute(
      "data-selected-saved-view-id",
      savedViewId,
    );
    await expect(topBar.locator(savedViewFamilySelector())).toHaveCount(0);
    await expect(statusStrip.getByTestId(saveStateTestId())).toBeVisible();
  }
});
