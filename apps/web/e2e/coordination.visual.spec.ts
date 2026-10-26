import { Buffer } from "node:buffer";
import {
  decisionSupersessionTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  lessonViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installOwnedRoute } from "./support/transport/requestInterception";
import {
  assertViewportVisualRegression,
  type VisualAnchor,
} from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import { createRichTimelineSource } from "./support/visual/timelineVisualFixture";
import {
  assertWorkbookGridVisualRegression,
  hideVisualTechnicalColumns,
} from "./support/visual/workbookCaptures";
import { normalizeWorkbookGridVisualState } from "./support/visual/workbookGeometry";
import {
  openContextualCreationFixture,
  retainedTimelineSourceLabel,
  retainTimelineContextualSource,
} from "./support/workbook/contextualCreate";
import {
  fillCoordinationMinimum,
  openCoordinationFixture,
  retainCoordinationUncertainResult,
} from "./support/workbook/coordinationCreate";
import { openDecisionReviewFixture } from "./support/workbook/decisionSupersession";
import { mountedGridCell } from "./support/workbook/keyboardGrid";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";

test("Capture Task Requests or Decisions, Parties link state, Communications Log, Handoff, Status Review, Lesson, and keyboard focus fixtures.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });

  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALCOORDINATIONREVIEW"),
    "browser.coordination-review coordination visual readiness",
  );
  const owner = await createIncidentMemberUser(page, incidentId, {
    display_name: "browser.coordination-review visual owner",
    email: uniqueEmail("coordination-review-visual-owner"),
    initial_password: "BackupRestoreVisual1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const party = await createViewRow(page, incidentId, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-PARTY"),
    "party.display_name": "browser.coordination-review Visual Party",
    "party.party_kind": "team",
  });
  const taskRow = await createViewRow(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    {
      client_txn_id: uniqueTxn("VISUALCOORDINATION-TASK"),
      "task.task_kind": "collection",
      "task.title": "Visual task request",
    },
  );
  const decision = await createViewRow(
    page,
    incidentId,
    decisionsViewSchemaId,
    {
      client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-DECISION"),
      "decision.decision_type": "containment",
      "decision.rationale": "browser.coordination-review visual rationale",
      "decision.summary": "browser.coordination-review visual decision",
    },
  );
  const comm = await createViewRow(page, incidentId, commLogViewSchemaId, {
    client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-COMM"),
    "comm_log.audience": "browser.coordination-review visual responders",
    "comm_log.channel_or_meeting": "browser.coordination-review visual bridge",
    "comm_log.comm_type": "briefing",
    "comm_log.decision_ids": {
      actions: [{ linked_record_id: decision.record_id, op: "add_record_ref" }],
      kind: "collection_actions_v1",
    },
    "comm_log.summary": "browser.coordination-review visual communication",
  });
  const handoff = await createViewRow(page, incidentId, handoffViewSchemaId, {
    client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-HANDOFF"),
    "handoff.current_state_summary":
      "browser.coordination-review visual handoff state",
    "handoff.incoming_owner_user_id": owner.user_id,
  });
  const status = await createViewRow(
    page,
    incidentId,
    statusReviewViewSchemaId,
    {
      client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-STATUS"),
      "status_review.current_state_summary":
        "browser.coordination-review visual status review state",
    },
  );
  const lesson = await createViewRow(page, incidentId, lessonViewSchemaId, {
    client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-LESSON"),
    "lesson.summary": "browser.coordination-review visual lesson",
  });

  await navigateVisualApplication(
    page,
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      taskRequestsViewSchemaId,
    )}`,
  );

  await expect(
    await mountedGridCell(
      page,
      taskRequestsViewSchemaId,
      taskRow.record_id,
      "task.title",
    ),
  ).toHaveText("Visual task request");
  await expect(
    await mountedGridCell(
      page,
      taskRequestsViewSchemaId,
      taskRow.record_id,
      "task.status",
    ),
  ).toHaveText("open");
  await normalizeWorkbookGridVisualState(page, taskRequestsViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await hideVisualTechnicalColumns(page, taskRequestsViewSchemaId, ["Owner"]);
  await assertWorkbookGridVisualRegression(
    page,
    "record-relationships-task-requests",
    taskRequestsViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );

  const linkedTask = await createViewRow(
    page,
    incidentId,
    taskRequestsViewSchemaId,
    {
      client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-LINKED-TASK"),
      "task.requester_party_id": party.record_id,
      "task.requester_party_text": "browser.coordination-review requester",
      "task.task_kind": "follow_up",
      "task.title": "browser.coordination-review party-linked task",
    },
  );

  const surfaceExpectations = [
    {
      expected: "browser.coordination-review party-linked task",
      fieldKey: "task.title",
      recordId: linkedTask.record_id,
      surface: taskRequestsViewSchemaId,
    },
    {
      expected: "browser.coordination-review Visual Party",
      fieldKey: "party.display_name",
      recordId: party.record_id,
      surface: partiesViewSchemaId,
    },
    {
      expected: "browser.coordination-review visual communication",
      fieldKey: "comm_log.summary",
      recordId: comm.record_id,
      surface: commLogViewSchemaId,
    },
    {
      expected: "browser.coordination-review visual handoff state",
      fieldKey: "handoff.current_state_summary",
      recordId: handoff.record_id,
      surface: handoffViewSchemaId,
    },
    {
      expected: "browser.coordination-review visual status review state",
      fieldKey: "status_review.current_state_summary",
      recordId: status.record_id,
      surface: statusReviewViewSchemaId,
    },
    {
      expected: "browser.coordination-review visual lesson",
      fieldKey: "lesson.summary",
      recordId: lesson.record_id,
      surface: lessonViewSchemaId,
    },
  ] as const;
  for (const expectation of surfaceExpectations) {
    await navigateVisualApplication(
      page,
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        expectation.surface,
      )}`,
    );

    await expect(
      page.getByTestId(gridShellTestId(expectation.surface)),
    ).toBeVisible();
    const cell = await mountedGridCell(
      page,
      expectation.surface,
      expectation.recordId,
      expectation.fieldKey,
    );
    await expect(cell).toHaveText(expectation.expected);
    await cell.focus();
    await normalizeWorkbookGridVisualState(page, expectation.surface, {
      scroll: { top: 0, left: "left" },
    });
  }

  const linkedRows = await queryViewRows(
    page,
    incidentId,
    taskRequestsViewSchemaId,
  );
  const linked = linkedRows.find(
    (candidate) => candidate.record_id === linkedTask.record_id,
  );
  const requesterPartyCell = linked?.cells["task.requester_party_id"] as
    | { value?: unknown }
    | undefined;
  expect(requesterPartyCell?.value).toBe(party.record_id);
  await test.info().attach("coordination-visual-fixture-matrix.json", {
    body: Buffer.from(
      JSON.stringify(
        {
          browser_zoom_percent: 100,
          device_scale_factor: 1,
          dynamic_masks: [
            "incident identifiers",
            "generated record identifiers",
            "clock-derived labels when visible",
          ],
          editor_state:
            "visual.fixture.edit_cell readonly editor fixture captured in the grid-adapter fixture.",
          fixture_ids: [
            "visual.fixture.task_requests_or_decisions",
            "visual.fixture.frozen_column",
            "visual.fixture.resize_handle",
            "visual.fixture.drag_fill_handle",
            "visual.fixture.edit_cell",
            "visual.fixture.tree_group_row",
          ],
          focus_state:
            "Each browser.coordination-review coordination/review surface focuses its row cell after deterministic query state is visible.",
          screenshot_scopes: [
            "task requests workbook grid",
            "grid-adapter design fixture",
          ],
          seed_id: "coordination-visual-deterministic-seed",
          surfaces: surfaceExpectations.map((entry) => entry.surface),
          viewport_css_px: "1440x900",
        },
        null,
        2,
      ),
      "utf8",
    ),
    contentType: "application/json",
  });
});

test("Capture Decision supersession review and accepted recovery at desktop and narrow widths", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  await openDecisionReviewFixture(
    page,
    (url) => navigateVisualApplication(page, url),
    {
      target: "2026-04-18T14:44:08Z",
      replacement: "2026-04-18T15:42:09Z",
    },
  );
  await hideVisualTechnicalColumns(page, decisionsViewSchemaId, ["Owner"]);
  const review = page.getByTestId(decisionSupersessionTestId("review"));
  const reviewAnchor: VisualAnchor = {
    locator: review,
    align: "start",
    scrollportSelector: `aside[data-view-schema-id="${decisionsViewSchemaId}"] [data-inspector-scroll-body]`,
  };
  await assertViewportVisualRegression(page, "decision-supersession-review", {
    anchor: reviewAnchor,
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await assertViewportVisualRegression(
    page,
    "decision-supersession-review-narrow",
    { anchor: reviewAnchor },
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await expect(review).toBeVisible();
  await page.getByTestId(decisionSupersessionTestId("confirm")).click();
  await openRecoveryItem(page, /^Decision supersession ·/);
  const recovery = page.getByTestId(decisionSupersessionTestId("recovery"));
  await expect(
    recovery.getByText(
      "Supersession accepted. Both Decisions and related projections refreshed.",
      { exact: true },
    ),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "decision-supersession-accepted");
});

test("Capture contextual Task and Decision authoring references and retained recovery", async ({
  page,
}) => {
  const capture = async (
    page: Page,
    name: string,
    options: { anchor?: VisualAnchor } = {},
  ) => {
    await assertViewportVisualRegression(page, name, options);
    await test.info().attach(`${name}-review`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  };
  for (const target of ["task_request", "decision"] as const) {
    await page.setViewportSize({ width: 1280, height: 720 });
    const { view, source } = await openContextualCreationFixture(
      page,
      target,
      (url) => navigateVisualApplication(page, url),
    );
    const form = page.getByRole("region", {
      name: target === "decision" ? "Create decision" : "Create task request",
      exact: true,
    });
    const anchor: VisualAnchor = {
      locator: form,
      align: "start",
      scrollportSelector: `aside[data-view-schema-id="${evidenceViewSchemaId}"] [data-inspector-scroll-body]`,
    };
    await capture(page, `contextual-${target}-authoring`, { anchor });
    await page.setViewportSize({ width: 768, height: 640 });
    await capture(page, `contextual-${target}-authoring-narrow`, { anchor });
    const choose = form.getByRole("button", {
      name:
        target === "decision" ? "Choose Support Refs" : "Choose Linked Records",
      exact: true,
    });
    await choose.click();
    const picker = form.getByRole("region", {
      name:
        target === "decision" ? "Choose Support Refs" : "Choose Linked Records",
      exact: true,
    });
    await expect(
      picker.getByRole("button", { name: "Apply references", exact: true }),
    ).toBeEnabled();
    await expect(
      picker.locator(`input[type="checkbox"][value="${source.record_id}"]`),
    ).toBeChecked();
    await capture(page, `contextual-${target}-references-narrow`, {
      anchor: { ...anchor, locator: picker },
    });
    await picker
      .getByRole("button", { name: "Cancel references", exact: true })
      .click();
    if (target === "decision") {
      let accepted = false;
      await installOwnedRoute(page, `**/views/${view}/rows`, async (route) => {
        const response = await route.fetch();
        expect(response.ok()).toBeTruthy();
        accepted = true;
        await route.fulfill({ response });
      });
      await installOwnedRoute(page, `**/views/${view}/query`, async (route) => {
        if (accepted) await route.abort("failed");
        else await route.continue();
      });
      await page.getByTestId(genericCreateSubmitTestId(view)).click();
      await expect(form).toHaveCount(0);
    } else
      await form
        .getByRole("button", { name: "Keep draft and close", exact: true })
        .click();
    await openRecoveryItem(
      page,
      /^(Task Requests|Decisions) (draft|creation) ·/,
    );
    const recovery = page.getByRole("region", {
      name: "Retained contextual creation",
      exact: true,
    });
    await expect(
      recovery.getByRole("button", {
        name:
          target === "decision"
            ? "Retry creation refresh"
            : "Resume contextual draft",
        exact: true,
      }),
    ).toBeEnabled();
    await capture(page, `contextual-${target}-recovery-narrow`);
    await page.setViewportSize({ width: 1280, height: 720 });
    await capture(page, `contextual-${target}-recovery`);
    const supplied = await createRichTimelineSource(page, `${target}-source`, {
      client_txn_id: uniqueTxn(`${target}-source`),
      "timeline.activity_synopsis_text": retainedTimelineSourceLabel,
    });
    const retainedSource = await retainTimelineContextualSource(
      page,
      target,
      (url) => navigateVisualApplication(page, url),
      supplied,
    );
    for (const viewport of [
      { width: 1280, height: 720 },
      { width: 768, height: 640 },
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
      await test
        .info()
        .attach(
          `contextual-${target}-timeline-source-${viewport.width}-review`,
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

test("Capture contextual coordination target authoring source selection and retained recovery", async ({
  page,
}) => {
  const capture = async (
    name: string,
    options: { anchor?: VisualAnchor } = {},
  ) => {
    await assertViewportVisualRegression(page, name, {
      ...options,
      ready: () =>
        normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
          scroll: { top: 0, left: "left" },
        }),
    });
    await test.info().attach(`${name}-review`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  };
  for (const variant of ["comm_log", "handoff", "status_review", "lesson"]) {
    await page.setViewportSize({ width: 1280, height: 720 });
    const supplied = await createRichTimelineSource(page, `${variant}-source`, {
      client_txn_id: uniqueTxn(`${variant}-source`),
      "timeline.activity_synopsis_text": "Reviewed coordination source",
    });
    const f = await openCoordinationFixture(
      page,
      variant,
      timelineViewSchemaId,
      (url) => navigateVisualApplication(page, url),
      supplied,
    );
    await fillCoordinationMinimum(f);
    await f.form.locator("input, textarea, select, button").first().focus();
    const anchor: VisualAnchor = {
      locator: f.form,
      align: "start",
      scrollportSelector: `aside[data-view-schema-id="${timelineViewSchemaId}"] [data-inspector-scroll-body]`,
    };
    await capture(`coordination-${variant}-authoring`, { anchor });
    if (variant === "lesson") {
      await page.setViewportSize({ width: 768, height: 640 });
      await f.form
        .getByRole("button", { name: "Choose source", exact: true })
        .click();
      const picker = f.form.getByRole("region", {
        name: "Choose source",
        exact: true,
      });
      await expect(
        picker.getByRole("button", { name: "Apply references", exact: true }),
      ).toBeEnabled();
      await capture("coordination-source-narrow", {
        anchor: { ...anchor, locator: picker },
      });
      await picker
        .getByRole("button", { name: "Cancel references", exact: true })
        .click();
      const { recovery } = await retainCoordinationUncertainResult(page, f);
      await page.setViewportSize({ width: 390, height: 480 });
      await capture("coordination-recovery-narrow");
      await page.setViewportSize({ width: 1280, height: 720 });
      await capture("coordination-recovery");
      await test.info().attach("coordination-recovery-tree", {
        body: await recovery.ariaSnapshot(),
        contentType: "text/plain",
      });
    }
  }
});
