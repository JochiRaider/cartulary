import type {
  GetJobResponse,
  ListReferencePacksResponse,
} from "@cartulary/protocol-ts/http";
import { referencePackAdminPanelTestId } from "@cartulary/ui-contracts";
import { expect, type Locator, type Page } from "@playwright/test";
import { DeploymentAdministration } from "../pages/deploymentAdministration";
import { rewriteSessionPresentation } from "./auth/sessionPresentation";
import type { BarrierOutcome } from "./runtime/cleanup";
import {
  awaitResponseBarrier,
  installOwnedRoute,
} from "./transport/requestInterception";

export { referencePackBundle } from "../../../../tools/harness/browser/reference-pack-fixture.mjs";

type Pack = ListReferencePacksResponse["data"]["pack_versions"][number];
type Job = GetJobResponse["data"];
const longPackKey = `type_registry.${"deployment_inventory_".repeat(4)}host`;
const longPackVersion = `2026.09.07-${"offline_verified_".repeat(3)}candidate`;
export function referencePackBrowserPack(
  key = longPackKey,
  version = longPackVersion,
): Pack {
  return {
    pack_key: key,
    pack_version: version,
    pack_kind: "type_registry",
    health: "verified_available",
    administratively_disabled: false,
    distribution_kind: "operator_imported",
    removed: false,
    missing_reason: null,
    last_failure_code: null,
    content_profile_id: "cartulary.reference_pack.type_registry.host.v1",
    content_profile_version: "1",
    pack_release_sequence: 1,
    source_profile_id: "test.fixture.v1",
    source_profile_sha256: "d".repeat(64),
    source_version: "1",
    source_as_of: null,
    license_expression: "MIT",
    redistribution: "allowed",
    trust_repository_id: "test.repo",
    last_verified_at: "2026-09-07T12:00:00Z",
    trust_valid_until: "2027-09-07T12:00:00Z",
    verified_signer_key_ids: ["c".repeat(64)],
    pack_contract_version: "cartulary.reference_pack_contract.v3",
    pack_version_state: "verified_available",
    active: false,
    pending_work: false,
    reproducibility_pinned: false,
    fallback_from_version: null,
    dependencies: [],
    source_identifier: null,
    manifest_sha256: "a".repeat(64),
    payload_sha256: "b".repeat(64),
    verification_method: "tuf_1_0_35_offline_bundle_v1",
    previous_active_version: null,
    imported_at: "2026-09-07T12:00:00Z",
    imported_by_user_id: null,
    activated_at: null,
    activated_by_user_id: null,
  };
}

export async function openReferencePacks(page: Page) {
  await page.goto("/deployment-administration");
  await new DeploymentAdministration(page).selectPanel("reference-packs");
  const panel = page.getByTestId(referencePackAdminPanelTestId());
  await expect(
    panel.getByText(/^(Reference packs loaded|No reference packs imported)$/),
  ).toBeVisible();
  await expect(
    panel.getByText("Checking current access and reconciling reference packs."),
  ).toHaveCount(0);
  return panel;
}
export async function expectReferencePackControlReachable(
  page: Page,
  control: Locator,
) {
  const layout = await page
    .getByTestId(referencePackAdminPanelTestId())
    .evaluate((element) => {
      const box = element.getBoundingClientRect();
      return {
        fits:
          box.left >= -1 &&
          box.right <= window.innerWidth + 1 &&
          element.scrollWidth <= element.clientWidth + 1 &&
          document.documentElement.scrollWidth <=
            document.documentElement.clientWidth + 1,
        left: box.left,
        right: box.right,
        width: window.innerWidth,
        scroll: element.scrollWidth,
        client: element.clientWidth,
      };
    });
  expect(layout).toMatchObject({ fits: true });
  await control.focus();
  await control.scrollIntoViewIfNeeded();
  await expect(control).toBeFocused();
  const box = await control.boundingBox();
  const viewport = page.viewportSize();
  expect(box).not.toBeNull();
  expect(viewport).not.toBeNull();
  if (!box || !viewport) return;
  expect(box.x).toBeGreaterThanOrEqual(-1);
  expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
  expect(box.y).toBeGreaterThanOrEqual(-1);
  expect(box.y + box.height).toBeLessThanOrEqual(viewport.height + 1);
}

/** Inline lifecycle results; activate, disable and remove never admit a Job. */
function committedLifecycleAction(
  pack: Pack,
  action: "activate" | "disable" | "remove",
  actorId: string,
): Pack {
  if (action === "activate")
    return {
      ...pack,
      active: true,
      activated_at: "2026-09-07T12:00:05Z",
      activated_by_user_id: actorId,
    };
  if (action === "disable")
    return {
      ...pack,
      active: false,
      administratively_disabled: true,
      pack_version_state: "disabled",
    };
  return {
    ...pack,
    active: false,
    removed: true,
    health: "missing",
    pack_version_state: "missing",
    missing_reason: "administrative_removal",
  };
}
const lifecycleActions = ["activate", "disable", "remove"] as const;
function lifecycleAction(value: string | undefined) {
  return lifecycleActions.find((action) => action === value) ?? null;
}

/** Complete HTTP fixtures for presentation states; production execution stays service-backed. */
export async function installReferencePackPresentation(
  page: Page,
  actorId: string,
) {
  // Shared worker fixtures can change the display name; preserve real identity and authorization.
  await installOwnedRoute(page, "**/api/v1/auth/session", async (route) => {
    await rewriteSessionPresentation(route, (session) => ({
      ...session,
      display_name: "Reference Pack operator",
    }));
  });
  const packs = [
    referencePackBrowserPack(),
    referencePackBrowserPack(longPackKey, "2"),
    referencePackBrowserPack("type_registry.process", "1"),
  ];
  let failAdmission = false;
  let loseLifecycleResponses = false;
  let failReads = false;
  let failCatalog = false;
  let admissionGate: Promise<BarrierOutcome> | null = null;
  // Exact replay of a committed lifecycle action returns its original result.
  const lifecycleReceipts = new Map<string, Pack>();
  const lifecycleRequests: { path: string; body: string }[] = [];
  let lifecycleCommits = 0;
  let sequence = 0;
  let current: Job | null = null;
  let family = "reference_packs_refreshed";
  let target = packs[0];
  const jobs = new Map<string, Job>();
  const admissions: string[] = [];
  let reads = 0;
  const setStatus = (status: Job["status"]) => {
    if (!current) throw new Error("No admitted fixture job");
    const terminal = ["succeeded", "failed", "canceled"].includes(status);
    const route = `/api/v1/reference-packs/${encodeURIComponent(target?.pack_key ?? "")}/${encodeURIComponent(target?.pack_version ?? "")}`;
    current = {
      ...current,
      status,
      cancelable: status === "queued" || status === "running",
      updated_at: terminal
        ? "2026-09-07T12:00:04Z"
        : status === "queued"
          ? "2026-09-07T12:00:00Z"
          : status === "running"
            ? "2026-09-07T12:00:02Z"
            : "2026-09-07T12:00:03Z",
      started_at: status === "queued" ? null : "2026-09-07T12:00:01Z",
      finished_at: terminal ? "2026-09-07T12:00:04Z" : null,
      retained_until: terminal ? "2026-09-15T12:00:04Z" : null,
      progress: {
        completed: status === "succeeded" ? 3 : status === "queued" ? 0 : 1,
        total: status === "queued" ? null : 3,
      },
      result_summary:
        status === "succeeded"
          ? {
              code: family,
              message: "Completed",
              resource_refs:
                family === "reference_packs_refreshed"
                  ? []
                  : [{ kind: "reference_pack_version", id: route, route }],
            }
          : status === "canceled"
            ? { code: "job_canceled", message: "Canceled" }
            : null,
      error_summary:
        status === "failed"
          ? {
              code: "reference_pack_verification_failed",
              message: "Verification failed",
              retryable: false,
              details: { reason_code: "checksum_mismatch" },
            }
          : null,
    };
    jobs.set(current.job_id, current);
  };
  await installOwnedRoute(
    page,
    "**/api/v1/reference-packs**",
    async (route) => {
      const url = new URL(route.request().url());
      if (route.request().method() === "GET") {
        if (failCatalog) {
          await route.abort("failed");
          return;
        }
        const suffix = url.pathname
          .slice("/api/v1/reference-packs".length)
          .split("/")
          .filter(Boolean)
          .map(decodeURIComponent);
        const found = packs.find(
          (pack) =>
            pack.pack_key === suffix[0] && pack.pack_version === suffix[1],
        );
        const rows =
          url.searchParams.get("search") === "process" ? [packs[2]] : packs;
        await route.fulfill({
          json: {
            data: suffix.length ? found : { pack_versions: rows },
            meta: {
              request_id: "catalog",
              ...(suffix.length
                ? {}
                : {
                    paging: { limit: 100, has_more: false, next_cursor: null },
                  }),
            },
          },
        });
        return;
      }
      admissions.push(route.request().postData() ?? "multipart");
      if (!(await awaitResponseBarrier(route, admissionGate))) return;
      if (failAdmission) {
        await route.abort("failed");
        return;
      }
      const action = url.pathname.split("/").at(-1);
      const parts = url.pathname.split("/").map(decodeURIComponent);
      const inline = lifecycleAction(action);
      if (inline) {
        const body = route.request().postData() ?? "";
        lifecycleRequests.push({ path: url.pathname, body });
        const receipt = `${url.pathname}\n${JSON.parse(body).client_txn_id}`;
        let pack = lifecycleReceipts.get(receipt);
        if (!pack) {
          const index = packs.findIndex(
            (candidate) =>
              candidate.pack_key === parts[4] &&
              candidate.pack_version === parts[5],
          );
          const prior = packs[index];
          if (!prior) throw new Error(`Unknown fixture pack ${url.pathname}`);
          pack = committedLifecycleAction(prior, inline, actorId);
          packs[index] = pack;
          lifecycleReceipts.set(receipt, pack);
          ++lifecycleCommits;
        }
        // A lost acknowledgement after commit leaves the client uncertain.
        if (loseLifecycleResponses) {
          await route.abort("failed");
          return;
        }
        await route.fulfill({
          status: 200,
          json: {
            data: { pack_version: pack },
            meta: { request_id: "inline" },
          },
        });
        return;
      }
      family =
        action === "import"
          ? "reference_pack_imported"
          : action === "reverify"
            ? "reference_pack_reverified"
            : "reference_packs_refreshed";
      target =
        packs.find(
          (pack) =>
            pack.pack_key === parts[4] && pack.pack_version === parts[5],
        ) ?? packs[0];
      const id = `11111111-1111-4111-8111-${String(++sequence).padStart(12, "0")}`;
      current = {
        job_id: id,
        scope: { kind: "deployment" },
        status_route: `/api/v1/jobs/${id}`,
        status: "queued",
        cancelable: true,
        submitted_by_user_id: actorId,
        submitted_at: "2026-09-07T12:00:00Z",
        updated_at: "2026-09-07T12:00:00Z",
        started_at: null,
        finished_at: null,
        retained_until: null,
        progress: { completed: 0, total: null },
        result_summary: null,
        error_summary: null,
      };
      jobs.set(id, current);
      await route.fulfill({
        status: 202,
        json: { data: current, meta: { request_id: "admitted" } },
      });
    },
  );
  await installOwnedRoute(page, "**/api/v1/jobs/**", async (route) => {
    const parts = new URL(route.request().url()).pathname.split("/");
    const id = parts[4] ?? "";
    if (route.request().method() === "GET") {
      ++reads;
      if (failReads) {
        await route.abort("failed");
        return;
      }
    } else setStatus("cancel_requested");
    await route.fulfill({
      status: 200,
      json: { data: jobs.get(id), meta: { request_id: "observed" } },
    });
  });
  return {
    packs,
    admissions,
    lifecycleRequests,
    lifecycleCommits: () => lifecycleCommits,
    setStatus,
    reads: () => reads,
    failAdmission: (value: boolean) => {
      failAdmission = value;
    },
    loseLifecycleResponses: (value: boolean) => {
      loseLifecycleResponses = value;
    },
    failReads: (value: boolean) => {
      failReads = value;
    },
    failCatalog: (value: boolean) => {
      failCatalog = value;
    },
    gateAdmission: (value: Promise<BarrierOutcome> | null) => {
      admissionGate = value;
    },
  };
}
