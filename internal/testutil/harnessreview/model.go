// Package harnessreview owns review admission, lifecycle and retained publication.
// Browser and image operations are private adapters, not session authorities.
package harnessreview

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

type failureMapping struct {
	Class, Reason string
	Exit          int
}
type Failure struct {
	Class     string  `json:"failure_class"`
	Reason    string  `json:"failure_reason"`
	Code      string  `json:"diagnostic_code"`
	Phase     *string `json:"phase"`
	Subject   *string `json:"subject_id"`
	Condition *string `json:"condition"`
	Recovery  *string `json:"recovery_id"`
}
type ReviewError struct {
	Record    Failure
	Exit      int
	Cause     error
	Secondary []Failure
}

func (e *ReviewError) Error() string { return e.Record.Code }
func (e *ReviewError) Unwrap() error { return e.Cause }
func pointer[T any](v T) *T          { return &v }
func failure(code string) *ReviewError {
	p, ok := failurePolicy[code]
	if !ok {
		panic("unknown review failure")
	}
	return &ReviewError{Record: Failure{Class: p.Class, Reason: p.Reason, Code: code}, Exit: p.Exit}
}
func contextual(code, phase, subject, condition, recovery string) *ReviewError {
	e := failure(code)
	e.Record.Phase = &phase
	e.Record.Subject = &subject
	e.Record.Condition = &condition
	e.Record.Recovery = &recovery
	return e
}
func normalized(err error, fallback string) *ReviewError {
	var review *ReviewError
	if errors.As(err, &review) {
		return review
	}
	e := failure(fallback)
	if fallback == "preparation_failed" {
		e = contextual(fallback, "prerequisites", "preparation_child", "unknown", "inspect_failure")
	}
	e.Cause = err
	return e
}
func (f Failure) validate() error {
	p, ok := failurePolicy[f.Code]
	if !ok || f.Class != p.Class || f.Reason != p.Reason {
		return harnessschema.ErrInvalid
	}
	fields := map[string]*string{"phase": f.Phase, "subject_id": f.Subject, "condition": f.Condition, "recovery_id": f.Recovery}
	count := 0
	for k, v := range fields {
		if v != nil {
			count++
			if !slices.Contains(preparationContext[k], *v) {
				return harnessschema.ErrInvalid
			}
		}
	}
	if count != 0 && count != 4 {
		return harnessschema.ErrInvalid
	}
	if slices.Contains([]string{"build_failed", "fixture_failed", "diagnostic_invalid", "preparation_failed"}, f.Code) && count != 4 {
		return harnessschema.ErrInvalid
	}
	return nil
}

type Counts struct {
	Images     int64 `json:"images"`
	Elements   int64 `json:"observed_elements"`
	Violations int64 `json:"axe_violations"`
	Incomplete int64 `json:"axe_incomplete"`
	Console    int64 `json:"console_errors"`
	Requests   int64 `json:"failed_requests"`
}

func (c *Counts) Add(v Counts) {
	c.Images += v.Images
	c.Elements += v.Elements
	c.Violations += v.Violations
	c.Incomplete += v.Incomplete
	c.Console += v.Console
	c.Requests += v.Requests
}

type Reference struct {
	Path      string `json:"path"`
	Bytes     int    `json:"bytes"`
	SHA256    string `json:"sha256"`
	MediaType string `json:"media_type"`
}
type PrivateReference struct {
	Kind string `json:"kind"`
	Path string `json:"absolute_path"`
}
type Result struct {
	Schema    string             `json:"schema_id"`
	Command   string             `json:"command_id"`
	Session   *string            `json:"session_id"`
	Operation *int               `json:"operation_id"`
	State     *string            `json:"state"`
	Epoch     *int               `json:"epoch"`
	Status    string             `json:"status"`
	Exit      int                `json:"exit_code"`
	Failures  []Failure          `json:"failures"`
	Receipt   *Reference         `json:"receipt"`
	Bundle    *string            `json:"bundle_id"`
	Private   []PrivateReference `json:"private_refs"`
}
type Receipt struct {
	Schema    string    `json:"schema_id"`
	Command   string    `json:"command_id"`
	Session   string    `json:"session_id"`
	Operation *int      `json:"operation_id"`
	Mode      string    `json:"mode"`
	State     string    `json:"state"`
	Status    string    `json:"status"`
	Exit      int       `json:"exit_code"`
	Started   string    `json:"started_at"`
	Finished  string    `json:"finished_at"`
	Duration  int64     `json:"duration_ms"`
	Failures  []Failure `json:"failures"`
	Counts    Counts    `json:"counts"`
	Bundle    *string   `json:"bundle_id"`
	Cleanup   string    `json:"cleanup"`
}
type Locator struct {
	Schema  string     `json:"schema_id"`
	Session string     `json:"session_id"`
	RunID   string     `json:"run_id"`
	Mode    string     `json:"mode"`
	State   string     `json:"state"`
	Created string     `json:"created_at"`
	Updated string     `json:"updated_at"`
	Receipt *Reference `json:"terminal_receipt"`
}

func commandID(command string) string {
	if strings.HasPrefix(command, "browser-design-review") {
		command = "ui-review"
	}
	return "cartulary.harness.command." + strings.ReplaceAll(command, "-", "_") + ".v1"
}
func result(command string) Result {
	return Result{Schema: "cartulary.ui_review_command_result.v2", Command: commandID(command), Status: "ok", Failures: []Failure{}, Private: []PrivateReference{}}
}
func errorResult(command string, err error) Result {
	v := result(command)
	e := normalized(err, "unsafe_artifact")
	v.Status = "error"
	v.Exit = e.Exit
	v.Failures = []Failure{e.Record}
	return v
}
func jsonDocument(schema string, value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	v, err := harnessschema.Decode(data)
	if err == nil {
		err = harnessschema.Validate(schema, v)
	}
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
func timestamp(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z") }
func appendFailure(values []Failure, f Failure) []Failure {
	b, _ := json.Marshal(f)
	for _, v := range values {
		a, _ := json.Marshal(v)
		if string(a) == string(b) {
			return values
		}
	}
	if len(values) >= LimitFailures {
		values[LimitFailures-1] = contextual("diagnostic_invalid", "cleanup", "preparation_child", "invalid_artifact", "inspect_failure").Record
		return values
	}
	return append(values, f)
}
