package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/text/language"
)

type userDateFormat struct {
	DisplayLayout string
	Hint          string
	DateLayouts   []string
}

func detectUserDateFormat() userDateFormat {
	tag := detectLocaleTag()
	if tag == language.Und {
		return dateFormatDMY()
	}

	region, _ := tag.Region()
	switch region.String() {
	case "US":
		return dateFormatMDY()
	case "CA", "CN", "JP", "KR", "HU", "LT":
		return dateFormatYMD()
	default:
		return dateFormatDMY()
	}
}

func detectLocaleTag() language.Tag {
	for _, key := range []string{"LC_ALL", "LC_TIME", "LANG"} {
		raw := strings.TrimSpace(os.Getenv(key))
		if raw == "" {
			continue
		}
		raw = normalizeLocale(raw)
		if raw == "" {
			continue
		}
		if tag, err := language.Parse(raw); err == nil {
			return tag
		}
	}
	return language.Und
}

func normalizeLocale(raw string) string {
	locale := raw
	if idx := strings.Index(locale, "."); idx >= 0 {
		locale = locale[:idx]
	}
	if idx := strings.Index(locale, "@"); idx >= 0 {
		locale = locale[:idx]
	}
	locale = strings.ReplaceAll(locale, "_", "-")
	return strings.TrimSpace(locale)
}

func dateFormatMDY() userDateFormat {
	return userDateFormat{
		DisplayLayout: "01/02/2006",
		Hint:          "MM/DD/YYYY",
		DateLayouts: []string{
			"1/2/2006",
			"01/02/2006",
			"1-2-2006",
			"01-02-2006",
			"1.2.2006",
			"01.02.2006",
		},
	}
}

func dateFormatDMY() userDateFormat {
	return userDateFormat{
		DisplayLayout: "02/01/2006",
		Hint:          "DD/MM/YYYY",
		DateLayouts: []string{
			"2/1/2006",
			"02/01/2006",
			"2-1-2006",
			"02-01-2006",
			"2.1.2006",
			"02.01.2006",
		},
	}
}

func dateFormatYMD() userDateFormat {
	return userDateFormat{
		DisplayLayout: "2006-01-02",
		Hint:          "YYYY-MM-DD",
		DateLayouts: []string{
			"2006-1-2",
			"2006-01-02",
			"2006/1/2",
			"2006/01/02",
			"2006.1.2",
			"2006.01.02",
		},
	}
}

func (m Model) formatDueDate(dueAt time.Time) string {
	if m.dateFormat.DisplayLayout == "" {
		return dueAt.UTC().Format("2006-01-02")
	}
	return dueAt.UTC().Format(m.dateFormat.DisplayLayout)
}

func (m Model) formatCommentDateTime(ts time.Time) string {
	dateLayout := m.dateFormat.DisplayLayout
	if dateLayout == "" {
		dateLayout = "2006-01-02"
	}
	// Comment timestamps are moments in time, so render in local time with HH:mm.
	return ts.In(time.Local).Format(dateLayout + " 15:04")
}

func (m Model) dueDatePlaceholder() string {
	hint := m.dateFormat.Hint
	if hint == "" {
		hint = "YYYY-MM-DD"
	}
	return fmt.Sprintf("Due Date (%s)", hint)
}

func (m Model) parseDueDateInput(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	layouts := make([]string, 0, len(m.dateFormat.DateLayouts)+6)
	layouts = append(layouts, m.dateFormat.DateLayouts...)
	layouts = append(layouts,
		"2006-01-02",
		"2006/01/02",
		"2 Jan 2006",
		"02 Jan 2006",
		"2 January 2006",
		"January 2 2006",
	)

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			dateOnlyUTC := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			return &dateOnlyUTC, nil
		}
	}

	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		v := t.UTC()
		return &v, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		v := t.UTC()
		return &v, nil
	}

	return nil, fmt.Errorf("due date must match %s (locale), YYYY-MM-DD, or RFC3339", m.dateFormat.Hint)
}
