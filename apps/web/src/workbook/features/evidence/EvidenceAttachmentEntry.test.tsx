import { timelineDraftEvidenceFileInputTestId } from "@cartulary/ui-contracts";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { DraftRowCreateButton } from "../../timeline/components/TimelineDraftRowActions";
import { createDraftRow } from "../../timeline/models/timelineRowModel";
import { EvidenceAttachmentEntry } from "./EvidenceAttachmentEntry";

const file = new File(["capture"], "capture.txt", { type: "text/plain" });
const props = {
  title: "original source",
  testId: "file-picker",
  disabledReason: null,
  busy: false,
};
afterEach(cleanup);

describe("Evidence attachment chooser", () => {
  it("keeps a draft invocation when the trailing draft is replaced", () => {
    const first = createDraftRow(1),
      next = createDraftRow(2);
    const attach = vi.fn(),
      create = vi.fn();
    const view = render(
      <DraftRowCreateButton
        row={first}
        onCreate={create}
        onFilesSelected={attach}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", {
        name: "Attach evidence to draft timeline row",
      }),
    );
    view.rerender(
      <DraftRowCreateButton
        row={next}
        onCreate={create}
        onFilesSelected={attach}
      />,
    );
    fireEvent.change(
      screen.getByTestId(timelineDraftEvidenceFileInputTestId()),
      { target: { files: [file] } },
    );
    expect(attach).toHaveBeenCalledTimes(1);
    expect(attach.mock.calls[0]?.[0]?.key).toBe(first.key);
    expect(create).not.toHaveBeenCalled();
  });
  it("captures the invoking callback across source replacement for both presentations", () => {
    for (const compact of [false, true]) {
      const original = vi.fn(),
        replacement = vi.fn();
      const view = render(
        <EvidenceAttachmentEntry
          {...props}
          compact={compact}
          onAttach={original}
        />,
      );
      fireEvent.click(screen.getByRole("button"));
      view.rerender(
        <EvidenceAttachmentEntry
          {...props}
          compact={compact}
          title="replacement source"
          onAttach={replacement}
        />,
      );
      fireEvent.change(screen.getByTestId(props.testId), {
        target: { files: [file] },
      });
      expect(original).toHaveBeenCalledExactlyOnceWith([file]);
      expect(replacement).not.toHaveBeenCalled();
      view.unmount();
    }
  });

  it("cancels without admission and refuses completion without an invocation", () => {
    const attach = vi.fn();
    render(<EvidenceAttachmentEntry {...props} onAttach={attach} />);
    const input = screen.getByTestId(props.testId);
    fireEvent.change(input, { target: { files: [file] } });
    expect(attach).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button"));
    fireEvent(input, new Event("cancel", { bubbles: true }));
    fireEvent.change(input, { target: { files: [file] } });
    expect(attach).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button"));
    fireEvent.change(input, { target: { files: [] } });
    expect(attach).not.toHaveBeenCalled();
  });

  it("borrows authoring focus and leaves text-only clipboard events native", () => {
    const attach = vi.fn();
    render(<EvidenceAttachmentEntry {...props} onAttach={attach} />);
    const button = screen.getByRole("button");
    expect(button.getAttribute("data-grid-editor-external-action")).toBe(
      "true",
    );
    expect(fireEvent.mouseDown(button)).toBe(false);
    const region = screen.getByRole("region");
    expect(fireEvent.paste(region, { clipboardData: { files: [] } })).toBe(
      true,
    );
    fireEvent.paste(region, { clipboardData: { files: [file] } });
    expect(attach).toHaveBeenCalledExactlyOnceWith([file]);
  });
});
