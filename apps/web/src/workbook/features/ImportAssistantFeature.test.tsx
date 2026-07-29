import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  approveWorkbookImportMapping,
  createWorkbookImportRegion,
  setWorkbookImportUnitSelection,
  uploadAndDiscoverWorkbookImport,
  type WorkbookImportDiscovery,
  type WorkbookImportUnitDiscovery,
} from "../../imports/importCoordinator";
import { readyExtensionAvailability } from "../../testing/extensionAvailabilityTestSupport";
import { ImportAssistantFeature } from "./ImportAssistantFeature";

vi.mock("../../imports/importCoordinator", async (importOriginal) => {
  const actual =
    await importOriginal<typeof import("../../imports/importCoordinator")>();
  return {
    ...actual,
    approveWorkbookImportMapping: vi.fn(),
    createWorkbookImportRegion: vi.fn(),
    setWorkbookImportUnitSelection: vi.fn(),
    uploadAndDiscoverWorkbookImport: vi.fn(),
  };
});

const sessionID = "00000000-0000-4000-8000-000000000001";
const unitID = "00000000-0000-4000-8000-000000000002";

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe("ImportAssistantFeature selection lifecycle", () => {
  it("reselects a skipped mapped unit without approving a second mapping", async () => {
    const discovery = workbookDiscovery();
    const first = discovery.units[0];
    if (first === undefined) {
      throw new Error("missing import discovery fixture");
    }
    vi.mocked(uploadAndDiscoverWorkbookImport).mockResolvedValue(discovery);
    vi.mocked(approveWorkbookImportMapping).mockResolvedValue({
      ...first.unit,
      unit_status: "mapped",
      mapping_fingerprint: "a".repeat(64),
    });
    vi.mocked(setWorkbookImportUnitSelection).mockImplementation(
      async ({ selected }) => ({
        ...first.unit,
        unit_status: selected ? "ready" : "skipped",
        mapping_fingerprint: "a".repeat(64),
      }),
    );

    render(
      <ImportAssistantFeature
        apiBase={undefined}
        availability={readyExtensionAvailability()}
        currentIncidentRole="admin"
        incidentId="incident-1"
        onNavigateToView={() => undefined}
      />,
    );

    const sourceInput = screen.getByLabelText("Source workbook");
    fireEvent.change(sourceInput, {
      target: {
        files: [
          new File(["summary\none\n"], "selection.csv", {
            type: "text/csv",
          }),
        ],
      },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Upload and discover" }),
    );
    fireEvent.click(
      await screen.findByRole("button", { name: "Approve mapping and select" }),
    );
    await waitFor(() => {
      expect(approveWorkbookImportMapping).toHaveBeenCalledTimes(1);
      expect(setWorkbookImportUnitSelection).toHaveBeenLastCalledWith(
        expect.objectContaining({ selected: true }),
      );
    });

    fireEvent.click(screen.getByRole("button", { name: "Skip unit" }));
    const reselectButton = await screen.findByRole("button", {
      name: "Reselect unit",
    });
    await waitFor(() => {
      expect(setWorkbookImportUnitSelection).toHaveBeenLastCalledWith(
        expect.objectContaining({ selected: false }),
      );
    });

    fireEvent.click(reselectButton);
    await waitFor(() => {
      expect(setWorkbookImportUnitSelection).toHaveBeenCalledTimes(3);
      expect(setWorkbookImportUnitSelection).toHaveBeenLastCalledWith(
        expect.objectContaining({ selected: true }),
      );
    });
    expect(approveWorkbookImportMapping).toHaveBeenCalledTimes(1);
    expect(
      screen.getByText("Unit reselected with its approved mapping."),
    ).toBeTruthy();
  });

  it("uses generated view targets and creates an operator-selected region", async () => {
    const discovery = xlsxWorkbookDiscovery();
    const created = operatorRegionDiscovery();
    vi.mocked(uploadAndDiscoverWorkbookImport).mockResolvedValue(discovery);
    vi.mocked(createWorkbookImportRegion).mockResolvedValue(created);

    render(
      <ImportAssistantFeature
        apiBase={undefined}
        availability={readyExtensionAvailability()}
        currentIncidentRole="admin"
        incidentId="incident-1"
        onNavigateToView={() => undefined}
      />,
    );

    fireEvent.change(screen.getByLabelText("Source workbook"), {
      target: {
        files: [
          new File(["xlsx"], "regions.xlsx", {
            type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
          }),
        ],
      },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Upload and discover" }),
    );

    await screen.findByRole("button", { name: "Create operator region" });
    expect(
      (screen.getByLabelText("Target view") as HTMLSelectElement).options,
    ).toHaveLength(14);
    fireEvent.change(screen.getByLabelText("Region start row"), {
      target: { value: "2" },
    });
    fireEvent.change(screen.getByLabelText("Region end column"), {
      target: { value: "2" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Create operator region" }),
    );

    await waitFor(() => {
      expect(createWorkbookImportRegion).toHaveBeenCalledWith(
        expect.objectContaining({
          baseUnitId: unitID,
          sessionId: sessionID,
          sourceRect: {
            startRow: 2,
            startColumn: 1,
            endRow: 4,
            endColumn: 2,
          },
        }),
      );
    });
    expect(
      await screen.findByText(
        "Operator region created. Review its mapping separately.",
      ),
    ).toBeTruthy();
    expect(screen.getByText(/^Unit 2:/u).tagName).toBe("H3");
  });
});

function workbookDiscovery(): WorkbookImportDiscovery {
  return {
    session: {
      assistant_profile: "phase2_workbook_import_v1",
      blocking_diagnostics: [],
      created_at: "2026-07-29T00:00:00Z",
      created_by_user_id: "00000000-0000-4000-8000-000000000003",
      import_session_id: sessionID,
      incident_id: "00000000-0000-4000-8000-000000000004",
      nonblocking_warning_codes: [],
      original_filename: "selection.csv",
      parser_profile_id: "cartulary.import.phase2_workbook_import.v1",
      parser_version: "phase11_import_adapter_v1",
      selected_unit_ids: [],
      session_status: "discovered",
      source_content_sha256: "b".repeat(64),
      source_file_kind: "csv",
    },
    units: [
      {
        unit: {
          data_start_row_ref: 2,
          header_row_ref: 1,
          import_session_id: sessionID,
          import_unit_id: unitID,
          inferred_column_count: 1,
          inferred_row_count: 1,
          locator: { file: true },
          locator_kind: "csv_file",
          source_rect_a1: "A1:A2",
          unit_status: "discovered",
          warning_codes: [],
        },
        preview: {
          columns: [
            {
              source_column_ordinal: 1,
              source_header_text: "summary",
            },
          ],
          data_start_row_ref: 2,
          header_row_ref: 1,
          import_session_id: sessionID,
          import_unit_id: unitID,
          inferred_column_count: 1,
          inferred_row_count: 1,
          locator: { file: true },
          locator_kind: "csv_file",
          preview_rows: [
            {
              cells: [
                {
                  cell_kind: "string",
                  display_text: "one",
                  source_column_ordinal: 1,
                },
              ],
              source_row_ref: 2,
            },
          ],
          source_rect_a1: "A1:A2",
          truncated: false,
          unit_status: "discovered",
          warning_codes: [],
        },
      },
    ],
  };
}

function xlsxWorkbookDiscovery(): WorkbookImportDiscovery {
  const discovery = workbookDiscovery();
  const first = discovery.units[0];
  if (first === undefined) {
    throw new Error("missing import discovery fixture");
  }
  return {
    ...discovery,
    session: {
      ...discovery.session,
      original_filename: "regions.xlsx",
      source_file_kind: "xlsx",
    },
    units: [
      {
        unit: {
          ...first.unit,
          locator: { sheet_name: "Data" },
          locator_kind: "xlsx_used_range",
          source_rect_a1: "A1:C4",
          inferred_column_count: 3,
          inferred_row_count: 3,
        },
        preview: {
          ...first.preview,
          locator: { sheet_name: "Data" },
          locator_kind: "xlsx_used_range",
          source_rect_a1: "A1:C4",
          inferred_column_count: 3,
          inferred_row_count: 3,
          columns: [
            { source_column_ordinal: 1, source_header_text: "summary" },
            { source_column_ordinal: 2, source_header_text: "details" },
            { source_column_ordinal: 3, source_header_text: "source" },
          ],
        },
      },
    ],
  };
}

function operatorRegionDiscovery(): WorkbookImportUnitDiscovery {
  const base = xlsxWorkbookDiscovery().units[0];
  if (base === undefined) {
    throw new Error("missing XLSX import discovery fixture");
  }
  const regionID = "00000000-0000-4000-8000-000000000005";
  return {
    unit: {
      ...base.unit,
      import_unit_id: regionID,
      locator: {
        sheet_name: "Data",
        base_unit_id: unitID,
        region_sequence: 1,
      },
      locator_kind: "operator_region",
      source_rect_a1: "A2:B4",
      header_row_ref: 2,
      data_start_row_ref: 3,
      inferred_column_count: 2,
      inferred_row_count: 2,
    },
    preview: {
      ...base.preview,
      import_unit_id: regionID,
      locator: {
        sheet_name: "Data",
        base_unit_id: unitID,
        region_sequence: 1,
      },
      locator_kind: "operator_region",
      source_rect_a1: "A2:B4",
      header_row_ref: 2,
      data_start_row_ref: 3,
      inferred_column_count: 2,
      inferred_row_count: 2,
      columns: [
        { source_column_ordinal: 1, source_header_text: "one" },
        { source_column_ordinal: 2, source_header_text: "two" },
      ],
    },
  };
}
