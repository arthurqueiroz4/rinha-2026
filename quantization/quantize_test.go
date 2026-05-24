package quantization

import (
	"reflect"
	"testing"

	"rinha2026/model"
)

func TestQuantize(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected int16
	}{
		{"zero", 0, 0},
		{"positive_small", 0.01, 100},
		{"positive_fraction", 0.0833, 833},
		{"positive_one", 1, 10000},
		{"negative_one", -1, -10000},
		{"negative_small", -0.0001, -1},
		{"round_half", 0.00005, 1},
		{"round_down", 0.00004, 0},
		{"max_boundary", 3.2767, 32767},
		{"min_boundary", -3.2768, -32768},
		{"overflow", 10, 32767},
		{"underflow", -10, -32768},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Quantize(tt.input)
			if got != tt.expected {
				t.Errorf("Quantize(%v) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestQuantizeVector(t *testing.T) {
	input := [14]float64{
		0.01, 0.0833, 0.05, 0.8261, 0.1667, -1, -1,
		0.0432, 0.25, 0, 1, 0, 0.2, 0.0416,
	}
	expected := [14]int16{
		100, 833, 500, 8261, 1667, -10000, -10000,
		432, 2500, 0, 10000, 0, 2000, 416,
	}

	got := QuantizeVector(input)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("QuantizeVector mismatch:\ngot:      %v\nexpected: %v", got, expected)
	}
}

func TestQuantizeReference(t *testing.T) {
	input := model.Reference{
		Vector: [14]float64{0.01, 0.0833, 0.05, 0.8261, 0.1667, -1, -1, 0.0432, 0.25, 0, 1, 0, 0.2, 0.0416},
		Label:  "legit",
	}
	expected := model.ReferenceQuantized{
		Vector: [14]int16{100, 833, 500, 8261, 1667, -10000, -10000, 432, 2500, 0, 10000, 0, 2000, 416},
		Label:  "legit",
	}

	got := QuantizeReference(input)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("QuantizeReference mismatch:\ngot:      %+v\nexpected: %+v", got, expected)
	}
}

func TestQuantizeReferences(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := QuantizeReferences(nil)
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %v", got)
		}

		got = QuantizeReferences([]model.Reference{})
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %v", got)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		input := []model.Reference{
			{Vector: [14]float64{0.01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, Label: "legit"},
			{Vector: [14]float64{-1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, Label: "fraud"},
		}
		expected := []model.ReferenceQuantized{
			{Vector: [14]int16{100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, Label: "legit"},
			{Vector: [14]int16{-10000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, Label: "fraud"},
		}

		got := QuantizeReferences(input)
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("QuantizeReferences mismatch:\ngot:      %+v\nexpected: %+v", got, expected)
		}
	})
}
