import { incidentAdministrationTestId } from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { useLayoutEffect, useRef, useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { observeAccountOperation } from "../../app/accountOperation";
import { IncidentAdminPanel } from "../../app/IncidentAdminPanel";
import type { SheetRef } from "../../shared/sheetRef";
import {
  metadataActorId as actorId,
  metadataIncidentId as incidentId,
  metadataDeferred,
  metadataIncident,
  metadataJSON,
} from "../../testing/incidentMetadataTestSupport";
import { workbookAuthorizationRecovery } from "../../testing/workbookAuthorizationTestSupport";
import { useSavedViewTestApplication } from "../../testing/workbookSavedViewTestSupport";
import { createWorkbookPreferenceAdapter } from "../adapters/createWorkbookPreferenceAdapter";
import { ActiveSurfaceSavedViewSelector } from "../components/ActiveSurfaceSavedViewSelector";
import { useWorkbookStartupController } from "../hooks/useWorkbookStartupController";
import { WorkbookPreferenceController } from "./WorkbookPreferenceController";
import {
  formatPreferencePointer,
  WorkbookPreferencesPanel,
} from "./WorkbookPreferencesPanel";

const timeline: SheetRef = {
  kind: "view_schema",
  id: "cartulary.view.timeline.v2",
};
const hosts = { kind: "view_schema", id: "cartulary.view.hosts.v1" } as const;
const now = "2026-08-01T00:00:00Z";
const envelope = (data: unknown) => ({
  data,
  meta: { request_id: "preferences" },
});
const resource = (home: boolean, pointer: SheetRef | null) => ({
  incident_id: incidentId,
  created_at: now,
  updated_at: now,
  ...(home
    ? { user_id: actorId, home_sheet_ref: pointer }
    : { updated_by_user_id: actorId, default_sheet_ref: pointer }),
});

// Regression coverage retained from the pre-refactor characterization.
function Surface({ refresh = 0 }: { refresh?: number }) {
  const [preferences] = useState(
    () =>
      new WorkbookPreferenceController({
        port: () =>
          createWorkbookPreferenceAdapter({
            apiBase: undefined,
            incidentId,
            actorId,
          }),
        isCurrent: () => true,
        lost: () => {},
        observe: observeAccountOperation,
        recover: async () => ({
          kind: "authorized",
          userId: actorId,
          role: "admin",
        }),
      }),
  );
  const selection = useRef(0);
  const startup = useWorkbookStartupController({
    incidentId,
    surfaceSelectionVersionRef: selection,
  });
  useLayoutEffect(() => {
    preferences.setAuthority({
      incidentId,
      actorId,
      lifetime: "session",
      role: "admin",
    });
    preferences.setSurface({
      sheetRef: startup.snapshot.startupSheetRef,
      label: null,
      available: true,
    });
    preferences.setInspectionActive(true);
  });
  useLayoutEffect(() => () => preferences.dispose(), [preferences]);
  const savedViews = useSavedViewTestApplication(
    incidentId,
    actorId,
    workbookAuthorizationRecovery(),
  );
  return (
    <>
      <button
        type="button"
        onClick={() => startup.commands.selectWorkbookSurface(hosts.id)}
      >
        Select Hosts
      </button>
      <ActiveSurfaceSavedViewSelector
        controller={savedViews.savedViewController}
        activeViewSchemaId={startup.snapshot.surface}
        chromeMode="base"
        currentIncidentRole="admin"
        currentUserId={actorId}
        selectedSheetRef={startup.snapshot.startupSheetRef}
        savedViewsResource={{
          kind: "ready",
          selectedSavedView: null,
          selectedSavedViewId: "",
          message: null,
        }}
        onSelectBaseSurface={startup.commands.selectWorkbookSurface}
        preferenceController={preferences}
      />
      <IncidentAdminPanel
        key={refresh}
        incidentId={incidentId}
        currentIncidentRole="admin"
        preferenceControls={
          <WorkbookPreferencesPanel controller={preferences} />
        }
      />
    </>
  );
}

describe("Workbook preference characterization", () => {
  let home: SheetRef | null;
  let fetchMock: ReturnType<
    typeof vi.fn<(url: string, init?: RequestInit) => Promise<Response>>
  >;
  beforeEach(() => {
    home = null;
    window.history.replaceState({}, "", `/?incident_id=${incidentId}`);
    fetchMock = vi.fn(async (url: string, init?: RequestInit) => {
      if (url.endsWith(`/incidents/${incidentId}`))
        return metadataJSON(envelope(metadataIncident()));
      const personal = url.endsWith("/me");
      if (init?.method === "PUT")
        home = JSON.parse(String(init.body)).home_sheet_ref;
      return metadataJSON(envelope(resource(personal, personal ? home : null)));
    });
    vi.stubGlobal("fetch", fetchMock);
  });
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });
  const open = () =>
    fireEvent.click(screen.getByRole("button", { name: "Saved view actions" }));
  const setHome = () => {
    if (!screen.queryByRole("button", { name: "Set as my home" }))
      fireEvent.click(screen.getByRole("button", { name: /^Startup/ }));
    fireEvent.click(screen.getByRole("button", { name: "Set as my home" }));
  };
  const writes = () =>
    fetchMock.mock.calls.filter(([, init]) => init?.method === "PUT");

  it("provides explicit clear actions for both inspected resources", async () => {
    render(<Surface />);
    await screen.findByText("Incident controls synced.");
    expect
      .soft(screen.queryByRole("button", { name: "Clear my home" }))
      .not.toBeNull();
    expect
      .soft(screen.queryByRole("button", { name: "Clear incident default" }))
      .not.toBeNull();
  });
  it("publishes summary and home while the default read is delayed", async () => {
    const delayed = metadataDeferred<Response>();
    const original = fetchMock.getMockImplementation();
    if (!original) throw new Error("Missing preference fetch fixture");
    fetchMock.mockImplementation((url, init) =>
      url.endsWith("/default")
        ? delayed.promise.then((response) => response.clone())
        : original(url, init),
    );
    render(<Surface />);
    await waitFor(() =>
      expect(
        screen.getByTestId(incidentAdministrationTestId("summary-key"))
          .textContent,
      ).toBe("IR-METADATA"),
    );
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref"))
        .textContent,
    ).toBe("Unset");
    await act(async () =>
      delayed.resolve(metadataJSON(envelope(resource(false, null)))),
    );
  });
  it("rejects wrong incident and wrong user preference observations", async () => {
    const original = fetchMock.getMockImplementation();
    if (!original) throw new Error("Missing preference fetch fixture");
    fetchMock.mockImplementation((url, init) =>
      url.includes("workbook-preferences")
        ? Promise.resolve(
            metadataJSON(
              envelope({
                ...resource(url.endsWith("/me"), hosts),
                ...(url.endsWith("/me")
                  ? { user_id: "00000000-0000-4000-8000-000000000099" }
                  : { incident_id: "00000000-0000-4000-8000-000000009999" }),
              }),
            ),
          )
        : original(url, init),
    );
    render(<Surface />);
    await screen.findByText("IR-METADATA");
    expect
      .soft(
        screen.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref"))
          .textContent,
      )
      .toBe("Unavailable");
    expect
      .soft(
        screen.getByTestId(
          incidentAdministrationTestId("pref-default-sheet-ref"),
        ).textContent,
      )
      .toBe("Unavailable");
  });
  it("keeps admission when a surface switch exposes another shortcut", async () => {
    const delayed = metadataDeferred<Response>();
    const original = fetchMock.getMockImplementation();
    if (!original) throw new Error("Missing preference fetch fixture");
    fetchMock.mockImplementation((url, init) =>
      init?.method === "PUT"
        ? delayed.promise.then((response) => response.clone())
        : original(url, init),
    );
    render(<Surface />);
    open();
    setHome();
    await waitFor(() => expect(writes()).toHaveLength(1));
    fireEvent.click(screen.getByRole("button", { name: "Select Hosts" }));
    open();
    setHome();
    expect(writes()).toHaveLength(1);
    await act(async () =>
      delayed.resolve(metadataJSON(envelope(resource(true, timeline)))),
    );
  });
  it("retains the captured target and acknowledgement after a surface switch", async () => {
    const delayed = metadataDeferred<Response>();
    const original = fetchMock.getMockImplementation();
    if (!original) throw new Error("Missing preference fetch fixture");
    fetchMock.mockImplementation((url, init) =>
      init?.method === "PUT"
        ? delayed.promise.then((response) => response.clone())
        : original(url, init),
    );
    render(<Surface />);
    open();
    setHome();
    await waitFor(() => expect(writes()).toHaveLength(1));
    fireEvent.click(screen.getByRole("button", { name: "Select Hosts" }));
    expect(JSON.parse(String(writes()[0]?.[1]?.body))).toEqual({
      home_sheet_ref: timeline,
    });
    await act(async () =>
      delayed.resolve(metadataJSON(envelope(resource(true, timeline)))),
    );
    expect(
      screen.queryAllByText(/Home.*(confirmed|updated)/)[0],
    ).not.toBeNull();
  });
  it("keeps a confirmed write when a subsequent preference read fails", async () => {
    const view = render(<Surface />);
    open();
    setHome();
    await screen.findAllByText("Home update confirmed.");
    const original = fetchMock.getMockImplementation();
    if (!original) throw new Error("Missing preference fetch fixture");
    fetchMock.mockImplementation((url, init) =>
      url.includes("workbook-preferences")
        ? Promise.resolve(
            metadataJSON({ error: { code: "internal_error" } }, 500),
          )
        : original(url, init),
    );
    view.rerender(<Surface refresh={1} />);
    await screen.findAllByText(
      "Displayed stored value is retained and may be stale. Refresh this resource.",
    );
    expect(screen.queryAllByText("Home update confirmed.")[0]).not.toBeNull();
    expect(writes()).toHaveLength(1);
  });
  it("updates the inspected value after the existing Set shortcut", async () => {
    render(<Surface />);
    await screen.findByText("Incident controls synced.");
    open();
    setHome();
    await screen.findAllByText("Home update confirmed.");
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref"))
        .textContent,
    ).toContain("cartulary.view.timeline.v2");
  });
  it("does not display loading or missing preference data as Unset", async () => {
    const delayed = metadataDeferred<Response>();
    const original = fetchMock.getMockImplementation();
    if (!original) throw new Error("Missing preference fetch fixture");
    fetchMock.mockImplementation((url, init) =>
      url.includes("workbook-preferences")
        ? delayed.promise.then((response) => response.clone())
        : original(url, init),
    );
    render(<Surface />);
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref"))
        .textContent,
    ).toBe("Loading…");
    await act(async () =>
      delayed.resolve(metadataJSON(envelope({ incident_id: incidentId }))),
    );
    expect
      .soft(
        screen.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref"))
          .textContent,
      )
      .toBe("Unavailable");
  });
  it("labels stored saved pointers from already authorized metadata without target discovery", () => {
    const id = "00000000-0000-4000-8000-000000000099";
    const pointer: SheetRef = { kind: "saved_view", id };
    expect(
      formatPreferencePointer(pointer, {
        sheetRef: timeline,
        label: "Timeline",
        available: true,
        savedViewLabels: [{ id, label: "Known investigation" }],
      }),
    ).toBe(`Saved view: Known investigation (${id})`);
    expect(
      formatPreferencePointer(pointer, {
        sheetRef: hosts,
        label: "Hosts",
        available: true,
      }),
    ).toBe(`Saved view: ${id}`);
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
