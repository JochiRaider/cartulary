package objectstore

import (
	"context"
	"errors"
	"io"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

type telemetryStore struct {
	inner          Store
	serviceVersion string
	tracer         trace.Tracer
	duration       metric.Float64Histogram
	transferBytes  metric.Int64Histogram
}

var errTypedStoreUnavailable = errors.New("objectstore: typed adapter boundary unavailable")

func InstrumentStore(inner Store, serviceVersion string) Store {
	if inner == nil {
		return nil
	}
	meter := telemetry.Meter(telemetry.ScopeObjectStore, serviceVersion)
	duration, _ := meter.Float64Histogram(
		"cartulary.objectstore.operation.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Object-store dependency operation duration."),
	)
	transferBytes, _ := meter.Int64Histogram(
		"cartulary.objectstore.transfer.bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Safe object-store transfer size."),
	)
	return &telemetryStore{
		inner:          inner,
		serviceVersion: serviceVersion,
		tracer:         telemetry.Tracer(telemetry.ScopeObjectStore, serviceVersion),
		duration:       duration,
		transferBytes:  transferBytes,
	}
}

func (s *telemetryStore) UploadTarget(ctx context.Context, key string, expiresAt time.Time) (UploadTarget, error) {
	ctx, span, started := s.start(ctx, "create_upload_target")
	result, err := s.inner.UploadTarget(ctx, key, expiresAt)
	s.finish(ctx, span, started, "create_upload_target", err, nil)
	return result, err
}

func (s *telemetryStore) CompleteUploadTarget(ctx context.Context, token string, body io.Reader, contentType string) error {
	ctx, span, started := s.start(ctx, "complete_upload_target")
	err := s.inner.CompleteUploadTarget(ctx, token, body, contentType)
	s.finish(ctx, span, started, "complete_upload_target", err, nil)
	return err
}

func (s *telemetryStore) PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	ctx, span, started := s.start(ctx, "put_object")
	err := s.inner.PutObject(ctx, key, body, size, contentType)
	s.finish(ctx, span, started, "put_object", err, safeTransferSize(size, err))
	return err
}

func (s *telemetryStore) ReadObject(ctx context.Context, key string, options ReadOptions) (io.ReadCloser, ObjectInfo, error) {
	operation := "get_object"
	if options.RangeStart != nil {
		operation = "get_object_range"
	}
	ctx, span, started := s.start(ctx, operation)
	reader, info, err := s.inner.ReadObject(ctx, key, options)
	s.finish(ctx, span, started, operation, err, safeReadTransferSize(info, options, err))
	return reader, info, err
}

func (s *telemetryStore) StatObject(ctx context.Context, key string) (ObjectInfo, error) {
	ctx, span, started := s.start(ctx, "head_object")
	info, err := s.inner.StatObject(ctx, key)
	s.finish(ctx, span, started, "head_object", err, nil)
	return info, err
}

func (s *telemetryStore) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	ctx, span, started := s.start(ctx, "list_prefix")
	result, err := s.inner.ListObjects(ctx, prefix)
	s.finish(ctx, span, started, "list_prefix", err, nil)
	return result, err
}

func (s *telemetryStore) DeleteObject(ctx context.Context, key string) error {
	ctx, span, started := s.start(ctx, "delete_object")
	err := s.inner.DeleteObject(ctx, key)
	s.finish(ctx, span, started, "delete_object", err, nil)
	return err
}

func (s *telemetryStore) Close() error {
	return s.inner.Close()
}

func (s *telemetryStore) CreateUploadTarget(ctx context.Context, request UploadTargetRequest) (UploadTarget, error) {
	typed, err := s.typedStore()
	if err != nil {
		return UploadTarget{}, err
	}
	ctx, span, started := s.start(ctx, "create_upload_target")
	result, err := typed.CreateUploadTarget(ctx, request)
	s.finish(ctx, span, started, "create_upload_target", err, nil)
	return result, err
}

func (s *telemetryStore) Put(ctx context.Context, request PutObjectRequest) (PutObjectResult, error) {
	typed, err := s.typedStore()
	if err != nil {
		return PutObjectResult{}, err
	}
	ctx, span, started := s.start(ctx, "put_object")
	result, err := typed.Put(ctx, request)
	s.finish(ctx, span, started, "put_object", err, safeTransferSize(result.SizeBytes, err))
	return result, err
}

func (s *telemetryStore) Head(ctx context.Context, request HeadObjectRequest) (ObjectInfo, error) {
	typed, err := s.typedStore()
	if err != nil {
		return ObjectInfo{}, err
	}
	ctx, span, started := s.start(ctx, "head_object")
	result, err := typed.Head(ctx, request)
	s.finish(ctx, span, started, "head_object", err, nil)
	return result, err
}

func (s *telemetryStore) Get(ctx context.Context, request GetObjectRequest) (io.ReadCloser, ObjectInfo, error) {
	typed, err := s.typedStore()
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	operation := "get_object"
	if request.RangeStart != nil {
		operation = "get_object_range"
	}
	ctx, span, started := s.start(ctx, operation)
	reader, info, err := typed.Get(ctx, request)
	s.finish(ctx, span, started, operation, err, safeReadTransferSize(info, ReadOptions{RangeStart: request.RangeStart, RangeEnd: request.RangeEnd}, err))
	return reader, info, err
}

func (s *telemetryStore) ListPrefix(ctx context.Context, request ListPrefixRequest) (ListPrefixResult, error) {
	typed, err := s.typedStore()
	if err != nil {
		return ListPrefixResult{}, err
	}
	ctx, span, started := s.start(ctx, "list_prefix")
	result, err := typed.ListPrefix(ctx, request)
	s.finish(ctx, span, started, "list_prefix", err, nil)
	return result, err
}

func (s *telemetryStore) Delete(ctx context.Context, request DeleteObjectRequest) error {
	typed, err := s.typedStore()
	if err != nil {
		return err
	}
	ctx, span, started := s.start(ctx, "delete_object")
	err = typed.Delete(ctx, request)
	s.finish(ctx, span, started, "delete_object", err, nil)
	return err
}

func (s *telemetryStore) EnsureBucketForDevTest(ctx context.Context, request EnsureBucketRequest) (EnsureBucketResult, error) {
	typed, err := s.typedStore()
	if err != nil {
		return EnsureBucketResult{}, err
	}
	ctx, span, started := s.start(ctx, "ensure_bucket_for_dev_test")
	result, err := typed.EnsureBucketForDevTest(ctx, request)
	s.finish(ctx, span, started, "ensure_bucket_for_dev_test", err, nil)
	return result, err
}

func (s *telemetryStore) typedStore() (TypedStore, error) {
	typed, ok := s.inner.(TypedStore)
	if !ok {
		return nil, errTypedStoreUnavailable
	}
	return typed, nil
}

func (s *telemetryStore) start(ctx context.Context, operation string) (context.Context, trace.Span, time.Time) {
	started := time.Now()
	ctx, span := s.tracer.Start(
		ctx,
		"cartulary.objectstore.operation",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(telemetry.SafeAttributes(
			attribute.String("cartulary.operation", operation),
		)...),
	)
	return ctx, span, started
}

func (s *telemetryStore) finish(ctx context.Context, span trace.Span, started time.Time, operation string, err error, transferSize *int64) {
	result := "success"
	attrs := telemetry.SafeAttributes(
		attribute.String("cartulary.operation", operation),
	)
	if err != nil {
		result = "failed"
		if errors.Is(err, context.Canceled) {
			result = "canceled"
		}
		attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.error_class", objectStoreErrorClass(err)))...)
		span.SetStatus(codes.Error, "")
	}
	attrs = telemetry.SafeAttributes(append(attrs, attribute.String("cartulary.result", result))...)
	span.SetAttributes(attrs...)
	span.End()
	s.duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
	if transferSize != nil {
		s.transferBytes.Record(ctx, *transferSize, metric.WithAttributes(telemetry.SafeAttributes(
			attribute.String("cartulary.operation", operation),
			attribute.String("cartulary.result", result),
		)...))
	}
}

func objectStoreErrorClass(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	adapterErr, ok := AsAdapterError(err)
	if !ok {
		return "internal_error"
	}
	switch adapterErr.Code {
	case ErrorCodeUnavailable, ErrorCodeAccessRejected, ErrorCodeRetryExhausted:
		return "dependency_unavailable"
	case ErrorCodeDeadlineExceeded:
		return "timeout"
	case ErrorCodeInvalidRequest:
		return "invariant_violation"
	default:
		return "internal_error"
	}
}

func safeTransferSize(size int64, err error) *int64 {
	if err != nil || size < 0 {
		return nil
	}
	return &size
}

func safeReadTransferSize(info ObjectInfo, options ReadOptions, err error) *int64 {
	if err != nil || info.Size < 0 {
		return nil
	}
	if options.RangeStart == nil {
		return &info.Size
	}
	start := *options.RangeStart
	if start < 0 || start >= info.Size {
		return nil
	}
	size := info.Size - start
	if options.RangeEnd != nil && *options.RangeEnd >= start {
		size = *options.RangeEnd - start + 1
		if size > info.Size-start {
			size = info.Size - start
		}
	}
	return &size
}
