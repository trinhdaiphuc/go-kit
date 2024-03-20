package collection

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

// Find returns the smallest index i at which x == a[i],
// or len(a) if there is no such index.
func Find[T comparable](a []T, x T) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}

// Contains tells whether a contains x.
func Contains[T comparable](a []T, x T) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func MapToArray[T comparable](a map[string]T) []T {
	result := make([]T, 0)
	for _, ele := range a {
		result = append(result, ele)
	}
	return result
}

func ToArrayString[T comparable](a []T, fn func(ele T) string) []string {
	result := make([]string, len(a))
	for idx, ele := range a {
		result[idx] = fn(ele)
	}
	return result
}

func ToArrayNumber[T comparable, N Number](a []T, fn func(ele T) (N, error)) ([]N, error) {
	result := make([]N, len(a))
	for idx, ele := range a {
		n, err := fn(ele)
		if err != nil {
			return nil, err
		}
		result[idx] = n
	}
	return result, nil
}

func DeDuplicate[T comparable](a []T) []T {
	if len(a) == 0 {
		return []T{}
	}

	var (
		result  = make([]T, 0)
		tempMap = make(map[T]bool)
	)

	for _, value := range a {
		tempMap[value] = true
	}

	for key := range tempMap {
		result = append(result, key)
	}

	return result
}

func InRange[N Number](number, from, to N) bool {
	return number >= from && number <= to
}
