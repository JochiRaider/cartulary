import { Buffer } from "node:buffer";
import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  pasteGridMatrix,
} from "@cartulary/test-utils/grid";
import {
  decisionSupersessionTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridGroupingSelectTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  savedViewCreateButtonTestId,
  savedViewNameInputTestId,
  savedViewSelectorTestId,
  savedViewStatusTestId,
  saveStateTestId,
  systemViewSwitcherTriggerTestId,
  workbookFilterPopoverTestId,
  workbookFocusAnchorTestId,
  workbookInspectorCloseButtonTestId,
  workbookQueryEntryTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  handoffViewSchemaId,
  lessonViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleSemanticGridCellFocus } from "./support/accessibility/grid";
import {
  expectNoFocusTrap,
  expectTabTraversalAdvancesFrom,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { openA11ySystemSurface } from "./support/accessibility/queryControls";
import { expectAndRecordContrast } from "./support/accessibility/records";
import {
  expectCurrentIncidentRole,
  expectWorkbookSavePresentation,
} from "./support/accessibility/status";
import { readCurrentSession } from "./support/auth/sessions";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import {
  openContextualCreationFixture,
  retainTimelineContextualSource,
} from "./support/workbook/contextualCreate";
import {
  fillCoordinationMinimum,
  openCoordinationFixture,
  retainCoordinationUncertainResult,
} from "./support/workbook/coordinationCreate";
import {
  expectDecisionControlReachable,
  openDecisionReviewFixture,
} from "./support/workbook/decisionSupersession";
import {
  expectCellTextOrValue,
  mountedGridCell,
  mountedGridTarget,
} from "./support/workbook/keyboardGrid";
import { createViewRow } from "./support/workbook/query";
import { openFilterPopover } from "./support/workbook/queryControls";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";
import { activateCandidateIdentities } from "./support/workbook/references";
import { openSavedViewActionMenu } from "./support/workbook/savedViews";

test("a11y.coordination-review.row-01 Verify coordination surfaces and full keyboard/clipboard controls meet keyboard reachability, focus visibility, accessible-name, ARIA, and non-color-only state expectations.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YCOORDINATIONREVIEW"),
    "a11y.coordination-review coordination accessibility",
  );
  const owner = await createIncidentMemberUser(page, incidentId, {
    display_name: "browser.coordination-review accessibility owner",
    email: uniqueEmail("coordination-review-a11y-owner"),
    initial_password: "BackupRestoreA11y1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const party = await createViewRow(page, incidentId, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.coordination-review-party"),
    "party.display_name": "a11y.coordination-review response party",
    "party.party_kind": "team",
  });
  const task = await createViewRow(page, incidentId, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.coordination-review-task"),
    "task.priority": "normal",
    "task.requester_party_id": party.record_id,
    "task.task_kind": "collection",
    "task.title": "a11y.coordination-review task alpha",
  });
  const urgentTask = await createViewRow(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.coordination-review-urgent-task"),
      "task.priority": "urgent",
      "task.task_kind": "follow_up",
      "task.title": "a11y.coordination-review task urgent",
    },
  );
  const clipboardRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.coordination-review-clipboard"),
      "timeline.activity_utc_text": "2026-06-12T10:00:00Z",
      "timeline.activity_synopsis_text":
        "a11y.coordination-review clipboard row",
    },
  );
  const decision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.coordination-review-decision"),
      "decision.decision_type": "containment",
      "decision.rationale": "a11y.coordination-review coordination rationale",
      "decision.summary": "a11y.coordination-review decision summary",
    },
  );
  const comm = await createViewRow(page, incidentId, commLogViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.coordination-review-comm"),
    "comm_log.audience": "a11y.coordination-review responders",
    "comm_log.channel_or_meeting": "a11y.coordination-review bridge",
    "comm_log.comm_type": "briefing",
    "comm_log.decision_ids": {
      actions: [{ linked_record_id: decision.record_id, op: "add_record_ref" }],
      kind: "collection_actions_v1",
    },
    "comm_log.summary": "a11y.coordination-review communications log",
  });
  const handoff = await createViewRow(page, incidentId, handoffViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.coordination-review-handoff"),
    "handoff.current_state_summary": "a11y.coordination-review handoff state",
    "handoff.incoming_owner_user_id": owner.user_id,
  });
  const status = await createViewRow(
    page,
    incidentId,
    statusReviewViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.coordination-review-status"),
      "status_review.current_state_summary":
        "a11y.coordination-review status review state",
    },
  );
  const lesson = await createViewRow(page, incidentId, lessonViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.coordination-review-lesson"),
    "lesson.summary": "a11y.coordination-review lesson summary",
  });

  const surfaces = [
    {
      expected: "a11y.coordination-review task alpha",
      fieldKey: "task.title",
      groupToken: "coordination",
      label: "Task Requests",
      row: task,
      viewSchemaId: taskRequestsViewSchemaId,
    },
    {
      expected: "a11y.coordination-review decision summary",
      fieldKey: "decision.summary",
      groupToken: "coordination",
      label: "Decisions",
      row: decision,
      viewSchemaId: decisionsViewSchemaId,
    },
    {
      expected: "a11y.coordination-review response party",
      fieldKey: "party.display_name",
      groupToken: "coordination",
      label: "Parties",
      row: party,
      viewSchemaId: partiesViewSchemaId,
    },
    {
      expected: "a11y.coordination-review communications log",
      fieldKey: "comm_log.summary",
      groupToken: "coordination",
      label: "Communications Log",
      row: comm,
      viewSchemaId: commLogViewSchemaId,
    },
    {
      expected: "a11y.coordination-review handoff state",
      fieldKey: "handoff.current_state_summary",
      groupToken: "coordination",
      label: "Handoff",
      row: handoff,
      viewSchemaId: handoffViewSchemaId,
    },
    {
      expected: "a11y.coordination-review status review state",
      fieldKey: "status_review.current_state_summary",
      groupToken: "coordination",
      label: "Status Review",
      row: status,
      viewSchemaId: statusReviewViewSchemaId,
    },
    {
      expected: "a11y.coordination-review lesson summary",
      fieldKey: "lesson.summary",
      groupToken: "coordination",
      label: "Lesson",
      row: lesson,
      viewSchemaId: lessonViewSchemaId,
    },
  ] as const;

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      taskRequestsViewSchemaId,
    )}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectCurrentIncidentRole(page, "admin");
  await expectWorkbookSavePresentation(page);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expectTabTraversalAdvancesFrom(
    page,
    page.getByTestId(systemViewSwitcherTriggerTestId()),
    4,
    true,
  );

  for (const surface of surfaces) {
    await test.step(`${surface.label} accessibility surface`, async () => {
      await openA11ySystemSurface(page, surface);
      await expect(page).toHaveURL(
        new RegExp(
          `view_schema_id=${encodeURIComponent(surface.viewSchemaId)}`,
        ),
      );
      await expect(
        page.getByTestId(savedViewSelectorTestId(surface.viewSchemaId)),
      ).toHaveAttribute("data-selected-sheet-ref-kind", "view_schema");
      await openFilterPopover(page, surface.viewSchemaId);
      await expect(
        page.getByTestId(gridFilterFieldTestId(surface.viewSchemaId)),
      ).toBeVisible();
      await expect(
        page.getByTestId(gridFilterApplyTestId(surface.viewSchemaId)),
      ).toBeVisible();
      await expect(
        page.getByTestId(gridGroupingSelectTestId(surface.viewSchemaId)),
      ).toBeVisible();
      await page.keyboard.press("Escape");
      await expect(
        page.getByTestId(workbookFilterPopoverTestId(surface.viewSchemaId)),
      ).toHaveCount(0);

      const sortHeader = await mountedGridTarget(
        page,
        surface.viewSchemaId,
        gridSortHeaderTestId(surface.viewSchemaId, surface.fieldKey),
      );
      await expectVisibleFocus(sortHeader);
      await expect(sortHeader).not.toHaveText("");

      const cell = await mountedGridCell(
        page,
        surface.viewSchemaId,
        surface.row.record_id,
        surface.fieldKey,
      );
      await expectVisibleSemanticGridCellFocus(cell);
      await expect(page.getByTestId(workbookFocusAnchorTestId())).toContainText(
        surface.viewSchemaId,
      );
      await expectCellTextOrValue(cell, surface.expected);
    });
  }

  await openA11ySystemSurface(page, {
    groupToken: "coordination",
    viewSchemaId: taskRequestsViewSchemaId,
  });

  const taskTitle = await mountedGridCell(
    page,
    taskRequestsViewSchemaId,
    task.record_id,
    "task.title",
  );
  const taskTitleCell = await expectVisibleSemanticGridCellFocus(taskTitle);
  const copiedTaskTitle = await taskTitleCell.evaluate((element) => {
    const data = new DataTransfer();
    element.dispatchEvent(
      new ClipboardEvent("copy", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
    return data.getData("text/plain");
  });
  expect(copiedTaskTitle).toBe("a11y.coordination-review task alpha");

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  const clipboardSummary = await mountedGridCell(
    page,
    timelineViewSchemaId,
    clipboardRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expectVisibleSemanticGridCellFocus(clipboardSummary);
  await pasteGridMatrix({
    fieldKey: "timeline.activity_synopsis_text",
    matrix: [
      [
        "a11y.coordination-review pasted timeline",
        "a11y.coordination-review-host",
      ],
    ],
    page,
    recordId: clipboardRow.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(clipboardSummary).toHaveText(
    "a11y.coordination-review pasted timeline",
  );

  await openA11ySystemSurface(page, {
    groupToken: "coordination",
    viewSchemaId: taskRequestsViewSchemaId,
  });

  await applyFilterChip(
    page,
    taskRequestsViewSchemaId,
    "task.priority",
    "urgent",
  );
  await assertActiveFilterChipVisible(
    page,
    taskRequestsViewSchemaId,
    "task.priority",
  );
  const priorityChip = page.getByTestId(
    workbookQueryEntryTestId(
      taskRequestsViewSchemaId,
      "filter",
      "task.priority",
    ),
  );
  await expect(priorityChip).toContainText("urgent");
  await expectVisibleFocus(priorityChip);

  const urgentTitle = await mountedGridCell(
    page,
    taskRequestsViewSchemaId,
    urgentTask.record_id,
    "task.title",
  );
  await expectCellTextOrValue(
    urgentTitle,
    "a11y.coordination-review task urgent",
  );

  await openSavedViewActionMenu(page, taskRequestsViewSchemaId);
  const savedViewName = page.getByTestId(
    savedViewNameInputTestId(taskRequestsViewSchemaId),
  );
  await expectVisibleFocus(savedViewName);
  await savedViewName.fill("a11y.coordination-review keyboard saved view");
  const createSavedView = page.getByTestId(
    savedViewCreateButtonTestId(taskRequestsViewSchemaId),
  );
  await expectVisibleFocus(createSavedView);
  await createSavedView.press("Enter");
  const savedStatus = page.getByTestId(
    savedViewStatusTestId(taskRequestsViewSchemaId),
  );
  await expect(savedStatus).toHaveAttribute("aria-live", "polite");
  await expect(savedStatus).toHaveText("Saved view created.");

  await test
    .info()
    .attach("a11y.coordination-review-01-readiness-matrix.json", {
      contentType: "application/json",
      body: Buffer.from(
        `${JSON.stringify(
          {
            checks: [
              "keyboard reachability",
              "visible focus",
              "accessible names",
              "system-view menu ARIA state",
              "status and saved-view live regions",
              "clipboard copy and paste",
              "non-color-only filter chip state",
            ],
            scenario_title: test.info().title,
            stable_identity_scope: "view_schema_id + record_id + field_key",
            surfaces: surfaces.map((surface) => ({
              field_key: surface.fieldKey,
              record_id: surface.row.record_id,
              surface: surface.label,
              view_schema_id: surface.viewSchemaId,
            })),
            viewport: "1440x900",
            zoom: "100%",
          },
          null,
          2,
        )}\n`,
      ),
    });

  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  await openFilterPopover(page, taskRequestsViewSchemaId);
  await expectAndRecordContrast(
    page,
    [
      gridFilterFieldTestId(taskRequestsViewSchemaId),
      gridFilterApplyTestId(taskRequestsViewSchemaId),
    ],
    "filter-popover",
  );
  await openSavedViewActionMenu(page, taskRequestsViewSchemaId);
  await expectAndRecordContrast(page, [
    systemViewSwitcherTriggerTestId(),
    savedViewSelectorTestId(taskRequestsViewSchemaId),
    savedViewNameInputTestId(taskRequestsViewSchemaId),
    savedViewCreateButtonTestId(taskRequestsViewSchemaId),
    savedViewStatusTestId(taskRequestsViewSchemaId),
    workbookQueryEntryTestId(
      taskRequestsViewSchemaId,
      "filter",
      "task.priority",
    ),
    gridGroupingSelectTestId(taskRequestsViewSchemaId),
    gridSortHeaderTestId(taskRequestsViewSchemaId, "task.title"),
    rowCellTestId(urgentTask.record_id, "task.title"),
    saveStateTestId(),
  ]);
});

test("a11y.decision-supersession review cancellation and shell recovery preserve keyboard access", async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await openDecisionReviewFixture(page);
  const editor = page.getByTestId(decisionSupersessionTestId("editor"));
  const confirm = page.getByTestId(decisionSupersessionTestId("confirm"));
  await expect(confirm).toBeFocused();
  const cancel = editor.getByRole("button", {
    name: "Cancel review",
    exact: true,
  });
  await cancel.focus();
  await cancel.press("Enter");
  const review = page.getByTestId(decisionSupersessionTestId("review-action"));
  await expectVisibleFocus(review);
  await review.press("Enter");
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [
      confirm,
      cancel,
      page.getByTestId(decisionSupersessionTestId("reason")),
    ]) {
      await expectDecisionControlReachable(page, control);
      await expectVisibleFocus(control);
    }
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectDecisionControlReachable(page, confirm);
    await expectVisibleFocus(confirm);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await expectDecisionControlReachable(page, cancel);
    },
  );
  await expectAllInteractiveControlsNamed(page);
  await confirm.focus();
  await confirm.press("Enter");
  const trigger = recoveryEntry(page);
  await trigger.focus();
  await expect(recoveryEntry(page)).toHaveText("Recovery (0)");
  await openRecoveryItem(page, /^Decision supersession ·/);
  const recovery = page.getByTestId(decisionSupersessionTestId("recovery"));
  await expect(
    page
      .getByRole("region", { name: "Recovery navigation", exact: true })
      .locator(":scope > h2"),
  ).toBeFocused();
  await expect(
    recovery.getByText(
      "Supersession accepted. Both Decisions and related projections refreshed.",
      { exact: true },
    ),
  ).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
});

test("a11y.contextual-create target fields reference cancellation and retained recovery remain keyboard accessible", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const actor = await readCurrentSession(page);
  for (const target of ["task_request", "decision"] as const) {
    await page.setViewportSize({ width: 1280, height: 720 });
    const { view, source } = await openContextualCreationFixture(page, target);
    const field = page.getByTestId(
      genericCreateFieldTestId(
        target === "decision" ? "decision.rationale" : "task.title",
      ),
    );
    const original = await field.inputValue();
    if (target === "task_request") {
      const chooseOwner = page.getByRole("button", {
        name: "Choose Owner",
        exact: true,
      });
      await chooseOwner.focus();
      await chooseOwner.press("Enter");
      const ownerPicker = page.getByRole("region", {
        name: "Choose Owner",
        exact: true,
      });
      const ownerCandidates = ownerPicker.getByRole("group", {
        name: "Owner",
        exact: true,
      });
      await expect(
        ownerCandidates.getByRole("radio", { checked: true }),
      ).toHaveValue(actor.user_id);
      await expect(ownerPicker.getByRole("list")).toContainText(
        actor.display_name,
      );
      await expect(
        ownerPicker.getByRole("button", { name: /^Remove selected Owner / }),
      ).toHaveAccessibleName(`Remove selected Owner ${actor.display_name}`);
      const applyOwner = ownerPicker.getByRole("button", {
        name: "Apply references",
        exact: true,
      });
      await applyOwner.focus();
      await applyOwner.press("Enter");
      await expect(chooseOwner).toBeFocused();
      await expect(
        page.getByRole("button", { name: /^Remove Owner / }),
      ).toHaveAccessibleName(`Remove Owner ${actor.display_name}`);
    }
    const submit = page.getByTestId(genericCreateSubmitTestId(view));
    await field.fill("");
    await submit.focus();
    await submit.press("Enter");
    await expect(field).toHaveAttribute("aria-invalid", "true");
    await expect(field).toHaveAccessibleDescription(/is required/);
    await field.fill(original);
    const timestamp = page.getByTestId(
      genericCreateFieldTestId(
        target === "decision" ? "decision.decided_at" : "task.due_at",
      ),
    );
    await timestamp.fill("unfinished");
    await submit.focus();
    await submit.press("Enter");
    await expect(timestamp).toHaveAttribute("aria-invalid", "true");
    await timestamp.fill("");
    const choose = page.getByRole("button", {
      name:
        target === "decision" ? "Choose Support Refs" : "Choose Linked Records",
      exact: true,
    });
    await choose.focus();
    await choose.press("Enter");
    const picker = page.getByRole("region", {
      name:
        target === "decision" ? "Choose Support Refs" : "Choose Linked Records",
      exact: true,
    });
    await expect(
      picker.getByRole("button", { name: "Apply references", exact: true }),
    ).toBeEnabled();
    const candidates = picker.getByRole("group", {
      name: target === "decision" ? "Support Refs" : "Linked Records",
      exact: true,
    });
    const originalIds = [source.record_id];
    const expectAcceptedSelection = async (ids: readonly string[]) => {
      await expect(
        picker.getByRole("status").filter({
          hasText:
            /^Page \d+: \d+ candidates; (more available|end of this query)\.$/,
        }),
      ).toBeVisible();
      await expect(
        candidates.locator(
          `input[type="checkbox"][value="${source.record_id}"]`,
        ),
      ).toBeVisible();
      await expect
        .poll(() =>
          candidates
            .getByRole("checkbox", { checked: true })
            .evaluateAll((inputs) =>
              inputs.map((input) => (input as HTMLInputElement).value).sort(),
            ),
        )
        .toEqual([...ids].sort());
    };
    await expectAcceptedSelection(originalIds);
    await activateCandidateIdentities(candidates, []);
    await expectAcceptedSelection([]);
    await candidates.press("Escape");
    await expect(picker).toBeHidden();
    await expect(choose).toBeFocused();
    await choose.press("Enter");
    await expectAcceptedSelection(originalIds);
    await candidates.press("Escape");
    for (const viewport of [
      { width: 1280, height: 720 },
      { width: 768, height: 640 },
      { width: 390, height: 480 },
    ]) {
      await page.setViewportSize(viewport);
      for (const control of [field, timestamp, choose, submit]) {
        await expectDecisionControlReachable(page, control);
        await expectVisibleFocus(control);
      }
      await testInfo.attach(`contextual-${target}-${viewport.width}`, {
        body: await page.screenshot({ animations: "disabled", caret: "hide" }),
        contentType: "image/png",
      });
    }
    await page.setViewportSize({ width: 1280, height: 720 });
    await withPresentation(page, { zoom: "200%" }, async () => {
      await expectDecisionControlReachable(page, submit);
    });
    await withPresentation(
      page,
      {
        textSpacing:
          "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
      },
      async () => {
        await expectDecisionControlReachable(page, choose);
      },
    );
    await page
      .getByRole("button", { name: "Keep draft and close", exact: true })
      .click();
    const summary = recoveryEntry(page);
    await summary.focus();
    await openRecoveryItem(page, /^(Task Requests|Decisions) draft ·/);
    const recovery = page.getByRole("region", {
      name: "Retained contextual creation",
      exact: true,
    });
    const resume = recovery.getByRole("button", {
      name: "Resume contextual draft",
      exact: true,
    });
    await resume.focus();
    await resume.press("Enter");
    await expect(field).toHaveValue(original);
    if (target === "task_request")
      await expect(
        recovery.getByRole("button", { name: /^Remove Owner / }),
      ).toHaveAccessibleName(`Remove Owner ${actor.display_name}`);
    await page.setViewportSize({ width: 390, height: 480 });
    await expectDecisionControlReachable(page, field);
    await expectDecisionControlReachable(page, submit);
    await expectAllInteractiveControlsNamed(page);
    await testInfo.attach(`contextual-${target}-recovery-tree`, {
      body: await recovery.ariaSnapshot(),
      contentType: "text/plain",
    });
    await recovery.press("Escape");
    await expect(summary).toBeFocused();
    await expect(recovery).not.toBeVisible();
    await page.setViewportSize({ width: 1280, height: 720 });
    const retainedSource = await retainTimelineContextualSource(page, target);
    for (const viewport of [
      { width: 1280, height: 720 },
      { width: 768, height: 640 },
      { width: 390, height: 480 },
    ]) {
      await page.setViewportSize(viewport);
      const context = retainedSource.recovery.getByText(
        /^Create in .* Source:/,
      );
      await context.scrollIntoViewIfNeeded();
      await expect(context).toContainText(retainedSource.label);
      const bounds = await context.boundingBox();
      expect(bounds).not.toBeNull();
      if (bounds)
        expect(bounds.x + bounds.width).toBeLessThanOrEqual(viewport.width);
      const origin = page
        .getByRole("region", { name: "Recovery navigation", exact: true })
        .getByText(`${retainedSource.label} · Draft retained`, { exact: true });
      await expect(origin).toBeVisible();
      expect(
        await origin.evaluate(
          (element) => element.scrollWidth <= element.clientWidth,
        ),
      ).toBe(true);
      const input = retainedSource.recovery.getByTestId(
        genericCreateFieldTestId(retainedSource.scalar),
      );
      await expectDecisionControlReachable(page, input);
      await expectVisibleFocus(input);
      await expect(input).toHaveValue(
        "Exact retained scalar after link removal",
      );
      await testInfo.attach(
        `contextual-${target}-timeline-source-${viewport.width}`,
        {
          body: await page.screenshot({
            animations: "disabled",
            caret: "hide",
          }),
          contentType: "image/png",
        },
      );
    }
  }
});

test("a11y.coordination all target fields source review and uncertain recovery support keyboard and narrow layouts", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const variant of ["comm_log", "handoff", "status_review", "lesson"]) {
    const f = await openCoordinationFixture(page, variant);
    const submit = f.form.getByTestId(
      genericCreateSubmitTestId(f.target.viewSchemaId),
    );
    await submit.focus();
    await submit.press("Enter");
    await expect(f.form.getByRole("alert").first()).toBeVisible();
    await fillCoordinationMinimum(f);
    for (const viewport of [
      { width: 1280, height: 720 },
      { width: 390, height: 480 },
    ]) {
      await page.setViewportSize(viewport);
      for (const field of f.target.fields.filter(
        (field) => field.createWritable,
      )) {
        const control =
          (field.readKind === "collection" &&
            field.fieldKey !== "handoff.open_risk_refs") ||
          field.directReferenceContractId
            ? f.form.getByRole("button", {
                name: `Choose ${field.label.toLowerCase()}`,
                exact: true,
              })
            : f.form.getByTestId(genericCreateFieldTestId(field.fieldKey));
        await expectDecisionControlReachable(page, control);
        await expectVisibleFocus(control);
      }
      await expectDecisionControlReachable(page, submit);
      await expectVisibleFocus(submit);
      await testInfo.attach(`coordination-${variant}-${viewport.width}`, {
        body: await page.screenshot({ animations: "disabled", caret: "hide" }),
        contentType: "image/png",
      });
    }
    const choose = f.form.getByRole("button", {
      name: "Choose source",
      exact: true,
    });
    await choose.focus();
    await choose.press("Enter");
    const picker = f.form.getByRole("region", {
      name: "Choose source",
      exact: true,
    });
    await expect(
      picker.getByRole("button", { name: "Apply references", exact: true }),
    ).toBeEnabled();
    await picker
      .getByRole("group", { name: "Source", exact: true })
      .press("Escape");
    await expect(choose).toBeFocused();
    await expectAllInteractiveControlsNamed(page);
    if (variant === "lesson") {
      const { recovery, summary } = await retainCoordinationUncertainResult(
        page,
        f,
      );
      const recover = recovery.getByRole("button", {
        name: "Recover submission",
        exact: true,
      });
      for (const profile of [
        { width: 1280, height: 720, zoom: "" },
        { width: 390, height: 320, zoom: "" },
        { width: 1280, height: 720, zoom: "200%" },
      ]) {
        await page.setViewportSize({
          width: profile.width,
          height: profile.height,
        });
        await withPresentation(page, { zoom: profile.zoom }, async () => {
          await expectDecisionControlReachable(page, recover);
          await expectVisibleFocus(recover);
          await expect(recover).toBeInViewport();
          await expect(page.getByTestId(saveStateTestId())).toBeInViewport();
          const bounds = await recovery.evaluate((panel) => {
            const host = panel.closest<HTMLElement>(
              '[aria-label="Workbook recovery"]',
            );
            if (!host) throw new Error("Missing shell recovery host");
            const parent = host.parentElement;
            const box = host.getBoundingClientRect();
            const available = parent?.getBoundingClientRect();
            return {
              contained:
                available !== undefined &&
                box.top >= available.top &&
                box.bottom <= available.bottom + 1,
              internalScroll: getComputedStyle(host).overflowY,
              documentFits:
                document.documentElement.scrollHeight <=
                document.documentElement.clientHeight + 1,
            };
          });
          expect(bounds.contained).toBe(true);
          expect(bounds.internalScroll).toBe("auto");
          expect(bounds.documentFits).toBe(true);
        });
      }

      await page.setViewportSize({ width: 390, height: 480 });
      await testInfo.attach("coordination-recovery-tree", {
        body: await recovery.ariaSnapshot(),
        contentType: "text/plain",
      });
      await recover.press("Escape");
      await expect(summary).toBeFocused();
      await expect(recovery).not.toBeVisible();
      const systemViews = page.getByRole("button", {
        name: "More views",
        exact: true,
      });
      const account = page.getByRole("button", {
        name: "Account and application navigation",
        exact: true,
      });
      await expectDecisionControlReachable(page, systemViews);
      await expectVisibleFocus(systemViews);
      await expectDecisionControlReachable(page, account);
      await expectVisibleFocus(account);
      await account.press("Enter");
      await expect(
        page.getByRole("menuitem", { name: "Incidents", exact: true }),
      ).toBeFocused();
      await page.keyboard.press("Escape");
      await expect(account).toBeFocused();
    }
  }

  // Retained editable content and acknowledged read-recovery share the same bounds.
  const f = await openCoordinationFixture(page, "lesson");
  await fillCoordinationMinimum(f);
  await page.getByTestId(workbookInspectorCloseButtonTestId(f.view)).click();
  await openRecoveryItem(page, /^Coordination draft ·/);
  const retained = page.getByRole("region", {
    name: "Retained Coordination authoring",
    exact: true,
  });
  await retained
    .getByRole("button", { name: "Resume Coordination draft", exact: true })
    .click();
  const checkRecoveryControls = async () => {
    for (const profile of [
      { width: 1280, height: 720, zoom: "" },
      { width: 390, height: 320, zoom: "" },
      { width: 1280, height: 720, zoom: "200%" },
    ]) {
      await page.setViewportSize({
        width: profile.width,
        height: profile.height,
      });
      await withPresentation(page, { zoom: profile.zoom }, async () => {
        await page.evaluate(
          () =>
            new Promise<void>((resolve) =>
              requestAnimationFrame(() =>
                requestAnimationFrame(() => resolve()),
              ),
            ),
        );
        for (const role of ["textbox", "combobox", "button"] as const) {
          for (const control of await retained.getByRole(role).all()) {
            if (!(await control.isVisible()) || !(await control.isEnabled()))
              continue;
            await expectDecisionControlReachable(page, control);
            await expectVisibleFocus(control);
            await expect(control).toBeInViewport();
          }
        }
        await expect(page.getByTestId(saveStateTestId())).toBeInViewport();
        expect(
          await page.evaluate(
            () =>
              document.documentElement.scrollHeight <=
              document.documentElement.clientHeight + 1,
          ),
        ).toBe(true);
      });
    }
  };
  await checkRecoveryControls();
  const refreshPath = `**/incidents/${f.incident}/views/${f.target.viewSchemaId}/query`;
  await installOwnedRoute(page, refreshPath, (route) => route.abort("failed"));
  let creations = 0;
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      new URL(request.url()).pathname.endsWith(
        `/views/${f.target.viewSchemaId}/rows`,
      )
    )
      creations += 1;
  });
  await retained
    .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
    .click();
  await expect(retained).toContainText(
    "Coordination created, but views need refresh.",
  );
  const retryRefresh = retained.getByRole("button", {
    name: "Retry refresh",
    exact: true,
  });
  await expect(retryRefresh).toBeEnabled();
  await expect(
    page
      .getByRole("region", { name: "Recovery navigation", exact: true })
      .locator(":scope > h2"),
  ).toBeFocused();
  await checkRecoveryControls();
  await removeOwnedRoute(page, refreshPath);
  await retryRefresh.click();
  await expect(retained).not.toBeVisible();
  expect(creations).toBe(1);
});
