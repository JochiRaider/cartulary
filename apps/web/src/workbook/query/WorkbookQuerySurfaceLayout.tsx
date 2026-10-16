import { type ComponentProps, useMemo } from "react";
import { WorkbookSurfaceLayout } from "../layout/WorkbookSurfaceLayout";
import { useWorkbookQueryPresentation } from "./WorkbookQueryBrowsingContext";
import { WorkbookQueryBrowsingControls } from "./WorkbookQueryBrowsingControls";

type LayoutProps = ComponentProps<typeof WorkbookSurfaceLayout>;

/** Workbook query navigation is attached explicitly at the surface boundary. */
export function WorkbookQuerySurfaceLayout(
  props: Omit<
    LayoutProps,
    | "presentationBinding"
    | "queryControls"
    | "workAreaOnly"
    | "queryObservation"
  >,
) {
  const browsing = useWorkbookQueryPresentation();
  const { viewSchemaId } = props;
  const query = browsing.find(viewSchemaId)?.getSnapshot();
  const presentationBinding = useMemo<
    NonNullable<LayoutProps["presentationBinding"]>
  >(
    () => ({
      bind: (token) =>
        browsing.bindPresentation(viewSchemaId, token, false, () => {}),
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
      queryObservation={{
        acceptedRevision: query?.resultRevision ?? 0,
        filterCount: query?.authored?.filters.length ?? 0,
        sortCount: query?.authored?.sort.length ?? 0,
        groupField: query?.authored?.groupBy ?? "",
        pending: query?.pending !== null || !query?.accepted,
      }}
      queryControls={
        <WorkbookQueryBrowsingControls viewSchemaId={viewSchemaId} />
      }
    />
  );
}
