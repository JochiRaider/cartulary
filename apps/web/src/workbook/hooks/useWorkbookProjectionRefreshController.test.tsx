import { render, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useWorkbookProjectionRefreshController } from "./useWorkbookProjectionRefreshController";

function ProjectionRefreshHarness({
  loadAssessmentSurface,
  loadEntities,
  loadGenericSurface,
  loadSessionRole,
  sheetReloadToken,
}: Parameters<typeof useWorkbookProjectionRefreshController>[0]) {
  useWorkbookProjectionRefreshController({
    loadAssessmentSurface,
    loadEntities,
    loadGenericSurface,
    loadSessionRole,
    sheetReloadToken,
  });
  return null;
}

describe("useWorkbookProjectionRefreshController", () => {
  it("separates initial session/entity loading from sheet projection refresh", async () => {
    const loadAssessmentSurface = vi.fn(async () => undefined);
    const loadEntities = vi.fn(async () => undefined);
    const loadGenericSurface = vi.fn(async () => undefined);
    const loadSessionRole = vi.fn(async () => undefined);
    const view = render(
      <ProjectionRefreshHarness
        loadAssessmentSurface={loadAssessmentSurface}
        loadEntities={loadEntities}
        loadGenericSurface={loadGenericSurface}
        loadSessionRole={loadSessionRole}
        sheetReloadToken={0}
      />,
    );
    await waitFor(() => {
      expect(loadEntities).toHaveBeenCalledTimes(1);
      expect(loadSessionRole).toHaveBeenCalledTimes(1);
      expect(loadGenericSurface).toHaveBeenCalledTimes(1);
      expect(loadAssessmentSurface).toHaveBeenCalledTimes(1);
    });

    view.rerender(
      <ProjectionRefreshHarness
        loadAssessmentSurface={loadAssessmentSurface}
        loadEntities={loadEntities}
        loadGenericSurface={loadGenericSurface}
        loadSessionRole={loadSessionRole}
        sheetReloadToken={1}
      />,
    );
    await waitFor(() => {
      expect(loadEntities).toHaveBeenCalledTimes(2);
      expect(loadSessionRole).toHaveBeenCalledTimes(1);
      expect(loadGenericSurface).toHaveBeenCalledTimes(2);
      expect(loadAssessmentSurface).toHaveBeenCalledTimes(2);
    });
  });
});
