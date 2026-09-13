import { createContext } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookNoteCreateOwner } from "./WorkbookNoteCreateOwner";
export const NoteCreateContext = createContext<{
  readonly owner: WorkbookNoteCreateOwner;
  readonly sheetRef: SheetRef;
} | null>(null);

export const noteSheetAttachment = Symbol("note-sheet");
