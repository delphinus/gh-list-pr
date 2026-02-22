package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

var emojiPattern = regexp.MustCompile(`:(\w+):`)

func emojiCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheDir, "gh", "gh-list-pr")
}

func emojiCachePath() string {
	return filepath.Join(emojiCacheDir(), "emoji.json")
}

func loadEmoji() (map[string]string, error) {
	cachePath := emojiCachePath()

	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < 7*24*time.Hour {
			data, err := os.ReadFile(cachePath)
			if err == nil {
				var emoji map[string]string
				if err := json.Unmarshal(data, &emoji); err == nil {
					return emoji, nil
				}
			}
		}
	}

	return fetchAndCacheEmoji(cachePath)
}

func fetchAndCacheEmoji(cachePath string) (map[string]string, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("create REST client: %w", err)
	}

	var raw map[string]string
	if err := client.Get("emojis", &raw); err != nil {
		return nil, fmt.Errorf("fetch emojis: %w", err)
	}

	emoji := make(map[string]string, len(raw))
	for name, url := range raw {
		emoji[name] = emojiURLToUnicode(url)
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return emoji, nil // return emoji even if cache write fails
	}
	data, _ := json.Marshal(emoji)
	_ = os.WriteFile(cachePath, data, 0o644)

	return emoji, nil
}

func emojiURLToUnicode(url string) string {
	// URL format: https://github.githubassets.com/images/icons/emoji/unicode/1f600.png?v8
	// or multi-codepoint: .../1f1ef-1f1f5.png?v8
	idx := strings.LastIndex(url, "/")
	if idx < 0 {
		return url
	}
	filename := url[idx+1:]
	dotIdx := strings.Index(filename, ".")
	if dotIdx < 0 {
		return url
	}
	hex := filename[:dotIdx]

	parts := strings.Split(hex, "-")
	var runes []rune
	for _, p := range parts {
		cp, err := strconv.ParseInt(p, 16, 32)
		if err != nil {
			return url
		}
		runes = append(runes, rune(cp))
	}
	return string(runes)
}

func replaceEmoji(s string, emoji map[string]string) string {
	return emojiPattern.ReplaceAllStringFunc(s, func(match string) string {
		name := match[1 : len(match)-1] // strip ':'
		if u, ok := emoji[name]; ok {
			return u
		}
		return match
	})
}
