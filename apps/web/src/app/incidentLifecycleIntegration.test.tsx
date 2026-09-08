import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { StrictMode, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { sessionResource } from "../testing/appShellTestSupport";
import {
  metadataActorId as actorId,
  metadataDeferred as deferred,
  metadataEnvelope as envelope,
  metadataIncident as incident,
  metadataIncidentId as incidentId,
  metadataJSON as json,
} from "../testing/incidentMetadataTestSupport";
import { AppSessionController } from "./appSessionController";
import { IncidentLifecyclePanel } from "./IncidentLifecyclePanel";
import type { IncidentLifecycleController } from "./incidentLifecycleController";
import { IncidentMetadataController } from "./incidentMetadataController";
import { IncidentResourceController } from "./incidentResourceController";
import { useIncidentLifecycle } from "./useIncidentLifecycle";

const sessions: AppSessionController[] = [];
afterEach(() => {
  cleanup();
  for (const s of sessions.splice(0)) s.dispose();
  vi.unstubAllGlobals();
});
async function setup() {
  let identity = sessionResource({
    user_id: actorId,
    memberships: [{ incident_id: incidentId, role: "admin" }],
  });
  let controller: IncidentLifecycleController | undefined;
  const session = new AppSessionController({
    session: async () => ({
      ok: true,
      status: 200,
      payload: { data: identity, meta: { request_id: "session" } },
    }),
    preferences: async () => ({ ok: false, status: 503, payload: {} }),
    extensions: async () => ({
      ok: true,
      status: 200,
      payload: { data: { extensions: [] }, meta: { request_id: "extensions" } },
    }),
    retireLifetime: () => controller?.retire(),
  });
  sessions.push(session);
  await session.refreshSession();
  const recovery = session.recoveryPort((id) => id === incidentId);
  const write = deferred<Response>();
  const fetch = vi.fn((_: unknown, init?: RequestInit) =>
    init?.method === "POST"
      ? write.promise
      : Promise.resolve(json(envelope(incident()))),
  );
  vi.stubGlobal("fetch", fetch);
  const accepted = vi.fn();
  const lost = vi.fn();
  function Harness() {
    const [open, setOpen] = useState(true);
    const workflow = useIncidentLifecycle({
      sessionController: session,
      recovery,
      currentIncidentId: () => incidentId,
      onIncidentAccessLost: lost,
      onSessionLost: () => session.sessionLost(),
      onResourceAccepted: accepted,
    });
    controller = workflow.controller;
    // Keep the same App binding while the borrowed surface is conditionally mounted.
    return (
      <>
        <button
          type="button"
          onClick={() => {
            workflow.bindSurface(
              open
                ? null
                : {
                    incidentId,
                    activeSection: "summary",
                    currentIncidentRole: "admin",
                  },
            );
            setOpen(!open);
          }}
        >
          Toggle controls
        </button>
        {open ? (
          <IncidentLifecyclePanel controller={workflow.controller} />
        ) : null}
      </>
    );
  }
  const view = render(
    <StrictMode>
      <Harness />
    </StrictMode>,
  );
  const activeController = controller;
  if (!activeController) throw new Error("Missing lifecycle binding");
  await act(async () => activeController.setActive(true));
  await waitFor(() =>
    expect(activeController.getSnapshot().read).toBe("ready"),
  );
  return {
    controller: activeController,
    session,
    write,
    fetch,
    accepted,
    lost,
    view,
    changeSession: async (next: typeof identity) => {
      identity = next;
      await act(async () => {
        await session.refreshSession();
      });
    },
    identity: () => identity,
  };
}
const close = (c: IncidentLifecycleController) =>
  act(() => {
    c.changeReason("  retained\nreason  ");
    c.propose("close");
    c.confirm();
  });

describe("Lifecycle App integration", () => {
  it("retains reason through drawer return and current-member demotion with renewed administrator review", async () => {
    const { controller, changeSession, identity } = await setup();
    act(() => controller.changeReason("  retained\nreason  "));
    fireEvent.click(screen.getByText("Toggle controls"));
    fireEvent.click(screen.getByText("Toggle controls"));
    await waitFor(() => expect(controller.getSnapshot().read).toBe("ready"));
    expect(controller.getSnapshot().draft.reason).toBe("  retained\nreason  ");
    await changeSession({
      ...identity(),
      memberships: [{ incident_id: incidentId, role: "viewer" }],
    });
    expect(controller.canPropose("close")).toBe(false);
    expect(
      screen.getByRole("textbox", { name: "Reason" }).hasAttribute("readonly"),
    ).toBe(true);
    expect(controller.getSnapshot().draft.reason).toBe("  retained\nreason  ");
    await changeSession({
      ...identity(),
      memberships: [{ incident_id: incidentId, role: "admin" }],
    });
    await waitFor(() => expect(controller.getSnapshot().read).toBe("ready"));
    expect(controller.canConfirm()).toBe(false);
    expect(controller.canPropose("close")).toBe(true);
  });
  it("keeps a hidden acknowledgement without hidden reads and clears it on account replacement", async () => {
    const { controller, fetch, write, changeSession, identity, accepted } =
      await setup();
    close(controller);
    await waitFor(() =>
      expect(fetch.mock.calls.some(([, init]) => init?.method === "POST")).toBe(
        true,
      ),
    );
    fireEvent.click(screen.getByText("Toggle controls"));
    const reads = fetch.mock.calls.length;
    await act(async () =>
      write.resolve(
        json(
          envelope(
            incident({
              status: "closed",
              closed_at: "2026-08-02T00:00:00Z",
              incident_version: 2,
            }),
          ),
        ),
      ),
    );
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(fetch).toHaveBeenCalledTimes(reads);
    expect(accepted).toHaveBeenCalledTimes(2);
    await changeSession({
      ...identity(),
      user_id: "00000000-0000-4000-8000-000000000009",
    });
    expect(controller.getSnapshot().operation.kind).toBe("idle");
    expect(controller.getSnapshot().draft.reason).toBe("");
  });
  it("rejects late old-session publication even when transport ignores abort", async () => {
    const { controller, fetch, write, session, accepted } = await setup();
    close(controller);
    await waitFor(() =>
      expect(fetch.mock.calls.some(([, init]) => init?.method === "POST")).toBe(
        true,
      ),
    );
    act(() => session.sessionLost());
    const count = accepted.mock.calls.length;
    await act(async () =>
      write.resolve(
        json(
          envelope(
            incident({
              status: "closed",
              closed_at: "2026-08-02T00:00:00Z",
              incident_version: 2,
            }),
          ),
        ),
      ),
    );
    expect(accepted).toHaveBeenCalledTimes(count);
    expect(controller.getSnapshot().draft.reason).toBe("");
    expect(controller.getSnapshot().operation.kind).toBe("idle");
  });
  it("fans out newer lifecycle state while retaining metadata bases and rejecting stale or foreign resources", async () => {
    const authority = {
      incidentId,
      actorId,
      lifetime: "test",
      role: "admin" as const,
    };
    let current = true;
    const broadcast = vi.fn();
    const metadata = new IncidentMetadataController({
      isCurrent: () => current,
      lost: () => {},
      recover: async () => ({
        kind: "authorized",
        userId: actorId,
        role: "admin",
      }),
      read: async () => ({ ok: true, resource: incident() }),
      patch: async () => ({
        ok: true,
        resource: incident({ incident_version: 2 }),
      }),
      publishResource: broadcast,
    });
    metadata.setAuthority(authority);
    metadata.setActive(true);
    await act(async () => {
      for (let i = 0; i < 20; ++i) await Promise.resolve();
    });
    metadata.change("severity", "My retained severity");
    const owner = new IncidentResourceController({
      isCurrent: () => current,
      accepted: (r) => metadata.acceptResource(r, false),
    });
    owner.accept(incident(), authority);
    const base = metadata.getSnapshot().draft?.base;
    const closed = incident({
      status: "closed",
      closed_at: "2026-08-02T00:00:00Z",
      incident_version: 3,
    });
    owner.accept(closed, authority);
    expect(metadata.getSnapshot().draft?.values.severity).toBe(
      "My retained severity",
    );
    expect(metadata.getSnapshot().draft?.base).toBe(base);
    expect(metadata.getSnapshot().draft?.reviewRequired).toBe(true);
    owner.accept(incident({ incident_version: 2 }), authority);
    expect(owner.getSnapshot()).toEqual(closed);
    expect(metadata.getSnapshot().resource).toEqual(closed);
    expect(owner.accept(incident({ incident_version: 3 }), authority)).toBe(
      false,
    );
    current = false;
    expect(owner.accept(incident({ incident_version: 4 }), authority)).toBe(
      false,
    );
    expect(owner.accept(incident({ incident_id: "another" }), authority)).toBe(
      false,
    );
    metadata.dispose();
  });
});
