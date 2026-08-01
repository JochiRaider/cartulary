import type { WorkbookSheetRef } from "../models/workbookStartup";
import type { WorkbookPortResult } from "./WorkbookPortResult";

export interface WorkbookPreferencePort {
  setDefaultSheet(input: {
    readonly sheetRef: WorkbookSheetRef;
    readonly signal: AbortSignal;
  }): Promise<WorkbookPortResult<void>>;
  setHomeSheet(input: {
    readonly sheetRef: WorkbookSheetRef;
    readonly signal: AbortSignal;
  }): Promise<WorkbookPortResult<void>>;
}
