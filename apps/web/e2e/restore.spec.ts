import { fileURLToPath } from "node:url";
import { workbookShellReadyTestId } from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import {
  type BrowserFixtureFailure,
  createBrowserFixtureProcess,
  fixtureLifecycleAttachment,
} from "../../../tools/harness/browser/index.mjs";
import { test as base, expect } from "./fixtures";
import { waitForCommittedRowSummary } from "./measurement/timingSupport";
import { applyCookies, requireCookie } from "./support/auth/browserSession";

type RestoreTarget = {
  backup_set_id: string;
  consistency_point_at: string;
  incident_id: string;
  origin: string;
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
}>({
  restoreProcess: [
    async ({ page }, use, testInfo) => {
      const runtimeRoot = process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT;
      if (!runtimeRoot)
        throw new Error("CARTULARY_WEB_E2E_RUNTIME_ROOT is required");
      const target = createBrowserFixtureProcess({
        command: "go",
        args: [
          "run",
          "./tools/recoverybrowserrestore",
          "--runtime-root",
          runtimeRoot,
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
        if (!page.isClosed())
          await page.goto("about:blank", { timeout: 5_000 });
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
