package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Colors for the date column, from the most recent to the oldest.
const (
	dateToday     = "\033[1;93m" // bold bright yellow
	dateYesterday = "\033[93m"   // bright yellow
	dateThisWeek  = "\033[33m"   // yellow
	dateThisMonth = ""           // default foreground
	dateThisYear  = brightBlack
	dateOlder     = "\033[38;5;240m" // dark gray
)

// annotateDates fills DateText and DateColor of each PR relative to now.
func annotateDates(prs []PullRequest, now time.Time) {
	ja := isJapaneseLocale()
	for i := range prs {
		t, err := time.Parse(time.RFC3339, prs[i].CreatedAt)
		if err != nil {
			prs[i].DateText = prs[i].CreatedAt
			continue
		}
		t = t.In(now.Location())
		days := daysBetween(t, now)
		prs[i].DateText = finderDate(t, days, ja)
		prs[i].DateColor = dateColor(days)
	}
}

// daysBetween returns the number of calendar days from t to now.
func daysBetween(t, now time.Time) int {
	y1, m1, d1 := t.Date()
	y2, m2, d2 := now.Date()
	from := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	to := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)
	return int(to.Sub(from).Hours() / 24)
}

// finderDate formats t like the list view of Finder does.
func finderDate(t time.Time, days int, ja bool) string {
	if ja {
		// Finder does not zero-pad the hour, which Go's "15" layout always does.
		hm := fmt.Sprintf("%d:%02d", t.Hour(), t.Minute())
		switch days {
		case 0:
			return "今日 " + hm
		case 1:
			return "昨日 " + hm
		}
		return t.Format("2006年1月2日 ") + hm
	}
	switch days {
	case 0:
		return t.Format("Today at 3:04 PM")
	case 1:
		return t.Format("Yesterday at 3:04 PM")
	}
	return t.Format("Jan 2, 2006 at 3:04 PM")
}

func dateColor(days int) string {
	switch {
	case days <= 0:
		return dateToday
	case days == 1:
		return dateYesterday
	case days < 7:
		return dateThisWeek
	case days < 30:
		return dateThisMonth
	case days < 365:
		return dateThisYear
	}
	return dateOlder
}

func isJapaneseLocale() bool {
	for _, key := range []string{"LC_ALL", "LC_TIME", "LANG"} {
		if v := os.Getenv(key); v != "" {
			return strings.HasPrefix(v, "ja")
		}
	}
	return false
}
