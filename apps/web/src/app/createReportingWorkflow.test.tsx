import { act, cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { ReportingFeature } from "../reporting/ReportingPanel";
import { sessionResource } from "../testing/appShellTestSupport";
import { AppSessionController } from "./appSessionController";
import { createReportingWorkflow } from "./createReportingWorkflow";

const incidentId = "10000000-0000-4000-8000-000000000001",
  actorId = "10000000-0000-4000-8000-000000000002";
const cleanupOwners: (() => void)[] = [];
afterEach(() => {
  cleanup();
  for (const dispose of cleanupOwners.splice(0)) dispose();
  vi.unstubAllGlobals();
});
const flush = async () => {
  for (let i = 0; i < 30; i++) await Promise.resolve();
};
function setup() {
  let role: "editor" | "viewer" = "editor",
    major = 1,
    claimed = true;
  const application = new AppSessionController({
    session: async () => ({
      ok: true,
      status: 200,
      payload: {
        meta: { request_id: "reporting" },
        data: sessionResource({
          user_id: actorId,
          memberships: [{ incident_id: incidentId, role }],
        }),
      },
    }),
    preferences: async () => ({ ok: false, status: 503, payload: {} }),
    extensions: async () => ({
      ok: true,
      status: 200,
      payload: {
        meta: { request_id: "extensions" },
        data: {
          extensions: [
            {
              profile_id: "snapshot_reporting",
              contract_major: major,
              claimable: true,
              claimed,
              capabilities: [],
              workspace_keys: [],
              route_families: [
                "/api/v1/snapshots",
                "/api/v1/releases",
                "/api/v1/incidents/{incident_id}/report-compositions",
              ],
            },
          ],
        },
      },
    }),
  });
  const workflow = createReportingWorkflow(() => ({
    accepting: () => true,
    sessionController: application,
    currentIncidentId: () => incidentId,
  }));
  const disconnect = application.subscribe(workflow.synchronize);
  cleanupOwners.push(() => {
    disconnect();
    workflow.dispose();
    application.dispose();
  });
  const props = {
    incidentId,
    currentIncidentRole: "editor" as const,
    activeSection: "snapshot-reports" as const,
    controller: workflow.controller,
    attachSurface: workflow.attachSurface,
  };
  const refresh = async () => {
    await act(async () => {
      await application.refreshSession();
      application.start();
      await flush();
    });
  };
  return {
    application,
    workflow,
    props,
    refresh,
    setRole: () => {
      role = "viewer";
    },
    unsupported: () => {
      major = 99;
    },
    withdraw: () => {
      claimed = false;
    },
  };
}
it("binds reporting controls to claimed supported routes and current incident roles", async () => {
  const h = setup();
  await h.refresh();
  render(<ReportingFeature {...h.props} />);
  expect(
    screen
      .getByRole("button", { name: "Create snapshot" })
      .hasAttribute("disabled"),
  ).toBe(false);
  expect(
    screen
      .getByRole("button", { name: "Render internal report" })
      .hasAttribute("disabled"),
  ).toBe(true);
  h.setRole();
  await h.refresh();
  expect(
    screen
      .getByRole("button", { name: "Create snapshot" })
      .hasAttribute("disabled"),
  ).toBe(true);
  expect(
    screen.getByText(/requires an editor, reviewer, or admin/).isConnected,
  ).toBe(true);
  h.unsupported();
  await h.refresh();
  expect(h.workflow.controller.getSnapshot().available).toBe(false);
  expect(
    screen
      .getByRole("button", { name: "Create snapshot" })
      .hasAttribute("disabled"),
  ).toBe(true);
});
it("retains exact reporting admission across drawer detachment and purges withdrawn profile state", async () => {
  const h = setup();
  await h.refresh();
  const fetch = vi.fn(() => new Promise<Response>(() => {}));
  vi.stubGlobal("fetch", fetch);
  const rendered = render(<ReportingFeature {...h.props} />);
  let pending!: Promise<void>;
  act(() => {
    pending = h.workflow.controller.createSnapshot();
  });
  expect(fetch).toHaveBeenCalledTimes(1);
  const attempt = h.workflow.controller.getSnapshot().attempt;
  rendered.unmount();
  await act(async () => {
    await pending;
  });
  render(<ReportingFeature {...h.props} />);
  expect(h.workflow.controller.getSnapshot().attempt).toBe(attempt);
  expect(
    screen.getByRole("button", { name: "Recover request" }).isConnected,
  ).toBe(true);
  expect(fetch).toHaveBeenCalledTimes(1);
  h.withdraw();
  await h.refresh();
  expect(h.workflow.controller.getSnapshot().attempt).toBeNull();
  expect(screen.queryByRole("button", { name: "Recover request" })).toBeNull();
});
