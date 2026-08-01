import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookPortResult } from "./WorkbookPortResult";

export interface WorkbookPreferencePort {
  setDefaultSheet(input: {
    readonly sheetRef: SheetRef;
    readonly signal: AbortSignal;
  }): Promise<WorkbookPortResult<void>>;
  setHomeSheet(input: {
    readonly sheetRef: SheetRef;
    readonly signal: AbortSignal;
  }): Promise<WorkbookPortResult<void>>;
}
