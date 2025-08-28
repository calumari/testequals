package testequals

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/calumari/jwalk"
)

var (
	TestEqualDirective              = jwalk.NewDirective("test.eq", unmarshalEqual)
	TestNotEqualDirective           = jwalk.NewDirective("test.ne", unmarshalNotEqual)
	TestNilDirective                = jwalk.NewDirective("test.nil", unmarshalNil(true))
	TestRequiredDirective           = jwalk.NewDirective("test.required", unmarshalRequired(true))
	TestAnyDirective                = jwalk.NewDirective("test.any", unmarshalAny)
	TestMatchStringDirective        = jwalk.NewDirective("test.regex", unmarshalMatchString)
	TestElementsMatchDirective      = jwalk.NewDirective("test.elementsMatch", unmarshalElementsMatch)
	TestLengthDirective             = jwalk.NewDirective("test.length", unmarshalLength)
	TestEmptyDirective              = jwalk.NewDirective("test.empty", unmarshalEmpty)
	TestLessThanDirective           = jwalk.NewDirective("test.lt", unmarshalLT(false))
	TestLessThanOrEqualDirective    = jwalk.NewDirective("test.lte", unmarshalLT(true))
	TestGreaterThanDirective        = jwalk.NewDirective("test.gt", unmarshalGT(false))
	TestGreaterThanOrEqualDirective = jwalk.NewDirective("test.gte", unmarshalGT(true))
	TestInDirective                 = jwalk.NewDirective("test.in", unmarshalIn)
	TestAndDirective                = jwalk.NewDirective("test.and", unmarshalAnd)
	TestOrDirective                 = jwalk.NewDirective("test.or", unmarshalOr)
	TestNorDirective                = jwalk.NewDirective("test.nor", unmarshalNor)
	TestNotDirective                = jwalk.NewDirective("test.not", unmarshalNot)
)

// Equal is a Rule that enforces strict deep equality (including object key
// exhaustiveness) for the subtree it wraps. It overrides the default subset
// semantics applied to object values by Tester. The typical JSON representation
// uses a "$eq" key (see examples/main.go) registered via jwalk.
type Equal struct{ expected any }

func (c *Equal) Test(rc *RuleContext, actual any) error {
	// Enforce strict deep equality (including object key set) while still
	// delegating nested comparisons (and directives) to the Tester. We cannot
	// simply call tester.Test(expected, actual) because Tester intentionally
	// applies subset semantics for documents. Instead we perform our own
	// document key set check, then use tester.Test for each value so nested
	// directives behave normally.
	switch exp := c.expected.(type) {
	case jwalk.Document:
		act, ok := actual.(jwalk.Document)
		if !ok {
			return mismatch(nil, fmt.Sprintf("expected jwalk.Document, got %T", actual))
		}
		// Build map of actual values for O(1) lookup and to detect extras.
		amap := make(map[string]any, len(act))
		for _, e := range act {
			amap[e.Key] = e.Value
		}
		for _, e := range exp {
			av, ok := amap[e.Key]
			if !ok {
				return mismatch([]string{keySeg(e.Key)}, "key not found")
			}
			// Compare the expected value against the actual using Tester semantics.
			if err := rc.Test(e.Value, av); err != nil {
				// err may be *MismatchError or *MultiError; Tester.Test already formats paths.
				return err
			}
			delete(amap, e.Key)
		}
		if len(amap) > 0 {
			// Report first extra key.
			for k := range amap { // deterministic enough for error reporting
				return mismatch([]string{keySeg(k)}, fmt.Sprintf("unexpected extra key %q (strict $eq)", k))
			}
		}
		return nil
	case jwalk.Array:
		// Arrays already strict in Tester, delegate.
		return rc.Test(exp, actual)
	default:
		// Primitive or directive-containing value: rely on Tester for deep equality.
		return rc.Test(exp, actual)
	}
}

var (
	ErrNil    = errors.New("expected nil")
	ErrNotNil = errors.New("expected non-nil")

	ErrImplicitNil    = errors.New("nil value requires explicit assertion")
	ErrImplicitNotNil = errors.New("non-nil value requires explicit assertion")
)

type Nil struct {
	// expected == true  => the "nil" directive type (NilDirective)
	// expected == false => the "not-nil" directive type (NotNilDirective)
	// wanted == true    => user explicitly asserted expectation
	// wanted == false   => user did NOT explicitly assert expectation
	expected bool
	wanted   bool
}

func (c *Nil) Test(rc *RuleContext, actual any) error {
	isNil := isNil(reflect.ValueOf(actual))

	// User did NOT explicitly assert (wanted == false):
	// If the value coincidentally matches the expectation, force explicitness.
	// If it does not match, we ignore (no enforcement).
	if !c.wanted {
		if isNil == c.expected {
			if isNil {
				return fmt.Errorf("value is nil but expectation not explicitly asserted ($nil true required): %w", ErrImplicitNil)
			}
			return fmt.Errorf("value is non-nil but expectation not explicitly asserted ($nil false required): %w", ErrImplicitNotNil)
		}
		return nil
	}

	// User explicitly asserted (wanted == true):
	// Enforce strictly; mismatch is an error.
	if isNil != c.expected {
		if c.expected {
			return fmt.Errorf("expected nil ($nil true), got non-nil (%T): %w", actual, ErrNil)
		}
		return fmt.Errorf("expected non-nil ($nil false), got nil: %w", ErrNotNil)
	}

	return nil
}

type Required struct{ want bool }

var ErrRequired = errors.New("required value missing")

func (c *Required) Test(rc *RuleContext, actual any) error {
	if !c.want {
		return nil // sanity check; should not happen
	}
	v := reflect.ValueOf(actual)
	if isNil(v) || isZero(v) {
		return fmt.Errorf("required value missing or zero (%T): %w", actual, ErrRequired)
	}
	return nil
}

func RequiredDirective(name string) *jwalk.Directive {
	return jwalk.NewDirective(name, unmarshalRequired(true))
}

type Any struct{}

func (c *Any) Test(rc *RuleContext, actual any) error {
	return nil
}

func AnyDirective(name string) *jwalk.Directive {
	return jwalk.NewDirective(name, unmarshalAny)
}

type MatchString struct{ re *regexp.Regexp }

func (c *MatchString) Test(rc *RuleContext, actual any) error {
	s, ok := actual.(string)
	if !ok {
		return fmt.Errorf("$regex expects string, got %T", actual)
	}
	if !c.re.MatchString(s) {
		return fmt.Errorf("string %q does not match pattern %q", s, c.re.String())
	}
	return nil
}

func MatchStringDirective(name string) *jwalk.Directive {
	return jwalk.NewDirective(name, unmarshalMatchString)
}

type ElementsMatch struct{ expected []any }

func (c *ElementsMatch) Test(rc *RuleContext, actual any) error {
	av := reflect.ValueOf(actual)
	if !isList(av) {
		return fmt.Errorf("$elementsMatch expects array/slice, got %T", actual)
	}
	al := av.Len()
	if al != len(c.expected) {
		return fmt.Errorf("$elementsMatch length mismatch: expected %d elements, got %d", len(c.expected), al)
	}
	used := make([]bool, al)
	for _, exp := range c.expected {
		matched := false
		for i := range al {
			if used[i] {
				continue
			}
			if err := rc.Test(exp, av.Index(i).Interface()); err == nil {
				used[i] = true
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("$elementsMatch could not find match for expected element %v", exp)
		}
	}
	return nil
}

type Length struct {
	eq  *int
	lt  *int
	lte *int
	gt  *int
	gte *int
}

func (c *Length) Test(rc *RuleContext, actual any) error {
	av := reflect.ValueOf(actual)
	if !isList(av) {
		return fmt.Errorf("$length expects array/slice, got %T", actual)
	}
	al := av.Len()

	if c.eq != nil && al != *c.eq {
		return fmt.Errorf("$length eq failed: got %d, expected == %d", al, *c.eq)
	}
	if c.lt != nil && (al >= *c.lt) {
		return fmt.Errorf("$length lt failed: got %d, expected < %d", al, *c.lt)
	}
	if c.lte != nil && (al > *c.lte) {
		return fmt.Errorf("$length lte failed: got %d, expected <= %d", al, *c.lte)
	}
	if c.gt != nil && (al <= *c.gt) {
		return fmt.Errorf("$length gt failed: got %d, expected > %d", al, *c.gt)
	}
	if c.gte != nil && (al < *c.gte) {
		return fmt.Errorf("$length gte failed: got %d, expected >= %d", al, *c.gte)
	}

	return nil
}

type Empty struct{ want bool }

// Adjust Empty.Test semantics:
// want == true  => require empty
// want == false => require non-empty
func (c *Empty) Test(rc *RuleContext, actual any) error {
	av := reflect.ValueOf(actual)
	empty := isEmpty(av)
	if c.want {
		if !empty {
			return fmt.Errorf("$empty true failed: value (%T)%v not empty", actual, actual)
		}
	} else {
		if empty {
			return fmt.Errorf("$empty false failed: value (%T)%v is empty", actual, actual)
		}
	}
	return nil
}

// NotEqual fails if actual deeply equals the expected value.
type NotEqual struct{ expected any }

func (c *NotEqual) Test(rc *RuleContext, actual any) error {
	handled, err := fastPrimitiveEqual(c.expected, actual)
	if err == nil {
		return fmt.Errorf("$ne failed: values are equal (%v)", c.expected)
	}
	if handled {
		return nil
	}
	if reflect.DeepEqual(c.expected, actual) {
		return fmt.Errorf("$ne failed: values are deeply equal (%v)", c.expected)
	}
	return nil
}

// Numeric comparator base type
type numericCompare struct {
	op   string
	ref  float64
	incl bool // for <= or >=
}

func (c *numericCompare) Test(rc *RuleContext, actual any) error {
	val, ok := toFloat64(actual)
	if !ok {
		return fmt.Errorf("$%s expects numeric value, got %T", c.op, actual)
	}
	switch c.op {
	case "lt":
		if !(val < c.ref || (c.incl && val == c.ref)) {
			if c.incl {
				return fmt.Errorf("$lte failed: got %v, expected <= %v", trimFloat(val), trimFloat(c.ref))
			}
			return fmt.Errorf("$lt failed: got %v, expected < %v", trimFloat(val), trimFloat(c.ref))
		}
	case "gt":
		if !(val > c.ref || (c.incl && val == c.ref)) {
			if c.incl {
				return fmt.Errorf("$gte failed: got %v, expected >= %v", trimFloat(val), trimFloat(c.ref))
			}
			return fmt.Errorf("$gt failed: got %v, expected > %v", trimFloat(val), trimFloat(c.ref))
		}
	default:
		return errors.New("unknown numeric comparator")
	}
	return nil
}

type InSet struct{ elems []any }

func (c *InSet) Test(rc *RuleContext, actual any) error {
	for _, e := range c.elems {
		if err := rc.Test(e, actual); err == nil {
			return nil
		}
	}
	return fmt.Errorf("$in failed: value %v not in %v", actual, c.elems)
}

type And struct{ rules []any }

type Or struct{ rules []any }

type Nor struct{ rules []any }

type Not struct{ rule any }

func (c *And) Test(rc *RuleContext, actual any) error {
	// $and requires all rules to pass. We can reuse the parent context because
	// failures always invalidate the whole operator; when collecting we append
	// each rule's mismatches directly.
	for _, r := range c.rules {
		if err := rc.Test(r, actual); err != nil {
			if !rc.inner.collect {
				return fmt.Errorf("$and failed: %w", err)
			}
		}
	}
	if rc.inner.collect && len(rc.inner.mismatches) > 0 {
		return &MultiError{Mismatches: rc.inner.mismatches}
	}
	return nil
}

func (c *Or) Test(rc *RuleContext, actual any) error {
	// $or succeeds if any rule passes. To avoid polluting the parent mismatch
	// list when one eventually passes, we evaluate each alternative in an
	// isolated child context and only merge on total failure.
	var firstErr error
	var failedChildren []*cmpCtx
	for _, r := range c.rules {
		childCtx := &cmpCtx{path: append([]string{}, rc.inner.path...), collect: rc.inner.collect}
		child := &RuleContext{runner: rc.runner, inner: childCtx}
		if err := child.Test(r, actual); err == nil {
			return nil // success, discard prior failures
		} else {
			if firstErr == nil {
				firstErr = err
			}
			failedChildren = append(failedChildren, childCtx)
		}
	}
	// All failed
	if rc.inner.collect {
		if len(failedChildren) == 0 {
			return errors.New("$or failed: no alternatives provided")
		}
		for _, ch := range failedChildren {
			rc.inner.mismatches = append(rc.inner.mismatches, ch.mismatches...)
		}
		return &MultiError{Mismatches: rc.inner.mismatches}
	}
	if firstErr == nil {
		return errors.New("$or failed: no alternatives provided")
	}
	return fmt.Errorf("$or failed: value did not satisfy any alternative; first error: %v", firstErr)
}

func (c *Nor) Test(rc *RuleContext, actual any) error {
	// $nor fails if any rule succeeds. We can short‑circuit immediately in
	// non‑collect mode. In collect mode we note all matching alternatives.
	matchedIdx := []int{}
	for i, r := range c.rules {
		if err := rc.Test(r, actual); err == nil {
			if !rc.inner.collect {
				return errors.New("$nor failed: value satisfied a forbidden alternative")
			}
			matchedIdx = append(matchedIdx, i)
		}
	}
	if rc.inner.collect && len(matchedIdx) > 0 {
		for _, i := range matchedIdx {
			rc.inner.mismatches = append(rc.inner.mismatches, mismatch(nil, fmt.Sprintf("$nor failed: alternative %d matched", i)))
		}
		return &MultiError{Mismatches: rc.inner.mismatches}
	}
	return nil
}

func (c *Not) Test(rc *RuleContext, actual any) error {
	if err := rc.Test(c.rule, actual); err == nil {
		return errors.New("$not failed: value matched negated condition")
	}
	return nil
}
