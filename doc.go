// Package testequals provides expressive, low‑noise comparison helpers for
// JSON‑like Go values (typically produced by the jwalk package) tuned for
// regression tests. The core design choice is that object (document) matching
// defaults to SUBSET semantics: every key present in the expected value must
// exist and match in the actual value, while extra keys in the actual value are
// ignored. This keeps tests resilient to additive changes (e.g. new metadata
// fields) and focused on intent.
//
// Arrays and primitive values are always compared strictly (length + order for
// arrays; value equality for primitives). To opt into strict deep equality for
// an object subtree (rejecting unexpected keys) wrap the value with the Equal
// rule or, when unmarshalling JSON via jwalk, register and use the "$eq" rule
// token as demonstrated in examples/main.go.
//
// A simple usage pattern:
//
//	expected := jwalk.D{ {Key: "user", Value: jwalk.D{ {Key: "name", Value: "Alice"} } } }
//	actual   := obtainValueUnderTest()
//	if err := testequals.Test(expected, actual); err != nil { t.Fatal(err) }
//
// To aggregate all mismatches instead of failing fast, construct a Tester with
// the WithCollectAll option. The returned *MultiError exposes every mismatch
// with a stable path syntax (dot segments for object keys and [index] for array
// positions), which is friendly for snapshotting or golden file diffs.
//
// Performance notes:
//   - Small documents (length <= SmallDocLinearThreshold) are scanned linearly
//     for simplicity and lower allocation.
//   - Larger documents switch to a transient map (internally pooled) for O(1)
//     key lookups.
//
// Extensibility: Implement the Rule interface to inject custom comparison
// operators; inside Rule.Test you can delegate back to *Tester.Test to reuse
// subset / strict semantics. Return *MismatchError for single failures or
// *MultiError to contribute multiple path‑aware mismatches.
package testequals
