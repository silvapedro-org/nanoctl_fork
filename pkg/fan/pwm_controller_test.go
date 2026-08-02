package fan

import "testing"

func TestApplyDutyFloor(t *testing.T) {
	tests := []struct {
		name     string
		output   float64
		minDuty  float64
		offBelow float64
		want     float64
	}{
		// minDuty 0 disables the floor entirely — the original behaviour.
		{"disabled passes through", 12, 0, 0, 12},
		{"disabled passes through zero", 0, 0, 0, 0},

		// The stall band: anything the fan cannot actually turn at is lifted.
		{"inside stall band lifted to floor", 12, 30, 5, 30},
		{"just under floor lifted", 29.9, 30, 5, 30},
		{"at floor untouched", 30, 30, 5, 30},
		{"above floor untouched", 55, 30, 5, 55},
		{"full untouched", 100, 30, 5, 100},

		// Below the off threshold we stop rather than idle at the floor.
		{"at off threshold stops", 5, 30, 5, 0},
		{"below off threshold stops", 1, 30, 5, 0},
		{"zero stops", 0, 30, 5, 0},

		// offBelow 0 means only a genuine zero stops the fan.
		{"no off threshold keeps fan at floor", 0.1, 30, 0, 30},
		{"no off threshold stops at zero", 0, 30, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applyDutyFloor(tt.output, tt.minDuty, tt.offBelow); got != tt.want {
				t.Errorf("applyDutyFloor(%v, %v, %v) = %v, want %v",
					tt.output, tt.minDuty, tt.offBelow, got, tt.want)
			}
		})
	}
}
