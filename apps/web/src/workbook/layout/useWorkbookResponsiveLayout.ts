import { useEffect, useState } from "react";
import {
  selectWorkbookBlockMode,
  selectWorkbookChromeMode,
  type WorkbookBlockMode,
  type WorkbookChromeMode,
} from "./workbookResponsiveLayout";

type WorkbookResponsiveLayout = {
  readonly blockMode: WorkbookBlockMode;
  readonly chromeMode: WorkbookChromeMode;
};

export function currentWorkbookViewportSize(): {
  readonly height: number;
  readonly width: number;
} {
  const viewport = window.visualViewport;
  const viewportWidth = viewport?.width ?? window.innerWidth;
  const viewportHeight = viewport?.height ?? window.innerHeight;
  const rootInlineSize = document.documentElement.clientWidth;
  const rootZoom = workbookRootZoomScale();
  const zoomAdjustedWidth = viewportWidth / rootZoom;
  return {
    height: rootZoom === 1 ? viewportHeight : viewportHeight / rootZoom,
    width:
      rootZoom === 1
        ? viewportWidth
        : rootInlineSize > 0
          ? Math.min(zoomAdjustedWidth, rootInlineSize)
          : zoomAdjustedWidth,
  };
}

function workbookRootZoomScale(): number {
  const root = document.documentElement;
  const zoomValue = root.style.zoom || getComputedStyle(root).zoom;
  const parsed = Number.parseFloat(zoomValue);
  if (!Number.isFinite(parsed) || parsed <= 0) return 1;
  return zoomValue.trim().endsWith("%") ? parsed / 100 : parsed;
}

function selectWorkbookResponsiveLayout(
  widthCssPx: number,
  heightCssPx: number,
): WorkbookResponsiveLayout {
  return {
    blockMode: selectWorkbookBlockMode(heightCssPx),
    chromeMode: selectWorkbookChromeMode(widthCssPx),
  };
}

export function useWorkbookResponsiveLayout(): WorkbookResponsiveLayout {
  const [layout, setLayout] = useState<WorkbookResponsiveLayout>(() => {
    const viewport = currentWorkbookViewportSize();
    return selectWorkbookResponsiveLayout(viewport.width, viewport.height);
  });

  useEffect(() => {
    const updateLayout = () => {
      const viewport = currentWorkbookViewportSize();
      setLayout(
        selectWorkbookResponsiveLayout(viewport.width, viewport.height),
      );
    };
    const rootStyleObserver = new MutationObserver(updateLayout);
    rootStyleObserver.observe(document.documentElement, {
      attributeFilter: ["style"],
      attributes: true,
    });
    window.addEventListener("resize", updateLayout);
    window.visualViewport?.addEventListener("resize", updateLayout);
    return () => {
      window.removeEventListener("resize", updateLayout);
      window.visualViewport?.removeEventListener("resize", updateLayout);
      rootStyleObserver.disconnect();
    };
  }, []);

  return layout;
}
