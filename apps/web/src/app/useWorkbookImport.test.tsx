import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { WorkbookImportPort } from "../imports/WorkbookImportController";
import { sessionResource } from "../testing/appShellTestSupport";
import { importTestIds as ids } from "../testing/workbookImportTestSupport";
import { AppSessionController } from "./appSessionController";
import { useWorkbookImport } from "./useWorkbookImport";

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
  let claimed = true;
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
              profile_id: "import",
              claimed,
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
    useWorkbookImport({
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
  } satisfies WorkbookImportPort;
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
    withdraw: () => {
      claimed = false;
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
describe("Workbook import application binding", () => {
  it("retains copyable work on role loss and purges it on confirmed profile withdrawal", async () => {
    const h = setup();
    await h.login();
    act(() =>
      h.hook.result.current.controller.chooseFile(
        new File(["source"], "source.csv"),
      ),
    );
    h.setRole();
    await act(async () => {
      await h.application.refreshSession();
    });
    expect(h.hook.result.current.controller.getSnapshot().canWrite).toBe(false);
    expect(h.hook.result.current.controller.getSnapshot().file?.name).toBe(
      "source.csv",
    );
    h.withdraw();
    await act(async () => {
      h.application.start();
      await flush();
    });
    expect(h.hook.result.current.controller.getSnapshot().file).toBeNull();
  });
  it("fences incident and account replacement before a pending write can be accepted", async () => {
    const h = setup();
    await h.login();
    act(() => {
      h.hook.result.current.controller.chooseFile(
        new File(["source"], "source.csv"),
      );
      void h.hook.result.current.controller.upload();
    });
    expect(h.send).toHaveBeenCalledTimes(1);
    h.setIncident(ids.secondUnit);
    act(() => h.bind());
    expect(h.hook.result.current.controller.getSnapshot().file).toBeNull();
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
    expect(h.hook.result.current.controller.getSnapshot().operation).toBeNull();
    expect(h.send).toHaveBeenCalledTimes(1);
  });
});
