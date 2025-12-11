package ical

import (
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/romanodesouza/galendario/internal/event"
)

type Calendar struct {
	cal *ics.Calendar
}

func NewCalendar(name string) *Calendar {
	cal := ics.NewCalendar()
	cal.SetName(name)

	return &Calendar{
		cal: cal,
	}
}

func (c *Calendar) AddEvents(events []event.Event) {
	for _, event := range events {
		ev := c.cal.AddEvent(icalUID(event))
		// Event has time confirmed
		if event.DateTime.Hour() != 0 {
			ev.SetStartAt(event.DateTime)
			ev.SetEndAt(event.DateTime.Add(2 * time.Hour))
		} else { // Event has no time confirmed - flag it as whole-day event
			ev.SetAllDayStartAt(event.DateTime)
		}
		ev.SetSummary(fmt.Sprintf("%s x %s", event.HomeTeam, event.AwayTeam))
		ev.SetLocation(event.Stadium)
		ev.SetDescription(event.Tournament)
		ev.SetDtStampTime(time.Now().UTC())
	}
}

func (c *Calendar) ICalEvents() []*ics.VEvent {
	return c.cal.Events()
}

func (c *Calendar) SerializeTo(w io.Writer) error {
	return c.cal.SerializeTo(w)
}

func icalUID(ev event.Event) string {
	seed := fmt.Sprintf("%d-%d-%d:%s:%s:%s",
		ev.DateTime.Year(),
		ev.DateTime.Month(),
		ev.DateTime.Day(),
		ev.Tournament,
		ev.HomeTeam,
		ev.AwayTeam,
	)
	h := sha256.New()
	h.Write([]byte(seed))
	return fmt.Sprintf("%x", h.Sum(nil))
}
