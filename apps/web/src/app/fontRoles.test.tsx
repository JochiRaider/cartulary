import { appRouteTestId, saveStateTestId } from "@cartulary/ui-contracts";
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  installLandingShellFetch,
  sessionResource,
} from "../testing/appShellTestSupport";
import { TimelineWorkbookRuntimeFixture } from "../testing/TimelineWorkbookRuntimeFixture";
import {
  cleanupTimelineWorkbookTestGlobals,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  timelineRow,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../workbook/models/workbookSurfaceRegistry";
import { App } from "./App";

describe("vendored font role activation", () => {
  afterEach(() => {
    cleanup();
    cleanupTimelineWorkbookTestGlobals();
  });

  it("activates the internal hyperlegible reading profile on the rendered app shell", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    installLandingShellFetch(fetchMock, {
      incidents: [],
      session: sessionResource(),
    });

    render(<App readingProfile="hyperlegible" themeId="dark_graphite" />);

    const shell = await screen.findByTestId(appRouteTestId("app-shell"));
    expect(shell.getAttribute("data-reading-profile")).toBe("hyperlegible");
    expect(shell.style.fontFamily).toBe(
      "var(--ct-typography-accessible-reading-fontFamily)",
    );
  });

  it("marks compact workbook metadata for the condensed font role", async () => {
    const fetchMock = installTimelineWorkbookTestGlobals();
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: timelineViewSchemaId,
        rows: [
          timelineRow({
            captureState: "rough",
            recordId: "record-1",
            rowVersion: 1,
            summary: "Alpha",
          }),
        ],
      }),
    );

    render(<TimelineWorkbookRuntimeFixture incidentId="incident-1" />);

    expect(
      (await screen.findByTestId(saveStateTestId())).getAttribute(
        "data-density-role",
      ),
    ).toBe("narrow-metadata");
  });
});
