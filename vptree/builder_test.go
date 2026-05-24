package vptree

import (
	_ "embed"
	"encoding/json"
	"testing"

	"rinha2026/model"
	"rinha2026/quantization"
)

//go:embed testdata/references.json
var referencesJSON []byte

func loadReferences(t *testing.T) []model.ReferenceQuantized {
	var refs []model.Reference
	if err := json.Unmarshal(referencesJSON, &refs); err != nil {
		t.Fatalf("failed to unmarshal references.json: %v", err)
	}
	return quantization.QuantizeReferences(refs)
}

func TestBuild(t *testing.T) {
	refs := loadReferences(t)

	t.Run("n_nodes", func(t *testing.T) {
		tree := Build(refs)
		if len(tree.Nodes) != len(refs) {
			t.Errorf("expected %d nodes, got %d", len(refs), len(tree.Nodes))
		}
	})

	t.Run("contains_all_vectors", func(t *testing.T) {
		tree := Build(refs)
		refSet := make(map[[14]int16]string)
		for _, r := range refs {
			refSet[r.Vector] = r.Label
		}

		for i, node := range tree.Nodes {
			label, ok := refSet[node.Vec]
			if !ok {
				t.Errorf("node %d: vector %v not found in references", i, node.Vec)
				continue
			}
			expectedLabel := false
			if label == "legit" {
				expectedLabel = true
			}
			if node.Label != expectedLabel {
				t.Errorf("node %d: expected Label=%v for label=%q, got Label=%v", i, expectedLabel, label, node.Label)
			}
		}
	})

	t.Run("leaf_nodes", func(t *testing.T) {
		tree := Build(refs)

		for i, node := range tree.Nodes {
			isLeaf := node.Left == noChildIdx && node.Right == noChildIdx
			isInternal := node.Left != noChildIdx || node.Right != noChildIdx

			if isLeaf && node.Threshold != 0 {
				t.Errorf("leaf node %d: expected Threshold=0, got %d", i, node.Threshold)
			}

			_ = isInternal // nós internos podem ter 1 ou 2 filhos (VP-tree não é AVL)
		}
	})

	t.Run("root_index", func(t *testing.T) {
		tree := Build(refs)
		if tree.Nodes[0].Left == noChildIdx && tree.Nodes[0].Right == noChildIdx && len(refs) > 1 {
			t.Error("root is a leaf but refs > 1")
		}
	})

	t.Run("thresholds_positive", func(t *testing.T) {
		if len(refs) <= 1 {
			t.Skip("not enough refs")
		}
		tree := Build(refs)
		for i, node := range tree.Nodes {
			if node.Left == noChildIdx && node.Right == noChildIdx {
				continue
			}
			if node.Threshold == 0 {
				t.Errorf("internal node %d: expected Threshold > 0, got 0", i)
			}
		}
	})
}
