# testequals

Subset-first comparison for JSON-like Go values, built on [jwalk](https://github.com/calumari/jwalk). Ideal for regression testing where stability and clarity matter.

## Why

In many real-world tests, the important part isn’t *everything* in the response. What matters is whether certain fields or values are present. Extra noise like timestamps, metadata, or IDs often gets in the way.

**testequals** makes subset comparison the default for objects so your regression tests stay stable, focused, and expressive. When strict checks matter, you can opt in explicitly to keep intent clear and noise low.

## Installation

```
go get github.com/calumari/testequals
```

## Quick Start

For a full runnable example demonstrating `$eq` and mismatch aggregation, see the [example](./examples/main.go).

## Core Semantics

| Type       | Semantics | Notes                                                                                    |
| ---------- | --------- | ---------------------------------------------------------------------------------------- |
| Objects    | Subset    | Every key in the expected object must exist and match in actual; extra keys are ignored. |
| Arrays     | Strict    | Length and element order must match exactly.                                             |
| Primitives | Strict    | Compared by value.                                                                       |

## Strict Segments with `$eq`

Use the `$eq` operator to require strict deep equality for a subtree in your expected document. Unlike the default subset check, `$eq` ensures there are no extra object keys.

Example expected fragment:
```jsonc
{
  "user": {
    "$eq": {
      "name": "Alice",
      "age": 30,
      "address": { "city": "Paris" }
    }
  }
}
```

If the actual value contains an extra field (e.g. `user.extra`), the failure will be reported as:

```
.user.extra: unexpected key present
```

## Custom Rules

You can implement the `Rule` interface to define custom comparison logic:

```go
type Rule interface { 
    Test(*testequals.Tester, any) error 
}
```

Return any of:

* `*MismatchError` – for a single mismatch
* `*MultiError` – for multiple mismatches
* any other `error` (will be wrapped into a path-aware mismatch)

Inside your rule, you can call `tester.Test` to reuse `testequals`’ core comparison logic.
