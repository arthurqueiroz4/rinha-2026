package vptree

import (
	"encoding/binary"
	"errors"
	"os"
)

func Persist(vp *VPTree) error {
	if vp == nil {
		return errors.New("nil VPTree")
	}

	// Cada nó serializa em 39 bytes flat (sem padding, little-endian)
	// [0:4] Left uint32, [4:8] Right uint32, [8:36] Vec[14]int16,
	// [36] Label byte, [37:39] Threshold uint16
	const nodeSize = 39
	buf := make([]byte, 4+len(vp.Nodes)*nodeSize)

	binary.LittleEndian.PutUint32(buf[0:4], uint32(len(vp.Nodes)))

	for i, node := range vp.Nodes {
		offset := 4 + i*nodeSize
		binary.LittleEndian.PutUint32(buf[offset+0:offset+4], node.Left)
		binary.LittleEndian.PutUint32(buf[offset+4:offset+8], node.Right)
		for j, v := range node.Vec {
			binary.LittleEndian.PutUint16(buf[offset+8+j*2:offset+10+j*2], uint16(v))
		}
		if node.Label {
			buf[offset+36] = 1
		}
		binary.LittleEndian.PutUint16(buf[offset+37:offset+39], node.Threshold)
	}

	if err := os.WriteFile("resources/vptree.bin", buf, 0644); err != nil {
		return errors.New("Error writing vptree.bin: " + err.Error())
	}
	return nil
}
