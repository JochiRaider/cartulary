//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
)

// Compatibility keys use immutable run configuration. They never contain
// allocation generations or borrower capabilities.
type rendererConsumerKey struct {
	Profile, Policy string
	Diagnostics     bool
}

func rendererCompatibility(profile, policy string, diagnostics bool) rendererConsumerKey {
	return rendererConsumerKey{profile, policy, diagnostics}
}
func (s *rendererSlot) compatibility() rendererConsumerKey {
	return rendererCompatibility(s.profile, s.policy, s.diagnostics)
}

type fixtureConsumer struct {
	stack, cluster string
	renderer       rendererConsumerKey
	objects        bool
	disposition    string
	sequence       uint64
}

func consumerOf(w WorkRequirement) fixtureConsumer {
	return fixtureConsumer{stack: sharedKey(w.Request), cluster: clusterKey(w.Request),
		renderer: rendererCompatibility(w.Request.RuntimeProfile, w.Renderer, w.Diagnostics),
		objects:  slices.Contains(w.Request.Services, "object_store")}
}

type consumerRegistry struct {
	mu         sync.Mutex
	units      map[string]fixtureConsumer
	registered bool
	used       bool
	sequence   uint64
	wake       chan struct{}
	stop       chan struct{}
	done       chan struct{}
	stopping   bool
	err        error
}

// RegisterConsumers is mandatory and single-use for compiled graph owners.
// Dedicated interactive owners omit registration and cannot retain resources.
func (p *Production) RegisterConsumers(requirements []WorkRequirement) error {
	c := &p.consumers
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.registered || c.used {
		return errors.New("fixture consumers already registered or admission started")
	}
	units := map[string]fixtureConsumer{}
	for _, w := range requirements {
		if w.Request.UnitID == "" {
			return errors.New("missing fixture consumer identity")
		}
		if _, exists := units[w.Request.UnitID]; exists {
			return errors.New("duplicate fixture consumer")
		}
		if err := validateServiceRequest(w.Request); err != nil {
			return err
		}
		if w.Renderer != "none" && w.Renderer != "pooled" && w.Renderer != "dedicated" {
			return errors.New("invalid consumer renderer policy")
		}
		units[w.Request.UnitID] = consumerOf(w)
	}
	c.units, c.registered = units, true
	c.wake, c.stop, c.done = make(chan struct{}, 1), make(chan struct{}), make(chan struct{})
	go p.consumerCleanup()
	return nil
}

func (p *Production) validateConsumer(w WorkRequirement) error {
	c := &p.consumers
	c.mu.Lock()
	defer c.mu.Unlock()
	c.used = true
	if !c.registered {
		return nil
	} // Standalone dedicated ownership, no reuse.
	actual, ok := c.units[w.Request.UnitID]
	if !ok || actual.disposition != "" || actual != consumerOf(w) {
		return errors.New("unknown, terminal or incompatible fixture consumer")
	}
	return nil
}

// ConsumerTerminal changes only logical accounting. Physical cleanup runs on
// the fixture owner's bounded dispatcher, never the graph event goroutine.
func (p *Production) ConsumerTerminal(unit, disposition string) error {
	c := &p.consumers
	c.mu.Lock()
	defer c.mu.Unlock()
	if !slices.Contains([]string{"passed", "failed", "skipped", "cached", "cancelled"}, disposition) {
		return errors.New("invalid fixture consumer disposition")
	}
	consumer, ok := c.units[unit]
	if !c.registered || !ok {
		return errors.New("unknown fixture consumer")
	}
	if consumer.disposition != "" {
		if consumer.disposition == disposition {
			return nil
		}
		return errors.New("contradictory fixture consumer disposition")
	}
	if c.stopping {
		return errors.New("fixture consumer notification after finalization")
	}
	p.workMu.Lock()
	active := p.work[unit] != nil
	p.workMu.Unlock()
	if active {
		return errors.New("fixture consumer terminal before borrow settlement")
	}
	c.sequence++
	consumer.sequence = c.sequence
	consumer.disposition = disposition
	c.units[unit] = consumer
	select {
	case c.wake <- struct{}{}:
	default:
	}
	return nil
}

func (p *Production) remaining(exclude string, matches func(fixtureConsumer) bool) int {
	c := &p.consumers
	c.mu.Lock()
	defer c.mu.Unlock()
	total := 0
	for id, v := range c.units {
		if id != exclude && v.disposition == "" && matches(v) {
			total++
		}
	}
	return total
}
func (p *Production) stackConsumers(key, exclude string) int {
	return p.remaining(exclude, func(v fixtureConsumer) bool { return v.stack == key })
}
func (p *Production) rendererConsumers(key rendererConsumerKey, exclude string) int {
	return p.remaining(exclude, func(v fixtureConsumer) bool { return v.renderer == key })
}
func (p *Production) lastTerminal(matches func(fixtureConsumer) bool) string {
	c := &p.consumers
	c.mu.Lock()
	defer c.mu.Unlock()
	last, sequence := "", uint64(0)
	for id, v := range c.units {
		if matches(v) && v.sequence > sequence {
			last, sequence = id, v.sequence
		}
	}
	return last
}
func (p *Production) consumerCleanup() {
	c := &p.consumers
	defer close(c.done)
	for {
		select {
		case <-c.stop:
			return
		case <-c.wake:
		}
		// Closing the dispatcher takes priority over a coalesced wake. Create
		// the allowance under the same lock so every joined operation's
		// deadline predates finalization's independent cleanup allowance.
		c.mu.Lock()
		if c.stopping {
			c.mu.Unlock()
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		c.mu.Unlock()
		err := p.retireUnused(ctx)
		cancel()
		if err != nil {
			c.mu.Lock()
			c.err = errors.Join(c.err, err)
			c.mu.Unlock()
		}
	}
}
func (p *Production) joinConsumerCleanup() error {
	c := &p.consumers
	c.mu.Lock()
	if !c.registered {
		c.mu.Unlock()
		return nil
	}
	if !c.stopping {
		c.stopping = true
		close(c.stop)
	}
	done := c.done
	c.mu.Unlock()
	<-done // Each dispatched cleanup has its own bounded allowance.
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.err
	for id, v := range c.units {
		if v.disposition == "" {
			err = errors.Join(err, fmt.Errorf("nonterminal fixture consumer %s", id))
		}
	}
	return CleanupFailure(err)
}

func (p *Production) retireUnused(ctx context.Context) error {
	p.retireMu.Lock()
	defer p.retireMu.Unlock()
	if _, err := p.retireAllocations(ctx, nil, true, Retirement{"terminal_borrower_release", ""}); err != nil {
		return err
	}
	// Shared/measurement services outlive both logical consumers and physical
	// dependents. The resource dependency evaluator prevents premature closure.
	p.servicesMu.Lock()
	keys := make([]string, 0, len(p.clusters))
	for key := range p.clusters {
		keys = append(keys, key)
	}
	p.servicesMu.Unlock()
	slices.Sort(keys)
	for _, key := range keys {
		if p.remaining("", func(v fixtureConsumer) bool { return v.cluster == key }) != 0 {
			continue
		}
		p.Options.residency.mu.Lock()
		record := p.Options.residency.allocations[key]
		p.Options.residency.mu.Unlock()
		if record != nil {
			if err := p.Options.residency.drain(ctx, []*residentRecord{record}, func(ctx context.Context) error { return p.retireCluster(ctx, key) }); err != nil {
				return err
			}
		}
	}
	if p.Options.Environment["CARTULARY_TEST_SERVICES_MODE"] != "attach" && p.remaining("", func(v fixtureConsumer) bool { return v.objects }) == 0 {
		p.Options.residency.mu.Lock()
		record := p.Options.residency.allocations["object_store:shared"]
		p.Options.residency.mu.Unlock()
		if record != nil {
			return p.Options.residency.drain(ctx, []*residentRecord{record}, p.Suite.Close)
		}
	}
	// External persistent owners retain their independent authority.
	return nil
}
