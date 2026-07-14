import { type APIResult, clientTxnID, fetchJSON } from "../services/browserApi";
import { requestMultipartJSON } from "../services/httpTransport";
import type {
  ReferencePackJobEnvelope,
  ReferencePackListEnvelope,
  ReferencePackQuery,
  ReferencePackVersion,
} from "./referencePackAdminModel";

export type ReferencePackAction = "activate" | "disable" | "reverify";

function referencePackListURL(options: {
  cursorToken?: string | null | undefined;
  query: ReferencePackQuery;
}) {
  const params = new URLSearchParams({ limit: "100" });
  const cursorToken = options.cursorToken?.trim() ?? "";
  if (cursorToken !== "") {
    params.set("cursor_token", cursorToken);
  }
  if (options.query.search !== "") {
    params.set("search", options.query.search);
  }
  if (options.query.packVersionState !== "") {
    params.set("pack_version_state", options.query.packVersionState);
  }
  if (options.query.verificationResult !== "") {
    params.set("verification_result", options.query.verificationResult);
  }
  if (options.query.active !== "") {
    params.set("active", options.query.active);
  }
  return `/api/v1/reference-packs?${params.toString()}`;
}

export function listReferencePacks(options: {
  cursorToken?: string | null | undefined;
  query: ReferencePackQuery;
}) {
  return fetchJSON<ReferencePackListEnvelope>(referencePackListURL(options));
}

export function loadReferencePackJob(jobID: string) {
  return fetchJSON<ReferencePackJobEnvelope>(`/api/v1/jobs/${jobID}`);
}

export async function importReferencePackBundle(
  file: File,
): Promise<APIResult<ReferencePackJobEnvelope>> {
  const form = new FormData();
  form.append(
    "metadata",
    new Blob(
      [
        JSON.stringify({
          client_txn_id: clientTxnID("reference-pack-import"),
        }),
      ],
      { type: "application/json" },
    ),
  );
  form.append("file", file);

  return requestMultipartJSON<ReferencePackJobEnvelope>(
    "/api/v1/reference-packs/import",
    form,
  ) as Promise<APIResult<ReferencePackJobEnvelope>>;
}

export function runReferencePackAction(
  pack: ReferencePackVersion,
  action: ReferencePackAction,
) {
  return fetchJSON<ReferencePackJobEnvelope | { data: unknown }>(
    `/api/v1/reference-packs/${encodeURIComponent(pack.pack_key)}/${encodeURIComponent(pack.pack_version)}/${action}`,
    {
      method: "POST",
      body: JSON.stringify({
        client_txn_id: clientTxnID(`reference-pack-${action}`),
      }),
    },
  );
}

export function refreshReferencePacks(options: {
  all: boolean;
  packKeys: readonly string[];
}) {
  const body: Record<string, unknown> = {
    client_txn_id: clientTxnID("reference-pack-refresh"),
  };
  if (!options.all) {
    body.pack_keys = [...options.packKeys];
  }
  return fetchJSON<ReferencePackJobEnvelope>(
    "/api/v1/reference-packs/refresh",
    {
      method: "POST",
      body: JSON.stringify(body),
    },
  );
}

export function cancelReferencePackJob(jobID: string) {
  return fetchJSON<ReferencePackJobEnvelope>(`/api/v1/jobs/${jobID}/cancel`, {
    method: "POST",
    body: JSON.stringify({
      client_txn_id: clientTxnID("reference-pack-cancel"),
    }),
  });
}
