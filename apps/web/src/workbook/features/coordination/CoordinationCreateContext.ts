import { createContext } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookCoordinationCreateOwner } from "./WorkbookCoordinationCreateOwner";
export const CoordinationCreateContext = createContext<{
  readonly owner: WorkbookCoordinationCreateOwner;
  readonly sheetRef: SheetRef;
} | null>(null);
