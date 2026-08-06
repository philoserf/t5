// Package calendar implements the Traveller Imperial Calendar (Book 1
// Appendix 02, p.262): a 365-day year with no leap days, written as a
// day-of-year and a year. Day 1 of each year is Holiday; the remaining 364 days
// form 52 seven-day weeks.
package calendar

import "fmt"

// DaysPerYear is the fixed length of an Imperial year.
const DaysPerYear = 365

// weekdays names the seven days of an Imperial week, in order. Day 1 (Holiday)
// stands outside the weekly cycle, which begins on day 2 (Wonday).
var weekdays = [7]string{"Wonday", "Tuday", "Thirday", "Forday", "Fiday", "Sixday", "Senday"}

// A Date is a day within an Imperial year. Day is 1..365.
type Date struct {
	Year int
	Day  int
}

// New returns a Date, reporting an error if day is outside 1..365.
func New(year, day int) (Date, error) {
	if day < 1 || day > DaysPerYear {
		return Date{}, fmt.Errorf("calendar: day %d out of range 1..%d", day, DaysPerYear)
	}

	return Date{Year: year, Day: day}, nil
}

// Weekday returns the day's name: "Holiday" for day 1, otherwise one of the
// seven weekdays (day 2 is Wonday, day 8 is Senday, day 9 is Wonday again). A
// Day outside 1..365 renders as "?" rather than panicking, since Date's fields
// are exported and a caller may build one without New.
func (d Date) Weekday() string {
	switch {
	case !d.valid():
		return "?"
	case d.Day == 1:
		return "Holiday"
	default:
		return weekdays[(d.Day-2)%7]
	}
}

// String renders the date in day-year form, zero-padded to three digits, e.g.
// "001-1105". As with Weekday, a Day outside 1..365 renders its field as "???"
// rather than a well-formed-looking record ("000-1105", "400-1105"), since
// Date's fields are exported and a caller may build one without New.
func (d Date) String() string {
	if !d.valid() {
		return fmt.Sprintf("???-%d", d.Year)
	}

	return fmt.Sprintf("%03d-%d", d.Day, d.Year)
}

// Add returns the date days later (or earlier, when days is negative), rolling
// over 365-day years. A Date whose Day is outside 1..365 is returned unchanged:
// it names no day, so there is no day to count from, and normalizing it as a
// zero-based offset would launder a caller's bad literal into a valid-looking
// but different date — erasing the very state Weekday and String flag.
func (d Date) Add(days int) Date {
	if !d.valid() {
		return d
	}

	total := (d.Day - 1) + days // zero-based day index from the start of the year

	return Date{
		Year: d.Year + floorDiv(total, DaysPerYear),
		Day:  floorMod(total, DaysPerYear) + 1,
	}
}

// valid reports whether Day names a real day of the Imperial year. Date's fields
// are exported, so a caller may build one without New.
func (d Date) valid() bool { return d.Day >= 1 && d.Day <= DaysPerYear }

// Age returns the age in whole Imperial years of someone born in birthYear as of
// currentYear (Book 1: age is the difference in years).
func Age(birthYear, currentYear int) int {
	return currentYear - birthYear
}

// floorDiv and floorMod divide toward negative infinity so date arithmetic
// works for dates before the given one.
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}

	return q
}

func floorMod(a, b int) int {
	m := a % b
	if m != 0 && (m < 0) != (b < 0) {
		m += b
	}

	return m
}
