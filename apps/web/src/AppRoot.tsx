import { StrictMode } from "react";

import { App } from "./App";

export function AppRoot() {
  return (
    <StrictMode>
      <style>
        {`
          .cartulary-shell :where(button, input, select, textarea, a[href], [tabindex]:not([tabindex="-1"])):focus {
            outline: 3px solid rgb(20 83 45);
            outline-offset: 2px;
          }
        `}
      </style>
      <App />
    </StrictMode>
  );
}
