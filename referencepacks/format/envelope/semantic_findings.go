package envelope

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
