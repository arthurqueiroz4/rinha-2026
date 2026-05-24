package vptree

import (
	"encoding/binary"
	"errors"
	"log"
	"os"
	"runtime"
	"unsafe"
)

func Load(filename string) (*VPTree, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.New("Error reading vptree.bin: " + err.Error())
	}

	if len(buf) < 4 {
		return nil, errors.New("file too small")
	}

	numNodes := binary.LittleEndian.Uint32(buf[0:4])
	const nodeSize = 39
	expectedLen := 4 + int(numNodes)*nodeSize
	if len(buf) != expectedLen {
		return nil, errors.New("corrupted file: size mismatch")
	}

	nodes := make([]Node, numNodes)
	for i := range nodes {
		offset := 4 + i*nodeSize
		nodes[i].Left = binary.LittleEndian.Uint32(buf[offset+0 : offset+4])
		nodes[i].Right = binary.LittleEndian.Uint32(buf[offset+4 : offset+8])
		for j := range nodes[i].Vec {
			nodes[i].Vec[j] = int16(binary.LittleEndian.Uint16(buf[offset+8+j*2 : offset+10+j*2]))
		}
		nodes[i].Label = buf[offset+36] != 0
		nodes[i].Threshold = binary.LittleEndian.Uint16(buf[offset+37 : offset+39])
	}

	size := int(unsafe.Sizeof(nodes[0])) * int(numNodes)
	sizeMB := float64(size) / (1024 * 1024)
	log.Printf("Loaded VPTree: %.2f MB (%d bytes)", sizeMB, size)

	buf = nil
	runtime.GC()
	return &VPTree{Nodes: nodes}, nil
}
