import { incidentLandingTestId } from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectAndRecordContrast } from "./support/accessibility/records";
import { expectNoPrivateDiagnostics } from "./support/accessibility/status";
import {
  expectCreationControlReachable,
  holdCreationFailure,
  openCreationPresentation,
} from "./support/incidents/creation";

test("a11y.incident-creation keyboard feedback and constrained recovery remain reachable", async ({
  testActorPage: page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.setViewportSize({ width: 1280, height: 720 });
  const form = await openCreationPresentation(page);
  const key = form.getByLabel("Incident key", { exact: true });
  const title = form.getByLabel("Title", { exact: true });
  await expect(key).toBeFocused();
  await expectVisibleFocus(key);
  await page.keyboard.press("Enter");
  await expect(key).toHaveAccessibleDescription("Incident key is required.");
  await expect(title).toHaveAccessibleDescription("Title is required.");
  await expect(form.getByRole("alert")).toHaveCount(1);
  await expect(key).toBeFocused();
  await key.fill("IR-ACCESSIBLE-CREATE");
  await title.fill("Long investigation title ".repeat(15));
  const summary = form.locator("summary");
  await summary.focus();
  await page.keyboard.press("Enter");
  await form
    .getByLabel("Description", { exact: true })
    .fill("First line\nSecond line with investigation context ".repeat(12));
  const severity = form.getByLabel("Severity", { exact: true });
  await severity.fill("x".repeat(129));
  await summary.focus();
  await page.keyboard.press("Enter");
  await title.focus();
  await page.keyboard.press("Enter");
  await expect(severity).toHaveAttribute("aria-invalid", "true");
  await expect(severity).toBeVisible();
  await expect(severity).toHaveAccessibleDescription(/shorter|Shorten/);
  await expect(title).toBeFocused();
  await severity.fill("high");
  const { release, requested } = await holdCreationFailure(page);
  await title.press("Enter");
  await requested.promise;
  await expect(form).toHaveAttribute("aria-busy", "true");
  const pendingSubmit = form.getByTestId(
    incidentLandingTestId("create-submit-button"),
  );
  await expect(pendingSubmit).toBeDisabled();
  await expect(pendingSubmit).toHaveAccessibleName("Create and open — pending");
  for (const [width, height] of [
    [1280, 720],
    [1024, 720],
    [768, 640],
    [640, 480],
  ]) {
    await page.setViewportSize({ width: width ?? 1280, height: height ?? 720 });
    await expectCreationControlReachable(page, title);
    await expectCreationControlReachable(
      page,
      form.getByLabel("External case", { exact: true }),
    );
    await expectCreationControlReachable(
      page,
      form.getByRole("button", { name: "Close new incident" }),
    );
  }
  release.release();
  const retry = form.getByRole("button", { name: "Retry creation" });
  await expect(retry).toBeVisible();
  await expect(form).toHaveAttribute("aria-busy", "false");
  await expectCreationControlReachable(page, retry);
  await expectVisibleFocus(retry);
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectCreationControlReachable(page, retry);
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await expectCreationControlReachable(page, retry);
      await expectAllInteractiveControlsNamed(page);
      await expectAndRecordContrast(page, [
        incidentLandingTestId("incident-key"),
        incidentLandingTestId("incident-title"),
        incidentLandingTestId("create-status"),
      ]);
      await expectNoPrivateDiagnostics(form);
      await testInfo.attach("incident-creation-accessibility-tree", {
        body: await form.ariaSnapshot(),
        contentType: "text/plain",
      });
    },
  );
  await page.keyboard.press("Escape");
  const trigger = page.getByTestId(incidentLandingTestId("create-open-button"));
  await expect(trigger).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(retry).toBeFocused();
  await page
    .getByRole("button", { name: "Account and application navigation" })
    .focus();
  await page.keyboard.press("Enter");
  await expect(
    page.getByRole("menuitem", { name: "Incidents", exact: true }),
  ).toBeFocused();
});
