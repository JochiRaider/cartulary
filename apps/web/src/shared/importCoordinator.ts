import {
  apiPath,
  clientTxnID,
  csrfCookieName,
  csrfHeaderName,
  extractError,
  readCookie,
} from "../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../services/workbookApi";

type JobResource = {
  readonly job_id: string;
  readonly status:
    | "queued"
    | "running"
    | "cancel_requested"
    | "succeeded"
    | "failed"
    | "canceled";
  readonly result_summary?: {
    readonly code?: string;
    readonly resource_refs?: unknown[];
  } | null;
};

type ImportUnit = { readonly import_unit_id: string };

export async function coordinateExtensionImport(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly file: File;
  readonly mappingPayload: (clientTxnId: string) => Record<string, unknown>;
  readonly transactionPrefix: string;
  readonly onProgress?: ((message: string) => void) | undefined;
}): Promise<void> {
  options.onProgress?.("Uploading import.");
  const uploadJob = await uploadImportSession(options);
  const discovered = await pollJob(options.apiBase, uploadJob.job_id);
  const sessionId = importSessionIdFromJob(discovered);
  if (sessionId === null) {
    throw new Error("import_session_not_returned");
  }

  options.onProgress?.("Preparing mapping.");
  const unitId = await firstImportUnitID(options.apiBase, sessionId);
  await postImportJSON(
    options.apiBase,
    `/api/v1/import-sessions/${sessionId}/units/${unitId}/mapping`,
    options.mappingPayload(clientTxnID(`${options.transactionPrefix}-mapping`)),
    "PUT",
  );
  await postImportJSON(
    options.apiBase,
    `/api/v1/import-sessions/${sessionId}/units/${unitId}/select`,
    { client_txn_id: clientTxnID(`${options.transactionPrefix}-select`) },
  );

  options.onProgress?.("Applying import.");
  const applyEnvelope = await postImportJSON<{ data: JobResource }>(
    options.apiBase,
    `/api/v1/import-sessions/${sessionId}/apply`,
    { client_txn_id: clientTxnID(`${options.transactionPrefix}-apply`) },
  );
  await pollJob(options.apiBase, applyEnvelope.data.job_id);
}

async function uploadImportSession(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly file: File;
  readonly transactionPrefix: string;
}): Promise<JobResource> {
  const form = new FormData();
  form.append(
    "metadata",
    new Blob(
      [
        JSON.stringify({
          incident_id: options.incidentId,
          client_txn_id: clientTxnID(`${options.transactionPrefix}-upload`),
        }),
      ],
      { type: "application/json" },
    ),
  );
  form.append("file", options.file, options.file.name);
  const result = await fetchUploadJSON<{ data: JobResource }>(
    options.apiBase,
    "/api/v1/import-sessions",
    form,
  );
  return result.data;
}

async function firstImportUnitID(
  apiBase: string | undefined,
  sessionId: string,
): Promise<string> {
  const result = await fetchWorkbookJSON<{
    readonly data: { readonly import_units: readonly ImportUnit[] };
  }>(apiPath(apiBase, `/api/v1/import-sessions/${sessionId}/units`));
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  const unitId =
    readEnvelope<{ data: { import_units: ImportUnit[] } }>(result.payload).data
      .import_units[0]?.import_unit_id ?? "";
  if (unitId === "") {
    throw new Error("import_unit_not_returned");
  }
  return unitId;
}

async function pollJob(
  apiBase: string | undefined,
  jobId: string,
): Promise<JobResource> {
  for (let attempt = 0; attempt < 30; attempt += 1) {
    const result = await fetchWorkbookJSON<{ data: JobResource }>(
      apiPath(apiBase, `/api/v1/jobs/${jobId}`),
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
    const job = readEnvelope<{ data: JobResource }>(result.payload).data;
    if (job.status === "succeeded") {
      return job;
    }
    if (job.status === "failed" || job.status === "canceled") {
      throw new Error(`job_${job.status}`);
    }
    await new Promise((resolve) => window.setTimeout(resolve, 750));
  }
  throw new Error("job_poll_timeout");
}

async function postImportJSON<T extends object = Record<string, unknown>>(
  apiBase: string | undefined,
  path: string,
  body: Record<string, unknown>,
  method = "POST",
): Promise<T> {
  const result = await fetchWorkbookJSON<T>(apiPath(apiBase, path), {
    method,
    body: JSON.stringify(body),
  });
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return readEnvelope<T>(result.payload);
}

async function fetchUploadJSON<T>(
  apiBase: string | undefined,
  path: string,
  body: FormData,
): Promise<T> {
  const headers = new Headers();
  const csrfToken = readCookie(csrfCookieName);
  if (csrfToken !== null && csrfToken !== "") {
    headers.set(csrfHeaderName, csrfToken);
  }
  const response = await fetch(apiPath(apiBase, path), {
    method: "POST",
    credentials: "include",
    headers,
    body,
  });
  const payload = (await response.json()) as T | { error?: unknown };
  if (!response.ok) {
    const error = extractError(payload);
    throw new Error(error?.code ?? "upload_failed");
  }
  return payload as T;
}

function importSessionIdFromJob(job: JobResource): string | null {
  const refs = job.result_summary?.resource_refs;
  if (!Array.isArray(refs)) {
    return null;
  }
  for (const ref of refs) {
    if (!ref || typeof ref !== "object") {
      continue;
    }
    const candidate = ref as Record<string, unknown>;
    if (
      candidate.kind === "import_session" &&
      typeof candidate.id === "string" &&
      candidate.id.trim() !== ""
    ) {
      return candidate.id;
    }
  }
  return null;
}
