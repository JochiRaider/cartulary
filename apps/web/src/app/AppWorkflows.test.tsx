import { act, cleanup, renderHook } from "@testing-library/react";
import { type ReactNode, StrictMode, Suspense, startTransition } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useWorkflowAttachment } from "../shared/useWorkflowAttachment";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import { sessionResource } from "../testing/appShellTestSupport";
import {
  metadataActorId,
  metadataIncident,
  metadataIncidentId,
} from "../testing/incidentMetadataTestSupport";
import { type AppWorkflowOptions, createAppWorkflows } from "./AppWorkflows";
import { AppSessionController } from "./appSessionController";
import { IncidentResourceController } from "./incidentResourceController";
import { useAppSession } from "./useAppSession";
import { useAppWorkflows } from "./useAppWorkflows";

const sessions: AppSessionController[] = [];
afterEach(() => {
  cleanup();
  for (const session of sessions.splice(0)) session.dispose();
  vi.unstubAllGlobals();
});
function setup() {
  let workflows: ReturnType<typeof createAppWorkflows> | undefined;
  const session = new AppSessionController({
    session: async () => ({
      ok: true,
      status: 200,
      payload: {
        meta: { request_id: "workflows" },
        data: sessionResource({
          user_id: metadataActorId,
          memberships: [{ incident_id: metadataIncidentId, role: "admin" }],
        }),
      },
    }),
    preferences: async () => ({ ok: false, status: 503, payload: {} }),
    extensions: async () => ({
      ok: true,
      status: 200,
      payload: { data: { extensions: [] }, meta: { request_id: "extensions" } },
    }),
    retireLifetime: () => workflows?.retire(),
  });
  sessions.push(session);
  const resources = new IncidentResourceController({
    isCurrent: (authority) =>
      session.getSnapshot().lifetime === authority.lifetime &&
      session.getSnapshot().session?.user_id === authority.actorId,
    accepted: (resource) => {
      workflows?.metadata.controller.acceptResource(resource, false);
      workflows?.lifecycle.controller.acceptResource(resource, false);
    },
  });
  const options: AppWorkflowOptions = {
    sessionController: session,
    recovery: session.recoveryPort(() => true),
    currentIncidentId: () => metadataIncidentId,
    onSessionLost: () => session.sessionLost(),
    onIncidentAccessLost: resources.retire,
    onResourceAccepted: resources.accept,
  };
  return {
    session,
    resources,
    options,
    assign: (value: ReturnType<typeof createAppWorkflows>) => {
      workflows = value;
    },
  };
}
const strict = ({ children }: { children: ReactNode }) => (
  <StrictMode>{children}</StrictMode>
);
describe("Application workflow composition", () => {
  it("constructs eight inert owners and retires and disposes each contribution once", async () => {
    const h = setup();
    const subscribe = vi.spyOn(h.session, "subscribe");
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const workflows = createAppWorkflows(() => h.options);
    h.assign(workflows);
    expect(subscribe).not.toHaveBeenCalled();
    expect(fetch).not.toHaveBeenCalled();
    const members = [
      workflows.metadata,
      workflows.lifecycle,
      workflows.membershipManagement,
      workflows.membershipAudit,
      workflows.savedViews,
      workflows.preferences,
      workflows.workbookImport,
      workflows.networkFlowImport,
    ];
    const sync = members.map((m) => vi.spyOn(m, "synchronize"));
    const retire = members.map((m) => vi.spyOn(m, "retire"));
    const dispose = members.map((m) => vi.spyOn(m, "dispose"));
    workflows.synchronize();
    sync.forEach((spy) => {
      expect(spy).toHaveBeenCalledTimes(1);
    });
    workflows.retire();
    retire.forEach((spy) => {
      expect(spy).toHaveBeenCalledTimes(1);
    });
    workflows.dispose();
    workflows.dispose();
    workflows.synchronize();
    dispose.forEach((spy) => {
      expect(spy).toHaveBeenCalledTimes(1);
    });
    expect(fetch).not.toHaveBeenCalled();
  });
  it("retires all old authority before publishing and suppresses a reentrant superseded session", async () => {
    const h = setup();
    const workflows = createAppWorkflows(() => h.options);
    h.assign(workflows);
    const release = workflows.connect(h.session);
    await h.session.refreshSession();
    let reentered = false;
    const snapshots: string[] = [];
    const unsubscribe = workflows.metadata.controller.subscribe(() => {
      if (
        !workflows.metadata.controller.getSnapshot().authority &&
        !reentered
      ) {
        reentered = true;
        h.session.sessionLost();
        // Retirement cannot admit an old presentation through a synchronous callback.
        workflows.lifecycle.attachSurface({
          incidentId: metadataIncidentId,
          activeSection: "summary",
          currentIncidentRole: "admin",
        });
      }
    });
    const published = h.session.subscribe(() => {
      const snapshot = h.session.getSnapshot();
      snapshots.push(snapshot.session?.user_id ?? "anonymous");
      if (snapshot.ended) {
        expect(
          workflows.savedViews.controller.getSnapshot().authority,
        ).toBeNull();
        expect(
          workflows.lifecycle.controller.getSnapshot().authority,
        ).toBeNull();
      }
    });
    h.session.authenticationCompleted(
      sessionResource({ user_id: "replacement", memberships: [] }),
      h.session.getSnapshot().revision,
    );
    expect(reentered).toBe(true);
    expect(snapshots).not.toContain("replacement");
    expect(h.session.getSnapshot().ended).toBe(true);
    unsubscribe();
    published();
    release();
    workflows.dispose();
  });
  it("replaces the committed session subscription and ignores callbacks from the old source", async () => {
    const a = setup(),
      b = setup();
    let oldListener: (() => void) | undefined;
    const original = a.session.subscribe;
    const unsubscribed = vi.fn();
    vi.spyOn(a.session, "subscribe").mockImplementation((listener) => {
      oldListener = listener;
      const stop = original(listener);
      return () => {
        unsubscribed();
        return stop();
      };
    });
    const hook = renderHook(({ options }) => useAppWorkflows(options), {
      initialProps: { options: a.options },
    });
    const identity = hook.result.current;
    await act(() => a.session.refreshSession());
    hook.rerender({ options: b.options });
    await act(() => b.session.refreshSession());
    expect(hook.result.current).toBe(identity);
    expect(unsubscribed).toHaveBeenCalledTimes(1);
    const before = identity.metadata.controller.getSnapshot();
    act(() => oldListener?.());
    expect(identity.metadata.controller.getSnapshot()).toBe(before);
    hook.unmount();
    identity.dispose();
  });
  it("keeps owners live through Strict Mode replay and disposes once at terminal departure", async () => {
    const h = setup();
    let dispose: ReturnType<typeof vi.spyOn> | undefined;
    const hook = renderHook(
      () => {
        const workflows = useAppWorkflows(h.options);
        h.assign(workflows);
        if (!dispose) dispose = vi.spyOn(workflows, "dispose");
        useAppSession(h.session, workflows.dispose);
        return workflows;
      },
      { wrapper: strict },
    );
    await act(async () => {
      await Promise.resolve();
    });
    expect(dispose).not.toHaveBeenCalled();
    const workflows = hook.result.current;
    workflows.metadata.controller.acceptResource(metadataIncident(), false);
    hook.unmount();
    await act(async () => {
      await Promise.resolve();
    });
    expect(dispose).toHaveBeenCalledTimes(1);
    workflows.synchronize();
    expect(workflows.metadata.controller.getSnapshot().authority).toBeNull();
  });
  it("does not subscribe dispatch or mutate a committed owner from an abandoned render", async () => {
    const h = setup();
    const subscription = vi.spyOn(h.session, "subscribe");
    const pending = new Promise<never>(() => {});
    const wrapper = ({ children }: { children: ReactNode }) => (
      <Suspense fallback={null}>{children}</Suspense>
    );
    const hook = renderHook(
      ({ abandon }) => {
        const w = useAppWorkflows(h.options);
        if (abandon) throw pending;
        return w;
      },
      { initialProps: { abandon: false }, wrapper },
    );
    await act(() => h.session.refreshSession());
    const owner = hook.result.current;
    const before = owner.metadata.controller.getSnapshot();
    const calls = subscription.mock.calls.length;
    hook.rerender({ abandon: true });
    expect(subscription).toHaveBeenCalledTimes(calls);
    expect(owner.metadata.controller.getSnapshot()).toBe(before);
    hook.unmount();
    owner.dispose();
  });
  it("rejects stale attachments mismatched subjects and bindings after retirement or disposal", async () => {
    const h = setup();
    const workflows = createAppWorkflows(() => h.options);
    h.assign(workflows);
    const release = workflows.connect(h.session);
    await h.session.refreshSession();
    const active = vi
      .spyOn(workflows.metadata.controller, "setActive")
      .mockImplementation(() => {});
    const binding: WorkbookIncidentControlsRendererProps = {
      incidentId: metadataIncidentId,
      activeSection: "summary",
      currentIncidentRole: "admin",
    };
    const old = workflows.metadata.attachSurface(binding);
    if (!old) throw new Error("Expected an authorized attachment");
    const current = workflows.metadata.attachSurface(binding);
    if (!current) throw new Error("Expected replacement attachment");
    active.mockClear();
    old.detach();
    expect(old.update({ ...binding, activeSection: "incident-fields" })).toBe(
      false,
    );
    expect(current.update({ ...binding, incidentId: "foreign" })).toBe(false);
    workflows.metadata
      .attachSurface({ ...binding, incidentId: "foreign" })
      ?.detach();
    expect(active).not.toHaveBeenCalled();
    expect(
      current.update({ ...binding, activeSection: "incident-fields" }),
    ).toBe(true);
    expect(active).toHaveBeenLastCalledWith(true);
    current.detach();
    current.detach();
    expect(active).toHaveBeenCalledTimes(2);
    expect(active).toHaveBeenLastCalledWith(false);
    const preferences = vi.spyOn(
      workflows.preferences.controller,
      "setSurface",
    );
    expect(
      workflows.preferences.attachWorkbook({
        incidentId: metadataIncidentId,
        actorId: "foreign",
        apiBase: undefined,
        surface: null,
        onAuthorizationRecovered: vi.fn(),
      }),
    ).toBeNull();
    expect(preferences).not.toHaveBeenCalled();
    const retired = workflows.metadata.attachSurface(binding);
    if (!retired) throw new Error("Expected retained attachment");
    workflows.retire();
    active.mockClear();
    expect(retired.update(binding)).toBe(false);
    retired.detach();
    workflows.dispose();
    workflows.metadata.attachSurface(binding)?.detach();
    expect(active).not.toHaveBeenCalled();
    release();
  });
  it("commits presentation updates and visibility without admitting abandoned render ports", async () => {
    const h = setup();
    const workflows = createAppWorkflows(() => h.options);
    h.assign(workflows);
    workflows.connect(h.session);
    await h.session.refreshSession();
    const active = vi
      .spyOn(workflows.metadata.controller, "setActive")
      .mockImplementation(() => {});
    const pending = new Promise<never>(() => {});
    const wrapper = ({ children }: { children: ReactNode }) => (
      <Suspense fallback={null}>{children}</Suspense>
    );
    const hook = renderHook(
      ({
        abandon,
        section,
      }: {
        abandon: boolean;
        section: "summary" | "incident-fields";
      }) => {
        useWorkflowAttachment(
          workflows.metadata.attachSurface,
          {
            incidentId: metadataIncidentId,
            activeSection: section,
            currentIncidentRole: "admin",
          },
          true,
        );
        if (abandon) throw pending;
      },
      { wrapper, initialProps: { abandon: false, section: "summary" } },
    );
    active.mockClear();
    act(() =>
      startTransition(() =>
        hook.rerender({ abandon: true, section: "incident-fields" }),
      ),
    );
    expect(active).not.toHaveBeenCalled();
    act(() => document.dispatchEvent(new Event("visibilitychange")));
    expect(active).toHaveBeenLastCalledWith(false);
    hook.rerender({ abandon: false, section: "incident-fields" });
    expect(active).toHaveBeenLastCalledWith(true);
    hook.unmount();
    expect(active).toHaveBeenLastCalledWith(false);
    workflows.dispose();
  });
});
