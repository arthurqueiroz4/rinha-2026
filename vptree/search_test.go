package vptree

import (
	_ "embed"
	"encoding/json"
	"testing"

	"rinha2026/model"
	"rinha2026/quantization"
)

//go:embed testdata/references.json
var searchReferencesJSON []byte

func loadSearchReferences(t *testing.T) []model.ReferenceQuantized {
	var refs []model.Reference
	if err := json.Unmarshal(searchReferencesJSON, &refs); err != nil {
		t.Fatalf("failed to unmarshal references.json: %v", err)
	}
	return quantization.QuantizeReferences(refs)
}

func countValid(result [5]int) int {
	c := 0
	for _, v := range result {
		if v >= 0 {
			c++
		}
	}
	return c
}

func TestSearch(t *testing.T) {
	refs := loadSearchReferences(t)
	tree := Build(refs)

	t.Run("k_5", func(t *testing.T) {
		query := refs[0].Vector
		result := Search(tree, query)
		if countValid(result) != 5 {
			t.Errorf("expected 5 results, got %d", countValid(result))
		}
	})

	t.Run("query_equals_vantage_point", func(t *testing.T) {
		query := refs[0].Vector
		result := Search(tree, query)
		dist := calculateDistance(&query, &tree.Nodes[result[0]].Vec)
		if dist != 0 {
			t.Errorf("expected distance 0, got %d for result %d", dist, result[0])
		}
	})

	t.Run("all_indices_valid", func(t *testing.T) {
		query := refs[0].Vector
		result := Search(tree, query)
		seen := make(map[int]bool)
		for _, idx := range result {
			if idx < 0 {
				continue
			}
			if idx >= len(refs) {
				t.Errorf("index %d out of range [0, %d)", idx, len(refs))
			}
			if seen[idx] {
				t.Errorf("duplicate index %d", idx)
			}
			seen[idx] = true
		}
		if len(seen) != 5 && len(refs) >= 5 {
			t.Errorf("expected 5 unique indices, got %d", len(seen))
		}
	})
}
