package kinematics

import (
	"math"
	"testing"
	"time"
)

func TestKilometersFromAU(t *testing.T) {
	cases := map[float64]float64{
		-1:  0,
		0:   0,
		0.2: 30_000_000,
		1:   150_000_000,
		10:  1_500_000_000,
	}

	for au, want := range cases {
		if got := KilometersFromAU(au); got != want {
			t.Errorf("KilometersFromAU(%v) = %v, want %v", au, got, want)
		}
	}
}

func TestLightDelay(t *testing.T) {
	cases := []struct {
		name       string
		distanceKM float64
		want       time.Duration
	}{
		{"negative", -1, 0},
		{"zero", 0, 0},
		{"orbit 0", 30_000_000, 100 * time.Second},
		{"one AU", 150_000_000, 500 * time.Second},
		{"orbit 7", 1_500_000_000, 5_000 * time.Second},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := LightDelay(c.distanceKM); got != c.want {
				t.Errorf("LightDelay(%v) = %v, want %v", c.distanceKM, got, c.want)
			}
		})
	}
}

func TestImpactTime(t *testing.T) {
	cases := []struct {
		name          string
		distanceKM    float64
		accelerationG float64
		want          time.Duration
	}{
		{"zero distance", 0, 1, 0},
		{"negative distance", -1, 1, 0},
		{"zero acceleration", 30_000_000, 0, 0},
		{"negative acceleration", 30_000_000, -1, 0},
		{"orbit 0 at 1G", 30_000_000, 1, 21 * time.Hour},
		{"orbit 0 at 2G", 30_000_000, 2, 15 * time.Hour},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertNear(t, ImpactTime(c.distanceKM, c.accelerationG), c.want, time.Hour)
		})
	}
}

func TestImpactTimeFormula(t *testing.T) {
	want := 21*time.Hour + 30*time.Minute + 59*time.Second + 666*time.Millisecond
	assertNear(t, ImpactTime(30_000_000, 1), want, time.Millisecond)
}

func TestStartStopTime(t *testing.T) {
	cases := []struct {
		name          string
		distanceKM    float64
		accelerationG float64
		want          time.Duration
	}{
		{"zero distance", 0, 1, 0},
		{"negative distance", -1, 1, 0},
		{"zero acceleration", 30_000_000, 0, 0},
		{"negative acceleration", 30_000_000, -1, 0},
		{"orbit 0 at 1G", 30_000_000, 1, 30 * time.Hour},
		{"orbit 0 at 2G", 30_000_000, 2, 21 * time.Hour},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertNear(t, StartStopTime(c.distanceKM, c.accelerationG), c.want, time.Hour)
		})
	}
}

func TestStartStopTimeFormula(t *testing.T) {
	want := 30*time.Hour + 25*time.Minute + 44*time.Second + 511*time.Millisecond
	assertNear(t, StartStopTime(30_000_000, 1), want, time.Millisecond)
}

func TestAccelerationScalesByInverseSquareRoot(t *testing.T) {
	const distanceKM = 150_000_000

	oneG := float64(StartStopTime(distanceKM, 1))
	fourG := float64(StartStopTime(distanceKM, 4))

	if ratio := oneG / fourG; math.Abs(ratio-2) > 1e-9 {
		t.Errorf("1G/4G time ratio = %v, want 2", ratio)
	}
}

func assertNear(t *testing.T, got, want, tolerance time.Duration) {
	t.Helper()

	difference := got - want
	if difference < 0 {
		difference = -difference
	}

	if difference > tolerance {
		t.Errorf("got %v, want %v (+/-%v)", got, want, tolerance)
	}
}
