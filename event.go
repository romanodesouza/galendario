package galendario

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
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
		event.Stadium = normalizeStadium(htmlquery.InnerText(div.LastChild))
		event.DateTime = parseDateTime(htmlquery.InnerText(div.FirstChild.NextSibling), loc)

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
	now := time.Now().In(loc)
	// Date and time
	t, err := time.ParseInLocation("02/01 às 15:04-2006", fmt.Sprintf("%s-%d", input, now.Year()), loc)
	// Date only
	if err != nil {
		t, _ = time.ParseInLocation("02/01-2006", fmt.Sprintf("%s-%d", input, now.Year()), loc)
	}
	return t
}

func normalizeStadium(input string) string {
	return strings.TrimSpace(input)
}

func normalizeTournament(input string) string {
	tournament := strings.ToLower(strings.TrimSpace(input))
	switch {
	case strings.Contains(tournament, "libertadores"):
		return "Copa Libertadores"
	case strings.Contains(tournament, "brasileir"):
		return "Campeonato Brasileiro"
	case strings.Contains(tournament, "do brasil"):
		return "Copa do Brasil"
	}
	return tournament
}

func normalizeTeam(input string) string {
	return strings.TrimSpace(input)
}
