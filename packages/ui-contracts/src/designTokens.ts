import { cartularyDesignTokenVars } from "./generated/design-tokens";

export {
  type CartularyDefaultThemeId,
  type CartularyDesignTokenVarName,
  cartularyDefaultThemeId,
  cartularyDesignThemeCssText,
  cartularyDesignTokenVars,
} from "./generated/design-tokens";

export type WorkbookGridDensity = "compact" | "default" | "comfortable";

export function workbookGridRowHeightPx(density: WorkbookGridDensity): number {
  const tokenName = `--ct-density-${density}-rowHeight` as const;
  const token = cartularyDesignTokenVars[tokenName];
  const match = /^(\d+)px$/.exec(token);
  if (match === null) {
    throw new Error(
      `Invalid fixed grid row-height token ${tokenName}: ${token}`,
    );
  }
  return Number(match[1]);
}
