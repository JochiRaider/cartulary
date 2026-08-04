export const testTimelineViewSchemaId = "cartulary.view.timeline.v2";

export function createBrowserPage(
  elements:
    | Record<string, Element | undefined>
    | (() => Record<string, Element | undefined>),
  options: {
    isVisible?: (
      testId: string,
      element: Element,
    ) => boolean | Promise<boolean>;
    onEvaluate?: (testId: string, element: Element) => void;
  } = {},
) {
  const resolveElement = (value: string) => {
    const resolvedElements =
      typeof elements === "function" ? elements() : elements;
    return resolvedElements[value];
  };
  return {
    getByTestId(value: string) {
      return {
        click: async () => {
          const element = resolveElement(value);
          if (element === undefined) {
            throw new Error(`Unknown test id ${value}`);
          }
          if (element instanceof HTMLElement) {
            element.click();
          }
        },
        evaluate: async (
          pageFunction: (element: Element, arg?: unknown) => unknown,
          arg?: unknown,
        ) => {
          const element = resolveElement(value);
          if (element === undefined) {
            throw new Error(`Unknown test id ${value}`);
          }
          options.onEvaluate?.(value, element);
          return pageFunction(element, arg);
        },
        fill: async () => undefined,
        isVisible: async () => {
          const element = resolveElement(value);
          if (element === undefined) {
            return false;
          }
          return options.isVisible?.(value, element) ?? element.isConnected;
        },
      };
    },
  };
}

export function rectFromBox(options: {
  height: number;
  left: number;
  top: number;
  width: number;
}) {
  return {
    bottom: options.top + options.height,
    height: options.height,
    left: options.left,
    right: options.left + options.width,
    top: options.top,
    width: options.width,
    x: options.left,
    y: options.top,
    toJSON: () => ({}),
  } as DOMRect;
}
