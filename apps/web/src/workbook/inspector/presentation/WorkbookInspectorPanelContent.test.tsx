import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { WorkbookInspectorPanelContent } from "./WorkbookInspectorPanelContent";

describe("Inspector panel states", () => {
  it("distinguishes loading and empty data from retained failure and concealment", () => {
    const populated = {
      kind: "populated" as const,
      content: <p>Authorized content</p>,
    };
    const { rerender } = render(
      <WorkbookInspectorPanelContent
        model={{ access: "readable", data: { state: "initial_loading" } }}
      />,
    );
    expect(screen.getByRole("status").textContent).toContain("Loading");
    rerender(
      <WorkbookInspectorPanelContent
        model={{
          access: "readable",
          data: {
            state: "ready",
            content: {
              kind: "empty",
              message: "No associations yet. Link an existing record.",
            },
          },
        }}
      />,
    );
    expect(
      screen.getByText("No associations yet. Link an existing record."),
    ).not.toBeNull();
    rerender(
      <WorkbookInspectorPanelContent
        model={{
          access: "readable",
          data: {
            state: "stale_failure",
            content: populated,
            message: "Could not refresh associations.",
          },
        }}
      />,
    );
    expect(screen.getByText("Authorized content")).not.toBeNull();
    expect(screen.getByText("Could not refresh associations.")).not.toBeNull();
    rerender(<WorkbookInspectorPanelContent model={{ access: "concealed" }} />);
    expect(screen.queryByText("Authorized content")).toBeNull();
    expect(screen.queryByText("Could not refresh associations.")).toBeNull();
    expect(
      screen.queryByText("No associations yet. Link an existing record."),
    ).toBeNull();
  });
});
