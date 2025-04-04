package collection

func ArrayToMap[K, V comparable](a []V, fn func(ele V) K) map[K]V {
	result := make(map[K]V)
	for _, ele := range a {
		result[fn(ele)] = ele
	}
	return result
}
