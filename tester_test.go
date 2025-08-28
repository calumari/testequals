package testequals

import (
	"testing"

	"github.com/calumari/jwalk"
	"github.com/stretchr/testify/assert"
)

type mockRule struct {
	err error
}

func (m *mockRule) Test(rc *RuleContext, actual any) error {
	return m.err
}

func TestTester_Test(t *testing.T) {
	t.Run("primitive values equal succeeds", func(t *testing.T) {
		tester := New()
		err := tester.Test(42, 42)
		assert.NoError(t, err)
	})

	t.Run("primitive values not equal returns error", func(t *testing.T) {
		tester := New()
		err := tester.Test(42, 43)
		assert.Error(t, err)
	})

	t.Run("nil values equal succeeds", func(t *testing.T) {
		tester := New()
		err := tester.Test(nil, nil)
		assert.NoError(t, err)
	})

	t.Run("nil vs non-nil returns error", func(t *testing.T) {
		tester := New()
		err := tester.Test(nil, 1)
		assert.Error(t, err)
	})

	t.Run("object subset linear scan succeeds", func(t *testing.T) {
		tester := New()
		exp := jwalk.Document{{Key: "a", Value: 1}}
		act := jwalk.Document{{Key: "a", Value: 1}, {Key: "b", Value: 2}}
		err := tester.Test(exp, act)
		assert.NoError(t, err)
	})

	t.Run("object linear scan missing key returns error", func(t *testing.T) {
		tester := New()
		exp := jwalk.Document{{Key: "a", Value: 1}, {Key: "b", Value: 2}}
		act := jwalk.Document{{Key: "a", Value: 1}}
		err := tester.Test(exp, act)
		assert.Error(t, err)
	})

	t.Run("object value linear scan mismatch returns error", func(t *testing.T) {
		tester := New()
		exp := jwalk.Document{{Key: "a", Value: 1}}
		act := jwalk.Document{{Key: "a", Value: 2}}
		err := tester.Test(exp, act)
		assert.Error(t, err)
	})

	t.Run("object subset succeeds", func(t *testing.T) {
		tester := New(WithLinearScanThreshold(0))
		exp := jwalk.Document{{Key: "a", Value: 1}}
		act := jwalk.Document{{Key: "a", Value: 1}, {Key: "b", Value: 2}}
		err := tester.Test(exp, act)
		assert.NoError(t, err)
	})

	t.Run("object missing key returns error", func(t *testing.T) {
		tester := New(WithLinearScanThreshold(0))
		exp := jwalk.Document{{Key: "a", Value: 1}, {Key: "b", Value: 2}}
		act := jwalk.Document{{Key: "a", Value: 1}}
		err := tester.Test(exp, act)
		assert.Error(t, err)
	})

	t.Run("object value mismatch returns error", func(t *testing.T) {
		tester := New(WithLinearScanThreshold(0))
		exp := jwalk.Document{{Key: "a", Value: 1}}
		act := jwalk.Document{{Key: "a", Value: 2}}
		err := tester.Test(exp, act)
		assert.Error(t, err)
	})

	t.Run("array equal succeeds", func(t *testing.T) {
		tester := New()
		exp := jwalk.Array{1, 2, 3}
		act := jwalk.Array{1, 2, 3}
		err := tester.Test(exp, act)
		assert.NoError(t, err)
	})

	t.Run("array length mismatch returns error", func(t *testing.T) {
		tester := New()
		exp := jwalk.Array{1, 2}
		act := jwalk.Array{1, 2, 3}
		err := tester.Test(exp, act)
		assert.Error(t, err)
	})

	t.Run("array value mismatch returns error", func(t *testing.T) {
		tester := New()
		exp := jwalk.Array{1, 2, 3}
		act := jwalk.Array{1, 2, 4}
		err := tester.Test(exp, act)
		assert.Error(t, err)
	})

	t.Run("expected document, got array returns error", func(t *testing.T) {
		tester := New()
		exp := jwalk.Document{{Key: "a", Value: 1}}
		act := jwalk.Array{1}
		err := tester.Test(exp, act)
		assert.Error(t, err)
	})

	t.Run("expected array, got document returns error", func(t *testing.T) {
		tester := New()
		exp := jwalk.Array{1}
		act := jwalk.Document{{Key: "a", Value: 1}}
		err := tester.Test(exp, act)
		assert.Error(t, err)
	})

	t.Run("Rule success returns nil", func(t *testing.T) {
		tester := New()
		rule := &mockRule{err: nil}
		err := tester.Test(rule, 123)
		assert.NoError(t, err)
	})

	t.Run("Rule returns error", func(t *testing.T) {
		tester := New()
		rule := &mockRule{err: assert.AnError}
		err := tester.Test(rule, 123)
		assert.Error(t, err)
	})
}
