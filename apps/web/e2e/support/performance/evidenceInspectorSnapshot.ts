import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import type { ViewRow } from "@cartulary/protocol-ts/http";
import { cartularyEvidenceInspectorContract as contract } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { authHeadersForStorageState } from "../auth/storageState";
import { apiBase } from "../runtime/configuration";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";
import { queryViewRows } from "../workbook/query";
import { parseFixtureRuntimeBundle } from "./runtimeBundle";

export const evidenceInspectorProfileId = "evidence_heavy_snapshot_v2";
export type EvidenceInspectorTarget = {
  record_id: string;
  row_version: number;
  summary: string;
  items: Array<{ record_id: string; row_version: number; title: string }>;
};

export function evidenceInspectorAccount() {
  const fixtureProfileId = process.env.CARTULARY_FIXTURE_PROFILE_ID;
  const snapshotKey = process.env.CARTULARY_FIXTURE_SNAPSHOT_KEY ?? "";
  const runtimePath = process.env.CARTULARY_PERFORMANCE_FIXTURE_RUNTIME_BUNDLE;
  if (fixtureProfileId !== evidenceInspectorProfileId || !runtimePath)
    throw new Error("Evidence inspector requires its admitted private fixture");
  const runtime = parseFixtureRuntimeBundle(readFileSync(runtimePath, "utf8"), {
    fixtureProfileId,
    snapshotKey,
    credentialSetId: "foreground_analysts",
    accountCount: 1,
  });
  const account = runtime.accounts[0];
  if (!account) throw new Error("Evidence inspector analyst is missing");
  return account;
}

export async function prepareEvidenceInspectorSnapshot(page: Page) {
  const headers = authHeadersForStorageState(
    await page.context().storageState(),
  );
  const request = atJsonOrigin(page.request, apiBase);
  const key = `AC045-PERF-${contract.fixture.seed}`;
  const incidents = await publicHttpOperation({
    operationID: "listVisibleIncidents",
    request,
    headers,
    query: { limit: 100, search: key },
  });
  if (!incidents.ok)
    throw new Error("Evidence inspector incident is inaccessible");
  const matches = incidents.payload.data.incidents.filter(
    (incident) => incident.incident_key === key,
  );
  const incidentId = matches[0]?.incident_id;
  if (matches.length !== 1 || !incidentId)
    throw new Error("Evidence inspector incident identity is ambiguous");
  const rows = await queryViewRows(page, incidentId, contract.view_schema_id, {
    filters: [
      {
        field_key: contract.workload.filter_field_key,
        op: contract.workload.filter_operator,
        arg: { value: contract.workload.filter_value },
      },
    ],
  });
  if (rows.length !== 1)
    throw new Error("Evidence inspector requires one linked target");
  const row = rows[0] as ViewRow;
  const summary = row.cells["timeline.activity_synopsis_text"]?.value;
  const collection = row.cells["timeline.attached_evidence_ids"]?.value as
    | { items?: Array<{ linked_record_id?: string }> }
    | undefined;
  const ids = collection?.items?.map((item) => item.linked_record_id);
  if (
    typeof summary !== "string" ||
    summary !== contract.fixture.target_summary ||
    row.row_version !== 2 ||
    row.cells["timeline.evidence_count"]?.value !==
      contract.fixture.linked_evidence_count ||
    !ids ||
    ids.length !== contract.fixture.linked_evidence_count ||
    ids.some((id) => typeof id !== "string") ||
    new Set(ids).size !== ids.length
  )
    throw new Error(
      "Evidence inspector target bindings differ from the workload",
    );
  const recordIds = ids as string[];
  const metadata = await publicHttpOperation({
    operationID: "queryEvidenceMetadata",
    request,
    headers,
    pathParameters: { incident_id: incidentId },
    body: { record_ids: recordIds as [string, ...string[]] },
  });
  if (!metadata.ok || metadata.payload.data.items.length !== recordIds.length)
    throw new Error("Evidence inspector metadata preflight failed");
  const items = metadata.payload.data.items;
  const titles = new Set(items.map((item) => item.title));
  for (let index = 1; index <= 100; index++)
    if (
      !titles.has(
        `${contract.fixture.evidence_title_prefix} ${String(index).padStart(5, "0")}`,
      )
    )
      throw new Error(
        "Evidence inspector attached population differs from construction order",
      );
  const validatedItems: EvidenceInspectorTarget["items"] = [];
  let verifiedFile: Buffer | undefined;
  for (const [index, item] of items.entries()) {
    if (
      typeof item.title !== "string" ||
      item.record_id !== recordIds[index] ||
      item.row_version !== 1 ||
      item.lifecycle_state !== "available" ||
      item.upload_state !== "available" ||
      !item.has_blob ||
      !item.preview_eligibility.eligible
    )
      throw new Error("Evidence inspector metadata population is not ready");
    validatedItems.push({
      record_id: item.record_id,
      row_version: item.row_version,
      title: item.title,
    });
    const issued = await publicHttpOperation({
      operationID: "issueEvidencePreviewHandle",
      request,
      headers,
      pathParameters: { record_id: item.record_id },
      body: {},
    });
    if (!issued.ok)
      throw new Error("Evidence inspector preview preflight was denied");
    const handle = issued.payload.data;
    const url = new URL(handle.href, apiBase);
    if (
      url.origin !== new URL(apiBase).origin ||
      !url.pathname.startsWith("/api/v1/evidence-handles/") ||
      handle.record_id !== item.record_id ||
      handle.incident_id !== incidentId ||
      handle.handle_kind !== "preview" ||
      handle.content_type !== contract.fixture.content_type ||
      handle.size_bytes !== contract.fixture.object_bytes ||
      !handle.sha256 ||
      !/^[a-f0-9]{64}$/u.test(handle.sha256)
    )
      throw new Error(
        "Evidence inspector preview handle has incompatible scope",
      );
    // Capability URLs and bodies never enter retained artifacts or failures.
    const response = await page.request
      .get(url.toString(), { headers })
      .catch(() => {
        throw new Error("Evidence inspector preview transport failed");
      });
    try {
      const body = await response.body().catch(() => {
        throw new Error("Evidence inspector preview body read failed");
      });
      if (
        response.status() !== 200 ||
        body.length !== contract.fixture.object_bytes ||
        createHash("sha256").update(body).digest("hex") !== handle.sha256 ||
        body.subarray(0, 8).toString("hex") !== "89504e470d0a1a0a" ||
        body.readUInt32BE(16) !== 2 ||
        body.readUInt32BE(20) !== 2
      )
        throw new Error(
          "Evidence inspector complete preview body verification failed",
        );
      verifiedFile ??= body;
    } finally {
      await response.dispose();
    }
  }
  const target: EvidenceInspectorTarget = {
    record_id: row.record_id,
    row_version: row.row_version,
    summary,
    items: validatedItems.slice(0, contract.workload.metadata_window_size),
  };
  if (!verifiedFile)
    throw new Error("Evidence preview preflight produced no verified file");
  return { incidentId, target, verifiedFile };
}
