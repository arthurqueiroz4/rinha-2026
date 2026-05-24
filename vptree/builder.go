package vptree

import (
	"math/rand/v2"
	"slices"

	"rinha2026/model"
)

func Build(vectors []model.ReferenceQuantized) *VPTree {
	rand.Shuffle(len(vectors), func(i, j int) {
		vectors[i], vectors[j] = vectors[j], vectors[i]
	})
	nodes := make([]Node, len(vectors))
	buildAux(nodes, vectors, 0)
	return &VPTree{Nodes: nodes}
}

// buildAux constrói recursivamente a VP-tree.
func buildAux(nodes []Node, vectors []model.ReferenceQuantized, idx uint32) uint32 {
	if len(vectors) == 1 {
		nodes[idx] = newLeafNode(vectors[0])
		return idx + 1
	}

	vantagePoint := vectors[0]
	distances := make([]Distance, 0, len(vectors)-1)

	for i := 1; i < len(vectors); i++ {
		distance := calculateDistance(&vantagePoint.Vector, &vectors[i].Vector)
		distances = append(distances, Distance{Idx: i, Distance: distance})
	}

	slices.SortFunc(distances, cmpDistance)

	median := len(distances) / 2

	leftRaw := make([]model.ReferenceQuantized, median)
	for idx, distance := range distances[:median] {
		leftRaw[idx] = vectors[distance.Idx]
	}

	rightRaw := make([]model.ReferenceQuantized, len(distances)-median)
	for idx, distance := range distances[median:] {
		rightRaw[idx] = vectors[distance.Idx]
	}

	threshold := distances[median].Distance

	nextIdx := idx + 1

	var leftChild, rightChild = noChildIdx, noChildIdx
	if len(leftRaw) > 0 {
		leftChild = nextIdx
		nextIdx = buildAux(nodes, leftRaw, nextIdx)
	}
	if len(rightRaw) > 0 {
		rightChild = nextIdx
		nextIdx = buildAux(nodes, rightRaw, nextIdx)
	}

	nodes[idx] = newNode(
		uint32(leftChild), uint32(rightChild), vantagePoint, threshold,
	)
	return nextIdx
}
