package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmojiURLToUnicode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"single_codepoint",
			"https://github.githubassets.com/images/icons/emoji/unicode/1f600.png?v8",
			"\U0001f600",
		},
		{
			"flag_multi_codepoint",
			"https://github.githubassets.com/images/icons/emoji/unicode/1f1ef-1f1f5.png?v8",
			"\U0001f1ef\U0001f1f5",
		},
		{
			"zwj_sequence",
			"https://github.githubassets.com/images/icons/emoji/unicode/1f468-200d-1f4bb.png?v8",
			"\U0001f468\u200d\U0001f4bb",
		},
		{
			"no_slash",
			"nopath",
			"nopath",
		},
		{
			"no_dot",
			"https://example.com/nodot",
			"https://example.com/nodot",
		},
		{
			"invalid_hex",
			"https://example.com/xyz.png",
			"https://example.com/xyz.png",
		},
		{
			"ascii_codepoint",
			"https://example.com/41.png",
			"A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := emojiURLToUnicode(tt.input); got != tt.want {
				t.Errorf("emojiURLToUnicode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestReplaceEmoji(t *testing.T) {
	emoji := map[string]string{
		"smile": "\U0001f604",
		"heart": "\u2764\ufe0f",
		"wave":  "\U0001f44b",
	}

	tests := []struct {
		name  string
		input string
		emoji map[string]string
		want  string
	}{
		{"no_emoji", "hello world", emoji, "hello world"},
		{"single_known", ":smile: hello", emoji, "\U0001f604 hello"},
		{"multiple_known", ":smile: and :heart:", emoji, "\U0001f604 and \u2764\ufe0f"},
		{"unknown_emoji", ":unknown: text", emoji, ":unknown: text"},
		{"known_and_unknown", ":smile: :unknown: :wave:", emoji, "\U0001f604 :unknown: \U0001f44b"},
		{"empty_string", "", emoji, ""},
		{"nil_map", ":smile: text", nil, ":smile: text"},
		{"time_pattern", "12:30:00", emoji, "12:30:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceEmoji(tt.input, tt.emoji); got != tt.want {
				t.Errorf("replaceEmoji(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEmojiCacheDir(t *testing.T) {
	got := emojiCacheDir()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".cache")
	}
	want := filepath.Join(cacheDir, "gh", "gh-list-pr")
	if got != want {
		t.Errorf("emojiCacheDir() = %q, want %q", got, want)
	}
}

func TestEmojiCachePath(t *testing.T) {
	got := emojiCachePath()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".cache")
	}
	want := filepath.Join(cacheDir, "gh", "gh-list-pr", "emoji.json")
	if got != want {
		t.Errorf("emojiCachePath() = %q, want %q", got, want)
	}
}
