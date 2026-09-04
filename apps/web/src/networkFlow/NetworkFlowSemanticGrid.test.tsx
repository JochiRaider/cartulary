import {
  networkAnalysisColumnActionTestId,
  networkAnalysisRowCellTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { NetworkFlowAcceptedGrid } from "./NetworkFlowSemanticGrid";
import type { NetworkFlowRow } from "./networkFlowClient";
import { NetworkFlowRequestError } from "./networkFlowErrors";

describe("NetworkFlowSemanticGrid accessibility", () => {
  afterEach(cleanup);

  it("restores semantic focus to the same field in the nearest replacement row and then the grid root", async () => {
    const initialRows = [flowRow("nfr_1", 1), flowRow("nfr_2", 2)];
    const { rerender } = renderGrid(initialRows);
    const selectedCell = semanticCell("nfr_2", "network_flow.src_ip");
    fireEvent.click(selectedCell);
    expect(
      await screen.findByTestId(networkAnalysisTestId("inspector")),
    ).toBeTruthy();

    const closeInspector = screen.getByTestId(
      networkAnalysisTestId("inspector-close"),
    );
    closeInspector.focus();
    fireEvent.click(closeInspector);
    await waitFor(() => expect(document.activeElement).toBe(selectedCell));
    expect(screen.queryByTestId(networkAnalysisTestId("inspector"))).toBeNull();

    const replacementRows = [flowRow("nfr_3", 3), flowRow("nfr_4", 4)];
    rerender(grid(replacementRows));
    const replacementCell = semanticCell("nfr_4", "network_flow.src_ip");
    await waitFor(() => expect(document.activeElement).toBe(replacementCell));
    expect(screen.queryByTestId(networkAnalysisTestId("inspector"))).toBeNull();

    rerender(grid(replacementRows, "filtered-query"));
    await waitFor(() =>
      expect(document.activeElement).toBe(screen.getByRole("grid")),
    );
  });

  it("supports keyboard column reordering with a polite semantic announcement", async () => {
    const user = userEvent.setup();
    renderGrid([flowRow("nfr_1", 1)]);
    await user.click(screen.getByText("Columns"));
    const moveSourceEarlier = screen.getByTestId(
      networkAnalysisColumnActionTestId("network_flow.src_ip", "move-earlier"),
    );
    moveSourceEarlier.focus();
    await user.keyboard("{Enter}");

    const headers = screen
      .getAllByRole("columnheader")
      .map((header) => header.textContent ?? "");
    expect(headers.indexOf("Source IP")).toBeLessThan(
      headers.indexOf("Flow end"),
    );
    expect(screen.getByText("Source IP column moved earlier.")).toBeTruthy();
    await user.click(screen.getByTestId(networkAnalysisTestId("layout-reset")));
  });

  it("maps typed protected-state loss to permission denial without retaining rows or inventing Retry", () => {
    const protectedError = new NetworkFlowRequestError({
      code: "authorization_denied",
      retryAction: "do_not_retry",
      retryable: false,
      safeMessage: "You no longer have access to this incident.",
      status: 403,
    });
    render(
      <div style={{ blockSize: 480, inlineSize: 1200 }}>
        <NetworkFlowAcceptedGrid
          error={protectedError}
          filtered={false}
          loadState="error"
          resetKey="access-lost"
          rows={[flowRow("protected-row", 1)]}
          sort={[]}
          onResetQuery={() => undefined}
          onRetry={() => undefined}
          onSelectionChange={() => undefined}
          onSortChange={() => undefined}
        />
      </div>,
    );

    expect(screen.getByRole("alert").textContent).toContain(
      "You no longer have access to this incident.",
    );
    expect(
      screen.queryByTestId(
        networkAnalysisRowCellTestId("protected-row", "network_flow.src_ip"),
      ),
    ).toBeNull();
    expect(screen.queryByRole("button", { name: "Retry" })).toBeNull();
  });

  it("retains authorized rows for a retryable stale refresh failure", async () => {
    const onRetry = vi.fn();
    const refreshError = new NetworkFlowRequestError({
      code: "network_flow_query_failed",
      retryAction: "retry_with_backoff",
      retryable: true,
      safeMessage: "Network Flow refresh failed.",
      status: 503,
    });
    render(
      <div style={{ blockSize: 480, inlineSize: 1200 }}>
        <NetworkFlowAcceptedGrid
          error={refreshError}
          filtered={false}
          loadState="error"
          resetKey="stale-refresh"
          rows={[flowRow("retained-row", 1)]}
          sort={[]}
          onResetQuery={() => undefined}
          onRetry={onRetry}
          onSelectionChange={() => undefined}
          onSortChange={() => undefined}
        />
      </div>,
    );

    expect(
      screen.getByTestId(
        networkAnalysisRowCellTestId("retained-row", "network_flow.src_ip"),
      ),
    ).toBeTruthy();
    expect(screen.getByRole("alert").textContent).toContain(
      "Network Flow refresh failed. Previously loaded rows may be stale.",
    );
    await userEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(onRetry).toHaveBeenCalledOnce();
  });
});

function renderGrid(rows: readonly NetworkFlowRow[]) {
  return render(grid(rows));
}

function grid(rows: readonly NetworkFlowRow[], resetKey = "stable-query") {
  return (
    <div style={{ blockSize: 480, inlineSize: 1200 }}>
      <NetworkFlowAcceptedGrid
        error={null}
        filtered={false}
        loadState="ready"
        resetKey={resetKey}
        rows={rows}
        sort={[]}
        onResetQuery={() => undefined}
        onRetry={() => undefined}
        onSelectionChange={() => undefined}
        onSortChange={() => undefined}
      />
    </div>
  );
}

function semanticCell(rowId: string, fieldKey: string): HTMLElement {
  const content = screen.getByTestId(
    networkAnalysisRowCellTestId(rowId, fieldKey),
  );
  const cell = content.closest('[role="gridcell"]');
  if (!(cell instanceof HTMLElement)) {
    throw new Error(`Missing semantic grid cell for ${rowId}:${fieldKey}`);
  }
  return cell;
}

function flowRow(rowId: string, sourceRowNumber: number): NetworkFlowRow {
  return {
    network_flow_row_id: rowId,
    network_flow_table_id: "nft_1",
    source_row_number: sourceRowNumber,
    "network_flow.flow_start_utc": "2026-07-16T00:00:00Z",
    "network_flow.flow_end_utc": "2026-07-16T00:01:00Z",
    "network_flow.src_ip": `192.0.2.${sourceRowNumber}`,
    "network_flow.src_port": 443,
    "network_flow.dst_ip": "203.0.113.20",
    "network_flow.dst_port": 8443,
    "network_flow.ip_protocol": 6,
    "network_flow.bytes_count": "100",
    "network_flow.packets_count": "2",
    "network_flow.input_interface": null,
    "network_flow.output_interface": null,
  } as NetworkFlowRow;
}
