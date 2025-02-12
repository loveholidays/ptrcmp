# ptrcmp

This linter detects direct comparisons between basic pointer types (like *int, *string) since developers usually want to compare the underlying values rather than memory addresses.

## Run 

```bash
go run main.go ./example
```

## Why use this linter?

This linter helps prevent subtle bugs by detecting direct comparisons between basic pointer types (like *int, *string, etc.). Such comparisons check if two pointers reference the exact same memory address rather than comparing the underlying values, which is rarely the intended behavior in application code.

For example:
```go
var one *int
var two *int

if one == two {
    // linter should highlight as comparisons between two "basic" ptrs i.e *int == *int
}

if *one == *two {
    // linter should ignore as its int == int
}
```

While pointer comparisons have valid uses (particularly in low-level code), they're often a source of bugs when working with basic types where value comparison is typically desired. This linter helps developers catch these cases early and encourages more explicit code by requiring them to either:

- Use value comparison with dereferencing (*x == *y)
- Explicitly acknowledge pointer comparison is intended by disabling the lint warning
