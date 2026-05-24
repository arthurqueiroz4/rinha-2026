package quantization

import (
	"math"

	"rinha2026/model"
)

const scale float64 = 10000

func Quantize(value float64) int16 {
	rounded := math.Round(value * scale)
	if rounded > math.MaxInt16 {
		return math.MaxInt16
	}
	if rounded < math.MinInt16 {
		return math.MinInt16
	}
	return int16(rounded)
}

func QuantizeVector(v [14]float64) [14]int16 {
	var result [14]int16
	for i, val := range v {
		result[i] = Quantize(val)
	}
	return result
}

func QuantizeReference(ref model.Reference) model.ReferenceQuantized {
	return model.ReferenceQuantized{
		Vector: QuantizeVector(ref.Vector),
		Label:  ref.Label,
	}
}

func QuantizeReferences(refs []model.Reference) []model.ReferenceQuantized {
	result := make([]model.ReferenceQuantized, len(refs))
	for i, ref := range refs {
		result[i] = QuantizeReference(ref)
	}
	return result
}
