package convert

func Many[A, B any](convertFn func(A) B, args []A) []B {
	res := make([]B, len(args))
	for i, a := range args {
		res[i] = convertFn(a)
	}
	return res
}
