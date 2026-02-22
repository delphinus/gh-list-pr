package main

import (
	"testing"
)

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"ascii", "hello", 5},
		{"fullwidth", "日本語", 6},
		{"mixed", "Hi日本", 6},
		{"emoji", "👍", 2},
		{"space", " ", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayWidth(tt.input); got != tt.want {
				t.Errorf("displayWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncatePad(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{"width_zero", "hello", 0, ""},
		{"width_negative", "hello", -1, ""},
		{"short_pad", "hi", 5, "hi   "},
		{"exact_fit", "hello", 5, "hello"},
		{"truncate", "hello world", 5, "hell\u2026"},
		{"width_one", "hello", 1, "\u2026"},
		{"empty_pad", "", 5, "     "},
		{"fullwidth_truncate", "日本語テスト", 5, "日本\u2026"},
		{"fullwidth_pad", "日本", 6, "日本  "},
		{"fullwidth_exact", "日本語", 6, "日本語"},
		{"fullwidth_boundary", "日本語", 5, "日本\u2026"},
		{"single_char_exact", "a", 1, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePad(tt.input, tt.width)
			if got != tt.want {
				t.Errorf("truncatePad(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
			}
			if tt.width > 0 {
				gotW := displayWidth(got)
				if gotW != tt.width {
					t.Errorf("truncatePad(%q, %d) result display width = %d, want %d", tt.input, tt.width, gotW, tt.width)
				}
			}
		})
	}
}

func TestFzfMargin(t *testing.T) {
	tests := []struct {
		name       string
		opt        options
		envDefault string
		want       int
	}{
		{"print_mode", options{print: true}, "", 0},
		{"no_options", options{}, "", 2},
		{"border_rounded", options{fzfOptions: "--border=rounded"}, "", 4},
		{"border_sharp", options{fzfOptions: "--border=sharp"}, "", 4},
		{"border_bold", options{fzfOptions: "--border=bold"}, "", 4},
		{"border_double", options{fzfOptions: "--border=double"}, "", 4},
		{"border_vertical", options{fzfOptions: "--border=vertical"}, "", 4},
		{"border_no_value", options{fzfOptions: "--border"}, "", 4},
		{"border_left", options{fzfOptions: "--border=left"}, "", 3},
		{"border_right", options{fzfOptions: "--border=right"}, "", 3},
		{"border_none", options{fzfOptions: "--border=none"}, "", 2},
		{"border_top", options{fzfOptions: "--border=top"}, "", 2},
		{"padding_single", options{fzfOptions: "--padding 2"}, "", 6},
		{"padding_two_values", options{fzfOptions: "--padding 1,3"}, "", 8},
		{"padding_four_values", options{fzfOptions: "--padding 1,2,3,4"}, "", 8},
		{"border_and_padding", options{fzfOptions: "--border=rounded --padding 1"}, "", 6},
		{"env_override", options{fzfOptions: "--border=none"}, "--border=rounded", 2},
		{"env_only", options{}, "--border=rounded", 4},
		{"invalid_padding", options{fzfOptions: "--padding abc"}, "", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FZF_DEFAULT_OPTS", tt.envDefault)
			if got := fzfMargin(tt.opt); got != tt.want {
				t.Errorf("fzfMargin() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTermWidth(t *testing.T) {
	t.Run("valid_columns", func(t *testing.T) {
		t.Setenv("COLUMNS", "120")
		if got := termWidth(); got != 120 {
			t.Errorf("termWidth() = %d, want 120", got)
		}
	})

	t.Run("invalid_columns", func(t *testing.T) {
		t.Setenv("COLUMNS", "abc")
		got := termWidth()
		if got <= 0 {
			t.Errorf("termWidth() = %d, want > 0", got)
		}
	})

	t.Run("empty_columns", func(t *testing.T) {
		t.Setenv("COLUMNS", "")
		got := termWidth()
		if got <= 0 {
			t.Errorf("termWidth() = %d, want > 0", got)
		}
	})

	t.Run("zero_columns", func(t *testing.T) {
		t.Setenv("COLUMNS", "0")
		got := termWidth()
		if got <= 0 {
			t.Errorf("termWidth() = %d, want > 0", got)
		}
	})

	t.Run("negative_columns", func(t *testing.T) {
		t.Setenv("COLUMNS", "-1")
		got := termWidth()
		if got <= 0 {
			t.Errorf("termWidth() = %d, want > 0", got)
		}
	})
}
