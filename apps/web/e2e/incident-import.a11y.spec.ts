import { Buffer } from "node:buffer";
import { incidentImportTestId } from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleFocus } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectAndRecordContrast } from "./support/accessibility/records";
import { expectNoPrivateDiagnostics } from "./support/accessibility/status";
import {
  expectImportControlReachable,
  installImportObservationFixture,
  openImportPresentation,
} from "./support/incidents/import";
import { responseBarrier } from "./support/runtime/cleanup";

test("a11y.incident-import upload recovery progress and cancellation remain keyboard reachable", async ({
  testActorPage: page,
  testActor,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const fixture = await installImportObservationFixture(
    page,
    testActor.user_id,
  );
  const form = await openImportPresentation(page);
  const file = form.getByLabel("Incident bundle file");
  const start = form.getByRole("button", { name: "Start import" });
  await start.focus();
  await page.keyboard.press("Enter");
  await expect(file).toHaveAttribute("aria-invalid", "true");
  await expect(file).toHaveAccessibleDescription(
    /Select an incident bundle file/,
  );
  await file.focus();
  await expectVisibleFocus(file);
  const chooser = page.waitForEvent("filechooser");
  await page.keyboard.press("Space");
  await (await chooser).setFiles({
    name: "Investigation.tar",
    mimeType: "application/x-tar",
    buffer: Buffer.from("presentation fixture"),
  });
  fixture.failAdmission(true);
  const pending = responseBarrier(page);
  fixture.gateAdmission(pending.promise);
  await start.focus();
  await page.keyboard.press("Enter");
  await expect(start).toBeDisabled();
  await expect(
    page.getByRole("progressbar", { name: "Upload admission" }),
  ).not.toHaveAttribute("value");
  pending.release();
  const retry = page.getByRole("button", { name: "Retry admission" });
  await expect(retry).toBeVisible();
  for (const [width, height] of [
    [1280, 720],
    [1024, 720],
    [768, 640],
    [640, 480],
  ]) {
    await page.setViewportSize({ width: width ?? 1280, height: height ?? 720 });
    await expectImportControlReachable(page, retry);
  }
  await expectVisibleFocus(retry);
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectImportControlReachable(page, retry);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await page.setViewportSize({ width: 768, height: 640 });
      await expectImportControlReachable(page, retry);
      await expectAllInteractiveControlsNamed(page);
      await expectAndRecordContrast(page, [
        incidentImportTestId("file"),
        incidentImportTestId("admission"),
      ]);
      await expectNoPrivateDiagnostics(form);
      // A second lost response must leave keyboard recovery focused.
      await page.keyboard.press("Enter");
      await expect(retry).toBeVisible();
      await expect(retry).toBeFocused();
      fixture.failAdmission(false);
      await page.keyboard.press("Enter");
      const detail = page.getByTestId(incidentImportTestId("detail"));
      await expect(detail.getByRole("heading")).toHaveText("Queued");
      await expect(detail.getByRole("heading")).toBeFocused();
      const pause = page.getByRole("button", { name: "Pause updates" });
      await expectImportControlReachable(page, pause);
      await page.keyboard.press("Enter");
      fixture.failReads(true);
      await page.getByRole("button", { name: "Refresh job status" }).focus();
      await page.keyboard.press("Enter");
      const retryRead = page.getByRole("button", { name: "Retry observation" });
      await expect(retryRead).toBeVisible();
      await expectImportControlReachable(page, retryRead);
      fixture.failReads(false);
      await page.keyboard.press("Enter");
      const cancel = page.getByRole("button", {
        name: "Cancel import",
        exact: true,
      });
      await expect(cancel).toBeEnabled();
      await expectImportControlReachable(page, cancel);
      await expectVisibleFocus(cancel);
      const checking = responseBarrier(page);
      fixture.gateReads(checking.promise);
      await page.keyboard.press("Enter");
      await expect(cancel).toHaveAttribute("aria-busy", "true");
      await expect(cancel).toBeFocused();
      expect(fixture.cancellationBodies).toHaveLength(0);
      fixture.failReads(true);
      checking.release();
      await expect(
        page
          .getByText(
            "The action could not be confirmed. Review the current job status and retry.",
            { exact: true },
          )
          .first(),
      ).toBeVisible();
      await expect(cancel).toBeFocused();
      expect(fixture.cancellationBodies).toHaveLength(0);
      fixture.gateReads(null);
      fixture.failReads(false);
      fixture.cancellation("rejected");
      await page.keyboard.press("Enter");
      await expect(detail).toContainText("Cancellation was rejected");
      await expect(cancel).toBeFocused();
      fixture.cancellation("requested");
      await page.keyboard.press("Enter");
      await expect(detail.getByRole("heading")).toHaveText(
        "Cancellation requested",
      );
      await expect(detail.getByRole("heading")).toBeFocused();
      await expect(
        page.getByRole("button", { name: "Open imported incident" }),
      ).toHaveCount(0);
      await testInfo.attach("incident-import-accessibility-tree", {
        body: await page.locator("[data-incident-import]").ariaSnapshot(),
        contentType: "text/plain",
      });
    },
  );
});
