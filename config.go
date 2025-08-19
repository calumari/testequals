package testequals

type Config struct {
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

func DefaultConfig() Config {
	return Config{
		SmallDocLinearThreshold: 8,
	}
}

type Option func(*Config)

// WithLinearScanThreshold sets the object size boundary (inclusive) at which
// the matching algorithm switches from a naïve nested loop to a pooled map.
// Lower values favor lower per‑comparison allocations for mid‑sized documents;
// higher values favor simplicity. A negative input is treated as 0.
func WithLinearScanThreshold(n int) Option {
	return func(c *Config) {
		c.SmallDocLinearThreshold = n
	}
}

// WithCollectAll enables aggregation of all mismatches. When set, Tester.Test
// returns a *MultiError whose Mismatches slice enumerates each individual
// *MismatchError in depth‑first traversal order.
func WithCollectAll() Option {
	return func(c *Config) {
		c.CollectAll = true
	}
}
