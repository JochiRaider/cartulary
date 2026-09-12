import { createContext } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookContextualTaskDecisionCreateOwner } from "./WorkbookContextualTaskDecisionCreateOwner";

export const ContextualCreateContext = createContext<{
  readonly owner: WorkbookContextualTaskDecisionCreateOwner;
  readonly sheetRef: SheetRef;
} | null>(null);
