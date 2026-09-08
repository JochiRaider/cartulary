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
  metadataActorId as actorId,
  metadataIncidentId as incidentId,
  metadataDeferred,
  metadataEnvelope,
  metadataIncident,
  metadataJSON,
} from "../testing/incidentMetadataTestSupport";
import { useIncidentControlsDrawer } from "../workbook/hooks/useIncidentControlsDrawer";
import { AppSessionController } from "./appSessionController";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { IncidentMetadataFeature } from "./IncidentMetadataPanel";
import type { IncidentMetadataController } from "./incidentMetadataController";
import { useIncidentMetadata } from "./useIncidentMetadata";

const sessions: AppSessionController[] = [];
afterEach(() => {
  cleanup();
  for (const s of sessions.splice(0)) s.dispose();
  vi.unstubAllGlobals();
});
async function setup() {
  let currentSession = sessionResource({
    user_id: actorId,
    is_deployment_admin: false,
    memberships: [{ incident_id: incidentId, role: "admin" }],
  });
  let controller: IncidentMetadataController | null = null;
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
  const publish = vi.fn();
  const writes: ReturnType<typeof metadataDeferred<Response>>[] = [];
  let resource = metadataIncident();
  let preferenceReads = 0;
  const fetch = vi.fn((url: unknown, init?: RequestInit) => {
    if (init?.method === "PATCH") {
      const write = metadataDeferred<Response>();
      writes.push(write);
      return write.promise;
    }
    if (String(url).includes("workbook-preferences")) {
      ++preferenceReads;
      return Promise.resolve(
        metadataJSON({
          data: { default_sheet_ref: null, home_sheet_ref: null },
        }),
      );
    }
    return Promise.resolve(metadataJSON(metadataEnvelope(resource)));
  });
  vi.stubGlobal("fetch", fetch);
  function Harness() {
    const [role, setRole] = useState<WorkbookIncidentRole>("admin");
    const editor = useIncidentMetadata({
      sessionController: session,
      recovery,
      currentIncidentId: () => incidentId,
      onIncidentAccessLost: lost,
      onSessionLost: () => session.sessionLost(),
    });
    controller = editor.controller;
    const drawer = useIncidentControlsDrawer(false, (section) => {
      if (section !== "incident-fields") editor.controller.setActive(false);
    });
    return (
      <>
        <button
          type="button"
          onClick={() =>
            drawer.accountIncidentControls.onSelectSection("incident-fields")
          }
        >
          Open metadata
        </button>
        <button
          type="button"
          onClick={() => drawer.closeDrawer({ restoreTriggerFocus: false })}
        >
          Close metadata
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
        {drawer.drawerSection === "incident-fields" ? (
          <IncidentMetadataFeature
            controller={editor.controller}
            bindSurface={editor.bindSurface}
            incidentId={incidentId}
            currentIncidentRole={role}
            activeSection="incident-fields"
            onAuthorizationRecovered={(r) => setRole(r.role)}
            onIncidentResourceAccepted={publish}
          />
        ) : null}
        {drawer.drawerSection === "summary" ? (
          <IncidentAdminPanel
            incidentId={incidentId}
            currentIncidentRole={role}
            activeSection="summary"
            acceptedIncident={editor.resource}
          />
        ) : null}
      </>
    );
  }
  render(
    <StrictMode>
      <Harness />
    </StrictMode>,
  );
  return {
    session,
    lost,
    publish,
    fetch,
    writes,
    preferences: () => preferenceReads,
    controller: () => {
      if (!controller) throw new Error("missing editor");
      return controller;
    },
    open: async () => {
      fireEvent.click(screen.getByText("Open metadata"));
      await screen.findByLabelText("Severity");
    },
    update: async (overrides: Partial<typeof currentSession>) => {
      currentSession = { ...currentSession, ...overrides };
      await act(async () => {
        await session.refreshSession();
      });
    },
    acknowledge: async () => {
      resource = metadataIncident({
        severity: "critical",
        incident_version: 2,
      });
      await act(async () =>
        writes[0]?.resolve(metadataJSON(metadataEnvelope(resource))),
      );
    },
  };
}
describe("Incident metadata lifetime integration", () => {
  it("retains exact drafts through real drawer transitions and publishes a late save into summary without preference reloads", async () => {
    const s = await setup();
    await s.open();
    fireEvent.change(screen.getByLabelText("Severity"), {
      target: { value: "critical" },
    });
    fireEvent.click(screen.getByText("Close metadata"));
    await s.open();
    expect(screen.getByLabelText("Severity")).toHaveProperty(
      "value",
      "critical",
    );
    fireEvent.click(screen.getByText("Save promoted fields"));
    await waitFor(() => expect(s.writes).toHaveLength(1));
    fireEvent.click(screen.getByText("Summary"));
    await screen.findByText("Incident controls synced.");
    const count = s.preferences();
    await s.acknowledge();
    expect(s.preferences()).toBe(count);
    expect(screen.getByText("critical")).toBeTruthy();
    expect(s.publish).toHaveBeenLastCalledWith(
      expect.objectContaining({ incident_version: 2 }),
    );
    await s.open();
    expect(screen.getByText("Saved promoted incident fields.")).toBeTruthy();
  });
  it("requires renewed review after current role changes while retaining authorized values and drafts", async () => {
    const s = await setup();
    await s.open();
    fireEvent.change(screen.getByLabelText("Severity"), {
      target: { value: "exact draft  " },
    });
    act(() => screen.getByLabelText("Severity").focus());
    await s.update({
      memberships: [{ incident_id: incidentId, role: "viewer" }],
    });
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Refresh" }),
    );
    expect(screen.queryByLabelText("Severity")).toBeNull();
    expect(screen.getByText("Original description")).toBeTruthy();
    expect(s.lost).not.toHaveBeenCalled();
    await s.update({
      memberships: [{ incident_id: incidentId, role: "reviewer" }],
    });
    expect(screen.getByLabelText("Severity")).toHaveProperty(
      "value",
      "exact draft  ",
    );
    expect(
      screen.getByText("Save promoted fields").getAttribute("aria-disabled"),
    ).toBe("true");
    fireEvent.click(screen.getByText("Use this version"));
    expect(s.controller().canSave()).toBe(true);
  });
  it("clears incident access without logging out and excludes a late acknowledgement after account replacement", async () => {
    const s = await setup();
    await s.open();
    fireEvent.change(screen.getByLabelText("Severity"), {
      target: { value: "critical" },
    });
    fireEvent.click(screen.getByText("Save promoted fields"));
    await waitFor(() => expect(s.writes).toHaveLength(1));
    await s.update({ memberships: [], is_deployment_admin: true });
    expect(s.lost).toHaveBeenCalled();
    expect(s.controller().getSnapshot().draft).toBeNull();
    expect(s.session.getSnapshot().session).not.toBeNull();
    await s.update({
      user_id: "00000000-0000-4000-8000-000000000099",
      memberships: [{ incident_id: incidentId, role: "admin" }],
    });
    s.publish.mockClear();
    await s.acknowledge();
    expect(s.publish).not.toHaveBeenCalled();
    expect(s.controller().getSnapshot().operation.kind).toBe("idle");
  });
});
