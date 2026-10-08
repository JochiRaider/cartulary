import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
} from "@cartulary/ui-contracts";
import {
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow, type SuppliedSourceFixture } from "./query";
import { openRecoveryItem } from "./recovery";
import {
  openGenericInspectorForRecord,
  openTimelineInspector,
} from "./rowMutations";

/** Fixture-only data through the same public creation routes used by analysts. */
export async function openContextualCreationFixture(
  page: Page,
  target: "task_request" | "decision",
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
  sourceContext: {
    viewSchemaId: typeof evidenceViewSchemaId | typeof timelineViewSchemaId;
    label: string;
  } = {
    viewSchemaId: evidenceViewSchemaId,
    label: "Reviewed Evidence source",
  },
  supplied?: SuppliedSourceFixture,
) {
  const incident =
    supplied?.incidentId ??
    (await createIncident(
      page,
      uniqueIncidentKey("CTD-PRESENTATION"),
      "Contextual creation review",
    ));
  const source =
    supplied?.source ??
    (await createViewRow(page, incident, sourceContext.viewSchemaId, {
      client_txn_id: uniqueTxn("source"),
      ...(sourceContext.viewSchemaId === evidenceViewSchemaId
        ? { "evidence.requested_at": "2025-03-01T10:00:00Z" }
        : {}),
      [sourceContext.viewSchemaId === timelineViewSchemaId
        ? "timeline.activity_synopsis_text"
        : "evidence.title"]: sourceContext.label,
    }));
  await navigate(
    `/?incident_id=${incident}&view_schema_id=${sourceContext.viewSchemaId}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(sourceContext.viewSchemaId)),
  ).toBeVisible();
  if (sourceContext.viewSchemaId === timelineViewSchemaId)
    await openTimelineInspector(page, source.record_id);
  else
    await openGenericInspectorForRecord(
      page,
      sourceContext.viewSchemaId,
      source.record_id,
    );
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        sourceContext.viewSchemaId,
        `create_related.${target}`,
      ),
    )
    .click();
  const view =
    target === "task_request"
      ? taskRequestsViewSchemaId
      : decisionsViewSchemaId;
  await expect(page.getByTestId(genericCreateSubmitTestId(view))).toBeVisible();
  if (target === "task_request") {
    await page
      .getByTestId(genericCreateFieldTestId("task.title"))
      .fill("Review related investigation");
    await page
      .getByTestId(genericCreateFieldTestId("task.task_kind"))
      .selectOption("follow_up");
  } else {
    await page
      .getByTestId(genericCreateFieldTestId("decision.summary"))
      .fill("Retain the reviewed context");
    await page
      .getByTestId(genericCreateFieldTestId("decision.decision_type"))
      .selectOption("containment");
    await page
      .getByTestId(genericCreateFieldTestId("decision.rationale"))
      .fill("The selected Evidence supports this decision.");
  }
  return { incident, source, view };
}

export const retainedTimelineSourceLabel =
  "Original Timeline activity: investigate the observed connection, preserve the analyst's contextual draft, and retain this readable source across reference removal and Recovery. " +
  "LongSourceContext".repeat(12);

/** Exercise retained source context through the existing owner and shell Recovery. */
export async function retainTimelineContextualSource(
  page: Page,
  target: "task_request" | "decision",
  navigate?: (url: string) => Promise<unknown>,
  supplied?: SuppliedSourceFixture,
) {
  const label = retainedTimelineSourceLabel;
  const fixture = await openContextualCreationFixture(
    page,
    target,
    navigate,
    {
      viewSchemaId: timelineViewSchemaId,
      label,
    },
    supplied,
  );
  let creations = 0;
  const observe = (request: import("@playwright/test").Request) => {
    if (
      request.method() === "POST" &&
      request.url().includes(`/views/${fixture.view}/rows`)
    )
      creations++;
  };
  page.on("request", observe);
  try {
    const form = page.getByRole("region", {
      name: target === "decision" ? "Create decision" : "Create task request",
      exact: true,
    });
    const fields =
      target === "decision"
        ? ["Support Refs", "Affected Records"]
        : ["Linked Records"];
    if (target === "decision") {
      await form
        .getByRole("button", { name: "Choose Affected Records", exact: true })
        .click();
      const picker = form.getByRole("region", {
        name: "Choose Affected Records",
        exact: true,
      });
      await picker
        .getByRole("group", { name: "Affected Records", exact: true })
        .locator(`input[value="${fixture.source.record_id}"]`)
        .check();
      await picker
        .getByRole("button", { name: "Apply references", exact: true })
        .click();
    }
    for (const field of fields) {
      const remove = form.getByRole("button", {
        name: `Remove ${field} ${label}`,
        exact: true,
      });
      await remove.focus();
      await remove.press("Enter");
      await expect(
        form.getByRole("button", { name: `Choose ${field}`, exact: true }),
      ).toBeFocused();
      await expect(
        form.getByRole("group", {
          name: `${field} selected references`,
          exact: true,
        }),
      ).toContainText("No references selected.");
    }
    const scalar = target === "decision" ? "decision.summary" : "task.title";
    await form
      .getByTestId(genericCreateFieldTestId(scalar))
      .fill("Exact retained scalar after link removal");
    await expect(form.getByText(/^Create in .* Source:/)).toContainText(label);
    const keep = form.getByRole("button", {
      name: "Keep draft and close",
      exact: true,
    });
    await keep.focus();
    await keep.press("Enter");
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .click();
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
    await expect(
      recovery.getByTestId(genericCreateFieldTestId(scalar)),
    ).toHaveValue("Exact retained scalar after link removal");
    await expect(recovery.getByText(/^Create in .* Source:/)).toContainText(
      label,
    );
    for (const field of fields)
      await expect(
        recovery.getByRole("group", {
          name: `${field} selected references`,
          exact: true,
        }),
      ).toContainText("No references selected.");
    expect(creations).toBe(0);
    return { ...fixture, recovery, label, scalar };
  } finally {
    page.off("request", observe);
  }
}
