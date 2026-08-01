import {
  assessmentCreateControlTestId,
  assessmentCreatePanelTestId,
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  rowCellTestId,
  surfaceTabTestId,
  workbookAddRowButtonTestId,
  workbookFilterPopoverTriggerTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createAppAuthorizationRecoveryPort } from "../app/api/appShellClient";
import { fetchJSON } from "../services/browserApi";
import { deferred, requireJSONBodyAt } from "../testing/fetchMockTestSupport";
import {
  errorEnvelope,
  timelineRow as fullTimelineRow,
  successEnvelope,
} from "../testing/timelineWorkbookTestSupport";
import {
  buildAssessmentCreatePayload,
  confidenceScoreFromBand,
} from "./models/assessmentWorkbookModel";
import { WorkbookShell as WorkbookShellImpl } from "./WorkbookShell";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

const authorizationRecovery = createAppAuthorizationRecoveryPort({
  loadCurrentSession: (signal) => fetchJSON("/api/v1/auth/session", { signal }),
});

function WorkbookShell(
  props: Omit<Parameters<typeof WorkbookShellImpl>[0], "authorizationRecovery">,
) {
  return (
    <WorkbookShellImpl
      {...props}
      authorizationRecovery={authorizationRecovery}
    />
  );
}

const assessmentsViewSchemaId = "cartulary.view.assessments.v1";
const hostsViewSchemaId = "cartulary.view.hosts.v1";
const identitiesViewSchemaId = "cartulary.view.identities.v1";
const timelineViewSchemaId = "cartulary.view.timeline.v2";

function assessmentWorkbookStartup() {
  return successEnvelope({
    incident_id: "00000000-0000-4000-8000-000000000001",
    extension_workspace_availability: {
      schema_id: "cartulary.extension_workspace_availability.v1",
      incident_id: "00000000-0000-4000-8000-000000000001",
      workspaces: [],
    },
    selected_sheet_ref: {
      kind: "view_schema",
      id: assessmentsViewSchemaId,
    },
    selected_view_schema_id: assessmentsViewSchemaId,
    selected_saved_view: null,
    source: "explicit",
    cleared_pointers: [],
    home_sheet_ref: null,
    default_sheet_ref: null,
  });
}

describe("Assessment workbook surface", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    window.history.replaceState(
      {},
      "",
      `/?incident_id=00000000-0000-4000-8000-000000000001&view_schema_id=${encodeURIComponent(
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

  it("maps band-first assessment create payloads without submitting the derived band", () => {
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
        subjectRecordId: "00000000-0000-4000-8000-000000000101",
        subjectType: "host",
        supportRecordIds: [
          "00000000-0000-4000-8000-000000000102",
          "00000000-0000-4000-8000-000000000102",
        ],
      },
      "txn-assessment-create",
    );

    expect(payload).toEqual({
      client_txn_id: "txn-assessment-create",
      "assessment.subject_ref": "00000000-0000-4000-8000-000000000101",
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
            linked_record_id: "00000000-0000-4000-8000-000000000102",
          },
        ],
      },
    });
    expect(payload).not.toHaveProperty("assessment.confidence_band");
  });

  it("submits assessment creates through the workbook UI", async () => {
    const createdRows: Array<Record<string, unknown>> = [];
    fetchMock.mockImplementation(async (input: RequestInfo | URL, init) => {
      const url = String(input);
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [
            {
              incident_id: "00000000-0000-4000-8000-000000000001",
              role: "admin",
            },
          ],
        });
      }
      if (
        url.endsWith("/api/v1/incidents/00000000-0000-4000-8000-000000000001")
      ) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
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
      if (
        url.endsWith(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/memberships",
        )
      ) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "00000000-0000-4000-8000-000000000001",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/workbook-preferences/",
        )
      ) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/workbook-startup",
        )
      ) {
        return assessmentWorkbookStartup();
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/saved-views",
        )
      ) {
        return successEnvelope({
          saved_views: [],
        });
      }
      if (url.includes(`/views/${hostsViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: hostsViewSchemaId,
          rows: [hostRow()],
        });
      }
      if (url.includes(`/views/${identitiesViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: identitiesViewSchemaId,
          rows: [],
        });
      }
      if (url.includes(`/views/${timelineViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [timelineRow()],
        });
      }
      if (
        url.includes(`/views/${assessmentsViewSchemaId}/rows`) &&
        init?.method === "POST"
      ) {
        const createdRow = assessmentRow({
          supportRecordIds: ["00000000-0000-4000-8000-000000000102"],
        });
        createdRows.push(createdRow);
        return successEnvelope(
          {
            view_schema_id: assessmentsViewSchemaId,
            change_set_id: "00000000-0000-4000-8000-000000008400",
            row: createdRow,
          },
          201,
        );
      }
      if (url.includes(`/views/${assessmentsViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: assessmentsViewSchemaId,
          rows: createdRows,
        });
      }
      return successEnvelope({});
    });

    render(<WorkbookShell incidentId="00000000-0000-4000-8000-000000000001" />);

    fireEvent.click(
      await screen.findByTestId(
        workbookAddRowButtonTestId(assessmentsViewSchemaId),
      ),
    );
    await screen.findByTestId(assessmentCreatePanelTestId());
    await waitFor(() => {
      expect(
        (
          screen.getByTestId(
            assessmentCreateControlTestId("subject"),
          ) as HTMLSelectElement
        ).value,
      ).toBe("00000000-0000-4000-8000-000000000101");
    });

    fireEvent.change(
      screen.getByTestId(assessmentCreateControlTestId("state")),
      {
        target: { value: "confirmed" },
      },
    );
    fireEvent.change(
      screen.getByTestId(assessmentCreateControlTestId("confidence-band")),
      {
        target: { value: "high" },
      },
    );
    fireEvent.change(
      screen.getByTestId(assessmentCreateControlTestId("rationale")),
      {
        target: { value: "Confirmed from support." },
      },
    );
    fireEvent.change(
      screen.getByTestId(assessmentCreateControlTestId("assessed-at")),
      {
        target: { value: "2026-04-24T12:00:00Z" },
      },
    );
    const supportSelect = screen.getByTestId(
      assessmentCreateControlTestId("support-refs"),
    ) as HTMLSelectElement;
    const supportOption = supportSelect.options.item(0);
    expect(supportOption).not.toBeNull();
    (supportOption as HTMLOptionElement).selected = true;
    fireEvent.change(supportSelect);
    fireEvent.click(
      screen.getByTestId(assessmentCreateControlTestId("submit")),
    );

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
      "assessment.subject_ref": "00000000-0000-4000-8000-000000000101",
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
          linked_record_id: "00000000-0000-4000-8000-000000000102",
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

  it("preserves stable selection and an editable subject-only follow-on draft", async () => {
    const original = assessmentRow({
      recordId: "00000000-0000-4000-8000-000000000301",
      state: "confirmed",
      supportRecordIds: ["00000000-0000-4000-8000-000000000103"],
    });
    const filtered = assessmentRow({
      recordId: "00000000-0000-4000-8000-000000000305",
      state: "cleared",
    });
    const created = assessmentRow({
      recordId: "00000000-0000-4000-8000-000000008401",
      state: "suspected",
    });
    const assessmentRows = [original, filtered];
    const createBodies: Record<string, unknown>[] = [];
    let createAttempt = 0;

    fetchMock.mockImplementation(async (input: RequestInfo | URL, init) => {
      const url = String(input);
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [
            {
              incident_id: "00000000-0000-4000-8000-000000000001",
              role: "admin",
            },
          ],
        });
      }
      if (
        url.endsWith("/api/v1/incidents/00000000-0000-4000-8000-000000000001")
      ) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
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
      if (
        url.endsWith(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/memberships",
        )
      ) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "00000000-0000-4000-8000-000000000001",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/workbook-preferences/",
        )
      ) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/workbook-startup",
        )
      ) {
        return assessmentWorkbookStartup();
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/saved-views",
        )
      ) {
        return successEnvelope({ saved_views: [] });
      }
      if (url.includes(`/views/${hostsViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: hostsViewSchemaId,
          rows: [hostRow()],
        });
      }
      if (url.includes(`/views/${identitiesViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: identitiesViewSchemaId,
          rows: [],
        });
      }
      if (url.includes(`/views/${timelineViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [timelineRow()],
        });
      }
      if (
        url.includes(`/views/${assessmentsViewSchemaId}/rows`) &&
        init?.method === "POST"
      ) {
        createAttempt += 1;
        createBodies.push(
          JSON.parse(String(init.body ?? "{}")) as Record<string, unknown>,
        );
        if (createAttempt === 1) {
          return errorEnvelope("invalid_mutation_payload", 400);
        }
        if (createAttempt === 2) {
          return errorEnvelope("authorization_denied", 403);
        }
        assessmentRows.push(created);
        return successEnvelope(
          {
            view_schema_id: assessmentsViewSchemaId,
            change_set_id: "00000000-0000-4000-8000-000000008402",
            row: created,
          },
          201,
        );
      }
      if (url.includes(`/views/${assessmentsViewSchemaId}/query`)) {
        const state = assessmentStateFilterValue(parseRequestBody(init));
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: assessmentsViewSchemaId,
          rows:
            state === null
              ? assessmentRows
              : assessmentRows.filter(
                  (row) =>
                    row.cells["assessment.assessment_state"]?.value === state,
                ),
        });
      }
      return successEnvelope({});
    });

    render(<WorkbookShell incidentId="00000000-0000-4000-8000-000000000001" />);

    const originalCell = await screen.findByTestId(
      rowCellTestId(
        "00000000-0000-4000-8000-000000000301",
        "assessment.assessment_state",
      ),
    );
    fireEvent.click(originalCell);
    expect(originalCell.closest("tr")?.getAttribute("aria-current")).toBe(
      "true",
    );

    applyAssessmentStateFilter("cleared");
    await expectAssessmentRecordIds(["00000000-0000-4000-8000-000000000305"]);
    fireEvent.click(
      screen.getByTestId(
        workbookInspectorToggleTestId(assessmentsViewSchemaId),
      ),
    );
    fireEvent.click(
      screen.getByTestId(
        workbookInspectorFeatureActionTestId(
          assessmentsViewSchemaId,
          "create_related.assessment",
        ),
      ),
    );

    expect(screen.getByText("Append follow-on assessment")).toBeTruthy();
    expect(
      screen.getByText(/00000000-0000-4000-8000-000000000103/u),
    ).toBeTruthy();
    expect(assessmentControlValue("subject")).toBe(
      "00000000-0000-4000-8000-000000000101",
    );
    expect(assessmentControlValue("subject-type")).toBe("host");
    expect(assessmentControlValue("state")).toBe("unknown");
    expect(assessmentControlValue("confidence-band")).toBe("unset");
    expect(assessmentControlValue("rationale")).toBe("");
    expect(assessmentControlValue("assessed-at")).toBe("");
    const initialSupportPicker = screen.getByTestId(
      assessmentCreateControlTestId("support-refs"),
    ) as HTMLSelectElement;
    expect(Array.from(initialSupportPicker.selectedOptions)).toHaveLength(0);
    expect(
      Array.from(initialSupportPicker.options).map((option) => ({
        label: option.text,
        value: option.value,
      })),
    ).toEqual([
      {
        label: "Supporting timeline row",
        value: "00000000-0000-4000-8000-000000000102",
      },
    ]);

    fireEvent.change(
      screen.getByTestId(assessmentCreateControlTestId("rationale")),
      { target: { value: "Cancelled rationale" } },
    );
    fireEvent.keyDown(screen.getByTestId(assessmentCreatePanelTestId()), {
      key: "Escape",
    });
    expect(screen.queryByTestId(assessmentCreatePanelTestId())).toBeNull();
    expect(createAttempt).toBe(0);

    fireEvent.click(
      screen.getByTestId(
        gridFilterChipTestId(
          assessmentsViewSchemaId,
          "assessment.assessment_state",
        ),
      ),
    );
    const restoredOriginalCell = await screen.findByTestId(
      rowCellTestId(
        "00000000-0000-4000-8000-000000000301",
        "assessment.assessment_state",
      ),
    );
    expect(
      restoredOriginalCell.closest("tr")?.getAttribute("aria-current"),
    ).toBe("true");

    fireEvent.click(
      screen.getByTestId(
        workbookInspectorToggleTestId(assessmentsViewSchemaId),
      ),
    );
    fireEvent.click(
      screen.getByTestId(
        workbookInspectorFeatureActionTestId(
          assessmentsViewSchemaId,
          "create_related.assessment",
        ),
      ),
    );
    fireEvent.change(
      screen.getByTestId(assessmentCreateControlTestId("state")),
      { target: { value: "suspected" } },
    );
    fireEvent.change(
      screen.getByTestId(assessmentCreateControlTestId("rationale")),
      { target: { value: "Fresh follow-on rationale" } },
    );
    const followOnSupportPicker = screen.getByTestId(
      assessmentCreateControlTestId("support-refs"),
    ) as HTMLSelectElement;
    const supportOption = followOnSupportPicker.options.item(0);
    expect(supportOption).not.toBeNull();
    (supportOption as HTMLOptionElement).selected = true;
    fireEvent.change(followOnSupportPicker);

    fireEvent.click(
      screen.getByTestId(assessmentCreateControlTestId("submit")),
    );
    await screen.findByText("invalid_mutation_payload");
    expect(assessmentControlValue("rationale")).toBe(
      "Fresh follow-on rationale",
    );
    expect(selectedAssessmentSupportRecordIds()).toEqual([
      "00000000-0000-4000-8000-000000000102",
    ]);
    expect(
      restoredOriginalCell.closest("tr")?.getAttribute("aria-current"),
    ).toBe("true");

    fireEvent.click(
      screen.getByTestId(assessmentCreateControlTestId("submit")),
    );
    await screen.findByText("authorization_denied");
    expect(assessmentControlValue("rationale")).toBe(
      "Fresh follow-on rationale",
    );

    fireEvent.click(
      screen.getByTestId(assessmentCreateControlTestId("submit")),
    );
    await screen.findByText("Assessment created.");
    await screen.findByTestId(
      rowCellTestId(
        "00000000-0000-4000-8000-000000008401",
        "assessment.assessment_state",
      ),
    );
    expect(
      restoredOriginalCell.closest("tr")?.getAttribute("aria-current"),
    ).toBe("true");
    expect(assessmentControlValue("rationale")).toBe("");
    expect(selectedAssessmentSupportRecordIds()).toEqual([]);

    expect(createBodies[0]).toMatchObject({
      "assessment.subject_ref": "00000000-0000-4000-8000-000000000101",
      "assessment.subject_type": "host",
      "assessment.assessment_state": "suspected",
      "assessment.confidence_score": null,
      "assessment.rationale": "Fresh follow-on rationale",
    });
    expect(createBodies[0]).not.toHaveProperty("assessment.assessed_at");
    expect(createBodies[0]).not.toHaveProperty("assessment.assessor");
    expect(createBodies[0]).not.toHaveProperty("assessment.supersedes");
    expect(createBodies[0]).not.toHaveProperty("supersedes");
    expect(createBodies[0]?.["assessment.support_refs"]).toEqual({
      kind: "collection_actions_v1",
      actions: [
        {
          op: "add_record_ref",
          linked_record_id: "00000000-0000-4000-8000-000000000102",
        },
      ],
    });
    expect(assessmentPayloadWithoutClientTxn(createBodies[1])).toEqual(
      assessmentPayloadWithoutClientTxn(createBodies[0]),
    );
    expect(assessmentPayloadWithoutClientTxn(createBodies[2])).toEqual(
      assessmentPayloadWithoutClientTxn(createBodies[0]),
    );
    expect(new Set(createBodies.map((body) => body.client_txn_id)).size).toBe(
      3,
    );

    fireEvent.click(
      screen.getByTestId(
        workbookInspectorCloseButtonTestId(assessmentsViewSchemaId),
      ),
    );
    expect(screen.queryByTestId(assessmentCreatePanelTestId())).toBeNull();
    expect(
      restoredOriginalCell.closest("tr")?.getAttribute("aria-current"),
    ).toBe("true");
  });

  it("ignores superseded assessment query responses after rapid filter changes", async () => {
    const staleUnfiltered = deferred<Response>();
    const clearedRow = assessmentRow({
      recordId: "00000000-0000-4000-8000-000000000302",
      state: "cleared",
    });
    const disprovenRow = assessmentRow({
      recordId: "00000000-0000-4000-8000-000000000304",
      state: "disproven",
    });
    const confirmedRow = assessmentRow({
      recordId: "00000000-0000-4000-8000-000000000303",
      state: "confirmed",
    });
    let assessmentQueryCount = 0;

    fetchMock.mockImplementation(async (input: RequestInfo | URL, init) => {
      const url = String(input);
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [
            {
              incident_id: "00000000-0000-4000-8000-000000000001",
              role: "admin",
            },
          ],
        });
      }
      if (
        url.endsWith("/api/v1/incidents/00000000-0000-4000-8000-000000000001")
      ) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
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
      if (
        url.endsWith(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/memberships",
        )
      ) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "00000000-0000-4000-8000-000000000001",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/workbook-preferences/",
        )
      ) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/workbook-startup",
        )
      ) {
        return assessmentWorkbookStartup();
      }
      if (url.includes(`/views/${hostsViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: hostsViewSchemaId,
          rows: [hostRow()],
        });
      }
      if (url.includes(`/views/${identitiesViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: identitiesViewSchemaId,
          rows: [],
        });
      }
      if (url.includes(`/views/${timelineViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
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
            incident_id: "00000000-0000-4000-8000-000000000001",
            view_schema_id: assessmentsViewSchemaId,
            rows: [disprovenRow],
          });
        }
        if (assessmentQueryCount > 1 && state === null) {
          return staleUnfiltered.promise;
        }
        if (state === "cleared") {
          return successEnvelope({
            incident_id: "00000000-0000-4000-8000-000000000001",
            view_schema_id: assessmentsViewSchemaId,
            rows: [clearedRow],
          });
        }
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: assessmentsViewSchemaId,
          rows: [clearedRow, disprovenRow, confirmedRow],
        });
      }
      return successEnvelope({});
    });

    render(<WorkbookShell incidentId="00000000-0000-4000-8000-000000000001" />);

    await expectAssessmentRecordIds([
      "00000000-0000-4000-8000-000000000302",
      "00000000-0000-4000-8000-000000000304",
      "00000000-0000-4000-8000-000000000303",
    ]);

    applyAssessmentStateFilter("disproven");
    await expectAssessmentRecordIds(["00000000-0000-4000-8000-000000000304"]);

    fireEvent.click(
      screen.getByTestId(
        gridFilterChipTestId(
          assessmentsViewSchemaId,
          "assessment.assessment_state",
        ),
      ),
    );
    applyAssessmentStateFilter("cleared");
    await expectAssessmentRecordIds(["00000000-0000-4000-8000-000000000302"]);

    staleUnfiltered.resolve(
      successEnvelope({
        incident_id: "00000000-0000-4000-8000-000000000001",
        view_schema_id: assessmentsViewSchemaId,
        rows: [clearedRow, disprovenRow, confirmedRow],
      }),
    );
    await flushMicrotasks();

    expect(currentAssessmentRecordIds()).toEqual([
      "00000000-0000-4000-8000-000000000302",
    ]);
  });

  it("ignores superseded entity query responses after rapid host filters", async () => {
    const staleHostQuery = deferred<Response>();
    let staleHostQueryStarted = false;

    fetchMock.mockImplementation(async (input: RequestInfo | URL, init) => {
      const url = String(input);
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [
            {
              incident_id: "00000000-0000-4000-8000-000000000001",
              role: "admin",
            },
          ],
        });
      }
      if (
        url.endsWith("/api/v1/incidents/00000000-0000-4000-8000-000000000001")
      ) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
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
      if (
        url.endsWith(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/memberships",
        )
      ) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "00000000-0000-4000-8000-000000000001",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/workbook-preferences/",
        )
      ) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (
        url.includes(
          "/api/v1/incidents/00000000-0000-4000-8000-000000000001/workbook-startup",
        )
      ) {
        return assessmentWorkbookStartup();
      }
      if (url.includes(`/views/${hostsViewSchemaId}/query`)) {
        const value = stringFilterValue(parseRequestBody(init));
        if (value === "older") {
          staleHostQueryStarted = true;
          return staleHostQuery.promise;
        }
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: hostsViewSchemaId,
          rows: [
            hostRow(
              value === "newer"
                ? {
                    recordId: "00000000-0000-4000-8000-000000000112",
                    displayName: "newer",
                  }
                : {
                    recordId: "00000000-0000-4000-8000-000000000111",
                    displayName: "initial",
                  },
            ),
          ],
        });
      }
      if (url.includes(`/views/${identitiesViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: identitiesViewSchemaId,
          rows: [],
        });
      }
      if (url.includes(`/views/${timelineViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: timelineViewSchemaId,
          rows: [timelineRow()],
        });
      }
      if (url.includes(`/views/${assessmentsViewSchemaId}/query`)) {
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          view_schema_id: assessmentsViewSchemaId,
          rows: [],
        });
      }
      return successEnvelope({});
    });

    render(<WorkbookShell incidentId="00000000-0000-4000-8000-000000000001" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(hostsViewSchemaId)),
    );
    await expectRecordIds(hostsViewSchemaId, [
      "00000000-0000-4000-8000-000000000111",
    ]);

    applyHostStateFilter("older");
    await waitFor(() => {
      expect(staleHostQueryStarted).toBe(true);
    });
    applyHostStateFilter("newer");
    await expectRecordIds(hostsViewSchemaId, [
      "00000000-0000-4000-8000-000000000112",
    ]);

    staleHostQuery.resolve(
      successEnvelope({
        incident_id: "00000000-0000-4000-8000-000000000001",
        view_schema_id: hostsViewSchemaId,
        rows: [
          hostRow({
            recordId: "00000000-0000-4000-8000-000000000113",
            displayName: "older",
          }),
        ],
      }),
    );
    await flushMicrotasks();

    expect(currentRecordIds(hostsViewSchemaId)).toEqual([
      "00000000-0000-4000-8000-000000000112",
    ]);
  });
});

function hostRow(options: { displayName?: string; recordId?: string } = {}) {
  const recordId = options.recordId ?? "00000000-0000-4000-8000-000000000101";
  const displayName = options.displayName ?? "Assessment Host";
  return {
    record_id: recordId,
    row_version: 1,
    cells: {
      "host.display_name": { value: displayName },
      "host.hostname": { value: `${recordId}.example.test` },
      "host.aad_device_id": { value: null },
      "host.fqdn": { value: `${recordId}.example.test` },
      "host.reusable_identifiers": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "host.aliases": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "host.host_state": { value: "canonical" },
      "host.linked_event_count": { value: 0 },
      "host.evidence_count": { value: 0 },
      "host.location": { value: null },
      "host.os_platform": { value: null },
      "host.business_owner": { value: null },
      "host.criticality": { value: null },
      "host.containment_status": { value: null },
      "host.edited_at": { value: "2026-04-24T12:00:00.000Z" },
    },
    group_values: {
      "host.host_state": "canonical",
      "host.criticality": null,
      "host.containment_status": null,
    },
  };
}

function timelineRow() {
  return fullTimelineRow({
    recordId: "00000000-0000-4000-8000-000000000102",
    rowVersion: 1,
    summary: "Supporting timeline row",
    captureState: "reviewed",
  });
}

function assessmentRow(
  options: {
    recordId?: string;
    state?: string;
    supportRecordIds?: string[];
  } = {},
) {
  const recordId = options.recordId ?? "00000000-0000-4000-8000-000000008403";
  const state = options.state ?? "confirmed";
  const supportRecordIds = options.supportRecordIds ?? [];
  return {
    record_id: recordId,
    row_version: 1,
    cells: {
      "assessment.subject_ref": {
        value: "00000000-0000-4000-8000-000000000101",
      },
      "assessment.subject_type": { value: "host" },
      "assessment.assessment_state": { value: state },
      "assessment.confidence_band": { value: "high" },
      "assessment.confidence_score": { value: 85 },
      "assessment.rationale": { value: "Confirmed from support." },
      "assessment.assessor": { value: "user-1" },
      "assessment.assessed_at": { value: "2026-04-24T12:00:00Z" },
      "assessment.support_refs": {
        value: {
          kind: "collection_value_v1",
          ordered: false,
          items: supportRecordIds.map((supportRecordId) => ({
            item_ref: `support:${supportRecordId}`,
            item_kind: "record_ref",
            linked_record_id: supportRecordId,
          })),
        },
      },
      "assessment.supporting_link_count": { value: supportRecordIds.length },
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
  fireEvent.click(
    screen.getByTestId(
      workbookFilterPopoverTriggerTestId(assessmentsViewSchemaId),
    ),
  );
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
  fireEvent.click(
    screen.getByTestId(workbookFilterPopoverTriggerTestId(hostsViewSchemaId)),
  );
  fireEvent.change(
    screen.getByTestId(gridFilterFieldTestId(hostsViewSchemaId)),
    {
      target: { value: "host.host_state" },
    },
  );
  fireEvent.change(
    screen.getByTestId(gridFilterValueTestId(hostsViewSchemaId)),
    {
      target: { value },
    },
  );
  fireEvent.click(screen.getByTestId(gridFilterApplyTestId(hostsViewSchemaId)));
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

function assessmentControlValue(
  control: Parameters<typeof assessmentCreateControlTestId>[0],
): string {
  return (
    screen.getByTestId(assessmentCreateControlTestId(control)) as
      | HTMLInputElement
      | HTMLSelectElement
      | HTMLTextAreaElement
  ).value;
}

function selectedAssessmentSupportRecordIds(): string[] {
  const picker = screen.getByTestId(
    assessmentCreateControlTestId("support-refs"),
  ) as HTMLSelectElement;
  return Array.from(picker.selectedOptions).map((option) => option.value);
}

function assessmentPayloadWithoutClientTxn(
  payload: Record<string, unknown> | undefined,
): Record<string, unknown> {
  if (payload === undefined) {
    return {};
  }
  const { client_txn_id: _clientTxnId, ...semanticPayload } = payload;
  return semanticPayload;
}

async function flushMicrotasks() {
  await Promise.resolve();
  await Promise.resolve();
}
