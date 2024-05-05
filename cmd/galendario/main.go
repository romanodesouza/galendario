package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/romanodesouza/galendario"
)

func main() {
	tournaments := flag.String("tournaments", "", "CSV tournament ids to fetch")
	flag.Parse()

	if tournaments == nil || *tournaments == "" {
		fmt.Fprintln(os.Stderr, "Please provide tournament ids to fetch")
		os.Exit(1)
	}

	ids := strings.Split(*tournaments, ",")
	rand.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})

	// Build calendar
	cal := ics.NewCalendar()
	cal.SetName("Galendário")

	// Location
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	chPages := make(chan io.ReadCloser)

	// Fetch pages
	wg.Add(1)
	go func() {
		defer close(chPages)
		defer wg.Done()

		for _, id := range ids {
			resp, err := fetchPage(id)
			if err != nil {
				log.Fatal(err)
			}
			chPages <- resp.Body
		}
	}()

	// Extract events
	wg.Add(1)
	go func() {
		defer wg.Done()

		for rc := range chPages {
			events, err := galendario.ExtractEvents(rc, loc)
			if err != nil {
				log.Fatal(err)
			}
			rc.Close()
			galendario.AddEventsToIcal(cal, events)
		}
	}()

	wg.Wait()

	// Output calendar
	cal.SerializeTo(os.Stdout)
}

const baseURL = "https://www.atletico.com.br/futebol/agenda"

func fetchPage(id string) (*http.Response, error) {
	body := url.Values{
		"filtro-campeonato": []string{id},
	}
	req, err := http.NewRequest("POST", baseURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("could not build POST request object for %s: %w", id, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could no make POST request to %s: %w", id, err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code from %s: %d", id, resp.StatusCode)
	}

	return resp, nil
}
