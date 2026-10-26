import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectDecisionControlReachable } from "./support/workbook/decisionSupersession";
import {
  openNoteFixture,
  retainNoteUncertainResult,
} from "./support/workbook/noteCreate";

test("a11y.linked-note authoring source selection and uncertain recovery remain keyboard accessible", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const f = await openNoteFixture(page);
  const title = f.form.getByRole("textbox", { name: "Title", exact: true });
  const body = f.form.getByRole("textbox", { name: "Body", exact: true });
  const tags = f.form.getByRole("textbox", {
    name: "Tags (one per line)",
    exact: true,
  });
  const submit = f.form.getByRole("button", {
    name: "Create Note",
    exact: true,
  });
  await expect(title).toBeFocused();
  await submit.focus();
  await submit.press("Enter");
  await expect(title).toHaveAttribute("aria-invalid", "true");
  await expect(title).toHaveAccessibleDescription(/title or body/);
  await title.fill("Retained investigation Note");
  await body.fill("Unfinished analysis remains attached to its chosen source.");
  const choose = f.form.getByRole("button", {
    name: "Choose source",
    exact: true,
  });
  await choose.focus();
  await choose.press("Enter");
  const picker = page.getByRole("region", {
    name: "Choose Note source",
    exact: true,
  });
  const candidates = picker.getByRole("group", {
    name: "Note source",
    exact: true,
  });
  await expect(
    picker.getByRole("button", { name: "Apply source", exact: true }),
  ).toBeEnabled();
  await picker
    .getByRole("button", { name: /^Remove selected Note source / })
    .click();
  await candidates.press("Escape");
  await expect(choose).toBeFocused();
  await expect(f.form).toContainText("Reviewed investigation source");
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [title, body, tags, choose, submit]) {
      await expectDecisionControlReachable(page, control);
      await expectVisibleFocus(control);
    }
    await testInfo.attach(`linked-note-authoring-${viewport.width}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectDecisionControlReachable(page, submit);
  });
  const { recovery, summary } = await retainNoteUncertainResult(
    page,
    f.source.record_id,
    f.view,
  );
  const recover = recovery.getByRole("button", {
    name: "Recover submission",
    exact: true,
  });
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    await expectDecisionControlReachable(page, recover);
    await expectVisibleFocus(recover);
  }
  await expectAllInteractiveControlsNamed(page);
  await testInfo.attach("linked-note-recovery-tree", {
    body: await recovery.ariaSnapshot(),
    contentType: "text/plain",
  });
  await recover.press("Escape");
  await expect(summary).toBeFocused();
  await expect(recovery).not.toBeVisible();
});
