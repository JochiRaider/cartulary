import { Buffer } from "node:buffer";
import type { GetJobResponse } from "@cartulary/protocol-ts/http";
import { incidentImportTestId } from "@cartulary/ui-contracts";
import { expect, type Locator, type Page } from "@playwright/test";
import { DeploymentAdministration } from "../../pages/deploymentAdministration";
import { rewriteSessionPresentation } from "../auth/sessionPresentation";

export { incidentBundleFixture as importBundleFixture } from "../../../../../tools/harness/fixtures/incident-bundle.mjs";

export async function openImportPresentation(page: Page) {
  await page.goto("/deployment-administration");
  await new DeploymentAdministration(page).selectPanel("incident-import");
  return page.getByTestId(incidentImportTestId("form"));
}

export async function expectImportControlReachable(
  page: Page,
  control: Locator,
) {
  await expect
    .poll(() =>
      page.locator("[data-incident-import]").evaluate((element) => {
        const box = element.getBoundingClientRect();
        return box.left >= 0 && box.right <= window.innerWidth + 1;
      }),
    )
    .toBe(true);
  if (await control.evaluate((element) => element === document.activeElement)) {
    await page.keyboard.press("Shift+Tab");
    await page.keyboard.press("Tab");
  } else await control.focus();
  await expect(control).toBeFocused();
  await expect
    .poll(() =>
      control.evaluate((element) => {
        const box = element.getBoundingClientRect();
        return (
          box.left >= 0 &&
          box.right <= window.innerWidth + 1 &&
          box.top >= 0 &&
          box.bottom <= window.innerHeight + 1
        );
      }),
    )
    .toBe(true);
  expect(
    await page.evaluate(
      () =>
        document.documentElement.scrollWidth <=
        document.documentElement.clientWidth + 1,
    ),
  ).toBe(true);
}

export async function installImportObservationFixture(
  page: Page,
  actorId: string,
) {
  // Other administrative fixtures may edit the shared worker's display name.
  // Keep this presentation label stable; current identity and authority stay real.
  await page.route("**/api/v1/auth/session", async (route) => {
    await rewriteSessionPresentation(route, (session) => ({
      ...session,
      display_name: "Import operator",
    }));
  });
  const importJobID = "00000000-0000-4000-8000-000000005003";
  const importedIncidentID = "00000000-0000-4000-8000-000000005002";
  function importJob(
    status: GetJobResponse["data"]["status"] = "queued",
    overrides: Partial<GetJobResponse["data"]> = {},
  ): GetJobResponse["data"] {
    const terminal = ["succeeded", "failed", "canceled"].includes(status);
    const job = {
      job_id: importJobID,
      scope: { kind: "deployment" as const },
      status_route: `/api/v1/jobs/${importJobID}`,
      status,
      cancelable: status === "queued" || status === "running",
      submitted_by_user_id: actorId,
      submitted_at: "2026-09-07T12:00:00Z",
      updated_at: "2026-09-07T12:00:01Z",
      started_at: status === "queued" ? null : "2026-09-07T12:00:01Z",
      finished_at: terminal ? "2026-09-07T12:00:01Z" : null,
      retained_until: terminal ? "2099-09-14T12:00:01Z" : null,
      progress: { completed: status === "succeeded" ? 1 : 0, total: null },
      error_summary:
        status === "failed"
          ? {
              code: "incident_bundle_import_rejected",
              message: "Import rejected",
              retryable: false,
            }
          : null,
      result_summary:
        status === "succeeded"
          ? {
              code: "incident_bundle_imported",
              message: "Imported",
              resource_refs: [
                {
                  kind: "incident",
                  id: importedIncidentID,
                  route: `/api/v1/incidents/${importedIncidentID}`,
                },
              ],
            }
          : status === "canceled"
            ? { code: "job_canceled", message: "Canceled" }
            : null,
      ...overrides,
    };
    return job;
  }
  const jobEnvelope = (job = importJob()): GetJobResponse => ({
    data: job,
    meta: { request_id: "request-import-presentation" },
  });
  let job = importJob();
  let readFailure = false;
  let readStatus = 200;
  let readCount = 0;
  let admissionFailure = false;
  let admissionGate: Promise<void> | null = null;
  let readGate: Promise<void> | null = null;
  let cancellation: "requested" | "rejected" | "lost" | "succeeded" =
    "requested";
  let admissionCount = 0;
  const cancellationBodies: string[] = [];
  await page.route("**/api/v1/incident-bundles/import", async (route) => {
    admissionCount += 1;
    if (admissionGate !== null) await admissionGate;
    if (admissionFailure) return route.abort("failed");
    return route.fulfill({ status: 202, json: jobEnvelope(importJob()) });
  });
  await page.route("**/api/v1/jobs/*", async (route) => {
    ++readCount;
    const captured = job;
    if (readGate !== null) await readGate;
    if (readFailure) return route.abort("failed");
    return route.fulfill({
      status: readStatus,
      json:
        readStatus === 200
          ? jobEnvelope(captured)
          : {
              error: {
                code: "job_not_found",
                message: "Unavailable",
                status: readStatus,
                retryable: false,
                details: {},
              },
            },
    });
  });
  await page.route("**/api/v1/jobs/*/cancel", async (route) => {
    cancellationBodies.push(route.request().postData() ?? "");
    if (cancellation === "lost") return route.abort("failed");
    if (cancellation === "rejected")
      return route.fulfill({
        status: 409,
        json: {
          error: {
            code: "job_cancel_rejected",
            message: "Rejected",
            status: 409,
            retryable: false,
            details: { reason: "not_cancelable" },
          },
        },
      });
    job = importJob(
      cancellation === "succeeded" ? "succeeded" : "cancel_requested",
      { progress: job.progress },
    );
    return route.fulfill({ status: 200, json: jobEnvelope(job) });
  });
  return {
    setJob: (next: typeof job) => {
      job = next;
    },
    readCount: () => readCount,
    readStatus: (value: number) => {
      readStatus = value;
    },
    failReads: (value: boolean) => {
      readFailure = value;
    },
    failAdmission: (value: boolean) => {
      admissionFailure = value;
    },
    gateAdmission: (value: Promise<void> | null) => {
      admissionGate = value;
    },
    gateReads: (value: Promise<void> | null) => {
      readGate = value;
    },
    cancellation: (value: typeof cancellation) => {
      cancellation = value;
    },
    admissionCount: () => admissionCount,
    cancellationBodies,
    importJob,
  };
}
