package vptree

import (
	"log"
	"math"
)

func calculateDistance(v1, v2 *[14]int16) uint16 {
	var sum int64
	for i := range len(v1) {
		d := int64(v1[i]) - int64(v2[i])
		sum += d * d
	}
	r := int64(math.Sqrt(float64(sum)))
	if r*r < sum {
		r++
	}
	if r > math.MaxUint16 {
		log.Println("Overflow")
	}
	return uint16(r)
}

type Distance struct {
	Idx      int
	Distance uint16
}

func cmpDistance(a, b Distance) int {
	if a.Distance < b.Distance {
		return -1
	}
	if a.Distance > b.Distance {
		return 1
	}
	return 0
}
