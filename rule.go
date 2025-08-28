package testequals

// Rule defines a pluggable comparison operator. Implementations receive the
// active Tester so they may delegate nested comparisons using existing subset /
// strict behavior. Return *MismatchError (single failure), *MultiError (many),
// or nil on success. Any other error value is converted into a path‑aware
// *MismatchError.
type Rule interface {
	Test(rc *RuleContext, actual any) error
}

// RuleContext exposes a controlled surface of cmpCtx for rules. It permits
// adding mismatches, temporary path segment pushes, and invoking nested
// comparisons that inherit the current path and aggregation behavior.
type RuleContext struct {
	runner testRunner
	inner  *cmpCtx
}

// Add records a mismatch at the current path. Returns the mismatch error when
// aggregation is disabled so callers may bail out early; otherwise returns nil.
func (rc *RuleContext) Add(msg string) error {
	return rc.inner.report(mismatch(rc.inner.path, msg))
}

// PushKey appends an object key to the path; the returned function must be
// deferred to restore the previous path.
func (rc *RuleContext) PushKey(k string) (pop func()) {
	rc.inner.push(keySeg(k))
	return func() { rc.inner.pop() }
}

// PushIndex appends an array index to the path; the returned function must be
// deferred to restore the previous path.
func (rc *RuleContext) PushIndex(i int) (pop func()) {
	rc.inner.push(indexSeg(i))
	return func() { rc.inner.pop() }
}

// Test performs a nested comparison using the shared context so any mismatches
// are path‑aware and aggregated according to CollectAll.
func (rc *RuleContext) Test(expected, actual any) error { return rc.runner.Test(expected, actual) }

// testRunner allows mocking Tester in unit tests.
type testRunner interface {
	Test(expected, actual any) error
}
