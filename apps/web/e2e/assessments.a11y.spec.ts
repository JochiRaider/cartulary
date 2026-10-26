import {
  assessmentCreateControlTestId,
  assessmentCreatePanelTestId,
  workbookAddRowButtonTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { expectDecisionControlReachable } from "./support/workbook/decisionSupersession";
import { createViewRow } from "./support/workbook/query";
import { activateCandidateIdentities } from "./support/workbook/references";

test("a11y.assessment deliberate subjects staged support and retained drafts remain reachable", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ASSESSMENT-A11Y"),
    "Accessible Assessment authoring",
  );
  const subject = await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("a11y-subject"),
    "host.display_name": "Assessment subject",
  });
  const support = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y-support"),
    "timeline.activity_synopsis_text": "Assessment supporting observation",
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${assessmentsViewSchemaId}`,
  );
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
  const subjectPicker = page.getByTestId(
    assessmentCreateControlTestId("subject"),
  );
  const rationale = page.getByTestId(
    assessmentCreateControlTestId("rationale"),
  );
  const timestamp = page.getByTestId(
    assessmentCreateControlTestId("assessed-at"),
  );
  const submit = page.getByTestId(assessmentCreateControlTestId("submit"));
  await expect(subjectPicker.getByRole("radio", { checked: true })).toHaveCount(
    0,
  );
  await subjectPicker
    .locator(`input[type="radio"][value="${subject.record_id}"]`)
    .check();
  await rationale.fill("Retained analyst judgment.");
  await timestamp.fill("unfinished");
  await submit.focus();
  await submit.press("Enter");
  await expect(timestamp).toHaveAttribute("aria-invalid", "true");
  await expect(timestamp).toHaveAccessibleDescription(
    /Enter an RFC3339 timestamp/,
  );
  await timestamp.fill("");
  const choose = page.getByRole("button", {
    name: "Choose support",
    exact: true,
  });
  await choose.focus();
  await choose.press("Enter");
  const candidates = page.getByTestId(
    assessmentCreateControlTestId("support-refs"),
  );
  await activateCandidateIdentities(candidates, support.record_id);
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
  ]) {
    await page.setViewportSize(viewport);
    await expectDecisionControlReachable(
      page,
      candidates.getByRole("checkbox").first(),
    );
    await testInfo.attach(`assessment-support-discovery-${viewport.width}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  await candidates.press("Escape");
  await expect(choose).toBeFocused();
  await expect(
    page.getByRole("region", { name: "Assessment supporting records" }),
  ).toContainText("No supporting records selected.");
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [
      subjectPicker.getByRole("radio", { checked: true }),
      rationale,
      choose,
      submit,
    ]) {
      await expectDecisionControlReachable(page, control);
      await expectVisibleFocus(control);
    }
    await testInfo.attach(`assessment-authoring-${viewport.width}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
    await subjectPicker.scrollIntoViewIfNeeded();
    await testInfo.attach(`assessment-subject-${viewport.width}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectDecisionControlReachable(page, submit);
  });
  await expectAllInteractiveControlsNamed(page);
  await page.getByTestId(assessmentCreatePanelTestId()).press("Escape");
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
  const resume = page.getByRole("button", {
    name: "Resume assessment draft",
    exact: true,
  });
  await resume.focus();
  await resume.press("Enter");
  await expect(rationale).toHaveValue("Retained analyst judgment.");
  await testInfo.attach("assessment-accessibility-tree", {
    body: await page.getByTestId(assessmentCreatePanelTestId()).ariaSnapshot(),
    contentType: "text/plain",
  });
});
