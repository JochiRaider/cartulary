import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  gridRowTestId,
  gridRowVersionAttribute,
  gridScrollportSelector,
  relationshipItemsTestId,
  rowCellTestId,
  saveStateTestId,
  timelineCollectionInputTestId,
  timelineMutationSubstrateReadyTestId,
  workbookRowContextMenuTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  installDeferredContinuityProbe,
  pendingContinuity,
} from "./support/timeline/deferredContinuity";
import { createTimelineFillers } from "./support/timeline/fixtures";
import { showTimelineCollectionColumns } from "./support/workbook/collections";
import { queryViewRows } from "./support/workbook/query";

const synopsis = "timeline.activity_synopsis_text";
const tags = "timeline.tags";

test("Timeline pending restoration yields to wheel native input external focus and a newer menu", async ({
  page,
}, info) => {
  test.setTimeout(180_000);
  page.setDefaultTimeout(10_000);
  await installDeferredContinuityProbe(page);
  await page.setViewportSize({ width: 1440, height: 720 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("CONTINUITY"),
    "Deferred Timeline continuity",
  );
  await createTimelineFillers(page, incidentId, "Continuity fact", 35);
  const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
  const source = rows[5];
  const other = rows[6];
  if (!source || !other) throw new Error("Missing continuity source rows");
  const observations = [];
  for (const intent of [
    "wheel",
    "input",
    "focus",
    "menu",
    "range",
    "uninterrupted",
  ] as const) {
    await page.goto(`/?incident_id=${incidentId}`);
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect
      .poll(() => page.evaluate(() => window.timelineContinuityProbe.handles))
      .toBeGreaterThan(0);
    const bulk = page.getByRole("checkbox", {
      name: `Select record ${other.record_id}`,
      exact: true,
    });
    await bulk.check();
    await showTimelineCollectionColumns(page, ["Tags"]);
    const input = page.getByTestId(
      timelineCollectionInputTestId(
        source.record_id,
        tags,
        intent === "input" ? "inspector" : "grid",
      ),
    );
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(source.record_id, tags, "grid"),
    });
    const collection = page
      .getByTestId(relationshipItemsTestId(source.record_id, tags, "grid"))
      .locator("xpath=ancestor::*[@role='gridcell'][1]");
    if (intent === "input") {
      await collection
        .getByRole("button", {
          name: "Inspect tag: pending-wheel",
          exact: true,
        })
        .click();
    } else {
      await collection
        .getByRole("button", { name: "Add tags token", exact: true })
        .click();
    }
    await input.fill(`pending-${intent}`);
    // DOM availability is gated only by this test. The production adapter keeps
    // its accepted semantic membership and its real cancellable focus request.
    await page.evaluate(
      ({ id, field }) => window.timelineContinuityProbe.hold(id, field),
      { id: source.record_id, field: synopsis },
    );
    const writes: string[] = [];
    const count = (request: import("@playwright/test").Request) => {
      if (
        request.method() === "PATCH" &&
        request.url().endsWith(`/records/${source.record_id}`)
      )
        writes.push(request.postData() ?? "");
    };
    page.on("request", count);
    try {
      // Native blur commits through the existing collection owner; no later
      // destination is introduced until the restoration request is pending.
      await input.evaluate((element: HTMLInputElement) => element.blur());
      await expect
        .poll(async () =>
          (await pendingContinuity(page, source.record_id)).some(
            (item) => item.result === null && !item.aborted,
          ),
        )
        .toBe(true);
      expect(
        await page.evaluate(
          (id) =>
            document
              .querySelector(`[data-testid="${id}"]`)
              ?.closest("[role='gridcell']")?.isConnected,
          rowCellTestId(source.record_id, synopsis),
        ),
      ).toBe(false);
      const scrollport = page.locator(gridScrollportSelector());
      if (intent === "wheel") {
        await scrollport.hover();
        await page.mouse.wheel(0, 100);
        await page.evaluate(
          () =>
            new Promise<void>((resolve) =>
              requestAnimationFrame(() =>
                requestAnimationFrame(() => resolve()),
              ),
            ),
        );
      } else if (intent === "input") {
        // Native input without keydown is deliberately independent of adapter
        // keyboard cancellation (e.g. IME, speech or an input-method update).
        await input.evaluate((element: HTMLInputElement) => {
          Object.getOwnPropertyDescriptor(
            HTMLInputElement.prototype,
            "value",
          )?.set?.call(element, "Newer raw Ω 東京");
          element.setSelectionRange(2, 7, "backward");
          element.dispatchEvent(
            new InputEvent("input", {
              bubbles: true,
              inputType: "insertText",
              data: "Ω",
            }),
          );
        });
      } else if (intent === "focus") {
        await page.getByRole("button", { name: "Hosts", exact: true }).focus();
      } else if (intent === "range") {
        await page
          .getByTestId(rowCellTestId(other.record_id, synopsis))
          .click();
        await page.keyboard.press("Escape");
        await page.keyboard.press("Shift+ArrowDown");
        await expect(
          page.locator(
            `${gridScrollportSelector()} [role="gridcell"][aria-selected="true"]`,
          ),
        ).toHaveCount(2);
      } else if (intent === "menu") {
        await page
          .getByTestId(rowCellTestId(other.record_id, synopsis))
          .click({ button: "right" });
        await expect(
          page.getByTestId(
            workbookRowContextMenuTestId(timelineViewSchemaId, other.record_id),
          ),
        ).toBeVisible();
      }
      if (intent !== "uninterrupted")
        await expect
          .poll(async () =>
            (await pendingContinuity(page, source.record_id)).some(
              (item) => item.aborted && item.result === "cancelled",
            ),
          )
          .toBe(true);
      const selection = () =>
        page
          .locator(
            `${gridScrollportSelector()} [role="gridcell"][aria-selected="true"]`,
          )
          .evaluateAll((cells) =>
            cells.map((cell) => ({
              row: cell.closest<HTMLElement>("[data-grid-record-id]")?.dataset
                .gridRecordId,
              field: cell.querySelector<HTMLElement>("[data-grid-field-key]")
                ?.dataset.gridFieldKey,
            })),
          );
      const selected = await selection();
      const before = await page.evaluate((selector) => {
        const grid = document.querySelector(selector);
        const active = document.activeElement;
        window.timelineContinuityProbe.destination = active;
        return {
          top: grid?.scrollTop,
          left: grid?.scrollLeft,
          focus: active?.getAttribute("data-testid") ?? active?.tagName,
          text: active instanceof HTMLInputElement ? active.value : null,
        };
      }, gridScrollportSelector());
      await page.evaluate(() => window.timelineContinuityProbe.release());
      // A real accepted query update releases normal adapter layout/registration
      // notifications, without another user focus or scrolling gesture.
      const current = (
        await queryViewRows(page, incidentId, timelineViewSchemaId)
      ).find((row) => row.record_id === source.record_id);
      if (!current) throw new Error("Missing refresh source");
      const refreshed = await page.request.patch(
        `${apiBase}/api/v1/records/${current.record_id}`,
        {
          headers: await csrfHeaders(page),
          data: {
            view_schema_id: timelineViewSchemaId,
            client_txn_id: uniqueTxn("continuity-refresh"),
            base_row_version: current.row_version,
            changes: [
              {
                field_key: "timeline.data_source_text",
                value: `Readiness update ${intent}`,
              },
            ],
          },
        },
      );
      expect(refreshed.ok()).toBe(true);
      await expect(
        page.getByTestId(gridRowTestId(timelineViewSchemaId, source.record_id)),
      ).toHaveAttribute(
        gridRowVersionAttribute,
        String(current.row_version + 1),
      );
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(
        page.getByTestId(rowCellTestId(source.record_id, synopsis)),
      ).toHaveCount(1);
      if (intent === "uninterrupted") {
        const cell = page
          .getByTestId(rowCellTestId(source.record_id, synopsis))
          .locator("xpath=ancestor::*[@role='gridcell'][1]");
        await expect(cell).toBeFocused();
      } else {
        await expect
          .poll(() =>
            page.evaluate(
              () =>
                document.activeElement ===
                window.timelineContinuityProbe.destination,
            ),
          )
          .toBe(true);
        await expect
          .poll(() =>
            scrollport.evaluate((node) => ({
              top: node.scrollTop,
              left: node.scrollLeft,
            })),
          )
          .toEqual({ top: before.top, left: before.left });
        if (intent === "menu")
          await expect(
            page.getByTestId(
              workbookRowContextMenuTestId(
                timelineViewSchemaId,
                other.record_id,
              ),
            ),
          ).toBeVisible();
        if (intent === "input") {
          await expect(input).toHaveValue("Newer raw Ω 東京");
          expect(
            await input.evaluate((node: HTMLInputElement) => [
              node.selectionStart,
              node.selectionEnd,
              node.selectionDirection,
            ]),
          ).toEqual([2, 7, "backward"]);
        }
      }
      await expect(bulk).toBeChecked();
      expect(await selection()).toEqual(selected);
      expect(writes).toHaveLength(1);
      observations.push({
        intent,
        observedRowVersion: current.row_version + 1,
        selected,
        before,
        requests: await pendingContinuity(page, source.record_id),
        writes,
      });
    } catch (error) {
      await info.attach(`continuity-debug-${intent}`, {
        body: JSON.stringify(
          await page.evaluate(() => ({
            observed: window.timelineContinuityProbe.observed.map(
              ({ target, result, signal }) => ({
                target,
                result,
                aborted: signal?.aborted,
              }),
            ),
            focus: document.activeElement?.outerHTML,
            text: document.body.innerText,
          })),
          null,
          2,
        ),
        contentType: "application/json",
      });
      throw error;
    } finally {
      page.off("request", count);
      await page.evaluate(() => window.timelineContinuityProbe.release());
    }
  }
  await info.attach("deferred-continuity-observations", {
    body: JSON.stringify(observations, null, 2),
    contentType: "application/json",
  });
});
