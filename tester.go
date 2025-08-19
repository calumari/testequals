package testequals

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/calumari/jwalk"
)

// Rule defines a pluggable comparison operator. Implementations receive the
// active Tester so they may delegate nested comparisons using existing subset /
// strict behavior. Return *MismatchError (single failure), *MultiError (many),
// or nil on success. Any other error value is converted into a path‑aware
// *MismatchError.
type Rule interface {
	Test(tester *Tester, actual any) error
}

// Tester performs comparisons between expected and actual values with subset
// semantics for object nodes (jwalk.D): every key present in the expected
// document must exist and match in the actual; additional keys in the actual
// document are ignored. Arrays (jwalk.A) and primitive values are strict. To
// enforce strict deep equality (rejecting extra object keys) for a subtree,
// wrap the expected value with an Equal Rule (or use the "$eq" JSON rule).
//
// A zero Tester must not be used; construct with New (or use the package level
// Default / Test helpers). A Tester is safe for concurrent use by multiple
// goroutines.
type Tester struct {
	cfg     Config
	mapPool sync.Pool
}

// Default is a shared Tester using DefaultConfig. It is safe for concurrent
// use. Mutating its Config is not supported; create a dedicated Tester via New
// for custom settings.
var Default = New()

// New constructs a Tester applying the provided Option values. Invalid option
// values (e.g. negative thresholds) are sanitized.
func New(opts ...Option) *Tester {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.SmallDocLinearThreshold < 0 {
		cfg.SmallDocLinearThreshold = 0
	}
	t := &Tester{cfg: cfg}
	t.mapPool.New = func() any { return make(map[string]any) }
	return t
}

// Test is a convenience wrapper that delegates to Default.Test.
func Test(expected, actual any) error { return Default.Test(expected, actual) }

// Test compares expected against actual using the Tester's semantics. On the
// first mismatch it returns a *MismatchError unless CollectAll is enabled, in
// which case all mismatches are aggregated and returned as *MultiError. The
// returned error is nil when actual satisfies (is a superset of) expected.
func (t *Tester) Test(expected, actual any) error {
	ctx := &cmpCtx{collect: t.cfg.CollectAll}
	if err := t.test(ctx, expected, actual); err != nil {
		return err
	}
	if ctx.collect && len(ctx.mismatches) > 0 {
		return &MultiError{Mismatches: ctx.mismatches}
	}
	return nil
}

func (t *Tester) test(ctx *cmpCtx, expected, actual any) error {
	switch exp := expected.(type) {
	case jwalk.D:
		actDoc, ok := actual.(jwalk.D)
		if !ok {
			return ctx.report(mismatch(ctx.path, fmt.Sprintf("expected jwalk.D, got %T", actual)))
		}
		return t.compareDocument(ctx, exp, actDoc)
	case jwalk.A:
		actArr, ok := actual.(jwalk.A)
		if !ok {
			return ctx.report(mismatch(ctx.path, fmt.Sprintf("expected jwalk.A, got %T", actual)))
		}
		return t.compareArray(ctx, exp, actArr)
	case Rule:
		if err := exp.Test(t, actual); err != nil {
			if merr, ok := err.(*MismatchError); ok {
				return ctx.report(mismatch(append(ctx.path, merr.Path...), merr.Message))
			}
			if multi, ok := err.(*MultiError); ok {
				for _, m := range multi.Mismatches {
					if r := ctx.report(mismatch(append(ctx.path, m.Path...), m.Message)); r != nil {
						return r
					}
				}
				return nil
			}
			return ctx.report(mismatch(ctx.path, err.Error()))
		}
		return nil
	default:
		if eq, handled, msg := fastPrimitiveEqual(expected, actual); handled {
			if !eq {
				return ctx.report(mismatch(ctx.path, msg))
			}
			return nil
		}
		if !reflect.DeepEqual(expected, actual) {
			return ctx.report(mismatch(ctx.path, fmt.Sprintf("expected (%T)%v, got (%T)%v", expected, expected, actual, actual)))
		}
		return nil
	}
}

func (t *Tester) compareDocument(ctx *cmpCtx, expected jwalk.D, actual jwalk.D) error {
	if len(expected) <= t.cfg.SmallDocLinearThreshold {
		for _, expEntry := range expected {
			found := false
			for _, actEntry := range actual {
				if expEntry.Key == actEntry.Key {
					found = true
					ctx.push(keySeg(expEntry.Key))
					err := t.test(ctx, expEntry.Value, actEntry.Value)
					ctx.pop()
					if err != nil {
						return err
					}
					break
				}
			}
			if !found {
				if err := ctx.reportAt(keySeg(expEntry.Key), "key not found"); err != nil {
					return err
				}
			}
		}
		return nil
	}
	m := t.mapPool.Get().(map[string]any)
	for k := range m {
		delete(m, k)
	}
	defer t.mapPool.Put(m)
	for _, entry := range actual {
		m[entry.Key] = entry.Value
	}
	for _, entry := range expected {
		actVal, exists := m[entry.Key]
		if !exists {
			if err := ctx.reportAt(keySeg(entry.Key), "key not found"); err != nil {
				return err
			}
			continue
		}
		ctx.push(keySeg(entry.Key))
		err := t.test(ctx, entry.Value, actVal)
		ctx.pop()
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Tester) compareArray(ctx *cmpCtx, expected jwalk.A, actual jwalk.A) error {
	if len(expected) != len(actual) {
		return ctx.report(mismatch(ctx.path, fmt.Sprintf("length mismatch: expected %d, got %d", len(expected), len(actual))))
	}
	for i := range expected {
		ctx.push(indexSeg(i))
		err := t.test(ctx, expected[i], actual[i])
		ctx.pop()
		if err != nil {
			return err
		}
	}
	return nil
}
