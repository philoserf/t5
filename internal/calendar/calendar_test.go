package calendar

import "testing"

func TestWeekday(t *testing.T) {
	cases := map[int]string{
		1:   "Holiday",
		2:   "Wonday",
		3:   "Tuday",
		8:   "Senday",
		9:   "Wonday", // the cycle repeats after Senday
		15:  "Senday",
		365: "Senday", // the last day of the year
	}
	for day, want := range cases {
		if got := (Date{Year: 1105, Day: day}).Weekday(); got != want {
			t.Errorf("day %d Weekday = %q, want %q", day, got, want)
		}
	}
}

func TestString(t *testing.T) {
	if got := (Date{Year: 1105, Day: 1}).String(); got != "001-1105" {
		t.Errorf("String() = %q, want 001-1105", got)
	}
	if got := (Date{Year: 1105, Day: 365}).String(); got != "365-1105" {
		t.Errorf("String() = %q, want 365-1105", got)
	}
}

func TestNew(t *testing.T) {
	if _, err := New(1105, 1); err != nil {
		t.Errorf("New(1105, 1) errored: %v", err)
	}
	for _, day := range []int{0, -1, 366} {
		if _, err := New(1105, day); err == nil {
			t.Errorf("New(1105, %d) should error", day)
		}
	}
}

func TestAdd(t *testing.T) {
	cases := []struct {
		start Date
		days  int
		want  Date
	}{
		{Date{1105, 1}, 7, Date{1105, 8}},
		{Date{1105, 365}, 1, Date{1106, 1}},     // roll to next year
		{Date{1105, 1}, -1, Date{1104, 365}},    // roll back a year
		{Date{1105, 100}, 365, Date{1106, 100}}, // exactly one year later
		{Date{1105, 200}, -400, Date{1104, 165}},
	}
	for _, c := range cases {
		if got := c.start.Add(c.days); got != c.want {
			t.Errorf("%s.Add(%d) = %s, want %s", c.start, c.days, got, c.want)
		}
	}
}

func TestAddRoundTrip(t *testing.T) {
	// Adding then subtracting any offset returns the original date.
	start := Date{1105, 173}
	for _, n := range []int{1, 7, 365, 1000, -50, -800} {
		if got := start.Add(n).Add(-n); got != start {
			t.Errorf("Add(%d).Add(%d) = %s, want %s", n, -n, got, start)
		}
	}
}

func TestAge(t *testing.T) {
	if got := Age(1100, 1105); got != 5 {
		t.Errorf("Age(1100, 1105) = %d, want 5", got)
	}
}
