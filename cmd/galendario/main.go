package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	resp, err := fetchPage(startDate, endDate)
	if err != nil {
		log.Fatal(err)
	}

	// Extract events
	events, err := event.ExtractEvents(resp.Body, loc)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	// Build calendar
	cal := ical.NewCalendar("Galendário")
	cal.AddEvents(events)

	// Print calendar
	if err := cal.SerializeTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

const baseURL = "https://www.atletico.com.br/futebol/agenda"

func fetchPage(startDate, endDate time.Time) (*http.Response, error) {
	body := url.Values{
		"data-inicio": []string{startDate.Format("02/01/2006")},
		"data-final":  []string{endDate.Format("02/01/2006")},
	}
	req, err := http.NewRequest("POST", baseURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("could not build POST request object for %s: %w", baseURL, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make POST request to %s: %w", baseURL, err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code from %s: %d", baseURL, resp.StatusCode)
	}

	return resp, nil
}

func startOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func endOfMonth(date time.Time) time.Time {
	firstDayOfNextMonth := startOfMonth(date).AddDate(0, 1, 0)
	return firstDayOfNextMonth.Add(-time.Second)
}
