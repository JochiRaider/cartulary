import { useEffect } from "react";

export function useWorkbookProjectionRefreshController({
  loadAssessmentSurface,
  loadEntities,
  loadGenericSurface,
  loadSessionRole,
  sheetReloadToken,
}: {
  readonly loadAssessmentSurface: () => Promise<void>;
  readonly loadEntities: () => Promise<void>;
  readonly loadGenericSurface: () => Promise<void>;
  readonly loadSessionRole: () => Promise<void>;
  readonly sheetReloadToken: number;
}) {
  useEffect(() => {
    void Promise.all([loadEntities(), loadSessionRole()]);
  }, [loadEntities, loadSessionRole]);

  useEffect(() => {
    if (sheetReloadToken === 0) {
      return;
    }
    void loadEntities();
  }, [loadEntities, sheetReloadToken]);

  useEffect(() => {
    void sheetReloadToken;
    void loadGenericSurface();
  }, [loadGenericSurface, sheetReloadToken]);

  useEffect(() => {
    void sheetReloadToken;
    void loadAssessmentSurface();
  }, [loadAssessmentSurface, sheetReloadToken]);
}
