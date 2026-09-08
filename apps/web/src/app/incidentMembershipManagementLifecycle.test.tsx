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
import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import { sessionResource } from "../testing/appShellTestSupport";
import {
  membershipEnvelope,
  membershipFixture,
} from "../testing/incidentMembershipManagementTestSupport";
import { useIncidentControlsDrawer } from "../workbook/hooks/useIncidentControlsDrawer";
import { AppSessionController } from "./appSessionController";
import { IncidentMembershipManagementFeature } from "./IncidentMembershipManagementPanel";
import type { IncidentMembershipManagementController } from "./incidentMembershipManagementController";
import { useIncidentMembershipManagement } from "./useIncidentMembershipManagement";

const incidentId = "00000000-0000-4000-8000-000000001001";
const actorId = "00000000-0000-4000-8000-000000000001";
const sessions: AppSessionController[] = [];
const json = (payload: unknown) =>
  new Response(JSON.stringify(payload), {
    headers: { "Content-Type": "application/json" },
  });
afterEach(() => {
  cleanup();
  for (const session of sessions.splice(0)) session.dispose();
  vi.unstubAllGlobals();
});
async function setup() {
  let currentSession = sessionResource({
    user_id: actorId,
    is_deployment_admin: false,
    memberships: [{ incident_id: incidentId, role: "admin" }],
  });
  let controller: IncidentMembershipManagementController | null = null;
  const session = new AppSessionController({
    session: async () => ({
      ok: true,
      status: 200,
      payload: { data: currentSession, meta: { request_id: "session" } },
    }),
    preferences: async () => ({
      ok: false,
      status: 503,
      payload: { error: { code: "service_unavailable" } },
    }),
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
  const lost = vi.fn();
  const onSessionLost = vi.fn(() => session.sessionLost());
  const reads: ((value: Response) => void)[] = [];
  vi.stubGlobal(
    "fetch",
    vi.fn(() => new Promise<Response>((resolve) => reads.push(resolve))),
  );
  function Harness() {
    const [role, setRole] = useState<WorkbookIncidentRole>("admin");
    const management = useIncidentMembershipManagement({
      sessionController: session,
      recovery,
      currentIncidentId: () => incidentId,
      onIncidentAccessLost: lost,
      onSessionLost,
    });
    controller = management.controller;
    const drawer = useIncidentControlsDrawer(false, (section) => {
      if (section !== "memberships") management.controller.setActive(false);
    });
    return (
      <>
        <button
          type="button"
          onClick={() =>
            drawer.accountIncidentControls.onSelectSection("memberships")
          }
        >
          Open management
        </button>
        <button
          type="button"
          onClick={() => drawer.closeDrawer({ restoreTriggerFocus: false })}
        >
          Close management
        </button>
        <button
          type="button"
          onClick={() =>
            drawer.accountIncidentControls.onSelectSection("summary")
          }
        >
          Summary
        </button>
        <span>Current role: {role}</span>
        {drawer.drawerSection === "memberships" ? (
          <IncidentMembershipManagementFeature
            controller={management.controller}
            bindSurface={management.bindSurface}
            incidentId={incidentId}
            currentIncidentRole={role}
            activeSection="memberships"
            onAuthorizationRecovered={(result) => setRole(result.role)}
            onSessionRoleChange={() => session.refreshSession().then(() => {})}
          />
        ) : null}
      </>
    );
  }
  const view = render(
    <StrictMode>
      <Harness />
    </StrictMode>,
  );
  const getController = () => {
    if (!controller) throw new Error("Memberships controller missing");
    return controller;
  };
  const finish = async (index: number) => {
    await act(async () => {
      const read = reads[index];
      if (!read) throw new Error("Memberships read missing");
      read(json(membershipEnvelope()));
    });
  };
  return {
    view,
    reads,
    lost,
    session,
    getController,
    finish,
    update: async (overrides: Partial<typeof currentSession>) => {
      currentSession = { ...currentSession, ...overrides };
      await act(async () => {
        await session.refreshSession();
      });
    },
  };
}
describe("Incident membership management lifetime integration", () => {
  it("retains one draft across actual drawer and section transitions under Strict Mode", async () => {
    const s = await setup();
    expect(s.reads).toHaveLength(0);
    fireEvent.click(screen.getByRole("button", { name: "Open management" }));
    await waitFor(() => expect(s.reads).toHaveLength(1));
    await s.finish(0);
    fireEvent.click(
      screen.getByRole("button", { name: /Change role for Analyst/u }),
    );
    fireEvent.change(
      screen.getByRole("combobox", { name: /Role for Analyst/u }),
      { target: { value: "reviewer" } },
    );
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await waitFor(() => expect(s.reads).toHaveLength(2));
    fireEvent.click(screen.getByRole("button", { name: "Close management" }));
    expect(s.getController().getSnapshot().page).toBeNull();
    await s.finish(1);
    expect(s.getController().getSnapshot().page).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Open management" }));
    await waitFor(() => expect(s.reads).toHaveLength(3));
    await s.finish(2);
    expect(
      screen.getByRole("combobox", { name: /Role for Analyst/u }),
    ).toHaveProperty("value", "reviewer");
    fireEvent.click(screen.getByRole("button", { name: "Summary" }));
    expect(s.getController().getSnapshot()).toMatchObject({
      page: null,
      editor: { role: "reviewer" },
    });
  });
  it("propagates current role loss and preserves ordinary reads while retiring privileged drafts", async () => {
    const s = await setup();
    fireEvent.click(screen.getByRole("button", { name: "Open management" }));
    await waitFor(() => expect(s.reads).toHaveLength(1));
    await s.finish(0);
    fireEvent.click(
      screen.getByRole("button", {
        name: /Remove incident access for Analyst/u,
      }),
    );
    await s.update({
      memberships: [{ incident_id: incidentId, role: "viewer" }],
    });
    expect(s.getController().getSnapshot()).toMatchObject({
      authority: { role: "viewer" },
      editor: null,
    });
    expect(screen.getByText("Current role: viewer")).toBeTruthy();
    expect(screen.getByText("Analyst")).toBeTruthy();
    expect(s.lost).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await waitFor(() => expect(s.reads).toHaveLength(2));
    await s.finish(1);
    expect(
      screen.queryByRole("button", { name: "Add existing account" }),
    ).toBeNull();
  });
  it("denies deployment-admin nonmember materialization and retires incident data without logging out", async () => {
    const s = await setup();
    fireEvent.click(screen.getByRole("button", { name: "Open management" }));
    await waitFor(() => expect(s.reads).toHaveLength(1));
    await s.update({ is_deployment_admin: true, memberships: [] });
    expect(s.getController().getSnapshot().authority).toBeNull();
    expect(s.lost).toHaveBeenCalledTimes(1);
    await s.finish(0);
    expect(s.getController().getSnapshot().page).toBeNull();
    expect(s.session.getSnapshot().session?.user_id).toBe(actorId);
  });
  it("fences hidden reads and clears all retained work on session retirement", async () => {
    const s = await setup();
    fireEvent.click(screen.getByRole("button", { name: "Open management" }));
    await waitFor(() => expect(s.reads).toHaveLength(1));
    await s.finish(0);
    fireEvent.click(
      screen.getByRole("button", { name: "Add existing account" }),
    );
    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "private@example.test" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await waitFor(() => expect(s.reads).toHaveLength(2));
    vi.spyOn(document, "visibilityState", "get").mockReturnValue("hidden");
    fireEvent(document, new Event("visibilitychange"));
    await s.finish(1);
    expect(s.getController().getSnapshot()).toMatchObject({
      page: null,
      editor: { email: "private@example.test" },
    });
    act(() => s.session.sessionLost());
    expect(s.getController().getSnapshot()).toMatchObject({
      page: null,
      authority: null,
      editor: null,
      operation: { kind: "idle" },
    });
    vi.restoreAllMocks();
  });
  it("settles an admitted self-demotion while closed and propagates its recovered role", async () => {
    const s = await setup();
    fireEvent.click(screen.getByRole("button", { name: "Open management" }));
    await waitFor(() => expect(s.reads).toHaveLength(1));
    await act(async () =>
      s.reads[0]?.(
        json(
          membershipEnvelope([
            membershipFixture({ user_id: actorId, role: "admin" }),
          ]),
        ),
      ),
    );
    fireEvent.click(
      screen.getByRole("button", { name: /Change role for Analyst/u }),
    );
    fireEvent.change(screen.getByRole("combobox"), {
      target: { value: "viewer" },
    });
    expect(screen.getByText(/lose membership administration/u)).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Save role" }));
    await waitFor(() => expect(s.reads).toHaveLength(2));
    fireEvent.click(screen.getByRole("button", { name: "Close management" }));
    await s.update({
      memberships: [{ incident_id: incidentId, role: "viewer" }],
    });
    await act(async () =>
      s.reads[1]?.(
        json({
          data: membershipFixture({
            user_id: actorId,
            role: "viewer",
            membership_version: 2,
          }),
          meta: { request_id: "demoted" },
        }),
      ),
    );
    expect(s.getController().getSnapshot()).toMatchObject({
      authority: { role: "viewer" },
      editor: null,
      operation: { kind: "confirmed" },
      page: null,
    });
    expect(screen.getByText("Current role: viewer")).toBeTruthy();
    expect(s.lost).not.toHaveBeenCalled();
  });
});
