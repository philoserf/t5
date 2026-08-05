package task

import "testing"

func TestEvaluateMishap(t *testing.T) {
	for _, tc := range []struct {
		name          string
		kind          MishapKind
		flux, rating  int
		wantTotal     int
		wantTriggered bool
	}{
		{"ordinary uses reliability", OrdinaryMishap, -3, 1, -2, true},
		{"safe dangerous task", DangerousMishap, -1, 2, 1, false},
		{"unsafe destructive task", DestructiveMishap, -4, -2, -6, true},
		{"zero is not negative", DangerousMishap, -2, 2, 0, false},
	} {
		got := EvaluateMishap(tc.kind, tc.flux, tc.rating)
		if got.Total != tc.wantTotal || got.Triggered != tc.wantTriggered {
			t.Errorf(
				"%s: EvaluateMishap = %+v, want total %d triggered %v",
				tc.name, got, tc.wantTotal, tc.wantTriggered,
			)
		}
	}
}
