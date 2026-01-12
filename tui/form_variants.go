package tui

import (
	"math/bits"

	"github.com/pzsp-teams/teams-cli/app"
)

// fieldInfo tracks a field's type and index for variant building
type fieldInfo struct {
	fieldType app.InputType
	index     int
}

// calculateVariants determines the variant layout based on flag conflicts
func calculateVariants(defs []app.FlagDef) [][]int {
	if len(defs) == 0 {
		return nil
	}

	adj := buildConflictGraphBits(defs)
	return findAllMISBits(adj)
}

func buildConflictGraphBits(defs []app.FlagDef) []uint64 {
	n := len(defs)
	if n > 64 {
		panic("bitset version supports up to 64 flags")
	}

	index := map[string]int{}
	for i := range defs {
		index[defs[i].Name] = i
	}

	adj := make([]uint64, n)

	for i := range defs {
		for _, name := range defs[i].ConflictsWith {
			if j, ok := index[name]; ok {
				adj[i] |= 1 << j
				adj[j] |= 1 << i
			}
		}
	}
	return adj
}

func findAllMISBits(adj []uint64) [][]int {
	n := len(adj)

	var result [][]int

	var dfs func(candidates, current, excluded uint64)

	dfs = func(candidates, current, excluded uint64) {
		if candidates == 0 {
			if excluded == 0 {
				result = append(result, bitsToSlice(current, n))
			}
			return
		}

		v := uint(bits.TrailingZeros64(candidates))
		vMask := uint64(1 << v)

		// Include v
		dfs(candidates&^adj[v]&^vMask,
			current|vMask,
			excluded&^adj[v])

		// Exclude v
		dfs(candidates&^vMask,
			current,
			excluded|vMask)
	}

	all := uint64(1<<n) - 1
	dfs(all, 0, 0)

	return result
}

func bitsToSlice(mask uint64, n int) []int {
	var out []int
	for i := range n {
		if mask&(1<<i) != 0 {
			out = append(out, i)
		}
	}
	return out
}

// buildVariants constructs variant groups from flag indices and field info
func buildVariants(def *app.CommandDef, fieldMap map[int]fieldInfo) [][]formField {
	variantsIdx := calculateVariants(def.Flags)

	var variants [][]formField
	for _, set := range variantsIdx {
		var variant []formField
		for _, flagIdx := range set {
			info := fieldMap[flagIdx]
			variant = append(variant, formField{
				fieldType: info.fieldType,
				index:     info.index,
				flagIndex: flagIdx,
			})
		}
		variants = append(variants, variant)
	}
	return variants
}
