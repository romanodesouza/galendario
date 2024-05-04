package galendario_test

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/romanodesouza/galendario"
)

func TestParseEvents(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		input   string
		want    []galendario.Event
		wantErr error
	}{
		{
			name:  "it should extract all events in asc order",
			input: "agenda.html",
			want: []galendario.Event{
				{
					Tournament: "Copa do Brasil",
					Stadium:    "Arena MRV",
					DateTime:   time.Date(2024, 04, 30, 21, 30, 0, 0, loc),
					HomeTeam:   "Atlético",
					AwayTeam:   "Sport",
				},
				{
					Tournament: "Copa do Brasil",
					Stadium:    "Arena Pernambuco",
					DateTime:   time.Date(2024, 05, 22, 19, 0, 0, 0, loc),
					HomeTeam:   "Sport",
					AwayTeam:   "Atlético",
				},
			},
			wantErr: nil,
		},
		{
			name:  "it should extract all events in asc order, except the finished ones",
			input: "agenda_finished_unfinished.html",
			want: []galendario.Event{
				{
					Tournament: "Copa Libertadores",
					Stadium:    "Gigante de Arroyito",
					DateTime:   time.Date(2024, 05, 7, 19, 0, 0, 0, loc),
					HomeTeam:   "Rosario Central",
					AwayTeam:   "Atlético",
				},
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
			wantErr: nil,
		},
		{
			name:    "it should return unexpected input error when input is not agenda",
			input:   "404.html",
			want:    nil,
			wantErr: galendario.ErrUnexpectedInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(fmt.Sprintf(".testdata/%s", tt.input))
			if err != nil {
				t.Fatal(err)
			}

			events, err := galendario.ParseEvents(f, loc)
			_ = f.Close()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: expected %v, got %v", tt.wantErr, err)
			}

			if diff := cmp.Diff(tt.want, events); diff != "" {
				t.Errorf("ParseEvents() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
