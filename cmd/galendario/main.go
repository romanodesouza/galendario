package main

import (
	"log"
	"os"
	"time"

	"github.com/romanodesouza/galendario/internal/event"
	"github.com/romanodesouza/galendario/internal/ical"
)

func main() {
	// Load location
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatal(err)
	}

	// Fetch events
	startDate := time.Now().In(loc)
	endDate := endOfMonth(startDate.AddDate(0, 3, 0))
	events, err := event.FetchAll(startDate, endDate)
	if err != nil {
		log.Fatal(err)
	}

	// Build calendar
	cal := ical.NewCalendar("Galendário")
	cal.AddEvents(events)

	// Print calendar
	if err := cal.SerializeTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func startOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func endOfMonth(date time.Time) time.Time {
	firstDayOfNextMonth := startOfMonth(date).AddDate(0, 1, 0)
	return firstDayOfNextMonth.Add(-time.Second)
}
