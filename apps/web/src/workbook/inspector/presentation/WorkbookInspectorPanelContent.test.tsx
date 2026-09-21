import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  type WorkbookInspectorNotice,
  WorkbookInspectorNoticeLedger,
} from "../workbookInspectorErrorModel";
import { WorkbookInspectorNoticeView } from "./WorkbookInspectorFeedback";
import { WorkbookInspectorRegionContent } from "./WorkbookInspectorPanelContent";

describe("Inspector panel states", () => {
  it("announces each owner transition once across remounts and never carries feedback across subjects", () => {
    const ledger = new WorkbookInspectorNoticeLedger();
    const notice: WorkbookInspectorNotice = {
      context: {
        authority: "actor/incident/1",
        subject: {
          kind: "record",
          viewSchemaId: "cartulary.view.notes.v1",
          recordId: "A",
        },
      },
      destination: {
        kind: "region",
        panel: "relationships",
        regionId: "sources",
      },
      attemptId: "read-1",
      transitionId: "failed-1",
      announcement: "assertive",
      feedback: {
        kind: "message",
        message: "Retry this read.",
        announcement: "none",
      },
    };
    const view = render(
      <WorkbookInspectorNoticeView
        notice={notice}
        consume={ledger.consume}
        visible={false}
      />,
    );
    expect(screen.getByRole("alert").textContent).toBe("Retry this read.");
    view.unmount();
    const attached = render(
      <WorkbookInspectorNoticeView
        notice={notice}
        consume={ledger.consume}
        visible={false}
      />,
    );
    expect(screen.queryByRole("alert")).toBeNull();
    attached.rerender(
      <WorkbookInspectorNoticeView
        notice={{ ...notice, transitionId: "failed-2" }}
        consume={ledger.consume}
        visible={false}
      />,
    );
    expect(screen.getByRole("alert").textContent).toBe("Retry this read.");
    attached.rerender(
      <WorkbookInspectorNoticeView
        notice={{
          ...notice,
          context: {
            ...notice.context,
            subject: {
              ...notice.context.subject,
              kind: "record",
              viewSchemaId: "cartulary.view.notes.v1",
              recordId: "B",
            },
          },
          announcement: "none",
        }}
        consume={ledger.consume}
        visible={false}
      />,
    );
    expect(screen.queryByRole("alert")).toBeNull();
    expect(screen.queryByText("Retry this read.")).toBeNull();
    attached.unmount();
  });
  it("distinguishes loading and empty data from retained failure and concealment", () => {
    const populated = {
      kind: "populated" as const,
      content: <p>Authorized content</p>,
    };
    const { rerender } = render(
      <WorkbookInspectorRegionContent
        model={{ access: "readable", data: { state: "initial_loading" } }}
      />,
    );
    expect(screen.getByRole("status").textContent).toContain("Loading");
    rerender(
      <WorkbookInspectorRegionContent
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
      <WorkbookInspectorRegionContent
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
    rerender(
      <WorkbookInspectorRegionContent model={{ access: "concealed" }} />,
    );
    expect(screen.queryByText("Authorized content")).toBeNull();
    expect(screen.queryByText("Could not refresh associations.")).toBeNull();
    expect(
      screen.queryByText("No associations yet. Link an existing record."),
    ).toBeNull();
  });
});
