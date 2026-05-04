package timeline

import "context"

type createTimingContextKey struct{}

type CreateTimingRecorder interface {
	MarkTimelineCreateTiming(name string)
}

func WithCreateTimingRecorder(ctx context.Context, recorder CreateTimingRecorder) context.Context {
	if recorder == nil {
		return ctx
	}
	return context.WithValue(ctx, createTimingContextKey{}, recorder)
}

func markCreateTiming(ctx context.Context, name string) {
	recorder, _ := ctx.Value(createTimingContextKey{}).(CreateTimingRecorder)
	if recorder == nil {
		return
	}
	recorder.MarkTimelineCreateTiming(name)
}
