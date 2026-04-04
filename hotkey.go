// Copyright (c) 2025 waynegeng
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see https://www.gnu.org/licenses/.

package main

import (
	"fmt"

	"github.com/gonutz/w32/v2"
	"github.com/waynegeng/Centerist/w32ex"
)

var (
	hotkeyRegistrations = make(map[int]*HotKey)
)

type HotKey struct {
	id, mod, vk int
	callback    func()
}

func (h HotKey) String() string { return fmt.Sprintf("mod=0x%x,vk=%d", h.mod, h.vk) }

func (h HotKey) Describe() string {
	var out string
	if h.mod&MOD_WIN == MOD_WIN {
		out += modKeyNames[MOD_WIN] + " + "
	}
	if h.mod&MOD_CONTROL == MOD_CONTROL {
		out += modKeyNames[MOD_CONTROL] + " + "
	}
	if h.mod&MOD_ALT == MOD_ALT {
		out += modKeyNames[MOD_ALT] + " + "
	}
	if h.mod&MOD_SHIFT == MOD_SHIFT {
		out += modKeyNames[MOD_SHIFT] + " + "
	}
	if v, ok := keyNames[h.vk]; ok {
		out += v
	} else {
		out += fmt.Sprintf("UNKNOWN KEY(0x%x)", h.vk)
	}
	return out
}

func RegisterHotKey(h HotKey) bool {
	if _, ok := hotkeyRegistrations[h.id]; ok {
		panic("hotkey id already registered")
	}
	ok := w32ex.RegisterHotKey(0, h.id, h.mod, h.vk)
	if ok {
		hotkeyRegistrations[h.id] = &h
	}
	return ok
}

func UnregisterAllHotKeys() {
	for id := range hotkeyRegistrations {
		w32ex.UnregisterHotKey(0, id)
	}
	hotkeyRegistrations = make(map[int]*HotKey)
}

func msgLoop() error {
	defer fmt.Println("event loop finished")
	for {
		var m w32.MSG
		c := w32.GetMessage(&m, 0, 0, 0)
		if c == -1 {
			return fmt.Errorf("GetMessage failed: %d", c)
		} else if c == 0 {
			// WM_QUIT received
			return nil
		}
		if m.Message == w32.WM_HOTKEY {
			h, ok := hotkeyRegistrations[int(m.WParam)]
			if !ok {
				return fmt.Errorf("hotkey without callback: %#v", m)
			}
			fmt.Printf("trace: hotkey id=%d (%s)\n", m.WParam, h)
			h.callback()
		} else if m.Message == wmAppReload {
			reloadHotkeys()
		} else {
			fmt.Printf("unhandled message received:0x%x %d\n", m.Message, m.Message)
			w32.TranslateMessage(&m)
			w32.DispatchMessage(&m)
		}
	}
}
