import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  cartularyEvidenceVerificationContract as contract,
  dataTestIdSelector,
  gridShellTestId,
  rowCellTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import { type Request, request } from "@playwright/test";
import {
  loginTrackedUserViaPage,
  logoutAndVerify,
  readCurrentSession,
} from "../support/auth/sessions";
import { authHeadersForStorageState } from "../support/auth/storageState";
import {
  evidenceInspectorAccount,
  prepareEvidenceInspectorSnapshot,
} from "../support/performance/evidenceInspectorSnapshot";
import { apiBase, webBase } from "../support/runtime/configuration";
import { publicHttpOperation } from "../support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "../support/transport/publicJsonClient";
import { queryViewRows } from "../support/workbook/query";
import {
  attachEvidenceVerificationObservation,
  measureEvidenceVerification,
} from "./evidenceVerificationTimingSupport";
import { expect, test } from "./fixtures";
import type { JobProgressCompletion } from "./jobProgressPaint";

// The physical preview preflight and File upload must not retain capabilities or bytes.
test.use({ trace: "off", screenshot: "off", video: "off" });
test.describe.configure({ mode: "serial", timeout: 5 * 60 * 1000 });
test("measures Evidence verification progress and cancel while another Timeline row accepts editing", async ({
  browser,
}, info) => {
  expect(info.retry).toBe(0);
  const context = await browser.newContext({
    viewport: { width: 1920, height: 1080 },
    deviceScaleFactor: 1,
  });
  context.setDefaultTimeout(10_000);
  context.setDefaultNavigationTimeout(30_000);
  const page = await context.newPage();
  let authenticated = false,
    attached = false,
    completion: JobProgressCompletion | undefined;
  let target: Parameters<typeof attachEvidenceVerificationObservation>[1];
  let primary: unknown;
  const cleanupErrors: unknown[] = [];
  const admissions: string[] = [],
    slots: Request[] = [],
    uploads: Request[] = [];
  const capture = (req: Request) => {
    const path = new URL(req.url()).pathname;
    if (
      req.method() === "POST" &&
      /^\/api\/v1\/object-blobs\/[^/]+\/verify$/u.test(path)
    )
      admissions.push(path.split("/")[4] ?? "");
    if (req.method() === "POST" && path === "/api/v1/object-blobs")
      slots.push(req);
    if (req.method() === "PUT") uploads.push(req);
  };
  try {
    await page.goto(webBase);
    await loginTrackedUserViaPage(page, evidenceInspectorAccount());
    authenticated = true;
    const snapshot = await prepareEvidenceInspectorSnapshot(page);
    const session = await readCurrentSession(page);
    expect(session.is_deployment_admin).toBe(false);
    expect(session.memberships).toEqual([
      { incident_id: snapshot.incidentId, role: "editor" },
    ]);
    const rows = await queryViewRows(
      page,
      snapshot.incidentId,
      contract.view_schema_id,
    );
    const [source, record] = rows;
    if (!source || !record || source.record_id >= record.record_id)
      throw new Error(
        "Evidence workload requires the first two default record identities",
      );
    expect(record.cells[contract.workload.field_key]?.value).toBeNull();
    expect(snapshot.verifiedFile.length).toBe(contract.workload.object_bytes);
    await page.goto(`${webBase}/?incident_id=${snapshot.incidentId}`);
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    await scrollGridCellIntoView({
      page,
      surface: contract.view_schema_id,
      recordId: record.record_id,
      cellKey: contract.workload.field_key,
    });
    const sourceCell = page.getByTestId(
      rowCellTestId(source.record_id, contract.workload.field_key),
    );
    const targetCell = page.getByTestId(
      rowCellTestId(record.record_id, contract.workload.field_key),
    );
    for (const cell of [sourceCell, targetCell])
      expect(
        await cell.evaluate((node) => {
          const b = node.getBoundingClientRect();
          return (
            b.width > 0 &&
            b.height > 0 &&
            b.left >= 0 &&
            b.right <= innerWidth &&
            b.top >= 0 &&
            b.bottom <= innerHeight
          );
        }),
      ).toBe(true);
    await sourceCell.click();
    const sourceEditor = page.getByTestId(
      timelineScalarEditorTestId({
        recordId: source.record_id,
        fieldKey: contract.workload.field_key,
        surface: "grid",
      }),
    );
    await expect(sourceEditor).toBeFocused();
    const editorId = timelineScalarEditorTestId({
      recordId: record.record_id,
      fieldKey: contract.workload.field_key,
      surface: "grid",
    });
    target = {
      sourceRecordId: source.record_id,
      recordId: record.record_id,
      rowVersion: record.row_version,
    };
    page.on("request", capture);
    completion = await measureEvidenceVerification(
      page,
      {
        recordId: record.record_id,
        rowVersion: record.row_version,
        fieldKey: contract.workload.field_key,
        draftValue: contract.workload.draft_value,
        gridSelector: dataTestIdSelector(
          gridShellTestId(contract.view_schema_id),
        ),
        editorSelector: dataTestIdSelector(editorId),
        progressSurfaceSelector: `[data-workbook-view-schema-id="${contract.view_schema_id}"]`,
      },
      async () => {
        await sourceEditor.evaluate(
          (node, input) => {
            const data = new DataTransfer();
            data.items.add(
              new File(
                [
                  Uint8Array.from(atob(input.base64), (value) =>
                    value.charCodeAt(0),
                  ),
                ],
                input.filename,
                {
                  type: input.contentType,
                },
              ),
            );
            node.dispatchEvent(
              new ClipboardEvent("paste", {
                bubbles: true,
                cancelable: true,
                clipboardData: data,
              }),
            );
          },
          {
            base64: snapshot.verifiedFile.toString("base64"),
            filename: contract.workload.filename,
            contentType: contract.workload.content_type,
          },
        );
        await page.waitForFunction(
          (mark) => performance.getEntriesByName(mark, "mark").length === 1,
          contract.admission_mark,
        );
        // Inspect the current position without scrolling or invoking locator auto-scroll.
        const point = await targetCell.evaluate((node) => {
          const field = node.closest<HTMLElement>("[data-grid-field-key]");
          if (!field) throw new Error("Missing edit target");
          const b = field.getBoundingClientRect(),
            x = b.left + b.width / 2,
            y = b.top + b.height / 2;
          if (
            b.left < 0 ||
            b.right > innerWidth ||
            b.top < 0 ||
            b.bottom > innerHeight ||
            !field.contains(document.elementFromPoint(x, y))
          )
            throw new Error("Edit target is not visible");
          return { x, y };
        });
        await page.mouse.click(point.x, point.y);
        await page.keyboard.press(contract.workload.draft_value);
      },
    );
    page.off("request", capture);
    expect(admissions).toEqual([completion.blobId]);
    expect(slots).toHaveLength(1);
    expect(uploads).toHaveLength(1);
    const surface = page.getByRole("region", {
      name: "Timeline attachments",
      exact: true,
    });
    await expect(
      surface.getByRole("button", { name: /^Attachments:/ }),
    ).toHaveAttribute("aria-expanded", "false");
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    await attachEvidenceVerificationObservation(info, target, completion);
    attached = true;
    expect(completion.elapsed).toBeLessThanOrEqual(contract.threshold_ms);
  } catch (error) {
    primary = error;
    if (!attached)
      await attachEvidenceVerificationObservation(
        info,
        target,
        completion,
        "Evidence verification measurement did not complete",
      );
  } finally {
    page.off("request", capture);
    try {
      await page.keyboard.press("Escape");
      // The measured action is bounded independently of observation. Let it drain;
      // a terminal transition may remove Cancel between visibility and a click.
      const jobId =
        completion?.jobId ??
        (await page.evaluate((mark) => {
          const entry = performance.getEntriesByName(mark, "mark")[0] as
            | PerformanceMark
            | undefined;
          return entry?.detail?.jobId as string | undefined;
        }, contract.admission_mark));
      if (jobId) {
        let terminalStatus = "";
        await expect
          .poll(
            async () => {
              const result = await publicHttpOperation({
                operationID: "getJob",
                request: atJsonOrigin(page.request, apiBase),
                headers: authHeadersForStorageState(
                  await context.storageState(),
                ),
                pathParameters: { job_id: jobId },
              });
              if (!result.ok)
                throw new Error(
                  "Verification cleanup could not read the admitted Job",
                );
              terminalStatus = result.payload.data.status;
              return terminalStatus;
            },
            { timeout: 130_000 },
          )
          .toMatch(/^(succeeded|failed|canceled)$/u);
        expect(terminalStatus).toBe("succeeded");
        await info.attach("Evidence verification terminal cleanup", {
          body: JSON.stringify({
            cancellation_requested: false,
            terminal_status: terminalStatus,
          }),
          contentType: "application/json",
        });
      }
    } catch (error) {
      cleanupErrors.push(error);
    }
    try {
      if (authenticated) {
        const state = await context.storageState();
        const auth = await request.newContext({
          baseURL: apiBase,
          extraHTTPHeaders: authHeadersForStorageState(state),
        });
        try {
          await logoutAndVerify(auth, state, "Evidence verification analyst");
        } finally {
          await auth.dispose();
        }
      }
    } catch (error) {
      cleanupErrors.push(error);
    } finally {
      await context.close();
    }
  }
  if (cleanupErrors.length)
    throw new AggregateError(
      [...(primary ? [primary] : []), ...cleanupErrors],
      "Evidence measurement or cleanup failed",
    );
  if (primary) throw primary;
});
