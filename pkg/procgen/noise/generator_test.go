package noise

import (
	"math"
	"testing"
)

func TestNoise2DDeterminism(t *testing.T) {
	seed := int64(12345)
	x, y := 1.5, 2.7

	v1 := Noise2D(x, y, seed)
	v2 := Noise2D(x, y, seed)

	if v1 != v2 {
		t.Errorf("noise should be deterministic: %f != %f", v1, v2)
	}
}

func TestNoise2DRange(t *testing.T) {
	seed := int64(99999)
	for i := 0; i < 1000; i++ {
		x := float64(i) * 0.1
		y := float64(i) * 0.13
		v := Noise2D(x, y, seed)
		if v < 0 || v > 1 {
			t.Errorf("noise value %f out of [0, 1] range at (%f, %f)", v, x, y)
		}
	}
}

func TestNoise2DSignedRange(t *testing.T) {
	seed := int64(99999)
	for i := 0; i < 1000; i++ {
		x := float64(i) * 0.1
		y := float64(i) * 0.13
		v := Noise2DSigned(x, y, seed)
		if v < -1 || v > 1 {
			t.Errorf("signed noise value %f out of [-1, 1] range at (%f, %f)", v, x, y)
		}
	}
}

func TestNoise2DDifferentSeeds(t *testing.T) {
	x, y := 5.0, 5.0
	v1 := Noise2D(x, y, 1)
	v2 := Noise2D(x, y, 2)

	if v1 == v2 {
		t.Error("different seeds should produce different noise values")
	}
}

func TestHashToFloatRange(t *testing.T) {
	seed := int64(42)
	for x := -100; x < 100; x++ {
		for y := -100; y < 100; y++ {
			v := HashToFloat(x, y, seed)
			if v < 0 || v > 1 {
				t.Errorf("hash value %f out of [0, 1] range at (%d, %d)", v, x, y)
			}
		}
	}
}

func TestSmoothstepBounds(t *testing.T) {
	// Smoothstep(0) should be 0
	if Smoothstep(0) != 0 {
		t.Errorf("smoothstep(0) should be 0, got %f", Smoothstep(0))
	}
	// Smoothstep(1) should be 1
	if Smoothstep(1) != 1 {
		t.Errorf("smoothstep(1) should be 1, got %f", Smoothstep(1))
	}
	// Smoothstep(0.5) should be 0.5
	if math.Abs(Smoothstep(0.5)-0.5) > 0.001 {
		t.Errorf("smoothstep(0.5) should be 0.5, got %f", Smoothstep(0.5))
	}
}

func TestLerp(t *testing.T) {
	if Lerp(0, 10, 0) != 0 {
		t.Error("lerp(0, 10, 0) should be 0")
	}
	if Lerp(0, 10, 1) != 10 {
		t.Error("lerp(0, 10, 1) should be 10")
	}
	if Lerp(0, 10, 0.5) != 5 {
		t.Error("lerp(0, 10, 0.5) should be 5")
	}
}
