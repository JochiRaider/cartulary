import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  cartularyReferencePackProgressContract as contract,
  dataTestIdSelector,
  gridShellTestId,
  referencePackAdminPanelTestId,
  referencePackRowTestId,
  rowCellTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookPresenceSummaryTestId,
} from "@cartulary/ui-contracts";
import type { Browser, Page, Request, TestInfo } from "@playwright/test";
import type { FixtureActorCredentials } from "../support/auth/sessions";
import { readCurrentSession } from "../support/auth/sessions";
import { authHeadersForStorageState } from "../support/auth/storageState";
import {
  type Ac043TrafficDriver,
  prepareAc043Snapshot,
  startAc043SnapshotTraffic,
} from "../support/performance/ac043Snapshot";
import {
  openReferencePacks,
  referencePackBundle,
} from "../support/referencePacks";
import { apiBase } from "../support/runtime/configuration";
import { publicHttpOperation } from "../support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "../support/transport/publicJsonClient";
import { expect, test } from "./fixtures";
import type { JobProgressCompletion } from "./jobProgressPaint";
import {
  attachReferencePackProgressObservation,
  measureReferencePackProgress,
  type ReferencePackProgressPurpose,
} from "./referencePackProgressTimingSupport";

test.use({
  viewport: { width: 1920, height: 1080 },
  deviceScaleFactor: 1,
  actionTimeout: 10_000,
  navigationTimeout: 30_000,
});
test.describe.configure({ mode: "serial", timeout: 5 * 60 * 1000 });

async function requireAvailableReferencePack(page: Page) {
  const version = await publicHttpOperation({
    operationID: "getReferencePackVersion",
    request: atJsonOrigin(page.request, apiBase),
    headers: authHeadersForStorageState(await page.context().storageState()),
    pathParameters: {
      pack_key: contract.workload.pack_key,
      pack_version: contract.workload.pack_version,
    },
  });
  if (!version.ok) throw new Error("Reference Pack version is unavailable");
  expect(version.payload.data).toMatchObject({
    pack_key: contract.workload.pack_key,
    pack_version: contract.workload.pack_version,
    manifest_sha256: expect.any(String),
    payload_sha256: contract.workload.payload_sha256,
    health: "verified_available",
    active: false,
    administratively_disabled: false,
    pending_work: false,
  });
}

async function runReferencePackProgress(
  purpose: ReferencePackProgressPurpose,
  browser: Browser,
  page: Page,
  testActor: FixtureActorCredentials,
  sessionTracker: Parameters<typeof startAc043SnapshotTraffic>[1],
  info: TestInfo,
) {
  const predicate = contract.predicates.find((value) => value.kind === purpose);
  if (!predicate) throw new Error("Unknown Reference Pack predicate");
  let admin: Page | undefined,
    traffic: Ac043TrafficDriver | null = null;
  let target: Parameters<typeof attachReferencePackProgressObservation>[2];
  let completion: JobProgressCompletion | undefined,
    attached = false,
    primary: unknown;
  const cleanupErrors: unknown[] = [];
  const requests: Request[] = [];
  const capture = (request: Request) => {
    const path = new URL(request.url()).pathname;
    if (
      request.method() === "POST" &&
      path ===
        (purpose === "import"
          ? "/api/v1/reference-packs/import"
          : purpose === "refresh"
            ? "/api/v1/reference-packs/refresh"
            : `/api/v1/reference-packs/${contract.workload.pack_key}/${contract.workload.pack_version}/reverify`)
    )
      requests.push(request);
  };
  try {
    expect(info.retry).toBe(0);
    const snapshot = await prepareAc043Snapshot(page, testActor);
    const foreground = await readCurrentSession(page);
    expect(foreground.is_deployment_admin).toBe(true);
    expect(foreground.memberships).toContainEqual({
      incident_id: snapshot.incidentId,
      role: "editor",
    });
    const firstPage = [...snapshot.timelineRows]
      .sort((a, b) =>
        a.record_id < b.record_id ? -1 : a.record_id > b.record_id ? 1 : 0,
      )
      .slice(0, 100);
    const record = firstPage[0];
    if (!record) throw new Error("Reference Pack grid target is missing");
    expect(record.cells[contract.workload.field_key]?.value).toBeNull();
    await page.goto(`/?incident_id=${snapshot.incidentId}`);
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    const excluded = new Set(firstPage.map((row) => row.record_id));
    traffic = await startAc043SnapshotTraffic(
      browser,
      sessionTracker,
      snapshot.incidentId,
      snapshot.timelineRows.filter((row) => !excluded.has(row.record_id)),
      snapshot.runtime.backgroundAccounts,
      predicate.predicate_id,
    );
    expect(traffic.qualifiedBackgroundSessions).toBe(24);
    expect(traffic.sessionCount).toBe(25);
    expect(traffic.updatesPerSecond).toBe(4.8);
    await expect(
      page.getByTestId(workbookPresenceSummaryTestId()),
    ).toHaveAttribute("aria-label", /^24 collaborators present on this sheet/);
    await scrollGridCellIntoView({
      page,
      surface: contract.view_schema_id,
      recordId: record.record_id,
      cellKey: contract.workload.field_key,
    });
    const cell = page.getByTestId(
      rowCellTestId(record.record_id, contract.workload.field_key),
    );
    const point = await cell.evaluate((node) => {
      const b = node.getBoundingClientRect(),
        x = b.left + b.width / 2,
        y = b.top + b.height / 2;
      if (
        b.width <= 0 ||
        b.height <= 0 ||
        b.left < 0 ||
        b.top < 0 ||
        b.right > innerWidth ||
        b.bottom > innerHeight ||
        !node.contains(document.elementFromPoint(x, y))
      )
        throw new Error("Reference Pack edit target is not wholly visible");
      return { x, y };
    });
    const popup = page.waitForEvent("popup");
    await page.evaluate(() =>
      window.open("/deployment-administration", "_blank"),
    );
    admin = await popup;
    const measuredAdmin = admin;
    await admin.setViewportSize({ width: 1920, height: 1080 });
    const panel = await openReferencePacks(admin);
    expect(
      await admin.evaluate(
        () => window.opener !== null && window.opener.document !== document,
      ),
    ).toBe(true);
    expect((await readCurrentSession(admin)).user_id).toBe(foreground.user_id);
    const bundle = referencePackBundle();
    expect(bundle.key).toBe(contract.workload.pack_key);
    expect(bundle.version).toBe(contract.workload.pack_version);
    expect(bundle.baseVersion).toBe(contract.workload.base_version);
    expect(bundle.payloadSha256).toBe(contract.workload.payload_sha256);
    target = {
      recordId: record.record_id,
      rowVersion: record.row_version,
      bundle: {
        key: bundle.key,
        version: bundle.version,
        payloadSha256: bundle.payloadSha256,
        sha256: bundle.sha256,
        preparationDate: bundle.preparationDate,
        bytes: bundle.upload.buffer.length,
      },
    };
    const upload = panel.getByLabel("Reference pack bundle", { exact: true });
    await upload.setInputFiles(bundle.upload);
    const selectedFile = await upload.evaluate(async (input) => {
      if (!(input instanceof HTMLInputElement) || input.files?.length !== 1)
        throw new Error("Reference Pack input must contain exactly one file");
      const file = input.files[0];
      if (!file) throw new Error("Reference Pack input file is missing");
      return {
        name: file.name,
        bytes: [...new Uint8Array(await file.arrayBuffer())],
      };
    });
    expect(selectedFile).toEqual({
      name: bundle.upload.name,
      bytes: [...bundle.upload.buffer],
    });
    const imported = panel.getByTestId(
      referencePackRowTestId(bundle.key, bundle.version),
    );
    await expect(imported).toHaveCount(0);
    if (purpose !== "import") {
      await panel.getByRole("button", { name: "Import", exact: true }).click();
      const preparation = panel.locator(
        '[data-job-purpose="reference_pack_import"]',
      );
      await expect(preparation).toHaveAttribute(
        "data-job-status",
        "succeeded",
        { timeout: 120_000 },
      );
      await expect(imported).toContainText("Verified, available");
      await requireAvailableReferencePack(admin);
      if (purpose === "refresh") await imported.getByRole("checkbox").check();
    }
    const button =
      purpose === "reverify"
        ? imported.getByRole("button", { name: "Reverify", exact: true })
        : panel.getByRole("button", {
            name: purpose === "import" ? "Import" : "Refresh selected",
            exact: true,
          });
    await button.scrollIntoViewIfNeeded();
    await button.focus();
    await expect(button).toBeFocused();
    admin.on("request", capture);
    const editorId = timelineScalarEditorTestId({
      recordId: record.record_id,
      fieldKey: contract.workload.field_key,
      surface: "grid",
    });
    completion = await measureReferencePackProgress(
      purpose,
      admin,
      {
        recordId: record.record_id,
        rowVersion: record.row_version,
        fieldKey: contract.workload.field_key,
        draftValue: contract.workload.draft_value,
        gridSelector: dataTestIdSelector(
          gridShellTestId(contract.view_schema_id),
        ),
        editorSelector: dataTestIdSelector(editorId),
        progressSurfaceSelector: dataTestIdSelector(
          referencePackAdminPanelTestId(),
        ),
      },
      async () => {
        await measuredAdmin.keyboard.press("Enter");
        await page.mouse.click(point.x, point.y);
        await page.keyboard.press(contract.workload.draft_value);
      },
    );
    admin.off("request", capture);
    expect(requests).toHaveLength(1);
    // Chromium's network report omits multipart file bodies. The input bytes
    // above and the successfully admitted version below bind this real upload;
    // the transport contract test separately checks exact multipart construction.
    if (purpose === "import")
      expect(requests[0]?.headers()["content-type"]).toMatch(
        /^multipart\/form-data; boundary=\S+$/u,
      );
    if (purpose === "refresh")
      expect(requests[0]?.postDataJSON()).toEqual({
        client_txn_id: expect.any(String),
        pack_keys: [...contract.workload.selected_pack_keys],
      });
    await traffic.stop();
    traffic = null;
    await attachReferencePackProgressObservation(
      info,
      predicate.predicate_id,
      target,
      completion,
    );
    attached = true;
    expect(completion.elapsed).toBeLessThanOrEqual(contract.threshold_ms);
  } catch (error) {
    primary = error;
    if (!attached)
      await attachReferencePackProgressObservation(
        info,
        predicate.predicate_id,
        target,
        completion,
        "Reference Pack measurement did not complete",
      );
  } finally {
    admin?.off("request", capture);
    try {
      await traffic?.stop();
    } catch (error) {
      cleanupErrors.push(error);
    }
    try {
      await page.keyboard.press("Escape");
      const jobId =
        completion?.jobId ??
        (await admin?.evaluate(
          (mark) =>
            (
              performance.getEntriesByName(mark, "mark")[0] as
                | PerformanceMark
                | undefined
            )?.detail?.jobId as string | undefined,
          contract.admission_mark,
        ));
      if (jobId && admin) {
        const cleanupAdmin = admin;
        let terminal = "";
        await expect
          .poll(
            async () => {
              const result = await publicHttpOperation({
                operationID: "getJob",
                request: atJsonOrigin(cleanupAdmin.request, apiBase),
                headers: authHeadersForStorageState(
                  await cleanupAdmin.context().storageState(),
                ),
                pathParameters: { job_id: jobId },
              });
              if (!result.ok)
                throw new Error(
                  "Reference Pack cleanup could not observe the admitted Job",
                );
              terminal = result.payload.data.status;
              return terminal;
            },
            { timeout: 120_000 },
          )
          .toMatch(/^(succeeded|failed|canceled)$/u);
        expect(terminal).toBe("succeeded");
        await requireAvailableReferencePack(admin);
        await info.attach("Reference Pack terminal cleanup", {
          body: JSON.stringify({
            job_purpose: purpose,
            terminal_status: terminal,
          }),
          contentType: "application/json",
        });
      }
    } catch (error) {
      cleanupErrors.push(error);
    }
    try {
      await admin?.close();
    } catch (error) {
      cleanupErrors.push(error);
    }
  }
  if (cleanupErrors.length)
    throw new AggregateError(
      [...(primary ? [primary] : []), ...cleanupErrors],
      "Reference Pack measurement or cleanup failed",
    );
  if (primary) throw primary;
}

test("measures Reference Pack import progress and cancel while Timeline accepts editing", async ({
  browser,
  testActorPage,
  testActor,
  sessionTracker,
}, info) => {
  await runReferencePackProgress(
    "import",
    browser,
    testActorPage,
    testActor,
    sessionTracker,
    info,
  );
});
test("measures Reference Pack reverify progress and cancel while Timeline accepts editing", async ({
  browser,
  testActorPage,
  testActor,
  sessionTracker,
}, info) => {
  await runReferencePackProgress(
    "reverify",
    browser,
    testActorPage,
    testActor,
    sessionTracker,
    info,
  );
});
test("measures Reference Pack refresh progress and cancel while Timeline accepts editing", async ({
  browser,
  testActorPage,
  testActor,
  sessionTracker,
}, info) => {
  await runReferencePackProgress(
    "refresh",
    browser,
    testActorPage,
    testActor,
    sessionTracker,
    info,
  );
});
