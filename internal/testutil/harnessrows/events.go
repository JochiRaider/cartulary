// Package harnessrows is the Go row runner's adapter: it turns the event
// stream of one `go test -json` invocation into the outcomes of the catalog
// rows the invocation runs (TH-HARNESS-REQ-872). It holds no execution, file
// or schema logic: the harness runtime runs the command, feeds its standard
// output here, and validates and publishes what comes back.
package harnessrows

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/testfailure"
)

const (
	// MaxEventBytes is the longest event line the stream may carry.
	MaxEventBytes  = 1 << 20
	maxMarkerBytes = 64 << 10
	maxPartials    = 128
)

// ErrEventTooLong reports an event line over MaxEventBytes. The stream's
// outcomes are then unknown, so none are published.
var ErrEventTooLong = errors.New("go test JSON event exceeds the 1 MiB line contract")

// Row is one catalog row of the invocation: its exact top-level tests.
type Row struct {
	ID        string   `json:"row_id"`
	Selectors []string `json:"selectors"`
}

// Observation is a test's terminal event.
type Observation struct {
	Action     string
	DurationMS int64
}

type marker struct {
	test      string
	envelope  testfailure.Envelope
	malformed bool
}

type lineState struct {
	test     string
	tail     string
	encoded  *strings.Builder
	bytes    int
	overflow bool
}

// Stream observes one invocation's standard output. It is an io.Writer, so
// the runtime can tee the child's output into it while the child runs; Finish
// returns what the stream established.
type Stream struct {
	rows    []Row
	line    []byte
	tooLong bool
	err     error

	terminal map[string]Observation
	order    []string // terminal tests in first-terminal order
	markers  []marker
	partial  map[[2]string]*lineState

	packageName    string
	packageOutcome *Observation
	packageElapsed *int64

	details failureDetails
	// build keeps the last 4 KiB of compiler output when the test binary
	// fails to build; no row then reports a result.
	build       []byte
	buildFailed bool
}

// NewStream observes an invocation of rows.
func NewStream(rows []Row) *Stream {
	s := &Stream{rows: rows, terminal: map[string]Observation{}, partial: map[[2]string]*lineState{}}
	s.details = newFailureDetails(rows)
	return s
}

// Write splits the output into lines and observes each event.
func (s *Stream) Write(p []byte) (int, error) {
	n := len(p)
	for len(p) > 0 {
		index := bytes.IndexByte(p, '\n')
		piece := p
		if index >= 0 {
			piece = p[:index]
		}
		if !s.tooLong {
			if len(s.line)+len(piece) > MaxEventBytes {
				s.tooLong, s.line = true, s.line[:0]
				s.err = ErrEventTooLong
			} else {
				s.line = append(s.line, piece...)
			}
		}
		if index < 0 {
			break
		}
		if !s.tooLong {
			s.event(s.line)
		}
		s.line, s.tooLong = s.line[:0], false
		p = p[index+1:]
	}
	return n, nil
}

// event observes one line. Blank lines, process diagnostics and JSON that is
// not an object are not events and cannot complete a fragment.
func (s *Stream) event(line []byte) {
	if len(bytes.TrimSpace(line)) == 0 {
		return
	}
	var fields map[string]any
	if json.Unmarshal(line, &fields) != nil || fields == nil {
		return
	}
	action, _ := fields["Action"].(string)
	switch action {
	case "build-output":
		output, _ := fields["Output"].(string)
		s.build = append(s.build, output...)
		s.build = slices.Clone(s.build[max(0, len(s.build)-detailTailBytes):])
		return
	case "build-fail":
		s.buildFailed = true
		return
	}
	pkg, _ := fields["Package"].(string)
	if s.packageName == "" && pkg != "" {
		s.packageName = pkg
	}
	test, isTest := fields["Test"].(string)
	terminal := action == "pass" || action == "fail" || action == "skip"
	if !isTest {
		// The package's own terminal event carries the test binary's elapsed time.
		if terminal {
			s.packageOutcome = &Observation{Action: action}
			s.packageElapsed = nil
			if seconds, ok := fields["Elapsed"].(float64); ok {
				ms := max(0, int64(math.Round(seconds*1000)))
				s.packageElapsed = &ms
			}
		}
		return
	}
	output, hasOutput := fields["Output"].(string)
	s.details.observe(test, action, output, hasOutput)
	// Package identity also keeps a malformed mixed-package stream from joining
	// fragments of equally named tests in different packages.
	key := [2]string{pkg, test}
	if hasOutput {
		s.output(key, test, output)
	}
	if terminal {
		if state := s.partial[key]; state != nil {
			s.finishLine(state, false)
		}
		delete(s.partial, key)
		duration := int64(0)
		if seconds, ok := fields["Elapsed"].(float64); ok {
			duration = max(0, int64(math.Round(seconds*1000)))
		}
		if _, seen := s.terminal[test]; !seen {
			s.order = append(s.order, test)
		}
		s.terminal[test] = Observation{Action: action, DurationMS: duration}
	}
}

// output reassembles failure markers. go test -json may split one logical
// output line across events and interleave other tests' output.
func (s *Stream) output(key [2]string, test, output string) {
	state := s.partial[key]
	delete(s.partial, key)
	if state == nil {
		state = &lineState{test: test}
	}
	pieces := strings.Split(output, "\n")
	for index, piece := range pieces {
		if state.encoded == nil {
			candidate := state.tail + piece
			if at := strings.Index(candidate, testfailure.Marker); at >= 0 {
				state.encoded = &strings.Builder{}
				state.encoded.WriteString(candidate[at+len(testfailure.Marker):])
				state.bytes = state.encoded.Len()
				state.tail = ""
			} else {
				state.tail = candidate[max(0, len(candidate)-(len(testfailure.Marker)-1)):]
			}
		} else if !state.overflow {
			state.bytes += len(piece)
			if state.bytes <= maxMarkerBytes {
				state.encoded.WriteString(piece)
			}
		}
		if state.bytes > maxMarkerBytes {
			state.overflow = true
			state.encoded.Reset()
		}
		if index < len(pieces)-1 {
			s.finishLine(state, true)
			state = &lineState{test: test}
		}
	}
	if state.encoded != nil || state.tail != "" {
		if len(s.partial) >= maxPartials {
			s.markers = append(s.markers, marker{test: test, malformed: true})
		} else {
			s.partial[key] = state
		}
	}
}

func (s *Stream) finishLine(state *lineState, terminated bool) {
	if state.encoded == nil {
		return
	}
	if !terminated || state.overflow {
		s.markers = append(s.markers, marker{test: state.test, malformed: true})
		return
	}
	envelope, err := testfailure.Decode([]byte(strings.TrimSpace(state.encoded.String())))
	s.markers = append(s.markers, marker{test: state.test, envelope: envelope, malformed: err != nil})
}

// Result is what one invocation's stream established.
type Result struct {
	Outcomes []Outcome
	// FailureDetails renders the bounded output of failed selected tests;
	// empty when none failed. It is diagnostic and must be redacted.
	FailureDetails string
	// Durations is the invocation's test durations record.
	Durations Durations
}

// Finish settles the stream. An event line over the contract bound means the
// outcomes are unknown and returns ErrEventTooLong.
func (s *Stream) Finish(packagePath string) (Result, error) {
	if len(s.line) > 0 && !s.tooLong {
		s.event(s.line)
		s.line = s.line[:0]
	}
	if s.err != nil {
		return Result{}, s.err
	}
	for _, state := range s.partial {
		s.finishLine(state, false)
	}
	clear(s.partial)
	outcomes, err := s.outcomes()
	if err != nil {
		return Result{}, err
	}
	details := s.details.render(s.order, s.terminal)
	if s.buildFailed {
		details = "[GO BUILD FAILURE] last 4 KiB of compiler output\n" + strings.ToValidUTF8(string(s.build), "\uFFFD") + "\n" + details
	}
	return Result{Outcomes: outcomes, FailureDetails: details, Durations: s.durations(packagePath)}, nil
}
