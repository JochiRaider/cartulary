import { type SheetRef, sheetRefsEqual } from "../../shared/sheetRef";
import type {
  WorkbookQueryState,
  WorkbookSavedViewLayoutJson,
} from "../models/workbookQuery";
import { savedViewJSONEqual } from "../models/workbookSavedViews";

type WorkbookSheet = Exclude<SheetRef, { kind: "extension_workspace" }>;
type ExtensionSheet = Extract<SheetRef, { kind: "extension_workspace" }>;
type Anchor = { readonly recordId: string; readonly fieldKey?: string } | Root;
type Root = { readonly recordId?: never; readonly fieldKey?: never };

export type WorkbookNavigationTarget =
  | ({ readonly sheetRef: WorkbookSheet } & Anchor)
  | ({ readonly sheetRef: SheetRef } & Root);

export type WorkbookReturnOrigin = {
  readonly incidentId: string;
  readonly invoker: "grid" | "view" | "work";
} & (
  | ({
      readonly sheetRef: WorkbookSheet;
      readonly viewSchemaId: string;
      readonly query: WorkbookQueryState;
      readonly layout: WorkbookSavedViewLayoutJson;
      readonly savedViewVersion?: number;
    } & Anchor)
  | ({
      readonly sheetRef: ExtensionSheet;
      readonly viewSchemaId?: never;
      readonly query?: never;
      readonly layout?: never;
      readonly savedViewVersion?: never;
    } & Root)
);

export type WorkbookSessionPin = {
  readonly incidentId: string;
  readonly label: string;
} & (
  | ({ readonly sheetRef: Extract<SheetRef, { kind: "view_schema" }> } & Anchor)
  | ({ readonly sheetRef: SheetRef } & Root)
);

export type WorkbookNavigationIntent = {
  readonly target: WorkbookNavigationTarget;
  readonly entry: "open" | "pin" | "return" | "base_record" | "base_surface";
  readonly inspect: boolean;
  readonly returning?: WorkbookReturnOrigin;
};

export function workbookNavigationIntentsEqual(
  a: WorkbookNavigationIntent,
  b: WorkbookNavigationIntent,
): boolean {
  const x = a.returning;
  const y = b.returning;
  return (
    a.entry === b.entry &&
    a.inspect === b.inspect &&
    sheetRefsEqual(a.target.sheetRef, b.target.sheetRef) &&
    a.target.recordId === b.target.recordId &&
    a.target.fieldKey === b.target.fieldKey &&
    (x === undefined || y === undefined
      ? x === y
      : x.incidentId === y.incidentId &&
        sheetRefsEqual(x.sheetRef, y.sheetRef) &&
        x.viewSchemaId === y.viewSchemaId &&
        x.recordId === y.recordId &&
        x.fieldKey === y.fieldKey &&
        x.savedViewVersion === y.savedViewVersion &&
        x.invoker === y.invoker &&
        savedViewJSONEqual(x.query, y.query) &&
        savedViewJSONEqual(x.layout, y.layout))
  );
}
