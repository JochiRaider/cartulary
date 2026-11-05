package graphprojection

// A budget is a counting destination for the canonical encoder. Its charge
// is the exact wire size, independent of the representation used by Go.
type resultBudgetV2 struct{ limit, remaining int }

func newResultBudgetV2(limit int) *resultBudgetV2 { return &resultBudgetV2{limit, limit} }
func (b *resultBudgetV2) exhausted() error {
	return resourceProjectionErrorV2("maximum_result_bytes", b.limit, b.limit+1)
}
func (b *resultBudgetV2) consume(n int) error {
	if n > b.remaining {
		return b.exhausted()
	}
	b.remaining -= n
	return nil
}
func (b *resultBudgetV2) WriteByte(byte) error { return b.consume(1) }
func (b *resultBudgetV2) WriteString(s string) (int, error) {
	if err := b.consume(len(s)); err != nil {
		return 0, err
	}
	return len(s), nil
}
func (b *resultBudgetV2) charge(v any) error {
	return writeCanonicalJSON(b, v)
}
func separatorBytesV2(previous int) int {
	if previous > 0 {
		return 1
	}
	return 0
}
func (b *resultBudgetV2) arrayEntry(v any, previous int) error {
	if err := b.consume(separatorBytesV2(previous)); err != nil {
		return err
	}
	return b.charge(v)
}
func (b *resultBudgetV2) mark() int {
	if b == nil {
		return 0
	}
	return b.remaining
}
func (b *resultBudgetV2) restore(mark int) {
	if b != nil {
		b.remaining = mark
	}
}

func (run *projectionWork) retainMember(m map[string]any, key string, v any) {
	if run.budget != nil {
		if err := run.budget.consume(len(canonicalJSONString(key)) + 1 + separatorBytesV2(len(m))); err != nil {
			run.issues.err = err
			return
		}
		if err := run.budget.charge(v); err != nil {
			run.issues.err = err
			return
		}
	}
	m[key] = v
}
