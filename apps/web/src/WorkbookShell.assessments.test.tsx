import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridSavedRowsSelector,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred, requireJSONBodyAt } from "./fetchMockTestSupport";
import { successEnvelope } from "./timelineWorkbookTestSupport";
import {
  buildAssessmentCreatePayload,
  confidenceScoreFromBand,
  WorkbookShell,
} from "./WorkbookShell";

const assessmentsViewSchemaId = "cartulary.view.assessments.v1";
const hostsViewSchemaId = "cartulary.view.hosts.v1";
const identitiesViewSchemaId = "cartulary.view.identities.v1";
const timelineViewSchemaId = "cartulary.view.timeline.v1";

describe("Assessment workbook surface", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    window.history.replaceState(
      {},
      "",
      `/?incident_id=incident-1&view_schema_id=${encodeURIComponent(
        assessmentsViewSchemaId,
      )}`,
    );
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal(
      "WebSocket",
      class {
        onmessage: ((event: MessageEvent) => void) | null = null;

        close() {}
      } as unknown as typeof WebSocket,
    );
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("Phase 4 U-4-WB-01 maps band-first assessment create payloads without submitting the derived band", () => {
    expect(confidenceScoreFromBand("unset")).toBeNull();
    expect(confidenceScoreFromBand("low")).toBe(25);
    expect(confidenceScoreFromBand("medium")).toBe(55);
    expect(confidenceScoreFromBand("high")).toBe(85);

    const payload = buildAssessmentCreatePayload(
      {
        assessedAt: "2026-04-24T12:00:00Z",
        assessmentState: "confirmed",
        confidenceBand: "medium",
        rationale: "Confirmed by support.",
        subjectRecordId: "host-1",
        subjectType: "host",
        supportRecordIds: ["support-1", "support-1"],
      },
      "txn-assessment-create",
    );

    expect(payload).toEqual({
      client_txn_id: "txn-assessment-create",
      "assessment.subject_ref": "host-1",
      "assessment.subject_type": "host",
      "assessment.assessment_state": "confirmed",
      "assessment.confidence_score": 55,
      "assessment.rationale": "Confirmed by support.",
      "assessment.assessed_at": "2026-04-24T12:00:00Z",
      "assessment.support_refs": {
        kind: "collection_actions_v1",
        actions: [
          {
            op: "add_record_ref",
            linked_record_id: "support-1",
          },
        ],
      },
    });
    expect(payload).not.toHaveProperty("assessment.confidence_band");
  });

  it("Phase 4 U-4-WB-02 submits assessment creates through the workbook UI", async () => {
    const createdRows: Array<Record<string, unknown>> = [];
    fetchMock.mockImplementation(async (input: RequestInfo | URL, init) => {
      const url = String(input);
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [{ incident_id: "incident-1", role: "admin" }],
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1")) {
        return successEnvelope({
          incident_id: "incident-1",
          incident_key: "IR-1",
          title: "Incident 1",
          description: null,
          severity: null,
          tlp: null,
          current_phase: null,
          primary_external_case_ref: null,
          incident_version: 1,
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1/memberships")) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "incident-1",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (url.includes("/api/v1/incidents/incident-1/workbook-preferences/")) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (url.includes(`/views/${hostsViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: hostsViewSchemaId,
          rows: [hostRow()],
        });
      }
      if (url.includes(`/views/${identitiesViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: identitiesViewSchemaId,
          rows: [],
        });
      }
      if (url.includes(`/views/${timelineViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [timelineRow()],
        });
      }
      if (
        url.includes(`/views/${assessmentsViewSchemaId}/rows`) &&
        init?.method === "POST"
      ) {
        createdRows.push(assessmentRow());
        return successEnvelope(
          {
            view_schema_id: assessmentsViewSchemaId,
            change_set_id: "change-set-1",
            row: assessmentRow(),
          },
          201,
        );
      }
      if (url.includes(`/views/${assessmentsViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: assessmentsViewSchemaId,
          rows: createdRows,
        });
      }
      return successEnvelope({});
    });

    render(<WorkbookShell incidentId="incident-1" />);

    await screen.findByTestId("assessment-create-panel");
    await waitFor(() => {
      expect(
        (screen.getByTestId("assessment-create-subject") as HTMLSelectElement)
          .value,
      ).toBe("host-1");
    });

    fireEvent.change(screen.getByTestId("assessment-create-state"), {
      target: { value: "confirmed" },
    });
    fireEvent.change(screen.getByTestId("assessment-create-confidence-band"), {
      target: { value: "high" },
    });
    fireEvent.change(screen.getByTestId("assessment-create-rationale"), {
      target: { value: "Confirmed from support." },
    });
    fireEvent.change(screen.getByTestId("assessment-create-assessed-at"), {
      target: { value: "2026-04-24T12:00:00Z" },
    });
    const supportSelect = screen.getByTestId(
      "assessment-create-support-refs",
    ) as HTMLSelectElement;
    const supportOption = supportSelect.options.item(0);
    expect(supportOption).not.toBeNull();
    (supportOption as HTMLOptionElement).selected = true;
    fireEvent.change(supportSelect);
    fireEvent.click(screen.getByTestId("assessment-create-submit"));

    await screen.findByText("Assessment created.");
    const createCallIndex = fetchMock.mock.calls.findIndex(([input]) =>
      String(input).includes(`/views/${assessmentsViewSchemaId}/rows`),
    );
    expect(createCallIndex).toBeGreaterThanOrEqual(0);
    const body = requireJSONBodyAt<Record<string, unknown>>(
      fetchMock,
      createCallIndex,
      "assessment create request body",
    );
    expect(body).toMatchObject({
      "assessment.subject_ref": "host-1",
      "assessment.subject_type": "host",
      "assessment.assessment_state": "confirmed",
      "assessment.confidence_score": 85,
      "assessment.rationale": "Confirmed from support.",
      "assessment.assessed_at": "2026-04-24T12:00:00Z",
    });
    expect(body).not.toHaveProperty("assessment.confidence_band");
    expect(body["assessment.support_refs"]).toEqual({
      kind: "collection_actions_v1",
      actions: [
        {
          op: "add_record_ref",
          linked_record_id: "support-1",
        },
      ],
    });
    const assessmentGrid = screen.getByTestId(
      gridShellTestId(assessmentsViewSchemaId),
    );
    expect(assessmentGrid.textContent).toContain("confirmed");
    expect(assessmentGrid.textContent).toContain("high");
    expect(assessmentGrid.textContent).toContain("1");
  });

  it("ignores superseded assessment query responses after rapid filter changes", async () => {
    const staleUnfiltered = deferred<Response>();
    const clearedRow = assessmentRow({
      recordId: "assessment-cleared",
      state: "cleared",
    });
    const disprovenRow = assessmentRow({
      recordId: "assessment-disproven",
      state: "disproven",
    });
    const confirmedRow = assessmentRow({
      recordId: "assessment-confirmed",
      state: "confirmed",
    });
    let assessmentQueryCount = 0;

    fetchMock.mockImplementation(async (input: RequestInfo | URL, init) => {
      const url = String(input);
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [{ incident_id: "incident-1", role: "admin" }],
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1")) {
        return successEnvelope({
          incident_id: "incident-1",
          incident_key: "IR-1",
          title: "Incident 1",
          description: null,
          severity: null,
          tlp: null,
          current_phase: null,
          primary_external_case_ref: null,
          incident_version: 1,
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1/memberships")) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "incident-1",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (url.includes("/api/v1/incidents/incident-1/workbook-preferences/")) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (url.includes(`/views/${hostsViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: hostsViewSchemaId,
          rows: [hostRow()],
        });
      }
      if (url.includes(`/views/${identitiesViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: identitiesViewSchemaId,
          rows: [],
        });
      }
      if (url.includes(`/views/${timelineViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [timelineRow()],
        });
      }
      if (url.includes(`/views/${assessmentsViewSchemaId}/query`)) {
        assessmentQueryCount += 1;
        const body = parseRequestBody(init);
        const state = assessmentStateFilterValue(body);
        if (state === "disproven") {
          return successEnvelope({
            incident_id: "incident-1",
            view_schema_id: assessmentsViewSchemaId,
            rows: [disprovenRow],
          });
        }
        if (assessmentQueryCount > 1 && state === null) {
          return staleUnfiltered.promise;
        }
        if (state === "cleared") {
          return successEnvelope({
            incident_id: "incident-1",
            view_schema_id: assessmentsViewSchemaId,
            rows: [clearedRow],
          });
        }
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: assessmentsViewSchemaId,
          rows: [clearedRow, disprovenRow, confirmedRow],
        });
      }
      return successEnvelope({});
    });

    render(<WorkbookShell incidentId="incident-1" />);

    await expectAssessmentRecordIds([
      "assessment-cleared",
      "assessment-disproven",
      "assessment-confirmed",
    ]);

    applyAssessmentStateFilter("disproven");
    await expectAssessmentRecordIds(["assessment-disproven"]);

    fireEvent.click(
      screen.getByTestId(
        gridFilterChipTestId(
          assessmentsViewSchemaId,
          "assessment.assessment_state",
        ),
      ),
    );
    applyAssessmentStateFilter("cleared");
    await expectAssessmentRecordIds(["assessment-cleared"]);

    staleUnfiltered.resolve(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: assessmentsViewSchemaId,
        rows: [clearedRow, disprovenRow, confirmedRow],
      }),
    );
    await flushMicrotasks();

    expect(currentAssessmentRecordIds()).toEqual(["assessment-cleared"]);
  });

  it("ignores superseded entity query responses after rapid host filters", async () => {
    const staleHostQuery = deferred<Response>();
    let staleHostQueryStarted = false;

    fetchMock.mockImplementation(async (input: RequestInfo | URL, init) => {
      const url = String(input);
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [{ incident_id: "incident-1", role: "admin" }],
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1")) {
        return successEnvelope({
          incident_id: "incident-1",
          incident_key: "IR-1",
          title: "Incident 1",
          description: null,
          severity: null,
          tlp: null,
          current_phase: null,
          primary_external_case_ref: null,
          incident_version: 1,
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1/memberships")) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "incident-1",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (url.includes("/api/v1/incidents/incident-1/workbook-preferences/")) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (url.includes(`/views/${hostsViewSchemaId}/query`)) {
        const value = stringFilterValue(parseRequestBody(init));
        if (value === "older") {
          staleHostQueryStarted = true;
          return staleHostQuery.promise;
        }
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: hostsViewSchemaId,
          rows: [
            hostRow(
              value === "newer"
                ? { recordId: "host-newer", displayName: "newer" }
                : { recordId: "host-initial", displayName: "initial" },
            ),
          ],
        });
      }
      if (url.includes(`/views/${identitiesViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: identitiesViewSchemaId,
          rows: [],
        });
      }
      if (url.includes(`/views/${timelineViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: timelineViewSchemaId,
          rows: [timelineRow()],
        });
      }
      if (url.includes(`/views/${assessmentsViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: assessmentsViewSchemaId,
          rows: [],
        });
      }
      return successEnvelope({});
    });

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(await screen.findByTestId("surface-tab-hosts"));
    await expectRecordIds("hosts", ["host-initial"]);

    applyHostStateFilter("older");
    await waitFor(() => {
      expect(staleHostQueryStarted).toBe(true);
    });
    applyHostStateFilter("newer");
    await expectRecordIds("hosts", ["host-newer"]);

    staleHostQuery.resolve(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: hostsViewSchemaId,
        rows: [hostRow({ recordId: "host-older", displayName: "older" })],
      }),
    );
    await flushMicrotasks();

    expect(currentRecordIds("hosts")).toEqual(["host-newer"]);
  });
});

function hostRow(options: { displayName?: string; recordId?: string } = {}) {
  const recordId = options.recordId ?? "host-1";
  const displayName = options.displayName ?? "Assessment Host";
  return {
    record_id: recordId,
    row_version: 1,
    cells: {
      "host.display_name": { value: displayName },
      "host.hostname": { value: `${recordId}.example.test` },
      "host.host_state": { value: "canonical" },
      "host.aliases": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "host.linked_event_count": { value: 0 },
    },
  };
}

function timelineRow() {
  return {
    record_id: "support-1",
    row_version: 1,
    cells: {
      "timeline.summary": { value: "Supporting timeline row" },
      "timeline.capture_state": { value: "reviewed" },
      "timeline.host_refs": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "timeline.identity_refs": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "timeline.tags": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
    },
  };
}

function assessmentRow(options: { recordId?: string; state?: string } = {}) {
  const recordId = options.recordId ?? "assessment-1";
  const state = options.state ?? "confirmed";
  return {
    record_id: recordId,
    row_version: 1,
    cells: {
      "assessment.subject_ref": { value: "host-1" },
      "assessment.subject_type": { value: "host" },
      "assessment.assessment_state": { value: state },
      "assessment.confidence_band": { value: "high" },
      "assessment.confidence_score": { value: 85 },
      "assessment.rationale": { value: "Confirmed from support." },
      "assessment.assessor": { value: "user-1" },
      "assessment.assessed_at": { value: "2026-04-24T12:00:00Z" },
      "assessment.supporting_link_count": { value: 1 },
    },
  };
}

function parseRequestBody(init: RequestInit | undefined) {
  return JSON.parse(String(init?.body ?? "{}")) as {
    filters?: Array<{
      arg?: { value?: unknown };
      field_key?: string;
    }>;
  };
}

function assessmentStateFilterValue(
  body: ReturnType<typeof parseRequestBody>,
): string | null {
  const filter = body.filters?.find(
    (candidate) => candidate.field_key === "assessment.assessment_state",
  );
  return typeof filter?.arg?.value === "string" ? filter.arg.value : null;
}

function stringFilterValue(body: ReturnType<typeof parseRequestBody>) {
  const [filter] = body.filters ?? [];
  return typeof filter?.arg?.value === "string" ? filter.arg.value : null;
}

function applyAssessmentStateFilter(value: string) {
  fireEvent.change(
    screen.getByTestId(gridFilterFieldTestId(assessmentsViewSchemaId)),
    {
      target: { value: "assessment.assessment_state" },
    },
  );
  fireEvent.change(
    screen.getByTestId(gridFilterValueTestId(assessmentsViewSchemaId)),
    {
      target: { value },
    },
  );
  fireEvent.click(
    screen.getByTestId(gridFilterApplyTestId(assessmentsViewSchemaId)),
  );
}

function applyHostStateFilter(value: string) {
  fireEvent.change(screen.getByTestId(gridFilterFieldTestId("hosts")), {
    target: { value: "host.host_state" },
  });
  fireEvent.change(screen.getByTestId(gridFilterValueTestId("hosts")), {
    target: { value },
  });
  fireEvent.click(screen.getByTestId(gridFilterApplyTestId("hosts")));
}

async function expectAssessmentRecordIds(expected: string[]) {
  await waitFor(() => {
    expect(currentAssessmentRecordIds()).toEqual(expected);
  });
}

function currentAssessmentRecordIds() {
  return currentRecordIds(assessmentsViewSchemaId);
}

async function expectRecordIds(
  surface: Parameters<typeof gridShellTestId>[0],
  expected: string[],
) {
  await waitFor(() => {
    expect(currentRecordIds(surface)).toEqual(expected);
  });
}

function currentRecordIds(surface: Parameters<typeof gridShellTestId>[0]) {
  const grid = screen.getByTestId(gridShellTestId(surface));
  return Array.from(grid.querySelectorAll(gridSavedRowsSelector())).map(
    (row) => row.getAttribute("data-grid-record-id") ?? "",
  );
}

async function flushMicrotasks() {
  await Promise.resolve();
  await Promise.resolve();
}
