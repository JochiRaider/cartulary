import type {
  ActivateReferencePackVersionRequest,
  ActivateReferencePackVersionResponse,
  CancelJobRequest,
  CancelJobResponse,
  DisableReferencePackVersionRequest,
  DisableReferencePackVersionResponse,
  GetJobResponse,
  ImportReferencePackResponse,
  ListReferencePacksResponse,
  RefreshReferencePacksRequest,
  RefreshReferencePacksResponse,
  ReverifyReferencePackVersionRequest,
  ReverifyReferencePackVersionResponse,
} from "@cartulary/protocol-ts/http";

import {
  clientTxnID,
  fetchHTTPOperation,
  fetchMultipartHTTPOperation,
} from "./browserApi";

export type ReferencePackAction = "activate" | "disable" | "reverify";

export type ReferencePackVersion =
  ListReferencePacksResponse["data"]["pack_versions"][number];

export type ReferencePackJobResource = GetJobResponse["data"];

export type ReferencePackPaging = NonNullable<
  ListReferencePacksResponse["meta"]["paging"]
>;

export type ReferencePackQuery = {
  active: string;
  packVersionState: string;
  search: string;
  verificationResult: string;
};

export function listReferencePacks(options: {
  cursorToken?: string | null | undefined;
  query: ReferencePackQuery;
}) {
  const cursorToken = options.cursorToken?.trim() ?? "";
  return fetchHTTPOperation<ListReferencePacksResponse>({
    operationID: "listReferencePacks",
    query: {
      limit: 100,
      ...(cursorToken === "" ? {} : { cursor_token: cursorToken }),
      ...(options.query.search === "" ? {} : { search: options.query.search }),
      ...(options.query.packVersionState === ""
        ? {}
        : { pack_version_state: options.query.packVersionState }),
      ...(options.query.verificationResult === ""
        ? {}
        : { verification_result: options.query.verificationResult }),
      ...(options.query.active === "" ? {} : { active: options.query.active }),
    },
  });
}

export function loadReferencePackJob(jobID: string) {
  return fetchHTTPOperation<GetJobResponse>({
    operationID: "getJob",
    pathParameters: { job_id: jobID },
  });
}

export async function importReferencePackBundle(file: File) {
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

  return fetchMultipartHTTPOperation<ImportReferencePackResponse>({
    body: form,
    operationID: "importReferencePack",
  });
}

function referencePackActionRequest(
  action: ReferencePackAction,
): ActivateReferencePackVersionRequest &
  DisableReferencePackVersionRequest &
  ReverifyReferencePackVersionRequest {
  return {
    client_txn_id: clientTxnID(`reference-pack-${action}`),
  };
}

function referencePackPathParameters(pack: ReferencePackVersion) {
  return {
    pack_key: pack.pack_key,
    pack_version: pack.pack_version,
  };
}

export function activateReferencePackVersion(pack: ReferencePackVersion) {
  return fetchHTTPOperation<ActivateReferencePackVersionResponse>({
    operationID: "activateReferencePackVersion",
    pathParameters: referencePackPathParameters(pack),
    init: {
      method: "POST",
      body: JSON.stringify(referencePackActionRequest("activate")),
    },
  });
}

export function disableReferencePackVersion(pack: ReferencePackVersion) {
  return fetchHTTPOperation<DisableReferencePackVersionResponse>({
    operationID: "disableReferencePackVersion",
    pathParameters: referencePackPathParameters(pack),
    init: {
      method: "POST",
      body: JSON.stringify(referencePackActionRequest("disable")),
    },
  });
}

export function reverifyReferencePackVersion(pack: ReferencePackVersion) {
  return fetchHTTPOperation<ReverifyReferencePackVersionResponse>({
    operationID: "reverifyReferencePackVersion",
    pathParameters: referencePackPathParameters(pack),
    init: {
      method: "POST",
      body: JSON.stringify(referencePackActionRequest("reverify")),
    },
  });
}

export function refreshReferencePacks(options: {
  all: boolean;
  packKeys: readonly string[];
}) {
  const body: RefreshReferencePacksRequest = {
    client_txn_id: clientTxnID("reference-pack-refresh"),
    ...(options.all ? {} : { pack_keys: [...options.packKeys] }),
  };
  return fetchHTTPOperation<RefreshReferencePacksResponse>({
    operationID: "refreshReferencePacks",
    init: {
      method: "POST",
      body: JSON.stringify(body),
    },
  });
}

export function cancelReferencePackJob(jobID: string) {
  const body: CancelJobRequest = {
    client_txn_id: clientTxnID("reference-pack-cancel"),
  };
  return fetchHTTPOperation<CancelJobResponse>({
    operationID: "cancelJob",
    pathParameters: { job_id: jobID },
    init: {
      method: "POST",
      body: JSON.stringify(body),
    },
  });
}
