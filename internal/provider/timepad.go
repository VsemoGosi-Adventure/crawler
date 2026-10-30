package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/VsemoGosi-Adventure/crawler/internal/dto"
)

var (
	htmlTagRegex   = regexp.MustCompile(`<[^>]*>`)
	lineBreakRegex = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</div>`)
	spaceRegex     = regexp.MustCompile(`\s+`)
)

type TimepadClient struct {
	httpClient *http.Client
	baseURL    string
	city       string
	token      string
	limit      int
}

func NewTimepadClient(city string, token ...string) *TimepadClient {
	var tok string
	if len(token) > 0 {
		tok = token[0]
	} else {
		tok = os.Getenv("TIMEPAD_TOKEN")
		if tok == "" {
			tok = os.Getenv("TIMEPAD_API_TOKEN")
		}
	}

	return &TimepadClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: "https://api.timepad.ru/v1/events.json",
		city:    city,
		token:   tok,
		limit:   50,
	}
}

func (c *TimepadClient) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

func (c *TimepadClient) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

type timepadResponse struct {
	Total  int            `json:"total"`
	Values []timepadEvent `json:"values"`
}

type timepadEvent struct {
	ID               int                      `json:"id"`
	Name             string                   `json:"name"`
	URL              string                   `json:"url"`
	Description      string                   `json:"description"`
	DescriptionShort string                   `json:"description_short"`
	DescriptionHTML  string                   `json:"description_html"`
	StartsAt         string                   `json:"starts_at"`
	EndsAt           *string                  `json:"ends_at"`
	TicketsLimit     *int                     `json:"tickets_limit"`
	AccessStatus     string                   `json:"access_status"`
	Location         *timepadLocation         `json:"location"`
	Categories       []timepadCategory        `json:"categories"`
	RegistrationData *timepadRegistrationData `json:"registration_data"`
}

type timepadLocation struct {
	Name        *string   `json:"name"`
	Venue       *string   `json:"venue"`
	City        *string   `json:"city"`
	Address     *string   `json:"address"`
	Coordinates []float64 `json:"coordinates"`
	Latitude    *float64  `json:"latitude"`
	Longitude   *float64  `json:"longitude"`
}

type timepadCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

type timepadRegistrationData struct {
	TicketsLimit       *int     `json:"tickets_limit"`
	IsRegistrationOpen *bool    `json:"is_registration_open"`
	PriceMin           *float64 `json:"price_min"`
	PriceMax           *float64 `json:"price_max"`
}

func (c *TimepadClient) Name() string {
	return "timepad"
}

func (c *TimepadClient) FetchEvents(ctx context.Context) ([]dto.CrawledEventMessage, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timepad url: %w", err)
	}

	limit := c.limit
	if limit <= 0 {
		limit = 50
	}

	q := u.Query()
	q.Set("limit", strconv.Itoa(limit))
	q.Set("fields", "location,description_short,description_html,categories,registration_data")

	if c.city != "" {
		q.Add("cities[]", c.city)
	}

	q.Set("starts_at_min", time.Now().Format(time.RFC3339))
	q.Set("access_statuses[]", "public")
	q.Set("sort", "+starts_at")

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create timepad request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute timepad request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("timepad api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var data timepadResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to decode timepad response: %w", err)
	}

	events := make([]dto.CrawledEventMessage, 0, len(data.Values))
	for _, ev := range data.Values {
		item, err := c.toCrawledEvent(ev)
		if err != nil {
			continue
		}
		events = append(events, item)
	}

	return events, nil
}

func (c *TimepadClient) toCrawledEvent(ev timepadEvent) (dto.CrawledEventMessage, error) {
	startsAt, err := parseTime(ev.StartsAt)
	if err != nil {
		return dto.CrawledEventMessage{}, fmt.Errorf("invalid starts_at %q: %w", ev.StartsAt, err)
	}

	var endsAtPtr *time.Time
	if ev.EndsAt != nil && *ev.EndsAt != "" {
		if t, err := parseTime(*ev.EndsAt); err == nil {
			endsAtPtr = &t
		}
	}

	eventURL := ev.URL
	if eventURL == "" {
		eventURL = fmt.Sprintf("https://timepad.ru/event/%d", ev.ID)
	}

	desc := strings.TrimSpace(ev.Description)
	if desc == "" {
		desc = strings.TrimSpace(ev.DescriptionShort)
	}
	if desc == "" && ev.DescriptionHTML != "" {
		desc = stripHTML(ev.DescriptionHTML)
	}

	var maxParticipants *int
	if ev.RegistrationData != nil && ev.RegistrationData.TicketsLimit != nil && *ev.RegistrationData.TicketsLimit > 0 {
		maxParticipants = ev.RegistrationData.TicketsLimit
	} else if ev.TicketsLimit != nil && *ev.TicketsLimit > 0 {
		maxParticipants = ev.TicketsLimit
	}

	loc := dto.LocationDTO{}
	if ev.Location != nil {
		if ev.Location.Name != nil && *ev.Location.Name != "" {
			loc.Name = *ev.Location.Name
		} else if ev.Location.Venue != nil && *ev.Location.Venue != "" {
			loc.Name = *ev.Location.Venue
		} else if ev.Location.City != nil && *ev.Location.City != "" {
			loc.Name = *ev.Location.City
		}

		city := ""
		if ev.Location.City != nil {
			city = strings.TrimSpace(*ev.Location.City)
		}
		addr := ""
		if ev.Location.Address != nil {
			addr = strings.TrimSpace(*ev.Location.Address)
		}

		if city != "" && addr != "" {
			if strings.Contains(strings.ToLower(addr), strings.ToLower(city)) {
				loc.Address = addr
			} else {
				loc.Address = fmt.Sprintf("%s, %s", city, addr)
			}
		} else if addr != "" {
			loc.Address = addr
		} else {
			loc.Address = city
		}

		if ev.Location.Latitude != nil && ev.Location.Longitude != nil {
			loc.Latitude = ev.Location.Latitude
			loc.Longitude = ev.Location.Longitude
		} else if len(ev.Location.Coordinates) >= 2 {
			lat := ev.Location.Coordinates[0]
			lon := ev.Location.Coordinates[1]
			loc.Latitude = &lat
			loc.Longitude = &lon
		}
	}

	tags := make([]string, 0, len(ev.Categories))
	seen := make(map[string]struct{})
	for _, cat := range ev.Categories {
		t := strings.TrimSpace(cat.Tag)
		if t == "" {
			t = strings.TrimSpace(cat.Name)
		}
		if t != "" {
			lower := strings.ToLower(t)
			if _, ok := seen[lower]; !ok {
				seen[lower] = struct{}{}
				tags = append(tags, t)
			}
		}
	}

	return dto.CrawledEventMessage{
		ExternalSource:   c.Name(),
		ExternalSourceID: strconv.Itoa(ev.ID),
		ExternalURL:      eventURL,
		Title:            ev.Name,
		Description:      desc,
		StartsAt:         startsAt,
		EndsAt:           endsAtPtr,
		MaxParticipants:  maxParticipants,
		JoinPolicy:       "open",
		Location:         loc,
		Tags:             tags,
	}, nil
}

func parseTime(value string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z0700",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date string %q", value)
}

func stripHTML(input string) string {
	text := lineBreakRegex.ReplaceAllString(input, " ")
	text = htmlTagRegex.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = spaceRegex.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}
