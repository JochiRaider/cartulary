import { Buffer } from "node:buffer";
import { expect } from "./fixtures";
import { withPresentation } from "./support/accessibility/presentation";
import {
  expectReferencePackControlReachable,
  installReferencePackPresentation,
  openReferencePacks,
} from "./support/referencePacks";
import { responseBarrier } from "./support/runtime/cleanup";
import { assertViewportVisualRegression } from "./support/visual/capture";
import { test } from "./support/visual/fixtures";

test("Capture Reference Pack catalog selection admission recovery and all job states.", async ({
  testActorPage: page,
  testActor,
}) => {
  const fixture = await installReferencePackPresentation(
    page,
    testActor.user_id,
  );
  await page.setViewportSize({ width: 1280, height: 900 });
  const panel = await openReferencePacks(page);
  await assertViewportVisualRegression(page, "reference-pack-catalog");
  await panel.getByRole("checkbox").first().check();
  await panel.getByLabel("Search reference packs").fill("process");
  await expect(
    panel.getByText("1 selected keys are outside the loaded rows:", {
      exact: true,
    }),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "reference-pack-selection");
  await panel.getByLabel("Reference pack bundle").setInputFiles({
    name: "Deployment-candidate.tar",
    mimeType: "application/x-tar",
    buffer: Buffer.from("presentation fixture"),
  });
  const pending = responseBarrier(page);
  fixture.gateAdmission(pending.promise);
  fixture.failAdmission(true);
  await panel.getByRole("button", { name: "Import", exact: true }).click();
  await expect(
    panel.getByText("Submitted; awaiting server acknowledgment.", {
      exact: true,
    }),
  ).toBeVisible();
  await assertViewportVisualRegression(page, "reference-pack-submitted");
  pending.release();
  const replay = panel.getByRole("button", { name: "Retry exact request" });
  await expect(replay).toBeVisible();
  await page.setViewportSize({ width: 640, height: 480 });
  await expectReferencePackControlReachable(page, replay);
  await assertViewportVisualRegression(page, "reference-pack-recovery-short");
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectReferencePackControlReachable(page, replay);
    await assertViewportVisualRegression(page, "reference-pack-recovery-zoom");
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await expectReferencePackControlReachable(page, replay);
      await assertViewportVisualRegression(
        page,
        "reference-pack-recovery-spacing",
      );
    },
  );
  await page.setViewportSize({ width: 1280, height: 900 });
  fixture.failAdmission(false);
  await replay.click();
  const operations = panel.getByRole("region", {
    name: "Operations in this session",
  });
  await expect(operations.getByText("Queued", { exact: true })).toBeVisible();
  await operations.scrollIntoViewIfNeeded();
  await assertViewportVisualRegression(page, "reference-pack-queued");
  fixture.setStatus("running");
  await expect(operations.getByText("Running", { exact: true })).toBeVisible();
  await assertViewportVisualRegression(page, "reference-pack-running");
  fixture.failReads(true);
  await expect(
    panel.getByRole("button", { name: "Retry observation" }),
  ).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "reference-pack-observation-recovery",
  );
  fixture.failReads(false);
  await panel.getByRole("button", { name: "Retry observation" }).click();
  await expect(
    panel.getByRole("button", { name: "Retry observation" }),
  ).toHaveCount(0);
  await panel.getByRole("button", { name: "Cancel operation" }).click();
  await expect(
    operations.getByText("Cancellation requested", { exact: true }),
  ).toBeVisible();
  await expect(
    operations.getByText(
      "Cancellation acknowledged. This does not establish cancellation or rollback.",
      { exact: true },
    ),
  ).toBeVisible();
  await operations.evaluate((element) =>
    element.scrollIntoView({ block: "start", behavior: "instant" }),
  );
  await assertViewportVisualRegression(page, "reference-pack-cancel-requested");
  fixture.setStatus("canceled");
  await expect(operations.getByText("Canceled", { exact: true })).toBeVisible();
  await operations.evaluate((element) =>
    element.scrollIntoView({ block: "start", behavior: "instant" }),
  );
  await assertViewportVisualRegression(page, "reference-pack-canceled");
  for (const [status, label, capture] of [
    ["succeeded", "Succeeded", "reference-pack-succeeded"],
    ["failed", "Failed", "reference-pack-failed"],
  ] as const) {
    await panel
      .getByRole("button", { name: "Refresh selected", exact: true })
      .click();
    await expect(operations.getByText("Queued", { exact: true })).toBeVisible();
    fixture.setStatus(status);
    await expect(operations.getByText(label, { exact: true })).toBeVisible();
    await operations.getByText(label, { exact: true }).scrollIntoViewIfNeeded();
    await assertViewportVisualRegression(page, capture);
  }
});
