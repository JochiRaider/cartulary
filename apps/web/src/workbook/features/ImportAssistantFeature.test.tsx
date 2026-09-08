import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ImportWriteAttempt } from "../../imports/importRequests";
import { WorkbookImportController } from "../../imports/WorkbookImportController";
import type { ImportWriteResult } from "../../services/importClient";
import type { DiscoveredImportUnit } from "../../services/importContractAdapter";
import { workbookImportTargets } from "../../services/importTargetContractAdapter";
import {
  importTestIds as ids,
  importTestJob,
  importTestMapping,
  importTestPreview,
  importTestScope,
  importTestSession,
  importTestUnit,
} from "../../testing/workbookImportTestSupport";
import { ImportAssistantFeature } from "./ImportAssistantFeature";

const controllers: WorkbookImportController[] = [];
afterEach(() => {
  cleanup();
  for (const c of controllers.splice(0)) c.dispose();
});
function setup(xlsx = false, selectionFailure = false) {
  let unit = importTestUnit(
    xlsx
      ? {
          locator_kind: "xlsx_used_range",
          locator: { sheet_name: "Sheet1" },
          source_rect_a1: "A1:A4",
          inferred_row_count: 3,
        }
      : {},
  );
  let units = [unit];
  let session = importTestSession();
  const send = vi.fn(
    async (a: ImportWriteAttempt): Promise<ImportWriteResult> => {
      if (a.kind === "upload")
        return {
          kind: "accepted",
          receipt: { kind: "job", job: importTestJob() },
        };
      if (a.kind === "mapping") {
        unit = {
          ...unit,
          approved_mapping: {
            ...importTestMapping,
            source_columns: [...importTestMapping.source_columns],
          },
          mapping_fingerprint: "a".repeat(64),
          unit_status: "mapped",
        };
        units = [unit];
        return { kind: "accepted", receipt: { kind: "unit", unit } };
      }
      if (a.kind === "region") {
        const region = importTestUnit({
          import_unit_id: ids.secondUnit,
          locator_kind: "operator_region",
          locator: { sheet_name: "Sheet1" },
        });
        units.push(region);
        return { kind: "accepted", receipt: { kind: "unit", unit: region } };
      }
      if (a.kind === "select" || a.kind === "skip") {
        if (selectionFailure) {
          selectionFailure = false;
          return {
            kind: "rejected",
            failure: {
              kind: "public",
              code: "invalid_import_state",
              reason: "unit_not_ready",
              field: null,
              status: 409,
              retryable: false,
            },
          };
        }
        const selected = a.kind === "select";
        unit = { ...unit, unit_status: selected ? "ready" : "skipped" };
        units = [unit];
        session = {
          ...session,
          session_status: selected ? "ready_to_apply" : "mapped",
          selected_unit_ids: selected ? [ids.unit] : [],
        };
        return {
          kind: "accepted",
          receipt: {
            kind: "selection",
            selection: {
              import_session_id: ids.session,
              session_status: session.session_status,
              selected_unit_ids: session.selected_unit_ids,
              unit,
            },
          },
        };
      }
      throw new Error("Unexpected test operation");
    },
  );
  const controller = new WorkbookImportController();
  controllers.push(controller);
  controller.bind({
    scope: importTestScope,
    available: true,
    closed: false,
    role: "admin",
    current: () => true,
    accessFailure: vi.fn(),
    client: {
      send,
      readJob: async () => ({
        kind: "received",
        value: importTestJob("succeeded"),
      }),
      readSession: async () => ({ kind: "received", value: session }),
      listUnits: async () => ({ kind: "received", value: units }),
      readUnit: async () => ({ kind: "received", value: unit }),
      preview: async (u: DiscoveredImportUnit) => ({
        kind: "received",
        value: importTestPreview(u),
      }),
    },
  });
  const view = render(
    <ImportAssistantFeature
      controller={controller}
      onNavigateToView={vi.fn()}
    />,
  );
  fireEvent.change(screen.getByLabelText("Source workbook"), {
    target: {
      files: [
        new File(
          ["Activity Synopsis\nObservation"],
          xlsx ? "source.xlsx" : "source.csv",
        ),
      ],
    },
  });
  fireEvent.click(screen.getByRole("button", { name: "Upload and discover" }));
  return { controller, send, view };
}
describe("ImportAssistantFeature selection lifecycle", () => {
  it("reselects a skipped mapped unit without approving a second mapping", async () => {
    const h = setup();
    fireEvent.click(
      await screen.findByRole("button", { name: "Approve mapping and select" }),
    );
    await waitFor(() =>
      expect(h.controller.getSnapshot().session?.selected_unit_ids).toEqual([
        ids.unit,
      ]),
    );
    fireEvent.click(screen.getByRole("button", { name: "Skip unit" }));
    const reselect = await screen.findByRole("button", {
      name: "Reselect unit",
    });
    await waitFor(() =>
      expect((reselect as HTMLButtonElement).disabled).toBe(false),
    );
    fireEvent.click(reselect);
    await waitFor(() =>
      expect(h.send.mock.calls.map(([a]) => a.kind)).toEqual([
        "upload",
        "mapping",
        "select",
        "skip",
        "select",
      ]),
    );
  });
  it("uses generated view targets and creates an operator-selected region", async () => {
    const h = setup(true);
    await screen.findByRole("button", { name: "Create operator region" });
    expect(
      (screen.getByLabelText("Target view") as HTMLSelectElement).options,
    ).toHaveLength(workbookImportTargets.length);
    fireEvent.change(screen.getByLabelText("Region start row"), {
      target: { value: "2" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Create operator region" }),
    );
    await screen.findByRole("heading", { name: "Unit 2: Sheet1" });
    expect(h.send.mock.calls.at(-1)?.[0]).toMatchObject({
      kind: "region",
      unitId: ids.unit,
      body: { source_rect: { start_row: 2, end_row: 4 } },
    });
    expect(
      within(screen.getByRole("region", { name: "Import unit 2" })).getByText(
        /Mapping has not been approved/,
      ),
    ).toBeTruthy();
  });
  it("keeps selection recovery beside the approved unit across presentation replacement", async () => {
    const h = setup(false, true);
    fireEvent.click(
      await screen.findByRole("button", { name: "Approve mapping and select" }),
    );
    await screen.findByRole("button", { name: "Retry selection" });
    h.view.unmount();
    render(
      <ImportAssistantFeature
        controller={h.controller}
        onNavigateToView={vi.fn()}
      />,
    );
    const unit = screen.getByRole("region", { name: "Import unit 1" });
    expect(within(unit).getByRole("alert").textContent).toContain(
      "Mapping approval is retained",
    );
    fireEvent.click(
      within(unit).getByRole("button", { name: "Retry selection" }),
    );
    await waitFor(() =>
      expect(h.send.mock.calls.map(([a]) => a.kind)).toEqual([
        "upload",
        "mapping",
        "select",
        "select",
      ]),
    );
  });
});
