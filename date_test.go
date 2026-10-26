package main

import (
	"testing"
	"time"
)

func TestAnnotateDates(t *testing.T) {
	now := time.Date(2026, 10, 26, 13, 30, 0, 0, time.UTC)

	tests := []struct {
		createdAt string
		ja        string
		en        string
		color     string
	}{
		{"2026-10-26T00:05:00Z", "今日 0:05", "Today at 12:05 AM", dateToday},
		{"2026-10-25T23:59:00Z", "昨日 23:59", "Yesterday at 11:59 PM", dateYesterday},
		{"2026-10-20T09:00:00Z", "2026年10月20日 9:00", "Oct 20, 2026 at 9:00 AM", dateThisWeek},
		{"2026-10-03T14:00:00Z", "2026年10月3日 14:00", "Oct 3, 2026 at 2:00 PM", dateThisMonth},
		{"2026-01-15T10:00:00Z", "2026年1月15日 10:00", "Jan 15, 2026 at 10:00 AM", dateThisYear},
		{"2024-12-31T10:00:00Z", "2024年12月31日 10:00", "Dec 31, 2024 at 10:00 AM", dateOlder},
	}

	for _, lang := range []string{"ja_JP.UTF-8", "en_US.UTF-8"} {
		t.Run(lang, func(t *testing.T) {
			t.Setenv("LC_ALL", "")
			t.Setenv("LC_TIME", "")
			t.Setenv("LANG", lang)
			for _, tt := range tests {
				prs := []PullRequest{{CreatedAt: tt.createdAt}}
				annotateDates(prs, now)
				want := tt.en
				if lang == "ja_JP.UTF-8" {
					want = tt.ja
				}
				if prs[0].DateText != want {
					t.Errorf("DateText(%s) = %q, want %q", tt.createdAt, prs[0].DateText, want)
				}
				if prs[0].DateColor != tt.color {
					t.Errorf("DateColor(%s) = %q, want %q", tt.createdAt, prs[0].DateColor, tt.color)
				}
			}
		})
	}
}

func TestAnnotateDatesUsesLocalCalendarDay(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	now := time.Date(2026, 10, 26, 8, 0, 0, 0, jst)
	// 2026-10-25T16:00Z is 01:00 on Oct 26 in JST, so it is today, not yesterday.
	prs := []PullRequest{{CreatedAt: "2026-10-25T16:00:00Z"}}
	t.Setenv("LC_ALL", "ja_JP.UTF-8")
	annotateDates(prs, now)
	if prs[0].DateText != "今日 1:00" {
		t.Errorf("DateText = %q, want %q", prs[0].DateText, "今日 1:00")
	}
}

func TestIsJapaneseLocale(t *testing.T) {
	tests := []struct {
		lcAll, lcTime, lang string
		want                bool
	}{
		{"", "", "ja_JP.UTF-8", true},
		{"", "", "en_US.UTF-8", false},
		{"", "", "", false},
		{"C", "", "ja_JP.UTF-8", false},
		{"", "ja_JP.UTF-8", "en_US.UTF-8", true},
	}
	for _, tt := range tests {
		t.Setenv("LC_ALL", tt.lcAll)
		t.Setenv("LC_TIME", tt.lcTime)
		t.Setenv("LANG", tt.lang)
		if got := isJapaneseLocale(); got != tt.want {
			t.Errorf("isJapaneseLocale(LC_ALL=%q, LC_TIME=%q, LANG=%q) = %v, want %v",
				tt.lcAll, tt.lcTime, tt.lang, got, tt.want)
		}
	}
}
