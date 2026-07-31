import {
  draftCellTestId,
  rowCellTestId,
  saveStateTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import * as viewContracts from "@cartulary/view-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { TimelineWorkbookRuntimeFixture } from "../testing/TimelineWorkbookRuntimeFixture";
import {
  changeInputValue,
  cleanupTimelineWorkbookTestGlobals,
  extractTimelineJSONBody,
  extractTimelinePatchBody,
  findWorkbookCell,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  type TimelineWorkbookFetchMock,
  timelineRow,
  waitForVisibleGridRowRecordIds,
} from "../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "./models/workbookSurfaceRegistry";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

type ViewRowNormalizer = (
  contract: viewContracts.ViewContract,
  row: unknown,
  source?: string,
) => {
  readonly cells: Readonly<Record<string, { readonly value: unknown }>>;
  readonly recordId: string;
  readonly rowVersion: number;
  readonly viewSchemaId: string;
};

const exactScenarioTitle =
  "Verify Timeline query response rows render full view_row_v1 cells and preserve row identity through create, patch, validation error, and refresh.";

const timelineContract =
  viewContracts.requireViewContract(timelineViewSchemaId);
const timelineSchemaFieldKeys = timelineContract.fields.map(
  (field) => field.fieldKey,
);

function sortedCellKeys(row: { cells: Record<string, unknown> }) {
  return Object.keys(row.cells).sort();
}

function expectFullTimelineCells(row: { cells: Record<string, unknown> }) {
  expect(sortedCellKeys(row)).toEqual([...timelineSchemaFieldKeys].sort());
  expect(row.cells).not.toHaveProperty("record_id");
  expect(row.cells).not.toHaveProperty("row_version");
}

function requireViewRowNormalizer() {
  const normalizer = (
    viewContracts as unknown as {
      normalizeViewRowV1?: ViewRowNormalizer;
    }
  ).normalizeViewRowV1;
  expect(normalizer).toBeTypeOf("function");
  return normalizer as ViewRowNormalizer;
}

describe("Timeline query row identity integration", () => {
  let fetchMock: TimelineWorkbookFetchMock;

  beforeEach(() => {
    fetchMock = installTimelineWorkbookTestGlobals();
  });

  afterEach(() => {
    cleanupTimelineWorkbookTestGlobals();
  });

  it(exactScenarioTitle, async () => {
    const alpha = timelineRow({
      recordId: "20000000-0000-4000-8000-000000000701",
      rowVersion: 1,
      occurredAt: "2026-04-10T10:00:00.000Z",
      summary: "Alpha timeline row",
      details: "Alpha hidden details",
      sourceText: "Alpha source text",
      captureState: "rough",
    });
    const beta = timelineRow({
      recordId: "20000000-0000-4000-8000-000000000702",
      rowVersion: 1,
      occurredAt: "2026-04-10T10:05:00.000Z",
      summary: "Beta timeline row",
      details: "Beta hidden details",
      sourceText: "Beta source text",
      captureState: "rough",
    });

    expectFullTimelineCells(alpha);
    expectFullTimelineCells(beta);

    const normalizeViewRowV1 = requireViewRowNormalizer();
    const betaWithAdditiveMembers = {
      ...beta,
      ignored_row_member: "ignored",
      cells: {
        ...beta.cells,
        "timeline.replacement_record_id": {
          value: null,
          ignored_cell_member: "ignored",
        },
      },
    };
    const interpreted = normalizeViewRowV1(
      timelineContract,
      betaWithAdditiveMembers,
      "unit timeline query row",
    );
    expect(interpreted.recordId).toBe("20000000-0000-4000-8000-000000000702");
    expect(interpreted.rowVersion).toBe(1);
    expect(interpreted.viewSchemaId).toBe(timelineViewSchemaId);
    expect(Object.keys(interpreted.cells).sort()).toEqual(
      [...timelineSchemaFieldKeys].sort(),
    );
    expect(interpreted.cells["timeline.replacement_record_id"]).toEqual({
      value: null,
    });
    expect(interpreted.cells).not.toHaveProperty("record_id");
    expect(interpreted.cells).not.toHaveProperty("row_version");

    const missingHiddenCell = {
      ...beta,
      cells: { ...beta.cells },
    };
    delete (missingHiddenCell.cells as Record<string, unknown>)[
      "timeline.attached_evidence_ids"
    ];
    expect(() =>
      normalizeViewRowV1(
        timelineContract,
        missingHiddenCell,
        "missing hidden timeline cell",
      ),
    ).toThrow(/missing cell timeline\.attached_evidence_ids/iu);

    expect(() =>
      normalizeViewRowV1(
        timelineContract,
        {
          ...beta,
          cells: {
            ...beta.cells,
            record_id: { value: "20000000-0000-4000-8000-000000000702" },
          },
        },
        "technical cell timeline row",
      ),
    ).toThrow(/technical cell record_id/iu);
    expect(() =>
      normalizeViewRowV1(
        timelineContract,
        {
          ...beta,
          cells: {
            ...beta.cells,
            "timeline.unknown": { value: "unknown" },
          },
        },
        "unknown cell timeline row",
      ),
    ).toThrow(/unknown cell timeline\.unknown/iu);

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [alpha, beta],
      }),
    );

    const { container, rerender } = render(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        reloadToken={0}
      />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000701",
      "20000000-0000-4000-8000-000000000702",
    ]);

    const betaPatch = timelineRow({
      recordId: "20000000-0000-4000-8000-000000000702",
      rowVersion: 2,
      occurredAt: "2026-04-10T10:05:00.000Z",
      summary: "Beta patched by record id",
      details: "Beta hidden details",
      sourceText: "Beta source text",
      captureState: "rough",
    });
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000701",
        row: betaPatch,
      }),
    );
    const betaSummary = (await findWorkbookCell(
      document.body,
      timelineViewSchemaId,
      "20000000-0000-4000-8000-000000000702",
      "timeline.activity_synopsis_text",
    )) as HTMLInputElement;
    await changeInputValue(betaSummary, "Beta patched by record id");
    fireEvent.blur(betaSummary);
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(extractTimelinePatchBody(fetchMock, 1)).toMatchObject({
      base_row_version: 1,
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Beta patched by record id",
        },
      ],
    });
    await waitFor(() => {
      expect(
        screen.getByTestId(
          timelineRowVersionTestId("20000000-0000-4000-8000-000000000702"),
        ).textContent,
      ).toBe("2");
    });

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [betaPatch, alpha],
      }),
    );
    rerender(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        reloadToken={1}
      />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000702",
      "20000000-0000-4000-8000-000000000701",
    ]);
    expect(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000702",
          "timeline.activity_synopsis_text",
        ),
      ).textContent,
    ).toBe("Beta patched by record id");

    const created = timelineRow({
      recordId: "20000000-0000-4000-8000-000000000703",
      rowVersion: 1,
      occurredAt: "2026-04-10T10:15:00.000Z",
      summary: "Created through draft row",
      details: "Created hidden details",
      sourceText: "Created source text",
      captureState: "rough",
    });
    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        view_schema_id: timelineViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000702",
        row: created,
      }),
    );
    const draftSummary = screen.getByTestId(
      draftCellTestId("timeline.activity_synopsis_text"),
    ) as HTMLInputElement;
    await changeInputValue(draftSummary, "Created through draft row");
    fireEvent.keyDown(draftSummary, { key: "Enter" });
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(4);
    });
    expect(extractTimelineJSONBody(fetchMock, 3)).toMatchObject({
      "timeline.activity_synopsis_text": "Created through draft row",
    });
    await waitFor(() => {
      expect(
        screen.getByTestId(
          rowCellTestId(
            "20000000-0000-4000-8000-000000000703",
            "timeline.activity_synopsis_text",
          ),
        ).textContent,
      ).toBe("Created through draft row");
    });

    fetchMock.mockResolvedValueOnce(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: timelineViewSchemaId,
        rows: [created, alpha, betaPatch],
      }),
    );
    rerender(
      <TimelineWorkbookRuntimeFixture
        incidentId="10000000-0000-4000-8000-000000000001"
        reloadToken={2}
      />,
    );
    await waitForVisibleGridRowRecordIds(container, [
      "20000000-0000-4000-8000-000000000703",
      "20000000-0000-4000-8000-000000000701",
      "20000000-0000-4000-8000-000000000702",
    ]);
    expect(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000702",
          "timeline.activity_synopsis_text",
        ),
      ).textContent,
    ).toBe("Beta patched by record id");
    expect(
      screen.getByTestId(
        rowCellTestId(
          "20000000-0000-4000-8000-000000000701",
          "timeline.activity_utc_text",
        ),
      ).textContent,
    ).toBe("2026-04-10T10:00:00.000Z");
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
  });
});
