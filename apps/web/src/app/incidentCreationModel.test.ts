import type { CreateIncidentResponse } from "@cartulary/protocol-ts/http";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred, jsonResponse } from "../testing/fetchMockTestSupport";
import { createIncident } from "./api/appShellClient";
import {
  IncidentCreationController,
  type IncidentCreationPorts,
} from "./incidentCreationModel";

const incident: CreateIncidentResponse["data"] = {
  incident_id: "00000000-0000-4000-8000-000000001401",
  incident_key: "IR-CREATE",
  title: "Creation draft",
  created_by_user_id: "00000000-0000-4000-8000-000000000001",
  updated_by_user_id: "00000000-0000-4000-8000-000000000001",
  created_at: "2026-04-20T12:00:00Z",
  updated_at: "2026-04-20T12:00:00Z",
  incident_version: 1,
  status: "active",
  closed_at: null,
  current_phase: null,
  description: null,
  severity: null,
  primary_external_case_ref: null,
  tlp: null,
};
const success = {
  ok: true as const,
  status: 201,
  payload: { data: incident, meta: { request_id: "request-test" } },
};
const rejection = (code: string, status = 400, details = {}) => ({
  ok: false as const,
  status,
  payload: { error: { code, status, details } },
});
const controllers: IncidentCreationController[] = [];
function setup(overrides: Partial<IncidentCreationPorts> = {}) {
  const ports = {
    create: vi.fn<IncidentCreationPorts["create"]>().mockResolvedValue(success),
    newTransactionId: vi
      .fn()
      .mockReturnValueOnce("txn-one")
      .mockReturnValue("txn-two"),
    isCurrentSession: vi.fn().mockReturnValue(true),
    sessionLost: vi.fn(),
    openIncident: vi
      .fn<IncidentCreationPorts["openIncident"]>()
      .mockResolvedValue("opened"),
    ...overrides,
  };
  const controller = new IncidentCreationController(ports);
  controllers.push(controller);
  controller.setSession("actor:session-one");
  controller.open();
  controller.change("incident_key", "IR-CREATE");
  controller.change("title", "Creation draft");
  return { controller, ports };
}
async function settle() {
  for (let step = 0; step < 6; step += 1) await Promise.resolve();
}
afterEach(() => {
  for (const controller of controllers.splice(0)) controller.dispose();
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe("incident creation operation", () => {
  it("captures only declared values and omits blank optional metadata", async () => {
    const { controller, ports } = setup();
    controller.change("incident_key", "  IR-E\u0301  ");
    controller.change("description", "  first\n\tsecond  ");
    controller.change("severity", "\u0085 \t");
    controller.change("tlp", "TLP:AMBER+STRICT");
    controller.submit();
    await settle();
    expect(ports.create).toHaveBeenCalledWith(
      {
        client_txn_id: "txn-one",
        incident_key: "  IR-E\u0301  ",
        title: "Creation draft",
        description: "  first\n\tsecond  ",
        tlp: "TLP:AMBER+STRICT",
      },
      expect.any(AbortSignal),
    );
  });

  it("guards synchronous dispatch and retains the exact uncertain request", async () => {
    const pending = deferred<typeof success>();
    const create = vi
      .fn<IncidentCreationPorts["create"]>()
      .mockReturnValueOnce(pending.promise)
      .mockResolvedValue({ ...success, status: 200 });
    const { controller } = setup({ create });
    controller.submit();
    controller.submit();
    controller.retry();
    expect(create).toHaveBeenCalledTimes(1);
    pending.reject(new TypeError("lost response"));
    await settle();
    const captured = create.mock.calls[0]?.[0];
    controller.change("title", "must not replace unresolved payload");
    controller.startNewAttempt();
    controller.close();
    controller.open();
    expect(controller.getSnapshot().draft.title).toBe("Creation draft");
    controller.retry();
    controller.retry();
    await settle();
    expect(create).toHaveBeenCalledTimes(2);
    expect(create.mock.calls[1]?.[0]).toBe(captured);
    expect(controller.getSnapshot().operation.kind).toBe("created");
  });

  it("times out observation and ignores a later original success", async () => {
    vi.useFakeTimers();
    const pending = deferred<typeof success>();
    const create = vi
      .fn<IncidentCreationPorts["create"]>()
      .mockReturnValueOnce(pending.promise)
      .mockResolvedValue(success);
    const { controller, ports } = setup({ create });
    controller.submit();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    expect(create.mock.calls[0]?.[1].aborted).toBe(true);
    controller.retry();
    await settle();
    expect(ports.openIncident).toHaveBeenCalledTimes(1);
    pending.resolve(success);
    await settle();
    expect(ports.openIncident).toHaveBeenCalledTimes(1);
    expect(create.mock.calls[1]?.[0]).toBe(create.mock.calls[0]?.[0]);
  });

  it("maps public optional errors and permits a corrected new attempt", async () => {
    const create = vi
      .fn<IncidentCreationPorts["create"]>()
      .mockResolvedValueOnce(
        rejection("invalid_incident_create", 400, {
          field: "severity",
          reason_code: "field_too_long",
        }),
      )
      .mockResolvedValue(success);
    const { controller } = setup({ create });
    controller.change("severity", "original");
    controller.submit();
    await settle();
    expect(controller.getSnapshot().detailsOpen).toBe(true);
    expect(controller.getSnapshot().errors.severity).toContain("Shorten");
    expect(controller.getSnapshot().draft.severity).toBe("original");
    controller.change("severity", "high");
    controller.submit();
    await settle();
    expect(create.mock.calls[1]?.[0]).toMatchObject({
      severity: "high",
      client_txn_id: "txn-two",
    });
  });

  it("requires explicit replacement after a transaction conflict", async () => {
    const create = vi
      .fn<IncidentCreationPorts["create"]>()
      .mockResolvedValue(rejection("client_txn_conflict", 409));
    const { controller } = setup({ create });
    controller.submit();
    await settle();
    controller.submit();
    expect(create).toHaveBeenCalledTimes(1);
    controller.startNewAttempt();
    controller.submit();
    await settle();
    expect(create.mock.calls[1]?.[0].client_txn_id).toBe("txn-two");
  });

  it("does not turn a rejected replay into proof of earlier rollback", async () => {
    const create = vi
      .fn<IncidentCreationPorts["create"]>()
      .mockRejectedValueOnce(new TypeError("lost response"))
      .mockResolvedValue(rejection("authorization_denied", 403));
    const { controller } = setup({ create });
    controller.submit();
    await settle();
    controller.retry();
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    expect(controller.getSnapshot().message).toContain(
      "may still have created",
    );
    controller.change("title", "edited");
    expect(controller.getSnapshot().draft.title).toBe("Creation draft");
  });

  it("retains confirmation when membership refresh fails and retries only handoff", async () => {
    const openIncident = vi
      .fn<IncidentCreationPorts["openIncident"]>()
      .mockResolvedValueOnce("unavailable")
      .mockResolvedValue("opened");
    const { controller, ports } = setup({ openIncident });
    controller.submit();
    await settle();
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "created",
      handoff: "failed",
    });
    controller.submit();
    controller.retry();
    await controller.openCreated();
    expect(ports.create).toHaveBeenCalledTimes(1);
    expect(openIncident).toHaveBeenCalledTimes(2);
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "created",
      handoff: "opened",
    });
  });

  it("retains collapse recovery without restoring automatic navigation on reopen", async () => {
    const pending = deferred<typeof success>();
    const { controller, ports } = setup({ create: () => pending.promise });
    controller.submit();
    controller.close();
    controller.open();
    pending.resolve(success);
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("created");
    expect(ports.openIncident).not.toHaveBeenCalled();
    await controller.openCreated();
    expect(ports.openIncident).toHaveBeenCalledTimes(1);
  });

  it("bounds handoff observation and prevents its late result from replacing a new draft", async () => {
    vi.useFakeTimers();
    const pending = deferred<"opened">();
    const { controller } = setup({ openIncident: () => pending.promise });
    controller.submit();
    await settle();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(controller.getSnapshot().operation).toMatchObject({
      kind: "created",
      handoff: "failed",
    });
    controller.startNewAttempt();
    controller.change("title", "new draft");
    pending.resolve("opened");
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("editing");
    expect(controller.getSnapshot().draft.title).toBe("new draft");
  });

  it("invalidates an in-progress handoff when a destination is chosen", async () => {
    const pending = deferred<"opened">();
    const openIncident = vi
      .fn<IncidentCreationPorts["openIncident"]>()
      .mockReturnValue(pending.promise);
    const { controller } = setup({ openIncident });
    controller.submit();
    await settle();
    controller.leaveSurface();
    expect(openIncident.mock.calls[0]?.[2]()).toBe(false);
    expect(openIncident.mock.calls[0]?.[1].aborted).toBe(true);
    pending.resolve("opened");
    await settle();
  });

  it("clears drafts and rejects late account and same-user session responses", async () => {
    const pending = deferred<typeof success>();
    const { controller, ports } = setup({ create: () => pending.promise });
    controller.submit();
    controller.setSession(null);
    controller.setSession("actor:session-two");
    controller.open();
    controller.change("title", "new session draft");
    pending.resolve(success);
    await settle();
    expect(controller.getSnapshot().draft.title).toBe("new session draft");
    expect(controller.getSnapshot().draft.incident_key).toBe("");
    expect(ports.openIncident).not.toHaveBeenCalled();
    controller.setSession("different-actor:session-two");
    expect(controller.getSnapshot().draft.title).toBe("");
  });

  it("clears account state on session loss and leaves forbidden errors local", async () => {
    const { controller, ports } = setup({
      create: vi
        .fn()
        .mockResolvedValueOnce(rejection("csrf_failed", 403))
        .mockResolvedValue(rejection("session_required", 401)),
    });
    controller.submit();
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("rejected");
    expect(controller.getSnapshot().draft.title).toBe("Creation draft");
    expect(ports.sessionLost).not.toHaveBeenCalled();
    controller.submit();
    await settle();
    expect(ports.sessionLost).toHaveBeenCalledTimes(1);
    expect(controller.getSnapshot().draft.title).toBe("");
  });

  it("requires Unicode-nonblank identity fields before dispatch", () => {
    const { controller, ports } = setup();
    controller.change("incident_key", "\u0085\u2003");
    controller.change("title", "\t ");
    controller.submit();
    expect(controller.getSnapshot().errors).toEqual({
      incident_key: "Incident key is required.",
      title: "Title is required.",
    });
    expect(ports.create).not.toHaveBeenCalled();
  });

  it("preserves explicit nulls and validates both typed success statuses", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(success.payload, 201))
      .mockResolvedValueOnce(jsonResponse(success.payload, 200));
    vi.stubGlobal("fetch", fetchMock);
    const request = {
      client_txn_id: "txn-captured",
      incident_key: "IR-CREATE",
      title: "Creation draft",
      description: null,
      severity: null,
      tlp: null,
      current_phase: null,
      primary_external_case_ref: null,
      initial_memberships: [],
    };
    for (const status of [201, 200])
      expect(
        await createIncident({ request, signal: new AbortController().signal }),
      ).toMatchObject({ ok: true, status });
    const body = JSON.parse(fetchMock.mock.calls[0]?.[1].body);
    expect(body).toEqual({
      client_txn_id: "txn-captured",
      incident_key: "IR-CREATE",
      title: "Creation draft",
      description: null,
      severity: null,
      tlp: null,
      current_phase: null,
      primary_external_case_ref: null,
    });
    expect(fetchMock.mock.calls[0]?.[0]).toBe("/api/v1/incidents");
  });

  it("classifies malformed success as uncertain at the public validation boundary", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse({ data: { incident_id: incident.incident_id } }, 201),
        ),
    );
    const { controller, ports } = setup({
      create: (request, signal) => createIncident({ request, signal }),
    });
    controller.submit();
    await vi.waitFor(() =>
      expect(controller.getSnapshot().operation.kind).toBe("uncertain"),
    );
    expect(ports.openIncident).not.toHaveBeenCalled();
    expect(controller.getSnapshot().draft.title).toBe("Creation draft");
  });
});
