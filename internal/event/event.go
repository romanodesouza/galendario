package event

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	baseURL = "https://www.atletico.com.br/futebol/agenda"
)

var (
	ErrUnexpectedInput = errors.New("unexpected html input")
)

type Event struct {
	Tournament string
	Stadium    string
	DateTime   time.Time
	HomeTeam   string
	AwayTeam   string
}

func (e *Event) AdjustYear(utcNow time.Time) {
	now := utcNow.In(e.DateTime.Location())

	if e.DateTime.Month() < now.Month() {
		e.DateTime = time.Date(e.DateTime.Year()+1, e.DateTime.Month(), e.DateTime.Day(),
			e.DateTime.Hour(), e.DateTime.Minute(), e.DateTime.Second(), e.DateTime.Nanosecond(), e.DateTime.Location())
	}
}

func FetchAll(startDate, endDate time.Time) ([]Event, error) {
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
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code from %s: %d", baseURL, resp.StatusCode)
	}

	events, err := ExtractEvents(resp.Body, startDate.Location())
	if err != nil {
		return nil, err
	}

	utcNow := time.Now().UTC()
	for _, event := range events {
		event.AdjustYear(utcNow)
	}

	return events, nil
}

func ExtractEvents(r io.Reader, loc *time.Location) ([]Event, error) {
	doc, err := htmlquery.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("ExtractEvents(): could not parse input: %w", err)
	}

	if !isCalendarPage(doc) {
		return nil, fmt.Errorf("ExtractEvents(): invalid html page: %w", ErrUnexpectedInput)
	}

	nodes, err := htmlquery.QueryAll(doc, `
	//section[@class="agenda-partidas"]
	//div[contains(concat(" ",normalize-space(@class)," ")," partida ") and not(contains(@class, "partida-finalizada"))]
	`)
	if err != nil {
		return nil, fmt.Errorf("ExtractEvents(): could not query main section node: %w", err)
	}

	events := make([]Event, len(nodes))
	for i, node := range nodes {
		var event Event

		div, err := htmlquery.Query(node, `//div[@class="partida-data"]`)
		switch {
		case err != nil:
			return nil, fmt.Errorf("ExtractEvents(): could not query 'div.partida-data': %w", err)
		case div == nil:
			return nil, fmt.Errorf("ExtractEvents(): missing expected 'div.partida-data' node: %w", ErrUnexpectedInput)
		}

		stadium := htmlquery.InnerText(div.LastChild)
		event.Stadium = normalizeStadium(stadium)

		dateTime := htmlquery.InnerText(div.FirstChild.NextSibling)
		event.DateTime = parseDateTime(dateTime, loc)
		if event.DateTime.IsZero() {
			return nil, fmt.Errorf(`ExtractEvents(): unexpected date/time format: "%s": %w`, dateTime, ErrUnexpectedInput)
		}

		div, err = htmlquery.Query(node, `//div[@class="partida-campeonato"]`)
		switch {
		case err != nil:
			return nil, fmt.Errorf("ExtractEvents(): could not query 'div.partida-campeonato': %w", err)
		case div == nil:
			return nil, fmt.Errorf("ExtractEvents(): missing expected 'div.partida-campeonato' node: %w", ErrUnexpectedInput)
		}
		event.Tournament = normalizeTournament(htmlquery.InnerText(div))

		abbr, err := htmlquery.QueryAll(node, `//div[@class="partida-placar"]//abbr[@title]`)
		switch {
		case err != nil:
			return nil, fmt.Errorf("ExtractEvents(): could not query 'abbr[@title]': %w", err)
		case abbr == nil:
			return nil, fmt.Errorf("ExtractEvents(): missing expected 'abbr[@title]' nodes: %w", ErrUnexpectedInput)
		case len(abbr) != 2:
			return nil, fmt.Errorf("ExtractEvents(): missing expected 2 'abbr[@title]' nodes: %w", ErrUnexpectedInput)
		}
		event.HomeTeam = normalizeTeam(htmlquery.SelectAttr(abbr[0], "title"))
		event.AwayTeam = normalizeTeam(htmlquery.SelectAttr(abbr[1], "title"))

		events[len(events)-(i+1)] = event
	}

	return events, nil
}

func isCalendarPage(doc *html.Node) bool {
	node, _ := htmlquery.Query(doc, "//title")
	title := strings.ToLower(htmlquery.InnerText(node))
	return strings.HasPrefix(title, "calendário de jogos")
}

func parseDateTime(input string, loc *time.Location) time.Time {
	input = strings.ToLower(input)
	now := time.Now().In(loc)
	// Date and time (21:00 format)
	t, err := time.ParseInLocation("02/01 às 15:04-2006", fmt.Sprintf("%s-%d", input, now.Year()), loc)
	// Date and time (21h format)
	if err != nil {
		t, err = time.ParseInLocation("02/01 às 15h04-2006", fmt.Sprintf("%s-%d", input, now.Year()), loc)
	}
	// Date only
	if err != nil {
		t, err = time.ParseInLocation("02/01-2006", fmt.Sprintf("%s-%d", input, now.Year()), loc)
	}
	// Date only ("a definir" format)
	if err != nil {
		t, _ = time.ParseInLocation("02/01 às a definir-2006", fmt.Sprintf("%s-%d", input, now.Year()), loc)
	}
	return t
}

func normalizeStadium(input string) string {
	return strings.TrimSpace(input)
}

func normalizeTournament(input string) string {
	input = strings.TrimSpace(input)
	tournament := strings.ToLower(input)
	switch {
	case strings.Contains(tournament, "libertadores"):
		return "Libertadores"
	case strings.Contains(tournament, "brasileir"):
		return "Brasileirão"
	case strings.Contains(tournament, "do brasil"):
		return "Copa do Brasil"
	case strings.Contains(tournament, "mineiro"):
		return "Campeonato Mineiro"
	case strings.Contains(tournament, "americana"):
		return "Sul-Americana"
	}
	return input
}

func normalizeTeam(input string) string {
	return strings.TrimSpace(input)
}
