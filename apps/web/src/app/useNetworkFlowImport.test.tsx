import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { sessionResource } from "../testing/appShellTestSupport";
import { useAppWorkflowsForTest } from "../testing/appWorkflowsTestSupport";
import { importTestIds as ids } from "../testing/workbookImportTestSupport";
import type { NetworkFlowImportPort } from "../workbook/features/NetworkFlowOperations";
import { AppSessionController } from "./appSessionController";

const useNetworkFlowImport = (
  options: Parameters<typeof useAppWorkflowsForTest>[0],
) => useAppWorkflowsForTest(options).networkFlowImport;

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
  let importMajor = 1;
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
              contract_major: 7,
              route_families: ["/api/v1/incidents/{incident_id}/network-flow"],
              workspace_keys: ["network_analysis"],
              capabilities: [],
            },
            {
              profile_id: "import",
              claimed: withdrawn !== "import",
              claimable: true,
              contract_major: importMajor,
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
    preview: unavailable,
    previewMapping: unavailable,
  } satisfies NetworkFlowImportPort;
  let attachment: ReturnType<
    ReturnType<typeof useNetworkFlowImport>["attachWorkbook"]
  > | null = null;
  const bind = (
    readiness:
      | "available"
      | "pending"
      | "invalid"
      | "unavailable" = "available",
  ) =>
    (attachment = hook.result.current.attachWorkbook({
      incidentId,
      readiness,
      role: "editor",
      closed: false,
      client,
      accessFailure: vi.fn(),
    }));
  return {
    detach: () => attachment?.detach(),
    unsupported: () => {
      importMajor = 99;
    },
    supported: () => {
      importMajor = 1;
    },
    application,
    hook,
    bind,
    send,
    setIncident: (id: string) => {
      act(() => {
        hook.result.current.retire();
        incidentId = id;
      });
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
    act(() => h.detach());
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
  it("distinguishes pending readiness from definitive loss and restores review without replay", async () => {
    const h = setup();
    await h.login();
    act(() => {
      void h.hook.result.current.controller.upload(
        new File(["source"], "source.csv"),
      );
    });
    const request = h.hook.result.current.controller.getSnapshot().write;
    for (let i = 0; i < 2; i++) {
      act(() => h.bind("pending"));
      expect(h.hook.result.current.controller.getSnapshot().access).toBe(
        "paused",
      );
      expect(h.hook.result.current.controller.getSnapshot().write).toBeNull();
      act(() => h.bind());
      expect(
        h.hook.result.current.controller.getSnapshot().write?.disposition,
      ).toBe("uncertain");
      expect(h.send).toHaveBeenCalledTimes(1);
    }
    expect(h.hook.result.current.controller.getSnapshot().write?.request).toBe(
      request?.request,
    );
    expect(request).not.toBeNull();
    act(() => h.bind("invalid"));
    act(() => h.bind());
    expect(h.hook.result.current.controller.getSnapshot().write).toBeNull();
    expect(h.send).toHaveBeenCalledTimes(1);
  });
  it("clears retained work when an unsupported profile arrives while detached", async () => {
    const h = setup();
    await h.login();
    act(() => {
      void h.hook.result.current.controller.upload(
        new File(["source"], "source.csv"),
      );
    });
    let attachment: ReturnType<typeof h.bind> = null;
    act(() => {
      attachment = h.bind();
    });
    act(() => attachment?.detach());
    h.unsupported();
    await act(async () => {
      h.application.refreshResources();
      await flush();
    });
    h.supported();
    await act(async () => {
      h.application.refreshResources();
      await flush();
      h.bind();
    });
    expect(h.hook.result.current.controller.getSnapshot().write).toBeNull();
    expect(h.send).toHaveBeenCalledTimes(1);
  });
});
