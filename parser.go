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

func ParseEvents(r io.Reader, loc *time.Location) ([]Event, error) {
	doc, err := htmlquery.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("ParseEvents(): could not parse input: %w", err)
	}

	if !isCalendarPage(doc) {
		return nil, fmt.Errorf("ParseEvents(): invalid html page: %w", ErrUnexpectedInput)
	}

	nodes, err := htmlquery.QueryAll(doc, `//section[@class="agenda-partidas"]//div[contains(concat(" ",normalize-space(@class)," ")," partida ") and not(contains(@class, "partida-finalizada"))]`)
	if err != nil {
		return nil, fmt.Errorf("ParseEvents(): could not query main section node: %w", err)
	}

	events := make([]Event, len(nodes))
	for i, node := range nodes {
		var event Event

		div, err := htmlquery.Query(node, `//div[@class="partida-data"]`)
		switch {
		case err != nil:
			return nil, fmt.Errorf("ParseEvents(): could not query 'div.partida-data': %w", err)
		case div == nil:
			return nil, fmt.Errorf("ParseEvents(): missing expected 'div.partida-data' node: %w", ErrUnexpectedInput)
		}
		event.Stadium = parseStadium(htmlquery.InnerText(div.LastChild))
		event.DateTime = parseDateTime(htmlquery.InnerText(div.FirstChild.NextSibling), loc)

		div, err = htmlquery.Query(node, `//div[@class="partida-campeonato"]`)
		switch {
		case err != nil:
			return nil, fmt.Errorf("ParseEvents(): could not query 'div.partida-campeonato': %w", err)
		case div == nil:
			return nil, fmt.Errorf("ParseEvents(): missing expected 'div.partida-campeonato' node: %w", ErrUnexpectedInput)
		}
		event.Tournament = parseTournament(htmlquery.InnerText(div))

		abbr, err := htmlquery.QueryAll(node, `//div[@class="partida-placar"]//abbr[@title]`)
		switch {
		case err != nil:
			return nil, fmt.Errorf("ParseEvents(): could not query 'abbr[@title]': %w", err)
		case abbr == nil:
			return nil, fmt.Errorf("ParseEvents(): missing expected 'abbr[@title]' nodes: %w", ErrUnexpectedInput)
		case len(abbr) != 2:
			return nil, fmt.Errorf("ParseEvents(): missing expected 2 'abbr[@title]' nodes: %w", ErrUnexpectedInput)
		}
		event.HomeTeam = parseTeam(htmlquery.SelectAttr(abbr[0], "title"))
		event.AwayTeam = parseTeam(htmlquery.SelectAttr(abbr[1], "title"))

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
	t, _ := time.ParseInLocation("02/01 às 15:04-2006", fmt.Sprintf("%s-%d", input, time.Now().Year()), loc)
	return t
}

func parseStadium(input string) string {
	return strings.TrimSpace(input)
}

func parseTournament(input string) string {
	tournament := strings.TrimSpace(input)
	switch {
	case strings.Contains(tournament, "Libertadores"):
		return "Libertadores"
	case strings.Contains(tournament, "Brasileiro"):
		return "Brasileirão"
	}
	return tournament
}

func parseTeam(input string) string {
	return strings.TrimSpace(input)
}
