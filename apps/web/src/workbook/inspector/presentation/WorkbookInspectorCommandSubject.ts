import { createContext } from "react";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";
export const WorkbookInspectorCommandSubject =
  createContext<WorkbookRecordSubject | null>(null);
