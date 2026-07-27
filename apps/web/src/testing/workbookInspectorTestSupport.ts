import {
  dataTestIdSelector,
  type EntityType,
  entityInspectorSubjectTestId,
  entityInspectorTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";
import { waitFor } from "@testing-library/react";
import {
  visibleGridRowRecordIds,
  workbookAsyncTimeoutMs,
} from "./timelineWorkbookTestSupport";

export type EntityInspectorExpectedSubject = {
  readonly entityType: EntityType;
  readonly recordId: string;
  readonly rowVersion: number;
  readonly viewSchemaId: WorkbookSurface;
};

type EntityInspectorReadiness = {
  readonly inspector: HTMLElement | null;
  readonly ready: boolean;
};

function selectedAttribute(
  element: HTMLElement | null,
  attributeName: string,
): string {
  return element?.getAttribute(attributeName) ?? "(none)";
}

function inspectEntityInspectorReadiness(
  container: HTMLElement,
  expected: EntityInspectorExpectedSubject,
): EntityInspectorReadiness {
  const inspector = container.querySelector<HTMLElement>(
    dataTestIdSelector(entityInspectorTestId(expected.entityType)),
  );
  const subject = container.querySelector<HTMLElement>(
    dataTestIdSelector(
      entityInspectorSubjectTestId(expected.entityType, expected.recordId),
    ),
  );
  const ready =
    inspector !== null &&
    subject !== null &&
    inspector.getAttribute("data-inspector-state") === "ready" &&
    inspector.getAttribute("data-view-schema-id") === expected.viewSchemaId &&
    inspector.getAttribute("data-record-id") === expected.recordId &&
    inspector.getAttribute("data-row-version") === String(expected.rowVersion);
  return { inspector, ready };
}

export function entityInspectorReadinessDiagnostic(
  container: HTMLElement,
  expected: EntityInspectorExpectedSubject,
): string {
  const { inspector } = inspectEntityInspectorReadiness(container, expected);
  const mountedRecordIds = visibleGridRowRecordIds(
    container,
    expected.viewSchemaId,
  );
  return [
    "Expected entity inspector subject",
    `view_schema_id=${expected.viewSchemaId}`,
    `entity_type=${expected.entityType}`,
    `record_id=${expected.recordId}`,
    `row_version=${expected.rowVersion}.`,
    "Mounted entity inspector subject",
    `view_schema_id=${selectedAttribute(inspector, "data-view-schema-id")}`,
    `record_id=${selectedAttribute(inspector, "data-record-id")}`,
    `row_version=${selectedAttribute(inspector, "data-row-version")}`,
    `inspector_state=${selectedAttribute(inspector, "data-inspector-state")}.`,
    `Mounted row record_ids=${mountedRecordIds.join(",") || "(none)"}.`,
    `Inspector selector=${entityInspectorTestId(expected.entityType)}`,
    `subject_selector=${entityInspectorSubjectTestId(
      expected.entityType,
      expected.recordId,
    )}.`,
  ].join(" ");
}

export async function waitForEntityInspectorReady(
  container: HTMLElement,
  expected: EntityInspectorExpectedSubject,
): Promise<HTMLElement> {
  let readyInspector: HTMLElement | null = null;
  await waitFor(
    () => {
      const readiness = inspectEntityInspectorReadiness(container, expected);
      if (!readiness.ready || readiness.inspector === null) {
        throw new Error("Entity inspector subject is not ready.");
      }
      readyInspector = readiness.inspector;
    },
    {
      onTimeout: (error) =>
        new Error(
          `${error.message}\n${entityInspectorReadinessDiagnostic(
            container,
            expected,
          )}`,
        ),
      timeout: workbookAsyncTimeoutMs,
    },
  );
  if (readyInspector === null) {
    throw new Error(entityInspectorReadinessDiagnostic(container, expected));
  }
  return readyInspector;
}
