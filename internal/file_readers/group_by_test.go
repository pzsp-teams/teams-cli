package file_readers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupBy_EmptySlice_ReturnsEmptyMap(t *testing.T) {
	got := GroupBy([]int{}, func(v int) int { return v })
	require.NotNil(t, got)
	require.Len(t, got, 0)
}

func TestGroupBy_GroupsByKeyAndPreservesOrderWithinGroup(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6}

	got := GroupBy(items, func(v int) string {
		if v%2 == 0 {
			return "even"
		}
		return "odd"
	})

	require.ElementsMatch(t, []string{"odd", "even"}, keys(got))

	require.Equal(t, []int{1, 3, 5}, got["odd"])
	require.Equal(t, []int{2, 4, 6}, got["even"])
}

func TestGroupBy_AllItemsSameKey(t *testing.T) {
	items := []string{"a", "b", "c"}

	got := GroupBy(items, func(s string) int { return 1 })

	require.Len(t, got, 1)
	require.Equal(t, []string{"a", "b", "c"}, got[1])
}

func TestGroupBy_DoesNotMutateInput(t *testing.T) {
	orig := []int{1, 2, 3}
	items := append([]int(nil), orig...)

	_ = GroupBy(items, func(v int) int { return v % 2 })

	require.Equal(t, orig, items)
}

func keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
