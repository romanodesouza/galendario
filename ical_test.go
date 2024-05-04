package galendario_test

import (
	"testing"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/google/go-cmp/cmp"
	"github.com/romanodesouza/galendario"
)

func TestAddEventsToIcal(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		events []galendario.Event
		want   []*ics.VEvent
	}{
		{
			name: "it should serialize a single event",
			events: []galendario.Event{
				{
					Tournament: "Copa Libertadores",
					Stadium:    "Gigante de Arroyito",
					DateTime:   time.Date(2024, 05, 7, 19, 0, 0, 0, loc),
					HomeTeam:   "Rosario Central",
					AwayTeam:   "Atlético",
				},
			},
			want: func() []*ics.VEvent {
				event := ics.NewEvent("test")
				event.SetStartAt(time.Date(2024, 05, 7, 19, 0, 0, 0, loc))
				event.SetEndAt(time.Date(2024, 05, 7, 19, 0, 0, 0, loc).Add(2 * time.Hour))
				event.SetSummary("Rosario Central vs Atlético")
				event.SetLocation("Gigante de Arroyito")
				event.SetDescription("Copa Libertadores")
				return []*ics.VEvent{event}
			}(),
		},
		{
			name: "it should serialize multiple events",
			events: []galendario.Event{
				{
					Tournament: "Copa Libertadores",
					Stadium:    "Campeón del Siglo",
					DateTime:   time.Date(2024, 05, 14, 19, 0, 0, 0, loc),
					HomeTeam:   "Peñarol",
					AwayTeam:   "Atlético",
				},
				{
					Tournament: "Copa Libertadores",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(2024, 05, 28, 19, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Caracas",
				},
			},
			want: func() []*ics.VEvent {
				event1 := ics.NewEvent("test")
				event1.SetStartAt(time.Date(2024, 05, 14, 19, 0, 0, 0, loc))
				event1.SetEndAt(time.Date(2024, 05, 14, 19, 0, 0, 0, loc).Add(2 * time.Hour))
				event1.SetSummary("Peñarol vs Atlético")
				event1.SetLocation("Campeón del Siglo")
				event1.SetDescription("Copa Libertadores")

				event2 := ics.NewEvent("test")
				event2.SetStartAt(time.Date(2024, 05, 28, 19, 0, 0, 0, loc))
				event2.SetEndAt(time.Date(2024, 05, 28, 19, 0, 0, 0, loc).Add(2 * time.Hour))
				event2.SetSummary("Atlético vs Caracas")
				event2.SetLocation("Arena MRV")
				event2.SetDescription("Copa Libertadores")

				return []*ics.VEvent{event1, event2}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cal := ics.NewCalendar()
			cal.SetName("Galendário")

			galendario.AddEventsToIcal(cal, tt.events)
			icalEvents := cal.Events()

			if len(icalEvents) != len(tt.want) {
				t.Fatalf("unexpected length of events, want %d, got %d", len(tt.want), len(icalEvents))
			}

			for i, event := range icalEvents {
				gotStartAt, _ := event.GetStartAt()
				wantStartAt, _ := tt.want[i].GetStartAt()

				if !wantStartAt.Equal(gotStartAt) {
					t.Errorf("unexpected event start time, want %v, got %v", wantStartAt, gotStartAt)
				}

				gotEndAt, _ := event.GetEndAt()
				wantEndAt, _ := tt.want[i].GetEndAt()

				if !wantEndAt.Equal(gotEndAt) {
					t.Errorf("unexpected event end time, want %v, got %v", wantEndAt, gotEndAt)
				}

				if diff := cmp.Diff(event.GetProperty(ics.ComponentPropertySummary).Value, tt.want[i].GetProperty(ics.ComponentPropertySummary).Value); diff != "" {
					t.Errorf("event summary mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(event.GetProperty(ics.ComponentPropertyLocation).Value, tt.want[i].GetProperty(ics.ComponentPropertyLocation).Value); diff != "" {
					t.Errorf("event location mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(event.GetProperty(ics.ComponentPropertyDescription).Value, tt.want[i].GetProperty(ics.ComponentPropertyDescription).Value); diff != "" {
					t.Errorf("event description mismatch (-want +got):\n%s", diff)
				}
			}

		})
	}
}
