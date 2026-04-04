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

// TODO make it possible to "go generate" on Windows (https://github.com/josephspurrier/goversioninfo/issues/52).
//go:generate /bin/bash -c "go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest -arm -64 -icon=assets/icon.ico - <<< '{}'"

package main

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"

	"github.com/getlantern/systray"
	"github.com/gonutz/w32/v2"

	"github.com/waynegeng/Centerist/w32ex"
)

var mainThreadID uint32

func main() {
	runtime.LockOSThread()
	mainThreadID = w32ex.GetCurrentThreadId()

	if !w32ex.SetProcessDPIAware() {
		panic("failed to set DPI aware")
	}

	autorun, err := AutoRunEnabled()
	if err != nil {
		panic(err)
	}
	fmt.Printf("autorun enabled=%v\n", autorun)
	printMonitors()

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("warn: config: %v, using defaults\n", err)
	}
	fmt.Printf("config: margin=%d, center=%q, centerFixedSize=%q\n",
		cfg.Margin, cfg.Hotkeys.Center, cfg.Hotkeys.CenterFixedSize)

	registerHotkeysFromConfig(cfg)

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt)
	go func() {
		<-exitCh
		fmt.Println("exit signal received")
		systray.Quit()
	}()

	initTray()
	if err := msgLoop(); err != nil {
		panic(err)
	}
}

func registerHotkeysFromConfig(cfg Config) {
	centerMod, centerVK, err := parseHotkey(cfg.Hotkeys.Center)
	if err != nil {
		fmt.Printf("error: invalid center hotkey %q: %v\n", cfg.Hotkeys.Center, err)
		showMessageBox(fmt.Sprintf("Invalid center hotkey %q: %v", cfg.Hotkeys.Center, err))
		return
	}
	fixedMod, fixedVK, err := parseHotkey(cfg.Hotkeys.CenterFixedSize)
	if err != nil {
		fmt.Printf("error: invalid centerFixedSize hotkey %q: %v\n", cfg.Hotkeys.CenterFixedSize, err)
		showMessageBox(fmt.Sprintf("Invalid centerFixedSize hotkey %q: %v", cfg.Hotkeys.CenterFixedSize, err))
		return
	}

	fixedMarginFunc := centerFixedMargin(cfg.Margin)

	hks := []HotKey{
		{id: 1, mod: centerMod, vk: centerVK, callback: func() {
			if _, err := resize(w32.GetForegroundWindow(), center); err != nil {
				fmt.Printf("warn: center: %v\n", err)
			}
		}},
		{id: 2, mod: fixedMod, vk: fixedVK, callback: func() {
			if _, err := resize(w32.GetForegroundWindow(), fixedMarginFunc); err != nil {
				fmt.Printf("warn: centerFixedSize: %v\n", err)
			}
		}},
	}

	var failedHotKeys []HotKey
	for _, hk := range hks {
		if !RegisterHotKey(hk) {
			failedHotKeys = append(failedHotKeys, hk)
		}
	}
	if len(failedHotKeys) > 0 {
		msg := "The following hotkey(s) are in use by another process:\n\n"
		for _, hk := range failedHotKeys {
			msg += "  - " + hk.Describe() + "\n"
		}
		msg += "\nTo use these hotkeys in Centerist, close the other process using the key combination(s)."
		showMessageBox(msg)
	}
}

func reloadHotkeys() {
	UnregisterAllHotKeys()
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("warn: reload config: %v\n", err)
		return
	}
	fmt.Printf("reloading hotkeys: margin=%d, center=%q, centerFixedSize=%q\n",
		cfg.Margin, cfg.Hotkeys.Center, cfg.Hotkeys.CenterFixedSize)
	registerHotkeysFromConfig(cfg)
}

func showMessageBox(text string) {
	w32.MessageBox(w32.GetActiveWindow(), text, "Centerist", w32.MB_ICONWARNING|w32.MB_OK)
}

type resizeFunc func(disp, cur w32.RECT) w32.RECT

func center(disp, cur w32.RECT) w32.RECT {
	w := (disp.Width() - cur.Width()) / 2
	h := (disp.Height() - cur.Height()) / 2
	return w32.RECT{
		Left:   disp.Left + w,
		Right:  disp.Left + w + cur.Width(),
		Top:    disp.Top + h,
		Bottom: disp.Top + h + cur.Height()}
}

func centerFixedMargin(margin int32) resizeFunc {
	return func(disp, _ w32.RECT) w32.RECT {
		return w32.RECT{
			Left:   disp.Left + margin,
			Top:    disp.Top + margin,
			Right:  disp.Right - margin,
			Bottom: disp.Bottom - margin,
		}
	}
}

func resize(hwnd w32.HWND, f resizeFunc) (bool, error) {
	if !isZonableWindow(hwnd) {
		fmt.Printf("warn: non-zonable window: %s\n", w32.GetWindowText(hwnd))
		return false, nil
	}
	rect := w32.GetWindowRect(hwnd)
	mon := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	hdc := w32.GetDC(hwnd)
	displayDPI := w32.GetDeviceCaps(hdc, w32.LOGPIXELSY)
	if !w32.ReleaseDC(hwnd, hdc) {
		return false, fmt.Errorf("failed to ReleaseDC:%d", w32.GetLastError())
	}
	var monInfo w32.MONITORINFO
	if !w32.GetMonitorInfo(mon, &monInfo) {
		return false, fmt.Errorf("failed to GetMonitorInfo:%d", w32.GetLastError())
	}

	ok, frame := w32.DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS(hwnd)
	if !ok {
		return false, fmt.Errorf("failed to DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS:%d", w32.GetLastError())
	}
	windowDPI := w32ex.GetDpiForWindow(hwnd)
	resizedFrame := resizeForDpi(frame, int32(windowDPI), int32(displayDPI))

	fmt.Printf("> window: 0x%x %#v (w:%d,h:%d) mon=0x%X(@ display DPI:%d)\n", hwnd, rect, rect.Width(), rect.Height(), mon, displayDPI)
	fmt.Printf("> DWM frame:        %#v (W:%d,H:%d) @ window DPI=%v\n", frame, frame.Width(), frame.Height(), windowDPI)
	fmt.Printf("> DPI-less frame:   %#v (W:%d,H:%d)\n", resizedFrame, resizedFrame.Width(), resizedFrame.Height())

	lExtra := resizedFrame.Left - rect.Left
	rExtra := -resizedFrame.Right + rect.Right
	tExtra := resizedFrame.Top - rect.Top
	bExtra := -resizedFrame.Bottom + rect.Bottom

	newPos := f(monInfo.RcWork, resizedFrame)

	newPos.Left -= lExtra
	newPos.Top -= tExtra
	newPos.Right += rExtra
	newPos.Bottom += bExtra

	if sameRect(rect, &newPos) {
		fmt.Println("no resize")
		return false, nil
	}

	fmt.Printf("> resizing to: %#v (W:%d,H:%d)\n", newPos, newPos.Width(), newPos.Height())
	if !w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL) {
		return false, fmt.Errorf("failed to normalize window ShowWindow:%d", w32.GetLastError())
	}
	if !w32.SetWindowPos(hwnd, 0, int(newPos.Left), int(newPos.Top), int(newPos.Width()), int(newPos.Height()), w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
		return false, fmt.Errorf("failed to SetWindowPos:%d", w32.GetLastError())
	}
	rect = w32.GetWindowRect(hwnd)
	fmt.Printf("> post-resize: %#v(W:%d,H:%d)\n", rect, rect.Width(), rect.Height())
	return true, nil
}

func resizeForDpi(src w32.RECT, from, to int32) w32.RECT {
	return w32.RECT{
		Left:   src.Left * to / from,
		Right:  src.Right * to / from,
		Top:    src.Top * to / from,
		Bottom: src.Bottom * to / from,
	}
}

func sameRect(a, b *w32.RECT) bool {
	return a != nil && b != nil && reflect.DeepEqual(*a, *b)
}
