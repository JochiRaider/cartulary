import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { StrictMode, useLayoutEffect, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { sessionResource } from "../testing/appShellTestSupport";
import { metadataJSON as json } from "../testing/incidentMetadataTestSupport";
import {
  preferenceActorId as actorId,
  defaultPreference,
  homePreference,
  preferenceIncidentId as incidentId,
  preferenceDeferred,
  preferenceTimeline,
} from "../testing/workbookPreferenceTestSupport";
import type { WorkbookPreferenceController } from "../workbook/preferences/WorkbookPreferenceController";
import { WorkbookPreferencesPanel } from "../workbook/preferences/WorkbookPreferencesPanel";
import { AppSessionController } from "./appSessionController";
import { useWorkbookPreferences } from "./useWorkbookPreferences";

const envelope = <T,>(data: T) => ({
  data,
  meta: { request_id: "preference" },
});
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
  let controller: WorkbookPreferenceController | undefined;
  let bindWorkbook:
    | ReturnType<typeof useWorkbookPreferences>["bindWorkbook"]
    | undefined;
  const session = new AppSessionController({
    session: async () => ({
      ok: true,
      status: 200,
      payload: envelope(identity),
    }),
    preferences: async () => ({ ok: false, status: 503, payload: {} }),
    extensions: async () => ({
      ok: true,
      status: 200,
      payload: envelope({ extensions: [] }),
    }),
    retireLifetime: () => controller?.retire(),
  });
  sessions.push(session);
  await session.refreshSession();
  const recovery = session.recoveryPort((id) => id === incidentId);
  const write = preferenceDeferred<Response>();
  const fetch = vi.fn((url: string, init?: RequestInit) =>
    init?.method === "PUT"
      ? write.promise
      : Promise.resolve(
          json(
            envelope(
              url.endsWith("/me") ? homePreference() : defaultPreference(),
            ),
          ),
        ),
  );
  vi.stubGlobal("fetch", fetch);
  const lost = vi.fn();
  const recovered = vi.fn();
  function Harness() {
    const [open, setOpen] = useState(true);
    const workflow = useWorkbookPreferences({
      sessionController: session,
      recovery,
      currentIncidentId: () => incidentId,
      onIncidentAccessLost: lost,
      onSessionLost: () => session.sessionLost(),
    });
    controller = workflow.controller;
    bindWorkbook = workflow.bindWorkbook;
    useLayoutEffect(() => {
      workflow.bindWorkbook({
        incidentId,
        actorId,
        apiBase: undefined,
        surface: {
          sheetRef: preferenceTimeline,
          label: "Timeline",
          available: true,
        },
        onAuthorizationRecovered: recovered,
      });
      // The workbook binds before the independently mounted drawer opens.
      workflow.controller.setInspectionActive(open);
    });
    return (
      <>
        <button type="button" onClick={() => setOpen(!open)}>
          Toggle controls
        </button>
        {open ? (
          <WorkbookPreferencesPanel controller={workflow.controller} />
        ) : null}
      </>
    );
  }
  render(
    <StrictMode>
      <Harness />
    </StrictMode>,
  );
  if (!controller) throw new Error("Missing preference binding");
  const bound = controller;
  await waitFor(() => expect(bound.getSnapshot().home.read).toBe("ready"));
  return {
    controller: bound,
    bindWorkbook: (
      binding: Parameters<
        ReturnType<typeof useWorkbookPreferences>["bindWorkbook"]
      >[0],
    ) => {
      if (!bindWorkbook) throw new Error("Missing workbook binding");
      bindWorkbook(binding);
    },
    session,
    write,
    fetch,
    recovered,
    lost,
    identity: () => identity,
    changeSession: async (next: typeof identity) => {
      identity = next;
      await act(async () => {
        await session.refreshSession();
      });
    },
  };
}
describe("Preference App integration", () => {
  it("retains pending work across drawer closure and confirms the captured target on return", async () => {
    const f = await setup();
    fireEvent.click(
      screen.getByRole("button", { name: "Set current surface as my home" }),
    );
    await waitFor(() =>
      expect(
        f.fetch.mock.calls.filter(([, init]) => init?.method === "PUT"),
      ).toHaveLength(1),
    );
    fireEvent.click(screen.getByText("Toggle controls"));
    expect(f.controller.hasDepartureWork()).toBe(true);
    await act(async () =>
      f.write.resolve(json(envelope(homePreference(preferenceTimeline)))),
    );
    fireEvent.click(screen.getByText("Toggle controls"));
    await screen.findByText("Home update confirmed.");
    expect(f.controller.getSnapshot().surface?.sheetRef).toEqual(
      preferenceTimeline,
    );
    expect(f.recovered).toHaveBeenCalledWith(
      expect.objectContaining({ role: "admin", userId: actorId }),
    );
  });
  it("retains authorized observations on demotion and clears protected state on membership loss", async () => {
    const f = await setup();
    await f.changeSession({
      ...f.identity(),
      memberships: [{ incident_id: incidentId, role: "viewer" }],
    });
    expect(f.controller.getSnapshot().home.resource).not.toBeNull();
    expect(f.controller.canWrite("home")).toBe(true);
    expect(f.controller.canWrite("default")).toBe(false);
    await f.changeSession({ ...f.identity(), memberships: [] });
    expect(f.controller.getSnapshot().authority).toBeNull();
    expect(f.controller.getSnapshot().home.resource).toBeNull();
  });
  it("fences late acknowledgements after replacement of the same actors session", async () => {
    const f = await setup();
    act(() => f.controller.clear("home"));
    await waitFor(() =>
      expect(f.controller.getSnapshot().home.transportPending).toBe(true),
    );
    await f.changeSession({
      ...f.identity(),
      authenticated_at: "2026-09-08T16:00:00Z",
    });
    await act(async () => f.write.resolve(json(envelope(homePreference()))));
    expect(f.controller.getSnapshot().home.operation.kind).toBe("idle");
    expect(f.controller.hasDepartureWork()).toBe(false);
  });
  it("ignores workbook publications addressed to another incident or actor", async () => {
    const f = await setup();
    const original = f.controller.getSnapshot().surface;
    for (const address of [
      { incidentId: "another-incident", actorId },
      { incidentId, actorId: "another-actor" },
    ]) {
      act(() =>
        f.bindWorkbook({
          ...address,
          apiBase: undefined,
          surface: {
            sheetRef: { kind: "view_schema", id: "cartulary.view.hosts.v1" },
            label: "Wrong lifetime",
            available: true,
          },
          onAuthorizationRecovered: () => {},
        }),
      );
      expect(f.controller.getSnapshot().surface).toBe(original);
    }
  });
});
