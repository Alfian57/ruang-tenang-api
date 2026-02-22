package model

import "testing"

func TestBreathingCycles_UsesFloorDivision(t *testing.T) {
	tech := &BreathingTechnique{
		InhaleDuration:     4,
		InhaleHoldDuration: 2,
		ExhaleDuration:     4,
		ExhaleHoldDuration: 2,
	}

	if got := tech.GetTotalCycleDuration(); got != 12 {
		t.Fatalf("expected total cycle duration 12, got %d", got)
	}

	if got := tech.GetCyclesForDuration(25); got != 2 {
		t.Fatalf("expected floor division result 2 cycles, got %d", got)
	}
}

func TestBreathingCycles_ZeroDurationReturnsZero(t *testing.T) {
	tech := &BreathingTechnique{}
	if got := tech.GetCyclesForDuration(1); got != 0 {
		t.Fatalf("expected 0 cycles for zero cycle duration, got %d", got)
	}
}
