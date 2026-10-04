package packformat

import (
	"bufio"
	"bytes"
	"container/heap"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
)

// DiagnosticScratch is a private, confined workspace separate from extracted
// content. Runs contain only issue digests. The caller owns and closes the
// workspace on every outcome, including cancellation and uncertain execution.
type DiagnosticScratch interface {
	Write(context.Context, string, func(io.Writer) error) error
	Open(context.Context, string) (io.ReadCloser, error)
	Remove(context.Context, string) error
}

const diagnosticChunk = 4096
const diagnosticRetained = 1000

type diagnosticRun struct {
	name   string
	count  int
	digest [32]byte
}

// A bounded prefix heap retains the only issue objects that can be published.
// Exact counts use a binary merge of sorted digest runs: at most 4096 digests,
// two read buffers, one write buffer and 64 run descriptors stay in memory.
// At most two copies of admitted digest records exist during a merge; obsolete
// runs are deleted immediately. No collection grows with the finding count.
type diagnosticCollector struct {
	ctx      context.Context
	check    Check
	scratch  DiagnosticScratch
	buffer   [][32]byte
	levels   [64]*diagnosticRun
	sequence uint64
	prefix   issueHeap
	retained map[string]bool
	failure  error
}

func newDiagnosticCollector(ctx context.Context, check Check, scratch DiagnosticScratch) *diagnosticCollector {
	return &diagnosticCollector{ctx: ctx, check: check, scratch: scratch, buffer: make([][32]byte, 0, diagnosticChunk), retained: make(map[string]bool, diagnosticRetained)}
}

func (c *diagnosticCollector) add(f Finding) (resultErr error) {
	if c.failure != nil {
		return c.failure
	}
	defer func() {
		if resultErr != nil {
			c.failure = resultErr
		}
	}()
	if err := c.ctx.Err(); err != nil {
		return err
	}
	issue, err := makeIssue(c.check, f)
	if err != nil {
		return err
	}
	if !c.retained[issue.ID] {
		if len(c.prefix) < diagnosticRetained {
			heap.Push(&c.prefix, issue)
			c.retained[issue.ID] = true
		} else if compareIssues(issue, c.prefix[0]) < 0 {
			delete(c.retained, c.prefix[0].ID)
			c.prefix[0] = issue
			c.retained[issue.ID] = true
			heap.Fix(&c.prefix, 0)
		}
	}
	var digest [32]byte
	if _, err := hex.Decode(digest[:], []byte(issue.ID[4:])); err != nil {
		return err
	}
	c.buffer = append(c.buffer, digest)
	if len(c.buffer) == diagnosticChunk {
		return c.flush()
	}
	return nil
}

func (c *diagnosticCollector) writeRun(write func(io.Writer) (int, error)) (*diagnosticRun, error) {
	if c.scratch == nil {
		return nil, errors.New("reference pack: diagnostic scratch unavailable")
	}
	c.sequence++
	name := fmt.Sprintf("diagnostics/run-%016x", c.sequence)
	count := 0
	digest := sha256.New()
	err := c.scratch.Write(c.ctx, name, func(destination io.Writer) error {
		writer := bufio.NewWriterSize(io.MultiWriter(destination, digest), 32768)
		var err error
		count, err = write(writer)
		if err != nil {
			return err
		}
		return writer.Flush()
	})
	if err != nil {
		return nil, err
	}
	var sum [32]byte
	copy(sum[:], digest.Sum(nil))
	return &diagnosticRun{name: name, count: count, digest: sum}, nil
}

func (c *diagnosticCollector) flush() error {
	if len(c.buffer) == 0 {
		return nil
	}
	slices.SortFunc(c.buffer, func(a, b [32]byte) int { return bytes.Compare(a[:], b[:]) })
	c.buffer = slices.Compact(c.buffer)
	run, err := c.writeRun(func(w io.Writer) (int, error) {
		for _, digest := range c.buffer {
			if err := c.ctx.Err(); err != nil {
				return 0, err
			}
			if _, err := w.Write(digest[:]); err != nil {
				return 0, err
			}
		}
		return len(c.buffer), nil
	})
	if err != nil {
		return err
	}
	c.buffer = c.buffer[:0]
	for level := range c.levels {
		if c.levels[level] == nil {
			c.levels[level] = run
			return nil
		}
		run, err = c.merge(c.levels[level], run)
		if err != nil {
			return err
		}
		c.levels[level] = nil
	}
	return errors.New("reference pack: diagnostic count overflow")
}

func (c *diagnosticCollector) merge(left, right *diagnosticRun) (result *diagnosticRun, resultErr error) {
	a, err := c.scratch.Open(c.ctx, left.name)
	if err != nil {
		return nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, a.Close()) }()
	b, err := c.scratch.Open(c.ctx, right.name)
	if err != nil {
		return nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, b.Close()) }()
	result, err = c.writeRun(func(w io.Writer) (int, error) {
		ah, bh := sha256.New(), sha256.New()
		al := &io.LimitedReader{R: io.TeeReader(a, ah), N: int64(left.count)*32 + 1}
		bl := &io.LimitedReader{R: io.TeeReader(b, bh), N: int64(right.count)*32 + 1}
		ar, br := bufio.NewReaderSize(al, 32768), bufio.NewReaderSize(bl, 32768)
		var av, bv, previous [32]byte
		aerr, berr := readDiagnosticDigest(ar, &av), readDiagnosticDigest(br, &bv)
		count := 0
		for aerr == nil || berr == nil {
			if err := c.ctx.Err(); err != nil {
				return 0, err
			}
			var next [32]byte
			if aerr == nil && (berr != nil || bytes.Compare(av[:], bv[:]) <= 0) {
				next = av
				aerr = readDiagnosticDigest(ar, &av)
			} else {
				next = bv
				berr = readDiagnosticDigest(br, &bv)
			}
			if count > 0 && previous == next {
				continue
			}
			if _, err := w.Write(next[:]); err != nil {
				return 0, err
			}
			previous = next
			count++
		}
		if !errors.Is(aerr, io.EOF) || !errors.Is(berr, io.EOF) {
			return 0, errors.New("reference pack: diagnostic run unavailable")
		}
		if al.N != 1 || bl.N != 1 || !bytes.Equal(ah.Sum(nil), left.digest[:]) || !bytes.Equal(bh.Sum(nil), right.digest[:]) {
			return 0, errors.New("reference pack: diagnostic run integrity failure")
		}
		return count, nil
	})
	if err != nil {
		return nil, err
	}
	if err := c.scratch.Remove(c.ctx, left.name); err != nil {
		return nil, err
	}
	if err := c.scratch.Remove(c.ctx, right.name); err != nil {
		return nil, err
	}
	return result, nil
}

func readDiagnosticDigest(reader io.Reader, target *[32]byte) error {
	_, err := io.ReadFull(reader, target[:])
	return err
}

func (c *diagnosticCollector) finish() (*ValidationSummary, error) {
	if c.failure != nil {
		return nil, c.failure
	}
	if err := c.ctx.Err(); err != nil {
		return nil, err
	}
	if len(c.prefix) == 0 {
		return successSummary(), nil
	}
	var run *diagnosticRun
	if c.sequence == 0 {
		// Small checks need no filesystem writes. This includes success and all
		// scalar guards. The fixed chunk is still the complete deduplication set.
		slices.SortFunc(c.buffer, func(a, b [32]byte) int { return bytes.Compare(a[:], b[:]) })
		c.buffer = slices.Compact(c.buffer)
		run = &diagnosticRun{count: len(c.buffer)}
	} else {
		if err := c.flush(); err != nil {
			return nil, err
		}
		for _, next := range c.levels {
			if next == nil {
				continue
			}
			if run == nil {
				run = next
				continue
			}
			var err error
			run, err = c.merge(run, next)
			if err != nil {
				return nil, err
			}
		}
		if err := c.scratch.Remove(c.ctx, run.name); err != nil {
			return nil, err
		}
	}
	if err := c.ctx.Err(); err != nil {
		return nil, err
	}
	ordered := []Issue(c.prefix)
	slices.SortFunc(ordered, compareIssues)
	primary := ordered[0].ID
	return &ValidationSummary{SchemaID: "cartulary.reference_pack_validation_summary.v1", Result: "failed", PrimaryIssueID: &primary, Truncated: run.count > diagnosticRetained, Total: run.count, Retained: len(ordered), Issues: ordered}, nil
}

type issueHeap []Issue

func (h issueHeap) Len() int           { return len(h) }
func (h issueHeap) Less(i, j int) bool { return compareIssues(h[i], h[j]) > 0 }
func (h issueHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *issueHeap) Push(value any)    { *h = append(*h, value.(Issue)) }
func (h *issueHeap) Pop() any {
	old := *h
	value := old[len(old)-1]
	*h = old[:len(old)-1]
	return value
}
