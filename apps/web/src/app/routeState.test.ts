import { describe, expect, it, vi } from "vitest";

import {
  buildAppRouteLocation,
  emptyAppRouteState,
  parseAppRouteState,
  readAppRouteState,
  writeAppRouteState,
} from "./routeState";

describe("app route state", () => {
  it("parses the root route without accidental debug state", () => {
    expect(
      parseAppRouteState({
        pathname: "/",
        search: "",
      }),
    ).toEqual(emptyAppRouteState);
  });

  it("parses incident and debug query state outside deployment administration", () => {
    expect(
      parseAppRouteState({
        pathname: "/",
        search: "?incident_id=%20incident-1%20&debug=harness",
      }),
    ).toEqual({
      incidentId: "incident-1",
      debugHarness: true,
      deploymentAdministration: false,
    });
  });

  it("parses deployment administration as a path-owned route", () => {
    expect(
      parseAppRouteState({
        pathname: "/deployment-administration",
        search: "?incident_id=incident-1&debug=harness",
      }),
    ).toEqual({
      incidentId: "",
      debugHarness: false,
      deploymentAdministration: true,
    });
  });

  it("reads root navigation independently of obsolete history flags", () => {
    window.history.replaceState({ cartularyIncidentDirectory: true }, "", "/");
    expect(readAppRouteState()).toEqual(emptyAppRouteState);
  });

  it("builds URLs while preserving non-route query parameters", () => {
    expect(
      buildAppRouteLocation(
        {
          incidentId: "incident-2",
          debugHarness: false,
          deploymentAdministration: false,
        },
        "?keep=1&incident_id=incident-1&surface=timeline&debug=harness",
      ),
    ).toEqual({
      historyState: {},
      url: "/?keep=1&incident_id=incident-2",
    });

    expect(
      buildAppRouteLocation(
        {
          incidentId: "",
          debugHarness: false,
          deploymentAdministration: true,
        },
        "?keep=1&incident_id=incident-1&surface=timeline&debug=harness",
      ),
    ).toEqual({
      historyState: {},
      url: "/deployment-administration?keep=1",
    });
  });

  it("clears incident-owned workbook startup parameters when changing application context", () => {
    const search =
      "?keep=1&incident_id=incident-1&view_schema_id=cartulary.view.hosts.v1&sheet_ref_kind=saved_view&sheet_ref_id=saved-1&extension_profile_id=network_flow_activity&workspace_key=network_analysis&surface=hosts";
    for (const route of [
      emptyAppRouteState,
      { ...emptyAppRouteState, incidentId: "incident-2" },
      { ...emptyAppRouteState, deploymentAdministration: true },
    ]) {
      const url = new URL(
        buildAppRouteLocation(route, search).url,
        "http://cartulary.test",
      );
      expect(Object.fromEntries(url.searchParams)).toEqual(
        route.incidentId === "incident-2"
          ? { keep: "1", incident_id: "incident-2" }
          : { keep: "1" },
      );
    }
  });

  it("writes push and replace history entries with the route-owned history state", () => {
    const pushState = vi.fn();
    const replaceState = vi.fn();
    const history = { pushState, replaceState };

    writeAppRouteState(
      {
        incidentId: "",
        debugHarness: false,
        deploymentAdministration: false,
      },
      "push",
      {
        currentSearch: "?incident_id=incident-1&surface=timeline",
        history,
      },
    );
    writeAppRouteState(
      {
        incidentId: "",
        debugHarness: true,
        deploymentAdministration: false,
      },
      "replace",
      {
        currentSearch: "",
        history,
      },
    );

    expect(pushState).toHaveBeenCalledWith({}, "", "/");
    expect(replaceState).toHaveBeenCalledWith({}, "", "/?debug=harness");
  });
});
