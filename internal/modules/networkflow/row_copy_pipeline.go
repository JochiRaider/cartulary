package networkflow

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/jackc/pgx/v5"
)

// Bounds for the pipelined import copy. The queue bound keeps encoded rows that
// wait for COPY at a few MiB; one admitted row larger than the bound still
// travels alone. The read-ahead size sets how much protected source one
// connection hand-off reads, and therefore how long each COPY statement runs.
const (
	pipelinedCopyQueueBytes  = 4 << 20
	pipelinedSourceReadAhead = 1 << 20
)

// rowCopyPipeline overlaps row production with PostgreSQL's COPY work. The
// producer validates and encodes rows on its own goroutine while COPY consumes
// them. The producer also reads the protected source through the same borrowed
// transaction, and one connection runs one statement at a time: a source read
// first ends the current COPY statement, the read runs, and the next statement
// starts with the rows produced after it. Rows are never reordered or dropped.
type rowCopyPipeline struct {
	mu          sync.Mutex
	cond        *sync.Cond
	queue       []pipelinedRow
	head        int
	queueBytes  int
	pushed      int64
	copied      int64
	copying     bool
	connHeld    int
	wantConn    bool
	done        bool
	producerErr error
	consumerErr error
	// producerFirst records that the producer failed before COPY did; COPY then
	// only reflects the abort, and the producer's classification is primary.
	producerFirst bool
	ctx           context.Context
}

type pipelinedRow struct {
	values []any
	size   int
}

func newRowCopyPipeline() *rowCopyPipeline {
	p := &rowCopyPipeline{}
	p.cond = sync.NewCond(&p.mu)
	return p
}

func (p *rowCopyPipeline) pending() int { return len(p.queue) - p.head }

// push queues one encoded row and blocks while the byte bound is full. It
// fails once COPY has failed, which stops the producer.
func (p *rowCopyPipeline) push(values []any, size int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for p.consumerErr == nil && p.pending() > 0 && p.queueBytes+size > pipelinedCopyQueueBytes {
		p.cond.Wait()
	}
	if p.consumerErr != nil {
		return p.consumerErr
	}
	if p.head > 0 && p.head == len(p.queue) {
		p.queue, p.head = p.queue[:0], 0
	}
	p.queue = append(p.queue, pipelinedRow{values: values, size: size})
	p.queueBytes += size
	p.pushed++
	p.cond.Broadcast()
	return nil
}

// acquireConnection gives the producer the borrowed connection for a source
// read. It waits until the current COPY statement has ended; COPY ends as soon
// as the rows already queued are sent. Acquisition is reentrant for the
// producer goroutine, the only caller.
func (p *rowCopyPipeline) acquireConnection() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.connHeld > 0 {
		p.connHeld++
		return nil
	}
	p.wantConn = true
	p.cond.Broadcast()
	for p.copying && p.consumerErr == nil {
		p.cond.Wait()
	}
	if p.consumerErr != nil {
		p.wantConn = false
		p.cond.Broadcast()
		return p.consumerErr
	}
	p.connHeld = 1
	return nil
}

func (p *rowCopyPipeline) releaseConnection() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.connHeld == 0 {
		return
	}
	p.connHeld--
	if p.connHeld == 0 {
		p.wantConn = false
		p.cond.Broadcast()
	}
}

func (p *rowCopyPipeline) finishProducer(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done, p.producerErr = true, err
	p.producerFirst = err != nil && p.consumerErr == nil
	p.cond.Broadcast()
}

// run executes produce on its own goroutine and COPYs its rows in order. The
// producer's error wins over the COPY error it caused, so owner
// classifications (source change, cancellation) survive.
func (p *rowCopyPipeline) run(ctx context.Context, tx pgx.Tx, columns []string, produce func(gate rowSourceGate) error) error {
	p.ctx = ctx
	stop := context.AfterFunc(ctx, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.cond.Broadcast()
	})
	defer stop()
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		var err error
		defer func() {
			if recovered := recover(); recovered != nil {
				err = fmt.Errorf("network flow row producer panic: %v", recovered)
			}
			p.finishProducer(err)
		}()
		err = produce(p)
	}()
	var copyErr error
	for {
		p.mu.Lock()
		for !p.done && (p.pending() == 0 || p.wantConn) {
			p.cond.Wait()
		}
		if p.pending() == 0 || p.producerErr != nil {
			p.mu.Unlock()
			break
		}
		p.copying = true
		p.mu.Unlock()
		written, err := tx.CopyFrom(ctx, pgx.Identifier{"network_flow_rows"}, columns, &pipelineCopySource{pipeline: p})
		p.mu.Lock()
		p.copying = false
		p.copied += written
		if err != nil {
			p.consumerErr = err
		}
		p.cond.Broadcast()
		p.mu.Unlock()
		if err != nil {
			copyErr = err
			break
		}
	}
	<-finished
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.producerErr != nil && (copyErr == nil || p.producerFirst) {
		return p.producerErr
	}
	if copyErr != nil {
		return fmt.Errorf("insert network flow row batch: %w", copyErr)
	}
	if p.copied != p.pushed || p.pending() != 0 {
		return errSourceChanged
	}
	return nil
}

// pipelineCopySource feeds one COPY statement. It ends the statement when the
// queue is empty and either the producer finished or needs the connection.
type pipelineCopySource struct {
	pipeline *rowCopyPipeline
	current  []any
}

func (s *pipelineCopySource) Next() bool {
	p := s.pipeline
	p.mu.Lock()
	defer p.mu.Unlock()
	for p.pending() == 0 && !p.done && !p.wantConn && p.consumerErr == nil && p.ctx.Err() == nil {
		p.cond.Wait()
	}
	if p.pending() == 0 || p.producerErr != nil || p.ctx.Err() != nil {
		s.current = nil
		return false
	}
	row := p.queue[p.head]
	p.queue[p.head] = pipelinedRow{}
	p.head++
	p.queueBytes -= row.size
	s.current = row.values
	p.cond.Broadcast()
	return true
}

func (s *pipelineCopySource) Values() ([]any, error) { return s.current, nil }

// Err aborts the statement with CopyFail when the producer failed or the
// operation was canceled, so a partial statement never completes.
func (s *pipelineCopySource) Err() error {
	p := s.pipeline
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.producerErr != nil {
		return p.producerErr
	}
	return p.ctx.Err()
}

// withConnection runs fn while the producer holds the borrowed connection,
// after any open COPY statement has ended.
func (p *rowCopyPipeline) withConnection(fn func() error) error {
	if err := p.acquireConnection(); err != nil {
		return err
	}
	defer p.releaseConnection()
	return fn()
}

// readAhead fills up to size bytes of protected source under one connection
// hand-off, so COPY statements span the rows of that whole range.
func (p *rowCopyPipeline) readAhead(reader io.Reader, size int) io.Reader {
	return &pipelineReadAhead{reader: reader, pipeline: p, buffer: make([]byte, 0, size)}
}

// pipelineReadAhead preserves the wrapped reader's bytes and error order: it
// returns every byte read before an error, then the error.
type pipelineReadAhead struct {
	reader   io.Reader
	pipeline *rowCopyPipeline
	buffer   []byte
	offset   int
	err      error
}

func (r *pipelineReadAhead) Read(destination []byte) (int, error) {
	if r.offset == len(r.buffer) {
		if r.err != nil {
			return 0, r.err
		}
		if len(destination) == 0 {
			return 0, nil
		}
		r.fill()
		if r.offset == len(r.buffer) {
			return 0, r.err
		}
	}
	n := copy(destination, r.buffer[r.offset:])
	r.offset += n
	return n, nil
}

func (r *pipelineReadAhead) fill() {
	r.buffer, r.offset = r.buffer[:0], 0
	if err := r.pipeline.acquireConnection(); err != nil {
		r.err = err
		return
	}
	defer r.pipeline.releaseConnection()
	for len(r.buffer) < cap(r.buffer) {
		n, err := r.reader.Read(r.buffer[len(r.buffer):cap(r.buffer)])
		r.buffer = r.buffer[:len(r.buffer)+n]
		if err != nil {
			r.err = err
			return
		}
		if n == 0 {
			return
		}
	}
}

// rowSourceGate gives the row producer the borrowed connection between COPY
// statements. The producer issues every statement of its source pass through
// withConnection or a readAhead reader; it never runs one while COPY holds the
// connection.
type rowSourceGate interface {
	withConnection(func() error) error
	readAhead(io.Reader, int) io.Reader
}
