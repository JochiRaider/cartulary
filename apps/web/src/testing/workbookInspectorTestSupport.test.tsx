import {
  entityInspectorTestId,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import { cleanup, render } from "@testing-library/react";
import { useEffect, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "../workbook/models/workbookSurfaceRegistry";
import { workbookAsyncTimeoutMs } from "./timelineWorkbookTestSupport";
import {
  entityInspectorReadinessDiagnostic,
  waitForEntityInspectorReady,
} from "./workbookInspectorTestSupport";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.useRealTimers();
});

const expectedSubject = {
  entityType: "host",
  recordId: "host-expected",
  rowVersion: 7,
  viewSchemaId: hostsViewSchemaId,
} as const;

function mountedSubject() {
  const { container, getByRole } = render(
    <section>
      <input
        aria-label="Unrelated editor"
        defaultValue="private-authoring-value"
      />
      <aside
        data-inspector-state="ready"
        data-record-id={expectedSubject.recordId}
        data-row-version={expectedSubject.rowVersion}
        data-testid={entityInspectorTestId("host")}
        data-view-schema-id={expectedSubject.viewSchemaId}
      >
        private-record-payload
      </aside>
    </section>,
  );
  return {
    container,
    inspector: getByRole("complementary"),
    editor: getByRole("textbox"),
  };
}

function EntityInspectorFixture({
  readyInitially = false,
  recordId,
  rowVersion,
}: {
  readonly readyInitially?: boolean;
  readonly recordId: string;
  readonly rowVersion: number;
}) {
  const [ready, setReady] = useState(readyInitially);
  useEffect(() => {
    if (readyInitially) {
      return;
    }
    const handle = window.setTimeout(() => setReady(true), 20);
    return () => window.clearTimeout(handle);
  }, [readyInitially]);
  return (
    <section>
      <table data-testid={gridShellTestId(hostsViewSchemaId)}>
        <tbody>
          {/* biome-ignore lint/a11y/noRedundantRoles: shared grid selectors intentionally target explicit row roles. */}
          <tr data-grid-record-id={recordId} role="row">
            <td />
          </tr>
        </tbody>
      </table>
      {ready ? (
        <aside
          data-inspector-state="ready"
          data-record-id={recordId}
          data-row-version={rowVersion}
          data-testid={entityInspectorTestId("host")}
          data-view-schema-id={hostsViewSchemaId}
        />
      ) : null}
    </section>
  );
}

describe("workbook inspector test support", () => {
  it("waits for the expected record despite identical inspector text", async () => {
    vi.useFakeTimers();
    const { container, inspector } = mountedSubject();
    inspector.setAttribute("data-record-id", "host-other");
    const ready = vi.fn();
    const observation = waitForEntityInspectorReady(
      container,
      expectedSubject,
    ).then(ready);
    await vi.advanceTimersByTimeAsync(50);
    expect(ready).not.toHaveBeenCalled();
    inspector.setAttribute("data-record-id", expectedSubject.recordId);
    await vi.advanceTimersByTimeAsync(50);
    await observation;
    expect(ready).toHaveBeenCalledExactlyOnceWith(inspector);
  });

  it("waits for the expected inspector surface despite matching record and version", async () => {
    vi.useFakeTimers();
    const { container, inspector } = mountedSubject();
    inspector.setAttribute("data-view-schema-id", timelineViewSchemaId);
    const ready = vi.fn();
    const observation = waitForEntityInspectorReady(
      container,
      expectedSubject,
    ).then(ready);
    await vi.advanceTimersByTimeAsync(50);
    expect(ready).not.toHaveBeenCalled();
    inspector.setAttribute("data-view-schema-id", hostsViewSchemaId);
    await vi.advanceTimersByTimeAsync(50);
    await observation;
    expect(ready).toHaveBeenCalledExactlyOnceWith(inspector);
  });

  it("requires the exact inspector version rather than a version floor", async () => {
    vi.useFakeTimers();
    const { container, inspector } = mountedSubject();
    inspector.setAttribute("data-row-version", "6");
    const ready = vi.fn();
    const observation = waitForEntityInspectorReady(
      container,
      expectedSubject,
    ).then(ready);
    await vi.advanceTimersByTimeAsync(50);
    expect(ready).not.toHaveBeenCalled();
    inspector.setAttribute("data-row-version", "8");
    await vi.advanceTimersByTimeAsync(50);
    expect(ready).not.toHaveBeenCalled();
    inspector.setAttribute("data-row-version", "7");
    await vi.advanceTimersByTimeAsync(50);
    await observation;
    expect(ready).toHaveBeenCalledExactlyOnceWith(inspector);
  });

  it("times out a non-ready subject with bounded payload-free diagnostics", async () => {
    vi.useFakeTimers();
    const { container, inspector } = mountedSubject();
    inspector.setAttribute("data-inspector-state", "loading");
    const ready = vi.fn();
    const rejected = vi.fn();
    const observation = waitForEntityInspectorReady(
      container,
      expectedSubject,
    ).then(ready, rejected);
    await vi.advanceTimersByTimeAsync(workbookAsyncTimeoutMs - 1);
    expect(ready).not.toHaveBeenCalled();
    expect(rejected).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    await observation;
    expect(rejected).toHaveBeenCalledOnce();
    const diagnostic = String(rejected.mock.calls[0]?.[0]);
    expect(diagnostic).toContain("record_id=host-expected");
    expect(diagnostic).toContain("row_version=7");
    expect(diagnostic).toContain("inspector_state=loading");
    expect(diagnostic).not.toContain("private-record-payload");
    expect(diagnostic).not.toContain("private-authoring-value");
  });

  it("observes readiness without changing focus or dispatching actions", async () => {
    vi.useFakeTimers();
    const { container, inspector, editor } = mountedSubject();
    editor.focus();
    inspector.setAttribute("data-inspector-state", "loading");
    const focus = vi.spyOn(HTMLElement.prototype, "focus");
    const click = vi.spyOn(HTMLElement.prototype, "click");
    const dispatch = vi.spyOn(EventTarget.prototype, "dispatchEvent");
    const observation = waitForEntityInspectorReady(container, expectedSubject);
    await vi.advanceTimersByTimeAsync(50);
    inspector.setAttribute("data-inspector-state", "ready");
    await vi.advanceTimersByTimeAsync(50);
    await observation;
    expect(document.activeElement).toBe(editor);
    expect(focus).not.toHaveBeenCalled();
    expect(click).not.toHaveBeenCalled();
    expect(dispatch).not.toHaveBeenCalled();
  });

  it("waits for delayed entity inspector readiness by stable subject identity", async () => {
    const { container } = render(
      <EntityInspectorFixture recordId="host-delayed" rowVersion={7} />,
    );

    await expect(
      waitForEntityInspectorReady(container, {
        entityType: "host",
        recordId: "host-delayed",
        rowVersion: 7,
        viewSchemaId: hostsViewSchemaId,
      }),
    ).resolves.toBeTruthy();
  });

  it("diagnoses expected and mounted inspector subjects without row payloads", () => {
    const { container } = render(
      <EntityInspectorFixture
        readyInitially
        recordId="host-mounted"
        rowVersion={8}
      />,
    );
    const diagnostic = entityInspectorReadinessDiagnostic(container, {
      entityType: "host",
      recordId: "host-expected",
      rowVersion: 9,
      viewSchemaId: hostsViewSchemaId,
    });

    expect(diagnostic).toContain("record_id=host-expected");
    expect(diagnostic).toContain("row_version=9");
    expect(diagnostic).toContain("record_id=host-mounted");
    expect(diagnostic).toContain("row_version=8");
    expect(diagnostic).toContain("Mounted row record_ids=host-mounted");
    expect(diagnostic).not.toContain("incident");
  });
});
