import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { sessionResource } from "../testing/appShellTestSupport";
import {
  deferred,
  errorResponse,
  jsonResponse,
} from "../testing/fetchMockTestSupport";
import {
  importActorID,
  importJob,
  importJobID,
  jobEnvelope,
} from "../testing/incidentImportTestSupport";
import {
  cancelImportJob,
  importIncidentBundle,
  readImportJob,
} from "./api/incidentImportClient";
import { IncidentImportPanel } from "./IncidentImportPanel";
import { IncidentImportController } from "./incidentImportModel";
import {
  useIncidentImport,
  useIncidentImportPresentation,
} from "./useIncidentImport";

const controllers: IncidentImportController[] = [];
beforeEach(() => vi.useFakeTimers());
function Presentation({
  controller,
}: {
  controller: IncidentImportController;
}) {
  return (
    <IncidentImportPanel
      binding={useIncidentImportPresentation(controller, true)}
    />
  );
}

afterEach(() => {
  cleanup();
  for (const c of controllers.splice(0)) c.dispose();
  vi.useRealTimers();
  vi.unstubAllGlobals();
});
function start() {
  const controller = new IncidentImportController({
    admit: importIncidentBundle,
    read: readImportJob,
    cancel: cancelImportJob,
    isCurrent: () => true,
    authorizationFailed: vi.fn(),
    openIncident: async () => "opened",
  });
  controllers.push(controller);
  controller.setAuthority({ lifetime: "test-session", actorId: importActorID });
  controller.setActive(true);
  render(<Presentation controller={controller} />);
  fireEvent.change(screen.getByLabelText("Incident bundle file"), {
    target: {
      files: [
        new File(["bundle"], "incident.zip", { type: "application/zip" }),
      ],
    },
  });
  return screen.getByRole("button", { name: "Start import" });
}
async function settle() {
  await act(async () => {
    for (let i = 0; i < 8; i++) await Promise.resolve();
    await vi.advanceTimersByTimeAsync(0);
  });
}

describe("incident import workflow", () => {
  it("admits one synchronous submission and locks unresolved file replacement", () => {
    const gate = deferred<Response>();
    const fetch = vi.fn().mockReturnValue(gate.promise);
    vi.stubGlobal("fetch", fetch);
    const submit = start();
    fireEvent.click(submit);
    fireEvent.click(submit);
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(
      (screen.getByLabelText("Incident bundle file") as HTMLInputElement)
        .disabled,
    ).toBe(true);
  });
  it("offers exact admission recovery after malformed success", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(jsonResponse({ data: {} }, 202)),
    );
    fireEvent.click(start());
    await settle();
    expect(
      screen.getByRole("button", { name: "Retry admission" }),
    ).toBeTruthy();
  });
  it("locks file replacement while admission is unresolved", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockReturnValue(deferred<Response>().promise),
    );
    fireEvent.click(start());
    expect(
      (screen.getByLabelText("Incident bundle file") as HTMLInputElement)
        .disabled,
    ).toBe(true);
  });
  it("offers recovery after a lost admission response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new TypeError("controlled transport loss")),
    );
    fireEvent.click(start());
    await settle();
    expect(
      screen.getByRole("button", { name: "Retry admission" }),
    ).toBeTruthy();
  });
  it("keeps a late read from replacing the selected newer job", async () => {
    const oldRead = deferred<Response>();
    const newer = importJob("running", {
      job_id: "00000000-0000-4000-8000-000000005004",
      status_route: "/api/v1/jobs/00000000-0000-4000-8000-000000005004",
      progress: { completed: 5, total: 10 },
    });
    let admissions = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((path: string) => {
        if (path === "/api/v1/incident-bundles/import")
          return Promise.resolve(
            jsonResponse(
              jobEnvelope(
                ++admissions === 1
                  ? importJob()
                  : { ...newer, status: "queued", started_at: null },
              ),
              202,
            ),
          );
        if (path === `/api/v1/jobs/${importJobID}`) return oldRead.promise;
        return Promise.resolve(jsonResponse(jobEnvelope(newer)));
      }),
    );
    fireEvent.click(start());
    await settle();
    fireEvent.change(screen.getByLabelText("Incident bundle file"), {
      target: { files: [new File(["second"], "second.zip")] },
    });
    fireEvent.click(screen.getByRole("button", { name: "Start import" }));
    await settle();
    oldRead.resolve(
      jsonResponse(
        jobEnvelope(
          importJob("running", { progress: { completed: 2, total: 10 } }),
        ),
      ),
    );
    await settle();
    await act(() => vi.advanceTimersByTimeAsync(1000));
    expect(screen.getByText("5 of 10 completed.")).toBeTruthy();
  });
  it("exposes retry after observation failure", async () => {
    vi.useFakeTimers();
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(jsonResponse(jobEnvelope(), 202))
        .mockImplementation(async () =>
          jsonResponse({ error: { code: "unavailable" } }, 503),
        ),
    );
    fireEvent.click(start());
    await settle();
    await act(() => vi.advanceTimersByTimeAsync(1000));
    expect(
      screen.getByRole("button", { name: "Retry observation" }),
    ).toBeTruthy();
  });
  it("never opens a result from a nonterminal or wrong-code job", async () => {
    const invalid = importJob("running", {
      result_summary: importJob("succeeded").result_summary,
    });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(jsonResponse(jobEnvelope(invalid), 202)),
    );
    fireEvent.click(start());
    await settle();
    expect(
      screen.queryByRole("button", { name: "Open imported incident" }),
    ).toBeNull();
  });
});

it("restores import access only after a new confirmed session observation", async () => {
  let observations = sessionResource({
    user_id: importActorID,
    is_deployment_admin: true,
  });
  const fetch = vi
    .fn()
    .mockResolvedValueOnce(errorResponse("authorization_denied", 403))
    .mockImplementation(async (path: string) =>
      jsonResponse(jobEnvelope(), path.endsWith("/import") ? 202 : 200),
    );
  vi.stubGlobal("fetch", fetch);
  function Live({
    observedSession,
  }: {
    observedSession: ReturnType<typeof sessionResource>;
  }) {
    const options = {
      authority: { lifetime: "same-session", actorId: importActorID },
      observedSession,
      active: true,
      isCurrent: () => true,
      authorizationFailed: vi.fn(),
      openIncident: async () => "opened" as const,
    };
    return <IncidentImportPanel binding={useIncidentImport(options)} />;
  }
  const mounted = render(<Live observedSession={observations} />);
  const choose = () =>
    fireEvent.change(screen.getByLabelText("Incident bundle file"), {
      target: { files: [new File(["bytes"], "archive.tar")] },
    });
  choose();
  fireEvent.click(screen.getByRole("button", { name: "Start import" }));
  await settle();
  expect(
    (screen.getByLabelText("Incident bundle file") as HTMLInputElement)
      .disabled,
  ).toBe(true);
  mounted.rerender(<Live observedSession={observations} />);
  choose();
  fireEvent.click(screen.getByRole("button", { name: "Start import" }));
  expect(fetch).toHaveBeenCalledTimes(1);
  observations = sessionResource({
    user_id: importActorID,
    is_deployment_admin: true,
  });
  mounted.rerender(<Live observedSession={observations} />);
  choose();
  fireEvent.click(screen.getByRole("button", { name: "Start import" }));
  await settle();
  expect(
    fetch.mock.calls.filter(([path]) => String(path).endsWith("/import")),
  ).toHaveLength(2);
  expect(
    screen.getByText("Import accepted. Observe its job below."),
  ).toBeTruthy();
});
