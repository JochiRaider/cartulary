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
  setWorkbookImportUnitSelection,
  uploadAndDiscoverWorkbookImport,
  type WorkbookImportDiscovery,
} from "../../imports/importCoordinator";
import { readyExtensionAvailability } from "../../testing/extensionAvailabilityTestSupport";
import { ImportAssistantFeature } from "./ImportAssistantFeature";

vi.mock("../../imports/importCoordinator", async (importOriginal) => {
  const actual =
    await importOriginal<typeof import("../../imports/importCoordinator")>();
  return {
    ...actual,
    approveWorkbookImportMapping: vi.fn(),
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
    vi.mocked(uploadAndDiscoverWorkbookImport).mockResolvedValue(discovery);
    vi.mocked(approveWorkbookImportMapping).mockResolvedValue({
      ...discovery.units[0]!.unit,
      unit_status: "mapped",
      mapping_fingerprint: "a".repeat(64),
    });
    vi.mocked(setWorkbookImportUnitSelection).mockImplementation(
      async ({ selected }) => ({
        ...discovery.units[0]!.unit,
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
