import type { CollectionActionsV1 } from "@cartulary/protocol-ts/http";
export function tagActionsPayload(
  tagNames: readonly [string, ...string[]],
): CollectionActionsV1 {
  const [firstTagName, ...remainingTagNames] = tagNames;
  return {
    kind: "collection_actions_v1",
    actions: [
      { op: "add_tag", tag_name: firstTagName },
      ...remainingTagNames.map((tagName) => ({
        op: "add_tag" as const,
        tag_name: tagName,
      })),
    ],
  };
}
