package calendar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"Eve/internal/logger"

	ics "github.com/arran4/golang-ical"
)

const (
	fetchTimeout         = 15 * time.Second
	maxICSBytes          = 8 << 20
	maxURLLength         = 500
	defaultTimedDuration = time.Hour
	defaultSummary       = "(sans titre)"

	userAgent = "Eve-Discord-Bot/2 (+ICS calendar)"
)

var httpClient = &http.Client{Timeout: fetchTimeout}

type Event struct {
	UID         string
	Summary     string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
	AllDay      bool
	Recurring   bool
}

func normalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("empty url")
	}
	if len(trimmed) > maxURLLength {
		return "", fmt.Errorf("url longer than %d characters", maxURLLength)
	}
	if strings.HasPrefix(strings.ToLower(trimmed), "webcal://") {
		trimmed = "https://" + trimmed[len("webcal://"):]
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return "", fmt.Errorf("unsupported url scheme %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("url has no host")
	}
	return parsed.String(), nil
}

func fetchEvents(ctx context.Context, rawURL string) ([]Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/calendar, text/plain;q=0.9, */*;q=0.5")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching ICS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("ICS endpoint returned HTTP %d", resp.StatusCode)
	}

	cal, err := ics.ParseCalendar(io.LimitReader(resp.Body, maxICSBytes))
	if err != nil {
		return nil, fmt.Errorf("parsing ICS: %w", err)
	}
	if len(cal.Components) == 0 && len(cal.CalendarProperties) == 0 {
		return nil, errors.New("response is not an iCalendar document")
	}
	return eventsFromCalendar(cal), nil
}

func eventsFromCalendar(cal *ics.Calendar) []Event {
	raw := cal.Events()
	zones := declaredZones(cal)
	out := make([]Event, 0, len(raw))
	for _, vEvent := range raw {
		event, ok := convertEvent(vEvent, zones)
		if !ok {
			continue
		}
		out = append(out, event)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out
}

func convertEvent(vEvent *ics.VEvent, zones map[string]*time.Location) (Event, bool) {
	uid := propertyText(vEvent, ics.ComponentPropertyUniqueId)

	startProp := vEvent.GetProperty(ics.ComponentPropertyDtStart)
	if startProp == nil {
		logger.Warn("Calendar: VEVENT without DTSTART ignored", "uid", uid)
		return Event{}, false
	}
	start, err := vEvent.GetStartAt()
	if err != nil {
		fallback, ok := timeFromProperty(startProp, zones)
		if !ok {
			logger.Warn("Calendar: VEVENT with unreadable DTSTART ignored",
				"uid", uid, "dtstart", strings.TrimSpace(startProp.Value), "error", err)
			return Event{}, false
		}
		logger.Debug("Calendar: DTSTART resolved from the feed's own VTIMEZONE",
			"uid", uid, "start", fallback.Format(time.RFC3339))
		start = fallback
	}

	allDay := isDateOnly(startProp)
	event := Event{
		UID:         uid,
		Summary:     propertyText(vEvent, ics.ComponentPropertySummary),
		Description: propertyText(vEvent, ics.ComponentPropertyDescription),
		Location:    propertyText(vEvent, ics.ComponentPropertyLocation),
		Start:       start,
		AllDay:      allDay,
		Recurring:   vEvent.HasProperty(ics.ComponentPropertyRrule),
	}
	if event.Summary == "" {
		event.Summary = defaultSummary
	}
	event.End = resolveEnd(vEvent, zones, start, allDay)
	return event, true
}

func resolveEnd(vEvent *ics.VEvent, zones map[string]*time.Location, start time.Time, allDay bool) time.Time {
	if endProp := vEvent.GetProperty(ics.ComponentPropertyDtEnd); endProp != nil {
		end, err := vEvent.GetEndAt()
		if err != nil {
			end, _ = timeFromProperty(endProp, zones)
		}
		if end.After(start) {
			return end
		}
	}
	if d, ok := parseICSDuration(propertyText(vEvent, ics.ComponentPropertyDuration)); ok && d > 0 {
		return start.Add(d)
	}
	if allDay {
		return start.AddDate(0, 0, 1)
	}
	return start.Add(defaultTimedDuration)
}

const tzOffsetTo = ics.ComponentProperty(ics.PropertyTzoffsetto)

func declaredZones(cal *ics.Calendar) map[string]*time.Location {
	var zones map[string]*time.Location
	for _, timezone := range cal.Timezones() {
		idProp := timezone.GetProperty(ics.ComponentPropertyTzid)
		if idProp == nil {
			continue
		}
		id := strings.TrimSpace(idProp.Value)
		if id == "" {
			continue
		}
		if _, err := time.LoadLocation(id); err == nil {
			continue
		}
		offset, ok := declaredOffset(timezone)
		if !ok {
			continue
		}
		if zones == nil {
			zones = make(map[string]*time.Location, 1)
		}
		zones[id] = time.FixedZone(id, offset)
	}
	return zones
}

func declaredOffset(timezone *ics.VTimezone) (int, bool) {
	var daylight int
	var hasDaylight bool
	for _, sub := range timezone.SubComponents() {
		switch component := sub.(type) {
		case *ics.Standard:
			if offset, ok := parseUTCOffset(component.GetProperty(tzOffsetTo)); ok {
				return offset, true
			}
		case *ics.Daylight:
			if offset, ok := parseUTCOffset(component.GetProperty(tzOffsetTo)); ok && !hasDaylight {
				daylight, hasDaylight = offset, true
			}
		}
	}
	return daylight, hasDaylight
}

func parseUTCOffset(prop *ics.IANAProperty) (int, bool) {
	if prop == nil {
		return 0, false
	}
	value := strings.TrimSpace(prop.Value)
	if len(value) != 5 && len(value) != 7 {
		return 0, false
	}
	sign := 1
	switch value[0] {
	case '+':
	case '-':
		sign = -1
	default:
		return 0, false
	}

	digits := value[1:]
	if _, err := strconv.Atoi(digits); err != nil {
		return 0, false
	}
	hours, _ := strconv.Atoi(digits[0:2])
	minutes, _ := strconv.Atoi(digits[2:4])
	seconds := 0
	if len(digits) == 6 {
		seconds, _ = strconv.Atoi(digits[4:6])
	}
	return sign * (hours*3600 + minutes*60 + seconds), true
}

func timeFromProperty(prop *ics.IANAProperty, zones map[string]*time.Location) (time.Time, bool) {
	if prop == nil || len(zones) == 0 {
		return time.Time{}, false
	}
	location := declaredZoneOf(prop, zones)
	if location == nil {
		return time.Time{}, false
	}

	value := strings.TrimSpace(prop.Value)
	for _, layout := range []string{"20060102T150405", "20060102"} {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func declaredZoneOf(prop *ics.IANAProperty, zones map[string]*time.Location) *time.Location {
	for key, values := range prop.ICalParameters {
		if !strings.EqualFold(key, "TZID") || len(values) == 0 {
			continue
		}
		if location, ok := zones[strings.Trim(strings.TrimSpace(values[0]), `"`)]; ok {
			return location
		}
	}
	return nil
}

func propertyText(vEvent *ics.VEvent, property ics.ComponentProperty) string {
	prop := vEvent.GetProperty(property)
	if prop == nil {
		return ""
	}
	return strings.TrimSpace(ics.FromText(prop.Value))
}

func isDateOnly(prop *ics.IANAProperty) bool {
	for key, values := range prop.ICalParameters {
		if !strings.EqualFold(key, "VALUE") {
			continue
		}
		for _, value := range values {
			if strings.EqualFold(strings.Trim(value, `"`), "DATE") {
				return true
			}
		}
	}
	value := strings.TrimSpace(prop.Value)
	if len(value) != 8 {
		return false
	}
	_, err := strconv.Atoi(value)
	return err == nil
}

var durationPattern = regexp.MustCompile(`^([+-])?P(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

func parseICSDuration(raw string) (time.Duration, bool) {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if value == "" {
		return 0, false
	}
	match := durationPattern.FindStringSubmatch(value)
	if match == nil {
		return 0, false
	}

	units := []time.Duration{7 * 24 * time.Hour, 24 * time.Hour, time.Hour, time.Minute, time.Second}
	var total time.Duration
	for i, unit := range units {
		part := match[i+2]
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return 0, false
		}
		total += time.Duration(n) * unit
	}
	if total == 0 {
		return 0, false
	}
	if match[1] == "-" {
		total = -total
	}
	return total, true
}
