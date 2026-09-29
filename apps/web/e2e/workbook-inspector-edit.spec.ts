// biome-ignore lint/correctness/noUndeclaredDependencies: the workspace root pins the browser accessibility engine used by the Make-owned harness.
import AxeBuilder from "@axe-core/playwright";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  cartularyDesignPresentation,
  dataTestIdSelector,
  entityInspectorTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  timelineInspectorTestId,
  timelineScalarEditorTestId,
  workbookFocusAnchorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorPanelTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  type InspectorPanelId,
  indicatorsViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createInspectorReadingFixture } from "./support/timeline/inspectorReadingFixture";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { fetchRecordHistoryCount } from "./support/workbook/history";
import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import {
  activateCommittedGridCell,
  openGenericInspectorForRecord,
  openTimelineInspector,
} from "./support/workbook/rowMutations";

async function fixture(page: Page, view: string = hostsViewSchemaId) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("IER"),
    "Inspector editing recovery",
  );
  const field =
    view === hostsViewSchemaId ? "host.display_name" : "evidence.title";
  const first = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("inspector-a"),
    [field]: "Inspected A",
  });
  const second = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("inspector-b"),
    [field]: "Inspected B",
  });
  await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
  await openGenericInspectorForRecord(page, view, first.record_id);
  return { incident, view, field, first, second };
}
async function editField(page: Page, view: string, field: string) {
  await page.locator(`[data-inspector-edit-field="${field}"]`).click();
  return page.getByTestId(genericEditValueTestId(view));
}

async function expectReferenceDefinitionList(
  page: Page,
  view: string,
  panel: InspectorPanelId,
) {
  const result = await new AxeBuilder({ page })
    .include(dataTestIdSelector(workbookInspectorPanelTestId(view, panel)))
    .withRules(["definition-list"])
    .analyze();
  expect(result.violations).toEqual([]);
  expect(
    result.incomplete.filter((item) => item.id === "definition-list"),
  ).toEqual([]);
}

async function expectReachableReferenceAction(page: Page, actionName: string) {
  const action = page.getByRole("button", { name: actionName, exact: true });
  await action.scrollIntoViewIfNeeded();
  await expect(action).toBeVisible();
  const geometry = await action.evaluate((element) => {
    const bounds = element.getBoundingClientRect();
    return {
      height: bounds.height,
      width: bounds.width,
      clippedText: element.scrollWidth > element.clientWidth + 1,
      clippedViewport:
        bounds.left < 0 ||
        bounds.right > window.innerWidth ||
        bounds.top < 0 ||
        bounds.bottom > window.innerHeight,
    };
  });
  expect(geometry.height).toBeGreaterThanOrEqual(28);
  expect(geometry.width).toBeGreaterThanOrEqual(28);
  expect(geometry.clippedText).toBe(false);
  expect(geometry.clippedViewport).toBe(false);
  return action;
}

async function placeInspectorActionNearBodyBottom(action: Locator) {
  await action.evaluate((element) => {
    const body = element.closest<HTMLElement>("[data-inspector-scroll-body]");
    if (!body) throw new Error("Missing inspector scroll body");
    const scale = body.getBoundingClientRect().height / body.offsetHeight;
    body.scrollTop +=
      (element.getBoundingClientRect().bottom -
        (body.getBoundingClientRect().bottom - 12)) /
      scale;
  });
}

async function outsideInspectorScroll(page: Page) {
  return page.evaluate((selector) => {
    const grid = document.querySelector<HTMLElement>(selector);
    return {
      documentX: window.scrollX,
      documentY: window.scrollY,
      gridTop: grid?.scrollTop ?? null,
      gridLeft: grid?.scrollLeft ?? null,
    };
  }, gridScrollportSelector());
}

async function inspectorFocusedControlGeometry(control: Locator) {
  return control.evaluate((element) => {
    if (!(element instanceof HTMLElement))
      throw new Error("Expected an HTML inspector control");
    const body = element.closest<HTMLElement>("[data-inspector-scroll-body]");
    const field = element.closest<HTMLElement>("[data-inspector-saved-field]");
    if (!body || !field)
      throw new Error("Missing inspector field or scroll body");
    const bodyBox = body.getBoundingClientRect();
    const scale = bodyBox.height / body.offsetHeight;
    const visual = window.visualViewport;
    const bounds = {
      left: Math.max(
        visual?.offsetLeft ?? 0,
        bodyBox.left + body.clientLeft * scale,
      ),
      top: Math.max(
        visual?.offsetTop ?? 0,
        bodyBox.top + body.clientTop * scale,
      ),
      right: Math.min(
        (visual?.offsetLeft ?? 0) + (visual?.width ?? window.innerWidth),
        bodyBox.left + (body.clientLeft + body.clientWidth) * scale,
      ),
      bottom: Math.min(
        (visual?.offsetTop ?? 0) + (visual?.height ?? window.innerHeight),
        bodyBox.top + (body.clientTop + body.clientHeight) * scale,
      ),
    };
    for (
      let ancestor = body.parentElement;
      ancestor;
      ancestor = ancestor.parentElement
    ) {
      const style = getComputedStyle(ancestor);
      const clipsX = ["auto", "scroll", "hidden", "clip"].includes(
        style.overflowX,
      );
      const clipsY = ["auto", "scroll", "hidden", "clip"].includes(
        style.overflowY,
      );
      if (clipsX || clipsY) {
        const clip = ancestor.getBoundingClientRect();
        if (clipsX) {
          bounds.left = Math.max(bounds.left, clip.left);
          bounds.right = Math.min(bounds.right, clip.right);
        }
        if (clipsY) {
          bounds.top = Math.max(bounds.top, clip.top);
          bounds.bottom = Math.min(bounds.bottom, clip.bottom);
        }
      }
    }
    const rect = element.getBoundingClientRect();
    const style = getComputedStyle(element);
    const targetScale =
      element.offsetWidth > 0 ? rect.width / element.offsetWidth : 1;
    const ring =
      ((Number.parseFloat(style.outlineWidth) || 0) +
        (Number.parseFloat(style.outlineOffset) || 0)) *
      targetScale;
    const fieldLabel = field.querySelector("dt")?.getBoundingClientRect();
    const savedValue = field
      .querySelector("[data-inspector-field-value]")
      ?.getBoundingClientRect();
    const contextTop = Math.min(
      fieldLabel?.top ?? rect.top,
      savedValue?.top ?? rect.top,
      rect.top,
    );
    const contextBottom = Math.max(
      fieldLabel?.bottom ?? rect.bottom,
      savedValue?.bottom ?? rect.bottom,
      rect.bottom,
    );
    return {
      focused: document.activeElement === element,
      fits: rect.height + ring * 2 <= bounds.bottom - bounds.top,
      contained:
        rect.top - ring >= bounds.top - 1 &&
        rect.bottom + ring <= bounds.bottom + 1 &&
        rect.left - ring >= bounds.left - 1 &&
        rect.right + ring <= bounds.right + 1,
      contextFits: contextBottom - contextTop <= bounds.bottom - bounds.top,
      contextContained:
        contextTop >= bounds.top - 1 && contextBottom <= bounds.bottom + 1,
    };
  });
}

test("a11y.inspector Timeline lower-edge Edit and retained Resume reveal the focused control", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1024, height: 720 });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("IERV"),
    "Inspector editor reveal",
  );
  const row = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("ierv-timeline"),
    "timeline.activity_synopsis_text": "Inspector editor reveal target",
    "timeline.data_source_text": "Accepted source context",
  });
  const writes: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      new URL(request.url()).pathname.endsWith(`/records/${row.record_id}`)
    )
      writes.push(request.url());
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${timelineViewSchemaId}`,
  );
  await openTimelineInspector(page, row.record_id);
  const edit = page
    .getByTestId(timelineInspectorTestId())
    .getByRole("button", { name: "Edit Data Source", exact: true });
  const control = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.data_source_text",
      recordId: row.record_id,
      surface: "inspector",
    }),
  );
  await placeInspectorActionNearBodyBottom(edit);
  const outsideScroll = await outsideInspectorScroll(page);
  await edit.click();
  await expect(control).toBeFocused();
  const pointer = await inspectorFocusedControlGeometry(control);
  expect.soft(pointer.fits && pointer.contained).toBe(true);
  expect.soft(!pointer.contextFits || pointer.contextContained).toBe(true);
  expect(await outsideInspectorScroll(page)).toEqual(outsideScroll);
  await test.info().attach("inspector-timeline-edit-1024", {
    body: await page.screenshot({ animations: "disabled", caret: "hide" }),
    contentType: "image/png",
  });

  await control.press("Escape");
  await expect(edit).toBeFocused();
  await placeInspectorActionNearBodyBottom(edit);
  await edit.press("Enter");
  await expect(control).toBeFocused();
  const keyboard = await inspectorFocusedControlGeometry(control);
  expect.soft(keyboard.fits && keyboard.contained).toBe(true);
  expect.soft(!keyboard.contextFits || keyboard.contextContained).toBe(true);

  const raw = "  exact local source\n  ";
  await control.fill(raw);
  await page.getByRole("button", { name: "Close editor", exact: true }).click();
  await openTimelineInspector(page, row.record_id);
  const resume = page.getByRole("button", {
    name: "Resume draft for Data Source",
    exact: true,
  });
  await placeInspectorActionNearBodyBottom(resume);
  await resume.click();
  await expect(control).toHaveValue(raw);
  await expect(control).toBeFocused();
  const resumed = await inspectorFocusedControlGeometry(control);
  expect.soft(resumed.fits && resumed.contained).toBe(true);
  expect.soft(!resumed.contextFits || resumed.contextContained).toBe(true);
  expect(await outsideInspectorScroll(page)).toEqual(outsideScroll);
  await test.info().attach("inspector-timeline-resume-1024", {
    body: await page.screenshot({ animations: "disabled", caret: "hide" }),
    contentType: "image/png",
  });
  expect(writes).toHaveLength(0);
});

test("a11y.inspector Evidence reference shortcut reveals saved field context on repeated activation", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("IERR"),
    "Inspector reference reveal",
  );
  const party = await createViewRow(page, incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("ierr-party"),
    "party.display_name": "Collector for reveal",
    "party.party_kind": "team",
  });
  const evidence = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("ierr-evidence"),
    "evidence.title": "Reference reveal target",
    "evidence.collector_party_id": party.record_id,
  });
  const writes: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      new URL(request.url()).pathname.endsWith(`/records/${evidence.record_id}`)
    )
      writes.push(request.url());
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${evidenceViewSchemaId}`,
  );
  await openGenericInspectorForRecord(
    page,
    evidenceViewSchemaId,
    evidence.record_id,
  );
  const inspector = page.getByTestId(
    workbookInspectorPanelTestId(evidenceViewSchemaId, "relationships"),
  );
  const shortcut = inspector.getByRole("button", {
    name: "Edit Collector Party",
    exact: true,
  });
  const control = page.getByTestId(
    genericEditValueTestId(evidenceViewSchemaId),
  );
  await shortcut.focus();
  await shortcut.press("Enter");
  await expect(control).toBeFocused();
  const first = await inspectorFocusedControlGeometry(control);
  expect.soft(first.fits && first.contained).toBe(true);
  expect.soft(!first.contextFits || first.contextContained).toBe(true);
  await shortcut.click();
  await expect(control).toBeFocused();
  const repeated = await inspectorFocusedControlGeometry(control);
  expect.soft(repeated.fits && repeated.contained).toBe(true);
  expect.soft(!repeated.contextFits || repeated.contextContained).toBe(true);
  expect(writes).toHaveLength(0);
});

test("a11y.inspector Entity edit keeps focus and context visible across field changes and enlarged layouts", async ({
  page,
}) => {
  const f = await fixture(page);
  const inspector = page.getByTestId(entityInspectorTestId("host"));
  const input = page.getByTestId(genericEditValueTestId(f.view));
  const location = inspector.getByRole("button", {
    name: "Edit Location",
    exact: true,
  });
  let writes = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      new URL(request.url()).pathname.endsWith(`/records/${f.first.record_id}`)
    )
      writes++;
  });
  const raw = "  unchanged local location  ";
  await location.click();
  await expect(input).toBeFocused();
  await input.fill(raw);
  await inspector
    .getByRole("button", { name: "Edit Business Owner", exact: true })
    .click();
  await expect(input).toBeFocused();
  expect((await inspectorFocusedControlGeometry(input)).contained).toBe(true);
  await input.press("Escape");
  await expect(
    inspector.getByRole("button", {
      name: "Edit Business Owner",
      exact: true,
    }),
  ).toBeFocused();
  const resume = inspector.getByRole("button", {
    name: "Resume draft for Location",
    exact: true,
  });
  await resume.click();
  await expect(input).toHaveValue(raw);
  await expect(input).toBeFocused();
  expect((await inspectorFocusedControlGeometry(input)).contained).toBe(true);
  await input.press("Escape");
  await expect(resume).toBeFocused();

  for (const layout of [
    { width: 320, zoom: 1, spacing: false },
    { width: 1280, zoom: 2, spacing: false },
    { width: 1024, zoom: 1, spacing: true },
  ]) {
    await page.setViewportSize({ width: layout.width, height: 720 });
    await page.evaluate(({ zoom, spacing }) => {
      document.documentElement.style.zoom = String(zoom);
      document.body.style.lineHeight = spacing ? "1.5" : "";
      document.body.style.letterSpacing = spacing ? "0.12em" : "";
      document.body.style.wordSpacing = spacing ? "0.16em" : "";
    }, layout);
    await resume.click();
    await expect(input).toHaveValue(raw);
    await expect(input).toBeFocused();
    const geometry = await inspectorFocusedControlGeometry(input);
    expect
      .soft(geometry.fits && geometry.contained, JSON.stringify(layout))
      .toBe(true);
    expect
      .soft(
        !geometry.contextFits || geometry.contextContained,
        JSON.stringify(layout),
      )
      .toBe(true);
    if (layout.zoom === 2) {
      await page.setViewportSize({ width: 1100, height: 720 });
      await expect(input).toBeFocused();
      await expect(async () => {
        const resized = await inspectorFocusedControlGeometry(input);
        expect(resized.fits && resized.contained).toBe(true);
      }).toPass();
    }
    const clear = inspector.getByRole("button", {
      name: "Clear Location",
      exact: true,
    });
    await input.press("Tab");
    await expect(clear).toBeFocused();
    expect
      .soft((await inspectorFocusedControlGeometry(clear)).contained)
      .toBe(true);
    const update = inspector.getByTestId(genericEditSubmitTestId(f.view));
    await clear.press("Tab");
    await expect(update).toBeFocused();
    expect
      .soft((await inspectorFocusedControlGeometry(update)).contained)
      .toBe(true);
    const close = inspector.getByRole("button", {
      name: "Close editor",
      exact: true,
    });
    await update.press("Tab");
    await expect(close).toBeFocused();
    expect
      .soft((await inspectorFocusedControlGeometry(close)).contained)
      .toBe(true);
    await close.press("Escape");
    await expect(resume).toBeFocused();
  }
  await page.evaluate(() => {
    document.documentElement.style.zoom = "";
    document.body.style.lineHeight = "";
    document.body.style.letterSpacing = "";
    document.body.style.wordSpacing = "";
  });
  await page.setViewportSize({ width: 1440, height: 900 });
  await resume.click();
  await expect(input).toBeFocused();
  const sections = inspector.getByRole("button", { name: /^Sections:/ });
  if (await sections.isVisible()) await sections.click();
  await inspector.getByRole("button", { name: "History", exact: true }).click();
  const historyAction = inspector.getByRole("button", {
    name: "Open history",
    exact: true,
  });
  await expect(historyAction).toBeFocused();
  await page.setViewportSize({ width: 1280, height: 720 });
  await expect(historyAction).toBeFocused();
  await openGenericInspectorForRecord(page, f.view, f.second.record_id);
  await expect(inspector).toContainText("Inspected B");
  await expect(input).toHaveCount(0);
  expect(writes).toBe(0);
  await openGenericInspectorForRecord(page, f.view, f.first.record_id);
  if (await sections.isVisible()) await sections.click();
  await inspector.getByRole("button", { name: "Details", exact: true }).click();
  await patchRecord(page, f.first.record_id, {
    view_schema_id: f.view,
    base_row_version: 1,
    client_txn_id: uniqueTxn("inspector-reveal-review"),
    changes: [{ field_key: "host.location", value: "New accepted location" }],
  });
  const review = inspector.getByRole("button", {
    name: "Review draft for Location",
    exact: true,
  });
  await review.click();
  const useSaved = inspector.getByRole("button", {
    name: "Use saved Location",
    exact: true,
  });
  await expect(useSaved).toBeFocused();
  expect((await inspectorFocusedControlGeometry(useSaved)).contained).toBe(
    true,
  );
  const keepDraft = inspector.getByRole("button", {
    name: "Keep draft Location",
    exact: true,
  });
  await useSaved.press("Tab");
  await expect(keepDraft).toBeFocused();
  expect((await inspectorFocusedControlGeometry(keepDraft)).contained).toBe(
    true,
  );
  const writesBeforeReviewDecision = writes;
  await keepDraft.click();
  await expect(input).toHaveValue(raw);
  await expect(input).toBeFocused();
  expect((await inspectorFocusedControlGeometry(input)).contained).toBe(true);
  expect(writes).toBe(writesBeforeReviewDecision);
});

test("a11y.generic reference summaries keep Party and Evidence shortcuts semantic and draft-safe", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("IRS"),
    "Inspector reference summaries",
  );
  const party = await createViewRow(page, incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("irs-party"),
    "party.display_name": "Reference summary collector",
    "party.party_kind": "team",
  });
  const evidenceTitle = `Reference summary evidence ${"long accepted title ".repeat(8)}`;
  const evidence = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("irs-evidence"),
    "evidence.title": evidenceTitle,
    "evidence.collector_party_id": party.record_id,
  });
  const lesson = await createViewRow(page, incident, lessonViewSchemaId, {
    client_txn_id: uniqueTxn("irs-lesson"),
    "lesson.summary": "Reference summary lesson",
  });
  await patchRecord(page, lesson.record_id, {
    view_schema_id: lessonViewSchemaId,
    base_row_version: lesson.row_version,
    client_txn_id: uniqueTxn("irs-lesson-evidence"),
    changes: [
      {
        field_key: "lesson.evidence_refs",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            { op: "add_record_ref", linked_record_id: evidence.record_id },
          ],
        },
      },
    ],
  });
  const writes: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      [evidence.record_id, lesson.record_id].some((recordId) =>
        new URL(request.url()).pathname.endsWith(`/records/${recordId}`),
      )
    )
      writes.push(request.url());
  });

  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${evidenceViewSchemaId}`,
  );
  await openGenericInspectorForRecord(
    page,
    evidenceViewSchemaId,
    evidence.record_id,
  );
  const relationships = page.getByTestId(
    workbookInspectorPanelTestId(evidenceViewSchemaId, "relationships"),
  );
  await expect(relationships.getByRole("term")).toHaveText([
    "Collector Party",
    "Source Party",
  ]);
  await expect(relationships).toContainText(party.record_id);
  await expectReferenceDefinitionList(
    page,
    evidenceViewSchemaId,
    "relationships",
  );
  const edit = relationships.getByRole("button", {
    name: "Edit Source Party",
    exact: true,
  });
  await edit.focus();
  await edit.press("Enter");
  const details = page.getByTestId(
    workbookInspectorPanelTestId(evidenceViewSchemaId, "details"),
  );
  await expect(
    details.getByRole("group", { name: "Unsaved change: Source Party" }),
  ).toBeVisible();
  const input = page.getByTestId(genericEditValueTestId(evidenceViewSchemaId));
  await expect(input).toBeFocused();
  expect(writes).toHaveLength(0);
  await input.fill(party.record_id);
  await details.getByRole("button", { name: "Close editor" }).click();
  const resume = details.getByRole("button", {
    name: "Resume draft for Source Party",
  });
  await expect(resume).toBeVisible();
  await resume.click();
  await expect(input).toHaveValue(party.record_id);
  await input.press("Escape");
  await expect(resume).toBeVisible();
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(evidenceViewSchemaId))
    .click();
  await openGenericInspectorForRecord(
    page,
    evidenceViewSchemaId,
    evidence.record_id,
  );
  await expect(resume).toBeVisible();
  expect(writes).toHaveLength(0);

  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${lessonViewSchemaId}`,
  );
  await openGenericInspectorForRecord(
    page,
    lessonViewSchemaId,
    lesson.record_id,
  );
  const evidencePanel = page.getByTestId(
    workbookInspectorPanelTestId(lessonViewSchemaId, "evidence"),
  );
  await expect(evidencePanel.getByRole("term")).toHaveText(["Evidence"]);
  await expect(evidencePanel).toContainText(evidenceTitle);
  await expectReferenceDefinitionList(page, lessonViewSchemaId, "evidence");
  const manage = evidencePanel.getByRole("button", {
    name: "Manage Evidence",
    exact: true,
  });
  await manage.focus();
  await manage.press("Space");
  await expect(
    page
      .getByTestId(workbookInspectorPanelTestId(lessonViewSchemaId, "details"))
      .getByRole("group", { name: "Unsaved change: Evidence" }),
  ).toBeVisible();
  await expect(
    page.getByTestId(genericEditValueTestId(lessonViewSchemaId)),
  ).toBeFocused();
  expect(writes).toHaveLength(0);
  await page.getByRole("button", { name: "Close editor" }).click();

  for (const [width, zoom] of [
    [1024, 1],
    [320, 1],
    [1280, 2],
  ] as const) {
    await page.setViewportSize({ width, height: 720 });
    await page.evaluate((nextZoom) => {
      document.documentElement.style.zoom = String(nextZoom);
      document.body.style.lineHeight = "1.5";
      document.body.style.letterSpacing = "0.12em";
      document.body.style.wordSpacing = "0.16em";
    }, zoom);
    await expectReachableReferenceAction(page, "Manage Evidence");
    await expect(evidencePanel).toContainText(evidenceTitle);
  }
  expect(writes).toHaveLength(0);
});

test("a11y.generic reference summary viewer actions remain disabled", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("IRSV"),
    "Inspector reference viewer",
  );
  const evidence = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("irsv-evidence"),
    "evidence.title": "Viewer reference evidence",
  });
  const password = "InspectorReferenceViewer1!";
  const viewer = await createIncidentMemberUser(page, incident, {
    email: uniqueEmail("inspector-reference-viewer"),
    display_name: "Inspector reference viewer",
    initial_password: password,
    role: "viewer",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const context = await browser.newContext();
  try {
    const viewerPage = await context.newPage();
    await sessionTracker.loginTrackedUser(viewerPage, {
      createdBy: "a11y.generic-reference-summary",
      email: viewer.email,
      password,
      purpose: "reference summary viewer read-only state",
      userId: viewer.user_id,
    });
    await viewerPage.setViewportSize({ width: 1024, height: 720 });
    await viewerPage.goto(
      `/?incident_id=${incident}&view_schema_id=${evidenceViewSchemaId}`,
    );
    await openGenericInspectorForRecord(
      viewerPage,
      evidenceViewSchemaId,
      evidence.record_id,
    );
    const relationships = viewerPage.getByTestId(
      workbookInspectorPanelTestId(evidenceViewSchemaId, "relationships"),
    );
    await expectReferenceDefinitionList(
      viewerPage,
      evidenceViewSchemaId,
      "relationships",
    );
    const edit = relationships.getByRole("button", {
      name: "Edit Collector Party",
      exact: true,
    });
    await expect(edit).toBeDisabled();
    await expect(relationships.getByRole("term")).toHaveText([
      "Collector Party",
      "Source Party",
    ]);
    await expect(
      viewerPage.getByTestId(genericEditValueTestId(evidenceViewSchemaId)),
    ).toHaveCount(0);
  } finally {
    await context.close();
  }
});

test("Inspector edits bind the selected record and retain dirty fields through saved changes and explicit return", async ({
  page,
}) => {
  const f = await fixture(page),
    input = await editField(page, f.view, "host.location");
  await expect(page.getByRole("combobox", { name: "Edit record" })).toHaveCount(
    0,
  );
  await expect(
    page.locator('[data-inspector-edit-field="host.fqdn"]'),
  ).toHaveCount(0);
  await input.fill("  unfinished location  ");
  await page.getByRole("button", { name: "Close editor", exact: true }).click();
  await expect(input).toHaveCount(0);
  await page
    .getByRole("button", { name: "Resume draft for Location", exact: true })
    .click();
  await expect(input).toHaveValue("  unfinished location  ");
  const inspector = page.getByTestId(entityInspectorTestId("host"));
  let historyReads = 0;
  let inspectorWrites = 0;
  page.on("request", (request) => {
    if (request.url().includes(`/records/${f.first.record_id}/history`))
      historyReads++;
    if (
      request.method() === "PATCH" &&
      new URL(request.url()).pathname.endsWith(`/records/${f.first.record_id}`)
    )
      inspectorWrites++;
  });
  for (const width of [1440, 760, 320]) {
    await page.setViewportSize({ width, height: 900 });
    const chooser = inspector.getByRole("button", {
      name: /^Sections:/,
      exact: true,
    });
    await expect(async () => {
      if (
        (await chooser.isVisible()) &&
        (await chooser.getAttribute("aria-expanded")) === "false"
      )
        await chooser.click();
      await inspector
        .getByRole("button", { name: "History", exact: true })
        .click({ timeout: 1000 });
    }).toPass({ timeout: 5000 });
    await expect(
      inspector.getByRole("button", { name: "Open history", exact: true }),
    ).toBeFocused();
    if (await chooser.isVisible()) {
      await expect(chooser).toContainText("History");
    } else {
      await expect(
        inspector.locator('[data-inspector-navigation-panel="history"]'),
      ).toHaveAttribute("aria-current", "location");
    }
    await expect(
      inspector.getByRole("button", { name: "Close inspector" }),
    ).toBeInViewport();
    await expect(input).toHaveValue("  unfinished location  ");
    if (await chooser.isVisible()) {
      await chooser.click();
      await expect(
        inspector.getByRole("button", { name: "History", exact: true }),
      ).toHaveAttribute("aria-current", "location");
      await inspector
        .getByRole("button", { name: "Details", exact: true })
        .focus();
      await page.keyboard.press("Escape");
      await expect(chooser).toBeFocused();
      await expect(inspector).toBeVisible();
      await chooser.click();
    }
    await inspector
      .getByRole("button", { name: "Details", exact: true })
      .click();
    await expect(input).toHaveValue("  unfinished location  ");
    await test.info().attach(`inspector-navigation-${width}`, {
      body: await page.screenshot({ animations: "disabled" }),
      contentType: "image/png",
    });
  }
  expect(historyReads).toBe(0);
  expect(inspectorWrites).toBe(0);
  await page.setViewportSize({ width: 1440, height: 900 });
  await patchRecord(page, f.first.record_id, {
    view_schema_id: f.view,
    base_row_version: 1,
    client_txn_id: uniqueTxn("unrelated"),
    changes: [{ field_key: "host.display_name", value: "Inspected A renamed" }],
  });
  await expect(page.getByTestId(entityInspectorTestId("host"))).toContainText(
    "Inspected A renamed",
  );
  await expect(input).toHaveValue("  unfinished location  ");
  await expect(page.getByTestId(genericEditSubmitTestId(f.view))).toBeEnabled();
  await editField(page, f.view, "host.business_owner");
  await editField(page, f.view, "host.location");
  await expect(page.getByTestId(genericEditSubmitTestId(f.view))).toBeEnabled();
  await openGenericInspectorForRecord(page, f.view, f.second.record_id);
  await editField(page, f.view, "host.location");
  await expect(input).toHaveValue("");
  await openGenericInspectorForRecord(page, f.view, f.first.record_id);
  await editField(page, f.view, "host.location");
  await patchRecord(page, f.first.record_id, {
    view_schema_id: f.view,
    base_row_version: 2,
    client_txn_id: uniqueTxn("same-field"),
    changes: [{ field_key: "host.location", value: "Concurrent location" }],
  });
  await expect(
    page.getByRole("button", { name: "Keep draft Location", exact: true }),
  ).toBeVisible();
  await expect(input).toHaveValue("  unfinished location  ");
  await page
    .getByRole("button", { name: "Keep draft Location", exact: true })
    .click();
  const sent = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${f.first.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(f.view)).click();
  expect((await sent).ok()).toBeTruthy();
  await expect(
    page.getByRole("region", { name: "Inspector changes" }),
  ).toContainText("Saved, version 4");
  const rows = await queryViewRows(page, f.incident, f.view);
  expect(
    rows.find((row) => row.record_id === f.first.record_id)?.cells[
      "host.location"
    ]?.value,
  ).toBe("unfinished location");
  expect(
    rows.find((row) => row.record_id === f.second.record_id)?.row_version,
  ).toBe(1);
  const historySections = inspector.getByRole("button", { name: /^Sections:/ });
  if (await historySections.isVisible()) await historySections.click();
  await inspector.getByRole("button", { name: "History", exact: true }).click();
  await inspector
    .getByRole("button", { name: "Open history", exact: true })
    .click();
  await expect(
    inspector.getByText("Event details", { exact: true }).first(),
  ).toBeVisible();
  for (const summary of await inspector
    .getByText("Event details", { exact: true })
    .all())
    await summary.click();
  const comparison = inspector
    .locator("[data-history-comparison]")
    .filter({ hasText: "Inspected A renamed" })
    .first();
  await expect(comparison).toHaveAttribute("data-history-comparison", "paired");
  await page.setViewportSize({ width: 320, height: 720 });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "200%";
  });
  await expect(comparison).toHaveAttribute(
    "data-history-comparison",
    "stacked",
  );
  await expect(comparison).toContainText("Before: Inspected A");
  await expect(comparison).toContainText("After: Inspected A renamed");
  await page.evaluate(() => {
    document.documentElement.style.zoom = "";
  });
  await page.setViewportSize({ width: 1440, height: 900 });
  await expect(comparison).toHaveAttribute("data-history-comparison", "paired");
  const spacing = await page.addStyleTag({
    content:
      "[data-history-comparison] { letter-spacing: .12em; word-spacing: .16em; line-height: 1.5; }",
  });
  await expect(comparison).toHaveAttribute("data-history-comparison", "paired");
  await spacing.evaluate((element) => element.parentNode?.removeChild(element));
  const body = Array.from(
    { length: 20 },
    (_, index) => `Narrative line ${index + 1}: preserve this accepted note.`,
  ).join("\n");
  const note = await createViewRow(page, f.incident, notesViewSchemaId, {
    client_txn_id: uniqueTxn("reading-note"),
    "note.title": "Long saved note",
    "note.body": body,
  });
  await switchOrdinarySheet(page, notesViewSchemaId);
  await openGenericInspectorForRecord(page, notesViewSchemaId, note.record_id);
  const fullValue = page.getByRole("button", {
    name: "Show full value for Body",
    exact: true,
  });
  await expect(fullValue).toBeVisible();
  const saved = page.locator('[data-inspector-saved-field="note.body"]');
  await expect(saved).toHaveAttribute("data-inspector-value-kind", "narrative");
  const preview = saved.locator("dd").first().locator("div").first();
  await expect(preview).toHaveText(body);
  const before = (await preview.boundingBox())?.height ?? 0;
  const narrativeEditor = await editField(page, notesViewSchemaId, "note.body");
  await narrativeEditor.fill("  retained note authoring\n");
  const editorIdentity = await narrativeEditor.elementHandle();
  await fullValue.click();
  expect(
    await narrativeEditor.evaluate(
      (element, original) => element === original,
      editorIdentity,
    ),
  ).toBe(true);
  await expect(narrativeEditor).toHaveValue("  retained note authoring\n");
  await page.getByRole("button", { name: "Close editor", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Show less for Body", exact: true }),
  ).toBeVisible();
  expect((await preview.boundingBox())?.height ?? 0).toBeGreaterThan(
    before * 2,
  );
  for (const lines of [5, 6, 7]) {
    const text = Array.from(
      { length: lines },
      (_, index) => `Line ${index + 1}`,
    ).join("\n");
    await patchRecord(page, note.record_id, {
      view_schema_id: notesViewSchemaId,
      base_row_version: lines - 4,
      client_txn_id: uniqueTxn("reading-lines"),
      changes: [{ field_key: "note.body", value: text }],
    });
    await expect(preview).toHaveText(text);
    if (lines <= 6) await expect(fullValue).toHaveCount(0);
    else await expect(fullValue).toBeVisible();
  }
  const sparse = await createInspectorReadingFixture(page);
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.goto(
    `/?incident_id=${sparse.incidentId}&view_schema_id=${timelineViewSchemaId}`,
  );
  await openTimelineInspector(page, sparse.target.record_id);
  await page.evaluate(() => document.fonts.ready);
  const reading = page.getByTestId(timelineInspectorTestId());
  await reading.getByRole("button", { name: "Details", exact: true }).click();
  const rawActivity = reading.locator(
    '[data-inspector-saved-field="timeline.raw_activity_text"] [data-inspector-field-value] > div[id]',
  );
  await expect(rawActivity).toHaveText(
    "browser.inspector-history visual inspector details",
  );
  const geometry = await rawActivity.evaluate((element) => {
    const body = element.closest("[data-inspector-scroll-body]");
    if (!(body instanceof HTMLElement) || !element.firstChild)
      throw new Error("Missing reading surface");
    const range = document.createRange();
    range.selectNodeContents(element);
    const line = range.getClientRects()[0];
    if (!line) throw new Error("Missing first activity line");
    return {
      scroll: body.scrollTop,
      lineTop: line.top,
      lineBottom: line.bottom,
      bodyTop: body.getBoundingClientRect().top,
      bodyBottom: body.getBoundingClientRect().bottom,
    };
  });
  expect(geometry.lineTop).toBeGreaterThanOrEqual(geometry.bodyTop);
  expect(geometry.lineBottom).toBeLessThanOrEqual(geometry.bodyBottom);
  expect((await reading.boundingBox())?.width).toBe(420);
  for (const override of cartularyDesignPresentation.inspector
    .fieldLayoutOverrides) {
    await expect(
      reading.locator(`[data-inspector-saved-field="${override.fieldKey}"]`),
    ).toHaveAttribute("data-inspector-field-layout", override.layout);
  }
  const action = reading.locator(
    '[data-inspector-edit-field="timeline.date_entered_text"]',
  );
  expect((await action.boundingBox())?.height).toBeGreaterThanOrEqual(
    cartularyDesignPresentation.inspector.fieldActionMinSizePx,
  );
  await test.info().attach("inspector-sparse-reading", {
    body: await page.screenshot({ animations: "disabled" }),
    contentType: "image/png",
  });
  await page.setViewportSize({ width: 320, height: 720 });
  await expect
    .poll(async () =>
      reading
        .locator('[data-inspector-saved-field="timeline.date_entered_text"]')
        .evaluate((element) => {
          const label = element.querySelector("dt"),
            value = element.querySelector("[data-inspector-field-value]");
          return (
            !!label &&
            !!value &&
            label.getBoundingClientRect().bottom <=
              value.getBoundingClientRect().top
          );
        }),
    )
    .toBe(true);
  await test.info().attach("inspector-sparse-reading-320", {
    body: await page.screenshot({ animations: "disabled" }),
    contentType: "image/png",
  });
  const whitespace = "\t\r\n".repeat(300);
  const whitespaceRow = await createViewRow(
    page,
    sparse.incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("reading-whitespace"),
      "timeline.activity_synopsis_text": "Exact whitespace inspection",
      "timeline.raw_activity_text": whitespace,
    },
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.goto(
    `/?incident_id=${sparse.incidentId}&view_schema_id=${timelineViewSchemaId}`,
  );
  await openTimelineInspector(page, whitespaceRow.record_id);
  await page.setViewportSize({ width: 320, height: 640 });
  const rawWhitespace = page.locator(
    '[data-inspector-saved-field="timeline.raw_activity_text"]',
  );
  await rawWhitespace
    .getByText("Inspect source whitespace", { exact: true })
    .click();
  await expect(rawWhitespace.locator("code")).toHaveText(
    JSON.stringify(whitespace),
  );
  const whitespaceWidth = await reading.evaluate((element) => ({
    scroll: element.scrollWidth,
    client: element.clientWidth,
  }));
  expect(whitespaceWidth.scroll).toBeLessThanOrEqual(whitespaceWidth.client);
});

test("Inspector uncertain recovery replays exact requests without consuming newer authoring", async ({
  page,
}) => {
  for (const [view, malformed] of [
    [hostsViewSchemaId, false],
    [evidenceViewSchemaId, true],
  ] as const) {
    const f = await fixture(page, view),
      input = await editField(page, view, f.field),
      bodies: string[] = [];
    await page.route(
      `**/api/v1/records/${f.first.record_id}`,
      async (route) => {
        if (route.request().method() !== "PATCH") {
          await route.continue();
          return;
        }
        bodies.push(route.request().postData() ?? "");
        const response = await route.fetch();
        expect(response.ok()).toBeTruthy();
        if (bodies.length > 1) await route.fulfill({ response });
        else if (malformed)
          await route.fulfill({
            response,
            body: JSON.stringify({
              data: { row: {} },
              meta: { request_id: "invalid-inspector-receipt" },
            }),
          });
        else await route.abort("failed");
      },
    );
    await input.fill("Captured inspector value");
    await page.getByTestId(genericEditSubmitTestId(view)).click();
    const retry = page.getByRole("button", {
      name: "Retry original change",
      exact: true,
    });
    await expect(retry).toBeVisible();
    await expect(
      page.getByRole("button", { name: /^View:/, exact: true }),
    ).toBeVisible();
    await input.fill("Newer unfinished value");
    await expect(
      page.getByRole("button", { name: "Unfinished work (2)", exact: true }),
    ).toBeVisible();
    await page.getByTestId(workbookInspectorCloseButtonTestId(view)).click();
    await retry.focus();
    await retry.press("Enter");
    await expect(
      page.getByRole("region", { name: "Inspector changes" }),
    ).toContainText("Saved, version 2");
    expect(bodies).toHaveLength(2);
    expect(bodies[1]).toBe(bodies[0]);
    expect(await fetchRecordHistoryCount(page, f.first.record_id)).toBe(2);
    await expect(
      page.getByTestId(workbookInspectorCloseButtonTestId(view)),
    ).toHaveCount(0);
    await switchOrdinarySheet(page, timelineViewSchemaId);
    await switchOrdinarySheet(page, view);
    await openGenericInspectorForRecord(page, view, f.first.record_id);
    await editField(page, view, f.field);
    await expect(
      page.getByTestId(genericEditSubmitTestId(view)),
    ).toBeDisabled();
    await expect(
      page.getByRole("button", { name: /^Keep draft/ }),
    ).toBeVisible();
    await expect(input).toHaveValue("Newer unfinished value");
  }
});

test("Inspector acknowledged refresh recovery preserves newer grid focus and never dispatches another patch", async ({
  page,
}) => {
  const f = await fixture(page),
    input = await editField(page, f.view, "host.location");
  let release: () => void = () => {},
    committed = false,
    failRefresh = true,
    count = 0;
  await page.route(`**/api/v1/records/${f.first.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") {
      await route.continue();
      return;
    }
    count++;
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    committed = true;
    await new Promise<void>((resolve) => {
      release = resolve;
    });
    await route.fulfill({ response });
  });
  await page.route(`**/views/${f.view}/query`, async (route) => {
    if (committed && failRefresh) await route.abort("failed");
    else await route.continue();
  });
  await input.fill("Accepted location");
  await page.getByTestId(genericEditSubmitTestId(f.view)).click();
  await expect.poll(() => committed).toBeTruthy();
  await page.getByTestId(workbookInspectorCloseButtonTestId(f.view)).click();
  const id = rowCellTestId(f.second.record_id, "host.display_name");
  await scrollGridTargetIntoView({ page, surface: f.view, targetTestId: id });
  const cell = page
    .getByTestId(id)
    .locator("xpath=ancestor::*[@role='gridcell'][1]");
  await activateCommittedGridCell(cell);
  release();
  const recovery = page.getByRole("button", {
    name: "Refresh saved change",
    exact: true,
  });
  await expect(recovery).toBeVisible();
  await expect(cell).toBeFocused();
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${f.view}:${f.second.record_id}:host.display_name`,
  );
  await expect(
    page.getByTestId(workbookInspectorCloseButtonTestId(f.view)),
  ).toHaveCount(0);
  failRefresh = false;
  await recovery.click();
  await expect(recovery).toHaveCount(0);
  expect(count).toBe(1);
  expect(await fetchRecordHistoryCount(page, f.first.record_id)).toBe(2);
});

test("a11y.inspector retained editing and recovery remain named keyboard reachable and bounded in narrow layouts", async ({
  page,
}, info) => {
  const f = await fixture(page),
    input = await editField(page, f.view, "host.location");
  await input.fill("Keyboard retained location");
  await page.getByTestId(workbookInspectorCloseButtonTestId(f.view)).click();
  await page.getByTestId(workbookInspectorToggleTestId(f.view)).click();
  const resume = page.getByRole("button", {
    name: "Resume draft for Location",
    exact: true,
  });
  await resume.focus();
  await resume.press("Enter");
  await expect(input).toBeFocused();
  await expect(input).toHaveValue("Keyboard retained location");
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const [width, zoom] of [
    [1280, 1],
    [320, 1],
    [1280, 2],
  ] as const) {
    await page.setViewportSize({ width, height: 720 });
    await page.evaluate((zoom) => {
      document.documentElement.style.zoom = String(zoom);
      if (zoom > 1) {
        document.body.style.lineHeight = "1.5";
        document.body.style.letterSpacing = "0.12em";
        document.body.style.wordSpacing = "0.16em";
      }
    }, zoom);
    await input.scrollIntoViewIfNeeded();
    await input.focus();
    await expect(input).toBeFocused();
    await expect(input).toBeInViewport({ ratio: 1 });
    await expect(
      page.getByRole("button", { name: "Discard draft", exact: true }),
    ).toHaveAccessibleName("Discard draft");
    await info.attach(`inspector-edit-${width}-${zoom}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  await info.attach("inspector-edit-accessibility-tree", {
    body: await page.getByTestId(entityInspectorTestId("host")).ariaSnapshot(),
    contentType: "text/plain",
  });
});

test("Reference selection reaches later real targets and retains staged choices through source and refresh failures", async ({
  page,
  workerAdmin,
}) => {
  test.setTimeout(180_000);
  const taskView = taskRequestsViewSchemaId,
    noteView = notesViewSchemaId,
    indicatorView = indicatorsViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("RSR"),
    "Reference selection recovery",
  );
  const task = await createViewRow(page, incident, taskView, {
    client_txn_id: uniqueTxn("rsr-task"),
    "task.title": "Reference owner",
    "task.task_kind": "question",
    "task.owner_user_id": workerAdmin.user_id,
  });
  const notes = [];
  for (let offset = 0; offset < 105; offset += 5) {
    notes.push(
      ...(await Promise.all(
        Array.from({ length: 5 }, (_, index) =>
          createViewRow(page, incident, noteView, {
            client_txn_id: uniqueTxn("rsr-note"),
            "note.title": `Reference ${String(offset + index).padStart(3, "0")}`,
          }),
        ),
      )),
    );
  }
  const indicator = await createViewRow(page, incident, indicatorView, {
    client_txn_id: uniqueTxn("rsr-indicator"),
    "indicator.display_value": "reference.example",
    "indicator.indicator_type": "domain_name",
    "indicator.value_kind": "atomic",
  });
  const retained = notes[0];
  if (!retained) throw new Error("Missing retained Note");
  let initialFailure = true,
    continuationFailure = true,
    refreshFailure = false;
  const reads: Record<string, unknown>[] = [],
    writes: Record<string, unknown>[] = [];
  await page.route(`**/views/${noteView}/query`, async (route) => {
    const body = route.request().postDataJSON();
    reads.push(body);
    if (!body.cursor_token && initialFailure) {
      initialFailure = false;
      return route.abort("failed");
    }
    if (body.cursor_token && continuationFailure) {
      continuationFailure = false;
      return route.abort("failed");
    }
    await route.continue();
  });
  await page.route(`**/views/${taskView}/query`, (route) =>
    refreshFailure ? route.abort("failed") : route.continue(),
  );
  await page.route(`**/api/v1/records/${task.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    writes.push(route.request().postDataJSON());
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    refreshFailure = true;
    await route.fulfill({ response });
  });
  await page.goto(`/?incident_id=${incident}&view_schema_id=${taskView}`);
  await openGenericInspectorForRecord(page, taskView, task.record_id);
  const input = await editField(page, taskView, "task.linked_record_ids");
  await input.fill(retained.record_id);
  const choose = page
    .getByTestId(workbookInspectorPanelTestId(taskView, "details"))
    .getByRole("button", { name: "Choose linked records", exact: true });
  await choose.focus();
  await choose.press("Enter");
  const popup = page.getByRole("dialog", {
    name: "Choose linked records",
    exact: true,
  });
  await popup.getByLabel("Reference surface").selectOption(noteView);
  await expect(popup.getByRole("alert")).toContainText("could not be loaded");
  await expect(input).toHaveValue(retained.record_id);
  await popup
    .getByRole("button", { name: "Retry references", exact: true })
    .click();
  await expect(popup).toContainText("Page 1: 100 candidates; more available");
  const list = popup.getByRole("listbox", {
    name: "Linked Records candidates",
  });
  const first = await list.locator("option").first().getAttribute("value");
  if (!first) throw new Error("Missing page one candidate");
  const selectedOnPage = () =>
    list.evaluate((element) =>
      Array.from(
        (element as HTMLSelectElement).selectedOptions,
        (option) => option.value,
      ),
    );
  await list.selectOption([...(await selectedOnPage()), first]);
  const next = popup.getByRole("button", { name: "Next", exact: true });
  await next.focus();
  await next.press("Enter");
  await expect(popup.getByRole("alert")).toContainText(
    "accepted page is retained",
  );
  await expect(next).toBeFocused();
  await expect(list.locator("option")).toHaveCount(100);
  const retry = popup.getByRole("button", {
    name: "Retry references",
    exact: true,
  });
  await retry.focus();
  await retry.press("Space");
  await expect(popup).toContainText("Page 2: 5 candidates; end of this source");
  await expect(retry).toBeFocused();
  await expect(retry).toHaveAttribute("aria-disabled", "true");
  const later = await list.locator("option").last().getAttribute("value");
  if (!later) throw new Error("Missing later candidate");
  await list.selectOption([...(await selectedOnPage()), later]);
  await popup
    .getByLabel("Reference filter field")
    .selectOption("note.created_by_user_id");
  const beforeFilter = reads.length;
  await popup
    .getByLabel("Reference filter value")
    .fill("00000000-0000-4000-8000-000000000999");
  expect(reads).toHaveLength(beforeFilter);
  await popup
    .getByRole("button", { name: "Apply filter", exact: true })
    .click();
  await expect(popup).toContainText("Page 1: 0 candidates; end of this source");
  await popup.getByLabel("Reference filter value").fill(workerAdmin.user_id);
  await popup
    .getByRole("button", { name: "Apply filter", exact: true })
    .click();
  await expect(popup).toContainText("Page 1: 100 candidates; more available");
  await expect(list.locator("option")).toHaveCount(100);
  const surface = popup.getByLabel("Reference surface");
  await surface.focus();
  await surface.selectOption(indicatorView);
  await expect(surface).toBeFocused();
  await expect(list.locator("option")).toHaveCount(1);
  await list.selectOption(`record:${indicator.record_id}`);
  expect(writes).toHaveLength(0);
  await expect(input).toHaveValue(retained.record_id);
  await popup
    .getByRole("button", { name: "Use selection", exact: true })
    .click();
  const selected = (await input.inputValue()).split("\n");
  expect(selected).toContain(retained.record_id);
  expect(selected).toContain(first.replace("record:", ""));
  expect(selected).toContain(later.replace("record:", ""));
  expect(selected).toContain(indicator.record_id);
  expect(writes).toHaveLength(0);
  await choose.click();
  await popup
    .getByRole("button", { name: "Cancel references", exact: true })
    .click();
  expect((await input.inputValue()).split("\n")).toEqual(selected);
  await page.getByTestId(genericEditSubmitTestId(taskView)).click();
  await expect(
    page.getByRole("region", { name: "Inspector changes" }),
  ).toContainText("Saved");
  expect(writes).toHaveLength(1);
  expect(writes[0]?.changes).toEqual([
    {
      field_key: "task.linked_record_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: selected.map((linked_record_id) => ({
          op: "add_record_ref",
          linked_record_id,
        })),
      },
    },
  ]);
  // A separate candidate retry after the acknowledgement is still read-only.
  await choose.click();
  await popup.getByLabel("Reference surface").selectOption(noteView);
  await expect(popup).toContainText("Page 1: 100 candidates");
  await popup.getByRole("button", { name: "Next", exact: true }).click();
  await expect(popup).toContainText("Page 2: 5 candidates");
  await popup
    .getByRole("button", { name: "Cancel references", exact: true })
    .click();
  expect(writes).toHaveLength(1);
  expect(reads.length).toBeLessThanOrEqual(10);
  refreshFailure = false;
  await page
    .getByRole("button", { name: "Refresh saved change", exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: "Refresh saved change", exact: true }),
  ).toHaveCount(0);
  expect(writes).toHaveLength(1);
  const saved = (await queryViewRows(page, incident, taskView)).find(
    (row) => row.record_id === task.record_id,
  );
  const value = saved?.cells["task.linked_record_ids"]?.value as {
    items: { item_ref: string }[];
  };
  expect(value.items).toHaveLength(selected.length);
  await page
    .getByRole("combobox", { name: "Collection edit action" })
    .selectOption("remove");
  const removal = page.getByTestId(genericEditValueTestId(taskView));
  const itemRef = await removal.locator("option").first().getAttribute("value");
  expect(value.items.map((item) => item.item_ref)).toContain(itemRef);
  if (!itemRef) throw new Error("Missing removal item_ref");
  await removal.selectOption(itemRef);
  const removalAccepted = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${task.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(taskView)).click();
  expect((await removalAccepted).ok()).toBe(true);
  await expect(
    page.getByRole("region", { name: "Inspector changes" }),
  ).toContainText("Saved, version 3");
  await expect.poll(() => writes.length).toBe(2);
  expect(writes[1]?.changes).toEqual([
    {
      field_key: "task.linked_record_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "remove_record_ref", item_ref: itemRef }],
      },
    },
  ]);
});

test("a11y.references native popup preserves keyboard focus and fits narrow zoomed and spaced layouts", async ({
  page,
}, info) => {
  const f = await fixture(page, evidenceViewSchemaId);
  await createViewRow(page, f.incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("rsr-a11y"),
    "party.display_name": "Keyboard accessible Party",
    "party.party_kind": "person",
  });
  const input = await editField(page, f.view, "evidence.source_party_id");
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const [width, zoom] of [
    [1280, 1],
    [390, 1],
    [1280, 2],
  ] as const) {
    await page.setViewportSize({ width, height: 720 });
    await page.evaluate((zoom) => {
      document.documentElement.style.zoom = String(zoom);
      document.body.style.lineHeight = "1.5";
      document.body.style.letterSpacing = "0.12em";
      document.body.style.wordSpacing = "0.16em";
    }, zoom);
    const trigger = page
      .getByTestId(workbookInspectorPanelTestId(f.view, "details"))
      .getByRole("button", { name: "Choose source party", exact: true });
    await trigger.scrollIntoViewIfNeeded();
    await trigger.focus();
    await trigger.press("Enter");
    const popup = page.getByRole("dialog", {
      name: "Choose source party",
      exact: true,
    });
    await expect(popup).toContainText(
      "Page 1: 1 candidates; end of this source",
    );
    expect(
      await popup.evaluate((element) => element.matches(":popover-open")),
    ).toBe(true);
    await expect(popup).toBeInViewport({ ratio: 1 });
    const list = popup.getByRole("listbox", {
      name: "Source Party candidates",
    });
    await list.focus();
    await expect(list).toBeFocused();
    const cancel = popup.getByRole("button", {
      name: "Cancel references",
      exact: true,
    });
    await cancel.focus();
    await cancel.press("Tab");
    expect(
      await popup.evaluate((element) =>
        element.contains(document.activeElement),
      ),
    ).toBe(true);
    await info.attach(`reference-picker-${width}-${zoom}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
    await info.attach(`reference-picker-tree-${width}-${zoom}`, {
      body: await popup.ariaSnapshot(),
      contentType: "text/plain",
    });
    await page.keyboard.press("Escape");
    await expect(popup).toHaveCount(0);
    await expect(input).toBeFocused();
  }
});

test("Ordinary reference picker restores focus after keyboard removal of its sole staged Party", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("RKR"),
    "Reference keyboard removal",
  );
  const evidence = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("rkr-evidence"),
    "evidence.title": "Collector focus evidence",
  });
  await createViewRow(page, incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("rkr-party"),
    "party.display_name": "Collector focus Party",
    "party.party_kind": "person",
  });
  const writes: Record<string, unknown>[] = [];
  await page.route(`**/api/v1/records/${evidence.record_id}`, (route) => {
    if (route.request().method() === "PATCH")
      writes.push(route.request().postDataJSON());
    return route.continue();
  });
  for (const [width, key] of [
    [1440, "Enter"],
    [1024, "Space"],
  ] as const) {
    await page.setViewportSize({ width, height: width === 1440 ? 900 : 720 });
    await page.goto(
      `/?incident_id=${incident}&view_schema_id=${evidenceViewSchemaId}`,
    );
    await openGenericInspectorForRecord(
      page,
      evidenceViewSchemaId,
      evidence.record_id,
    );
    const input = await editField(
      page,
      evidenceViewSchemaId,
      "evidence.collector_party_id",
    );
    const trigger = page
      .getByTestId(
        workbookInspectorPanelTestId(evidenceViewSchemaId, "details"),
      )
      .getByRole("button", {
        name: "Choose collector party",
        exact: true,
      });
    await trigger.focus();
    await page.keyboard.press("Enter");
    const popup = page.getByRole("dialog", {
      name: "Choose collector party",
      exact: true,
    });
    const candidates = popup.getByRole("listbox", {
      name: "Collector Party candidates",
    });
    await expect(candidates).toBeVisible();
    await candidates.focus();
    await page.keyboard.press("ArrowDown");
    const remove = popup.getByRole("button", { name: /^Remove selected / });
    await expect(remove).toHaveCount(1);
    await page.keyboard.press("Tab");
    await expect(remove).toBeFocused();
    const scrollBefore = await page.evaluate((selector) => {
      const grid = document.querySelector<HTMLElement>(selector);
      return {
        pageX: window.scrollX,
        pageY: window.scrollY,
        gridX: grid?.scrollLeft ?? 0,
        gridY: grid?.scrollTop ?? 0,
      };
    }, gridScrollportSelector());
    await page.keyboard.press(key);
    await expect(remove).toHaveCount(0);
    await expect(candidates).toBeFocused();
    await expect(input).toHaveValue("");
    expect(
      await page.evaluate((selector) => {
        const grid = document.querySelector<HTMLElement>(selector);
        return {
          pageX: window.scrollX,
          pageY: window.scrollY,
          gridX: grid?.scrollLeft ?? 0,
          gridY: grid?.scrollTop ?? 0,
        };
      }, gridScrollportSelector()),
    ).toEqual(scrollBefore);
    expect(writes).toHaveLength(0);
    await page.keyboard.press("Escape");
    await expect(popup).toHaveCount(0);
    await expect(input).toBeFocused();
    expect(writes).toHaveLength(0);
  }
});

test("Reference target deletion merge and membership removal preserve exact choices and owner admission", async ({
  page,
  workerAdmin,
  workerAdminRequest,
}) => {
  const f = await fixture(page, evidenceViewSchemaId);
  const partyView = partiesViewSchemaId,
    taskView = taskRequestsViewSchemaId;
  const party = await createViewRow(page, f.incident, partyView, {
    client_txn_id: uniqueTxn("rsr-deleted-party"),
    "party.display_name": "Chosen before deletion",
    "party.party_kind": "person",
  });
  const input = await editField(page, f.view, "evidence.source_party_id");
  const panel = page.getByTestId(
    workbookInspectorPanelTestId(f.view, "details"),
  );
  await panel.getByRole("button", { name: "Choose source party" }).click();
  const picker = page.getByRole("dialog", { name: "Choose source party" });
  await picker
    .getByRole("listbox", { name: "Source Party candidates" })
    .selectOption(`party:${party.record_id}`);
  await picker.getByRole("button", { name: "Use selection" }).click();
  const deleted = await publicHttpOperation({
    operationID: "deleteRecord",
    pathParameters: { record_id: party.record_id },
    body: {
      base_row_version: 1,
      client_txn_id: uniqueTxn("rsr-delete-target"),
      reason: "Reference eligibility evidence",
    },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(deleted.ok).toBe(true);
  const response = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${f.first.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(f.view)).click();
  expect((await response).ok()).toBe(false);
  await expect(input).toHaveValue(party.record_id);
  await panel.getByRole("button", { name: "Choose source party" }).click();
  await expect(picker).toContainText(
    "Page 1: 0 candidates; end of this source",
  );
  await expect(picker).toContainText("Chosen before deletion");
  await picker.getByRole("button", { name: "Cancel references" }).click();
  await expect(input).toHaveValue(party.record_id);
  const task = await createViewRow(page, f.incident, taskView, {
    client_txn_id: uniqueTxn("rsr-member-task"),
    "task.title": "Member target",
    "task.task_kind": "question",
    "task.owner_user_id": workerAdmin.user_id,
  });
  const member = await createIncidentMemberUser(page, f.incident, {
    email: uniqueEmail("rsr-membership"),
    display_name: "Selected incident member",
    initial_password: "ReferenceMembership1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await switchOrdinarySheet(page, taskView);
  await openGenericInspectorForRecord(page, taskView, task.record_id);
  const owner = await editField(page, taskView, "task.owner_user_id");
  const taskPanel = page.getByTestId(
    workbookInspectorPanelTestId(taskView, "details"),
  );
  await taskPanel
    .getByRole("button", { name: "Choose owner", exact: true })
    .click();
  const members = page.getByRole("dialog", {
    name: "Choose owner",
    exact: true,
  });
  await members
    .getByRole("listbox", { name: "Owner candidates" })
    .selectOption(`incident_member:${member.user_id}`);
  await members.getByRole("button", { name: "Use selection" }).click();
  expect(
    (
      await workerAdminRequest.delete(
        `/api/v1/incidents/${f.incident}/memberships/${member.user_id}`,
        { data: { base_membership_version: 1 } },
      )
    ).status(),
  ).toBe(204);
  const rejected = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${task.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(taskView)).click();
  expect((await rejected).ok()).toBe(false);
  await expect(owner).toHaveValue(member.user_id);
  await taskPanel
    .getByRole("button", { name: "Choose owner", exact: true })
    .click();
  await expect(
    members.getByRole("option", { name: new RegExp(member.user_id) }),
  ).toHaveCount(0);
  await expect(members).toContainText("Selected incident member");
  await members.getByRole("button", { name: "Cancel references" }).click();
  await expect(owner).toHaveValue(member.user_id);
  const loser = await createViewRow(page, f.incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("rsr-merge-loser"),
    "host.display_name": "Chosen before merge",
  });
  const survivor = await createViewRow(page, f.incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("rsr-merge-survivor"),
    "host.display_name": "Surviving target",
  });
  const linked = await editField(page, taskView, "task.linked_record_ids");
  await taskPanel
    .getByRole("button", { name: "Choose linked records", exact: true })
    .click();
  const records = page.getByRole("dialog", {
    name: "Choose linked records",
    exact: true,
  });
  await records.getByLabel("Reference surface").selectOption(hostsViewSchemaId);
  await records
    .getByRole("listbox", { name: "Linked Records candidates" })
    .selectOption(`record:${loser.record_id}`);
  await records.getByRole("button", { name: "Use selection" }).click();
  const merged = await publicHttpOperation({
    operationID: "mergeEntityRecord",
    pathParameters: { survivor_record_id: survivor.record_id },
    body: {
      client_txn_id: uniqueTxn("rsr-merge"),
      loser_record_id: loser.record_id,
      loser_base_row_version: 1,
      survivor_base_row_version: 1,
      reason: "Reference target merge evidence",
    },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(merged.ok).toBe(true);
  await taskPanel
    .getByRole("button", { name: "Choose linked records", exact: true })
    .click();
  await expect(records).toContainText(
    "Page 1: 1 candidates; end of this source",
  );
  await expect(
    records.getByRole("option", { name: new RegExp(loser.record_id) }),
  ).toHaveCount(0);
  await expect(records).toContainText("Chosen before merge");
  await records.getByRole("button", { name: "Cancel references" }).click();
  await expect(linked).toHaveValue(loser.record_id);
  const mergeResult = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${task.record_id}`),
  );
  await page.getByTestId(genericEditSubmitTestId(taskView)).click();
  const mergeResponse = await mergeResult;
  expect(mergeResponse.ok()).toBe(true);
  expect(
    mergeResponse.request().postDataJSON().changes[0].action_payload.actions,
  ).toEqual([{ op: "add_record_ref", linked_record_id: loser.record_id }]);
  const savedTask = (await queryViewRows(page, f.incident, taskView)).find(
    (row) => row.record_id === task.record_id,
  );
  const links = savedTask?.cells["task.linked_record_ids"]?.value as {
    items: { item_ref: string }[];
  };
  expect(links.items.map((item) => item.item_ref)).toEqual([
    `record_ref:${loser.record_id}`,
  ]);
});
