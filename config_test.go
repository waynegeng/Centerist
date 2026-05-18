package main

import "testing"

func TestFormatParseHotkeyRoundTrip(t *testing.T) {
	cases := []string{
		"Alt+C",
		"Alt+Enter",
		"Ctrl+Shift+F5",
		"Win+1",
	}
	for _, s := range cases {
		mod, vk, err := parseHotkey(s)
		if err != nil {
			t.Fatalf("parseHotkey(%q): %v", s, err)
		}
		got, ok := formatHotkey(mod&^MOD_NOREPEAT, vk)
		if !ok {
			t.Fatalf("formatHotkey failed for %q", s)
		}
		mod2, vk2, err := parseHotkey(got)
		if err != nil {
			t.Fatalf("parseHotkey(%q) after format: %v", got, err)
		}
		if mod2&^MOD_NOREPEAT != mod&^MOD_NOREPEAT || vk2 != vk {
			t.Fatalf("round trip %q -> %q: mod/vk mismatch", s, got)
		}
	}
}
