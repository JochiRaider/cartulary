import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  gridShellTestId,
  relationshipItemsTestId,
  saveStateTestId,
  timelineCollectionInputTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import {
  editTimelineSummary,
  expectServerSummaries,
  installPatchController,
  successfulPatchCalls,
} from "./support/collaboration/replay";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createTimelineFillers } from "./support/timeline/fixtures";
import { showTimelineCollectionColumns } from "./support/workbook/collections";
import { createViewRow, waitForViewRow } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { observeSaveEvents, saveEvents } from "./support/workbook/saveStatus";

// Browser-only observation port; do not pull the application's CSS graph into
// the separate E2E TypeScript project through a production-runtime type import.
type DiagnosticRuntime = {
  beginExplicitMutation(): () => void;
  getSnapshot(): unknown;
  notifyPendingChanged(): void;
};

type DiagnosticWindow = Window & {
  taf: {
    runtime: DiagnosticRuntime | null;
    counts: Record<string, number>;
  };
};

test("Timeline failed refresh stays stale beside terminal save recovery", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TAFQUERY"),
    "Timeline independent query and save feedback",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("taf-query-row"),
    "timeline.activity_synopsis_text": "Authoritative original",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  const patches = await installPatchController(page);
  const queryRoute = `**/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`;
  let failedReads = 0;
  try {
    patches.failNextPatch(409, "client_txn_conflict", {
      recordId: row.record_id,
    });
    await editTimelineSummary(page, row.record_id, "Retained rejected draft", {
      outcome: "queued",
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await page.route(queryRoute, async (route) => {
      failedReads++;
      await route.abort("failed");
    });
    const refresh = page
      .getByRole("group", { name: "Workbook browsing" })
      .getByRole("button", { name: "Refresh", exact: true });
    await refresh.click();
    await expect(
      page.locator('[data-grid-data-state="stale_error"]'),
    ).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    expect(failedReads).toBe(1);
    expect(patches.calls).toHaveLength(1);
    await openRecoveryItem(page, /^Queued edit recovery ·/);
    await page.getByTestId(workbookEditRecoveryDiscardButtonTestId()).click();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.locator('[data-grid-data-state="stale_error"]'),
    ).toBeVisible();
    expect(patches.calls).toHaveLength(1);
    await page.unroute(queryRoute);
    await refresh.click();
    await expect(
      page.locator('[data-grid-data-state="stale_error"]'),
    ).toHaveCount(0);
    await expectServerSummaries(page, incidentId, {
      [row.record_id]: "Authoritative original",
    });
    expect(patches.calls).toHaveLength(1);
  } finally {
    await page.unroute(queryRoute);
    await patches.dispose();
  }
});

test("Timeline autosave feedback production characterization", async ({
  page,
}, testInfo) => {
  test.setTimeout(300_000);
  await page.setViewportSize({ width: 1440, height: 900 });
  // Diagnostic commit observations only. AC-043 runs without this hook.
  await page.addInitScript(() => {
    type Fiber = {
      type?: unknown;
      flags: number;
      child?: Fiber;
      sibling?: Fiber;
      memoizedProps?: Record<string, unknown> & {
        mutationRuntime?: DiagnosticRuntime;
        binding?: { kind?: string };
      };
      alternate?: Fiber;
    };
    const observed = window as unknown as DiagnosticWindow & {
      __REACT_DEVTOOLS_GLOBAL_HOOK__: unknown;
    };
    observed.taf = { runtime: null, counts: {} };
    let previous = new WeakSet<object>();
    observed.__REACT_DEVTOOLS_GLOBAL_HOOK__ = {
      supportsFiber: true,
      inject: () => 1,
      onCommitFiberUnmount: () => {},
      onCommitFiberRoot: (_renderer: number, root: { current: Fiber }) => {
        const counts = observed.taf.counts;
        const increment = (key: string) => {
          counts[key] = (counts[key] ?? 0) + 1;
        };
        increment("commits");
        const current = new WeakSet<object>();
        const visit = (fiber: Fiber | undefined) => {
          if (!fiber) return;
          current.add(fiber);
          const props = fiber.memoizedProps;
          if (props?.mutationRuntime?.beginExplicitMutation)
            observed.taf.runtime = props.mutationRuntime;
          if (!previous.has(fiber)) {
            if (
              typeof fiber.type === "function" &&
              (fiber.flags & 1) !== 0 &&
              props?.binding?.kind === "collection"
            )
              increment("collectionRenders");
            for (const key of ["columns", "rows"])
              if (
                Array.isArray(props?.[key]) &&
                fiber.alternate &&
                props?.[key] !== fiber.alternate.memoizedProps?.[key]
              )
                increment(`${key}Replacements`);
          }
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
    uniqueIncidentKey("TAF"),
    "Timeline autosave feedback",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("taf-row"),
    "timeline.activity_utc_text": "2026-04-01T00:00:00Z",
    "timeline.activity_synopsis_text": "Autosave baseline",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await observeSaveEvents(page);
  await showTimelineCollectionColumns(page, ["Tags"]);
  const settle = () =>
    page.evaluate(
      () =>
        new Promise<void>((resolve) =>
          requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
        ),
    );
  const counts = () =>
    page.evaluate(() => ({
      ...(window as unknown as DiagnosticWindow).taf.counts,
    }));
  const delta = (
    before: Record<string, number>,
    after: Record<string, number>,
  ) =>
    Object.fromEntries(
      [...new Set([...Object.keys(before), ...Object.keys(after)])].map(
        (key) => [key, (after[key] ?? 0) - (before[key] ?? 0)],
      ),
    );
  const observations: unknown[] = [];
  const patches = await installPatchController(page);
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  try {
    for (const loaded of [1, 100, 200, 300]) {
      if (loaded === 100) {
        await createTimelineFillers(page, incidentId, "taf-window", 299, {
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
        const close = page.getByTestId(
          workbookInspectorCloseButtonTestId(timelineViewSchemaId),
        );
        if (await close.count()) await close.click();
        await scrollGridTargetIntoView({
          page,
          surface: timelineViewSchemaId,
          targetTestId: relationshipItemsTestId(
            row.record_id,
            "timeline.tags",
            "grid",
          ),
        });
        const cell = page
          .getByRole("group", { name: "Tags collection cell", exact: true })
          .filter({
            has: page.getByTestId(
              relationshipItemsTestId(row.record_id, "timeline.tags", "grid"),
            ),
          });
        await cell
          .locator("xpath=ancestor::*[@role='gridcell'][1]")
          .click({ position: { x: 1, y: 1 } });
        if (inspectorOpen)
          await page
            .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
            .click();
        await cell
          .getByRole("button", { name: "Add tags token", exact: true })
          .click();
        const input = page.getByTestId(
          timelineCollectionInputTestId(row.record_id, "timeline.tags", "grid"),
        );
        const rawDraft = `retained raw Ω draft ${loaded} ${inspectorOpen}`;
        await input.fill(rawDraft);
        await input.evaluate((element: HTMLInputElement) =>
          element.setSelectionRange(3, 8, "backward"),
        );
        const original = await input.elementHandle();
        await settle();
        const before = await counts();
        const unchangedSnapshot = await page.evaluate(() => {
          const runtime = (window as unknown as DiagnosticWindow).taf.runtime;
          if (!runtime) throw new Error("Production runtime was not observed");
          const previous = runtime.getSnapshot();
          runtime.notifyPendingChanged();
          return previous === runtime.getSnapshot();
        });
        await settle();
        const unchangedWork = delta(before, await counts());
        expect(unchangedSnapshot).toBe(true);
        expect(unchangedWork.commits ?? 0).toBe(0);
        const beforeStatus = await counts();
        await page.evaluate(async () => {
          const runtime = (window as unknown as DiagnosticWindow).taf.runtime;
          if (!runtime) throw new Error("Production runtime was not observed");
          const finishFirst = runtime.beginExplicitMutation();
          const finishSecond = runtime.beginExplicitMutation();
          await new Promise(requestAnimationFrame);
          finishFirst();
          await new Promise(requestAnimationFrame);
          finishSecond();
        });
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
        await settle();
        const statusWork = delta(beforeStatus, await counts());
        for (const work of [
          "columnsReplacements",
          "rowsReplacements",
          "collectionRenders",
        ])
          expect(
            statusWork[work] ?? 0,
            `${work} during save status updates`,
          ).toBe(0);
        expect(
          await original?.evaluate((element: HTMLInputElement) => ({
            connected: element.isConnected,
            focused: document.activeElement === element,
            text: element.value,
            start: element.selectionStart,
            end: element.selectionEnd,
            direction: element.selectionDirection,
          })),
        ).toEqual({
          connected: true,
          focused: true,
          text: rawDraft,
          start: 3,
          end: 8,
          direction: "backward",
        });
        const beforeCollection = patches.calls.length;
        const beforeRowChange = await counts();
        const collection = patches.holdNextPatch({ recordId: row.record_id });
        try {
          await input.press(inspectorOpen ? "Tab" : "Enter");
          await collection.waitForHit;
          await expect(page.getByTestId(saveStateTestId())).toHaveText(
            "Syncing",
          );
          await expect(input).toHaveValue(rawDraft);
        } finally {
          collection.release();
          await collection.waitForCompletion;
        }
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
        expect(
          successfulPatchCalls(patches.calls.slice(beforeCollection)),
        ).toHaveLength(1);
        const saved = await waitForViewRow(
          page,
          incidentId,
          timelineViewSchemaId,
          row.record_id,
        );
        expect(JSON.stringify(saved.cells["timeline.tags"])).toContain(
          rawDraft,
        );
        const summary = `Autosave ${loaded} ${inspectorOpen}`;
        const firstCall = patches.calls.length;
        const held = patches.holdNextPatch({ recordId: row.record_id });
        let notice: unknown;
        try {
          await editTimelineSummary(page, row.record_id, summary, {
            outcome: "queued",
            commit: "Tab",
          });
          await held.waitForHit;
          await expect(page.getByTestId(saveStateTestId())).toHaveText(
            "Syncing",
          );
          const stack = page.getByRole("complementary", {
            name: "Workbook notices",
            exact: true,
          });
          await expect(stack).toHaveCount(0);
          const rect = (await stack.count()) ? await stack.boundingBox() : null;
          const grid = await page
            .getByTestId(gridShellTestId(timelineViewSchemaId))
            .boundingBox();
          notice = {
            text: (await stack.count()) ? await stack.innerText() : null,
            rect,
            overlapsGrid: !!(
              rect &&
              grid &&
              rect.x < grid.x + grid.width &&
              rect.x + rect.width > grid.x &&
              rect.y < grid.y + grid.height &&
              rect.y + rect.height > grid.y
            ),
            status: await page
              .getByTestId(workbookShellSlotTestId("status-strip"))
              .innerText(),
          };
          if (loaded === 100 && !inspectorOpen)
            await testInfo.attach("held-save-full-workbook", {
              body: await page.screenshot(),
              contentType: "image/png",
            });
        } finally {
          held.release();
          await held.waitForCompletion;
        }
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
        await expectServerSummaries(page, incidentId, {
          [row.record_id]: summary,
        });
        await editTimelineSummary(page, row.record_id, `${summary} rapid`, {
          outcome: "accepted",
        });
        await expectServerSummaries(page, incidentId, {
          [row.record_id]: `${summary} rapid`,
        });
        expect(
          successfulPatchCalls(patches.calls.slice(firstCall)),
        ).toHaveLength(2);
        observations.push({
          loaded,
          inspectorOpen,
          unchangedSnapshot,
          unchangedWork,
          statusWork,
          notice,
          scalarWrites: 2,
          collectionWrites: 1,
          rowChangeWork: delta(beforeRowChange, await counts()),
          events: await saveEvents(page),
        });
      }
    }
  } finally {
    await patches.dispose();
    await testInfo.attach("autosave-feedback-observations", {
      body: JSON.stringify(
        {
          kind: "diagnostic-production-characterization-not-timing",
          observations,
        },
        null,
        2,
      ),
      contentType: "application/json",
    });
  }
});
