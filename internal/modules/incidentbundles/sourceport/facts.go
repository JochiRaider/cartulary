package sourceport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// ErrFactUnavailable reports a fact read that the catalog's recorded order does
// not admit, or a fact its family never published. It is a programming error
// and fails closed.
var ErrFactUnavailable = errors.New("incident bundle source fact is unavailable")

// ImportFacts carries read-only facts that a family publishes from its own
// preparation. A family earlier in the catalog's recorded order publishes; a
// later family reads while it prepares, so a cross-family consistency check is
// proven before the final transaction (Core 01 REQ-01-641). Facts are sealed
// sorted key runs in the operation's preparation store and are abandoned with
// it.
type ImportFacts struct {
	registry *factRegistry
	family   string
}

type factRegistry struct {
	mu        sync.Mutex
	store     PreparationStore
	order     map[string]int
	published map[string]incidentportability.SortedKeyRuns
}

// NewImportFacts starts an operation's fact registry over its preparation
// store. order is the catalog's recorded family order.
func NewImportFacts(store PreparationStore, order []string) *ImportFacts {
	positions := make(map[string]int, len(order))
	for index, family := range order {
		positions[family] = index
	}
	return &ImportFacts{registry: &factRegistry{store: store, order: positions, published: map[string]incidentportability.SortedKeyRuns{}}}
}

// forFamily scopes the registry to one preparing family. Only the source
// adapter scopes, so a family cannot publish or read as another family.
func (f *ImportFacts) forFamily(family string) *ImportFacts {
	if f == nil || f.registry == nil {
		return nil
	}
	return &ImportFacts{registry: f.registry, family: family}
}

func factKey(family, name string) string { return family + "/" + name }

func factPrefix(family, name string) string {
	return "prepared/core/" + family + "/facts/" + name + "/"
}

// Publish stages the keys produce emits as the preparing family's fact name: a
// sealed, sorted, deduplicated set. A family publishes each name once.
func (f *ImportFacts) Publish(ctx context.Context, name string, produce func(add func(key string) error) error) error {
	if f == nil || f.registry == nil || f.family == "" || strings.TrimSpace(name) == "" || strings.ContainsAny(name, "/.") || produce == nil {
		return ErrFactUnavailable
	}
	registry, key := f.registry, factKey(f.family, name)
	registry.mu.Lock()
	_, known := registry.order[f.family]
	_, exists := registry.published[key]
	registry.mu.Unlock()
	if !known || registry.store == nil {
		return ErrFactUnavailable
	}
	if exists {
		return fmt.Errorf("%w: %s published twice", ErrFactUnavailable, key)
	}
	stager := incidentportability.NewSortedKeyStager(ctx, registry.store, factPrefix(f.family, name), 0)
	if err := produce(stager.Add); err != nil {
		return err
	}
	runs, err := stager.Finish()
	if err != nil {
		return err
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.published[key]; exists {
		return fmt.Errorf("%w: %s published twice", ErrFactUnavailable, key)
	}
	registry.published[key] = runs
	return nil
}

// Fact returns the published fact name of family for reading while the scoped
// family prepares. Only a family earlier in the recorded order may be read.
func (f *ImportFacts) Fact(family, name string) (incidentportability.SortedKeySet, error) {
	if f == nil || f.registry == nil || f.family == "" {
		return incidentportability.SortedKeySet{}, ErrFactUnavailable
	}
	registry := f.registry
	registry.mu.Lock()
	defer registry.mu.Unlock()
	publisher, publisherKnown := registry.order[family]
	reader, readerKnown := registry.order[f.family]
	if !publisherKnown || !readerKnown || publisher >= reader {
		return incidentportability.SortedKeySet{}, fmt.Errorf("%w: %s may not read %s", ErrFactUnavailable, f.family, factKey(family, name))
	}
	runs, ok := registry.published[factKey(family, name)]
	if !ok {
		return incidentportability.SortedKeySet{}, fmt.Errorf("%w: %s was not published", ErrFactUnavailable, factKey(family, name))
	}
	return incidentportability.SortedKeySet{Source: factReader{store: registry.store, prefix: factPrefix(family, name)}, Runs: runs}, nil
}

// factReader confines reads to one published fact's runs.
type factReader struct {
	store  PreparationStore
	prefix string
}

func (r factReader) Members() []incidentportability.Member {
	result := []incidentportability.Member{}
	for _, member := range r.store.Members() {
		if strings.HasPrefix(member.Path, r.prefix) {
			result = append(result, member)
		}
	}
	return result
}

func (r factReader) OpenMember(ctx context.Context, path string) (io.ReadCloser, incidentportability.Member, error) {
	if !strings.HasPrefix(path, r.prefix) {
		return nil, incidentportability.Member{}, ErrFactUnavailable
	}
	return r.store.OpenMember(ctx, path)
}

// RecordAttributionFact is published by the records family: one key per
// admitted historical envelope actor, (record, column, actor).
const (
	RecordsFamilyID       = "records"
	RecordAttributionFact = "record_attribution"
)

// RecordAttributionKey encodes one record attribution fact key. Keys of one
// record share the RecordAttributionScope prefix and sort after it.
func RecordAttributionKey(recordID uuid.UUID, column string, actorID uuid.UUID) string {
	return RecordAttributionScope(recordID) + column + "|" + actorID.String()
}

// RecordAttributionScope is the key prefix of one record's attribution facts;
// as a key of its own it sorts before every attribution key of the record.
func RecordAttributionScope(recordID uuid.UUID) string { return recordID.String() + "|" }
