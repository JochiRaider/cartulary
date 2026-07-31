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

function currentViewportSize(): {
  readonly height: number;
  readonly width: number;
} {
  const viewport = window.visualViewport;
  if (!viewport) {
    return { height: 720, width: 1280 };
  }
  return {
    height: viewport.height,
    width: viewport.width,
  };
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
    const viewport = currentViewportSize();
    return selectWorkbookResponsiveLayout(viewport.width, viewport.height);
  });

  useEffect(() => {
    const updateLayout = () => {
      const viewport = currentViewportSize();
      setLayout(
        selectWorkbookResponsiveLayout(viewport.width, viewport.height),
      );
    };
    window.addEventListener("resize", updateLayout);
    window.visualViewport?.addEventListener("resize", updateLayout);
    return () => {
      window.removeEventListener("resize", updateLayout);
      window.visualViewport?.removeEventListener("resize", updateLayout);
    };
  }, []);

  return layout;
}
