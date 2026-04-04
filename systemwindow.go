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
	"strings"

	"github.com/gonutz/w32/v2"
	"github.com/waynegeng/Centerist/w32ex"
)

const (
	GWL_EXSTYLE = -20
	GWL_STYLE   = -16
)

func isZonableWindow(hwnd w32.HWND) bool {
	if hwnd == 0 {
		return false
	}
	return isStandardWindow(hwnd) && hasNoVisibleOwner(hwnd)
}

func hasNoVisibleOwner(hwnd w32.HWND) bool {
	owner := w32.GetWindow(hwnd, w32.GW_OWNER)
	if owner == 0 {
		return true
	}
	if !w32.IsWindowVisible(owner) {
		return true
	}
	rect := w32.GetWindowRect(owner)
	if rect == nil {
		return false
	}
	return rect.Width() == 0 || rect.Height() == 0
}

func isStandardWindow(hwnd w32.HWND) bool {
	// adapted from https://github.com/microsoft/PowerToys/blob/7d0304fd06939d9f552e75be9c830db22f8ff9e2/src/modules/fancyzones/FancyZonesLib/util.cpp#L403
	if w32ex.GetAncestor(hwnd, w32ex.GA_ROOT) != hwnd ||
		!w32.IsWindowVisible(hwnd) {
		return false
	}

	for _, sysWindow := range []w32.HWND{w32.GetDesktopWindow(), w32ex.GetShellWindow()} {
		if hwnd == sysWindow {
			return false
		}
	}

	style := w32.GetWindowLong(hwnd, GWL_STYLE)
	// a window with think frame and minimize/maximize buttons
	if uint32(style)&w32.WS_POPUP == w32.WS_POPUP &&
		style&w32.WS_THICKFRAME == w32.WS_THICKFRAME &&
		style&w32.WS_MINIMIZEBOX == 0 &&
		style&w32.WS_MAXIMIZEBOX == 0 {
		return false
	}
	exStyle := w32.GetWindowLong(hwnd, GWL_EXSTYLE)
	if uint32(style)&w32.WS_CHILD == w32.WS_CHILD ||
		style&w32.WS_DISABLED == w32.WS_DISABLED ||
		exStyle&w32.WS_EX_TOOLWINDOW == w32.WS_EX_TOOLWINDOW ||
		exStyle&w32.WS_EX_NOACTIVATE == w32.WS_EX_NOACTIVATE {
		return false
	}

	className, ok := w32.GetClassName(hwnd)
	if !ok {
		panic("GetClassName failed")
	}
	return !isSystemClassName(className)
}

func isSystemClassName(className string) bool {
	// adapted from https://github.com/microsoft/PowerToys/blob/7d0304fd06939d9f552e75be9c830db22f8ff9e2/tools/FancyZones_zonable_tester/main.cpp#L135
	for _, c := range []string{
		"SysListView32",
		"WorkerW",
		"Shell_TrayWnd",
		"Shell_SecondaryTrayWnd",
		"Progman",
	} {
		if strings.EqualFold(c, className) {
			return true
		}
	}
	return false
}
