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
    void loadSessionRole();
  }, [loadSessionRole]);

  useEffect(() => {
    void sheetReloadToken;
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
