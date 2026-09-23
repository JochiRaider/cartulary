import {
  gridGroupingSelectTestId,
  gridSortHeaderTestId,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  notesViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";
import { createSavedView } from "./support/workbook/savedViews";

function gate() {
  let release = () => {};
  const ready = new Promise<void>((resolve) => {
    release = resolve;
  });
  return { ready, release };
}

test("Saved-view discovery publishes bounded pages and preserves addressed selection through delayed reads and retry", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("SVD"),
    "Saved-view bounded discovery",
  );
  const startup = await createSavedView(page, incident, {
    display_name: "Startup selected outside discovery page",
    scope: "private",
    view_schema_id: timelineViewSchemaId,
    query_json: { sort: [], filters: [] },
    layout_json: {},
  });
  for (let batch = 0; batch < 6; batch++) {
    await Promise.all(
      Array.from({ length: batch === 5 ? 2 : 10 }, (_, i) =>
        createSavedView(page, incident, {
          display_name: `Discovery ${batch * 10 + i} — a long analyst configuration label`,
          scope: "shared",
          view_schema_id: timelineViewSchemaId,
          query_json: {
            sort: [],
            filters: [],
            group_by: "timeline.capture_state",
          },
          layout_json: {
            layout_schema_id: "cartulary.layout.v2",
            frozen_through_field_key: null,
            column_order: requireViewContract(timelineViewSchemaId).fields.map(
              (field) => field.fieldKey,
            ),
            column_widths: [],
            hidden_field_keys: ["timeline.analyst_text"],
          },
        }),
      ),
    );
  }
  // Newer unrelated schema entries cannot consume the active schema's page.
  await Promise.all(
    Array.from({ length: 8 }, (_, i) =>
      createSavedView(page, incident, {
        display_name: `Notes ${i}`,
        scope: "private",
        view_schema_id: notesViewSchemaId,
        query_json: { sort: [], filters: [] },
        layout_json: {},
      }),
    ),
  );
  const first = gate();
  const later = gate();
  let firstPending = false;
  let laterPending = false;
  let continuationAttempts = 0;
  const requests: URL[] = [];
  await page.route(
    new RegExp(`/api/v1/incidents/${incident}/saved-views(?:\\?|$)`),
    async (route) => {
      if (route.request().method() !== "GET") {
        await route.continue();
        return;
      }
      const url = new URL(route.request().url());
      requests.push(url);
      expect(url.searchParams.get("view_schema_id")).toBe(timelineViewSchemaId);
      expect(url.searchParams.get("limit")).toBe("50");
      if (requests.length === 1) {
        const response = await route.fetch();
        firstPending = true;
        await first.ready;
        await route.fulfill({ response });
        return;
      }
      if (url.searchParams.has("cursor_token")) {
        continuationAttempts++;
        if (continuationAttempts === 1) {
          await route.fulfill({
            status: 500,
            contentType: "application/json",
            body: JSON.stringify({
              error: {
                code: "internal_error",
                status: 500,
                message: "Discovery temporarily unavailable",
                details: {},
                request_id: "svd-continuation",
                retryable: true,
              },
            }),
          });
          return;
        }
        if (continuationAttempts === 2) {
          const response = await route.fetch();
          laterPending = true;
          await later.ready;
          await route.fulfill({ response });
          return;
        }
      }
      await route.continue();
    },
  );
  await page.goto(
    `/?incident_id=${incident}&sheet_ref_kind=saved_view&sheet_ref_id=${startup.saved_view_id}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const trigger = page.getByTestId(
    savedViewSelectorTestId(timelineViewSchemaId),
  );
  await expect(trigger).toHaveAttribute(
    "data-selected-saved-view-id",
    startup.saved_view_id,
  );
  expect(requests).toHaveLength(0);
  const analystHeader = page.getByTestId(
    gridSortHeaderTestId(timelineViewSchemaId, "timeline.analyst_text"),
  );
  await expect(analystHeader).toBeVisible();
  await trigger.press("Enter");
  const browser = page.getByRole("dialog", {
    name: "Saved views",
    exact: true,
  });
  await expect.poll(() => firstPending).toBe(true);
  await expect(
    browser.getByRole("option", { name: "Unsaved view", exact: true }),
  ).toBeVisible();
  await expect(trigger).toContainText(startup.display_name);
  first.release();
  await expect(browser.getByRole("option")).toHaveCount(51);
  expect(requests).toHaveLength(1);
  await expect(
    page.getByTestId(
      savedViewOptionTestId(timelineViewSchemaId, startup.saved_view_id),
    ),
  ).toHaveCount(0);
  await browser.getByRole("button", { name: "Next", exact: true }).click();
  await expect(
    browser.getByRole("button", { name: "Retry page" }),
  ).toBeVisible();
  await expect(browser.getByRole("option")).toHaveCount(51);
  await expect(trigger).toHaveAttribute(
    "data-selected-saved-view-id",
    startup.saved_view_id,
  );
  await browser.getByRole("button", { name: "Retry page" }).click();
  await expect.poll(() => laterPending).toBe(true);
  await expect(browser.getByRole("option")).toHaveCount(51);
  const base = browser.getByRole("option", {
    name: "Unsaved view",
    exact: true,
  });
  await base.focus();
  await base.press("End");
  await expect(browser.getByRole("option").last()).toBeFocused();
  await expect(trigger).toHaveAttribute(
    "data-selected-saved-view-id",
    startup.saved_view_id,
  );
  later.release();
  await expect(browser.getByRole("option")).toHaveCount(4);
  expect(requests[1]?.searchParams.get("cursor_token")).toBe(
    requests[2]?.searchParams.get("cursor_token"),
  );
  expect(requests).toHaveLength(3);
  await browser.getByRole("button", { name: "Previous", exact: true }).click();
  await expect(browser.getByRole("option")).toHaveCount(51);
  await browser.getByRole("button", { name: "Next", exact: true }).click();
  await expect(browser.getByRole("option")).toHaveCount(4);
  const candidate = browser
    .getByRole("option")
    .filter({ hasText: "Discovery" })
    .first();
  const id = await candidate.getAttribute("data-saved-view-id");
  if (!id) throw new Error("Candidate is missing its resource identity");
  const query = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().includes(`/views/${timelineViewSchemaId}/query`),
  );
  await candidate.focus();
  await candidate.press("Enter");
  expect((await query).postDataJSON().group_by).toBe("timeline.capture_state");
  await expect(trigger).toHaveAttribute("data-selected-saved-view-id", id);
  await expect(
    page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
  ).toHaveValue("timeline.capture_state");
  await expect(analystHeader).toHaveCount(0);
  await expect(browser).toHaveCount(0);
  await expect(trigger).toBeFocused();
  await trigger.press("Enter");
  await expect(browser).toBeVisible();
  await browser.press("Escape");
  await expect(trigger).toBeFocused();
  await expect(trigger).toHaveAttribute("data-selected-saved-view-id", id);
});
