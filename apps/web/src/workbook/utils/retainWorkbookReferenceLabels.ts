/** Retain presentation only for feature-supplied identities, never visited pages. */
export function retainWorkbookReferenceLabels(
  identities: readonly string[],
  labels: Readonly<Record<string, string>>,
): Readonly<Record<string, string>> {
  return Object.fromEntries(
    [...new Set(identities)].flatMap((id) =>
      Object.hasOwn(labels, id) ? [[id, labels[id] as string]] : [],
    ),
  );
}
