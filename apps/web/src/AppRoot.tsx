import {
  cartularyDefaultThemeId,
  cartularyDesignThemeCssText,
} from "@cartulary/ui-contracts";
import { StrictMode } from "react";

import { App } from "./App";

export function AppRoot() {
  return (
    <StrictMode>
      <style>{cartularyDesignThemeCssText}</style>
      <style>
        {`
          .cartulary-shell :where(button, input, select, textarea, a[href], [tabindex]:not([tabindex="-1"])):focus {
            outline: var(--ct-component-focus-ring-border);
            outline-offset: var(--ct-component-focus-ring-offset);
          }
        `}
      </style>
      <App themeId={cartularyDefaultThemeId} />
    </StrictMode>
  );
}
