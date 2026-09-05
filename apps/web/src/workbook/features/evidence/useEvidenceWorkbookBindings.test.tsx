import {
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import type { ComponentProps } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { EvidenceHandleOutcome } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { useEvidenceWorkbookBindings } from "./useEvidenceWorkbookBindings";

const row = {
  record_id: "evidence-1",
  row_version: 1,
  cells: {
    "evidence.title": { value: "Investigation screenshot" },
    "evidence.lifecycle_state": { value: "available" },
    "evidence.upload_state": { value: "available" },
  },
} as unknown as WorkbookQueryRow;
const otherRow = { ...row, record_id: "evidence-2" };
function deferred() {
  let resolve!: (outcome: EvidenceHandleOutcome) => void;
  const promise = new Promise<EvidenceHandleOutcome>((accept) => {
    resolve = accept;
  });
  return { promise, resolve };
}
const accepted = (suffix: string): EvidenceHandleOutcome => ({
  kind: "accepted",
  value: {
    href: `/api/v1/evidence-handles/${suffix}`,
    filename: "evidence.txt",
    previewKind: "text_inline",
  },
});
const rejected: EvidenceHandleOutcome = {
  kind: "rejected",
  failure: { kind: "terminal", message: "private_payload" },
};
function defaults() {
  return {
    mutationCommands: { issueHandle: vi.fn(), attach: vi.fn() },
    mutation: {
      beginMutation: vi.fn(() => vi.fn()),
    },
    onRefresh: vi.fn(),
    ownerBindings: ["evidence_lifecycle"] as const,
    resetKey: "surface-1",
    rows: [row, otherRow],
    subjectRecordId: row.record_id,
    canRead: true,
    attachDisabledReason: null,
    density: "default" as const,
    onInspect: vi.fn(),
    onRestoreFocus: vi.fn(),
  };
}
function Harness(props: Parameters<typeof useEvidenceWorkbookBindings>[0]) {
  const bindings = useEvidenceWorkbookBindings(props);
  return (
    <>
      {props.rows.map((record) => (
        <div key={record.record_id}>{bindings.renderRowActions(record)}</div>
      ))}
      {bindings.renderInspector(props.rows[0] ?? row)}
      {bindings.overlay}
      {bindings.announcements}
    </>
  );
}
function clickPreview(id = row.record_id) {
  fireEvent.click(screen.getByTestId(evidencePreviewButtonTestId(id)));
}
function clickDownload(id = row.record_id) {
  fireEvent.click(screen.getByTestId(evidenceDownloadButtonTestId(id)));
}
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("Evidence workbook bindings", () => {
  it("keeps the latest preview intent and does not claim loaded bytes on issuance", async () => {
    const first = deferred();
    const second = deferred();
    const props = defaults();
    props.mutationCommands.issueHandle
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    render(
      <Harness {...props} attachDisabledReason="This incident is closed." />,
    );
    expect(
      (
        screen.getByTestId(
          evidenceAttachFileInputTestId(row.record_id),
        ) as HTMLInputElement
      ).disabled,
    ).toBe(true);
    clickPreview();
    clickPreview(otherRow.record_id);
    expect(
      screen.getByTestId(evidenceAccessMessageTestId(row.record_id))
        .textContent,
    ).toBe("Available");
    await act(async () => second.resolve(accepted("newer")));
    expect(screen.queryByText("Preview loaded inline.")).toBeNull();
    expect(
      screen
        .getByTestId(evidencePreviewFrameTestId(otherRow.record_id))
        .getAttribute("src"),
    ).toContain("newer");
    await act(async () => first.resolve(accepted("obsolete")));
    expect(
      screen
        .getByTestId(evidencePreviewFrameTestId(otherRow.record_id))
        .getAttribute("src"),
    ).toContain("newer");
    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(screen.queryByTestId(evidencePreviewPanelTestId())).toBeNull();
    expect(props.onRestoreFocus).toHaveBeenCalledWith(otherRow.record_id);
  });

  it("discards pending preview completions after close retarget replacement deletion surface or access changes", async () => {
    const changes: Array<
      (props: ComponentProps<typeof Harness>) => ComponentProps<typeof Harness>
    > = [
      (props) => ({ ...props, subjectRecordId: otherRow.record_id }),
      (props) => ({ ...props, rows: [{ ...row, row_version: 2 }, otherRow] }),
      (props) => ({ ...props, rows: [otherRow] }),
      (props) => ({ ...props, resetKey: "surface-2" }),
      (props) => ({ ...props, canRead: false }),
    ];
    for (const change of [null, ...changes]) {
      const pending = deferred();
      const props = defaults();
      props.mutationCommands.issueHandle.mockReturnValue(pending.promise);
      const view = render(<Harness {...props} />);
      clickPreview();
      expect(screen.getByTestId(evidencePreviewPanelTestId())).toBeTruthy();
      if (change === null)
        fireEvent.click(screen.getByRole("button", { name: "Close" }));
      else view.rerender(<Harness {...change(props)} />);
      await act(async () => pending.resolve(accepted("obsolete")));
      expect(screen.queryByTestId(evidencePreviewPanelTestId())).toBeNull();
      expect(screen.queryByText("Preview opened.")).toBeNull();
      cleanup();
    }
  });

  it("orders record feedback and download effects and discards completions after unmount or access loss", async () => {
    const anchorClick = vi
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(() => undefined);
    const older = deferred();
    const newer = deferred();
    const preview = deferred();
    const props = defaults();
    props.mutationCommands.issueHandle
      .mockReturnValueOnce(older.promise)
      .mockReturnValueOnce(newer.promise)
      .mockReturnValueOnce(preview.promise);
    const view = render(<Harness {...props} />);
    clickDownload();
    clickDownload();
    clickPreview();
    await act(async () => newer.resolve(accepted("download")));
    expect(anchorClick).toHaveBeenCalledTimes(1);
    expect(
      screen.getByTestId(evidenceAccessMessageTestId(row.record_id))
        .textContent,
    ).toBe("Pending");
    await act(async () => older.resolve(accepted("obsolete")));
    expect(anchorClick).toHaveBeenCalledTimes(1);
    view.unmount();
    await act(async () => preview.resolve(accepted("unmounted")));
    const download = deferred();
    const denied = deferred();
    const secured = defaults();
    secured.mutationCommands.issueHandle
      .mockReturnValueOnce(download.promise)
      .mockReturnValueOnce(denied.promise);
    render(<Harness {...secured} />);
    clickDownload();
    clickPreview();
    await act(async () =>
      denied.resolve({
        kind: "rejected",
        failure: { kind: "authorization_lost", message: "private" },
      }),
    );
    await act(async () => download.resolve(accepted("revoked")));
    expect(anchorClick).toHaveBeenCalledTimes(1);
    expect(screen.queryByTestId(evidencePreviewPanelTestId())).toBeNull();
    expect(secured.onRefresh).toHaveBeenCalledTimes(1);
    expect(
      (
        screen.getByTestId(
          evidenceDownloadButtonTestId(row.record_id),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    cleanup();
    const retargetedDownload = deferred();
    const retargeted = defaults();
    retargeted.mutationCommands.issueHandle.mockReturnValue(
      retargetedDownload.promise,
    );
    const retargetedView = render(<Harness {...retargeted} />);
    clickDownload();
    retargetedView.rerender(
      <Harness {...retargeted} subjectRecordId={otherRow.record_id} />,
    );
    await act(async () => retargetedDownload.resolve(accepted("retargeted")));
    expect(anchorClick).toHaveBeenCalledTimes(1);
    expect(
      screen.getByTestId(evidenceAccessMessageTestId(row.record_id))
        .textContent,
    ).toBe("Available");
    cleanup();
    const invalidatedDownload = deferred();
    const invalidated = defaults();
    invalidated.mutationCommands.issueHandle.mockReturnValue(
      invalidatedDownload.promise,
    );
    const invalidatedView = render(<Harness {...invalidated} />);
    clickDownload();
    invalidatedView.rerender(
      <Harness
        {...invalidated}
        rows={[
          {
            ...row,
            cells: {
              ...row.cells,
              "evidence.upload_state": {
                ...row.cells["evidence.upload_state"],
                value: "failed",
              },
            },
          },
          otherRow,
        ]}
      />,
    );
    await act(async () =>
      invalidatedDownload.resolve(accepted("invalidated-blob")),
    );
    expect(anchorClick).toHaveBeenCalledTimes(1);
    expect(
      screen.getByTestId(evidenceAccessMessageTestId(row.record_id))
        .textContent,
    ).toBe("Inconsistent");
  });

  it("shares safe feedback across distinct row and inspector views and announces each outcome once", async () => {
    const props = defaults();
    props.mutationCommands.issueHandle.mockResolvedValue(rejected);
    render(<Harness {...props} />);
    clickPreview();
    await act(async () => undefined);
    expect(screen.getAllByRole("alert")).toHaveLength(1);
    expect(screen.getByRole("alert").textContent).toContain(
      "Investigation screenshot:",
    );
    clickPreview();
    expect(props.mutationCommands.issueHandle).toHaveBeenCalledTimes(1);
    expect(screen.queryByText("private_payload")).toBeNull();
    expect(
      screen
        .getByTestId(evidenceAccessMessageTestId(row.record_id))
        .getAttribute("aria-label"),
    ).toContain("Inspect evidence details");
    expect(
      screen.getByTestId(
        evidenceAccessMessageTestId(row.record_id, "inspector"),
      ).textContent,
    ).toContain("could not");
    fireEvent.click(
      screen.getByTestId(evidenceAccessMessageTestId(row.record_id)),
    );
    expect(props.onInspect).toHaveBeenCalledWith(row.record_id);
    expect(
      screen.getByTestId(
        evidenceAttachFileInputTestId(row.record_id, "inspector"),
      ),
    ).toBeTruthy();
  });

  it("preserves attachment settlement and suppresses stale attachment feedback", async () => {
    const pending = deferred();
    const props = defaults();
    props.mutationCommands.attach.mockReturnValue(pending.promise);
    const view = render(<Harness {...props} />);
    fireEvent.change(
      screen.getByTestId(evidenceAttachFileInputTestId(row.record_id)),
      { target: { files: [new File(["bytes"], "evidence.txt")] } },
    );
    expect(props.mutation.beginMutation).toHaveBeenCalledTimes(1);
    view.rerender(<Harness {...props} resetKey="surface-2" />);
    await act(async () => pending.resolve(rejected));
    expect(
      props.mutation.beginMutation.mock.results[0]?.value,
    ).toHaveBeenCalledOnce();
    expect(screen.getByRole("alert").textContent).toBe("");
  });
});
