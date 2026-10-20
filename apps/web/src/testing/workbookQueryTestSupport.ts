import {
  type RenderHookOptions,
  render,
  renderHook,
} from "@testing-library/react";
import { WorkbookQueryBrowsingProvider } from "../workbook/query/WorkbookQueryBrowsingContext";

/** Isolated query consumers use the same required owner as the workbook shell. */
export function renderWithWorkbookQueryBrowsing(
  ui: Parameters<typeof render>[0],
) {
  return render(ui, { wrapper: WorkbookQueryBrowsingProvider });
}

export function renderHookWithWorkbookQueryBrowsing<Result, Props>(
  callback: (props: Props) => Result,
  options?: Pick<RenderHookOptions<Props>, "initialProps">,
) {
  return renderHook(callback, {
    ...options,
    wrapper: WorkbookQueryBrowsingProvider,
  });
}
