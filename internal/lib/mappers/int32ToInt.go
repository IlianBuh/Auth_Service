package mappers

func Int32ToInt(vals ...int32) []int {
	res := make([]int, len(vals))

	for i, v := range vals {
		res[i] = int(v)
	}

	return res
}
