export type AppRouteState = {
  incidentId: string;
  debugHarness: boolean;
  deploymentAdministration: boolean;
  manualIncidentDirectory: boolean;
};

export type AppRouteWriteMode = "push" | "replace";

export type AppRouteLocation = {
  historyState: Record<string, true>;
  url: string;
};

type BrowserHistoryWriter = Pick<History, "pushState" | "replaceState">;

const incidentDirectoryHistoryFlag = "cartularyIncidentDirectory";

export const emptyAppRouteState: AppRouteState = {
  incidentId: "",
  debugHarness: false,
  deploymentAdministration: false,
  manualIncidentDirectory: false,
};

export function parseAppRouteState(options: {
  historyState?: unknown;
  pathname: string;
  search: string;
}): AppRouteState {
  const params = new URLSearchParams(options.search);
  const deploymentAdministration =
    options.pathname === "/deployment-administration";
  const incidentId = deploymentAdministration
    ? ""
    : (params.get("incident_id") ?? "").trim();
  const historyState =
    typeof options.historyState === "object" && options.historyState !== null
      ? (options.historyState as Record<string, unknown>)
      : null;

  return {
    incidentId,
    debugHarness:
      !deploymentAdministration && params.get("debug") === "harness",
    deploymentAdministration,
    manualIncidentDirectory:
      !deploymentAdministration &&
      incidentId === "" &&
      historyState?.[incidentDirectoryHistoryFlag] === true,
  };
}

export function readAppRouteState(): AppRouteState {
  return parseAppRouteState({
    historyState: window.history.state,
    pathname: window.location.pathname,
    search: window.location.search,
  });
}

export function buildAppRouteLocation(
  next: AppRouteState,
  currentSearch: string,
): AppRouteLocation {
  const params = new URLSearchParams(currentSearch);

  if (next.deploymentAdministration) {
    params.delete("incident_id");
    params.delete("surface");
    params.delete("debug");
    const query = params.toString();
    return {
      historyState: {},
      url:
        query === ""
          ? "/deployment-administration"
          : `/deployment-administration?${query}`,
    };
  }

  if (next.incidentId === "") {
    params.delete("incident_id");
    params.delete("surface");
  } else {
    params.set("incident_id", next.incidentId);
  }

  if (next.debugHarness) {
    params.set("debug", "harness");
  } else {
    params.delete("debug");
  }

  const query = params.toString();
  return {
    historyState: next.manualIncidentDirectory
      ? { [incidentDirectoryHistoryFlag]: true }
      : {},
    url: query === "" ? "/" : `/?${query}`,
  };
}

export function writeAppRouteState(
  next: AppRouteState,
  mode: AppRouteWriteMode,
  options: {
    currentSearch?: string;
    history?: BrowserHistoryWriter;
  } = {},
) {
  const target = buildAppRouteLocation(
    next,
    options.currentSearch ?? window.location.search,
  );
  const history = options.history ?? window.history;
  if (mode === "push") {
    history.pushState(target.historyState, "", target.url);
    return;
  }
  history.replaceState(target.historyState, "", target.url);
}
