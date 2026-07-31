import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ExtensionAvailabilityTag } from "../../extensions/extensionAvailability";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  useWorkbookStartupAdmission,
  type WorkbookStartupAvailabilityPort,
  type WorkbookStartupSavedViewStatePort,
  type WorkbookStartupSelectionPort,
} from "./useWorkbookStartupAdmission";

type Deferred<T> = {
  readonly promise: Promise<T>;
  readonly resolve: (value: T) => void;
};

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

function response(data: unknown, status = 200): Response {
  return new Response(JSON.stringify({ data }), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function baseStartup(viewSchemaId = hostsViewSchemaId) {
  return {
    incident_id: "incident-1",
    extension_workspace_availability: { workspaces: [] },
    selected_sheet_ref: { kind: "view_schema", id: viewSchemaId },
    selected_view_schema_id: viewSchemaId,
    selected_saved_view: null,
    source: "default",
    cleared_pointers: [],
    home_sheet_ref: null,
    default_sheet_ref: null,
  };
}

function savedViewStartup() {
  return {
    ...baseStartup(hostsViewSchemaId),
    selected_sheet_ref: { kind: "saved_view", id: "saved-view-1" },
    selected_saved_view: {
      saved_view_id: "saved-view-1",
      view_schema_id: hostsViewSchemaId,
      display_name: "Hosts by owner",
      scope: "private",
      query_json: { sort: [], filters: [] },
      layout_json: {},
      owner_user_id: "user-1",
      saved_view_version: 3,
    },
  };
}

function extensionStartup() {
  return {
    ...baseStartup(),
    selected_sheet_ref: {
      kind: "extension_workspace",
      extension_profile_id: "network_flow_activity",
      workspace_key: "network_analysis",
    },
    selected_view_schema_id: null,
    selected_saved_view: null,
  };
}

function ports(
  options: {
    readonly accept?: boolean;
    readonly renderable?: boolean;
    readonly reserve?: ExtensionAvailabilityTag | null;
  } = {},
) {
  let selectionVersion = 0;
  const events: string[] = [];
  const availabilityPort: WorkbookStartupAvailabilityPort = {
    reserve: vi.fn(
      () =>
        options.reserve ?? {
          epochId: "a".repeat(64),
          generation: 1n,
        },
    ),
    acceptWorkbookStartup: vi.fn(() => options.accept ?? true),
    isRenderable: vi.fn(() => options.renderable ?? true),
  };
  const selectionPort: WorkbookStartupSelectionPort = {
    readSelectionVersion: vi.fn(() => selectionVersion),
    applyStartupIdentity: vi.fn(() => {
      events.push("identity");
    }),
    selectTimeline: vi.fn(() => {
      events.push("fallback");
    }),
  };
  const savedViewStatePort: WorkbookStartupSavedViewStatePort = {
    upsertSavedView: vi.fn(() => {
      events.push("saved-view");
    }),
    applyQueryStateForSurface: vi.fn(() => {
      events.push("query");
    }),
    applyLayoutStateForSurface: vi.fn(() => {
      events.push("layout");
    }),
  };
  return {
    availabilityPort,
    events,
    savedViewStatePort,
    selectionPort,
    setSelectionVersion: (value: number) => {
      selectionVersion = value;
    },
  };
}

function renderAdmission(
  ownedPorts: ReturnType<typeof ports>,
  onAvailabilityChange = vi.fn(),
  incidentId = "incident-1",
) {
  return renderHook(
    ({ currentIncidentId }: { readonly currentIncidentId: string }) =>
      useWorkbookStartupAdmission({
        incidentId: currentIncidentId,
        urlParams: new URLSearchParams(
          "view_schema_id=cartulary.view.hosts.v1",
        ),
        availabilityPort: ownedPorts.availabilityPort,
        selectionPort: ownedPorts.selectionPort,
        savedViewStatePort: ownedPorts.savedViewStatePort,
        onAvailabilityChange,
      }),
    { initialProps: { currentIncidentId: incidentId } },
  );
}

describe("useWorkbookStartupAdmission", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("applies an ordinary base-surface startup identity", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.resolve(response(baseStartup()))),
    );
    const ownedPorts = ports();
    const onAvailabilityChange = vi.fn();
    renderAdmission(ownedPorts, onAvailabilityChange);

    await waitFor(() => {
      expect(
        ownedPorts.selectionPort.applyStartupIdentity,
      ).toHaveBeenCalledWith({
        sheetRef: { kind: "view_schema", id: hostsViewSchemaId },
        viewSchemaId: hostsViewSchemaId,
      });
    });
    expect(onAvailabilityChange).toHaveBeenCalledOnce();
    expect(ownedPorts.selectionPort.selectTimeline).not.toHaveBeenCalled();
  });

  it("hydrates a saved view in upsert-query-layout-identity order", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.resolve(response(savedViewStartup()))),
    );
    const ownedPorts = ports();
    renderAdmission(ownedPorts);

    await waitFor(() => {
      expect(ownedPorts.selectionPort.applyStartupIdentity).toHaveBeenCalled();
    });
    expect(ownedPorts.events).toEqual([
      "saved-view",
      "query",
      "layout",
      "identity",
    ]);
  });

  it("keeps a user selection made while startup is delayed", async () => {
    const pending = deferred<Response>();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => pending.promise),
    );
    const ownedPorts = ports();
    const onAvailabilityChange = vi.fn();
    renderAdmission(ownedPorts, onAvailabilityChange);
    await waitFor(() => expect(fetch).toHaveBeenCalledOnce());
    ownedPorts.setSelectionVersion(1);

    await act(async () => {
      pending.resolve(response(baseStartup()));
      await pending.promise;
    });

    await waitFor(() => expect(onAvailabilityChange).toHaveBeenCalledOnce());
    expect(
      ownedPorts.selectionPort.applyStartupIdentity,
    ).not.toHaveBeenCalled();
    expect(
      ownedPorts.savedViewStatePort.upsertSavedView,
    ).not.toHaveBeenCalled();
  });

  it("admits only the newest of two overlapping startup requests", async () => {
    const first = deferred<Response>();
    const second = deferred<Response>();
    const fetchMock = vi
      .fn()
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    vi.stubGlobal("fetch", fetchMock);
    const ownedPorts = ports();
    const rendered = renderAdmission(ownedPorts);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledOnce());
    rendered.rerender({ currentIncidentId: "incident-2" });
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));

    await act(async () => {
      second.resolve(response(baseStartup(timelineViewSchemaId)));
      await second.promise;
    });
    await waitFor(() => {
      expect(
        ownedPorts.selectionPort.applyStartupIdentity,
      ).toHaveBeenCalledOnce();
    });

    await act(async () => {
      first.resolve(response(baseStartup(hostsViewSchemaId)));
      await first.promise;
    });
    expect(
      ownedPorts.selectionPort.applyStartupIdentity,
    ).toHaveBeenCalledOnce();
    expect(ownedPorts.selectionPort.applyStartupIdentity).toHaveBeenCalledWith({
      sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
      viewSchemaId: timelineViewSchemaId,
    });
  });

  it("falls back exactly once when availability admission is stale", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.resolve(response(baseStartup()))),
    );
    const ownedPorts = ports({ accept: false });
    const onAvailabilityChange = vi.fn();
    renderAdmission(ownedPorts, onAvailabilityChange);

    await waitFor(() => {
      expect(ownedPorts.selectionPort.selectTimeline).toHaveBeenCalledWith(
        "availability_rejected",
      );
    });
    expect(ownedPorts.selectionPort.selectTimeline).toHaveBeenCalledOnce();
    expect(onAvailabilityChange).toHaveBeenCalledOnce();
  });

  it("falls back exactly once when the selected extension is not renderable", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.resolve(response(extensionStartup()))),
    );
    const ownedPorts = ports({ renderable: false });
    renderAdmission(ownedPorts);

    await waitFor(() => {
      expect(ownedPorts.selectionPort.selectTimeline).toHaveBeenCalledWith(
        "selected_extension_not_renderable",
      );
    });
    expect(ownedPorts.selectionPort.selectTimeline).toHaveBeenCalledOnce();
  });

  it("applies no response state for a non-OK result", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() => Promise.resolve(response({ error: {} }, 503))),
    );
    const ownedPorts = ports();
    renderAdmission(ownedPorts);

    await waitFor(() => expect(fetch).toHaveBeenCalledOnce());
    await act(async () => {
      await Promise.resolve();
    });
    expect(
      ownedPorts.availabilityPort.acceptWorkbookStartup,
    ).not.toHaveBeenCalled();
    expect(
      ownedPorts.selectionPort.applyStartupIdentity,
    ).not.toHaveBeenCalled();
  });

  it("applies no selection-owned state for a malformed result", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve(
          response({
            extension_workspace_availability: { workspaces: [] },
            selected_sheet_ref: "invalid",
          }),
        ),
      ),
    );
    const ownedPorts = ports();
    const onAvailabilityChange = vi.fn();
    renderAdmission(ownedPorts, onAvailabilityChange);

    await waitFor(() => expect(onAvailabilityChange).toHaveBeenCalledOnce());
    expect(
      ownedPorts.selectionPort.applyStartupIdentity,
    ).not.toHaveBeenCalled();
    expect(
      ownedPorts.savedViewStatePort.upsertSavedView,
    ).not.toHaveBeenCalled();
  });

  it("applies no callback or response state after effect teardown", async () => {
    const pending = deferred<Response>();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => pending.promise),
    );
    const ownedPorts = ports();
    const onAvailabilityChange = vi.fn();
    const rendered = renderAdmission(ownedPorts, onAvailabilityChange);
    await waitFor(() => expect(fetch).toHaveBeenCalledOnce());
    rendered.unmount();

    await act(async () => {
      pending.resolve(response(baseStartup()));
      await pending.promise;
    });
    expect(onAvailabilityChange).not.toHaveBeenCalled();
    expect(
      ownedPorts.availabilityPort.acceptWorkbookStartup,
    ).not.toHaveBeenCalled();
    expect(
      ownedPorts.selectionPort.applyStartupIdentity,
    ).not.toHaveBeenCalled();
  });

  it("does not hydrate a late saved view after a later selection", async () => {
    const pending = deferred<Response>();
    vi.stubGlobal(
      "fetch",
      vi.fn(() => pending.promise),
    );
    const ownedPorts = ports();
    const onAvailabilityChange = vi.fn();
    renderAdmission(ownedPorts, onAvailabilityChange);
    await waitFor(() => expect(fetch).toHaveBeenCalledOnce());
    ownedPorts.setSelectionVersion(2);

    await act(async () => {
      pending.resolve(response(savedViewStartup()));
      await pending.promise;
    });

    await waitFor(() => expect(onAvailabilityChange).toHaveBeenCalledOnce());
    expect(
      ownedPorts.savedViewStatePort.upsertSavedView,
    ).not.toHaveBeenCalled();
    expect(
      ownedPorts.savedViewStatePort.applyQueryStateForSurface,
    ).not.toHaveBeenCalled();
    expect(
      ownedPorts.savedViewStatePort.applyLayoutStateForSurface,
    ).not.toHaveBeenCalled();
    expect(
      ownedPorts.selectionPort.applyStartupIdentity,
    ).not.toHaveBeenCalled();
  });
});
