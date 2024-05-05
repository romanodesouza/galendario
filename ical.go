package galendario

import (
	"crypto/sha256"
	"fmt"
	"time"

	ics "github.com/arran4/golang-ical"
)

func AddEventsToIcal(cal *ics.Calendar, events []Event) {
	for _, event := range events {
		ev := cal.AddEvent(icalUID(event))
		// Event has time confirmed
		if event.DateTime.Hour() != 0 {
			ev.SetStartAt(event.DateTime)
			ev.SetEndAt(event.DateTime.Add(2 * time.Hour))
		} else { // Event has no time confirmed - flag it as whole day event
			ev.SetAllDayStartAt(event.DateTime)
		}
		ev.SetSummary(fmt.Sprintf("%s vs %s", event.HomeTeam, event.AwayTeam))
		ev.SetLocation(event.Stadium)
		ev.SetDescription(event.Tournament)
	}
}

func icalUID(event Event) string {
	seed := fmt.Sprintf("%d-%d-%d:%s:%s:%s",
		event.DateTime.Year(),
		event.DateTime.Month(),
		event.DateTime.Day(),
		event.Tournament,
		event.HomeTeam,
		event.AwayTeam,
	)
	h := sha256.New()
	h.Write([]byte(seed))
	return fmt.Sprintf("%x", h.Sum(nil))
}
