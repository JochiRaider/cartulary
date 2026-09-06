import {
  cartularyDefaultThemeId,
  cartularyDesignThemeCssText,
} from "@cartulary/ui-contracts";
import ReactDOM from "react-dom/client";
import { NetworkFlowGridLoadFixture } from "./NetworkFlowGridLoadFixture";

const root = document.getElementById("root");
if (!root) throw new Error("missing measurement root");
ReactDOM.createRoot(root).render(
  <>
    <style>{cartularyDesignThemeCssText}</style>
    <style>{`body { margin: 0; background: var(--ct-colors-canvas); }`}</style>
    <main
      className="cartulary-shell"
      data-cartulary-theme={cartularyDefaultThemeId}
      style={{
        minHeight: "100vh",
        color: "var(--ct-colors-ink)",
        fontFamily: "var(--ct-typography-ui-fontFamily)",
        fontSize: "var(--ct-typography-ui-fontSize)",
        lineHeight: "var(--ct-typography-ui-lineHeight)",
      }}
    >
      <section
        style={{
          width: "min(78rem, 100%)",
          margin: "2rem auto",
          padding: "2rem",
        }}
      >
        <NetworkFlowGridLoadFixture />
      </section>
    </main>
  </>,
);
