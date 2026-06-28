import { describe, expect, it, vi } from "vitest";

import {
  buildAppRouteLocation,
  emptyAppRouteState,
  parseAppRouteState,
  writeAppRouteState,
} from "./routeState";

describe("app route state", () => {
  it("parses the root route without accidental manual-directory or debug state", () => {
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
      manualIncidentDirectory: false,
    });
  });

  it("parses deployment administration as a path-owned route", () => {
    expect(
      parseAppRouteState({
        historyState: { cartularyIncidentDirectory: true },
        pathname: "/deployment-administration",
        search: "?incident_id=incident-1&debug=harness",
      }),
    ).toEqual({
      incidentId: "",
      debugHarness: false,
      deploymentAdministration: true,
      manualIncidentDirectory: false,
    });
  });

  it("uses history state only for explicit incident-directory navigation", () => {
    expect(
      parseAppRouteState({
        historyState: { cartularyIncidentDirectory: true },
        pathname: "/",
        search: "",
      }),
    ).toMatchObject({
      manualIncidentDirectory: true,
    });
  });

  it("builds URLs while preserving non-route query parameters", () => {
    expect(
      buildAppRouteLocation(
        {
          incidentId: "incident-2",
          debugHarness: false,
          deploymentAdministration: false,
          manualIncidentDirectory: false,
        },
        "?keep=1&incident_id=incident-1&surface=timeline&debug=harness",
      ),
    ).toEqual({
      historyState: {},
      url: "/?keep=1&incident_id=incident-2&surface=timeline",
    });

    expect(
      buildAppRouteLocation(
        {
          incidentId: "",
          debugHarness: false,
          deploymentAdministration: true,
          manualIncidentDirectory: false,
        },
        "?keep=1&incident_id=incident-1&surface=timeline&debug=harness",
      ),
    ).toEqual({
      historyState: {},
      url: "/deployment-administration?keep=1",
    });
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
        manualIncidentDirectory: true,
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
        manualIncidentDirectory: false,
      },
      "replace",
      {
        currentSearch: "",
        history,
      },
    );

    expect(pushState).toHaveBeenCalledWith(
      { cartularyIncidentDirectory: true },
      "",
      "/",
    );
    expect(replaceState).toHaveBeenCalledWith({}, "", "/?debug=harness");
  });
});
