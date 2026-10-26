import {
  rowCellTestId,
  timelineEvidenceFileInputTestId,
  timelineScalarEditorTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";
import { openTimelineSurface, tinyPNG } from "./support/evidence/browser";
import {
  chooseEvidenceFile,
  openTimelineAttachmentFeedback,
} from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import { fetchOwnedJobResponse } from "./support/jobs/ownership";
import { responseBarrier } from "./support/runtime/cleanup";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { rejectUnexpectedPageErrors } from "./support/runtime/observations";
import {
  awaitResponseBarrier,
  installOwnedRoute,
} from "./support/transport/requestInterception";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

test.beforeEach(({ page }) => {
  rejectUnexpectedPageErrors(page);
});

test("keeps verification cancel visible and honors cancellation when completion races observation", async ({
  page,
}, info) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("VERIFY-CANCEL"),
    "Verification cancellation",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("verification-source"),
    "timeline.activity_synopsis_text": "File source",
  });
  const other = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("verification-other"),
    "timeline.activity_synopsis_text": "Another editable row",
  });
  // Preserve real admission, worker execution and responses. Only observation
  // delivery is delayed to exercise the completed-worker/cancel-intent race.
  let releaseObservation = () => {};
  const observationGateBarrier = responseBarrier(page);
  releaseObservation = observationGateBarrier.release;
  const observationGate = observationGateBarrier.promise;
  await installOwnedRoute(page, "**/api/v1/jobs/*", async (route) => {
    if (route.request().method() !== "GET") return route.continue();
    const response = await fetchOwnedJobResponse(page, route);
    if (!(await awaitResponseBarrier(route, observationGate))) return;
    await route.fulfill({ response });
  });
  try {
    await openTimelineSurface(page, incidentId);
    await openTimelineInspector(page, source.record_id);
    await chooseEvidenceFile(
      page.getByTestId(timelineEvidenceFileInputTestId(source.record_id)),
      {
        name: "verification-race.png",
        mimeType: "image/png",
        buffer: tinyPNG(),
      },
    );
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .click();
    const attachments = page.getByRole("region", {
      name: "Timeline attachments",
      exact: true,
    });
    await expect(
      attachments.getByRole("button", { name: /^Attachments:/ }),
    ).toHaveAttribute("aria-expanded", "false");
    const verification = attachments.getByRole("group", {
      name: "File verification: verification-race.png",
      exact: true,
    });
    await expect(verification.getByRole("progressbar")).toBeVisible();
    const jobId = await verification.getAttribute("data-job-id");
    expect(jobId).toMatch(/^[a-f0-9-]{36}$/);
    await expect
      .poll(async () => {
        const response = await page.request.get(
          `${apiBase}/api/v1/jobs/${jobId}`,
        );
        expect(response.ok()).toBeTruthy();
        return (await response.json()).data.status;
      })
      .toBe("succeeded");
    await page
      .getByTestId(
        rowCellTestId(other.record_id, "timeline.activity_synopsis_text"),
      )
      .click();
    const editor = page.getByTestId(
      timelineScalarEditorTestId({
        recordId: other.record_id,
        fieldKey: "timeline.activity_synopsis_text",
        surface: "grid",
      }),
    );
    await editor.fill("Uncommitted grid draft");
    await expect(editor).toBeFocused();
    await expect(
      verification.getByRole("button", {
        name: "Cancel verification",
        exact: true,
      }),
    ).toBeVisible();
    await info.attach("verification-progress-with-grid-draft", {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    const cancelled = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === `/api/v1/jobs/${jobId}/cancel`,
    );
    await verification
      .getByRole("button", { name: "Cancel verification", exact: true })
      .click();
    const cancellationResponse = await cancelled;
    expect(cancellationResponse.status()).toBe(409);
    expect((await cancellationResponse.json()).error).toMatchObject({
      code: "job_cancel_rejected",
      details: { reason_code: "already_terminal" },
    });
    releaseObservation();
    await expect(verification).toHaveCount(0);
    await openTimelineAttachmentFeedback(page);
    await expect(
      attachments.getByRole("button", { name: "Resume", exact: true }),
    ).toBeEnabled();
    expect(
      await queryViewRows(page, incidentId, evidenceViewSchemaId),
    ).toHaveLength(0);
    const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
    expect(
      rows.find((row) => row.record_id === source.record_id)?.cells[
        "timeline.evidence_count"
      ]?.value,
    ).toBe(0);

    releaseObservation();
  } finally {
    observationGateBarrier.cancel();
  }
});
