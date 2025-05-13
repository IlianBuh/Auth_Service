package mappers

func IntToInt32(vals ...int) []int32 {
	res := make([]int32, len(vals))

	for i, v := range vals {
		res[i] = int32(v)
	}

	return res
}
