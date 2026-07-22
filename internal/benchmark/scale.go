// Package benchmark provides Traveller's shared benchmark scales and formulas
// (Book 1 pp. 36-43).
package benchmark

// ScaleEntry is one level on the shared Fame, Danger, and Threat scale.
type ScaleEntry struct {
	Level       int
	Descriptor  string
	Alternative string
}

var scale = [...]ScaleEntry{
	{0, "Unknown", ""},
	{1, "Parent, Person", "1 person"},
	{2, "Close Family", "10 people; ship crew"},
	{3, "Extended Family", "100 people"},
	{4, "Neighborhood", "1,000 people"},
	{5, "Town", "10,000 people"},
	{6, "City", "100,000 people"},
	{7, "Large City", "One million people"},
	{8, "Regional", "Corporation"},
	{9, "Continental", ""},
	{10, "World", "Large Corporation"},
	{11, "World Complex", ""},
	{12, "World System", "Powerful Corporation"},
	{13, "Inner System", ""},
	{14, "System", ""},
	{15, "Greater System", ""},
	{16, "Outer System", ""},
	{17, "Systems", ""},
	{18, "Many Systems", ""},
	{19, "Subsector", ""},
	{20, "Sector", "MegaCorporation"},
	{21, "Domain", ""},
	{22, "Domains", ""},
	{23, "Many Domains", ""},
	{24, "Empire", ""},
	{25, "Beyond Empire", ""},
	{26, "Several Empires", ""},
	{27, "This Spiral Arm", ""},
	{28, "Many Spiral Arms", ""},
	{29, "The Galaxy", ""},
	{30, "Several Galaxies", ""},
	{31, "Many Galaxies", ""},
	{32, "The Universe", ""},
	{33, "Present Reality", ""},
	{34, "All Past Realities", ""},
	{35, "All Future Realities", ""},
	{36, "All Reality", ""},
}

// Scale returns one entry on the shared Fame, Danger, and Threat scale (Book 1
// p.36). Levels outside 0-36 have no benchmark.
func Scale(level int) (ScaleEntry, bool) {
	if level < 0 || level >= len(scale) {
		return ScaleEntry{}, false
	}

	return scale[level], true
}

// Risk combines the Probability, Severity, and Imminence Flux values of a
// danger (Book 1 p.36).
type Risk struct {
	Probability int
	Severity    int
	Imminence   int
}

// Level returns the sum of the risk's three components.
func (r Risk) Level() int { return r.Probability + r.Severity + r.Imminence }

// RequiresAction reports whether the risk is positive and must be addressed.
func (r Risk) RequiresAction() bool { return r.Level() > 0 }

// RiskDescriptors gives the meanings of a Flux value in the three Risk
// columns. Values outside -6..+6 have no entry.
func RiskDescriptors(flux int) (string, string, string, bool) {
	if flux < -6 || flux > 6 {
		return "", "", "", false
	}

	index := flux + 6

	return probabilityDescriptors[index], severityDescriptors[index], imminenceDescriptors[index], true
}

var probabilityDescriptors = [...]string{
	"Impossible", "Highly Improbable", "Improbable", "Highly Unlikely", "Unlikely", "Not Likely",
	"Either Way", "Possible", "Likely", "Probable", "Very Probable", "Almost Certain", "Certain",
}

var severityDescriptors = [...]string{
	"None", "Trivial", "Negligible", "Very Minor", "Minor", "Mild", "Temporary",
	"Strong", "Major", "Severe", "Very Severe", "Devastating", "Total",
}

var imminenceDescriptors = [...]string{
	"Far Future", "Centuries", "Lifetime", "Generation", "Decades", "Years", "Months",
	"Weeks", "Days", "Hours", "Minutes", "Seconds", "Now",
}
