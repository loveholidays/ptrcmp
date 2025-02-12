package example

func exampleComparison() {
	var one *int
	var two *int

	if one == two {
		// linter should highlight as comparisons between two "basic" ptrs i.e *int == *int
	}

	if *one == *two {
		// linter should ignore as its int == int
	}
}
