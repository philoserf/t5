package shipcombat

import (
	"testing"

	"github.com/philoserf/t5/internal/shipgen"
)

func TestCanInterfere(t *testing.T) {
	for _, c := range []struct {
		rangeR int
		want   bool
	}{
		{-1, false},
		{0, true},
		{6, true},
		{7, true},
		{8, false},
	} {
		if got := CanInterfere(c.rangeR); got != c.want {
			t.Errorf("CanInterfere(R=%d) = %v, want %v", c.rangeR, got, c.want)
		}
	}
}

func TestInterferenceAgainst(t *testing.T) {
	for _, c := range []struct {
		name   string
		kind   AttackKind
		rangeR int
		want   InterferenceOptions
	}{
		{
			"missile", MissileAttack, 7,
			InterferenceOptions{AntiMissileFire: true, SubstituteTarget: true},
		},
		{
			"beam", BeamAttack, 7,
			InterferenceOptions{AntiBeamFire: true, SubstituteTarget: true},
		},
		{
			"other", OtherAttack, 7,
			InterferenceOptions{AvailableDefense: true, SubstituteTarget: true},
		},
		{"outside range", MissileAttack, 8, InterferenceOptions{}},
		{"unknown attack", AttackKind(99), 7, InterferenceOptions{}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := InterferenceAgainst(c.kind, c.rangeR); got != c.want {
				t.Errorf("InterferenceAgainst(%d, R=%d) = %+v, want %+v", c.kind, c.rangeR, got, c.want)
			}
		})
	}
}

func TestProtectedByDefense(t *testing.T) {
	for _, c := range []struct {
		name                     string
		separation, defenseRange int
		want                     bool
	}{
		{"inside", 5, 7, true},
		{"at limit", 7, 7, true},
		{"outside", 8, 7, false},
		{"invalid separation", -1, 7, false},
		{"invalid defense range", 1, -1, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := ProtectedByDefense(c.separation, c.defenseRange); got != c.want {
				t.Errorf("ProtectedByDefense(%d, %d) = %v, want %v", c.separation, c.defenseRange, got, c.want)
			}
		})
	}
}

func TestDecoyInterferes(t *testing.T) {
	for _, c := range []struct {
		name       string
		rangeR     int
		decoy      shipgen.Missile
		targetSize int
		want       bool
	}{
		{"one smaller", 7, missile(shipgen.Decoy, 4), 5, true},
		{"same size", 7, missile(shipgen.Decoy, 5), 5, true},
		{"too small", 7, missile(shipgen.Decoy, 3), 5, false},
		{"too far", 8, missile(shipgen.Decoy, 5), 5, false},
		{"not a decoy", 7, missile(shipgen.Explosive, 5), 5, false},
		{"invalid target size", 7, missile(shipgen.Decoy, 5), -1, false},
		{"invalid design", 7, shipgen.Missile{
			Spec:     shipgen.MissileSpec{Type: shipgen.Decoy, Size: 5},
			Problems: []shipgen.Problem{{Kind: shipgen.MissileRoundSizeMismatch}},
		}, 5, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := DecoyInterferes(c.rangeR, c.decoy, c.targetSize); got != c.want {
				t.Errorf("DecoyInterferes(R=%d, Size-%d, target Size-%d) = %v, want %v",
					c.rangeR, c.decoy.Spec.Size, c.targetSize, got, c.want)
			}
		})
	}
}

func missile(typ shipgen.MissileType, size int) shipgen.Missile {
	return shipgen.Missile{Spec: shipgen.MissileSpec{Type: typ, Size: size}}
}
