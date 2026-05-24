package vptree

type Result struct {
	Idx      uint32
	Distance uint16
}

type ResultBuffer struct {
	items [5]Result
	len   int
}

type PriorityQueue struct {
	candidates        [5]Result
	count             int8
	farthestCandidate int8
}

func (pq *PriorityQueue) Add(idx uint32, dist uint16) {
	r := Result{Idx: idx, Distance: dist}
	if pq.count < 5 {
		pq.candidates[pq.count] = r
		if pq.count == 0 || r.Distance > pq.candidates[pq.farthestCandidate].Distance {
			pq.farthestCandidate = pq.count
		}
		pq.count++
		return
	}

	if r.Distance > pq.candidates[pq.farthestCandidate].Distance {
		return
	}

	pq.candidates[pq.farthestCandidate] = r
	max := pq.candidates[0].Distance
	var maxIdx int8
	for i := int8(1); i < 5; i++ {
		if pq.candidates[i].Distance > max {
			max = pq.candidates[i].Distance
			maxIdx = i
		}
	}
	pq.farthestCandidate = maxIdx
}

func (pq *PriorityQueue) max() uint16 {
	return pq.candidates[pq.farthestCandidate].Distance
}

// func (rb *ResultBuffer) tryAdd(idx uint32, dist uint16) {
// 	if rb.len < 5 {
// 		rb.items[rb.len] = Result{Idx: idx, Distance: dist}
// 		rb.len++
// 		rb.siftUp(rb.len - 1)
// 	} else if dist < rb.items[0].Distance {
// 		rb.items[0] = Result{Idx: idx, Distance: dist}
// 		rb.siftDown(0)
// 	}
// }
//
// func (rb *ResultBuffer) max() uint16 {
// 	if rb.len == 0 {
// 		return 65535
// 	}
// 	return rb.items[0].Distance
// }
//
// func (rb *ResultBuffer) siftUp(i int) {
// 	for i > 0 {
// 		parent := (i - 1) / 2
// 		if rb.items[parent].Distance >= rb.items[i].Distance {
// 			break
// 		}
// 		rb.items[parent], rb.items[i] = rb.items[i], rb.items[parent]
// 		i = parent
// 	}
// }
//
// func (rb *ResultBuffer) siftDown(i int) {
// 	for {
// 		left := 2*i + 1
// 		right := 2*i + 2
// 		largest := i
// 		if left < rb.len && rb.items[left].Distance > rb.items[largest].Distance {
// 			largest = left
// 		}
// 		if right < rb.len && rb.items[right].Distance > rb.items[largest].Distance {
// 			largest = right
// 		}
// 		if largest == i {
// 			break
// 		}
// 		rb.items[i], rb.items[largest] = rb.items[largest], rb.items[i]
// 		i = largest
// 	}
// }

func Search(tree *VPTree, query [14]int16) [5]int {
	var rb PriorityQueue
	searchAux(tree, 0, &query, &rb)

	var result [5]int
	for i := range result {
		result[i] = -1
	}
	for i := range 5 {
		result[i] = int(rb.candidates[i].Idx)
	}
	return result
}

func searchAux(tree *VPTree, idx uint32, query *[14]int16, rb *PriorityQueue) {
	node := tree.Nodes[idx]
	dist := calculateDistance(query, &node.Vec)

	rb.Add(idx, dist)

	if node.Left == noChildIdx && node.Right == noChildIdx {
		return
	}

	var diff uint16
	if dist > node.Threshold {
		diff = dist - node.Threshold
	} else {
		diff = node.Threshold - dist
	}

	if dist <= node.Threshold {
		if node.Left != noChildIdx {
			searchAux(tree, node.Left, query, rb)
		}

		if node.Right != noChildIdx && diff < rb.max() {
			searchAux(tree, node.Right, query, rb)
		}
	} else {
		if node.Right != noChildIdx {
			searchAux(tree, node.Right, query, rb)
		}
		if node.Left != noChildIdx && diff < rb.max() {
			searchAux(tree, node.Left, query, rb)
		}
	}
}
