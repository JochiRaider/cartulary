import { Buffer } from "node:buffer";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import {
  expectReferencePackControlReachable,
  installReferencePackPresentation,
  openReferencePacks,
} from "./support/referencePacks";
import { responseBarrier } from "./support/runtime/cleanup";

test("a11y.reference-packs keyboard selection upload observation and cancellation recovery remain reachable", async ({
  testActorPage: page,
  testActor,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const fixture = await installReferencePackPresentation(
    page,
    testActor.user_id,
  );
  const panel = await openReferencePacks(page);
  const file = panel.getByLabel("Reference pack bundle");
  await expectReferencePackControlReachable(page, file);
  await expectVisibleFocus(file);
  const chooser = page.waitForEvent("filechooser");
  await page.keyboard.press("Space");
  await (await chooser).setFiles({
    name: "Candidate.tar",
    mimeType: "application/x-tar",
    buffer: Buffer.from("presentation bytes"),
  });
  await page.keyboard.press("Tab");
  const start = panel.getByRole("button", { name: "Import", exact: true });
  await expect(start).toBeFocused();
  const pending = responseBarrier(page);
  fixture.gateAdmission(pending.promise);
  fixture.failAdmission(true);
  await page.keyboard.press("Enter");
  await expect(file).toBeDisabled();
  pending.release();
  const retry = panel.getByRole("button", { name: "Retry exact request" });
  await expect(retry).toBeVisible();
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 1024, height: 720 },
    { width: 768, height: 640 },
    { width: 640, height: 480 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    await expectReferencePackControlReachable(page, retry);
    await expectVisibleFocus(retry);
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectReferencePackControlReachable(page, retry);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await expectReferencePackControlReachable(page, retry);
      await expectAllInteractiveControlsNamed(page);
      const failedRetry = responseBarrier(page);
      fixture.gateAdmission(failedRetry.promise);
      const admissionsBeforeRetry = fixture.admissions.length;
      await page.keyboard.press("Enter");
      await expect(retry).toHaveAttribute("aria-disabled", "true");
      await expect
        .poll(() => fixture.admissions.length)
        .toBe(admissionsBeforeRetry + 1);
      failedRetry.release();
      await expect(retry).toHaveAttribute("aria-disabled", "false");
      await expect(retry).toBeFocused();
      fixture.gateAdmission(null);
      fixture.failAdmission(false);
      fixture.failReads(true);
      await page.keyboard.press("Enter");
      const observation = panel.getByRole("button", {
        name: "Retry observation",
      });
      await expect(observation).toBeVisible();
      await expectReferencePackControlReachable(page, observation);
      fixture.failReads(false);
      await page.keyboard.press("Enter");
      await expect(panel.getByRole("progressbar")).not.toHaveAttribute("value");
      const cancel = panel.getByRole("button", { name: "Cancel operation" });
      await expectReferencePackControlReachable(page, cancel);
      await expectVisibleFocus(cancel);
      await page.keyboard.press("Enter");
      await expect(
        panel.getByText("Cancellation requested", { exact: true }),
      ).toBeVisible();
      await expect(cancel).toHaveCount(0);
      fixture.setStatus("canceled");
      const dismiss = panel.getByRole("button", { name: "Dismiss operation" });
      await expect(dismiss).toBeVisible();
      await expectReferencePackControlReachable(page, dismiss);
      await page.keyboard.press("Enter");
      await expect(dismiss).toHaveCount(0);
      const selected = panel.getByRole("checkbox").first();
      await expectReferencePackControlReachable(page, selected);
      await page.keyboard.press("Space");
      await expect(selected).toBeChecked();
      await expect(panel.getByRole("checkbox").nth(1)).toBeChecked();
      for (const viewport of [
        { width: 390, height: 480 },
        { width: 640, height: 480 },
        { width: 768, height: 640 },
      ]) {
        await page.setViewportSize(viewport);
        const activate = panel
          .getByRole("button", { name: "Activate", exact: true })
          .first();
        await expectReferencePackControlReachable(page, activate);
        await expectVisibleFocus(activate);
        const label = panel
          .locator("td .rp-cell-label")
          .filter({ hasText: "Version actions" })
          .first();
        await expect(label).toBeVisible();
        const box = await label.boundingBox();
        expect(box?.width).toBeGreaterThan(50);
        expect(box?.height).toBeLessThan(100);
      }
      await testInfo.attach("reference-pack-accessibility-tree", {
        body: await panel.ariaSnapshot(),
        contentType: "text/plain",
      });
    },
  );
});
