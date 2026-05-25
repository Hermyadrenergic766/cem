package main

import "testing"

func TestSuggestTool(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"claude", "claude"},     // exact
		{"cluade", "claude"},     // transposition
		{"claud", "claude"},      // missing char
		{"agi", "agy"},           // 1 sub
		{"cursr", "cursor"},      // missing 'o'
		{"gpt5", "gpt"},          // extra char (dist 1)
		{"xyzqqq", ""},           // too far
		{"", ""},                 // empty
	}
	for _, c := range cases {
		got := suggestTool(c.input)
		if got != c.want {
			t.Errorf("suggestTool(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
