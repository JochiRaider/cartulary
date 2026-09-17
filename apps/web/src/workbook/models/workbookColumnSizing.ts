import { cartularyDesignPresentation } from "@cartulary/ui-contracts";

export const workbookColumnSizing =
  cartularyDesignPresentation.workbookColumnSizing;

export function isWorkbookColumnWidth(value: number): boolean {
  return (
    Number.isSafeInteger(value) &&
    value >= workbookColumnSizing.minimumWidthPx &&
    value <= workbookColumnSizing.maximumWidthPx
  );
}

export function parseWorkbookColumnWidth(text: string): number | null {
  if (text.trim() === "") return null;
  const value = Number(text);
  return isWorkbookColumnWidth(value) ? value : null;
}
