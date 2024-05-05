package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/romanodesouza/galendario"
)

func main() {
	// Load location
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatal(err)
	}

	// Fetch page
	resp, err := fetchPage(loc)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Extract events
	events, err := galendario.ExtractEvents(resp.Body, loc)
	if err != nil {
		log.Fatal(err)
	}

	// Build calendar
	cal := ics.NewCalendar()
	cal.SetName("Galendário")
	galendario.AddEventsToIcal(cal, events)

	// Print calendar
	cal.SerializeTo(os.Stdout)
}

const baseURL = "https://www.atletico.com.br/futebol/agenda"

func fetchPage(loc *time.Location) (*http.Response, error) {
	now := time.Now().In(loc)
	body := url.Values{
		"data-inicio": []string{now.Format("02/01/2006")},
	}
	req, err := http.NewRequest("POST", baseURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("could not build POST request object for %s: %w", baseURL, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make POST request to %s: %w", baseURL, err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code from %s: %d", baseURL, resp.StatusCode)
	}

	return resp, nil
}
