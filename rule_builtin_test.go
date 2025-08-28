package testequals

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/calumari/jwalk"
	"github.com/stretchr/testify/assert"
)

// fakeTester implements testRunner allowing isolation from real Tester logic in unit tests.
type fakeTester struct{ calls []struct{ exp, act any } }

func (f *fakeTester) Test(e, a any) error {
	f.calls = append(f.calls, struct{ exp, act any }{e, a})
	if r, ok := e.(Rule); ok {
		return r.Test(&RuleContext{runner: f, inner: &cmpCtx{}}, a)
	}
	if !reflect.DeepEqual(e, a) {
		return mismatch(nil, "values differ")
	}
	return nil
}

func newRC(r testRunner) *RuleContext { return &RuleContext{runner: r, inner: &cmpCtx{}} }

func TestEqualRule(t *testing.T) {
	t.Run("actual Document does not match expected type returns error", func(t *testing.T) {
		ft := &fakeTester{}
		r := &Equal{expected: jwalk.Document{{Key: "a", Value: 1}}}
		assert.Error(t, r.Test(newRC(ft), 5))
	})

	t.Run("Document exact succeeds", func(t *testing.T) {
		ft := &fakeTester{}
		r := &Equal{expected: jwalk.Document{{Key: "a", Value: 1}}}
		assert.NoError(t, r.Test(newRC(ft), jwalk.Document{{Key: "a", Value: 1}}))
	})

	t.Run("Document missing key returns error", func(t *testing.T) {
		ft := &fakeTester{}
		r := &Equal{expected: jwalk.Document{{Key: "a", Value: 1}, {Key: "x", Value: 2}}}
		assert.Error(t, r.Test(newRC(ft), jwalk.Document{{Key: "a", Value: 1}}))
	})

	t.Run("Document extra key returns error", func(t *testing.T) {
		ft := &fakeTester{}
		r := &Equal{expected: jwalk.Document{{Key: "a", Value: 1}}}
		assert.Error(t, r.Test(newRC(ft), jwalk.Document{{Key: "a", Value: 1}, {Key: "x", Value: 2}}))
	})

	t.Run("Document value mismatch returns error", func(t *testing.T) {
		ft := &fakeTester{}
		r := &Equal{expected: jwalk.Document{{Key: "a", Value: 1}}}
		assert.Error(t, r.Test(newRC(ft), jwalk.Document{{Key: "a", Value: 2}}))
	})

	t.Run("Array exact succeeds", func(t *testing.T) {
		ft := &fakeTester{}
		r := &Equal{expected: jwalk.Array{1, 2, 3}}
		assert.NoError(t, r.Test(newRC(ft), jwalk.Array{1, 2, 3}))
	})

	t.Run("Array length mismatch returns error", func(t *testing.T) {
		ft := &fakeTester{}
		r := &Equal{expected: jwalk.Array{1, 2, 3}}
		assert.Error(t, r.Test(newRC(ft), jwalk.Array{1, 2}))
	})

	t.Run("Array element mismatch returns error", func(t *testing.T) {
		ft := &fakeTester{}
		r := &Equal{expected: jwalk.Array{1, 2, 3}}
		assert.Error(t, r.Test(newRC(ft), jwalk.Array{1, 4, 3}))
	})

	t.Run("primitive exact succeeds", func(t *testing.T) {
		ft := &fakeTester{}
		r := &Equal{expected: 5}
		assert.NoError(t, r.Test(newRC(ft), 5))
	})
}

func TestNilRule(t *testing.T) {
	t.Run("implicit nil requires explicit assertion returns error", func(t *testing.T) {
		c := &Nil{expected: true, wanted: false}
		err := c.Test(newRC(&fakeTester{}), nil)
		assert.Error(t, err)
	})

	t.Run("implicit not nil requires explicit assertion returns error", func(t *testing.T) {
		c := &Nil{expected: false, wanted: false}
		err := c.Test(newRC(&fakeTester{}), 5)
		assert.Error(t, err)
	})

	t.Run("implicit nil with no assertion succeeds", func(t *testing.T) {
		c := &Nil{expected: true, wanted: false}
		err := c.Test(newRC(&fakeTester{}), 5)
		assert.NoError(t, err)
	})

	t.Run("explicit nil fails", func(t *testing.T) {
		c := &Nil{expected: true, wanted: true}
		err := c.Test(newRC(&fakeTester{}), 5)
		assert.Error(t, err)
	})

	t.Run("explicit nil succeeds", func(t *testing.T) {
		c := &Nil{expected: true, wanted: true}
		err := c.Test(newRC(&fakeTester{}), nil)
		assert.NoError(t, err)
	})

	t.Run("explicit not nil succeeds", func(t *testing.T) {
		c := &Nil{expected: false, wanted: true}
		err := c.Test(newRC(&fakeTester{}), 5)
		assert.NoError(t, err)
	})

	t.Run("explicit not nil fails", func(t *testing.T) {
		c := &Nil{expected: false, wanted: true}
		err := c.Test(newRC(&fakeTester{}), nil)
		assert.Error(t, err)
	})
}

func TestRequiredRule(t *testing.T) {
	t.Run("required non-zero fails when not expected", func(t *testing.T) {
		c := &Required{want: false}
		err := c.Test(newRC(&fakeTester{}), 5)
		assert.NoError(t, err)
	})

	t.Run("required non-zero succeeds", func(t *testing.T) {
		c := &Required{want: true}
		err := c.Test(newRC(&fakeTester{}), 5)
		assert.NoError(t, err)
	})

	t.Run("required zero returns error", func(t *testing.T) {
		c := &Required{want: true}
		err := c.Test(newRC(&fakeTester{}), 0)
		assert.Error(t, err)
	})
}

func TestAnyRule(t *testing.T) {
	t.Run("any succeeds", func(t *testing.T) {
		c := &Any{}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), struct{}{}))
	})
}

func TestMatchStringRule(t *testing.T) {
	t.Run("prefix match succeeds", func(t *testing.T) {
		re := &MatchString{re: regexp.MustCompile("^abc")}
		assert.NoError(t, re.Test(newRC(&fakeTester{}), "abcdef"))
	})
	t.Run("type mismatch returns error", func(t *testing.T) {
		re := &MatchString{re: regexp.MustCompile("^abc")}
		assert.Error(t, re.Test(newRC(&fakeTester{}), 123))
	})
	t.Run("pattern mismatch returns error", func(t *testing.T) {
		re := &MatchString{re: regexp.MustCompile("^abc")}
		assert.Error(t, re.Test(newRC(&fakeTester{}), "zzz"))
	})
}

func TestElementsMatchRule(t *testing.T) {
	t.Run("actual not array returns error", func(t *testing.T) {
		c := &ElementsMatch{expected: []any{1, 2, 3}}
		assert.Error(t, c.Test(newRC(&fakeTester{}), 5))
	})

	t.Run("unordered exact multiset succeeds", func(t *testing.T) {
		c := &ElementsMatch{expected: []any{1, 2, 3}}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), []int{3, 2, 1}))
	})

	t.Run("length mismatch returns error", func(t *testing.T) {
		c := &ElementsMatch{expected: []any{1, 2, 3}}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{1, 2}))
	})

	t.Run("missing element returns error", func(t *testing.T) {
		c := &ElementsMatch{expected: []any{1, 2, 3}}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 4}))
	})
}

func TestLengthRule(t *testing.T) {
	t.Run("actual not array returns error", func(t *testing.T) {
		c := &Length{eq: toPtr(3)}
		assert.Error(t, c.Test(newRC(&fakeTester{}), 5))
	})

	t.Run("exact length eq succeeds", func(t *testing.T) {
		c := &Length{eq: toPtr(3)}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 3}))
	})

	t.Run("eq mismatch returns error", func(t *testing.T) {
		c := &Length{eq: toPtr(3)}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{1, 2}))
	})

	t.Run("length lt succeeds", func(t *testing.T) {
		c := &Length{lt: toPtr(3)}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), []int{1, 2}))
	})

	t.Run("length lt equal returns error", func(t *testing.T) {
		c := &Length{lt: toPtr(3)}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 3}))
	})

	t.Run("length lte succeeds", func(t *testing.T) {
		c := &Length{lte: toPtr(3)}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 3}))
	})

	t.Run("length lte greater returns error", func(t *testing.T) {
		c := &Length{lte: toPtr(3)}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 3, 4}))
	})

	t.Run("length gt succeeds", func(t *testing.T) {
		c := &Length{gt: toPtr(3)}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 3, 4}))
	})

	t.Run("length gt equal returns error", func(t *testing.T) {
		c := &Length{gt: toPtr(3)}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 3}))
	})

	t.Run("length gte succeeds", func(t *testing.T) {
		c := &Length{gte: toPtr(3)}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 3}))
	})

	t.Run("length gte less returns error", func(t *testing.T) {
		c := &Length{gte: toPtr(3)}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{1, 2}))
	})

	t.Run("multiple constraints succeeds", func(t *testing.T) {
		c := &Length{gt: toPtr(2), lt: toPtr(5)}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 3}))
	})

	t.Run("multiple constraints returns error", func(t *testing.T) {
		c := &Length{gt: toPtr(2), lt: toPtr(5)}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{1, 2, 3, 4, 5}))
	})
}

func TestEmptyRule(t *testing.T) {
	t.Run("empty true succeeds", func(t *testing.T) {
		c := &Empty{want: true}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), []int{}))
	})

	t.Run("empty true on non-empty returns error", func(t *testing.T) {
		c := &Empty{want: true}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{1}))
	})

	t.Run("empty false succeeds", func(t *testing.T) {
		c := &Empty{want: false}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), []int{1}))
	})

	t.Run("empty false on empty returns error", func(t *testing.T) {
		c := &Empty{want: false}
		assert.Error(t, c.Test(newRC(&fakeTester{}), []int{}))
	})
}

func TestNotEqualRule(t *testing.T) {
	t.Run("different value succeeds", func(t *testing.T) {
		c := &NotEqual{expected: 5}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), 6))
	})

	t.Run("equal primitives returns error", func(t *testing.T) {
		c := &NotEqual{expected: 5}
		assert.Error(t, c.Test(newRC(&fakeTester{}), 5))
	})

	t.Run("equal Documents returns error", func(t *testing.T) {
		c := &NotEqual{expected: jwalk.Document{{Key: "a", Value: 1}}}
		assert.Error(t, c.Test(newRC(&fakeTester{}), jwalk.Document{{Key: "a", Value: 1}}))
	})
}

func TestLessThanRule(t *testing.T) {
	t.Run("actual not numeric returns error", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "lt", ref: 10}).Test(newRC(&fakeTester{}), "abc"))
	})

	t.Run("value less than passes", func(t *testing.T) {
		assert.NoError(t, (&numericCompare{op: "lt", ref: 10}).Test(newRC(&fakeTester{}), 5))
	})

	t.Run("equal value fails", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "lt", ref: 10}).Test(newRC(&fakeTester{}), 10))
	})

	t.Run("op not recognized returns error", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "unknown", ref: 10}).Test(newRC(&fakeTester{}), 5))
	})
}

func TestLessThanOrEqualRule(t *testing.T) {
	t.Run("actual not numeric returns error", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "lt", ref: 10, incl: true}).Test(newRC(&fakeTester{}), "abc"))
	})

	t.Run("less passes", func(t *testing.T) {
		assert.NoError(t, (&numericCompare{op: "lt", ref: 10, incl: true}).Test(newRC(&fakeTester{}), 5))
	})

	t.Run("equal passes", func(t *testing.T) {
		assert.NoError(t, (&numericCompare{op: "lt", ref: 10, incl: true}).Test(newRC(&fakeTester{}), 10))
	})

	t.Run("greater fails", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "lt", ref: 10, incl: true}).Test(newRC(&fakeTester{}), 11))
	})

	t.Run("op not recognized returns error", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "unknown", ref: 10, incl: true}).Test(newRC(&fakeTester{}), 5))
	})
}

func TestGreaterThanRule(t *testing.T) {
	t.Run("actual not numeric returns error", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "gt", ref: 10}).Test(newRC(&fakeTester{}), "abc"))
	})

	t.Run("greater passes", func(t *testing.T) {
		assert.NoError(t, (&numericCompare{op: "gt", ref: 10}).Test(newRC(&fakeTester{}), 11))
	})

	t.Run("equal fails", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "gt", ref: 10}).Test(newRC(&fakeTester{}), 10))
	})

	t.Run("op not recognized returns error", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "unknown", ref: 10}).Test(newRC(&fakeTester{}), 5))
	})
}

func TestGreaterThanOrEqualRule(t *testing.T) {
	t.Run("actual not numeric returns error", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "gt", ref: 10, incl: true}).Test(newRC(&fakeTester{}), "abc"))
	})

	t.Run("greater passes", func(t *testing.T) {
		assert.NoError(t, (&numericCompare{op: "gt", ref: 10, incl: true}).Test(newRC(&fakeTester{}), 11))
	})

	t.Run("equal passes", func(t *testing.T) {
		assert.NoError(t, (&numericCompare{op: "gt", ref: 10, incl: true}).Test(newRC(&fakeTester{}), 10))
	})

	t.Run("less fails", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "gt", ref: 10, incl: true}).Test(newRC(&fakeTester{}), 9))
	})

	t.Run("op not recognized returns error", func(t *testing.T) {
		assert.Error(t, (&numericCompare{op: "unknown", ref: 10, incl: true}).Test(newRC(&fakeTester{}), 5))
	})
}

func TestInRule(t *testing.T) {
	t.Run("membership succeeds", func(t *testing.T) {
		c := &InSet{elems: []any{1, 2, 3}}
		assert.NoError(t, c.Test(newRC(&fakeTester{}), 2))
	})

	t.Run("non-membership returns error", func(t *testing.T) {
		c := &InSet{elems: []any{1, 2, 3}}
		assert.Error(t, c.Test(newRC(&fakeTester{}), 5))
	})
}

func TestAndRule(t *testing.T) {
	t.Run("all pass", func(t *testing.T) {
		r := &And{rules: []any{&NotEqual{expected: 1}, &NotEqual{expected: 2}}}
		ft := &fakeTester{}
		assert.NoError(t, r.Test(newRC(ft), 3))
	})

	t.Run("one fails", func(t *testing.T) {
		r := &And{rules: []any{&NotEqual{expected: 1}, &NotEqual{expected: 2}}}
		ft := &fakeTester{}
		assert.Error(t, r.Test(newRC(ft), 2))
	})
}

func TestOrRule(t *testing.T) {
	t.Run("first passes", func(t *testing.T) {
		r := &Or{rules: []any{5, 6}}
		ft := &fakeTester{}
		assert.NoError(t, r.Test(newRC(ft), 5))
	})

	t.Run("second passes", func(t *testing.T) {
		r := &Or{rules: []any{5, 6}}
		ft := &fakeTester{}
		assert.NoError(t, r.Test(newRC(ft), 6))
	})

	t.Run("all fail", func(t *testing.T) {
		r := &Or{rules: []any{5, 6}}
		ft := &fakeTester{}
		assert.Error(t, r.Test(newRC(ft), 7))
	})
}

func TestNorRule(t *testing.T) {
	t.Run("none match passes", func(t *testing.T) {
		r := &Nor{rules: []any{5, 6}}
		ft := &fakeTester{}
		assert.NoError(t, r.Test(newRC(ft), 7))
	})

	t.Run("match fails", func(t *testing.T) {
		r := &Nor{rules: []any{5, 6}}
		ft := &fakeTester{}
		assert.Error(t, r.Test(newRC(ft), 5))
	})
}

func TestNotRule(t *testing.T) {
	t.Run("different passes", func(t *testing.T) {
		r := &Not{rule: 5}
		ft := &fakeTester{}
		assert.NoError(t, r.Test(newRC(ft), 7))
	})

	t.Run("equal fails", func(t *testing.T) {
		r := &Not{rule: 5}
		ft := &fakeTester{}
		assert.Error(t, r.Test(newRC(ft), 5))
	})
}

func toPtr(i int) *int { return &i }
