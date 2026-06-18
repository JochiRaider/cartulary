import {
  cartularyDefaultThemeId,
  cartularyDesignThemeCssText,
} from "@cartulary/ui-contracts";
import { StrictMode } from "react";

import { App, type CartularyReadingProfile } from "./App";

type AppRootProps = {
  readonly readingProfile?: CartularyReadingProfile | undefined;
};

export function AppRoot({ readingProfile = "default" }: AppRootProps = {}) {
  return (
    <StrictMode>
      <style>{cartularyDesignThemeCssText}</style>
      <style>
        {`
          :root {
            --ct-app-viewport-block-size: 100vh;
          }

          @supports (height: 100dvh) {
            :root {
              --ct-app-viewport-block-size: 100dvh;
            }
          }

          html,
          body,
          #root {
            block-size: 100%;
            min-height: 100%;
            margin: 0;
          }

          body {
            background: var(--ct-colors-canvas);
          }

          #root {
            min-height: var(--ct-app-viewport-block-size);
          }

          .cartulary-grid :where(button, input, select, textarea, a[href], [tabindex]:not([tabindex="-1"])) {
            scroll-margin-block-start: var(--cartulary-grid-scroll-margin-block-start);
            scroll-margin-block-end: var(--cartulary-grid-scroll-margin-block-end);
          }

          .cartulary-shell :where(button, input, select, textarea, a[href], [tabindex]:not([tabindex="-1"])):focus {
            outline: var(--ct-component-focus-ring-border);
            outline-offset: var(--ct-component-focus-ring-offset);
          }
        `}
      </style>
      <App readingProfile={readingProfile} themeId={cartularyDefaultThemeId} />
    </StrictMode>
  );
}
