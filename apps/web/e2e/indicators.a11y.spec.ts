import {
  indicatorCreateTestId,
  indicatorLifecycleTestId,
  indicatorObservationTestId,
} from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectDecisionControlReachable } from "./support/workbook/decisionSupersession";
import {
  createCanonicalObservationFixture,
  openCanonicalProposal,
} from "./support/workbook/indicatorCanonicalCreate";
import {
  createLifecycleFixture,
  openLifecycleEditor,
} from "./support/workbook/indicatorLifecycle";
import {
  createObservationFixture,
  openObservationEditor,
} from "./support/workbook/indicatorObservations";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";

test("a11y.indicator-lifecycle UTC errors and retained recovery remain keyboard reachable", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const { incidentId, indicator } = await createLifecycleFixture(page);
  const editor = await openLifecycleEditor(
    page,
    incidentId,
    indicator.record_id,
  );
  const submit = editor.getByRole("button", {
    name: "Append lifecycle interval",
    exact: true,
  });
  await submit.focus();
  await submit.press("Enter");
  const from = editor.getByLabel("Effective from (UTC)", { exact: true });
  await expect(from).toHaveAttribute("aria-invalid", "true");
  await expect(from).toHaveAccessibleDescription(/UTC/);
  await from.fill("2026-09-11T12:00");
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [
      from,
      editor.getByLabel("Confidence (optional)", { exact: true }),
      submit,
    ]) {
      await expectDecisionControlReachable(page, control);
      await expectVisibleFocus(control);
    }
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectDecisionControlReachable(page, submit);
  });
  await expectAllInteractiveControlsNamed(page);
  await submit.focus();
  await submit.press("Enter");
  await expect(
    page.getByText("Interval saved. Indicator and history refreshed.", {
      exact: true,
    }),
  ).toBeVisible();
  const trigger = recoveryEntry(page);
  await trigger.focus();
  await openRecoveryItem(page, /^Indicator interval ·/);
  const recovery = page.getByTestId(indicatorLifecycleTestId("recovery"));
  const recoveryHeading = page
    .getByRole("region", { name: "Recovery navigation", exact: true })
    .locator(":scope > h2");
  await expect(recoveryHeading).toHaveAccessibleName("Indicator interval");
  await expect(recoveryHeading).toBeFocused();
  await expect(
    recovery.getByText("Interval saved. Indicator and history refreshed.", {
      exact: true,
    }),
  ).toBeVisible();
  await testInfo.attach("indicator-lifecycle-accessibility-tree", {
    body: await recovery.ariaSnapshot(),
    contentType: "text/plain",
  });
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
});

test("a11y.indicator-observations exact source selection and retained recovery remain keyboard reachable", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const { incidentId, source } = await createObservationFixture(page);
  await page.goto(`/?incident_id=${incidentId}`);
  const editor = await openObservationEditor(page, source.record_id);
  const text = editor.getByRole("textbox", {
    name: "Saved source text",
    exact: true,
  });
  await text.scrollIntoViewIfNeeded();
  await text.focus();
  await page.keyboard.press("Control+Home");
  await page.keyboard.press("Shift+End");
  expect(
    await text.evaluate((node) => {
      const input = node as HTMLTextAreaElement;
      return [input.selectionStart, input.selectionEnd];
    }),
  ).toEqual([0, 13]);
  await expect(text).toHaveAttribute("aria-readonly", "true");
  // Typing, deletion and line insertion cannot alter the saved source preview.
  const sourceText = await text.inputValue();
  for (const key of ["x", "Backspace", "Delete", "Enter"]) {
    await page.keyboard.press(key);
    await expect(text).toHaveValue(sourceText);
  }
  await page.keyboard.press("Control+Home");
  await page.keyboard.press("Shift+End");
  const useSelection = editor.getByRole("button", {
    name: "Use selected text",
    exact: true,
  });
  await useSelection.focus();
  await useSelection.press("Enter");
  await expect(
    page.getByTestId(indicatorObservationTestId("preview")),
  ).toContainText("alpha.example");
  const submit = editor.getByRole("button", {
    name: "Create observation",
    exact: true,
  });
  await expect(submit).toBeEnabled();
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [text, useSelection, submit]) {
      await expectDecisionControlReachable(page, control);
      await expectVisibleFocus(control);
    }
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectDecisionControlReachable(page, submit);
  });
  await expectAllInteractiveControlsNamed(page);
  await submit.focus();
  await submit.press("Enter");
  await expect(
    page.getByText("Observation change saved. Records and history refreshed.", {
      exact: true,
    }),
  ).toBeVisible();
  const trigger = recoveryEntry(page);
  await trigger.focus();
  await openRecoveryItem(page, /^Indicator observation ·/);
  const recovery = page.getByTestId(indicatorObservationTestId("recovery"));
  const recoveryHeading = page
    .getByRole("region", { name: "Recovery navigation", exact: true })
    .locator(":scope > h2");
  await expect(recoveryHeading).toHaveAccessibleName("Indicator observation");
  await expect(recoveryHeading).toBeFocused();
  await expect(
    recovery.getByText(
      "Observation change saved. Records and history refreshed.",
      { exact: true },
    ),
  ).toBeVisible();
  await testInfo.attach("indicator-observation-accessibility-tree", {
    body: await recovery.ariaSnapshot(),
    contentType: "text/plain",
  });
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
});

test("a11y.canonical-indicator proposal validation and separate resolution remain keyboard reachable", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const fixture = await createCanonicalObservationFixture(page);
  await page.goto(`/?incident_id=${fixture.incidentId}`);
  const { form, editor } = await openCanonicalProposal(
    page,
    fixture.source.record_id,
  );
  const value = form.getByRole("textbox", {
    name: "Canonical value",
    exact: true,
  });
  const submit = form.getByRole("button", {
    name: "Create canonical Indicator",
    exact: true,
  });
  await value.fill("");
  await submit.focus();
  await submit.press("Enter");
  await expect(value).toBeFocused();
  await expect(value).toHaveAttribute("aria-invalid", "true");
  await expect(value).toHaveAccessibleDescription("Required.");
  await value.fill("NEW.EXAMPLE");
  await form.getByText("Additional canonical details", { exact: true }).click();
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [
      form.getByLabel("Indicator type", { exact: true }),
      form.getByLabel("Value kind", { exact: true }),
      value,
      submit,
    ]) {
      await expectDecisionControlReachable(page, control);
      await expectVisibleFocus(control);
    }
    await testInfo.attach(`canonical-proposal-${viewport.width}-review`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectDecisionControlReachable(page, submit);
  });
  await expectAllInteractiveControlsNamed(page);
  await submit.focus();
  await submit.press("Enter");
  await expect(
    editor.getByText("Indicator and history refreshed.", { exact: true }),
  ).toBeVisible();
  await expect(
    editor.getByText("Not linked to this Indicator.", { exact: true }),
  ).toBeVisible();
  const resolve = editor.getByRole("button", {
    name: "Resolve observation to this Indicator",
    exact: true,
  });
  await page.setViewportSize({ width: 390, height: 480 });
  await expectDecisionControlReachable(page, resolve);
  await expectVisibleFocus(resolve);
  await testInfo.attach("canonical-available-narrow-review", {
    body: await page.screenshot({ animations: "disabled", caret: "hide" }),
    contentType: "image/png",
  });
  await testInfo.attach("canonical-create-accessibility-tree", {
    body: await editor.ariaSnapshot(),
    contentType: "text/plain",
  });
  const trigger = recoveryEntry(page);
  await expectDecisionControlReachable(page, trigger);
  expect((await trigger.boundingBox())?.width).toBeGreaterThanOrEqual(32);
  await trigger.focus();
  await openRecoveryItem(page, /^Canonical Indicator creation ·/);
  await expect(
    page.getByTestId(indicatorCreateTestId("recovery")),
  ).toBeVisible();
  const canonicalRecoveryHeading = page
    .getByRole("region", { name: "Recovery navigation", exact: true })
    .locator(":scope > h2");
  await expect(canonicalRecoveryHeading).toHaveAccessibleName(
    "Canonical Indicator creation",
  );
  await expect(canonicalRecoveryHeading).toBeFocused();
  await expectDecisionControlReachable(
    page,
    page.getByRole("button", {
      name: "Close recovery",
      exact: true,
    }),
  );
  await testInfo.attach("canonical-recovery-narrow-review", {
    body: await page.screenshot({ animations: "disabled", caret: "hide" }),
    contentType: "image/png",
  });
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
});
