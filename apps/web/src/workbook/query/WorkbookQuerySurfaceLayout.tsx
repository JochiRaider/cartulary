import { type ComponentProps, useMemo } from "react";
import { WorkbookSurfaceLayout } from "../layout/WorkbookSurfaceLayout";
import { useWorkbookBrowsingRegistry } from "./WorkbookQueryBrowsingContext";
import { WorkbookQueryBrowsingControls } from "./WorkbookQueryBrowsingControls";

type LayoutProps = ComponentProps<typeof WorkbookSurfaceLayout>;

/** Workbook query navigation is attached explicitly at the surface boundary. */
export function WorkbookQuerySurfaceLayout(
  props: Omit<
    LayoutProps,
    "presentationBinding" | "queryControls" | "workAreaOnly"
  >,
) {
  const browsing = useWorkbookBrowsingRegistry();
  const { viewSchemaId } = props;
  const presentationBinding = useMemo<
    NonNullable<LayoutProps["presentationBinding"]>
  >(
    () => ({
      update: (token, ready, detach) =>
        browsing.updatePresentation(viewSchemaId, token, ready, detach),
      unbind: (token) => browsing.unbindPresentation(viewSchemaId, token),
    }),
    [browsing, viewSchemaId],
  );
  return (
    <WorkbookSurfaceLayout
      {...props}
      presentationBinding={presentationBinding}
      queryControls={
        <WorkbookQueryBrowsingControls viewSchemaId={viewSchemaId} />
      }
    />
  );
}
