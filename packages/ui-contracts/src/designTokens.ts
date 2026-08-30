import { cartularyDesignTokenVars } from "./generated/design-tokens";

export {
  type CartularyDefaultThemeId,
  type CartularyDesignTokenVarName,
  cartularyDefaultThemeId,
  cartularyDesignThemeCssText,
  cartularyDesignTokenVars,
} from "./generated/design-tokens";

export type WorkbookGridDensity = "compact" | "default" | "comfortable";

export type WorkbookGridDensityMetrics = {
  readonly cellPaddingBlockCssPx: number;
  readonly cellPaddingInlineCssPx: number;
  readonly fontSizeCssPx: number;
  readonly lineHeight: number;
  readonly rowHeightCssPx: number;
};

export function workbookGridDensityMetrics(
  density: WorkbookGridDensity,
): WorkbookGridDensityMetrics {
  const prefix = `--ct-density-${density}` as const;
  const padding = cartularyDesignTokenVars[`${prefix}-cellPadding`];
  const paddingMatch = /^(0|[1-9]\d*)px (0|[1-9]\d*)px$/.exec(padding);
  if (paddingMatch === null) {
    throw new Error(
      `Invalid grid density cell-padding token ${prefix}-cellPadding: ${padding}`,
    );
  }
  return {
    cellPaddingBlockCssPx: Number(paddingMatch[1]),
    cellPaddingInlineCssPx: Number(paddingMatch[2]),
    fontSizeCssPx: fixedDensityMetricPx(`${prefix}-fontSize`),
    lineHeight: unitlessDensityMetric(`${prefix}-lineHeight`),
    rowHeightCssPx: fixedDensityMetricPx(`${prefix}-rowHeight`),
  };
}

export function workbookGridRowHeightPx(density: WorkbookGridDensity): number {
  return workbookGridDensityMetrics(density).rowHeightCssPx;
}

function fixedDensityMetricPx(
  tokenName:
    | `--ct-density-${WorkbookGridDensity}-fontSize`
    | `--ct-density-${WorkbookGridDensity}-rowHeight`,
): number {
  const token = cartularyDesignTokenVars[tokenName];
  const match = /^(0|[1-9]\d*)px$/.exec(token);
  if (match === null) {
    throw new Error(`Invalid fixed grid density token ${tokenName}: ${token}`);
  }
  return Number(match[1]);
}

function unitlessDensityMetric(
  tokenName: `--ct-density-${WorkbookGridDensity}-lineHeight`,
): number {
  const token = cartularyDesignTokenVars[tokenName];
  if (!/^(?:0|[1-9]\d*)(?:\.\d+)?$/.test(token)) {
    throw new Error(
      `Invalid grid density line-height token ${tokenName}: ${token}`,
    );
  }
  const value = Number(token);
  if (!Number.isFinite(value) || value <= 0) {
    throw new Error(
      `Invalid grid density line-height token ${tokenName}: ${token}`,
    );
  }
  return value;
}

type FixedLayoutMetricToken =
  | "--ct-layout-baseMinHeight"
  | "--ct-layout-baseMinWidth"
  | "--ct-layout-compactMinHeight"
  | "--ct-layout-compactMinWidth"
  | "--ct-layout-inspectorDefaultWidth"
  | "--ct-layout-inspectorMinWidth"
  | "--ct-layout-narrowMinWidth";

export type WorkbookLayoutMetrics = {
  readonly baseMinHeightCssPx: number;
  readonly baseMinWidthCssPx: number;
  readonly compactMinHeightCssPx: number;
  readonly compactMinWidthCssPx: number;
  readonly inspectorDefaultWidthCssPx: number;
  readonly inspectorEffectiveMaxWidthCssPx: number;
  readonly inspectorMinWidthCssPx: number;
  readonly narrowMinWidthCssPx: number;
};

export function workbookLayoutMetrics(
  viewportWidthCssPx: number,
): WorkbookLayoutMetrics {
  const inspectorMinWidthCssPx = fixedLayoutMetricPx(
    "--ct-layout-inspectorMinWidth",
  );
  const inspectorMaximum = minPxVwLayoutMetric("--ct-layout-inspectorMaxWidth");
  const finiteViewportWidth = Number.isFinite(viewportWidthCssPx)
    ? Math.max(0, viewportWidthCssPx)
    : 0;
  const inspectorEffectiveMaxWidthCssPx = Math.max(
    inspectorMinWidthCssPx,
    Math.min(
      inspectorMaximum.maximumCssPx,
      (finiteViewportWidth * inspectorMaximum.viewportWidthPercent) / 100,
    ),
  );
  return {
    baseMinHeightCssPx: fixedLayoutMetricPx("--ct-layout-baseMinHeight"),
    baseMinWidthCssPx: fixedLayoutMetricPx("--ct-layout-baseMinWidth"),
    compactMinHeightCssPx: fixedLayoutMetricPx("--ct-layout-compactMinHeight"),
    compactMinWidthCssPx: fixedLayoutMetricPx("--ct-layout-compactMinWidth"),
    inspectorDefaultWidthCssPx: clamp(
      fixedLayoutMetricPx("--ct-layout-inspectorDefaultWidth"),
      inspectorMinWidthCssPx,
      inspectorEffectiveMaxWidthCssPx,
    ),
    inspectorEffectiveMaxWidthCssPx,
    inspectorMinWidthCssPx,
    narrowMinWidthCssPx: fixedLayoutMetricPx("--ct-layout-narrowMinWidth"),
  };
}

function fixedLayoutMetricPx(tokenName: FixedLayoutMetricToken): number {
  const token = cartularyDesignTokenVars[tokenName];
  const match = /^(0|[1-9]\d*)(?:\.(\d+))?px$/.exec(token);
  if (match === null) {
    throw new Error(`Invalid css_px_length_v1 token ${tokenName}: ${token}`);
  }
  const value = Number(token.slice(0, -2));
  if (!Number.isFinite(value) || value < 0 || value > 9999) {
    throw new Error(`Invalid css_px_length_v1 token ${tokenName}: ${token}`);
  }
  return value;
}

function minPxVwLayoutMetric(tokenName: "--ct-layout-inspectorMaxWidth"): {
  readonly maximumCssPx: number;
  readonly viewportWidthPercent: number;
} {
  const token = cartularyDesignTokenVars[tokenName];
  const match = /^min\(((?:0|[1-9]\d*)(?:\.\d+)?)px, ([1-9]\d?|100)vw\)$/.exec(
    token,
  );
  if (match === null) {
    throw new Error(`Invalid css_min_px_vw_v1 token ${tokenName}: ${token}`);
  }
  const maximumCssPx = Number(match[1]);
  const viewportWidthPercent = Number(match[2]);
  if (maximumCssPx > 9999) {
    throw new Error(`Invalid css_min_px_vw_v1 token ${tokenName}: ${token}`);
  }
  return { maximumCssPx, viewportWidthPercent };
}

function clamp(value: number, minimum: number, maximum: number): number {
  return Math.min(maximum, Math.max(minimum, value));
}
