package event_test

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/romanodesouza/galendario/internal/event"
)

func TestExtractEvents(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().In(loc)
	year := now.Year()

	tests := []struct {
		name    string
		input   string
		want    []event.Event
		wantErr error
	}{
		{
			name:  "it should extract all events in asc order",
			input: "agenda.html",
			want: []event.Event{
				{
					Tournament: "Campeonato Mineiro",
					Stadium:    "Mário Helênio",
					DateTime:   time.Date(year, 1, 19, 16, 0, 0, 0, loc),
					HomeTeam:   "Aymorés",
					AwayTeam:   "Atlético",
				},
				{
					Tournament: "Copa do Brasil",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 4, 30, 21, 30, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Sport",
				},
				{
					Tournament: "Copa do Brasil",
					Stadium:    "Arena Pernambuco",
					DateTime:   time.Date(year, 5, 22, 19, 0, 0, 0, loc),
					HomeTeam:   "Sport",
					AwayTeam:   "Atlético",
				},
			},
			wantErr: nil,
		},
		{
			name:  "it should extract all events in asc order, except the finished ones",
			input: "agenda_finished_unfinished.html",
			want: []event.Event{
				{
					Tournament: "Copa Libertadores",
					Stadium:    "Gigante de Arroyito",
					DateTime:   time.Date(year, 5, 7, 19, 0, 0, 0, loc),
					HomeTeam:   "Rosario Central",
					AwayTeam:   "Atlético",
				},
				{
					Tournament: "Copa Libertadores",
					Stadium:    "Campeón del Siglo",
					DateTime:   time.Date(year, 5, 14, 19, 0, 0, 0, loc),
					HomeTeam:   "Peñarol",
					AwayTeam:   "Atlético",
				},
				{
					Tournament: "Copa Libertadores",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 5, 28, 19, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Caracas",
				},
			},
			wantErr: nil,
		},
		{
			name:    "it should return unexpected input error when input is not agenda",
			input:   "404.html",
			want:    nil,
			wantErr: event.ErrUnexpectedInput,
		},
		{
			name:  "it should extract all events with date and time or date only",
			input: "agenda_dates_without_time.html",
			want: []event.Event{
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 5, 11, 21, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Grêmio",
				},
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 5, 19, 16, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Bahia",
				},
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 9, 29, 0, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Vitória",
				},
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Castelão",
					DateTime:   time.Date(year, 10, 5, 0, 0, 0, 0, loc),
					HomeTeam:   "Fortaleza",
					AwayTeam:   "Atlético",
				},
			},
			wantErr: nil,
		},
		{
			name:  "it should extract all events with date and time or date only, handling multiple time formats",
			input: "agenda_dates_without_time_multiple_formats.html",
			want: []event.Event{
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 5, 11, 21, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Grêmio",
				},
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 5, 19, 16, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Bahia",
				},
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 9, 29, 0, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Vitória",
				},
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Castelão",
					DateTime:   time.Date(year, 10, 5, 0, 0, 0, 0, loc),
					HomeTeam:   "Fortaleza",
					AwayTeam:   "Atlético",
				},
				{
					Tournament: "Copa Libertadores",
					Stadium:    "Nuevo Gasómetro",
					DateTime:   time.Date(year, 8, 13, 21, 30, 0, 0, loc),
					HomeTeam:   "San Lorenzo",
					AwayTeam:   "Atlético",
				},
			},
			wantErr: nil,
		},
		{
			name:  "it should extract all events with date and time or date only, handling 'a definir' time format",
			input: "agenda_a_definir_format.html",
			want: []event.Event{
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 5, 11, 21, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Grêmio",
				},
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 5, 19, 16, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Bahia",
				},
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(year, 9, 29, 0, 0, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Vitória",
				},
				{
					Tournament: "Campeonato Brasileiro",
					Stadium:    "Castelão",
					DateTime:   time.Date(year, 10, 5, 0, 0, 0, 0, loc),
					HomeTeam:   "Fortaleza",
					AwayTeam:   "Atlético",
				},
				{
					Tournament: "Copa Libertadores",
					Stadium:    "Nuevo Gasómetro",
					DateTime:   time.Date(year, 8, 13, 0, 0, 0, 0, loc),
					HomeTeam:   "San Lorenzo",
					AwayTeam:   "Atlético",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(fmt.Sprintf(".testdata/%s", tt.input))
			if err != nil {
				t.Fatal(err)
			}

			events, err := event.ExtractEvents(f, loc)
			_ = f.Close()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: expected %v, got %v", tt.wantErr, err)
			}

			if diff := cmp.Diff(tt.want, events); diff != "" {
				t.Errorf("ExtractEvents() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
