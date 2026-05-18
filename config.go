package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type HotkeyConfig struct {
	Center          string `json:"center"`
	CenterFixedSize string `json:"centerFixedSize"`
}

type Config struct {
	Margin  int32        `json:"margin"`
	Hotkeys HotkeyConfig `json:"hotkeys"`
}

func defaultConfig() Config {
	return Config{
		Margin: 50,
		Hotkeys: HotkeyConfig{
			Center:          "Alt+C",
			CenterFixedSize: "Alt+Enter",
		},
	}
}

func configDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine config directory: %w", err)
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}
	return filepath.Join(appData, "Centerist"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func loadConfig() (Config, error) {
	p, err := configPath()
	if err != nil {
		return defaultConfig(), err
	}

	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		cfg := defaultConfig()
		if saveErr := saveConfig(cfg); saveErr != nil {
			fmt.Printf("warn: failed to save default config: %v\n", saveErr)
		}
		return cfg, nil
	}
	if err != nil {
		return defaultConfig(), fmt.Errorf("failed to read config: %w", err)
	}

	cfg := defaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig(), fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

func saveConfig(cfg Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

var modNames = map[string]int{
	"alt":   MOD_ALT,
	"ctrl":  MOD_CONTROL,
	"shift": MOD_SHIFT,
	"win":   MOD_WIN,
}

var vkNames = map[string]int{
	"a": 0x41, "b": 0x42, "c": 0x43, "d": 0x44, "e": 0x45,
	"f": 0x46, "g": 0x47, "h": 0x48, "i": 0x49, "j": 0x4A,
	"k": 0x4B, "l": 0x4C, "m": 0x4D, "n": 0x4E, "o": 0x4F,
	"p": 0x50, "q": 0x51, "r": 0x52, "s": 0x53, "t": 0x54,
	"u": 0x55, "v": 0x56, "w": 0x57, "x": 0x58, "y": 0x59,
	"z": 0x5A,
	"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33, "4": 0x34,
	"5": 0x35, "6": 0x36, "7": 0x37, "8": 0x38, "9": 0x39,
	"enter":     0x0D,
	"return":    0x0D,
	"space":     0x20,
	"tab":       0x09,
	"escape":    0x1B,
	"esc":       0x1B,
	"backspace": 0x08,
	"delete":    0x2E,
	"insert":    0x2D,
	"home":      0x24,
	"end":       0x23,
	"pageup":    0x21,
	"pagedown":  0x22,
	"left":      0x25,
	"up":        0x26,
	"right":     0x27,
	"down":      0x28,
	"f1":        0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73,
	"f5": 0x74, "f6": 0x75, "f7": 0x76, "f8": 0x77,
	"f9": 0x78, "f10": 0x79, "f11": 0x7A, "f12": 0x7B,
}

// vkCanonicalName maps virtual-key codes to display names accepted by parseHotkey.
var vkCanonicalName = map[int]string{
	0x08: "Backspace",
	0x09: "Tab",
	0x0D: "Enter",
	0x1B: "Esc",
	0x20: "Space",
	0x21: "PageUp",
	0x22: "PageDown",
	0x23: "End",
	0x24: "Home",
	0x25: "Left",
	0x26: "Up",
	0x27: "Right",
	0x28: "Down",
	0x2D: "Insert",
	0x2E: "Delete",
	0x70: "F1", 0x71: "F2", 0x72: "F3", 0x73: "F4",
	0x74: "F5", 0x75: "F6", 0x76: "F7", 0x77: "F8",
	0x78: "F9", 0x79: "F10", 0x7A: "F11", 0x7B: "F12",
}

func vkDisplayName(vk int) (string, bool) {
	if vk >= 0x41 && vk <= 0x5A {
		return string(rune(vk)), true
	}
	if vk >= 0x30 && vk <= 0x39 {
		return string(rune(vk)), true
	}
	if name, ok := vkCanonicalName[vk]; ok {
		return name, true
	}
	return "", false
}

func formatHotkey(mod, vk int) (string, bool) {
	if name, ok := vkDisplayName(vk); ok {
		var parts []string
		if mod&MOD_CONTROL != 0 {
			parts = append(parts, "Ctrl")
		}
		if mod&MOD_ALT != 0 {
			parts = append(parts, "Alt")
		}
		if mod&MOD_SHIFT != 0 {
			parts = append(parts, "Shift")
		}
		if mod&MOD_WIN != 0 {
			parts = append(parts, "Win")
		}
		parts = append(parts, name)
		return strings.Join(parts, "+"), true
	}
	return "", false
}

func parseHotkey(s string) (mod int, vk int, err error) {
	parts := strings.Split(s, "+")
	if len(parts) == 0 {
		return 0, 0, fmt.Errorf("empty hotkey string")
	}

	for _, part := range parts {
		key := strings.TrimSpace(strings.ToLower(part))
		if key == "" {
			continue
		}
		if m, ok := modNames[key]; ok {
			mod |= m
		} else if v, ok := vkNames[key]; ok {
			if vk != 0 {
				return 0, 0, fmt.Errorf("multiple non-modifier keys in hotkey %q", s)
			}
			vk = v
		} else {
			return 0, 0, fmt.Errorf("unknown key %q in hotkey %q", part, s)
		}
	}
	if vk == 0 {
		return 0, 0, fmt.Errorf("no key specified in hotkey %q", s)
	}
	mod |= MOD_NOREPEAT
	return mod, vk, nil
}
