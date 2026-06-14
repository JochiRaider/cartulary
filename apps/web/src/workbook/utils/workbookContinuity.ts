export type ScrollPosition = {
  top: number;
  left: number;
};

export type RectLike = {
  top: number;
  left: number;
  right: number;
  bottom: number;
  width: number;
  height: number;
};

export type ViewportAnchor = {
  top: number;
  left: number;
  width: number;
  height: number;
};

export type ViewportSnapshot = {
  scroll: ScrollPosition | null;
  anchor: ViewportAnchor | null;
};

export function captureViewportAnchor(
  containerRect: RectLike,
  elementRect: RectLike,
): ViewportAnchor {
  return {
    top: elementRect.top - containerRect.top,
    left: elementRect.left - containerRect.left,
    width: elementRect.width,
    height: elementRect.height,
  };
}

export function computeRestoredViewportScroll(options: {
  preservedScroll: ScrollPosition | null;
  currentScroll: ScrollPosition | null;
  preservedAnchor: ViewportAnchor | null;
  containerRect: RectLike;
  elementRect: RectLike;
}): ScrollPosition | null {
  const {
    currentScroll,
    preservedAnchor,
    preservedScroll,
    containerRect,
    elementRect,
  } = options;
  if (preservedScroll === null || currentScroll === null) {
    return null;
  }

  const currentTop = elementRect.top - containerRect.top;
  const currentLeft = elementRect.left - containerRect.left;

  let nextTop = currentScroll.top;
  if (preservedAnchor !== null) {
    nextTop += currentTop - preservedAnchor.top;
  }

  const predictedTop = currentTop - (nextTop - currentScroll.top);
  const predictedBottom = predictedTop + elementRect.height;
  if (predictedTop < 0) {
    nextTop += roundVisibilityDelta(predictedTop);
  } else if (predictedBottom > containerRect.height) {
    nextTop += Math.ceil(predictedBottom - containerRect.height);
  }

  let nextLeft = currentScroll.left;
  const predictedLeft = currentLeft;
  const predictedRight = predictedLeft + elementRect.width;
  if (predictedLeft < 0) {
    nextLeft += roundVisibilityDelta(predictedLeft);
  } else if (predictedRight > containerRect.width) {
    nextLeft += Math.ceil(predictedRight - containerRect.width);
  }

  return {
    top: Math.max(0, nextTop),
    left: Math.max(0, nextLeft),
  };
}

export function isRectFullyVisibleWithinContainer(
  containerRect: RectLike,
  elementRect: RectLike,
) {
  const top = elementRect.top - containerRect.top;
  const left = elementRect.left - containerRect.left;
  const bottom = elementRect.bottom - containerRect.top;
  const right = elementRect.right - containerRect.left;
  return (
    top >= 0 &&
    left >= 0 &&
    bottom <= containerRect.height &&
    right <= containerRect.width
  );
}

function roundVisibilityDelta(delta: number) {
  if (delta === 0) {
    return 0;
  }
  return delta > 0 ? Math.ceil(delta) : -Math.ceil(-delta);
}
