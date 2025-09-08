package p

import "time"

type DateRange struct {
	Start, End time.Time
}

// TODO: expand as more use cases arise
func (r *DateRange) Format() string {
	return r.formatMonthShortName()
}

func (r *DateRange) Days() int {
	dur := r.End.Sub(r.Start)
	return int(dur.Hours() / 24)
}

func (r *DateRange) Hours() int {
	dur := r.End.Sub(r.Start)
	return int(dur.Hours())
}

func (r *DateRange) formatMonthShortName() string {
	if r.Start.After(r.End) {
		return ""
	}

	if r.Start.Year() < r.End.Year() {
		return Format(
			"%s %d - %s %d",
			r.Start.Format("Jan"),
			r.Start.Year(),
			r.End.Format("Jan"),
			r.End.Year(),
		)
	}

	return Format("%s - %s %d", r.Start.Format("Jan"), r.End.Format("Jan"), r.End.Year())
}
