package vptree

import (
	"rinha2026/model"
)

const noChildIdx uint32 = 1<<32 - 1

type VPTree struct {
	Nodes []Node
}

type Node struct {
	Left      uint32
	Right     uint32
	Vec       [14]int16
	Label     bool
	Threshold uint16
}

func newNode(left, right uint32, ref model.ReferenceQuantized, threshold uint16) Node {
	label := ref.Label == "legit"
	return Node{
		Left:      left,
		Right:     right,
		Vec:       ref.Vector,
		Label:     label,
		Threshold: threshold,
	}
}

func newLeafNode(ref model.ReferenceQuantized) Node {
	return newNode(noChildIdx, noChildIdx, ref, 0)
}
