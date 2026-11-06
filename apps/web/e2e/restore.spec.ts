import { createHash, randomUUID } from "node:crypto";
import { readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  incidentControlsMenuItemTestId,
  incidentControlsTriggerTestId,
  networkAnalysisSavedGraphTestId,
  networkAnalysisTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page, TestInfo } from "@playwright/test";
import {
  type BrowserFixtureFailure,
  createBrowserFixtureProcess,
  fixtureLifecycleAttachment,
} from "../../../tools/harness/browser/index.mjs";
import { test as base, expect } from "./fixtures";
import { waitForCommittedRowSummary } from "./measurement/timingSupport";
import { applyCookies, requireCookie } from "./support/auth/browserSession";
import {
  createPortableInvestigation,
  type PortableInvestigation,
} from "./support/extensions/network_flow_activity/portableInvestigation";
import {
  type NetworkFlowFixtureFields,
  networkFlowFixtureSetup,
} from "./support/extensions/network_flow_activity/workspace";

type RestoreTarget = {
  backup_set_id: string;
  consistency_point_at: string;
  incident_id: string;
  origin: string;
  transferred_origin: string;
  restored_incident_ids: string[];
  schema_id: string;
  timeline_summary: string;
  user_email: string;
  user_password: string;
};

type RestoreProcess = {
  target: ReturnType<typeof createBrowserFixtureProcess>;
  ready: boolean;
};

const test = base.extend<{
  restoreProcess: RestoreProcess;
  restoreTarget: RestoreTarget;
  portableSource: PortableInvestigation | null;
}>({
  portableSource: null,
  restoreProcess: [
    async ({ browserScope, portableSource }, use, testInfo) => {
      const runtimeRoot = process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT;
      if (!runtimeRoot)
        throw new Error("CARTULARY_WEB_E2E_RUNTIME_ROOT is required");
      const bundleName = portableSource ? `portable-${randomUUID()}.zip` : null;
      if (bundleName && portableSource)
        await writeFile(
          path.join(runtimeRoot, bundleName),
          portableSource.bundle,
          { mode: 0o600, flag: "wx" },
        );
      const target = createBrowserFixtureProcess({
        command: "go",
        args: [
          "run",
          "./tools/recoverybrowserrestore",
          "--runtime-root",
          runtimeRoot,
          ...(bundleName ? ["--incident-bundle", bundleName] : []),
        ],
        cwd: fileURLToPath(new URL("../../..", import.meta.url)),
      });
      const owned: RestoreProcess = { target, ready: false };
      const failures: BrowserFixtureFailure[] = [];
      try {
        // Yield ownership before waiting for readiness. A startup timeout in
        // restoreTarget must still unwind this fixture on its teardown budget.
        await use(owned);
      } catch (error) {
        failures.push({
          phase: "startup",
          failure_class: "harness",
          failure_reason: "tool_diagnostic_failure",
          error,
        });
      }
      for (const error of testInfo.errors) {
        const interrupted = testInfo.status === "interrupted";
        failures.push({
          phase: owned.ready || interrupted ? "body" : "startup",
          failure_class: interrupted
            ? "interrupted"
            : owned.ready
              ? "product"
              : "harness",
          failure_reason: interrupted
            ? "cancelled_or_interrupted"
            : owned.ready
              ? "test_assertion_failure"
              : "tool_diagnostic_failure",
          error,
        });
      }
      try {
        // Restored sessions belong to the disposable deployment. The common
        // scope closes their browser transport without registering those actors
        // against the ordinary deployment's borrowed provisioner.
        await browserScope.dispose();
      } catch (error) {
        failures.push({
          phase: "cleanup",
          failure_class: "harness",
          failure_reason: "cleanup_error",
          error,
        });
      }
      try {
        await target.stop();
      } catch (error) {
        failures.push({
          phase: "cleanup",
          failure_class: "harness",
          failure_reason: "cleanup_error",
          error,
        });
      }
      if (bundleName)
        await rm(path.join(runtimeRoot, bundleName), { force: true });
      const report = target.report(failures);
      // Inline, redacted attempt evidence survives without an attachment path.
      await testInfo.attach(fixtureLifecycleAttachment, {
        contentType: "application/json",
        body: Buffer.from(JSON.stringify(report)),
      });
      const fixtureFailures = report.failures.filter(
        (failure) => failure.phase !== "body",
      );
      if (fixtureFailures.length)
        throw new AggregateError(
          fixtureFailures.map((failure) => new Error(failure.message)),
          fixtureFailures.map((failure) => failure.message).join("\n"),
        );
    },
    { timeout: 100_000 },
  ],
  restoreTarget: [
    async ({ restoreProcess }, use) => {
      const ready = await restoreProcess.target.ready;
      restoreProcess.ready = true;
      await use(ready as RestoreTarget);
    },
    { timeout: 100_000 },
  ],
});

const portabilityTest = test.extend<NetworkFlowFixtureFields>({
  ...networkFlowFixtureSetup,
  portableSource: [
    async ({ page, networkFlowMinimalCSV }, use) => {
      await use(await createPortableInvestigation(page, networkFlowMinimalCSV));
    },
    { scope: "test", timeout: 90_000 },
  ],
});

portabilityTest(
  "IR transferred investigation reopens selected history and retained report files",
  async ({ page, portableSource, restoreTarget }, info) => {
    await verifyPortableDestination(
      "transferred",
      { page, portableSource, restoreTarget },
      info,
    );
  },
);

portabilityTest(
  "IR restored investigation reopens selected history and retained report files",
  async ({ page, portableSource, restoreTarget }, info) => {
    await verifyPortableDestination(
      "restored",
      { page, portableSource, restoreTarget },
      info,
    );
  },
);

async function verifyPortableDestination(
  stage: "transferred" | "restored",
  {
    page,
    portableSource,
    restoreTarget: ready,
  }: {
    page: Page;
    portableSource: PortableInvestigation | null;
    restoreTarget: RestoreTarget;
  },
  info: TestInfo,
) {
  info.setTimeout(120_000);
  if (!portableSource)
    throw new Error("Portable investigation setup is required");
  expect(ready.incident_id).toBe(portableSource.incidentId);
  const origin =
    stage === "transferred" ? ready.transferred_origin : ready.origin;
  await loginTargetLocalSession(
    page,
    origin,
    ready.user_email,
    ready.user_password,
  );
  const graphID = portableSource.graph.graph_view.graph_view_id;
  const root = `${origin}/api/v1/incidents/${ready.incident_id}/network-flow`;
  await expect
    .poll(
      async () => {
        const response = await page.request.get(
          `${root}/graph-views/${graphID}/result`,
        );
        if (response.status() === 502) {
          const failure = await response.json();
          expect(failure.error.code).toBe(
            "network_flow_graph_projection_failed",
          );
          expect(failure.error.details.reason_code).toBe(
            "projection_unavailable",
          );
          return false;
        }
        expect(response.ok(), await response.text()).toBeTruthy();
        const selected = (await response.json()).data;
        expect(selected.result).toEqual(portableSource.graph.result);
        expect(selected.graph_view.selected_source_binding).toEqual(
          portableSource.graph.graph_view.selected_source_binding,
        );
        return true;
      },
      { timeout: 30_000 },
    )
    .toBe(true);
  const refreshes: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().includes(`/graph-views/${graphID}/refresh`)
    )
      refreshes.push(request.url());
  });
  await page.goto(`${origin}/?incident_id=${ready.incident_id}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page.getByRole("button", { name: "More views", exact: true }).click();
  await page.getByTestId(networkAnalysisTestId("tab")).click();
  await page.getByRole("button", { name: "Graph", exact: true }).click();
  await page.getByRole("button", { name: "Saved graphs", exact: true }).click();
  const graphs = page.getByRole("region", {
    name: "Saved Network Flow graphs",
  });
  const retainedGraph = graphs.getByTestId(
    networkAnalysisSavedGraphTestId(graphID),
  );
  await expect(retainedGraph).toContainText("Selected result");
  await retainedGraph.click();
  const result = graphs.getByTestId(
    networkAnalysisTestId("saved-graph-result"),
  );
  await expect(result).toHaveAttribute(
    "data-projection-result-id",
    portableSource.graph.result.graph_projection_result.projection_result_id,
  );
  await expect(result).toBeVisible();
  expect(refreshes).toEqual([]);
  await page.screenshot({
    path: info.outputPath(`${stage}-selected-graph.png`),
  });
  await page.getByLabel("Account and application navigation").click();
  await page.getByTestId(incidentControlsTriggerTestId()).click();
  await page
    .getByTestId(incidentControlsMenuItemTestId("snapshot-reports"))
    .click();
  const reports = page.getByRole("region", {
    name: "Transferred report files",
  });
  await reports
    .getByRole("button", { name: "Load transferred report files" })
    .click();
  const downloadReportFile = async (filename: string): Promise<Buffer> => {
    const link = reports.getByRole("link", { name: filename, exact: true });
    await expect(link).toBeVisible();
    const [download] = await Promise.all([
      page.waitForEvent("download"),
      link.click(),
    ]);
    const saved = info.outputPath(`${stage}-${filename}`);
    await download.saveAs(saved);
    return readFile(saved);
  };
  // output_sha256 identifies the approved render manifest, not an arbitrary
  // Mermaid member. Verify that binding before checking the selected diagram.
  const manifestBytes = await downloadReportFile("manifest.json");
  expect(
    createHash("sha256")
      .update("cartulary.render_bundle_manifest.v3\n")
      .update(manifestBytes)
      .digest("hex"),
  ).toBe(portableSource.reportSHA256);
  const manifest = JSON.parse(manifestBytes.toString()) as {
    files: { path: string; sha256: string }[];
  };
  const diagram = manifest.files.find(
    (file) => file.path === "diagrams/browser-retained-flow.mmd",
  );
  if (!diagram) throw new Error("Selected graph report source absent");
  const bytes = await downloadReportFile(path.basename(diagram.path));
  expect(createHash("sha256").update(bytes).digest("hex")).toBe(diagram.sha256);
  expect(bytes.toString()).toContain("flowchart");
  await page.screenshot({ path: info.outputPath(`${stage}-report-files.png`) });
}

test("restore recovers workbook surface and executes a built-in workbook query", async ({
  page,
  restoreTarget: ready,
}) => {
  test.setTimeout(120_000);
  const incidentId = ready.incident_id;
  const summary = ready.timeline_summary;
  expect(ready.schema_id).toBe("cartulary.restore.browser_restore_target.v1");
  expect(ready.backup_set_id).toMatch(
    /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u,
  );
  expect(ready.restored_incident_ids).toContain(incidentId);

  await loginTargetLocalSession(
    page,
    ready.origin,
    ready.user_email,
    ready.user_password,
  );
  await page.goto(`${ready.origin}/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await waitForCommittedRowSummary(page, {
    expectedSummary: summary,
    surface: timelineViewSchemaId,
    timeoutMs: 5_000,
  });

  const queryResponse = await page.request.post(
    `${ready.origin}/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
    { data: {} },
  );
  expect(queryResponse.ok(), await queryResponse.text()).toBeTruthy();
  const query = (await queryResponse.json()) as {
    data: { rows: Array<{ cells: Record<string, { value: unknown }> }> };
  };
  expect(
    query.data.rows.some(
      (row) => row.cells["timeline.activity_synopsis_text"]?.value === summary,
    ),
  ).toBe(true);
});

async function loginTargetLocalSession(
  page: Parameters<typeof applyCookies>[0],
  origin: string,
  email: string,
  password: string,
) {
  await page.context().clearCookies();
  const response = await page.request.post(`${origin}/api/v1/auth/login`, {
    data: {
      password,
      username: email,
    },
  });
  expect(response.ok(), await response.text()).toBeTruthy();
  await applyCookies(
    page,
    requireCookie(response, "cartulary_session"),
    requireCookie(response, "cartulary_csrf"),
  );
}
