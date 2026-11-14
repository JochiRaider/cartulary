import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { cartularyLargeGridFixture } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { afterEach, describe, expect, it, vi } from "vitest";

import {
  establishAc043PresenceClient,
  parseAc043RuntimeBundle,
  prepareAc043Snapshot,
} from "./ac043Snapshot";

const admissionMocks = vi.hoisted(() => ({
  login: vi.fn(),
  logout: vi.fn(),
  operation: vi.fn(),
}));
vi.mock("../auth/sessions", () => ({
  loginBootstrapControlPlaneContext: admissionMocks.login,
  logoutAndVerify: admissionMocks.logout,
}));
vi.mock("../transport/publicHttpOperationClient", () => ({
  publicHttpOperation: admissionMocks.operation,
}));

const snapshotKey = "a".repeat(64);

function runtimeBundle() {
  return {
    schema_id: "cartulary.performance_fixture_runtime.v2",
    fixture_profile_id: "timeline_large_grid_snapshot_v2",
    snapshot_key: snapshotKey,
    credential_sets: [
      {
        set_id: "background_analysts",
        credential_kind: "email_password",
        credentials: Array.from({ length: 24 }, (_, index) => ({
          principal: `private-${index}@example.test`,
          secret: `private-runtime-password-${String(index).padStart(2, "0")}`,
        })),
      },
    ],
  };
}

describe("AC-043 snapshot runtime bundle", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
    vi.resetAllMocks();
  });
  it("preserves the typed supported-envelope fixture and traffic shape", async () => {
    expect(cartularyLargeGridFixture).toEqual({
      fixtureId: "cartulary.perf.large_grid.v1",
      seed: 20260405,
      timelineRows: 20_000,
      hostRows: 1_000,
      identityRows: 1_000,
      tagAssignments: 1_000,
      mentionAssignments: 1_000,
      linkAssignments: 1_000,
      analystSessions: 25,
      backgroundSessions: 24,
      backgroundUpdateIntervalMs: 5_000,
      backgroundUpdatesPerSecond: 4.8,
      trafficTraceId: "cartulary.perf.live_updates_25sessions.v1",
    });
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-13T12:00:00Z"));
    const sent: unknown[] = [];
    const listeners = new Map<string, (event?: unknown) => void>();
    class FixtureSocket {
      constructor(readonly url: string) {
        expect(url).toContain("/ws/v1/incidents/incident-fixture");
      }
      addEventListener(kind: string, listener: (event?: unknown) => void) {
        listeners.set(kind, listener);
      }
      send(raw: string) {
        sent.push(JSON.parse(raw));
      }
      close() {}
    }
    vi.stubGlobal("WebSocket", FixtureSocket);
    vi.stubGlobal("window", { setTimeout, clearTimeout });
    const page = {
      goto: vi.fn(async () => undefined),
      evaluate: vi.fn(
        async (run: (input: unknown) => Promise<void>, input: unknown) => {
          const result = run(input);
          listeners.get("open")?.();
          listeners.get("message")?.({
            data: JSON.stringify({ type: "hello_ack", payload: {} }),
          });
          await result;
          listeners.get("message")?.({
            data: JSON.stringify({ type: "ping", payload: {} }),
          });
        },
      ),
    };
    await establishAc043PresenceClient(
      page as unknown as Page,
      "incident-fixture",
      0,
    );
    expect(sent).toEqual([
      {
        type: "hello",
        incident_id: "incident-fixture",
        emitted_at: "2026-07-13T12:00:00.000Z",
        payload: {
          client_instance_id: "ac043-background-1",
          presence: {
            mode: "viewing",
            sheet_ref: {
              kind: "view_schema",
              id: "cartulary.view.timeline.v2",
            },
          },
        },
      },
      {
        type: "pong",
        incident_id: "incident-fixture",
        emitted_at: "2026-07-13T12:00:00.000Z",
        payload: {},
      },
    ]);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("retires foreground admission transports after operation and logout failures", async () => {
    const root = mkdtempSync(join(tmpdir(), "cartulary-ac043-admission-"));
    try {
      const bundlePath = join(root, "runtime.json");
      writeFileSync(bundlePath, JSON.stringify(runtimeBundle()));
      vi.stubEnv(
        "CARTULARY_FIXTURE_PROFILE_ID",
        "timeline_large_grid_snapshot_v2",
      );
      vi.stubEnv("CARTULARY_FIXTURE_SNAPSHOT_KEY", snapshotKey);
      vi.stubEnv("CARTULARY_PERFORMANCE_FIXTURE_RUNTIME_BUNDLE", bundlePath);
      for (const failureAt of ["none", "discovery", "membership"] as const) {
        for (const logoutFails of [false, true]) {
          for (const disposeFails of [false, true]) {
            vi.resetAllMocks();
            const events: string[] = [];
            const primary = new Error("admission failure");
            const retirement = new Error("logout failure");
            const disposal = new Error("disposal failure");
            const nextPhase = new Error("stop before snapshot reads");
            const request = {
              dispose: vi.fn(async () => {
                events.push("dispose");
                if (disposeFails) throw disposal;
              }),
            };
            const storageState = { cookies: [], origins: [] };
            admissionMocks.login.mockImplementation(async () => {
              events.push("login");
              return { request, storageState };
            });
            admissionMocks.logout.mockImplementation(async () => {
              events.push("logout");
              if (logoutFails) throw retirement;
            });
            admissionMocks.operation.mockImplementation(
              async (input: {
                operationID: string;
                request: unknown;
                body?: unknown;
              }) => {
                if (input.operationID === "getIncidentWorkbookStartup") {
                  events.push("startup");
                  throw nextPhase;
                }
                expect(input.request).toBe(request);
                if (input.operationID === "listVisibleIncidents") {
                  events.push("discovery");
                  if (failureAt === "discovery") throw primary;
                  return {
                    ok: true,
                    payload: {
                      data: {
                        incidents: [
                          {
                            incident_id: "incident-1",
                            incident_key: "AC043-PERF-20260405",
                          },
                        ],
                      },
                    },
                  };
                }
                expect(input.operationID).toBe("createIncidentMembership");
                expect(input.body).toEqual({
                  email: "analyst@example.test",
                  role: "reviewer",
                  client_txn_id: expect.any(String),
                });
                events.push("membership");
                if (failureAt === "membership") throw primary;
                return { ok: true };
              },
            );
            const outcome = await prepareAc043Snapshot(
              { request: {} } as unknown as Page,
              {
                user_id: "analyst-1",
                email: "analyst@example.test",
                password: "fixture-password",
              },
              { foregroundRole: "reviewer" },
            ).then(
              () => undefined,
              (error: unknown) => error,
            );
            expect(admissionMocks.login).toHaveBeenCalledOnce();
            expect(admissionMocks.logout).toHaveBeenCalledOnce();
            expect(admissionMocks.logout).toHaveBeenCalledWith(
              request,
              storageState,
              "AC-043 foreground snapshot admission",
            );
            expect(request.dispose).toHaveBeenCalledOnce();
            expect(events).toEqual([
              "login",
              "discovery",
              ...(failureAt === "discovery" ? [] : ["membership"]),
              "logout",
              "dispose",
              ...(failureAt === "none" && !logoutFails && !disposeFails
                ? ["startup"]
                : []),
            ]);
            if (logoutFails || disposeFails) {
              expect(outcome).toBeInstanceOf(AggregateError);
              expect((outcome as AggregateError).errors).toEqual([
                ...(failureAt === "none" ? [] : [primary]),
                ...(logoutFails ? [retirement] : []),
                ...(disposeFails ? [disposal] : []),
              ]);
              expect((outcome as AggregateError).cause).toBe(
                failureAt === "none" ? undefined : primary,
              );
            } else {
              expect(outcome).toBe(failureAt === "none" ? nextPhase : primary);
            }
          }
        }
      }
    } finally {
      rmSync(root, { recursive: true, force: true });
    }
  });

  it("accepts exactly 24 private accounts for the admitted snapshot", () => {
    const parsed = parseAc043RuntimeBundle(JSON.stringify(runtimeBundle()), {
      fixtureProfileId: "timeline_large_grid_snapshot_v2",
      snapshotKey,
    });

    expect(parsed.backgroundAccounts).toHaveLength(24);
    expect(parsed.fixtureProfileId).toBe("timeline_large_grid_snapshot_v2");
    expect(parsed.snapshotKey).toBe(snapshotKey);
  });

  it("rejects stale identity, duplicate accounts, and extra fields", () => {
    const stale = runtimeBundle();
    stale.snapshot_key = "b".repeat(64);
    expect(() =>
      parseAc043RuntimeBundle(JSON.stringify(stale), {
        fixtureProfileId: "timeline_large_grid_snapshot_v2",
        snapshotKey,
      }),
    ).toThrow(/identity is inconsistent/u);

    const duplicate = runtimeBundle();
    const credentialSet = duplicate.credential_sets.at(0);
    const firstCredential = credentialSet?.credentials.at(0);
    if (!credentialSet || !firstCredential) {
      throw new Error("runtime bundle fixture is incomplete");
    }
    credentialSet.credentials[1] = firstCredential as {
      principal: string;
      secret: string;
    };
    expect(() =>
      parseAc043RuntimeBundle(JSON.stringify(duplicate), {
        fixtureProfileId: "timeline_large_grid_snapshot_v2",
        snapshotKey,
      }),
    ).toThrow(/duplicate accounts/u);

    expect(() =>
      parseAc043RuntimeBundle(
        JSON.stringify({ ...runtimeBundle(), runtime_path: "/private" }),
        {
          fixtureProfileId: "timeline_large_grid_snapshot_v2",
          snapshotKey,
        },
      ),
    ).toThrow(/invalid field set/u);
  });
});
