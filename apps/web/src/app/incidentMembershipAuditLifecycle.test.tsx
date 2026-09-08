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
import {
  auditEnvelope,
  auditEvent,
} from "../testing/administrativeAuditTestSupport";
import { sessionResource } from "../testing/appShellTestSupport";
import { useIncidentControlsDrawer } from "../workbook/hooks/useIncidentControlsDrawer";
import { AppSessionController } from "./appSessionController";
import { IncidentMembershipAuditFeature } from "./IncidentMembershipAuditPanel";
import type { IncidentMembershipAuditController } from "./incidentMembershipAuditController";
import { useIncidentMembershipAudit } from "./useIncidentMembershipAudit";

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
  let controller: IncidentMembershipAuditController | null = null;
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
    const audit = useIncidentMembershipAudit({
      sessionController: session,
      recovery,
      currentIncidentId: () => incidentId,
      onIncidentAccessLost: lost,
      onSessionLost,
    });
    controller = audit.controller;
    const drawer = useIncidentControlsDrawer(false, (section) => {
      if (section !== "membership-audit") audit.controller.setActive(false);
    });
    return (
      <>
        <button
          type="button"
          onClick={() =>
            drawer.accountIncidentControls.onSelectSection("membership-audit")
          }
        >
          Open audit
        </button>
        <button
          type="button"
          onClick={() => drawer.closeDrawer({ restoreTriggerFocus: false })}
        >
          Close audit
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
        {drawer.drawerSection === "membership-audit" ? (
          <IncidentMembershipAuditFeature
            controller={audit.controller}
            bindSurface={audit.bindSurface}
            incidentId={incidentId}
            currentIncidentRole={role}
            activeSection="membership-audit"
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
    if (!controller) throw new Error("Audit controller missing");
    return controller;
  };
  const finish = async (index: number) => {
    await act(async () => {
      const read = reads[index];
      if (!read) throw new Error("Audit read missing");
      read(
        json(
          auditEnvelope([
            auditEvent({
              scope_kind: "incident",
              scope_id: incidentId,
              action_code: "membership_created",
              target_kind: "incident_membership",
            }),
          ]),
        ),
      );
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
describe("Incident membership audit lifetime integration", () => {
  it("retains only filters across actual drawer and section transitions under Strict Mode", async () => {
    const s = await setup();
    expect(s.reads).toHaveLength(0);
    fireEvent.click(screen.getByRole("button", { name: "Open audit" }));
    await waitFor(() => expect(s.reads).toHaveLength(1));
    await s.finish(0);
    fireEvent.change(screen.getByRole("combobox", { name: "Action code" }), {
      target: { value: "membership_deleted" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Apply filters" }));
    await waitFor(() => expect(s.reads).toHaveLength(2));
    fireEvent.click(screen.getByRole("button", { name: "Close audit" }));
    expect(s.getController().getSnapshot().page).toBeNull();
    await s.finish(1);
    expect(s.getController().getSnapshot().page).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Open audit" }));
    await waitFor(() => expect(s.reads).toHaveLength(3));
    expect(
      screen.getByRole("combobox", { name: "Action code" }),
    ).toHaveProperty("value", "membership_deleted");
    fireEvent.click(screen.getByRole("button", { name: "Summary" }));
    await s.finish(2);
    expect(s.getController().getSnapshot().page).toBeNull();
  });
  it("synchronously clears accepted session role loss while preserving visible incident navigation", async () => {
    const s = await setup();
    fireEvent.click(screen.getByRole("button", { name: "Open audit" }));
    await waitFor(() => expect(s.reads).toHaveLength(1));
    await s.finish(0);
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await waitFor(() => expect(s.reads).toHaveLength(2));
    await s.update({
      memberships: [{ incident_id: incidentId, role: "viewer" }],
    });
    expect(s.getController().getSnapshot()).toMatchObject({
      page: null,
      authority: null,
    });
    expect(screen.getByText("Current role: viewer")).toBeTruthy();
    expect(s.lost).not.toHaveBeenCalled();
    await s.finish(1);
    expect(s.getController().getSnapshot().page).toBeNull();
    await s.update({
      memberships: [{ incident_id: incidentId, role: "admin" }],
    });
    await waitFor(() => expect(s.reads).toHaveLength(3));
  });
  it("does not grant incident access through deployment administration and delegates hidden incident loss", async () => {
    const s = await setup();
    fireEvent.click(screen.getByRole("button", { name: "Open audit" }));
    await waitFor(() => expect(s.reads).toHaveLength(1));
    await s.update({ is_deployment_admin: true, memberships: [] });
    expect(s.getController().getSnapshot().authority).toBeNull();
    expect(s.lost).toHaveBeenCalledTimes(1);
    await s.finish(0);
    expect(s.getController().getSnapshot().page).toBeNull();
    expect(s.reads).toHaveLength(1);
  });
  it("clears protected filters and late publication when the application retires the session", async () => {
    const s = await setup();
    fireEvent.click(screen.getByRole("button", { name: "Open audit" }));
    await waitFor(() => expect(s.reads).toHaveLength(1));
    fireEvent.change(screen.getByRole("textbox", { name: "Target ID" }), {
      target: { value: "private" },
    });
    act(() => s.session.sessionLost());
    await s.finish(0);
    expect(s.getController().getSnapshot()).toMatchObject({
      page: null,
      authority: null,
      inputs: { target_id: "" },
    });
  });
});
