import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { WorkbookBatchRecordChoices } from "./WorkbookBatchRecordChoices";

afterEach(cleanup);
it("Batch record choices page and search locally with at most twenty explicit review controls", () => {
  const onReview = vi.fn();
  const receipt = {
    viewSchemaId: "cartulary.view.timeline.v2",
    changeSetId: "change",
    conflicts: [],
    rows: Array.from({ length: 105 }, (_, index) => ({
      record_id: `row-${index}`,
      row_version: 2,
      cells: {
        "timeline.activity_synopsis_text": {
          value: `Incident observation ${index}`,
          display: `Incident observation ${index}`,
        },
      },
    })),
  };
  render(<WorkbookBatchRecordChoices receipt={receipt} onReview={onReview} />);
  expect(
    screen.getAllByRole("button", { name: /Review this change:/ }),
  ).toHaveLength(20);
  fireEvent.click(screen.getByRole("button", { name: "Next records" }));
  expect(
    screen.getByRole("button", {
      name: "Review this change: Incident observation 20",
    }),
  ).not.toBeNull();
  fireEvent.change(
    screen.getByRole("textbox", { name: "Find a returned record" }),
    { target: { value: "observation 104" } },
  );
  expect(
    screen.getAllByRole("button", { name: /Review this change:/ }),
  ).toHaveLength(1);
  expect(onReview).not.toHaveBeenCalled();
  fireEvent.click(
    screen.getByRole("button", {
      name: "Review this change: Incident observation 104",
    }),
  );
  expect(onReview).toHaveBeenCalledExactlyOnceWith({
    recordId: "row-104",
    label: "Incident observation 104",
  });
});
it("Batch record choices omit absent changes and non-Timeline consumers", () => {
  const receipt = {
    viewSchemaId: "cartulary.view.timeline.v2",
    changeSetId: null,
    conflicts: [],
    rows: [],
  };
  const rendered = render(
    <WorkbookBatchRecordChoices receipt={receipt} onReview={vi.fn()} />,
  );
  expect(screen.queryByRole("region")).toBeNull();
  rendered.rerender(
    <WorkbookBatchRecordChoices
      receipt={{
        ...receipt,
        viewSchemaId: "cartulary.view.hosts.v1",
        changeSetId: "entity-change",
        rows: [{ record_id: "host", row_version: 1, cells: {} }],
      }}
      onReview={vi.fn()}
    />,
  );
  expect(
    screen.queryByRole("button", { name: /Review this change/ }),
  ).toBeNull();
});
