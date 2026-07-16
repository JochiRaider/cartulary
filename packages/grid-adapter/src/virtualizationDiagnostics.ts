let virtualizationDisabledForDiagnostics = false;

export function isGridVirtualizationEnabled(): boolean {
  return !virtualizationDisabledForDiagnostics;
}

export function setGridVirtualizationDisabledForDiagnostics(
  disabled: boolean,
): void {
  virtualizationDisabledForDiagnostics = disabled;
}
