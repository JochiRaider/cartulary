import type { QueryWorkbookViewRequest } from "@cartulary/protocol-ts/http";
import { changeGrouping } from "@cartulary/test-utils/grid";
import {
  gridGroupingSelectTestId,
  gridGroupRowsSelector,
  gridGroupRowTestId,
  workbookQueryEntryTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";

import { createIncident } from "./support/incidents/fixtures";

import { responseBarrier } from "./support/runtime/cleanup";

import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";

import {
  awaitResponseBarrier,
  installOwnedRoute,
} from "./support/transport/requestInterception";

import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";

import { seed } from "./support/workbook/queryBrowsing";

test("Workbook Group retains requested keyboard choice during delayed replacement", async ({
  page,
  testActor,
}, testInfo) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-GROUP-PENDING"),
    "Workbook pending group choice",
  );
  await seed(page, incident, timelineViewSchemaId, 2, testActor.user_id);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
  );
  const grouping = page.getByTestId(
    gridGroupingSelectTestId(timelineViewSchemaId),
  );
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  const acceptedChip = page.getByTestId(
    workbookQueryEntryTestId(
      timelineViewSchemaId,
      "group",
      "timeline.capture_state",
    ),
  );
  const acceptedGroup = page.getByTestId(
    gridGroupRowTestId(timelineViewSchemaId, "timeline.capture_state", "rough"),
  );
  await expect(acceptedChip).toBeVisible();
  await expect(acceptedGroup).toBeVisible();
  let releaseOlder = () => {};
  let releaseLatest = () => {};
  const olderGateBarrier = responseBarrier(page);
  releaseOlder = olderGateBarrier.release;
  const olderGate = olderGateBarrier.promise;
  const latestGateBarrier = responseBarrier(page);
  releaseLatest = latestGateBarrier.release;
  const latestGate = latestGateBarrier.promise;
  const held = new Set<string>();
  const settled = new Set<string>();
  const requested: QueryWorkbookViewRequest[] = [];
  await installOwnedRoute(
    page,
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      const request = route
        .request()
        .postDataJSON() as QueryWorkbookViewRequest;
      const groupBy = request.group_by;
      if (
        groupBy !== "timeline.has_evidence" &&
        groupBy !== "timeline.has_unresolved_mentions"
      ) {
        await route.continue();
        return;
      }
      requested.push(request);
      const response = await route.fetch();
      held.add(groupBy);
      if (
        !(await awaitResponseBarrier(
          route,
          groupBy === "timeline.has_evidence" ? olderGate : latestGate,
        ))
      )
        return;
      try {
        await route.fulfill({ response });
      } finally {
        settled.add(groupBy);
      }
    },
  );
  try {
    await grouping.click();
    await grouping.press("ArrowDown");
    await grouping.press("Enter");
    await expect.poll(() => held.has("timeline.has_evidence")).toBe(true);
    await testInfo.attach("group-pending-baseline", {
      body: JSON.stringify({
        selected: await grouping.inputValue(),
        focused: await grouping.evaluate(
          (element) => document.activeElement === element,
        ),
        acceptedChip: await acceptedChip.isVisible(),
        acceptedGroup: await acceptedGroup.isVisible(),
        requested,
      }),
      contentType: "application/json",
    });
    await expect(grouping).toHaveValue("timeline.has_evidence");
    await expect(grouping).toBeFocused();
    await expect(acceptedChip).toBeVisible();
    await expect(acceptedGroup).toBeVisible();
    await grouping.press("ArrowDown");
    await grouping.press("Enter");
    await expect
      .poll(() => held.has("timeline.has_unresolved_mentions"))
      .toBe(true);
    await expect(grouping).toHaveValue("timeline.has_unresolved_mentions");
    await expect(grouping).toBeFocused();
    await expect(acceptedChip).toBeVisible();
    await expect(acceptedGroup).toBeVisible();
    releaseLatest();
    const latestChip = page.getByTestId(
      workbookQueryEntryTestId(
        timelineViewSchemaId,
        "group",
        "timeline.has_unresolved_mentions",
      ),
    );
    await expect(latestChip).toBeVisible();
    await expect(
      page.locator(
        gridGroupRowsSelector(
          timelineViewSchemaId,
          "timeline.has_unresolved_mentions",
        ),
      ),
    ).not.toHaveCount(0);
    await expect(grouping).toBeFocused();
    releaseOlder();
    await expect.poll(() => settled.has("timeline.has_evidence")).toBe(true);
    await expect(grouping).toHaveValue("timeline.has_unresolved_mentions");
    await expect(latestChip).toBeVisible();
    await expect(acceptedChip).toHaveCount(0);
    await expect(acceptedGroup).toHaveCount(0);
    await expect(grouping).toBeFocused();
    expect(requested.map((entry) => entry.group_by)).toEqual([
      "timeline.has_evidence",
      "timeline.has_unresolved_mentions",
    ]);
    await latestChip.click();
    await expect(grouping).toBeFocused();
    await grouping.press("Escape");
    await expect(latestChip).toBeFocused();
    await expect(grouping).toHaveValue("timeline.has_unresolved_mentions");

    releaseLatest();
    releaseOlder();
  } finally {
    latestGateBarrier.cancel();
    olderGateBarrier.cancel();
  }
});

test("Workbook Group retains requested keyboard choice after failed replacement", async ({
  page,
  testActor,
}, testInfo) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-GROUP-FAIL"),
    "Workbook failed group choice",
  );
  await seed(page, incident, timelineViewSchemaId, 2, testActor.user_id);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
  );
  const grouping = page.getByTestId(
    gridGroupingSelectTestId(timelineViewSchemaId),
  );
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  const acceptedChip = page.getByTestId(
    workbookQueryEntryTestId(
      timelineViewSchemaId,
      "group",
      "timeline.capture_state",
    ),
  );
  const acceptedGroup = page.getByTestId(
    gridGroupRowTestId(timelineViewSchemaId, "timeline.capture_state", "rough"),
  );
  await expect(acceptedChip).toBeVisible();
  await expect(acceptedGroup).toBeVisible();
  const requests: QueryWorkbookViewRequest[] = [];
  let failEvidence = true;
  let failNone = false;
  let releaseNone = () => {};
  const noneGateBarrier = responseBarrier(page);
  releaseNone = noneGateBarrier.release;
  const noneGate = noneGateBarrier.promise;
  await installOwnedRoute(
    page,
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      const request = route
        .request()
        .postDataJSON() as QueryWorkbookViewRequest;
      requests.push(request);
      if (request.group_by === "timeline.has_evidence" && failEvidence) {
        failEvidence = false;
        await route.abort("failed");
        return;
      }
      if (request.group_by === undefined && failNone) {
        failNone = false;
        if (!(await awaitResponseBarrier(route, noneGate))) return;
        await route.abort("failed");
        return;
      }
      await route.continue();
    },
  );
  await grouping.click();
  await grouping.press("ArrowDown");
  await grouping.press("Enter");
  await expect
    .poll(() => requests.at(-1)?.group_by)
    .toBe("timeline.has_evidence");
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  const retry = browsing.getByRole("button", { name: "Retry", exact: true });
  await expect(retry).toBeVisible();
  await testInfo.attach("group-failed-baseline", {
    body: JSON.stringify({
      selected: await grouping.inputValue(),
      focused: await grouping.evaluate(
        (element) => document.activeElement === element,
      ),
      acceptedChip: await acceptedChip.isVisible(),
      acceptedGroup: await acceptedGroup.isVisible(),
      acceptedGroupCount: await page
        .locator(
          gridGroupRowsSelector(timelineViewSchemaId, "timeline.capture_state"),
        )
        .count(),
      requested: requests.at(-1),
    }),
    contentType: "application/json",
  });
  await expect(grouping).toHaveValue("timeline.has_evidence");
  await expect(grouping).toBeFocused();
  await expect(acceptedChip).toBeVisible();
  await expect(acceptedGroup).toBeVisible();
  await expect(grouping).toHaveAttribute("aria-describedby", /unapplied/u);
  const groupStatus = page.locator(
    `[id="${gridGroupingSelectTestId(timelineViewSchemaId)}-unapplied"]`,
  );
  await expect(groupStatus).toContainText(
    "retained results grouped by Capture State",
  );
  await retry.click();
  const evidenceChip = page.getByTestId(
    workbookQueryEntryTestId(
      timelineViewSchemaId,
      "group",
      "timeline.has_evidence",
    ),
  );
  await expect(evidenceChip).toBeVisible();
  await expect(
    page.locator(
      gridGroupRowsSelector(timelineViewSchemaId, "timeline.has_evidence"),
    ),
  ).not.toHaveCount(0);
  await expect(grouping).toHaveValue("timeline.has_evidence");
  await expect(acceptedChip).toHaveCount(0);
  expect(requests.at(-1)?.group_by).toBe("timeline.has_evidence");

  failNone = true;
  const beforeNone = requests.length;
  await grouping.click();
  await grouping.selectOption("");
  try {
    await expect.poll(() => requests.length).toBeGreaterThan(beforeNone);
    expect(requests.at(-1)?.group_by).toBeUndefined();
    await expect(grouping).toHaveValue("");
    await expect(evidenceChip).toBeVisible();
    await expect(
      page.locator(
        gridGroupRowsSelector(timelineViewSchemaId, "timeline.has_evidence"),
      ),
    ).not.toHaveCount(0);

    releaseNone();
  } finally {
    noneGateBarrier.cancel();
  }
  await expect(retry).toBeVisible();
  await expect(grouping).toHaveValue("");
  await expect(evidenceChip).toBeVisible();
  await browsing.getByRole("button", { name: "Revert", exact: true }).click();
  await expect(grouping).toHaveValue("timeline.has_evidence");
  await expect(evidenceChip).toBeVisible();
  await expect(groupStatus).toHaveCount(0);

  await grouping.click();
  await grouping.selectOption("");
  await expect(grouping).toHaveValue("");
  await expect(evidenceChip).toHaveCount(0);
  await expect(
    page.locator(
      gridGroupRowsSelector(timelineViewSchemaId, "timeline.has_evidence"),
    ),
  ).toHaveCount(0);
  expect(requests.at(-1)?.group_by).toBeUndefined();
});

test("Workbook Group pointer selection stays scoped to Hosts", async ({
  page,
  testActor,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-GROUP-HOST"),
    "Workbook Host group choice",
  );
  await seed(page, incident, hostsViewSchemaId, 2, testActor.user_id);
  await seed(page, incident, timelineViewSchemaId, 1, testActor.user_id);
  const requests: QueryWorkbookViewRequest[] = [];
  await installOwnedRoute(
    page,
    `**/incidents/${incident}/views/${hostsViewSchemaId}/query`,
    async (route) => {
      requests.push(route.request().postDataJSON() as QueryWorkbookViewRequest);
      await route.continue();
    },
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(hostsViewSchemaId)}`,
  );
  const hostGrouping = page.getByTestId(
    gridGroupingSelectTestId(hostsViewSchemaId),
  );
  await expect(hostGrouping).toHaveValue("");
  await hostGrouping.click();
  await hostGrouping.selectOption("host.host_state");
  await expect(hostGrouping).toHaveValue("host.host_state");
  await expect(
    page.getByTestId(
      workbookQueryEntryTestId(hostsViewSchemaId, "group", "host.host_state"),
    ),
  ).toBeVisible();
  await expect(
    page.locator(gridGroupRowsSelector(hostsViewSchemaId, "host.host_state")),
  ).not.toHaveCount(0);
  expect(requests.at(-1)?.group_by).toBe("host.host_state");
  await switchOrdinarySheet(page, timelineViewSchemaId);
  await expect(
    page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
  ).toHaveValue("");
  await expect(hostGrouping).toHaveCount(0);
});
