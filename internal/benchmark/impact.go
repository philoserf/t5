package benchmark

// SpeedEntry is one impact-speed benchmark (Book 1 p.37).
type SpeedEntry struct {
	Speed int
	// KPH is the printed kilometers-per-hour cell. The book leaves it blank
	// above Speed 16, so KPH 0 for Speeds 17-20 means "no printed kph", not
	// zero speed; only Speed 0 is genuinely 0 kph.
	KPH        int
	HitsPerTon int
	Descriptor string
}

var impactSpeeds = [...]SpeedEntry{
	{0, 0, 0, "Still"},
	{1, 5, 1, "Creep"},
	{2, 10, 4, "Xslow"},
	{3, 20, 9, "Slow"},
	{4, 30, 16, "Standard"},
	{5, 50, 25, "Cruise"},
	{6, 100, 36, "Fast"},
	{7, 300, 49, "Vfast"},
	{8, 500, 64, ""},
	{9, 700, 81, ""},
	{10, 1_000, 100, "Sonic"},
	{11, 2_000, 121, "Supersonic"},
	{12, 3_000, 144, "Hypersonic"},
	{13, 5_000, 169, ""},
	{14, 10_000, 196, ""},
	{15, 20_000, 225, ""},
	{16, 30_000, 256, "Meteoric"},
	{17, 0, 289, ""},
	{18, 0, 324, ""},
	{19, 0, 361, ""},
	{20, 0, 400, ""},
}

// ImpactSpeed returns the printed benchmark for Speed 0-20.
func ImpactSpeed(speed int) (SpeedEntry, bool) {
	if speed < 0 || speed >= len(impactSpeeds) {
		return SpeedEntry{}, false
	}

	return impactSpeeds[speed], true
}

// ImpactHits returns Speed squared times displacement tons. Fractional tons
// are retained; non-positive speed or tonnage produces no hits.
func ImpactHits(speed int, tons float64) float64 {
	if speed <= 0 || tons <= 0 {
		return 0
	}

	return float64(speed*speed) * tons
}
