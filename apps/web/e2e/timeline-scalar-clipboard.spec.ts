import {
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  draftCellTestId,
  gridFieldCellSelector,
  gridRowTestId,
  gridShellTestId,
  rowCellTestId,
  saveStateTestId,
  surfaceTabTestId,
  timelineScalarEditorTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import {
  expectServerTimelineCells,
  installPatchController,
} from "./support/collaboration/replay";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { createViewRow, queryViewRows } from "./support/workbook/query";

const synopsis = "timeline.activity_synopsis_text";
const source = "timeline.data_source_text";
const raw = "timeline.raw_activity_text";
const originalText = "alpha beta gamma";

async function seed(page: Page) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("SCALAR-CLIPBOARD"),
    "Scalar clipboard production evidence",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("scalar-clipboard"),
    [synopsis]: originalText,
    [source]: originalText,
    [raw]: originalText,
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  return { incidentId, recordId: row.record_id };
}

async function activate(
  page: Page,
  recordId: string,
  fieldKey: string,
  surface: "grid" | "inspector",
) {
  await scrollGridCellIntoView({
    page,
    surface: timelineViewSchemaId,
    recordId,
    cellKey: fieldKey,
  });
  await page.getByTestId(rowCellTestId(recordId, fieldKey)).click();
  if (surface === "inspector")
    await page
      .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
      .click();
  const input = page.getByTestId(
    timelineScalarEditorTestId({ recordId, fieldKey, surface }),
  );
  await expect(input).toBeVisible();
  await input.focus();
  return input;
}

async function selection(
  input: Locator,
  start: number,
  end = start,
  direction: "forward" | "backward" = "forward",
) {
  await input.evaluate(
    (element: HTMLInputElement | HTMLTextAreaElement, range) =>
      element.setSelectionRange(range.start, range.end, range.direction),
    { start, end, direction },
  );
}

async function paste(page: Page, text: string) {
  await page.evaluate((value) => navigator.clipboard.writeText(value), text);
  await page.keyboard.press("Control+v");
}

async function state(input: Locator) {
  return input.evaluate((element: HTMLInputElement | HTMLTextAreaElement) => ({
    value: element.value,
    start: element.selectionStart,
    end: element.selectionEnd,
    direction: element.selectionDirection,
    focused: document.activeElement === element,
    connected: element.isConnected,
    scrollTop: element.scrollTop,
  }));
}

test("Timeline scalar clipboard production characterization", async ({
  page,
}, testInfo) => {
  test.setTimeout(240_000);
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.addInitScript(() => {
    type Fiber = {
      type?: unknown;
      flags: number;
      child?: Fiber;
      sibling?: Fiber;
      memoizedProps?: Record<string, unknown>;
      alternate?: Fiber;
    };
    const probe = {
      commits: 0,
      renders: 0,
      rows: 0,
      columns: 0,
      events: [] as unknown[],
    };
    const observed = window as unknown as {
      scalarClipboardProbe: typeof probe;
      __REACT_DEVTOOLS_GLOBAL_HOOK__: unknown;
    };
    observed.scalarClipboardProbe = probe;
    let previous = new WeakSet<object>();
    observed.__REACT_DEVTOOLS_GLOBAL_HOOK__ = {
      supportsFiber: true,
      inject: () => 1,
      onCommitFiberUnmount: () => {},
      onCommitFiberRoot: (_renderer: number, root: { current: Fiber }) => {
        probe.commits++;
        const current = new WeakSet<object>();
        const visit = (fiber: Fiber | undefined) => {
          if (!fiber) return;
          current.add(fiber);
          if (!previous.has(fiber)) {
            if (typeof fiber.type === "function" && (fiber.flags & 1) !== 0)
              probe.renders++;
            for (const key of ["rows", "columns"] as const)
              if (
                Array.isArray(fiber.memoizedProps?.[key]) &&
                fiber.alternate &&
                fiber.memoizedProps?.[key] !==
                  fiber.alternate.memoizedProps?.[key]
              )
                probe[key]++;
          }
          visit(fiber.child);
          visit(fiber.sibling);
        };
        visit(root.current);
        previous = current;
      },
    };
    for (const type of [
      "copy",
      "cut",
      "paste",
      "beforeinput",
      "input",
      "compositionstart",
      "compositionend",
    ])
      document.addEventListener(
        type,
        (event) => {
          const element = event.target;
          if (
            !(
              element instanceof HTMLInputElement ||
              element instanceof HTMLTextAreaElement
            )
          )
            return;
          setTimeout(() =>
            probe.events.push({
              type,
              inputType: event instanceof InputEvent ? event.inputType : null,
              composing:
                event instanceof InputEvent ? event.isComposing : false,
              prevented: event.defaultPrevented,
              value: element.value,
              start: element.selectionStart,
              end: element.selectionEnd,
            }),
          );
        },
        true,
      );
  });
  const observations: unknown[] = [];
  const requests: { url: string; body: unknown }[] = [];
  page.on("request", (request) => {
    if (
      ["PATCH", "POST"].includes(request.method()) &&
      (/\/records\/[^/]+$/.test(request.url()) ||
        request.url().endsWith("/clipboard-paste"))
    )
      requests.push({ url: request.url(), body: request.postDataJSON() });
  });
  try {
    for (const surface of ["grid", "inspector"] as const) {
      for (const fieldKey of surface === "grid"
        ? [source, synopsis]
        : [raw, synopsis]) {
        const { recordId } = await seed(page);
        const input = await activate(page, recordId, fieldKey, surface);
        const element = await input.elementHandle();
        const initialRequests = requests.length;
        const record = async (
          action: string,
          expected: Record<string, unknown> = {},
        ) => {
          const observation = {
            surface,
            fieldKey,
            action,
            ...(await state(input)),
            clipboard: await page.evaluate(() =>
              navigator.clipboard.readText(),
            ),
            sameElement: await input.evaluate(
              (current, previous) => current === previous,
              element,
            ),
            requests: requests.length,
            work: await page.evaluate(
              () =>
                (window as unknown as { scalarClipboardProbe: unknown })
                  .scalarClipboardProbe,
            ),
          };
          observations.push(observation);
          expect(observation).toMatchObject({
            connected: true,
            focused: true,
            sameElement: true,
            ...expected,
          });
        };
        for (const [name, start, end, direction] of [
          ["partial copy", 6, 10, "forward"],
          ["select-all copy", 0, originalText.length, "forward"],
          ["backward copy", 6, 10, "backward"],
          ["collapsed copy", 6, 6, "forward"],
        ] as const) {
          await selection(input, start, end, direction);
          await page.evaluate(() =>
            navigator.clipboard.writeText("clipboard sentinel"),
          );
          await page.keyboard.press("Control+c");
          await record(name, {
            value: originalText,
            start,
            end,
            direction,
            clipboard:
              name === "collapsed copy"
                ? "clipboard sentinel"
                : name === "select-all copy"
                  ? originalText
                  : "beta",
            requests: initialRequests,
          });
        }
        await selection(input, 6, 10);
        await paste(page, "B🙂");
        await record("middle replacement", {
          value: "alpha B🙂 gamma",
          start: 9,
          end: 9,
        });
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
        await record("paste accepted", {
          value: "alpha B🙂 gamma",
          start: 9,
          end: 9,
          requests: initialRequests + 1,
        });
        await page.keyboard.type("Z");
        await record("continued typing", {
          value: "alpha B🙂Z gamma",
          start: 10,
          end: 10,
        });
        await page.keyboard.press("Control+z");
        await record("undo typing", {
          value: "alpha B🙂 gamma",
          start: 9,
          end: 9,
        });
        await page.keyboard.press("Control+z");
        await record("undo paste", {
          value: originalText,
          start: 6,
          end: 10,
          requests: initialRequests + 1,
        });
        await page.keyboard.press("Control+Shift+z");
        await record("redo paste", { value: "alpha B🙂 gamma" });
        await page.keyboard.press("Control+Shift+z");
        await record("redo typing", {
          value: "alpha B🙂Z gamma",
          requests: initialRequests + 1,
        });
        await selection(input, 0, 5, "backward");
        await page.keyboard.press("Control+x");
        await record("cut", {
          value: " B🙂Z gamma",
          clipboard: "alpha",
          start: 0,
          end: 0,
        });
        await selection(input, 0);
        await paste(page, '雪\t"quoted"\r\nline\n');
        const prefix =
          fieldKey === source ? '雪\t"quoted" line' : '雪\t"quoted"\nline\n';
        await record("beginning delimiters", {
          value: `${prefix} B🙂Z gamma`,
          start: prefix.length,
          end: prefix.length,
        });
        await selection(input, (await input.inputValue()).length);
        await paste(page, " END");
        await record("end paste", { value: `${prefix} B🙂Z gamma END` });
        await input.press("Enter");
        await expect(page.getByTestId(saveStateTestId())).not.toHaveText(
          "Syncing",
        );
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
        expect(requests).toHaveLength(initialRequests + 3);
        expect(
          requests.slice(initialRequests).map((request) => request.body),
        ).toEqual([
          expect.objectContaining({
            changes: [{ field_key: fieldKey, value: "alpha B🙂 gamma" }],
          }),
          expect.objectContaining({
            changes: [{ field_key: fieldKey, value: `${prefix} B🙂Z gamma` }],
          }),
          expect.objectContaining({
            changes: [
              { field_key: fieldKey, value: `${prefix} B🙂Z gamma END` },
            ],
          }),
        ]);
        observations.push({
          surface,
          fieldKey,
          action: "departure",
          saveState: await page.getByTestId(saveStateTestId()).innerText(),
          feedback: await page.getByRole("alert").allTextContents(),
          requests: requests.length,
        });
      }
    }
  } finally {
    await testInfo.attach("scalar-clipboard-observations", {
      body: JSON.stringify({ observations, requests }, null, 2),
      contentType: "application/json",
    });
  }
});

test("Timeline scalar native paste joins rapid departure and retains rejected drafts", async ({
  page,
}, info) => {
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  const observations: unknown[] = [];
  for (const surface of ["grid", "inspector"] as const) {
    for (const departure of ["Enter", "Tab", "blur", "Escape"] as const) {
      const { incidentId, recordId } = await seed(page);
      const input = await activate(page, recordId, synopsis, surface);
      const patches = await installPatchController(page);
      const held = patches.holdNextPatch({ recordId });
      try {
        await selection(input, 6, 10);
        await paste(page, "pasted Ω");
        if (departure === "blur")
          await input.evaluate((element: HTMLElement) => element.blur());
        else await page.keyboard.press(departure);
        await held.waitForHit;
        if (departure === "Enter" || departure === "Tab") {
          await expect(input).toBeFocused();
          await expect(input).toHaveValue("alpha pasted Ω gamma");
        }
        held.release();
        await held.waitForCompletion;
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
        await expectServerTimelineCells(page, incidentId, recordId, {
          [synopsis]: "alpha pasted Ω gamma",
        });
        expect(patches.calls).toHaveLength(1);
        if (surface === "grid" && departure !== "blur")
          await expect(input).toHaveCount(0);
        observations.push({
          surface,
          departure,
          calls: patches.calls,
          focus: await page.evaluate(() =>
            document.activeElement?.getAttribute("data-testid"),
          ),
        });
      } finally {
        held.release();
        await patches.dispose();
      }
    }
    const { recordId } = await seed(page);
    const input = await activate(page, recordId, synopsis, surface);
    const patches = await installPatchController(page);
    try {
      patches.failNextPatch(422, "invalid_request", { recordId });
      await selection(input, 0, originalText.length);
      await paste(page, "  exact rejected Ω\t\n  ");
      await page.keyboard.press("Tab");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      await expect(input).toHaveValue("  exact rejected Ω\t\n  ");
      await input.focus();
      await page.keyboard.press("Enter");
      expect(patches.calls).toHaveLength(1);
      await expect(input).toHaveValue("  exact rejected Ω\t\n  ");
      observations.push({
        surface,
        rejected: await state(input),
        calls: patches.calls,
      });
    } finally {
      await patches.dispose();
    }
  }
  await info.attach("scalar-paste-departure", {
    body: JSON.stringify(observations, null, 2),
    contentType: "application/json",
  });
});

test("Timeline scalar older receipts preserve newer equal edits and independent surfaces", async ({
  page,
}, info) => {
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  const observations: unknown[] = [];
  for (const surface of ["grid", "inspector"] as const) {
    const { incidentId, recordId } = await seed(page);
    const input = await activate(page, recordId, synopsis, surface);
    const element = await input.elementHandle();
    const patches = await installPatchController(page);
    const held = patches.holdNextPatch({ recordId });
    try {
      await selection(input, 0, originalText.length);
      await paste(page, "same paste Ω");
      await held.waitForHit;
      await input.press("Enter");
      // A later equal-text replacement fences this pending departure.
      await selection(input, 0, "same paste Ω".length);
      await paste(page, "same paste Ω");
      await page.keyboard.type(" newer");
      const before = await state(input);
      held.release();
      await held.waitForCompletion;
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(input).toBeFocused();
      expect(await state(input)).toEqual(before);
      expect(
        await input.evaluate((node, original) => node === original, element),
      ).toBe(true);
      // The newer identical paste is an admitted operation even when the driver
      // can settle it authoritatively without another changed value on the wire.
      const admissions = await page.evaluate(
        () =>
          performance.getEntriesByName(
            "cartulary.workbook.pending_unit_admitted",
          ).length,
      );
      expect(admissions).toBe(2);
      await input.press("Enter");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expectServerTimelineCells(page, incidentId, recordId, {
        [synopsis]: "same paste Ω newer",
      });
      observations.push({ surface, before, admissions, calls: patches.calls });
    } finally {
      held.release();
      await patches.dispose();
    }
  }

  const { incidentId, recordId } = await seed(page);
  await activate(page, recordId, synopsis, "grid");
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  const grid = await activate(page, recordId, synopsis, "grid");
  const patches = await installPatchController(page);
  const held = patches.holdNextPatch({ recordId });
  try {
    await selection(grid, 0, originalText.length);
    await paste(page, "grid submitted");
    await held.waitForHit;
    await page.keyboard.type(" newer grid");
    await page
      .getByRole("button", { name: "Find in loaded rows", exact: true })
      .click();
    const find = page.getByRole("textbox", {
      name: "Find in loaded rows",
      exact: true,
    });
    await expect(find).toBeFocused();
    const inspector = page.getByTestId(
      timelineScalarEditorTestId({
        recordId,
        fieldKey: synopsis,
        surface: "inspector",
      }),
    );
    await inspector.focus();
    await selection(inspector, 0, (await inspector.inputValue()).length);
    await page.keyboard.type("independent Inspector Ω");
    held.release();
    await held.waitForCompletion;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(grid).toHaveValue("grid submitted newer grid");
    await expect(inspector).toHaveValue("independent Inspector Ω");
    await expect(inspector).toBeFocused();
    expect(patches.calls).toHaveLength(1);
    await selection(inspector, (await inspector.inputValue()).length);
    await paste(page, " pasted");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectServerTimelineCells(page, incidentId, recordId, {
      [synopsis]: "independent Inspector Ω pasted",
    });
    await expect(grid).toHaveValue("grid submitted newer grid");
    observations.push({
      independent: {
        grid: await state(grid),
        inspector: await state(inspector),
      },
      calls: patches.calls,
    });
  } finally {
    held.release();
    await patches.dispose();
  }
  const detached = await seed(page);
  const detachedInput = await activate(
    page,
    detached.recordId,
    synopsis,
    "grid",
  );
  const detachedPatches = await installPatchController(page);
  const detachedHold = detachedPatches.holdNextPatch({
    recordId: detached.recordId,
  });
  try {
    await selection(detachedInput, 0, originalText.length);
    await paste(page, "accepted while detached");
    await detachedHold.waitForHit;
    await page.keyboard.type(" retained newer");
    await page.getByTestId(surfaceTabTestId(notesViewSchemaId)).click();
    detachedHold.release();
    await detachedHold.waitForCompletion;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
    const restored = await activate(page, detached.recordId, synopsis, "grid");
    await expect(restored).toHaveValue(
      "accepted while detached retained newer",
    );
    expect(detachedPatches.calls).toHaveLength(1);
    await restored.press("Enter");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectServerTimelineCells(
      page,
      detached.incidentId,
      detached.recordId,
      { [synopsis]: "accepted while detached retained newer" },
    );
    observations.push({ detached: true, calls: detachedPatches.calls });
  } finally {
    detachedHold.release();
    await detachedPatches.dispose();
  }
  await info.attach("scalar-revision-continuity", {
    body: JSON.stringify(observations, null, 2),
    contentType: "application/json",
  });
});

test("Timeline scalar native paste promotes one creation record and preserves current authoring", async ({
  page,
}, info) => {
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  const observations: unknown[] = [];
  for (const departure of ["continue", "Enter", "Tab"] as const) {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("SCALAR-CAPTURE"),
      "Scalar paste promotion",
    );
    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    const draftId = draftCellTestId(synopsis);
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: draftId,
    });
    const input = page.getByTestId(draftId);
    await input.focus();
    const held = await holdBrowserRequest(page, {
      method: "POST",
      path: `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
    });
    const requests: unknown[] = [];
    const record = (request: import("@playwright/test").Request) => {
      if (
        request.method() === "PATCH" ||
        (request.method() === "POST" && request.url().endsWith("/rows"))
      )
        requests.push({ url: request.url(), body: request.postDataJSON() });
    };
    page.on("request", record);
    try {
      await paste(page, "capture Ω\nsecond");
      if (departure !== "continue") await page.keyboard.press(departure);
      await held.waitForHit;
      if (departure === "continue") {
        await selection(input, 8);
        await page.keyboard.type("new ");
      }
      const before = await state(input);
      held.release();
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
      expect(rows).toHaveLength(1);
      const recordId = rows[0]?.record_id;
      if (!recordId) throw new Error("Missing promoted record");
      expect(held.hitCount()).toBe(1);
      if (departure === "continue") {
        const promoted = page.getByTestId(
          timelineScalarEditorTestId({
            recordId,
            fieldKey: synopsis,
            surface: "grid",
          }),
        );
        await expect(promoted).toBeFocused();
        await expect(promoted).toHaveValue(before.value);
        expect(await state(promoted)).toMatchObject({
          start: before.start,
          end: before.end,
        });
        await page.keyboard.type("attached ");
        await promoted.press("Enter");
        await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
        await expectServerTimelineCells(page, incidentId, recordId, {
          [synopsis]: "capture new attached Ω\nsecond",
        });
      } else if (departure === "Enter") {
        // Explicit Enter at the last record intentionally advances to the next
        // capture row. Paste alone above must remain on the promoted record.
        await expect(page.getByTestId(draftId)).toBeFocused();
        await expect(page.getByTestId(draftId)).toHaveValue("");
      } else {
        await expect(page.getByTestId(draftId)).not.toBeFocused();
      }
      expect(
        await queryViewRows(page, incidentId, timelineViewSchemaId),
      ).toHaveLength(1);
      observations.push({
        departure,
        before,
        requests,
        focus: await page.evaluate(() =>
          document.activeElement?.getAttribute("data-testid"),
        ),
      });
    } finally {
      held.release();
      await held.dispose();
      page.off("request", record);
    }
  }
  await info.attach("scalar-paste-promotion", {
    body: JSON.stringify(observations, null, 2),
    contentType: "application/json",
  });
});

test("Timeline scalar paste retains multiline scrolling composition and readable copy", async ({
  page,
}, info) => {
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  const { incidentId, recordId } = await seed(page);
  const input = await activate(page, recordId, raw, "inspector");
  const patches = await installPatchController(page);
  const held = patches.holdNextPatch({ recordId });
  const cdp = await page.context().newCDPSession(page);
  const element = await input.elementHandle();
  try {
    await selection(input, 0, originalText.length);
    await paste(page, "line Ω\n".repeat(40));
    await held.waitForHit;
    await selection(input, 15);
    await cdp.send("Input.imeSetComposition", {
      text: "仮",
      selectionStart: 1,
      selectionEnd: 1,
    });
    await input.evaluate((node: HTMLTextAreaElement) => {
      node.scrollTop = 40;
    });
    const composing = await state(input);
    held.release();
    await held.waitForCompletion;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    expect(await state(input)).toEqual(composing);
    expect(await input.evaluate((node, old) => node === old, element)).toBe(
      true,
    );
    expect(patches.calls).toHaveLength(1);
    await cdp.send("Input.insertText", { text: "確定" });
    await page.keyboard.type(" continued");
    const completed = await state(input);
    await input.press("Enter");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectServerTimelineCells(page, incidentId, recordId, {
      [raw]: completed.value,
    });
    const lifecycle = await currentLifecycle(page, incidentId);
    expect(
      (
        await lifecycleAction(page, incidentId, "closeIncident", {
          client_txn_id: uniqueTxn("scalar-readonly"),
          base_incident_version: lifecycle.incident_version,
          reason: "Readable scalar copy",
        })
      ).ok,
    ).toBe(true);
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)).getByRole("grid"),
    ).toHaveAttribute("aria-readonly", "true");
    await scrollGridCellIntoView({
      page,
      surface: timelineViewSchemaId,
      recordId,
      cellKey: raw,
    });
    const readableCell = page
      .getByTestId(gridRowTestId(timelineViewSchemaId, recordId))
      .locator(gridFieldCellSelector(raw))
      .first();
    await expect(readableCell).toBeVisible();
    await readableCell.click();
    if (!(await input.count()))
      await page
        .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
        .click();
    await expect(input).toHaveAttribute("readonly", "");
    await input.focus();
    await selection(input, 0, 4, "backward");
    await page.keyboard.press("Control+c");
    expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(
      "line",
    );
    const count = patches.calls.length;
    await page.keyboard.press("Control+x");
    await paste(page, "cannot insert");
    await expect(input).toHaveValue(completed.value);
    expect(patches.calls).toHaveLength(count);
    await info.attach("scalar-composition-readonly", {
      body: JSON.stringify(
        {
          composing,
          completed,
          readonly: await state(input),
          calls: patches.calls,
        },
        null,
        2,
      ),
      contentType: "application/json",
    });
  } finally {
    held.release();
    await patches.dispose();
    await cdp.detach();
  }
});

test("Timeline scalar paste replays captured requests and refreshes accepted writes without resend", async ({
  page,
}, info) => {
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  const { incidentId, recordId } = await seed(page);
  const input = await activate(page, recordId, synopsis, "grid");
  const route = `**/api/v1/records/${recordId}`;
  const requests: string[] = [],
    changes: string[] = [];
  await page.route(route, async (route) => {
    if (route.request().method() !== "PATCH") {
      await route.fallback();
      return;
    }
    requests.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    changes.push((await response.json()).data.change_set_id);
    if (requests.length === 1) await route.abort("failed");
    else await route.fulfill({ response });
  });
  const query = `**/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`;
  try {
    await selection(input, 6, 10);
    await paste(page, "captured Ω");
    await expect.poll(() => requests.length).toBe(2);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    expect(requests[1]).toBe(requests[0]);
    expect(changes[1]).toBe(changes[0]);
    let failedReads = 0;
    await page.route(query, async (route) => {
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
    expect(failedReads).toBe(1);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    expect(requests).toHaveLength(2);
    await page.unroute(query);
    await refresh.click();
    await expect(
      page.locator('[data-grid-data-state="stale_error"]'),
    ).toHaveCount(0);
    await expectServerTimelineCells(page, incidentId, recordId, {
      [synopsis]: "alpha captured Ω gamma",
    });
    expect(requests).toHaveLength(2);
    await info.attach("scalar-paste-recovery", {
      body: JSON.stringify({ requests, changes, failedReads }, null, 2),
      contentType: "application/json",
    });
  } finally {
    await page.unroute(query);
    await page.unroute(route);
  }
});
