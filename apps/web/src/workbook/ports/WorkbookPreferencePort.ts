import type { SheetRef } from "../../shared/sheetRef";
import type {
  DefaultPreference,
  HomePreference,
  PreferenceResult,
} from "../preferences/workbookPreferenceModel";

export interface WorkbookPreferencePort {
  readHome(input: {
    readonly signal: AbortSignal;
  }): Promise<PreferenceResult<HomePreference>>;
  readDefault(input: {
    readonly signal: AbortSignal;
  }): Promise<PreferenceResult<DefaultPreference>>;
  setDefaultSheet(input: {
    readonly sheetRef: SheetRef | null;
    readonly signal: AbortSignal;
  }): Promise<PreferenceResult<DefaultPreference>>;
  setHomeSheet(input: {
    readonly sheetRef: SheetRef | null;
    readonly signal: AbortSignal;
  }): Promise<PreferenceResult<HomePreference>>;
}
