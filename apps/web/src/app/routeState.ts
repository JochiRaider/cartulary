export type AppRouteState = {
  incidentId: string;
  debugHarness: boolean;
  deploymentAdministration: boolean;
};

export type AppRouteWriteMode = "push" | "replace";

export type AppRouteLocation = {
  historyState: Record<string, true>;
  url: string;
};

type BrowserHistoryWriter = Pick<History, "pushState" | "replaceState">;

export const emptyAppRouteState: AppRouteState = {
  incidentId: "",
  debugHarness: false,
  deploymentAdministration: false,
};

export function parseAppRouteState(options: {
  pathname: string;
  search: string;
}): AppRouteState {
  const params = new URLSearchParams(options.search);
  const deploymentAdministration =
    options.pathname === "/deployment-administration";
  const incidentId = deploymentAdministration
    ? ""
    : (params.get("incident_id") ?? "").trim();

  return {
    incidentId,
    debugHarness:
      !deploymentAdministration && params.get("debug") === "harness",
    deploymentAdministration,
  };
}

export function readAppRouteState(): AppRouteState {
  return parseAppRouteState({
    pathname: window.location.pathname,
    search: window.location.search,
  });
}

export function buildAppRouteLocation(
  next: AppRouteState,
  currentSearch: string,
): AppRouteLocation {
  const params = new URLSearchParams(currentSearch);
  // Sheet selection belongs to one incident. Application navigation starts a
  // fresh workbook context; sign-in at an explicit URL does not rewrite it.
  if (
    next.deploymentAdministration ||
    next.incidentId === "" ||
    next.incidentId !== params.get("incident_id")
  ) {
    for (const key of [
      "view_schema_id",
      "sheet_ref_kind",
      "sheet_ref_id",
      "extension_profile_id",
      "workspace_key",
      "surface",
    ])
      params.delete(key);
  }

  if (next.deploymentAdministration) {
    params.delete("incident_id");
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
    historyState: {},
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
