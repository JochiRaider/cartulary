import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { ImportClient } from "../services/importClient";
import type { CommittedImportSource } from "../services/importContractAdapter";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { jsonResponse } from "../testing/fetchMockTestSupport";
import {
  importTestIds as ids,
  importTestEnvelope,
  importTestMapping,
} from "../testing/workbookImportTestSupport";
import { CommittedSourcesPanel } from "./CommittedSourcesPanel";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
const source: CommittedImportSource = {
  schema_id: "cartulary.imports.committed_source.v1",
  committed_source_id: ids.unit,
  incident_id: ids.incident,
  content_sha256: "b".repeat(64),
  byte_size: 123,
  source_file_kind: "csv",
  source_media_type: "text/csv",
  original_filename: "<img src=x onerror=alert(1)>.csv",
  parser_profile_id: "cartulary.import.workbook.v2",
  parser_version: "workbook_import_adapter_v2",
  locator_kind: "csv_file",
  locator: "",
  source_rect_a1: "A1:A2",
  header_row_ref: 1,
  data_start_row_ref: 2,
  approved_mapping: {
    ...importTestMapping,
    source_columns: importTestMapping.source_columns.map((column) => ({
      ...column,
    })),
  },
  mapping_fingerprint: "a".repeat(64),
  target_kind: "view_schema",
  target_view_schema_id: "cartulary.view.timeline.v2",
  extension_profile_id: "",
  origin_import_session_id: ids.session,
  origin_import_unit_id: ids.unit,
  created_by_user_id: ids.actor,
  committed_at: "2026-11-06T12:00:00Z",
};
const client = () =>
  new ImportClient({
    availability: readyExtensionAvailability(ids.incident),
    incidentId: ids.incident,
  });

it("inspects distinct application identities and renders source text inertly with deliberate download", async () => {
  const second = { ...source, committed_source_id: ids.secondUnit };
  const summary = (value: CommittedImportSource) => ({
    committed_source_id: value.committed_source_id,
    content_sha256: value.content_sha256,
    original_filename: value.original_filename,
    target_kind: value.target_kind,
    target_view_schema_id: value.target_view_schema_id,
    extension_profile_id: value.extension_profile_id,
    committed_at: value.committed_at,
  });
  const fetch = vi
    .fn()
    .mockResolvedValueOnce(
      jsonResponse({
        data: { committed_sources: [summary(source), summary(second)] },
        meta: {
          request_id: "sources",
          paging: { limit: 50, has_more: false, next_cursor: null },
        },
      }),
    )
    .mockResolvedValueOnce(jsonResponse(importTestEnvelope(second)));
  vi.stubGlobal("fetch", fetch);
  const view = render(
    <CommittedSourcesPanel
      client={client()}
      initialSourceId={null}
      recordId={ids.unit}
    />,
  );
  const buttons = await screen.findAllByRole("button", {
    name: source.original_filename,
  });
  expect(buttons).toHaveLength(2);
  expect(String(fetch.mock.calls[0]?.[0])).toContain(`record_id=${ids.unit}`);
  fireEvent.click(buttons[1] as HTMLElement);
  const download = await screen.findByRole("link", {
    name: "Download original source",
  });
  expect(download.getAttribute("href")).toBe(
    `/api/v1/incidents/${ids.incident}/committed-sources/${ids.secondUnit}/content`,
  );
  expect(view.container.querySelector("img")).toBeNull();
  expect(fetch).toHaveBeenCalledTimes(2);
  expect(screen.getByRole("heading", { name: "Source history" })).toBe(
    document.activeElement,
  );
  expect(screen.getByText(ids.actor)).toBeTruthy();
});

it("withdraws source details on rejected reads and ignores a detached observation", async () => {
  let settle!: (value: Response) => void;
  const fetch = vi
    .fn()
    .mockImplementationOnce(
      () =>
        new Promise<Response>((resolve) => {
          settle = resolve;
        }),
    )
    .mockResolvedValue(
      jsonResponse(
        {
          error: {
            code: "authorization_denied",
            details: {},
            retryable: false,
          },
          meta: { request_id: "denied" },
        },
        403,
      ),
    );
  vi.stubGlobal("fetch", fetch);
  const view = render(
    <CommittedSourcesPanel client={client()} initialSourceId={ids.unit} />,
  );
  await waitFor(() => expect(fetch).toHaveBeenCalledTimes(1));
  view.unmount();
  render(
    <CommittedSourcesPanel
      client={client()}
      initialSourceId={ids.secondUnit}
    />,
  );
  settle(jsonResponse(importTestEnvelope(source)));
  await screen.findByRole("alert");
  expect(screen.queryByText(source.original_filename)).toBeNull();
  expect(
    screen.queryByRole("link", { name: "Download original source" }),
  ).toBeNull();
  expect(
    screen.getByRole("button", { name: "Retry source read" }),
  ).toBeTruthy();
});
