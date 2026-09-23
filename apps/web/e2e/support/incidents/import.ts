import { Buffer } from "node:buffer";
import { createHash, randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";
import type { GetJobResponse } from "@cartulary/protocol-ts/http";
import { incidentImportTestId } from "@cartulary/ui-contracts";
import { expect, type Locator, type Page } from "@playwright/test";
import { DeploymentAdministration } from "../../pages/deploymentAdministration";
import { rewriteSessionPresentation } from "../auth/sessionPresentation";

// Test-only current-format empty-workbook source, using the authored machine inventory.
// Actual admission, validation, publication and membership run on the server.
export function importBundleFixture(status: "active" | "closed" = "active") {
  const incidentId = randomUUID();
  const actorId = randomUUID();
  const key = `IMPORT-${incidentId}`;
  const timestamp = "2026-01-02T03:04:05+00:00";
  const catalog: {
    families: { paths: { logical_path: string; versions: number[] }[] }[];
    special_consumers: { logical_path: string; versions: number[] }[];
  } = JSON.parse(
    readFileSync(
      new URL(
        "../../../../../contracts/incident-bundles/source_catalog.json",
        import.meta.url,
      ),
      "utf8",
    ),
  );
  const sources = new Map<string, Buffer>();
  for (const source of [
    ...catalog.families.flatMap((family) => family.paths),
    ...catalog.special_consumers,
  ]) {
    if (source.versions.includes(4) && !source.logical_path.includes("*"))
      sources.set(source.logical_path, Buffer.alloc(0));
  }
  const json = (value: unknown) => Buffer.from(`${JSON.stringify(value)}\n`);
  sources.set(
    "data/incident.json",
    json({
      id: incidentId,
      incident_key: key,
      incident_key_canonical: key,
      title: `Imported ${status} investigation`,
      description: null,
      status,
      severity: null,
      tlp: null,
      current_phase: null,
      primary_external_case_ref: null,
      created_by_user_id: actorId,
      created_at: timestamp,
      updated_at: timestamp,
      updated_by_user_id: actorId,
      incident_version: 1,
      closed_at: status === "closed" ? timestamp : null,
    }),
  );
  sources.set(
    "data/actors.ndjson",
    json({ actor_id: actorId, display_name: "Portable operator" }),
  );
  sources.set("data/reference_pack_refs.json", json([]));
  sources.set(
    "data/timeline_time_profiles.ndjson",
    json({
      incident_id: incidentId,
      enabled: false,
      local_offset_minutes: null,
      local_label: null,
      profile_version: 1,
      updated_at: timestamp,
      updated_by_user_id: actorId,
    }),
  );
  const hash = (bytes: Buffer) =>
    createHash("sha256").update(bytes).digest("hex");
  const files = [...sources]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([path, bytes]) => ({
      path,
      sha256: `sha256:${hash(bytes)}`,
      size_bytes: bytes.length,
      required: true,
    }));
  sources.set(
    "manifest.json",
    json({
      bundle_format: "cartulary.incident_bundle",
      bundle_version: 4,
      bundle_id: randomUUID(),
      incident_id: incidentId,
      incident_key: key,
      exported_at: timestamp,
      source_change_set_high_watermark: `cartulary.source_boundary.v1:${hash(Buffer.from(JSON.stringify(files)))}`,
      history_mode: "full",
      blob_mode: "full",
      reference_pack_mode: "refs_only",
      optional_sections: [],
      required_capabilities: [],
      files,
    }),
  );
  sources.set(
    "integrity/checksums.sha256",
    Buffer.from(
      files.map((file) => `${file.sha256.slice(7)}  ${file.path}\n`).join(""),
    ),
  );
  const chunks: Buffer[] = [];
  for (const [name, bytes] of sources) {
    const header = Buffer.alloc(512);
    if (Buffer.byteLength(name) >= 100)
      throw new Error("Test TAR path exceeds header capacity");
    header.write(name, 0, 100);
    const octal = (value: number, offset: number, width: number) =>
      header.write(
        `${value.toString(8).padStart(width - 1, "0")}\0`,
        offset,
        width,
      );
    octal(0o644, 100, 8);
    octal(0, 108, 8);
    octal(0, 116, 8);
    octal(bytes.length, 124, 12);
    octal(0, 136, 12);
    header.fill(32, 148, 156);
    header.write("0", 156);
    header.write("ustar\0", 257);
    header.write("00", 263);
    header.write(
      `${header
        .reduce((sum, byte) => sum + byte, 0)
        .toString(8)
        .padStart(6, "0")}\0 `,
      148,
      8,
    );
    chunks.push(
      header,
      bytes,
      Buffer.alloc((512 - (bytes.length % 512)) % 512),
    );
  }
  chunks.push(Buffer.alloc(1024));
  return {
    incidentId,
    key,
    upload: {
      name: `incident-${status}.tar`,
      mimeType: "application/x-tar",
      buffer: Buffer.concat(chunks),
    },
  };
}

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
