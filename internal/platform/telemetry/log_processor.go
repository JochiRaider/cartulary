package telemetry

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type dropNewLogProcessorConfig struct {
	MaxQueueSize       int
	MaxExportBatchSize int
	ExportInterval     time.Duration
	ExportTimeout      time.Duration
	SignalKind         string
	OnDrop             func(context.Context, string, string)
}

type dropNewLogProcessor struct {
	exporter      sdklog.Exporter
	queue         chan sdklog.Record
	flushRequests chan chan error
	stop          chan struct{}
	done          chan struct{}
	shutdownOnce  sync.Once
	stopped       atomic.Bool
	maxBatchSize  int
	exportTimeout time.Duration
	signalKind    string
	onDrop        func(context.Context, string, string)
}

func newDropNewLogProcessor(exporter sdklog.Exporter, cfg dropNewLogProcessorConfig) sdklog.Processor {
	if exporter == nil {
		exporter = noopLogExporter{}
	}
	maxQueueSize := cfg.MaxQueueSize
	if maxQueueSize < 1 {
		maxQueueSize = 1
	}
	maxBatchSize := cfg.MaxExportBatchSize
	if maxBatchSize < 1 || maxBatchSize > maxQueueSize {
		maxBatchSize = maxQueueSize
	}
	interval := cfg.ExportInterval
	if interval <= 0 {
		interval = time.Second
	}
	processor := &dropNewLogProcessor{
		exporter:      exporter,
		queue:         make(chan sdklog.Record, maxQueueSize),
		flushRequests: make(chan chan error),
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
		maxBatchSize:  maxBatchSize,
		exportTimeout: cfg.ExportTimeout,
		signalKind:    cfg.SignalKind,
		onDrop:        cfg.OnDrop,
	}
	go processor.run(interval)
	return processor
}

func (p *dropNewLogProcessor) Enabled(context.Context, sdklog.EnabledParameters) bool {
	return p != nil && !p.stopped.Load()
}

func (p *dropNewLogProcessor) OnEmit(ctx context.Context, record *sdklog.Record) error {
	if p == nil || record == nil || p.stopped.Load() {
		return nil
	}
	cloned := record.Clone()
	select {
	case p.queue <- cloned:
	default:
		if p.onDrop != nil {
			p.onDrop(ctx, p.signalKind, QueueFullDropReason)
		}
	}
	return nil
}

func (p *dropNewLogProcessor) ForceFlush(ctx context.Context) error {
	if p == nil || p.stopped.Load() {
		return nil
	}
	response := make(chan error, 1)
	select {
	case p.flushRequests <- response:
	case <-ctx.Done():
		return context.Cause(ctx)
	}
	select {
	case err := <-response:
		return err
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

func (p *dropNewLogProcessor) Shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}
	var shutdownErr error
	p.shutdownOnce.Do(func() {
		p.stopped.Store(true)
		close(p.stop)
		select {
		case <-p.done:
		case <-ctx.Done():
			shutdownErr = context.Cause(ctx)
		}
		shutdownErr = errors.Join(shutdownErr, p.exporter.Shutdown(ctx))
	})
	return shutdownErr
}

func (p *dropNewLogProcessor) run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(p.done)

	for {
		select {
		case <-ticker.C:
			_ = p.drain(context.Background())
		case response := <-p.flushRequests:
			response <- p.drain(context.Background())
		case <-p.stop:
			_ = p.drain(context.Background())
			return
		}
	}
}

func (p *dropNewLogProcessor) drain(ctx context.Context) error {
	var err error
	for {
		batch := p.dequeueBatch()
		if len(batch) == 0 {
			return err
		}
		err = errors.Join(err, p.exportBatch(ctx, batch))
	}
}

func (p *dropNewLogProcessor) dequeueBatch() []sdklog.Record {
	batch := make([]sdklog.Record, 0, p.maxBatchSize)
	for len(batch) < p.maxBatchSize {
		select {
		case record := <-p.queue:
			batch = append(batch, record)
		default:
			return batch
		}
	}
	return batch
}

func (p *dropNewLogProcessor) exportBatch(ctx context.Context, batch []sdklog.Record) error {
	if len(batch) == 0 {
		return nil
	}
	if p.exportTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.exportTimeout)
		defer cancel()
	}
	return p.exporter.Export(ctx, batch)
}

type noopLogExporter struct{}

func (noopLogExporter) Export(context.Context, []sdklog.Record) error { return nil }

func (noopLogExporter) Shutdown(context.Context) error { return nil }

func (noopLogExporter) ForceFlush(context.Context) error { return nil }
