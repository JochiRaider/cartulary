import { Buffer } from "node:buffer";
// biome-ignore lint/correctness/noUndeclaredDependencies: the workspace root pins the browser accessibility engine used by the Make-owned harness.
import AxeBuilder from "@axe-core/playwright";
import type {
  EvidenceCreateRequest,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import {
  evidenceAccessMessageTestId,
  evidenceAccessStateTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridScrollportSelector,
  gridShellTestId,
  timelineEvidenceFileInputTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import {
  expectNoFocusTrap,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import {
  expectAndRecordContrast,
  expectLocatorContrast,
} from "./support/accessibility/records";
import { expectNoPrivateDiagnostics } from "./support/accessibility/status";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  createEvidenceFixtureRow,
  createUploadedEvidenceFixture,
  type EvidenceUploadOptions,
} from "./support/evidence/fixtures";
import { chooseEvidenceFile } from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import { responseBarrier } from "./support/runtime/cleanup";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  awaitResponseBarrier,
  installOwnedRoute,
  installPublicErrorResponse,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import { expectDecisionControlReachable } from "./support/workbook/decisionSupersession";
import { mountedGridTarget } from "./support/workbook/keyboardGrid";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";
import {
  openTimelineEvidenceFixture,
  retainTimelineEvidencePartialResult,
} from "./support/workbook/timelineRelatedEvidence";

function evidenceAccessStateContainer(page: Page, recordId: string): Locator {
  return page.getByTestId(evidenceAccessStateTestId(recordId));
}

type EvidenceA11yRowStateKey =
  | "available"
  | "quarantined"
  | "pending_receipt"
  | "requested";

async function expectEvidenceControlsPainted(page: Page, recordId: string) {
  await mountedGridTarget(
    page,
    evidenceViewSchemaId,
    evidenceAccessStateTestId(recordId),
  );
  const shell = page.getByTestId(gridShellTestId(evidenceViewSchemaId));
  const shellBox = await shell.boundingBox();
  const scrollportBox = await shell
    .locator(gridScrollportSelector())
    .boundingBox();
  expect(scrollportBox?.width).toBeLessThanOrEqual((shellBox?.width ?? 0) + 1);
  const state = page.getByTestId(evidenceAccessStateTestId(recordId));
  // Access and attachment are siblings within the same horizontally scrolling cell.
  await state.locator("..").scrollIntoViewIfNeeded();
  const result = await state.locator("..").evaluate((element) =>
    Array.from(element.querySelectorAll("button")).map((button) => {
      const box = button.getBoundingClientRect();
      const hit = document.elementFromPoint(
        box.x + box.width / 2,
        box.y + box.height / 2,
      );
      let unclipped =
        box.width > 0 &&
        box.height > 0 &&
        (hit === button || button.contains(hit));
      const clipping: string[] = [];
      for (
        let parent = button.parentElement;
        parent;
        parent = parent.parentElement
      ) {
        const style = getComputedStyle(parent);
        const clip = parent.getBoundingClientRect();
        if (
          ["hidden", "clip", "auto", "scroll"].includes(style.overflowY) &&
          (box.top < clip.top - 1 || box.bottom > clip.bottom + 1)
        ) {
          unclipped = false;
          clipping.push(
            `block ${parent.tagName}: ${box.top}..${box.bottom} outside ${clip.top}..${clip.bottom}`,
          );
        }
        if (
          ["hidden", "clip", "auto", "scroll"].includes(style.overflowX) &&
          (box.left < clip.left - 1 || box.right > clip.right + 1)
        ) {
          unclipped = false;
          clipping.push(
            `inline ${parent.tagName}: ${box.left}..${box.right} outside ${clip.left}..${clip.right}`,
          );
        }
      }
      return { label: button.textContent, unclipped, clipping };
    }),
  );
  expect(result).toHaveLength(4);
  expect(
    result.filter((button) => !button.unclipped),
    "Every production Evidence control must be painted without clipping",
  ).toEqual([]);
}

async function expectEvidenceAccessState(
  page: Page,
  recordId: string,
  stateKey: EvidenceA11yRowStateKey,
  options: {
    messageText?: RegExp | string;
  } = {},
) {
  await mountedGridTarget(
    page,
    evidenceViewSchemaId,
    evidencePreviewButtonTestId(recordId),
  );
  await expect(evidenceAccessStateContainer(page, recordId)).toHaveAttribute(
    "data-evidence-state-key",
    stateKey,
  );
  const previewButton = page.getByTestId(evidencePreviewButtonTestId(recordId));
  const downloadButton = page.getByTestId(
    evidenceDownloadButtonTestId(recordId),
  );
  await expect(previewButton).toHaveText("Preview");
  await expect(downloadButton).toHaveText("Download");
  if (stateKey === "available") {
    await expect(previewButton).toBeEnabled();
    await expect(downloadButton).toBeEnabled();
  } else {
    await expect(previewButton).toBeDisabled();
    await expect(downloadButton).toBeDisabled();
  }
  if (options.messageText !== undefined) {
    const message = page.getByTestId(evidenceAccessMessageTestId(recordId));
    await expect(message).toBeVisible();
    if (options.messageText !== undefined) {
      await expect(message).toContainText(options.messageText);
    }
  }
}

async function createUploadedA11yEvidence(
  page: Page,
  incidentId: string,
  options: EvidenceUploadOptions,
): Promise<ViewRow> {
  return createUploadedEvidenceFixture(page, incidentId, {
    collectorPartyText: "browser.evidence-workflow accessibility fixture",
    ...options,
  });
}

async function createA11yEvidenceRow(
  page: Page,
  incidentId: string,
  options: {
    lifecycleState: NonNullable<
      EvidenceCreateRequest["evidence.lifecycle_state"]
    >;
    requestedAt: string;
    storageRef: string;
    title: string;
    txnPrefix: string;
  },
): Promise<ViewRow> {
  return createEvidenceFixtureRow(page, incidentId, {
    collectorPartyText: "browser.evidence-workflow accessibility fixture",
    ...options,
  });
}

test("a11y.Evidence access states preview and download expose names and announced outcomes", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YEVIDENCEWORKFLOW"),
    "a11y.evidence-workflow.row-01 evidence access",
  );
  const requested = await createA11yEvidenceRow(page, incidentId, {
    lifecycleState: "requested",
    requestedAt: "2026-06-07T10:00:00Z",
    storageRef: "case://a11y.evidence-workflow/requested",
    title: "01 requested evidence",
    txnPrefix: "a11y.evidence-workflow-requested",
  });
  const pending = await createA11yEvidenceRow(page, incidentId, {
    lifecycleState: "pending_receipt",
    requestedAt: "2026-06-07T10:05:00Z",
    storageRef: "case://a11y.evidence-workflow/pending",
    title: "02 pending evidence",
    txnPrefix: "a11y.evidence-workflow-pending",
  });
  const blocked = await createA11yEvidenceRow(page, incidentId, {
    lifecycleState: "quarantined",
    requestedAt: "2026-06-07T10:10:00Z",
    storageRef: "case://a11y.evidence-workflow/quarantined",
    title: "03 quarantined evidence",
    txnPrefix: "a11y.evidence-workflow-blocked",
  });
  const availablePreview = await createUploadedA11yEvidence(page, incidentId, {
    body: Buffer.from("a11y.evidence-workflow preview evidence\n", "utf8"),
    contentType: "text/plain",
    filename: "a11y.evidence-workflow-preview.txt",
    requestedAt: "2026-06-07T10:15:00Z",
    title: "04 available preview evidence",
    txnPrefix: "a11y.evidence-workflow-preview",
  });
  const downloadHandle = await createUploadedA11yEvidence(page, incidentId, {
    body: Buffer.from("a11y.evidence-workflow download evidence\n", "utf8"),
    contentType: "text/plain",
    filename: "a11y.evidence-workflow-download.txt",
    requestedAt: "2026-06-07T10:20:00Z",
    title: "05 download handle evidence",
    txnPrefix: "a11y.evidence-workflow-download",
  });
  const previewBlocked = await createUploadedA11yEvidence(page, incidentId, {
    body: Buffer.from(
      "<!doctype html><title>a11y.evidence-workflow unsupported preview</title>",
      "utf8",
    ),
    contentType: "text/html",
    filename: "a11y.evidence-workflow-preview-blocked.html",
    requestedAt: "2026-06-07T10:25:00Z",
    title: "06 preview blocked evidence",
    txnPrefix: "a11y.evidence-workflow-preview-blocked",
  });
  const failedHandle = await createUploadedA11yEvidence(page, incidentId, {
    body: Buffer.from(
      "a11y.evidence-workflow failed handle evidence\n",
      "utf8",
    ),
    contentType: "text/plain",
    filename: "a11y.evidence-workflow-failed.txt",
    requestedAt: "2026-06-07T10:30:00Z",
    title: "07 failed handle evidence",
    txnPrefix: "a11y.evidence-workflow-failed",
  });
  const inconsistentHandle = await createUploadedA11yEvidence(
    page,
    incidentId,
    {
      body: Buffer.from(
        "a11y.evidence-workflow inconsistent handle evidence\n",
        "utf8",
      ),
      contentType: "text/plain",
      filename: "a11y.evidence-workflow-inconsistent.txt",
      requestedAt: "2026-06-07T10:35:00Z",
      title: "08 inconsistent handle evidence",
      txnPrefix: "a11y.evidence-workflow-inconsistent",
    },
  );
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      evidenceViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
  ).toBeVisible();
  await expectEvidenceAccessState(page, requested.record_id, "requested", {
    messageText: "Requested",
  });
  await expectEvidenceAccessState(page, pending.record_id, "pending_receipt", {
    messageText: "Pending receipt",
  });
  await expectEvidenceAccessState(page, blocked.record_id, "quarantined", {
    messageText: "Quarantined",
  });
  await expectAndRecordContrast(
    page,
    [
      evidenceAccessMessageTestId(requested.record_id),
      evidenceAccessMessageTestId(pending.record_id),
      evidenceAccessMessageTestId(blocked.record_id),
    ],
    "unavailable-states",
  );
  await expectEvidenceAccessState(
    page,
    availablePreview.record_id,
    "available",
  );
  const paintedAction = await page
    .getByTestId(evidencePreviewButtonTestId(availablePreview.record_id))
    .evaluate((button) => {
      const box = button.getBoundingClientRect();
      let parent = button.parentElement;
      while (parent) {
        const style = getComputedStyle(parent);
        const clip = parent.getBoundingClientRect();
        if (
          ["hidden", "clip", "auto", "scroll"].includes(style.overflowY) &&
          (box.top < clip.top - 1 || box.bottom > clip.bottom + 1)
        )
          return false;
        parent = parent.parentElement;
      }
      return box.height > 0;
    });
  expect(
    paintedAction,
    "Evidence Preview must fit its painted row bounds",
  ).toBe(true);
  const previewButton = page.getByTestId(
    evidencePreviewButtonTestId(availablePreview.record_id),
  );
  await expectVisibleFocus(
    page.getByTestId(evidenceDownloadButtonTestId(availablePreview.record_id)),
  );
  await expectVisibleFocus(previewButton);
  await previewButton.click();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(availablePreview.record_id)),
  ).toBeVisible();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(availablePreview.record_id)),
  ).toHaveAttribute("title", /Evidence preview/u);
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(availablePreview.record_id)),
  ).not.toHaveAttribute("aria-live");
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(availablePreview.record_id)),
  ).toHaveText("Preview open");
  await expectAndRecordContrast(
    page,
    [
      evidenceAccessMessageTestId(availablePreview.record_id),
      evidencePreviewButtonTestId(availablePreview.record_id),
    ],
    "preview-open",
  );
  await page
    .getByTestId(evidencePreviewPanelTestId())
    .getByRole("button", { name: "Close" })
    .click();
  await expectEvidenceAccessState(page, downloadHandle.record_id, "available");
  const downloadButton = page.getByTestId(
    evidenceDownloadButtonTestId(downloadHandle.record_id),
  );
  await expectVisibleFocus(
    page.getByTestId(evidencePreviewButtonTestId(downloadHandle.record_id)),
  );
  await expectVisibleFocus(downloadButton);
  const downloadPromise = page.waitForEvent("download");
  await downloadButton.click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toBe(
    "a11y.evidence-workflow-download.txt",
  );
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
  ).not.toHaveAttribute("aria-live");
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
  ).toHaveText("Download requested");
  await expectAndRecordContrast(
    page,
    [
      evidenceAccessMessageTestId(downloadHandle.record_id),
      evidenceDownloadButtonTestId(downloadHandle.record_id),
    ],
    "download-requested",
  );
  await expectEvidenceAccessState(page, previewBlocked.record_id, "available");
  const unsupportedPreviewButton = page.getByTestId(
    evidencePreviewButtonTestId(previewBlocked.record_id),
  );
  await expectVisibleFocus(unsupportedPreviewButton);
  await expectVisibleFocus(
    page.getByTestId(evidenceDownloadButtonTestId(previewBlocked.record_id)),
  );
  await unsupportedPreviewButton.click();
  const previewBlockedMessage = page.getByTestId(
    evidenceAccessMessageTestId(previewBlocked.record_id),
  );
  await expect(previewBlockedMessage).not.toHaveAttribute("aria-live");
  await expect(
    page.getByRole("status").filter({ hasText: /evidence:/u }),
  ).toHaveCount(1);
  await expect(previewBlockedMessage).toContainText("No preview");
  await expectNoPrivateDiagnostics(previewBlockedMessage);
  await expectAndRecordContrast(
    page,
    [evidenceAccessMessageTestId(previewBlocked.record_id)],
    "preview-blocked",
  );
  await installPublicErrorResponse(page, {
    method: "POST",
    status: 409,
    code: "evidence_access_unavailable",
    path: `/api/v1/evidence-records/${failedHandle.record_id}/preview-handle`,
    details: { reason_code: "blob_failed" },
  });
  await (
    await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidencePreviewButtonTestId(failedHandle.record_id),
    )
  ).click();
  const failedMessage = page.getByTestId(
    evidenceAccessMessageTestId(failedHandle.record_id),
  );
  await expect(failedMessage).not.toHaveAttribute("aria-live");
  await expect(
    page.getByRole("status").filter({ hasText: /evidence:/u }),
  ).toHaveCount(1);
  await expect(failedMessage).toContainText("Upload failed");
  await expectNoPrivateDiagnostics(failedMessage);
  await expectAndRecordContrast(
    page,
    [evidenceAccessMessageTestId(failedHandle.record_id)],
    "failed-handle",
  );
  await installPublicErrorResponse(page, {
    method: "POST",
    status: 409,
    code: "evidence_access_unavailable",
    path: `/api/v1/evidence-records/${inconsistentHandle.record_id}/preview-handle`,
    details: { reason_code: "evidence_inconsistent" },
  });
  await (
    await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidencePreviewButtonTestId(inconsistentHandle.record_id),
    )
  ).click();
  const inconsistentMessage = page.getByTestId(
    evidenceAccessMessageTestId(inconsistentHandle.record_id),
  );
  await expect(inconsistentMessage).not.toHaveAttribute("aria-live");
  await expect(
    page.getByRole("status").filter({ hasText: /evidence:/u }),
  ).toHaveCount(1);
  await expect(inconsistentMessage).toContainText("Inconsistent");
  await expectNoPrivateDiagnostics(inconsistentMessage);
  await expectAndRecordContrast(
    page,
    [evidenceAccessMessageTestId(inconsistentHandle.record_id)],
    "inconsistent-handle",
  );
  await expectEvidenceControlsPainted(page, availablePreview.record_id);
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
});

test("a11y.Evidence preview close and pending cancellation preserve focus across density and layout", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YEVIDENCEWORKFLOW"),
    "a11y.evidence-workflow.row-01 evidence access",
  );
  const availablePreview = await createUploadedA11yEvidence(page, incidentId, {
    body: Buffer.from("a11y.evidence-workflow preview evidence\n", "utf8"),
    contentType: "text/plain",
    filename: "a11y.evidence-workflow-preview.txt",
    requestedAt: "2026-06-07T10:15:00Z",
    title: "04 available preview evidence",
    txnPrefix: "a11y.evidence-workflow-preview",
  });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      evidenceViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
  ).toBeVisible();
  const preferences = await page.request.get(
    `${apiBase}/api/v1/account/preferences`,
  );
  const originalPreferences = (await preferences.json()).data;
  const setDensity = async (density: string) => {
    const response = await page.request.get(
      `${apiBase}/api/v1/account/preferences`,
    );
    const current = (await response.json()).data;
    const updated = await page.request.put(
      `${apiBase}/api/v1/account/preferences`,
      {
        headers: await csrfHeaders(page),
        data: {
          base_preferences_version: current.preferences_version,
          client_txn_id: uniqueTxn("evidence-access-density"),
          density_mode: density,
        },
      },
    );
    expect(updated.ok()).toBeTruthy();
    await page.reload();
  };
  try {
    for (const density of ["compact", "default", "comfortable"]) {
      await setDensity(density);
      await expectEvidenceControlsPainted(page, availablePreview.record_id);
      await withPresentation(
        page,
        {
          textSpacing: `* { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; } p { margin-block-end: 2em !important; }`,
        },
        async () => {
          await expectEvidenceControlsPainted(page, availablePreview.record_id);
        },
      );
    }
    await setDensity("default");
    for (const width of [1440, 1024, 768]) {
      await page.setViewportSize({
        width,
        height: width >= 1280 ? 900 : width >= 1024 ? 720 : 640,
      });
      await expectEvidenceControlsPainted(page, availablePreview.record_id);
      const state = page.getByTestId(
        evidenceAccessMessageTestId(availablePreview.record_id),
      );
      await state.focus();
      await state.press("Enter");
      const inspectorPreview = page.getByTestId(
        evidencePreviewButtonTestId(availablePreview.record_id, "inspector"),
      );
      await expect(inspectorPreview).toBeVisible();
      await inspectorPreview.focus();
      await inspectorPreview.press("Enter");
      const panel = page.getByTestId(evidencePreviewPanelTestId());
      const close = panel.getByRole("button", { name: "Close" });
      await expect(close).toBeVisible();
      await close.focus();
      await expect(close).toBeFocused();
      await close.press("Escape");
      await expect(panel).toHaveCount(0);
      await expect(inspectorPreview).toBeFocused();
      if (width === 768) {
        await installPublicErrorResponse(page, {
          method: "POST",
          status: 409,
          code: "evidence_access_unavailable",
          path: `/api/v1/evidence-records/${availablePreview.record_id}/preview-handle`,
          details: { reason_code: "unsupported_preview" },
        });
        await inspectorPreview.press("Enter");
        await expect(
          page.getByTestId(
            evidenceAccessMessageTestId(
              availablePreview.record_id,
              "inspector",
            ),
          ),
        ).toHaveText("This file type cannot be previewed.");
        await expect(
          page
            .getByRole("status")
            .filter({ hasText: "This file type cannot be previewed." }),
        ).toHaveCount(1);
        await expect.soft(inspectorPreview).toBeFocused();
      }
      await inspectorPreview.press("Escape");
      await expect(inspectorPreview).toHaveCount(0);
    }
    await page.setViewportSize({ width: 1440, height: 900 });
    await withPresentation(page, { zoom: "200%" }, async () => {
      await expectEvidenceControlsPainted(page, availablePreview.record_id);
    });
    await page.setViewportSize({ width: 1440, height: 900 });
    const availableState = page.getByTestId(
      evidenceAccessStateTestId(availablePreview.record_id),
    );
    const attach = availableState.locator("..").getByRole("button", {
      name: /Attach file/u,
    });
    await attach.focus();
    const chooserPromise = page.waitForEvent("filechooser");
    await attach.press("Enter");
    await (await chooserPromise).setFiles([]);
    const heldBarrier = responseBarrier(page);
    const release = heldBarrier.release;
    const held = heldBarrier.promise;
    const completedBarrier = responseBarrier(page);
    const finished = completedBarrier.release;
    const completed = completedBarrier.promise;
    const previewPath = `**/api/v1/evidence-records/${availablePreview.record_id}/preview-handle`;
    await installOwnedRoute(page, previewPath, async (route) => {
      const response = await route.fetch();
      if (!(await awaitResponseBarrier(route, held))) return;
      await route.fulfill({ response });
      finished();
    });
    const preview = page.getByTestId(
      evidencePreviewButtonTestId(availablePreview.record_id),
    );
    await preview.focus();
    await preview.press("Enter");
    const pendingPanel = page.getByTestId(evidencePreviewPanelTestId());
    await expect(pendingPanel).toContainText("Opening preview…");
    await pendingPanel.getByRole("button", { name: "Close" }).click();
    release();
    await completed;
    await expect(pendingPanel).toHaveCount(0);
    await expect(preview).toBeFocused();
    await removeOwnedRoute(page, previewPath);
  } finally {
    await setDensity(originalPreferences.density_mode);
  }
});

test("a11y.closed incidents retain readable preview and download controls", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YEVIDENCEWORKFLOW"),
    "a11y.evidence-workflow.row-01 evidence access",
  );
  const availablePreview = await createUploadedA11yEvidence(page, incidentId, {
    body: Buffer.from("a11y.evidence-workflow preview evidence\n", "utf8"),
    contentType: "text/plain",
    filename: "a11y.evidence-workflow-preview.txt",
    requestedAt: "2026-06-07T10:15:00Z",
    title: "04 available preview evidence",
    txnPrefix: "a11y.evidence-workflow-preview",
  });
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      evidenceViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
  ).toBeVisible();
  const incidentResponse = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}`,
  );
  const incident = (await incidentResponse.json()).data;
  const closed = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/close`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_incident_version: incident.incident_version,
        client_txn_id: uniqueTxn("evidence-access-close"),
        reason: "Evidence read access verification",
      },
    },
  );
  expect(closed.ok()).toBeTruthy();
  await page.reload();
  await expectEvidenceControlsPainted(page, availablePreview.record_id);
  await expect(
    page.getByTestId(evidenceAttachFileInputTestId(availablePreview.record_id)),
  ).toBeDisabled();
  await expect(
    page.getByTestId(evidencePreviewButtonTestId(availablePreview.record_id)),
  ).toBeEnabled();
  await expect(
    page.getByTestId(evidenceDownloadButtonTestId(availablePreview.record_id)),
  ).toBeEnabled();
  await page
    .getByTestId(evidencePreviewButtonTestId(availablePreview.record_id))
    .click();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(availablePreview.record_id)),
  ).toBeVisible();
  await page
    .getByTestId(evidencePreviewPanelTestId())
    .getByRole("button", { name: "Close" })
    .click();
  const closedDownload = page.waitForEvent("download");
  await page
    .getByTestId(evidenceDownloadButtonTestId(availablePreview.record_id))
    .click();
  expect((await closedDownload).suggestedFilename()).toBe(
    "a11y.evidence-workflow-preview.txt",
  );
});

test("a11y.timeline-related-evidence metadata references and retained partial success remain keyboard accessible", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const { form, party } = await openTimelineEvidenceFixture(page);
  const collector = form.getByTestId(
    genericCreateFieldTestId("evidence.collector_party_text"),
  );
  const timestamp = form.getByTestId(
    genericCreateFieldTestId("evidence.received_at"),
  );
  const lifecycle = form.getByTestId(
    genericCreateFieldTestId("evidence.lifecycle_state"),
  );
  const submit = form.getByTestId(
    genericCreateSubmitTestId(evidenceViewSchemaId),
  );
  await expect(lifecycle).toHaveValue("");
  await expect(lifecycle.locator("option:checked")).toHaveText(
    "Use server default (Requested)",
  );
  await timestamp.fill("unfinished timestamp");
  await submit.focus();
  await submit.press("Enter");
  await expect(timestamp).toHaveAttribute("aria-invalid", "true");
  await expect(timestamp).toHaveAccessibleDescription(
    /timestamp|RFC3339|date/i,
  );
  await timestamp.fill("");
  const choose = form.getByRole("button", {
    name: "Choose Collector Party",
    exact: true,
  });
  await choose.focus();
  await choose.press("Enter");
  const picker = form.getByRole("region", {
    name: "Choose Collector Party",
    exact: true,
  });
  const candidates = picker.getByRole("group", {
    name: "Collector Party",
    exact: true,
  });
  await expect(
    picker.getByRole("button", { name: "Apply Party", exact: true }),
  ).toBeEnabled();
  await candidates
    .locator(`input[type="radio"][value="${party.record_id}"]`)
    .check();
  await candidates.press("Escape");
  await expect(choose).toBeFocused();
  await choose.press("Enter");
  await expect(candidates.getByRole("radio", { checked: true })).toHaveCount(0);
  await candidates.press("Escape");
  await expect(collector).toHaveValue("Response team collection log");
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [collector, timestamp, lifecycle, choose, submit]) {
      await expectDecisionControlReachable(page, control);
      await expectVisibleFocus(control);
    }
    await testInfo.attach(`timeline-evidence-${viewport.width}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectDecisionControlReachable(page, submit);
  });
  await withPresentation(
    page,
    {
      textSpacing:
        "* { line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important; } p { margin-bottom: 2em !important; }",
    },
    async () => {
      await expectDecisionControlReachable(page, choose);
    },
  );
  const { recovery, summary } = await retainTimelineEvidencePartialResult(page);
  const review = recovery.getByRole("button", {
    name: "Review original Timeline link",
    exact: true,
  });
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    await expectDecisionControlReachable(page, review);
    await expectVisibleFocus(review);
  }
  await review.press("Enter");
  await expect(review).toBeFocused();
  const link = recovery.getByRole("button", {
    name: "Link created Evidence",
    exact: true,
  });
  await expectDecisionControlReachable(page, link);
  await expectAllInteractiveControlsNamed(page);
  await testInfo.attach("timeline-evidence-recovery-tree", {
    body: await recovery.ariaSnapshot(),
    contentType: "text/plain",
  });
  await link.press("Escape");
  await expect(summary).toBeFocused();
  await expect(recovery).not.toBeVisible();
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
});

test("a11y.evidence-file-recovery keeps stage recovery local reachable and keyboard operable", async ({
  page,
}, info) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("A11Y-EUR"),
    "Evidence file recovery accessibility",
  );
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "timeline.activity_synopsis_text": "Original recovery source",
  });
  let slots = 0;
  await installOwnedRoute(page, "**/api/v1/object-blobs", async (route) => {
    slots++;
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    if (slots === 1) await route.abort("failed");
    else await route.fulfill({ response });
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${timelineViewSchemaId}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await openTimelineInspector(page, source.record_id);
  await chooseEvidenceFile(
    page.getByTestId(timelineEvidenceFileInputTestId(source.record_id)),
    {
      name: "local-recovery-with-a-long-original-filename-for-narrow-workbook-controls.txt",
      mimeType: "text/plain",
      buffer: Buffer.from("Recovery file"),
    },
  );
  const recovery = page.getByRole("group", {
    name: "Inspector file recovery: local-recovery-with-a-long-original-filename-for-narrow-workbook-controls.txt",
    exact: true,
    includeHidden: true,
  });
  await expect(recovery.getByRole("status")).toContainText(
    "Upload preparation is uncertain",
  );
  await expect(recovery.getByRole("status")).toHaveAttribute(
    "aria-live",
    "off",
  );
  const resume = recovery.getByRole("button", { name: "Resume", exact: true });
  const discard = recovery.getByRole("button", {
    name: "Discard retained file work",
    exact: true,
  });
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    for (const control of [resume, discard]) {
      await expectDecisionControlReachable(page, control);
      await expectVisibleFocus(control);
    }
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await info.attach(`evidence-file-recovery-${viewport.width}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectDecisionControlReachable(page, resume);
  });
  await expectLocatorContrast(resume, "file-recovery-resume");
  await expectAndRecordContrast(page, [
    {
      id: "file-recovery-status",
      selector:
        '[aria-label="Inspector file recovery: local-recovery-with-a-long-original-filename-for-narrow-workbook-controls.txt"] [role="status"]',
    },
  ]);
  await resume.focus();
  await resume.press("Enter");
  await expect(recovery).toContainText("Evidence attached.");
  expect(slots).toBe(2);
  await expect(discard).toHaveCount(0);
  // The action disappears on completion: focus stays local, never on body.
  expect(await page.evaluate(() => document.activeElement?.tagName)).not.toBe(
    "BODY",
  );
  const history = page.getByText("Completed attachments (1)", { exact: true });
  await history.focus();
  await history.press("Enter");
  await expect(recovery).toBeVisible();
  await history.press("Enter");
  await expect(recovery).toBeHidden();
  await expect(
    page.getByRole("button", { name: /^Attachments:/ }),
  ).toHaveAttribute("aria-expanded", "false");
  expect(slots).toBe(2);
  const rows = await queryViewRows(page, incident, timelineViewSchemaId);
  expect(
    rows.find((row) => row.record_id === source.record_id)?.cells[
      "timeline.evidence_count"
    ]?.value,
  ).toBe(1);
});

test("a11y.evidence-verification exposes compact progress and keyboard cancellation across narrow layouts", async ({
  page,
}, info) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("A11Y-VERIFY"),
    "Evidence verification accessibility",
  );
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "timeline.activity_synopsis_text": "Verification source",
  });
  let release = () => {};
  const gateBarrier = responseBarrier(page);
  release = gateBarrier.release;
  const gate = gateBarrier.promise;
  await installOwnedRoute(page, "**/api/v1/jobs/*", async (route) => {
    if (route.request().method() !== "GET") return route.continue();
    const response = await route.fetch();
    if (!(await awaitResponseBarrier(route, gate))) return;
    await route.fulfill({ response });
  });
  try {
    await page.goto(
      `/?incident_id=${incident}&view_schema_id=${timelineViewSchemaId}`,
    );
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await openTimelineInspector(page, source.record_id);
    const filename =
      "verification-with-a-long-original-filename-for-narrow-workbook-controls.txt";
    await chooseEvidenceFile(
      page.getByTestId(timelineEvidenceFileInputTestId(source.record_id)),
      {
        name: filename,
        mimeType: "text/plain",
        buffer: Buffer.from("Verification file"),
      },
    );
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .click();
    const surface = page.getByRole("region", {
      name: "Timeline attachments",
      exact: true,
    });
    const cancel = surface.getByRole("button", {
      name: "Cancel verification",
      exact: true,
    });
    await expect(cancel).toBeVisible();
    for (const viewport of [
      { width: 1280, height: 720 },
      { width: 768, height: 640 },
      { width: 390, height: 480 },
    ]) {
      await page.setViewportSize(viewport);
      await expectDecisionControlReachable(page, cancel);
      await expectVisibleFocus(cancel);
      await expect(
        surface.getByRole("progressbar", {
          name: "File verification progress",
        }),
      ).toBeVisible();
      await expect(
        surface.getByRole("button", { name: /^Attachments:/ }),
      ).toHaveAttribute("aria-expanded", "false");
      await info.attach(`verification-progress-${viewport.width}`, {
        body: await page.screenshot({ animations: "disabled", caret: "hide" }),
        contentType: "image/png",
      });
    }
    await expectAndRecordContrast(page, [
      {
        id: "verification-cancel",
        selector:
          '[aria-label="Timeline attachments"] [data-job-cancel="true"]',
      },
      {
        id: "verification-status",
        selector:
          '[aria-label="Timeline attachments"] [data-job-progress="true"]',
      },
    ]);
    const axe = await new AxeBuilder({ page })
      .include('[aria-label="Timeline attachments"]')
      .withTags(["wcag2a", "wcag2aa", "wcag21aa"])
      .analyze();
    expect(axe.violations).toEqual([]);
    const response = page.waitForResponse(
      (r) =>
        new URL(r.url()).pathname.endsWith("/cancel") &&
        r.request().method() === "POST",
    );
    await cancel.focus();
    await cancel.press("Enter");
    await response;
    release();
    await expect(cancel).toHaveCount(0);

    release();
  } finally {
    gateBarrier.cancel();
  }
});
