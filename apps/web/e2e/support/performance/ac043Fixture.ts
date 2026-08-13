import type {
  ViewRow,
  PasteWorkbookClipboardRequest as WorkbookClipboardPasteRequest,
} from "@cartulary/protocol-ts/http";
import {
  type CartularyAc043PredicateId,
  cartularyAc043PerformanceContract,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Browser, BrowserContext, Page } from "@playwright/test";

import { csrfHeaders } from "../auth/browserSession";
import { createIncidentMemberUser } from "../incidents/memberships";
import { apiBase, webBase } from "../runtime/configuration";
import { uniqueEmail, uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

const batchSize = 500;
const fixture = cartularyAc043PerformanceContract.fixture;

export type Ac043Fixture = {
  fixtureId: typeof fixture.fixtureId;
  identityRows: ViewRow[];
  hostRows: ViewRow[];
  targetRows: readonly [ViewRow, ViewRow, ...ViewRow[]];
  timelineRows: ViewRow[];
};

export type Ac043TrafficDriver = {
  qualifiedBackgroundSessions: number;
  sessionCount: number;
  updatesPerSecond: number;
  stop: () => Promise<void>;
};

type LoginSessionTracker = {
  loginTrackedUser: (
    page: Page,
    details: {
      createdBy: string;
      email: string;
      password: string;
      purpose: string;
      userId: string;
    },
  ) => Promise<void>;
};

export function ac043FixtureCacheKey(input: {
  migrationDigest: string;
  sourceContractDigest: string;
}) {
  return [
    "cartulary.ac043.fixture.v1",
    input.migrationDigest,
    input.sourceContractDigest,
    fixture.seed,
  ].join(":");
}

export function timelineFixtureMatrix(startIndex: number, count: number) {
  return Array.from({ length: count }, (_, offset) => {
    const index = startIndex + offset;
    const associated = index % 20 === 0;
    const associationIndex = Math.floor(index / 20);
    const suffix = String(associationIndex).padStart(4, "0");
    return [
      associated
        ? `Performance Timeline ${String(index).padStart(5, "0")} perf-host-${suffix} perf-identity-${suffix}@example.test https://fixture-${suffix}.example.test/trace`
        : `Performance Timeline ${String(index).padStart(5, "0")}`,
      associated ? `perf-host-${suffix}` : "",
      associated ? `perf-identity-${suffix}@example.test` : "",
      associated ? `perf-tag-${suffix}` : "",
      associated ? `https://fixture-${suffix}.example.test/trace` : "",
    ];
  });
}

export async function assembleAc043Fixture(
  page: Page,
  incidentId: string,
): Promise<Ac043Fixture> {
  const hostRows = await pasteCreatedRows(page, incidentId, {
    columns: ["host.display_name", "host.hostname"],
    matrix: Array.from({ length: fixture.hostRows }, (_, index) => {
      const suffix = String(index).padStart(4, "0");
      return [`Performance Host ${suffix}`, `perf-host-${suffix}`];
    }),
    startFieldKey: "host.display_name",
    viewSchemaId: "cartulary.view.hosts.v1",
  });
  const identityRows = await pasteCreatedRows(page, incidentId, {
    columns: ["identity.display_name", "identity.upn"],
    matrix: Array.from({ length: fixture.identityRows }, (_, index) => {
      const suffix = String(index).padStart(4, "0");
      return [
        `Performance Identity ${suffix}`,
        `perf-identity-${suffix}@example.test`,
      ];
    }),
    startFieldKey: "identity.display_name",
    viewSchemaId: "cartulary.view.identities.v1",
  });
  const timelineRows: ViewRow[] = [];
  for (let start = 0; start < fixture.timelineRows; start += batchSize) {
    timelineRows.push(
      ...(await pasteCreatedRows(page, incidentId, {
        columns: [
          "timeline.activity_synopsis_text",
          "timeline.host_refs",
          "timeline.identity_refs",
          "timeline.tags",
          "timeline.data_source_text",
        ],
        matrix: timelineFixtureMatrix(
          start,
          Math.min(batchSize, fixture.timelineRows - start),
        ),
        startFieldKey: "timeline.activity_synopsis_text",
        viewSchemaId: "cartulary.view.timeline.v2",
      })),
    );
  }
  requireExactRows(hostRows, fixture.hostRows, "Host");
  requireExactRows(identityRows, fixture.identityRows, "Identity");
  requireExactRows(timelineRows, fixture.timelineRows, "Timeline");
  if (timelineRows.length < 26) {
    throw new Error("AC-043 fixture lacks deterministic measurement targets");
  }
  return {
    fixtureId: fixture.fixtureId,
    hostRows,
    identityRows,
    targetRows: [
      timelineRows[10],
      timelineRows[11],
      ...timelineRows.slice(12, 36),
    ] as [ViewRow, ViewRow, ...ViewRow[]],
    timelineRows,
  };
}

export async function startAc043Traffic(
  browser: Browser,
  sessionTracker: LoginSessionTracker,
  page: Page,
  incidentId: string,
  rows: readonly ViewRow[],
  predicateId: CartularyAc043PredicateId,
): Promise<Ac043TrafficDriver> {
  if (rows.length < fixture.backgroundSessions) {
    throw new Error("AC-043 background row pool is undersized");
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
    for (let index = 0; index < fixture.backgroundSessions; index += 1) {
      const member = await createIncidentMemberUser(page, incidentId, {
        display_name: `AC-043 Background Analyst ${index + 1}`,
        email: uniqueEmail(`ac043-background-${index + 1}`),
        initial_password: "Ac043BackgroundPass!7",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      });
      const context = await browser.newContext();
      const backgroundPage = await context.newPage();
      await sessionTracker.loginTrackedUser(backgroundPage, {
        createdBy: predicateId,
        email: member.email,
        password: member.initial_password,
        purpose: `AC-043 traffic session ${index + 1}`,
        userId: member.user_id,
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

async function pasteCreatedRows(
  page: Page,
  incidentId: string,
  input: {
    columns: [string, ...string[]];
    matrix: string[][];
    startFieldKey: string;
    viewSchemaId: WorkbookClipboardPasteRequest["view_schema_id"];
  },
) {
  const rows: ViewRow[] = [];
  for (let start = 0; start < input.matrix.length; start += batchSize) {
    const matrix = input.matrix.slice(start, start + batchSize);
    const request: WorkbookClipboardPasteRequest = {
      client_txn_id: uniqueTxn("ac043-fixture"),
      clipboard_text: matrix
        .map((values) => values.map(escapeTSVCell).join("\t"))
        .join("\n"),
      columns: input.columns,
      format: "tsv",
      start_field_key: input.startFieldKey,
      targets: matrix.map(() => ({ kind: "create" })) as [
        { kind: "create" },
        ...Array<{ kind: "create" }>,
      ],
      view_schema_id: input.viewSchemaId,
    };
    const response = await publicHttpOperation({
      body: request,
      headers: await csrfHeaders(page),
      operationID: "pasteWorkbookClipboard",
      pathParameters: {
        incident_id: incidentId,
        view_schema_id: input.viewSchemaId,
      },
      request: atJsonOrigin(page.request, apiBase),
    });
    if (!response.ok) {
      throw new Error(
        `AC-043 ${input.viewSchemaId} fixture paste failed with HTTP ${response.status}`,
      );
    }
    rows.push(...response.payload.data.rows);
  }
  return rows;
}

function requireExactRows(rows: ViewRow[], expected: number, label: string) {
  if (rows.length !== expected) {
    throw new Error(
      `AC-043 ${label} fixture expected ${expected} rows, got ${rows.length}`,
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
    throw new Error(`AC-043 ${label} fixture contains an unstable row`);
  }
}

function escapeTSVCell(value: string) {
  return value
    .replaceAll("\t", " ")
    .replaceAll("\n", " ")
    .replaceAll("\r", " ");
}
