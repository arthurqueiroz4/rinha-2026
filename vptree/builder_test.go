package vptree

import (
	_ "embed"
	"encoding/json"
	"slices"
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

	t.Run("tree_is_connected", func(t *testing.T) {
		tree := Build(refs)
		visited := make([]bool, len(tree.Nodes))
		var walk func(uint32)
		walk = func(idx uint32) {
			if idx == noChildIdx || visited[idx] {
				return
			}
			visited[idx] = true
			walk(tree.Nodes[idx].Left)
			walk(tree.Nodes[idx].Right)
		}
		walk(0)

		for i, v := range visited {
			if !v {
				t.Errorf("node %d is not reachable from root", i)
			}
		}
	})

	t.Run("no_cycles", func(t *testing.T) {
		tree := Build(refs)
		visiting := make([]bool, len(tree.Nodes))
		var walk func(uint32)
		walk = func(idx uint32) {
			if idx == noChildIdx {
				return
			}
			if visiting[idx] {
				t.Errorf("cycle detected at node %d", idx)
				return
			}
			visiting[idx] = true
			walk(tree.Nodes[idx].Left)
			walk(tree.Nodes[idx].Right)
			visiting[idx] = false
		}
		walk(0)
	})

	t.Run("child_indices_valid", func(t *testing.T) {
		tree := Build(refs)
		n := uint32(len(tree.Nodes))
		for i, node := range tree.Nodes {
			if node.Left != noChildIdx && node.Left >= n {
				t.Errorf("node %d: Left=%d out of range [0, %d)", i, node.Left, n)
			}
			if node.Right != noChildIdx && node.Right >= n {
				t.Errorf("node %d: Right=%d out of range [0, %d)", i, node.Right, n)
			}
		}
	})

	t.Run("left_nodes_closer_than_threshold", func(t *testing.T) {
		tree := Build(refs)
		n := uint32(len(tree.Nodes))
		var collect func(uint32) []uint32
		collect = func(idx uint32) []uint32 {
			if idx == noChildIdx {
				return nil
			}
			node := tree.Nodes[idx]
			result := []uint32{idx}
			result = append(result, collect(node.Left)...)
			result = append(result, collect(node.Right)...)
			return result
		}

		for i := range n {
			node := tree.Nodes[i]
			if node.Left == noChildIdx {
				continue
			}
			leftIndices := collect(node.Left)
			for _, li := range leftIndices {
				dist := calculateDistance(&node.Vec, &tree.Nodes[li].Vec)
				if dist > node.Threshold {
					t.Errorf("node %d: left descendant %d has distance %d > threshold %d",
						i, li, dist, node.Threshold)
				}
			}
		}
	})

	t.Run("right_nodes_farther_than_threshold", func(t *testing.T) {
		tree := Build(refs)
		n := uint32(len(tree.Nodes))
		var collect func(uint32) []uint32
		collect = func(idx uint32) []uint32 {
			if idx == noChildIdx {
				return nil
			}
			node := tree.Nodes[idx]
			result := []uint32{idx}
			result = append(result, collect(node.Left)...)
			result = append(result, collect(node.Right)...)
			return result
		}

		for i := range n {
			node := tree.Nodes[i]
			if node.Right == noChildIdx {
				continue
			}
			rightIndices := collect(node.Right)
			for _, ri := range rightIndices {
				dist := calculateDistance(&node.Vec, &tree.Nodes[ri].Vec)
				if dist < node.Threshold {
					t.Errorf("node %d: right descendant %d has distance %d < threshold %d",
						i, ri, dist, node.Threshold)
				}
			}
		}
	})

	t.Run("search_returns_nearest_neighbors", func(t *testing.T) {
		tree := Build(refs)
		maxAllowedRatio := 2.0
		failures := 0

		for i := range refs {
			query := refs[i].Vector
			result := Search(tree, query)

			allDists := make([]struct {
				idx  int
				dist uint16
			}, len(refs))
			for j := range refs {
				allDists[j].idx = j
				allDists[j].dist = calculateDistance(&query, &refs[j].Vector)
			}
			slices.SortFunc(allDists, func(a, b struct {
				idx  int
				dist uint16
			}) int {
				if a.dist < b.dist {
					return -1
				}
				if a.dist > b.dist {
					return 1
				}
				return 0
			})

			resultDist := calculateDistance(&query, &tree.Nodes[result[4]].Vec)
			bruteDist := allDists[4].dist

			if bruteDist > 0 {
				ratio := float64(resultDist) / float64(bruteDist)
				if ratio > maxAllowedRatio {
					failures++
					t.Logf("query %d: search 5th neighbor dist=%d, brute=%d, ratio=%.2f",
						i, resultDist, bruteDist, ratio)
				}
			}
		}

		if failures > len(refs)/10 {
			t.Errorf("too many inaccurate queries: %d/%d (>10%%)", failures, len(refs))
		}
	})

	t.Run("single_element_tree", func(t *testing.T) {
		single := []model.ReferenceQuantized{
			{Vector: [14]int16{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400}, Label: "legit"},
		}
		tree := Build(single)
		if len(tree.Nodes) != 1 {
			t.Fatalf("expected 1 node, got %d", len(tree.Nodes))
		}
		node := tree.Nodes[0]
		if node.Left != noChildIdx || node.Right != noChildIdx {
			t.Errorf("single node should be a leaf")
		}
		result := Search(tree, single[0].Vector)
		if result[0] != 0 {
			t.Errorf("expected result[0]=0, got %d", result[0])
		}
	})

	t.Run("two_element_tree", func(t *testing.T) {
		pair := []model.ReferenceQuantized{
			{Vector: [14]int16{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400}, Label: "legit"},
			{Vector: [14]int16{9000, 9000, 9000, 9000, 9000, 9000, 9000, 9000, 9000, 9000, 9000, 9000, 9000, 9000}, Label: "fraud"},
		}
		tree := Build(pair)
		if len(tree.Nodes) != 2 {
			t.Fatalf("expected 2 nodes, got %d", len(tree.Nodes))
		}
		root := tree.Nodes[0]
		if root.Threshold == 0 {
			t.Error("root of 2-node tree should have non-zero threshold")
		}

		result := Search(tree, pair[0].Vector)
		if result[0] != 0 && (len(result) < 2 || result[1] != 0) {
			t.Errorf("query for refs[0] should find node 0, got %v", result)
		}
	})

	t.Run("synthetic_1000_elements", func(t *testing.T) {
		const n = 1000
		refs := make([]model.ReferenceQuantized, n)
		for i := range n {
			var vec [14]int16
			for d := range 14 {
				vec[d] = int16((i*7+d*13+5) % 9000)
			}
			label := "fraud"
			if i%3 == 0 {
				label = "legit"
			}
			refs[i] = model.ReferenceQuantized{Vector: vec, Label: label}
		}

		tree := Build(refs)

		t.Run("node_count", func(t *testing.T) {
			if len(tree.Nodes) != n {
				t.Errorf("expected %d nodes, got %d", n, len(tree.Nodes))
			}
		})

		t.Run("connected", func(t *testing.T) {
			visited := make([]bool, len(tree.Nodes))
			var walk func(uint32)
			walk = func(idx uint32) {
				if idx == noChildIdx || visited[idx] {
					return
				}
				visited[idx] = true
				walk(tree.Nodes[idx].Left)
				walk(tree.Nodes[idx].Right)
			}
			walk(0)
			unreachable := 0
			for i, v := range visited {
				if !v {
					unreachable++
					t.Logf("node %d unreachable", i)
				}
			}
			if unreachable > 0 {
				t.Errorf("%d/%d nodes unreachable from root", unreachable, n)
			}
		})

		t.Run("partitioning_correct", func(t *testing.T) {
			var collectSubtree func(uint32) []uint32
			collectSubtree = func(idx uint32) []uint32 {
				if idx == noChildIdx {
					return nil
				}
				ids := []uint32{idx}
				ids = append(ids, collectSubtree(tree.Nodes[idx].Left)...)
				ids = append(ids, collectSubtree(tree.Nodes[idx].Right)...)
				return ids
			}

			for i := uint32(0); i < uint32(n); i++ {
				node := tree.Nodes[i]
				if node.Left == noChildIdx && node.Right == noChildIdx {
					continue
				}
				for _, li := range collectSubtree(node.Left) {
					dist := calculateDistance(&node.Vec, &tree.Nodes[li].Vec)
					if dist > node.Threshold {
						t.Errorf("node %d: left desc %d dist=%d > threshold=%d",
							i, li, dist, node.Threshold)
						return
					}
				}
				for _, ri := range collectSubtree(node.Right) {
					dist := calculateDistance(&node.Vec, &tree.Nodes[ri].Vec)
					if dist < node.Threshold {
						t.Errorf("node %d: right desc %d dist=%d < threshold=%d",
							i, ri, dist, node.Threshold)
						return
					}
				}
			}
		})

		t.Run("search_exact_for_self", func(t *testing.T) {
			sampled := []int{0, n / 4, n / 2, 3 * n / 4, n - 1}
			for _, si := range sampled {
				query := refs[si].Vector
				result := Search(tree, query)
				found := false
				for _, idx := range result {
					if idx >= 0 {
						dist := calculateDistance(&query, &tree.Nodes[idx].Vec)
						if dist == 0 {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("query for refs[%d] did not find exact match (dist=0) in top 5", si)
				}
			}
		})
	})
}
