import { createContext, useContext } from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type {
  WorkbookSessionNavigation,
  WorkbookSessionPin,
} from "./WorkbookSessionNavigation";

export type WorkbookNavigationTarget = {
  readonly sheetRef: SheetRef;
  readonly recordId?: string;
  readonly fieldKey?: string;
};
export type WorkbookInspectValue = {
  readonly revision: number;
  readonly viewSchemaId: string;
  readonly recordId: string;
  readonly fieldKey: string;
};
export type WorkbookNavigationActions = {
  readonly open: (target: WorkbookNavigationTarget, inspect?: boolean) => void;
};
export type WorkbookWorkbench = WorkbookNavigationActions & {
  readonly navigationReady: boolean;
  readonly inspectValue: WorkbookInspectValue | null;
  readonly requestInspectValue: () => boolean;
  readonly acknowledgeInspectValue: (revision: number) => void;
  readonly session: WorkbookSessionNavigation;
  readonly openPin: (pin: WorkbookSessionPin) => void;
  readonly pinCurrentView: () => void;
  readonly pinRecord: (
    viewSchemaId: string,
    recordId: string,
    label: string,
    fieldKey?: string,
  ) => void;
  readonly returnToOrigin: (inspect?: boolean) => void;
  readonly registerInspector: (view: string, handler: () => void) => () => void;
  readonly registerInspectorFocus: (
    view: string,
    recordId: string,
    focus: () => boolean,
  ) => () => void;
  readonly pinViewLabel: "Pin view" | "Pin base surface";
  readonly cancelNavigation: () => void;
  readonly message: string | null;
  readonly retry: (() => void) | null;
  readonly openBase: (() => void) | null;
};
export const WorkbookWorkbenchContext = createContext<WorkbookWorkbench | null>(
  null,
);
export const useWorkbookWorkbench = () => useContext(WorkbookWorkbenchContext);
