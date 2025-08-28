package testequals

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/calumari/jwalk"
)

type cmpCtx struct {
	path       []string
	collect    bool
	mismatches []*MismatchError
}

func (c *cmpCtx) report(m *MismatchError) error {
	if c.collect {
		c.mismatches = append(c.mismatches, m)
		return nil
	}
	return m
}

func (c *cmpCtx) reportAt(seg, msg string) error {
	c.push(seg)
	m := mismatch(c.path, msg)
	c.pop()
	return c.report(m)
}

func (c *cmpCtx) push(seg string) {
	c.path = append(c.path, seg)
}

func (c *cmpCtx) pop() {
	c.path = c.path[:len(c.path)-1]
}

type TesterOptions struct {
	// SmallDocLinearThreshold controls the size cutoff for using linear scan vs
	// map lookup when matching object (jwalk.D) values. For documents whose
	// number of fields is <= this value a simple nested loop is used (avoids a
	// transient map allocation and can be faster for tiny objects). Larger
	// documents use a pooled map to achieve O(1) lookups. Negative values are
	// coerced to 0 during Tester construction.
	SmallDocLinearThreshold int
	// CollectAll causes Tester.Test to aggregate all mismatches and return a
	// *MultiError instead of failing fast on the first *MismatchError.
	CollectAll bool
}

func DefaultConfig() TesterOptions {
	return TesterOptions{
		SmallDocLinearThreshold: 8,
	}
}

type TesterOption func(*TesterOptions)

// WithLinearScanThreshold sets the object size boundary (inclusive) at which
// the matching algorithm switches from a naïve nested loop to a pooled map.
// Lower values favor lower per‑comparison allocations for mid‑sized documents;
// higher values favor simplicity. A negative input is treated as 0.
func WithLinearScanThreshold(n int) TesterOption {
	return func(c *TesterOptions) {
		c.SmallDocLinearThreshold = n
	}
}

// WithCollectAll enables aggregation of all mismatches. When set, Tester.Test
// returns a *MultiError whose Mismatches slice enumerates each individual
// *MismatchError in depth‑first traversal order.
func WithCollectAll() TesterOption {
	return func(c *TesterOptions) {
		c.CollectAll = true
	}
}

// Tester performs comparisons between expected and actual values with subset
// semantics for object nodes (jwalk.Document): every key present in the expected
// document must exist and match in the actual; additional keys in the actual
// document are ignored. Arrays (jwalk.Array) and primitive values are strict. To
// enforce strict deep equality (rejecting extra object keys) for a subtree,
// wrap the expected value with an Equal Rule (or use the "$eq" JSON rule).
//
// A zero Tester must not be used; construct with New (or use the package level
// Default / Test helpers). A Tester is safe for concurrent use by multiple
// goroutines.
type Tester struct {
	options TesterOptions
	mapPool sync.Pool
}

// defaultTester is a shared Tester using DefaultConfig. It is safe for concurrent
// use. Mutating its Config is not supported; create a dedicated Tester via New
// for custom settings.
var defaultTester atomic.Pointer[Tester]

func init() {
	defaultTester.Store(New())
}

func DefaultTester() *Tester {
	return defaultTester.Load()
}

func SetDefaultTester(t *Tester) {
	defaultTester.Store(t)
}

// New constructs a Tester applying the provided Option values. Invalid option
// values (e.g. negative thresholds) are sanitized.
func New(opts ...TesterOption) *Tester {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.SmallDocLinearThreshold < 0 {
		cfg.SmallDocLinearThreshold = 0
	}
	t := &Tester{options: cfg}
	t.mapPool.New = func() any { return make(map[string]any) }
	return t
}

// Test is a convenience wrapper that delegates to Default.Test.
func Test(expected, actual any) error {
	return DefaultTester().Test(expected, actual)
}

// Test compares expected against actual using the Tester's semantics. On the
// first mismatch it returns a *MismatchError unless CollectAll is enabled, in
// which case all mismatches are aggregated and returned as *MultiError. The
// returned error is nil when actual satisfies (is a superset of) expected.
func (t *Tester) Test(expected, actual any) error {
	ctx := &cmpCtx{collect: t.options.CollectAll}
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
	case jwalk.Document:
		actDoc, ok := actual.(jwalk.Document)
		if !ok {
			return ctx.report(mismatch(ctx.path, fmt.Sprintf("expected jwalk.Document, got %T", actual)))
		}
		return t.compareDocument(ctx, exp, actDoc)
	case jwalk.Array:
		actArr, ok := actual.(jwalk.Array)
		if !ok {
			return ctx.report(mismatch(ctx.path, fmt.Sprintf("expected jwalk.Array, got %T", actual)))
		}
		return t.compareArray(ctx, exp, actArr)
	case Rule:
		if err := exp.Test(&RuleContext{runner: t, inner: ctx}, actual); err != nil {
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
		handled, err := fastPrimitiveEqual(expected, actual)
		if err != nil {
			return ctx.report(mismatch(ctx.path, err.Error()))
		}
		if handled {
			return nil
		}
		if !reflect.DeepEqual(expected, actual) {
			return ctx.report(mismatch(ctx.path, fmt.Sprintf("expected (%T)%v, got (%T)%v", expected, expected, actual, actual)))
		}
		return nil
	}
}

func (t *Tester) compareDocument(ctx *cmpCtx, expected jwalk.Document, actual jwalk.Document) error {
	if len(expected) <= t.options.SmallDocLinearThreshold {
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

func (t *Tester) compareArray(ctx *cmpCtx, expected jwalk.Array, actual jwalk.Array) error {
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
