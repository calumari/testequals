package testequals

// Equal is a Rule that enforces strict deep equality (including object key
// exhaustiveness) for the subtree it wraps. It overrides the default subset
// semantics applied to object values by Tester. The typical JSON representation
// uses a "$eq" key (see examples/main.go) registered via jwalk.
type Equal struct{ expected any }

func (c *Equal) Test(tester *Tester, actual any) error {
	return tester.Test(actual, c.expected)
}

func NewEqual(expected any) *Equal { return &Equal{expected} }
