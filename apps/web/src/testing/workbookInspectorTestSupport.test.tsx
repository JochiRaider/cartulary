import {
  entityInspectorSubjectTestId,
  entityInspectorTestId,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import { cleanup, render } from "@testing-library/react";
import { useEffect, useState } from "react";
import { afterEach, describe, expect, it } from "vitest";
import { hostsViewSchemaId } from "../workbook/models/workbookSurfaceRegistry";
import {
  entityInspectorReadinessDiagnostic,
  waitForEntityInspectorReady,
} from "./workbookInspectorTestSupport";

afterEach(() => {
  cleanup();
});

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
        >
          <span data-testid={entityInspectorSubjectTestId("host", recordId)} />
        </aside>
      ) : null}
    </section>
  );
}

describe("workbook inspector test support", () => {
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
