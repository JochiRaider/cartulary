import {
  buildHTTPOperationPath,
  httpOperationBindings,
} from "@cartulary/protocol-ts/http";

type RequestPageLike = {
  evaluate?: (
    pageFunction: (arg?: unknown) => unknown,
    arg?: unknown,
  ) => Promise<unknown>;
  request?: {
    post: (
      url: string,
      options: {
        data?: unknown;
        headers?: Record<string, string>;
      },
    ) => Promise<{
      ok: () => boolean;
      status?: () => number;
    }>;
  };
};

export async function fillDownGridCells(options: {
  apiBase: string;
  clientTxnId?: string | undefined;
  csrfHeaders?: Record<string, string> | undefined;
  fieldKey: string;
  incidentId: string;
  page: RequestPageLike;
  targetRecords: readonly {
    readonly baseRowVersion: number;
    readonly recordId: string;
  }[];
  surface: string;
  value: string;
}) {
  if (options.page.request === undefined) {
    if (options.page.evaluate === undefined) {
      throw new Error(
        `fillDownGridCells(${options.surface}) requires page.evaluate() or page.request.post() support`,
      );
    }
  }
  const operationID = "applyWorkbookBulkMutation" as const;
  const path = buildHTTPOperationPath(operationID, {
    incident_id: options.incidentId,
    view_schema_id: options.surface,
  });
  const apiURL = `${options.apiBase}${path}`;
  const data = {
    view_schema_id: options.surface,
    client_txn_id:
      options.clientTxnId ?? `${options.surface}-fill-down-${Date.now()}`,
    kind: "fill_down_v1",
    field_key: options.fieldKey,
    value: options.value,
    targets: options.targetRecords.map((target) => ({
      record_id: target.recordId,
      base_row_version: target.baseRowVersion,
    })),
  };
  const headers = {
    "content-type": "application/json",
    ...(options.csrfHeaders ?? {}),
  };
  if (options.page.evaluate !== undefined) {
    const response = (await options.page.evaluate(
      async (arg) => {
        const request = arg as {
          data: unknown;
          headers: Record<string, string>;
          method: string;
          url: string;
        };
        const result = await fetch(request.url, {
          method: request.method,
          credentials: "include",
          headers: request.headers,
          body: JSON.stringify(request.data),
        });
        return { ok: result.ok, status: result.status };
      },
      {
        data,
        headers,
        method: httpOperationBindings[operationID].method,
        url: path,
      },
    )) as { ok?: unknown; status?: unknown };
    return {
      ok: () => response.ok === true,
      status: () =>
        typeof response.status === "number" ? response.status : Number.NaN,
    };
  }
  if (options.page.request === undefined) {
    throw new Error(
      `fillDownGridCells(${options.surface}) requires page.request.post() support`,
    );
  }
  const requestOptions: {
    data: unknown;
    headers?: Record<string, string>;
  } = { data };
  if (options.csrfHeaders !== undefined) {
    requestOptions.headers = options.csrfHeaders;
  }
  return options.page.request.post(apiURL, requestOptions);
}

export function assertRecordFieldMutationAnchor(options: {
  actualRecordId: string;
  body: Record<string, unknown>;
  expectedRecordId: string;
  expectedValue?: unknown;
  fieldKey: string;
}) {
  const { actualRecordId, body, expectedRecordId, expectedValue, fieldKey } =
    options;
  if (actualRecordId !== expectedRecordId) {
    throw new Error(
      `Expected mutation for record_id ${expectedRecordId}, received ${actualRecordId}`,
    );
  }
  const changes = Array.isArray(body.changes) ? body.changes : [];
  const change = changes.find(
    (candidate): candidate is { field_key: string; value?: unknown } =>
      typeof candidate === "object" &&
      candidate !== null &&
      "field_key" in candidate &&
      candidate.field_key === fieldKey,
  );
  if (!change) {
    throw new Error(`Expected mutation body to include field_key ${fieldKey}`);
  }
  if ("expectedValue" in options && change.value !== expectedValue) {
    throw new Error(
      `Expected ${fieldKey} mutation value ${String(expectedValue)}, received ${String(change.value)}`,
    );
  }
}
