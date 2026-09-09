import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { sessionResource } from "../testing/appShellTestSupport";
import { importTestIds as ids } from "../testing/workbookImportTestSupport";
import type { NetworkFlowImportPort } from "../workbook/features/NetworkFlowFeature";
import { AppSessionController } from "./appSessionController";
import { useNetworkFlowImport } from "./useNetworkFlowImport";

const controllers: AppSessionController[] = [];
afterEach(() => {
  cleanup();
  for (const c of controllers.splice(0)) c.dispose();
});
const flush = async () => {
  for (let i = 0; i < 30; i++) await Promise.resolve();
};
function setup() {
  let incidentId: string = ids.incident;
  let session = sessionResource({
    user_id: ids.actor,
    memberships: [{ incident_id: ids.incident, role: "editor" }],
  });
  let withdrawn = "";
  const application = new AppSessionController({
    session: async () => ({
      ok: true,
      status: 200,
      payload: { meta: { request_id: "test" }, data: session },
    }),
    preferences: async () => ({
      ok: false,
      status: 503,
      payload: { error: { code: "service_unavailable" } },
    }),
    extensions: async () => ({
      ok: true,
      status: 200,
      payload: {
        meta: { request_id: "test" },
        data: {
          extensions: [
            {
              profile_id: "network_flow_activity",
              claimed: withdrawn !== "network_flow_activity",
              claimable: true,
              contract_major: 5,
              route_families: ["/api/v1/incidents/{incident_id}/network-flow"],
              workspace_keys: ["network_analysis"],
              capabilities: [],
            },
            {
              profile_id: "import",
              claimed: withdrawn !== "import",
              claimable: true,
              contract_major: 1,
              route_families: ["/api/v1/import-sessions"],
              workspace_keys: [],
              capabilities: [],
            },
          ],
        },
      },
    }),
  });
  controllers.push(application);
  const hook = renderHook(() =>
    useNetworkFlowImport({
      sessionController: application,
      currentIncidentId: () => incidentId,
    }),
  );
  const send = vi.fn(() => new Promise<never>(() => {}));
  const unavailable = async () => ({
    kind: "failed" as const,
    failure: {
      kind: "transport" as const,
      status: 0,
      code: "interrupted",
      reason: null,
      field: null,
      retryable: true,
    },
  });
  const client = {
    send,
    readJob: unavailable,
    readSession: unavailable,
    listUnits: unavailable,
    readUnit: unavailable,
    preview: unavailable,
    previewMapping: unavailable,
  } satisfies NetworkFlowImportPort;
  const bind = () =>
    hook.result.current.bindWorkbook({
      incidentId,
      available: true,
      role: "editor",
      closed: false,
      client,
      accessFailure: vi.fn(),
    });
  return {
    application,
    hook,
    bind,
    send,
    setIncident: (id: string) => {
      incidentId = id;
    },
    setRole: () => {
      session = {
        ...session,
        memberships: [{ incident_id: ids.incident, role: "viewer" }],
      };
    },
    withdraw: (profileId: string) => {
      withdrawn = profileId;
    },
    login: async () => {
      await act(async () => {
        await application.refreshSession();
        await flush();
        bind();
      });
    },
  };
}

describe("Network Flow import application binding", () => {
  it("retains submitted intent on write loss and retires either withdrawn extension claim", async () => {
    for (const profile of ["import", "network_flow_activity"]) {
      const h = setup();
      await h.login();
      act(() => {
        void h.hook.result.current.controller.upload(
          new File(["source"], "source.csv"),
        );
      });
      expect(h.send).toHaveBeenCalledTimes(1);
      h.setRole();
      await act(async () => {
        await h.application.refreshSession();
      });
      expect(h.hook.result.current.controller.getSnapshot().canWrite).toBe(
        false,
      );
      expect(
        h.hook.result.current.controller.getSnapshot().write?.request.kind,
      ).toBe("upload");
      expect(
        h.hook.result.current.controller.getSnapshot().write?.disposition,
      ).toBe("uncertain");
      h.withdraw(profile);
      await act(async () => {
        h.application.start();
        await flush();
      });
      expect(h.hook.result.current.controller.getSnapshot().write).toBeNull();
    }
  });
  it("fences incident and account replacement before accepting a pending analytical write", async () => {
    const h = setup();
    await h.login();
    act(() => {
      void h.hook.result.current.controller.upload(
        new File(["source"], "source.csv"),
      );
    });
    expect(h.send).toHaveBeenCalledTimes(1);
    h.setIncident(ids.secondUnit);
    act(() => h.bind());
    expect(h.hook.result.current.controller.getSnapshot().write).toBeNull();
    await act(async () => {
      h.application.authenticationCompleted(
        sessionResource({
          user_id: ids.secondUnit,
          memberships: [{ incident_id: ids.secondUnit, role: "editor" }],
        }),
        h.application.getSnapshot().revision,
      );
      await flush();
      h.bind();
    });
    expect(h.hook.result.current.controller.getSnapshot().discovery).toBeNull();
    expect(h.send).toHaveBeenCalledTimes(1);
  });
  it("hides protected state while the shell is absent and restores only explicit recovery", async () => {
    const h = setup();
    await h.login();
    act(() => {
      void h.hook.result.current.controller.upload(
        new File(["source"], "source.csv"),
      );
    });
    act(() => h.hook.result.current.bindWorkbook(null));
    expect(h.hook.result.current.controller.getSnapshot()).toMatchObject({
      access: "paused",
      write: null,
      draft: null,
    });
    act(() => h.bind());
    expect(
      h.hook.result.current.controller.getSnapshot().write?.disposition,
    ).toBe("uncertain");
    expect(h.send).toHaveBeenCalledTimes(1);
  });
});
