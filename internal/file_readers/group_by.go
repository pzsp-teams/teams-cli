package file_readers

// GroupBy groups items in a slice by a key function
func GroupBy[T any, K comparable](items []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T, len(items))
	for _, item := range items {
		key := keyFunc(item)
		values := result[key]
		values = append(values, item)
		result[key] = values
	}
	return result
}
