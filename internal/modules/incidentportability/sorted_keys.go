package incidentportability

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// DefaultSortedKeyRunEntries bounds one in-memory run of owner keys. A run is
// sorted and deduplicated before it is sealed, so preparation memory stays
// bounded by the run, never by the number of source rows.
const DefaultSortedKeyRunEntries = 1 << 16

// SortedKeyRuns are sealed preparation members, each holding distinct owner
// keys in ascending byte order. MergeSortedKeys yields their distinct union in
// order. Owners use them for collection-wide checks (membership, ordering and
// distinct references) with bounded ordered passes (Core 01 REQ-01-686).
type SortedKeyRuns struct {
	Runs []PreparedRowStream
}

// Keys reports the number of keys staged across runs, before cross-run
// deduplication.
func (r SortedKeyRuns) Keys() int64 {
	var total int64
	for _, run := range r.Runs {
		total += run.Rows
	}
	return total
}

// SortedKeyStager buffers owner keys and seals each full buffer as one sorted,
// deduplicated run. Add may write a run synchronously, so it must not be
// called while another member write to the same destination is in progress.
type SortedKeyStager struct {
	ctx         context.Context
	destination MemberWriter
	prefix      string
	limit       int
	buffer      []string
	runs        []PreparedRowStream
	err         error
}

// NewSortedKeyStager stages runs at prefix plus a run ordinal.
func NewSortedKeyStager(ctx context.Context, destination MemberWriter, prefix string, runEntries int) *SortedKeyStager {
	if runEntries <= 0 {
		runEntries = DefaultSortedKeyRunEntries
	}
	stager := &SortedKeyStager{ctx: ctx, destination: destination, prefix: prefix, limit: runEntries}
	if destination == nil || strings.TrimSpace(prefix) == "" {
		stager.err = ErrPreparedStream
	}
	return stager
}

// Add buffers key. Keys are single lines; the owner chooses an encoding whose
// byte order is the order it needs.
func (s *SortedKeyStager) Add(key string) error {
	if s.err != nil {
		return s.err
	}
	if key == "" || strings.IndexByte(key, '\n') >= 0 || len(key) > MaxNDJSONLogicalLineBytes {
		s.err = ErrPreparedStream
		return s.err
	}
	s.buffer = append(s.buffer, key)
	if len(s.buffer) >= s.limit {
		s.err = s.flush()
	}
	return s.err
}

// Finish seals the last run and returns every run.
func (s *SortedKeyStager) Finish() (SortedKeyRuns, error) {
	if s.err == nil {
		s.err = s.flush()
	}
	if s.err != nil {
		return SortedKeyRuns{}, s.err
	}
	return SortedKeyRuns{Runs: slices.Clone(s.runs)}, nil
}

func (s *SortedKeyStager) flush() error {
	if len(s.buffer) == 0 {
		return nil
	}
	slices.Sort(s.buffer)
	keys := slices.Compact(s.buffer)
	path := fmt.Sprintf("%s%06d.keys", s.prefix, len(s.runs))
	run, err := StagePreparedRows(s.ctx, s.destination, path, sortedKeyCodec{}, func(emit func(string) error) error {
		for _, key := range keys {
			if err := emit(key); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.runs = append(s.runs, run)
	s.buffer = s.buffer[:0]
	return nil
}

// MergeSortedKeys visits the distinct union of runs in ascending byte order.
// It holds one key per run, and every run is read to EOF, which completes its
// seal verification and count check.
func MergeSortedKeys(ctx context.Context, source MemberReader, runs SortedKeyRuns, visit func(key string) error) error {
	if visit == nil {
		return ErrPreparedStream
	}
	return MergeSortedKeySets(ctx, []SortedKeySet{{Source: source, Runs: runs}}, func(key string, _ []bool) error {
		return visit(key)
	})
}

// SortedKeySet is one set of sorted runs and the preparation store that holds
// them; sets of different families or owners are read through their own
// sources.
type SortedKeySet struct {
	Source MemberReader
	Runs   SortedKeyRuns
}

// MergeSortedKeySets visits the distinct union of several key sets in
// ascending byte order. For each key, present reports which sets contain it.
// Memory holds one key per run; every run is read to EOF.
func MergeSortedKeySets(ctx context.Context, sets []SortedKeySet, visit func(key string, present []bool) error) error {
	if visit == nil {
		return ErrPreparedStream
	}
	var readers []*preparedRowReader[string]
	closeAll := func() error {
		var err error
		for _, reader := range readers {
			err = errors.Join(err, reader.close())
		}
		return err
	}
	merge := &sortedKeyHeap{}
	for index, set := range sets {
		for _, run := range set.Runs.Runs {
			if run.Rows == 0 {
				if err := requireEmptyStream(run); err != nil {
					return errors.Join(err, closeAll())
				}
				continue
			}
			reader, err := openPreparedRows(ctx, set.Source, run, PreparedRowCodec[string](sortedKeyCodec{}))
			if err != nil {
				return errors.Join(err, closeAll())
			}
			readers = append(readers, reader)
			if reader.next() {
				heap.Push(merge, sortedKeyHead{key: reader.row, reader: reader, set: index})
			} else if err := reader.finish(); err != nil {
				return errors.Join(err, closeAll())
			}
		}
	}
	present := make([]bool, len(sets))
	for merge.Len() > 0 {
		if err := ctx.Err(); err != nil {
			return errors.Join(err, closeAll())
		}
		key := (*merge)[0].key
		clear(present)
		for merge.Len() > 0 && (*merge)[0].key == key {
			head := heap.Pop(merge).(sortedKeyHead)
			present[head.set] = true
			if head.reader.next() {
				if head.reader.row <= head.key {
					return errors.Join(ErrPreparedStream, closeAll())
				}
				heap.Push(merge, sortedKeyHead{key: head.reader.row, reader: head.reader, set: head.set})
			} else if err := head.reader.finish(); err != nil {
				return errors.Join(err, closeAll())
			}
		}
		if err := visit(key, present); err != nil {
			return errors.Join(err, closeAll())
		}
	}
	return closeAll()
}

type sortedKeyCodec struct{}

func (sortedKeyCodec) Encode(key string) ([]byte, error)  { return []byte(key), nil }
func (sortedKeyCodec) Decode(line []byte) (string, error) { return string(line), nil }
func (sortedKeyCodec) Columns() []string                  { return nil }
func (sortedKeyCodec) Values(key string) ([]any, error)   { return []any{key}, nil }

type sortedKeyHead struct {
	key    string
	reader *preparedRowReader[string]
	set    int
}

type sortedKeyHeap []sortedKeyHead

func (h sortedKeyHeap) Len() int           { return len(h) }
func (h sortedKeyHeap) Less(i, j int) bool { return h[i].key < h[j].key }
func (h sortedKeyHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *sortedKeyHeap) Push(x any)        { *h = append(*h, x.(sortedKeyHead)) }
func (h *sortedKeyHeap) Pop() any {
	old := *h
	last := old[len(old)-1]
	*h = old[:len(old)-1]
	return last
}
