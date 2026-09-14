import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  draftCellTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  rowCellTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { fetchRecordHistoryCount } from "./support/workbook/history";
import {
  commitOrdinary,
  ordinaryEvidenceView as evidence,
  fillOrdinaryField,
  openOrdinaryFixture,
  ordinaryField,
  retainOrdinaryUncertainty,
  switchOrdinarySheet,
} from "./support/workbook/ordinaryCreate";
import { queryViewRows, waitForViewRowByCell } from "./support/workbook/query";

test("Ordinary creation admits the fourteen-schema minimum matrix through production controls", async ({
  page,
  workerAdmin,
}) => {
  const { incident } = await openOrdinaryFixture(page);
  const minima = {
    hosts: { "host.aad_device_id": "4837b1bf-a528-4a2c-8b7e-42a159de3464" },
    identities: {
      "identity.sid": "S-1-5-21-123456789-123456789-123456789-1001",
    },
    parties: {
      "party.display_name": "Ordinary party",
      "party.party_kind": "person",
    },
    task_requests: {
      "task.title": "Ordinary task",
      "task.task_kind": "question",
    },
    decisions: {
      "decision.summary": "Ordinary decision",
      "decision.decision_type": "scope",
      "decision.rationale": "Ordinary rationale",
    },
    comm_log: {
      "comm_log.comm_type": "briefing",
      "comm_log.audience": "Operations",
      "comm_log.channel_or_meeting": "Bridge",
      "comm_log.summary": "Ordinary communication",
    },
    handoff: {
      "handoff.incoming_owner_user_id": workerAdmin.user_id,
      "handoff.current_state_summary": "Ordinary handoff",
    },
    status_review: { "status_review.current_state_summary": "Ordinary review" },
    lesson: { "lesson.summary": "Ordinary lesson" },
    findings: { "finding.statement": "Ordinary finding" },
    investigative_queries: {
      "investigative_query.platform": "SQL",
      "investigative_query.purpose": "Ordinary purpose",
      "investigative_query.query_text": "select\n  value",
    },
    forensic_keywords: {
      "forensic_keyword.pattern": "needle",
      "forensic_keyword.reason": "Ordinary reason",
    },
    evidence: { "evidence.title": "Ordinary evidence" },
    indicators: {
      "indicator.indicator_type": "domain_name",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "EXAMPLE.TEST",
    },
  };
  for (const [name, values] of Object.entries(minima))
    await test.step(name, async () => {
      const view = `cartulary.view.${name}.v1`;
      await switchOrdinarySheet(page, view);
      await commitOrdinary(page, view);
      await expect(
        page.getByRole("region", { name: "Row creation" }).getByRole("alert"),
      ).toBeVisible();
      expect(await queryViewRows(page, incident, view)).toHaveLength(0);
      for (const [field, value] of Object.entries(values))
        await fillOrdinaryField(page, view, field, value);
      await commitOrdinary(page, view);
      await expect(
        page.getByRole("region", { name: "Row creation" }),
      ).toContainText("Row accepted.");
      const rows = await queryViewRows(page, incident, view);
      expect(rows).toHaveLength(1);
      const row = rows[0];
      if (!row) throw new Error("Missing created row");
      expect(row.row_version).toBe(1);
      if (name === "task_requests")
        expect(row.cells["task.status"]?.value).toBe("open");
      if (name === "decisions")
        expect(row.cells["decision.status"]?.value).toBe("proposed");
      if (name === "evidence")
        expect(row.cells["evidence.requested_at"]?.value).not.toBeNull();
      if (name === "forensic_keywords")
        expect(row.cells["forensic_keyword.case_sensitive"]?.value).toBe(false);
      expect(await fetchRecordHistoryCount(page, row.record_id)).toBe(1);
    });
});

test("Ordinary grid and inspector share retained authoring and admit same-frame Commit once", async ({
  page,
}) => {
  const view = "cartulary.view.parties.v1",
    { incident } = await openOrdinaryFixture(page, view);
  await fillOrdinaryField(
    page,
    view,
    "party.display_name",
    "  Retained party  ",
  );
  await fillOrdinaryField(page, view, "party.party_kind", "person");
  await fillOrdinaryField(
    page,
    view,
    "party.notes",
    "  exact hidden notes\nnext line  ",
  );
  await page.getByTestId(workbookInspectorToggleTestId(view)).click();
  await switchOrdinarySheet(page, timelineViewSchemaId);
  await switchOrdinarySheet(page, view);
  expect(
    await (await ordinaryField(page, view, "party.display_name")).inputValue(),
  ).toBe("  Retained party  ");
  expect(
    await (await ordinaryField(page, view, "party.notes")).inputValue(),
  ).toBe("  exact hidden notes\nnext line  ");
  const bodies: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith(`/views/${view}/rows`)
    )
      bodies.push(request.postData() ?? "");
  });
  const inspectorCommit = page.getByRole("button", {
    name: "Commit draft row",
    exact: true,
  });
  const gridCommitId = genericCreateSubmitTestId(view);
  await scrollGridTargetIntoView({
    page,
    surface: view,
    targetTestId: gridCommitId,
  });
  const gridCommit = await page.getByTestId(gridCommitId).elementHandle();
  if (!gridCommit) throw new Error("Missing grid Commit");
  await inspectorCommit.evaluate((button, gridButton) => {
    (button as HTMLButtonElement).click();
    (gridButton as HTMLButtonElement).click();
  }, gridCommit);
  const row = await waitForViewRowByCell(
    page,
    incident,
    view,
    "party.display_name",
    "Retained party",
  );
  await expect(
    page.getByRole("region", { name: "Row creation" }),
  ).toContainText("Row accepted.");
  expect(bodies).toHaveLength(1);
  expect(row.cells["party.notes"]?.value).toBe("exact hidden notes\nnext line");
});

test("Ordinary uncertain recovery preserves next authoring and leaves Timeline capture available", async ({
  page,
}) => {
  for (const malformed of [false, true]) {
    const { incident } = await openOrdinaryFixture(page);
    await fillOrdinaryField(
      page,
      evidence,
      "evidence.title",
      "Original evidence",
    );
    const { bodies } = await retainOrdinaryUncertainty(
      page,
      incident,
      evidence,
      malformed,
    );
    await fillOrdinaryField(
      page,
      evidence,
      "evidence.title",
      "  Next unsent evidence  ",
    );
    await switchOrdinarySheet(page, timelineViewSchemaId);
    const synopsis = "timeline.activity_synopsis_text";
    const draft = draftCellTestId(synopsis);
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: draft,
    });
    await page.getByTestId(draft).click();
    await page.keyboard.type("Capture during ordinary recovery");
    await page.keyboard.press("Enter");
    const timeline = await waitForViewRowByCell(
      page,
      incident,
      timelineViewSchemaId,
      synopsis,
      "Capture during ordinary recovery",
    );
    await expect(
      page.getByTestId(rowCellTestId(timeline.record_id, synopsis)),
    ).toContainText("Capture during ordinary recovery");
    await switchOrdinarySheet(page, evidence);
    const input = await ordinaryField(page, evidence, "evidence.title");
    await expect(input).toHaveValue("  Next unsent evidence  ");
    const recovery = page.getByRole("button", {
      name: "Recover submission",
      exact: true,
    });
    await recovery.focus();
    await recovery.press("Enter");
    await expect(
      page.getByRole("region", { name: "Row creation" }),
    ).toContainText("Row accepted.");
    expect(bodies).toHaveLength(2);
    expect(bodies[1]).toBe(bodies[0]);
    await expect(input).toHaveValue("  Next unsent evidence  ");
    const rows = await queryViewRows(page, incident, evidence);
    expect(rows).toHaveLength(1);
    expect(await fetchRecordHistoryCount(page, rows[0]?.record_id ?? "")).toBe(
      1,
    );
  }
});

test("Ordinary accepted results survive detached completion and recover failed refresh by reads", async ({
  page,
}) => {
  const { incident } = await openOrdinaryFixture(page);
  await fillOrdinaryField(
    page,
    evidence,
    "evidence.title",
    "Accepted while detached",
  );
  let committed = false,
    count = 0,
    release = () => {};
  const delay = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route(
    `**/incidents/${incident}/views/${evidence}/rows`,
    async (route) => {
      count++;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      await delay;
      await route.fulfill({ response });
    },
  );
  await commitOrdinary(page, evidence);
  await expect.poll(() => committed).toBe(true);
  await fillOrdinaryField(
    page,
    evidence,
    "evidence.title",
    "Next draft survives refresh",
  );
  await switchOrdinarySheet(page, timelineViewSchemaId);
  release();
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  const queryPattern = `**/incidents/${incident}/views/${evidence}/query`;
  await page.route(queryPattern, (route) =>
    route.fulfill({
      status: 500,
      contentType: "application/json",
      body: JSON.stringify({
        error: { code: "projection_failed", request_id: "failed-read" },
      }),
    }),
  );
  await switchOrdinarySheet(page, evidence);
  const refresh = page.getByRole("button", {
    name: "Refresh accepted result",
    exact: true,
  });
  await expect(refresh).toBeEnabled();
  await expect(
    page.getByRole("region", { name: "Row creation" }),
  ).toContainText("Row accepted.");
  await page.unroute(queryPattern);
  await refresh.focus();
  await refresh.press("Enter");
  await expect(
    page.getByTestId(genericCreateFieldTestId("evidence.title")),
  ).toHaveValue("Next draft survives refresh");
  expect(count).toBe(1);
  expect(await queryViewRows(page, incident, evidence)).toHaveLength(1);
});
