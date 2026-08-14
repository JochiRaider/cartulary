import { readFileSync } from "node:fs";

import type {
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import {
  type CartularyAc043PredicateId,
  cartularyAc043PerformanceContract,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Browser, BrowserContext, Page } from "@playwright/test";

import { csrfHeaders } from "../auth/browserSession";
import {
  loginBootstrapControlPlaneContext,
  loginTrackedUserViaPage,
  logoutAndVerify,
  readCurrentSession,
} from "../auth/sessions";
import type { WorkerAdminEntry } from "../auth/workerAdmin";
import { apiBase, webBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

const fixture = cartularyAc043PerformanceContract.fixture;
const fixtureIncidentKey = `AC043-PERF-${fixture.seed}`;
const fixtureProfileId = "ac043_large_grid_snapshot_v1";
const pageLimit = 500;

type BackgroundAccount = {
  email: string;
  password: string;
};

type PagingMeta = {
  has_more: boolean;
  limit: number;
  next_cursor: string | null;
};

export type Ac043Snapshot = {
  fixtureId: typeof fixture.fixtureId;
  incidentId: string;
  timelineRows: ViewRow[];
  runtime: {
    backgroundAccounts: BackgroundAccount[];
    fixtureProfileId: typeof fixtureProfileId;
    snapshotKey: string;
  };
};

export type Ac043TrafficDriver = {
  qualifiedBackgroundSessions: number;
  sessionCount: number;
  updatesPerSecond: number;
  stop: () => Promise<void>;
};

type LoginSessionTracker = {
  captureCurrentSession: (
    page: Page,
    details: {
      createdBy: string;
      email: string;
      purpose: string;
      userId: string;
    },
  ) => Promise<void>;
};

export function parseAc043RuntimeBundle(
  contents: string,
  expected: { fixtureProfileId: string; snapshotKey: string },
) {
  let decoded: unknown;
  try {
    decoded = JSON.parse(contents);
  } catch {
    throw new Error("AC-043 private runtime bundle contains invalid JSON");
  }
  const bundle = requireRecord(decoded, "AC-043 private runtime bundle");
  requireExactKeys(
    bundle,
    ["schema_id", "fixture_profile_id", "snapshot_key", "background_accounts"],
    "AC-043 private runtime bundle",
  );
  if (
    bundle.schema_id !== "cartulary.performance_fixture_runtime.v1" ||
    bundle.fixture_profile_id !== expected.fixtureProfileId ||
    bundle.snapshot_key !== expected.snapshotKey ||
    !/^[a-f0-9]{64}$/u.test(expected.snapshotKey)
  ) {
    throw new Error("AC-043 private runtime bundle identity is inconsistent");
  }
  if (!Array.isArray(bundle.background_accounts)) {
    throw new Error("AC-043 private runtime bundle accounts must be an array");
  }
  const emails = new Set<string>();
  const accounts = bundle.background_accounts.map((rawAccount, index) => {
    const label = `AC-043 private runtime account ${index + 1}`;
    const account = requireRecord(rawAccount, label);
    requireExactKeys(account, ["email", "password"], label);
    const email = requirePrivateString(account.email, `${label} email`);
    const password = requirePrivateString(
      account.password,
      `${label} password`,
    );
    if (!email.includes("@") || email.length > 254) {
      throw new Error(`${label} email is invalid`);
    }
    if (password.length < 24 || password.length > 256) {
      throw new Error(`${label} password is invalid`);
    }
    if (emails.has(email)) {
      throw new Error(
        "AC-043 private runtime bundle contains duplicate accounts",
      );
    }
    emails.add(email);
    return { email, password };
  });
  if (accounts.length !== fixture.backgroundSessions) {
    throw new Error(
      `AC-043 private runtime bundle expected ${fixture.backgroundSessions} accounts, got ${accounts.length}`,
    );
  }
  return {
    backgroundAccounts: accounts,
    fixtureProfileId,
    snapshotKey: expected.snapshotKey,
  } as const;
}

export async function prepareAc043Snapshot(
  page: Page,
  workerAdmin: WorkerAdminEntry,
): Promise<Ac043Snapshot> {
  const fixtureProfile = requiredEnvironment("CARTULARY_FIXTURE_PROFILE_ID");
  const snapshotKey = requiredEnvironment("CARTULARY_FIXTURE_SNAPSHOT_KEY");
  if (
    fixtureProfile !== fixtureProfileId ||
    !/^[a-f0-9]{64}$/u.test(snapshotKey)
  ) {
    throw new Error(
      "AC-043 snapshot execution has an invalid fixture identity",
    );
  }
  const runtime = parseAc043RuntimeBundle(
    readFileSync(
      requiredEnvironment("CARTULARY_PERFORMANCE_FIXTURE_RUNTIME_BUNDLE"),
      "utf8",
    ),
    { fixtureProfileId: fixtureProfile, snapshotKey },
  );
  const incidentId = await admitForegroundWorker(workerAdmin);
  await validateDefaultView(page, incidentId);
  const [timelineRows, hostRows, identityRows] = await Promise.all([
    queryAllViewRows(page, incidentId, timelineViewSchemaId, {
      sort: [
        { direction: "asc", field_key: "timeline.activity_synopsis_text" },
      ],
    }),
    queryAllViewRows(page, incidentId, hostsViewSchemaId, {
      sort: [{ direction: "asc", field_key: "host.hostname" }],
    }),
    queryAllViewRows(page, incidentId, identitiesViewSchemaId, {
      sort: [{ direction: "asc", field_key: "identity.upn" }],
    }),
  ]);
  validateSnapshotRows(timelineRows, hostRows, identityRows);
  return {
    fixtureId: fixture.fixtureId,
    incidentId,
    timelineRows,
    runtime,
  };
}

export async function startAc043SnapshotTraffic(
  browser: Browser,
  sessionTracker: LoginSessionTracker,
  incidentId: string,
  rows: readonly ViewRow[],
  accounts: readonly BackgroundAccount[],
  predicateId: CartularyAc043PredicateId,
): Promise<Ac043TrafficDriver> {
  if (rows.length < fixture.backgroundSessions) {
    throw new Error("AC-043 background row pool is undersized");
  }
  if (accounts.length !== fixture.backgroundSessions) {
    throw new Error("AC-043 background account pool is inconsistent");
  }
  const contexts: BrowserContext[] = [];
  const timers: Array<
    ReturnType<typeof setTimeout> | ReturnType<typeof setInterval>
  > = [];
  const pendingUpdates = new Set<Promise<void>>();
  const trafficSessions: Array<{ page: Page; rowState: ViewRow }> = [];
  let trafficFailure: Error | null = null;
  let stopped = false;
  const update = async (backgroundPage: Page, rowState: ViewRow) => {
    if (stopped) return;
    const cookies = await backgroundPage.context().cookies(apiBase);
    const response = await publicHttpOperation({
      body: {
        base_row_version: rowState.row_version,
        client_txn_id: uniqueTxn("ac043-traffic"),
        changes: [
          {
            field_key: "timeline.data_source_text",
            value: `performance traffic ${rowState.row_version + 1}`,
          },
        ],
        view_schema_id: timelineViewSchemaId,
      },
      headers: {
        ...(await csrfHeaders(backgroundPage)),
        cookie: cookies
          .map((cookie) => `${cookie.name}=${cookie.value}`)
          .join("; "),
      },
      operationID: "patchRecord",
      pathParameters: { record_id: rowState.record_id },
      request: atJsonOrigin(backgroundPage.request, apiBase),
    });
    if (!response.ok) {
      throw new Error(
        `AC-043 traffic mutation failed with HTTP ${response.status}`,
      );
    }
    rowState.row_version = response.payload.data.row.row_version;
  };
  const firstUpdates: Array<Promise<void>> = [];
  try {
    for (const [index, account] of accounts.entries()) {
      const context = await browser.newContext();
      const backgroundPage = await context.newPage();
      await loginTrackedUserViaPage(backgroundPage, account);
      const session = await readCurrentSession(backgroundPage);
      await sessionTracker.captureCurrentSession(backgroundPage, {
        createdBy: predicateId,
        email: account.email,
        purpose: `AC-043 traffic session ${index + 1}`,
        userId: session.user_id,
      });
      await establishPresenceClient(backgroundPage, incidentId, index);
      contexts.push(backgroundPage.context());
      const assignedRow = rows[index];
      if (assignedRow === undefined) {
        throw new Error(`AC-043 background row ${index + 1} is missing`);
      }
      trafficSessions.push({
        page: backgroundPage,
        rowState: { ...assignedRow },
      });
    }
    for (const [index, trafficSession] of trafficSessions.entries()) {
      let updateChain = Promise.resolve();
      const scheduleUpdate = () => {
        const scheduled = updateChain.then(() =>
          update(trafficSession.page, trafficSession.rowState),
        );
        updateChain = scheduled.catch((error: unknown) => {
          trafficFailure ??=
            error instanceof Error ? error : new Error(String(error));
        });
        const pending = updateChain.finally(() =>
          pendingUpdates.delete(pending),
        );
        pendingUpdates.add(pending);
        return scheduled;
      };
      const initialDelay =
        (fixture.backgroundUpdateIntervalMs / fixture.backgroundSessions) *
        index;
      firstUpdates.push(
        new Promise<void>((resolve, reject) => {
          timers.push(
            setTimeout(() => {
              scheduleUpdate().then(resolve, reject);
              timers.push(
                setInterval(() => {
                  void scheduleUpdate();
                }, fixture.backgroundUpdateIntervalMs),
              );
            }, initialDelay),
          );
        }),
      );
    }
    await Promise.all(firstUpdates);
    if (firstUpdates.length !== fixture.backgroundSessions) {
      throw new Error(
        `AC-043 traffic qualification expected ${fixture.backgroundSessions} sessions, got ${firstUpdates.length}`,
      );
    }
  } catch (error) {
    stopped = true;
    for (const timer of timers) clearTimeout(timer);
    await closeBackgroundContexts(contexts).catch(() => undefined);
    throw error;
  }
  return {
    qualifiedBackgroundSessions: firstUpdates.length,
    sessionCount: contexts.length + 1,
    updatesPerSecond: fixture.backgroundUpdatesPerSecond,
    stop: async () => {
      stopped = true;
      for (const timer of timers) clearTimeout(timer);
      await Promise.allSettled([...pendingUpdates]);
      await closeBackgroundContexts(contexts);
      if (trafficFailure !== null) {
        throw trafficFailure;
      }
    },
  };
}

async function admitForegroundWorker(workerAdmin: WorkerAdminEntry) {
  const controlPlane = await loginBootstrapControlPlaneContext();
  try {
    const incidents = await publicHttpOperation({
      operationID: "listVisibleIncidents",
      query: { limit: 100, search: fixtureIncidentKey },
      request: controlPlane.request,
    });
    if (!incidents.ok) {
      throw new Error(
        `AC-043 snapshot incident query failed with HTTP ${incidents.status}`,
      );
    }
    const matches = incidents.payload.data.incidents.filter(
      (incident) => incident.incident_key === fixtureIncidentKey,
    );
    if (matches.length !== 1) {
      throw new Error(
        `AC-043 snapshot incident query resolved ${matches.length} workspaces`,
      );
    }
    const incidentId = matches[0]?.incident_id;
    if (typeof incidentId !== "string" || incidentId === "") {
      throw new Error("AC-043 snapshot incident lacks an identity");
    }
    const membership = await publicHttpOperation({
      body: {
        client_txn_id: uniqueTxn("ac043-foreground-membership"),
        email: workerAdmin.email,
        role: "editor",
      },
      operationID: "createIncidentMembership",
      pathParameters: { incident_id: incidentId },
      request: controlPlane.request,
    });
    if (!membership.ok) {
      throw new Error(
        `AC-043 foreground membership failed with HTTP ${membership.status}`,
      );
    }
    return incidentId;
  } finally {
    await logoutAndVerify(
      controlPlane.request,
      controlPlane.storageState,
      "AC-043 foreground snapshot admission",
    );
    await controlPlane.request.dispose();
  }
}

async function validateDefaultView(page: Page, incidentId: string) {
  const response = await publicHttpOperation({
    operationID: "getIncidentWorkbookStartup",
    pathParameters: { incident_id: incidentId },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `AC-043 default-view validation failed with HTTP ${response.status}`,
    );
  }
  if (
    response.payload.data.selected_view_schema_id !== timelineViewSchemaId ||
    response.payload.data.source !== "timeline"
  ) {
    throw new Error("AC-043 snapshot Timeline fallback is inconsistent");
  }
}

async function queryAllViewRows(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  baseRequest: QueryWorkbookViewRequest,
) {
  const rows: ViewRow[] = [];
  const cursors = new Set<string>();
  let cursor: string | null = null;
  do {
    const payload: QueryWorkbookViewResponse = await querySnapshotPage(
      page,
      incidentId,
      viewSchemaId,
      {
        ...baseRequest,
        limit: pageLimit,
        ...(cursor === null ? {} : { cursor_token: cursor }),
      },
    );
    rows.push(...payload.data.rows);
    const paging: PagingMeta | undefined = payload.meta.paging;
    if (!paging?.has_more) break;
    if (!paging.next_cursor || cursors.has(paging.next_cursor)) {
      throw new Error(`AC-043 ${viewSchemaId} snapshot cursor is invalid`);
    }
    cursors.add(paging.next_cursor);
    cursor = paging.next_cursor;
  } while (rows.length <= fixture.timelineRows);
  return rows;
}

async function querySnapshotPage(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  body: QueryWorkbookViewRequest,
): Promise<QueryWorkbookViewResponse> {
  const response = await publicHttpOperation({
    body,
    operationID: "queryWorkbookView",
    pathParameters: {
      incident_id: incidentId,
      view_schema_id: viewSchemaId,
    },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `AC-043 ${viewSchemaId} snapshot query failed with HTTP ${response.status}`,
    );
  }
  return response.payload;
}

function validateSnapshotRows(
  timelineRows: readonly ViewRow[],
  hostRows: readonly ViewRow[],
  identityRows: readonly ViewRow[],
) {
  requireExactRows(timelineRows, fixture.timelineRows, "Timeline");
  requireExactRows(hostRows, fixture.hostRows, "Host");
  requireExactRows(identityRows, fixture.identityRows, "Identity");
  timelineRows.forEach((row, index) => {
    const summary = row.cells["timeline.activity_synopsis_text"]?.value;
    const prefix = `Performance Timeline ${String(index).padStart(5, "0")}`;
    if (typeof summary !== "string" || !summary.startsWith(prefix)) {
      throw new Error(`AC-043 Timeline snapshot row ${index} is inconsistent`);
    }
    if (index % 20 === 0) {
      const suffix = String(index / 20).padStart(4, "0");
      if (
        !summary.includes(`perf-host-${suffix}`) ||
        !summary.includes(`perf-identity-${suffix}@example.test`) ||
        !summary.includes(`https://fixture-${suffix}.example.test/trace`)
      ) {
        throw new Error(
          `AC-043 Timeline snapshot relationship row ${index} is inconsistent`,
        );
      }
    }
  });
  hostRows.forEach((row, index) => {
    if (
      row.cells["host.hostname"]?.value !==
      `perf-host-${String(index).padStart(4, "0")}`
    ) {
      throw new Error(`AC-043 Host snapshot row ${index} is inconsistent`);
    }
  });
  identityRows.forEach((row, index) => {
    if (
      row.cells["identity.upn"]?.value !==
      `perf-identity-${String(index).padStart(4, "0")}@example.test`
    ) {
      throw new Error(`AC-043 Identity snapshot row ${index} is inconsistent`);
    }
  });
}

function requireExactRows(
  rows: readonly ViewRow[],
  expected: number,
  label: string,
) {
  if (rows.length !== expected) {
    throw new Error(
      `AC-043 ${label} snapshot expected ${expected} rows, got ${rows.length}`,
    );
  }
  if (
    rows.some(
      (row) =>
        row.record_id.trim() === "" ||
        !Number.isInteger(row.row_version) ||
        row.row_version < 1,
    )
  ) {
    throw new Error(`AC-043 ${label} snapshot contains an unstable row`);
  }
}

async function closeBackgroundContexts(contexts: readonly BrowserContext[]) {
  const closeResults = await Promise.allSettled(
    contexts.map((context) => withTimeout(context.close(), 5_000)),
  );
  const failed = closeResults.filter((result) => result.status === "rejected");
  if (failed.length > 0) {
    throw new Error(
      `AC-043 background cleanup failed for ${failed.length} contexts`,
    );
  }
}

function withTimeout<T>(operation: Promise<T>, timeoutMs: number): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    const timeout = setTimeout(
      () => reject(new Error("AC-043 background context close timed out")),
      timeoutMs,
    );
    operation.then(
      (value) => {
        clearTimeout(timeout);
        resolve(value);
      },
      (error: unknown) => {
        clearTimeout(timeout);
        reject(error);
      },
    );
  });
}

async function establishPresenceClient(
  page: Page,
  incidentId: string,
  sessionIndex: number,
) {
  const routeURL = `${webBase}/__cartulary_ac043_presence__`;
  await page.route(routeURL, async (route) => {
    await route.fulfill({
      body: "<!doctype html><html><body>AC-043 presence client</body></html>",
      contentType: "text/html",
      status: 200,
    });
  });
  await page.goto(routeURL);
  const socketURL = new URL(`/ws/v1/incidents/${incidentId}`, apiBase);
  socketURL.protocol = socketURL.protocol === "https:" ? "wss:" : "ws:";
  await page.evaluate(
    ({ clientInstanceId, url }) =>
      new Promise<void>((resolve, reject) => {
        const socket = new WebSocket(url);
        const timeout = window.setTimeout(() => {
          socket.close();
          reject(new Error("AC-043 presence client handshake timed out"));
        }, 10_000);
        socket.addEventListener("open", () => {
          socket.send(
            JSON.stringify({
              payload: {
                client_instance_id: clientInstanceId,
                presence: {
                  mode: "viewing",
                  sheet_ref: {
                    id: "cartulary.view.timeline.v2",
                    kind: "view_schema",
                  },
                },
              },
              type: "hello",
            }),
          );
        });
        socket.addEventListener("message", (event) => {
          if (typeof event.data !== "string") return;
          const message = JSON.parse(event.data) as {
            payload?: unknown;
            type?: unknown;
          };
          if (message.type === "ping") {
            socket.send(JSON.stringify({ payload: {}, type: "pong" }));
            return;
          }
          if (message.type === "hello_ack") {
            window.clearTimeout(timeout);
            Object.assign(window, { __cartularyAc043Socket: socket });
            resolve();
            return;
          }
          if (message.type === "error") {
            window.clearTimeout(timeout);
            socket.close();
            reject(new Error("AC-043 presence client was rejected"));
          }
        });
        socket.addEventListener("error", () => {
          window.clearTimeout(timeout);
          reject(new Error("AC-043 presence client socket failed"));
        });
      }),
    {
      clientInstanceId: `ac043-background-${sessionIndex + 1}`,
      url: socketURL.toString(),
    },
  );
}

function requiredEnvironment(name: string) {
  const value = process.env[name]?.trim() ?? "";
  if (value === "") {
    throw new Error(`AC-043 snapshot execution requires ${name}`);
  }
  return value;
}

function requireRecord(value: unknown, label: string) {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value as Record<string, unknown>;
}

function requireExactKeys(
  value: Record<string, unknown>,
  keys: readonly string[],
  label: string,
) {
  if (
    JSON.stringify(Object.keys(value).sort()) !==
    JSON.stringify([...keys].sort())
  ) {
    throw new Error(`${label} has an invalid field set`);
  }
}

function requirePrivateString(value: unknown, label: string) {
  if (typeof value !== "string" || value === "") {
    throw new Error(`${label} is invalid`);
  }
  return value;
}
