// Package startupproof owns the private, durable managed-start acquisition receipt.
package startupproof

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"sync"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

const SchemaID = "cartulary.test_services.startup_journal.v1"
const Filename = "suite-startup.json"
const ControlEnv = "CARTULARY_TEST_SERVICES_STARTUP_CONTROL_FD"

type Binding struct {
	RuntimeRoot  string   `json:"runtime_root"`
	RuntimeLease string   `json:"runtime_lease"`
	RunRoot      string   `json:"run_root"`
	RunID        string   `json:"run_id"`
	SuiteID      string   `json:"suite_id"`
	Services     []string `json:"services"`
	Dependencies []string `json:"dependencies"`
}
type Attempt struct {
	ID           string            `json:"id"`
	Service      string            `json:"service"`
	Labels       map[string]string `json:"labels"`
	CleanupFiles []string          `json:"cleanup_files"`
	ContainerID  string            `json:"container_id"`
	State        string            `json:"state"`
}
type Document struct {
	SchemaID string    `json:"schema_id"`
	Binding  Binding   `json:"binding"`
	State    string    `json:"state"`
	Helper   string    `json:"helper"`
	Attempts []Attempt `json:"attempts"`
}
type Journal struct {
	mu       sync.Mutex
	path     string
	document Document
}

func decode(data []byte) (Document, error) {
	var d Document
	value, err := harnessschema.Decode(data)
	if err != nil {
		return d, err
	}
	if err := harnessschema.Validate(SchemaID, value); err != nil {
		return d, err
	}
	err = json.Unmarshal(data, &d)
	return d, err
}
func Read(path string) (Document, error) {
	data, err := harnesslocal.ReadFile(path, 65536, false)
	if err != nil {
		return Document{}, err
	}
	return decode(data)
}
func Create(path string, binding Binding) error {
	d := Document{SchemaID: SchemaID, Binding: binding, State: "prepared", Attempts: []Attempt{}}
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	if _, err := decode(data); err != nil {
		return err
	}
	return harnesslocal.PublishNew(path, data)
}
func Open(path string, binding Binding) (*Journal, error) {
	d, err := Read(path)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(binding, d.Binding) || d.State != "prepared" {
		return nil, errors.New("startup binding mismatch")
	}
	return &Journal{path: path, document: d}, nil
}
func (j *Journal) update(change func(*Document) error) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	data, err := json.Marshal(j.document)
	if err != nil {
		return err
	}
	next, err := decode(data)
	if err != nil {
		return err
	}
	if err := change(&next); err != nil {
		return err
	}
	data, err = json.Marshal(next)
	if err != nil {
		return err
	}
	if _, err := decode(data); err != nil {
		return err
	}
	if err := harnesslocal.ReplaceFile(j.path, data); err != nil {
		return err
	}
	j.document = next
	return nil
}
func (j *Journal) Register(proof string) error {
	return j.update(func(d *Document) error {
		if d.State != "prepared" || proof == "" {
			return errors.New("invalid startup registration")
		}
		d.State, d.Helper = "registered", proof
		return nil
	})
}
func (j *Journal) Begin(service string, labels map[string]string, files []string) error {
	return j.update(func(d *Document) error {
		if d.State != "registered" || !slices.Contains(d.Binding.Services, service) || len(d.Attempts) >= 6 {
			return errors.New("startup acquisition not admitted")
		}
		for _, a := range d.Attempts {
			if a.Service == service && a.State != "settled" {
				return errors.New("prior acquisition unresolved")
			}
		}
		var id [16]byte
		if _, err := rand.Read(id[:]); err != nil {
			return err
		}
		if files == nil {
			files = []string{}
		}
		d.Attempts = append(d.Attempts, Attempt{ID: hex.EncodeToString(id[:]), Service: service, Labels: labels, CleanupFiles: files, State: "acquiring"})
		return nil
	})
}
func (j *Journal) Acquired(service, id string) error {
	return j.update(func(d *Document) error {
		for i := len(d.Attempts) - 1; i >= 0; i-- {
			a := &d.Attempts[i]
			if a.Service == service && a.State == "acquiring" && id != "" {
				a.ContainerID, a.State = id, "acquired"
				return nil
			}
		}
		return errors.New("missing acquisition intent")
	})
}
func (j *Journal) Settled(service string) error {
	return j.update(func(d *Document) error {
		for i := len(d.Attempts) - 1; i >= 0; i-- {
			a := &d.Attempts[i]
			if a.Service == service && a.State == "acquired" {
				a.State = "settled"
				return nil
			}
		}
		return errors.New("acquisition outcome uncertain")
	})
}

// NotSubmitted is used only before the provider call, never for its error result.
func (j *Journal) NotSubmitted(service string) error {
	return j.update(func(d *Document) error {
		for i := len(d.Attempts) - 1; i >= 0; i-- {
			a := &d.Attempts[i]
			if a.Service == service && a.State == "acquiring" {
				a.State = "settled"
				return nil
			}
		}
		return errors.New("missing unsubmitted intent")
	})
}
func (j *Journal) Finish(ready bool) error {
	return j.update(func(d *Document) error {
		if d.State != "registered" {
			return errors.New("startup already terminal")
		}
		d.State = "settled"
		for _, a := range d.Attempts {
			if a.State != "settled" {
				d.State = "unresolved"
			}
		}
		if ready {
			d.State = "published"
		}
		return nil
	})
}

// Only after the exact helper producer has been settled may these states prove
// no outstanding acquisition. An in-flight create is deliberately not absence.
func (d Document) SettledWithoutLease() bool {
	if d.State == "prepared" {
		return len(d.Attempts) == 0 && d.Helper == ""
	}
	if d.State == "registered" && len(d.Attempts) == 0 {
		return d.Helper != ""
	}
	return d.State == "settled" && slices.IndexFunc(d.Attempts, func(a Attempt) bool { return a.State != "settled" }) < 0
}
