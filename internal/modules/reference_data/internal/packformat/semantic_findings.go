package packformat

import "strconv"

// semanticChecks keeps validation and diagnostic enumeration on one path.
// The stream has already passed its closed schema, so independent field
// predicates remain meaningful after another semantic predicate fails.
type semanticChecks struct {
	path    string
	id      *string
	emit    FindingSink
	invalid bool
	err     error
}

func (c *semanticChecks) require(ok bool, suffix string) {
	if ok {
		return
	}
	c.invalid = true
	if c.emit != nil && c.err == nil {
		c.err = c.emit(Finding{Path: c.path + suffix, EntryID: c.id})
	}
}

func (c *semanticChecks) finish(code string) error {
	if c.err != nil {
		return c.err
	}
	if c.invalid {
		return fail(code)
	}
	return nil
}

func (c *semanticChecks) ordered(value any, suffix string, normalize func(string) string, valid func(string) bool) {
	previous := ""
	seen := map[string]bool{}
	for i, item := range value.([]any) {
		v := item.(string)
		key := normalize(v)
		order := key + "\x00" + v
		c.require(key != "" && !seen[key] && valid(v) && order > previous, suffix+"["+strconv.Itoa(i)+"]")
		seen[key] = true
		previous = order
	}
}

func (c *semanticChecks) usage(value any, suffix string) {
	for i, item := range value.([]any) {
		c.require(multiline(item.(string)), suffix+"["+strconv.Itoa(i)+"]")
	}
}
