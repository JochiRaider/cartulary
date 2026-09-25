import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  relationshipItemsTestId,
  saveStateTestId,
  surfaceTabTestId,
  timelineCollectionInputTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
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
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { retainAutosaveFeedbackCase } from "./support/workbook/autosaveFeedback";
import { showTimelineCollectionColumns } from "./support/workbook/collections";
import { createViewRow, waitForViewRow } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { observeSaveEvents, saveEvents } from "./support/workbook/saveStatus";

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

// Collection commit counts belong to controlled React tests. This browser row
// observes public operations and editor continuity across their response boundary.
test("Timeline autosave feedback production characterization", async ({
  page,
}, testInfo) => {
  test.setTimeout(300_000);
  await page.setViewportSize({ width: 1440, height: 900 });
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
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  for (const loaded of [1, 100, 200, 300]) {
    for (const inspectorOpen of [false, true]) {
      const observation = {
        loaded,
        inspectorOpen,
        phase: "setup",
        recordId: row.record_id,
        fieldKey: "timeline.tags",
        surface: "grid",
        noteWrites: 0,
        collectionWrites: 0,
        scalarWrites: 0,
        editor: null as null | {
          connected: boolean;
          focused: boolean;
          draftMatches: boolean;
          start: number | null;
          end: number | null;
          direction: string | null;
        },
        gates: {
          notes: { started: false, released: false, completed: false },
          collection: { started: false, released: false, completed: false },
          scalar: { started: false, released: false, completed: false },
        },
        status: "",
        events: [] as Awaited<ReturnType<typeof saveEvents>>,
      };
      let refreshRequestEvidence = () => {};
      const releaseGate =
        (name: keyof typeof observation.gates, release: () => void) => () => {
          observation.gates[name].released = true;
          release();
        };
      const releases: (() => void)[] = [];
      const disposers: (() => Promise<void>)[] = [];
      const phase = <T>(name: string, action: () => Promise<T>) => {
        observation.phase = name;
        return test.step(`${loaded}/${inspectorOpen}: ${name}`, action, {
          timeout: 25_000,
        });
      };
      await retainAutosaveFeedbackCase(
        observation,
        async () => {
          if (loaded === 100 && !inspectorOpen)
            await createTimelineFillers(page, incidentId, "taf-window", 299, {
              occurredAtStart: "2026-04-02T00:00:00Z",
            });
          const patches = await installPatchController(page);
          disposers.push(patches.dispose);
          // An ordinary Notes create remains in flight while Timeline is active.
          // The test never discovers or invokes an application runtime handle.
          const notePath = `/api/v1/incidents/${incidentId}/views/${notesViewSchemaId}/rows`;
          const note = await holdBrowserRequest(page, {
            method: "POST",
            path: notePath,
          });
          const releaseNote = releaseGate("notes", note.release);
          releases.push(releaseNote);
          let collectionStart: number | null = null;
          let scalarStart: number | null = null;
          refreshRequestEvidence = () => {
            observation.noteWrites = note.hitCount();
            observation.gates.notes.started = note.hitCount() > 0;
            if (collectionStart !== null)
              observation.collectionWrites =
                (scalarStart ?? patches.calls.length) - collectionStart;
            if (scalarStart !== null)
              observation.scalarWrites = patches.calls.length - scalarStart;
          };
          disposers.push(note.dispose);
          await phase("hold-notes-operation", async () => {
            await page.getByTestId(surfaceTabTestId(notesViewSchemaId)).click();
            const title = page.getByTestId(
              genericCreateFieldTestId("note.title"),
            );
            if (!(await title.isVisible()))
              await page
                .getByTestId(workbookInspectorToggleTestId(notesViewSchemaId))
                .click();
            await title.fill(`Status ${loaded} ${inspectorOpen}`);
            await page
              .getByTestId(genericCreateSubmitTestId(notesViewSchemaId))
              .click();
            await expect.poll(note.hitCount, { timeout: 25_000 }).toBe(1);
            await expect(page.getByTestId(saveStateTestId())).toHaveText(
              "Syncing",
            );
          });
          const input = page.getByTestId(
            timelineCollectionInputTestId(
              row.record_id,
              "timeline.tags",
              "grid",
            ),
          );
          const rawDraft = `retained raw Ω draft ${loaded} ${inspectorOpen}`;
          await phase("editor-ready", async () => {
            await page
              .getByTestId(surfaceTabTestId(timelineViewSchemaId))
              .click();
            await expect(
              page.getByTestId(gridShellTestId(timelineViewSchemaId)),
            ).toBeVisible();
            // Each navigation starts a query window. Observe accepted pages rather
            // than guessing when its effects or the inspector have settled.
            const firstPage = loaded === 1 ? 1 : 100;
            await expect(browsing).toContainText(`${firstPage} records loaded`);
            for (let count = 100; count < loaded; count += 100) {
              await browsing
                .getByRole("button", { name: "Load more", exact: true })
                .click();
              await expect(browsing).toContainText(
                `${count + 100} records loaded`,
              );
            }
            const close = page.getByTestId(
              workbookInspectorCloseButtonTestId(timelineViewSchemaId),
            );
            if (await close.isVisible()) await close.click();
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
                  relationshipItemsTestId(
                    row.record_id,
                    "timeline.tags",
                    "grid",
                  ),
                ),
              });
            await cell
              .locator("xpath=ancestor::*[@role='gridcell'][1]")
              .click({ position: { x: 1, y: 1 } });
            if (inspectorOpen)
              await page
                .getByTestId(
                  workbookInspectorToggleTestId(timelineViewSchemaId),
                )
                .click();
            await expect(close).toHaveCount(inspectorOpen ? 1 : 0);
            await cell
              .getByRole("button", { name: "Add tags token", exact: true })
              .click();
            await input.fill(rawDraft);
            await input.evaluate((element: HTMLInputElement) =>
              element.setSelectionRange(3, 8, "backward"),
            );
            await expect(input).toBeFocused();
            await expect(input).toHaveValue(rawDraft);
          });
          const original = await input.elementHandle();
          if (!original)
            throw new Error("The authored native editor is absent");
          disposers.push(() => original.dispose());
          const readEditor = () =>
            original.evaluate(
              (element: HTMLInputElement, expected: string) => ({
                connected: element.isConnected,
                focused: document.activeElement === element,
                draftMatches: element.value === expected,
                start: element.selectionStart,
                end: element.selectionEnd,
                direction: element.selectionDirection,
              }),
              rawDraft,
            );
          const expectedEditor = {
            connected: true,
            focused: true,
            draftMatches: true,
            start: 3,
            end: 8,
            direction: "backward",
          };
          observation.editor = await readEditor();
          expect(observation.editor).toEqual(expectedEditor);
          await phase("notes-response-and-continuity", async () => {
            const response = page.waitForResponse(
              (response) =>
                new URL(response.url()).pathname === notePath &&
                response.request().method() === "POST",
              { timeout: 25_000 },
            );
            releaseNote();
            const settled = await response;
            observation.gates.notes.completed = true;
            expect(settled.ok()).toBe(true);
            await expect(page.getByTestId(saveStateTestId())).toHaveText(
              "Saved",
            );
            observation.noteWrites = note.hitCount();
            expect(observation.noteWrites).toBe(1);
            await expect
              .poll(async () => {
                observation.editor = await readEditor();
                return observation.editor;
              })
              .toEqual(expectedEditor);
          });
          const beforeCollection = patches.calls.length;
          collectionStart = beforeCollection;
          const collection = patches.holdNextPatch({ recordId: row.record_id });
          const releaseCollection = releaseGate(
            "collection",
            collection.release,
          );
          releases.push(releaseCollection);
          await phase("collection-held", async () => {
            await input.press(inspectorOpen ? "Tab" : "Enter");
            await collection.waitForHit;
            observation.gates.collection.started = true;
            await expect(page.getByTestId(saveStateTestId())).toHaveText(
              "Syncing",
            );
            await expect(input).toHaveValue(rawDraft);
          });
          await phase("collection-response", async () => {
            releaseCollection();
            await collection.waitForCompletion;
            observation.gates.collection.completed = true;
            await expect(page.getByTestId(saveStateTestId())).toHaveText(
              "Saved",
            );
            observation.collectionWrites =
              patches.calls.length - beforeCollection;
            expect(observation.collectionWrites).toBe(1);
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
          });
          const summary = `Autosave ${loaded} ${inspectorOpen}`;
          const beforeScalar = patches.calls.length;
          scalarStart = beforeScalar;
          const scalar = patches.holdNextPatch({ recordId: row.record_id });
          const releaseScalar = releaseGate("scalar", scalar.release);
          releases.push(releaseScalar);
          await phase("scalar-held", async () => {
            await editTimelineSummary(page, row.record_id, summary, {
              outcome: "queued",
              commit: "Tab",
            });
            await scalar.waitForHit;
            observation.gates.scalar.started = true;
            await expect(page.getByTestId(saveStateTestId())).toHaveText(
              "Syncing",
            );
            await expect(
              page.getByRole("complementary", {
                name: "Workbook notices",
                exact: true,
              }),
            ).toHaveCount(0);
            if (loaded === 100 && !inspectorOpen)
              await testInfo.attach("held-save-full-workbook", {
                body: await page.screenshot(),
                contentType: "image/png",
              });
          });
          await phase("scalar-responses", async () => {
            releaseScalar();
            await scalar.waitForCompletion;
            observation.gates.scalar.completed = true;
            await expect(page.getByTestId(saveStateTestId())).toHaveText(
              "Saved",
            );
            await expectServerSummaries(page, incidentId, {
              [row.record_id]: summary,
            });
            await editTimelineSummary(page, row.record_id, `${summary} rapid`, {
              outcome: "accepted",
            });
            await expectServerSummaries(page, incidentId, {
              [row.record_id]: `${summary} rapid`,
            });
            observation.scalarWrites = patches.calls.length - beforeScalar;
            expect(observation.scalarWrites).toBe(2);
            expect(
              successfulPatchCalls(patches.calls.slice(beforeScalar)),
            ).toHaveLength(2);
          });
          observation.phase = "complete";
        },
        async (active) => {
          refreshRequestEvidence();
          // The case exists before assertions. Attach before releasing gates and
          // omit payloads, credentials and authored text from retained metadata.
          active.status =
            (await page
              .getByTestId(saveStateTestId())
              .textContent({ timeout: 1_000 })
              .catch(() => null)) ?? "unavailable";
          active.events = (await saveEvents(page).catch(() => [])).slice(-16);
          await testInfo.attach(
            `autosave-feedback-${loaded}-${inspectorOpen}-${active.phase}`,
            {
              body: JSON.stringify({
                kind: "semantic-autosave-continuity",
                ...active,
              }),
              contentType: "application/json",
            },
          );
        },
        async () => {
          for (const release of releases) release();
          const results = await Promise.allSettled(
            disposers.map((dispose) => dispose()),
          );
          const failure = results.find(
            (result) => result.status === "rejected",
          );
          if (failure?.status === "rejected") throw failure.reason;
        },
      );
    }
  }
});
