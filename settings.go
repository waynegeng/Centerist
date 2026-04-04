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
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
	"github.com/waynegeng/Centerist/w32ex"
)

const (
	settingsWndClass = "CenteristSettings"
	wmAppReload      = 0x8000 + 1 // WM_APP + 1

	idBtnSave   = 1001
	idBtnCancel = 1002
)

var (
	settingsOpenFlag int32
	hEditCenter      w32.HWND
	hEditFixed       w32.HWND
	hEditMargin      w32.HWND

	settingsWndProcCallback = syscall.NewCallback(settingsWndProc)
)

func showSettingsDialog() {
	if !atomic.CompareAndSwapInt32(&settingsOpenFlag, 0, 1) {
		return
	}

	go func() {
		runtime.LockOSThread()
		defer func() {
			atomic.StoreInt32(&settingsOpenFlag, 0)
			runtime.UnlockOSThread()
		}()

		cfg, _ := loadConfig()
		inst := w32.GetModuleHandle("")

		className, _ := syscall.UTF16PtrFromString(settingsWndClass)
		var wc w32.WNDCLASSEX
		wc.Size = uint32(unsafe.Sizeof(wc))
		wc.WndProc = settingsWndProcCallback
		wc.Instance = inst
		wc.ClassName = className
		wc.Background = w32.HBRUSH(16) // COLOR_BTNFACE + 1
		wc.Cursor = w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_ARROW))
		w32.RegisterClassEx(&wc)

		hwnd := w32.CreateWindowExStr(
			0,
			settingsWndClass,
			"Centerist - Hotkey Settings",
			0x00C80000,             // WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU
			0x80000000, 0x80000000, // CW_USEDEFAULT
			430, 240,
			0, 0, inst, nil,
		)
		if hwnd == 0 {
			fmt.Printf("failed to create settings window: %d\n", w32.GetLastError())
			return
		}

		hFont := uintptr(w32.GetStockObject(w32.DEFAULT_GUI_FONT))
		setFont := func(h w32.HWND) {
			w32.SendMessage(h, 0x0030, hFont, 1) // WM_SETFONT
		}

		lbl1 := w32.CreateWindowExStr(0, "STATIC", "Center window:",
			w32.WS_CHILD|w32.WS_VISIBLE, 20, 22, 150, 20, hwnd, 0, inst, nil)
		setFont(lbl1)

		hEditCenter = w32.CreateWindowExStr(w32.WS_EX_CLIENTEDGE, "EDIT", cfg.Hotkeys.Center,
			w32.WS_CHILD|w32.WS_VISIBLE|0x00010000, // WS_TABSTOP
			180, 20, 220, 24, hwnd, 0, inst, nil)
		setFont(hEditCenter)

		lbl2 := w32.CreateWindowExStr(0, "STATIC", "Center & Resize:",
			w32.WS_CHILD|w32.WS_VISIBLE, 20, 57, 150, 20, hwnd, 0, inst, nil)
		setFont(lbl2)

		hEditFixed = w32.CreateWindowExStr(w32.WS_EX_CLIENTEDGE, "EDIT", cfg.Hotkeys.CenterFixedSize,
			w32.WS_CHILD|w32.WS_VISIBLE|0x00010000,
			180, 55, 220, 24, hwnd, 0, inst, nil)
		setFont(hEditFixed)

		lbl3 := w32.CreateWindowExStr(0, "STATIC", "Margin (px):",
			w32.WS_CHILD|w32.WS_VISIBLE, 20, 92, 150, 20, hwnd, 0, inst, nil)
		setFont(lbl3)

		hEditMargin = w32.CreateWindowExStr(w32.WS_EX_CLIENTEDGE, "EDIT", strconv.Itoa(int(cfg.Margin)),
			w32.WS_CHILD|w32.WS_VISIBLE|0x00010000,
			180, 90, 220, 24, hwnd, 0, inst, nil)
		setFont(hEditMargin)

		btnSave := w32.CreateWindowExStr(0, "BUTTON", "Save",
			w32.WS_CHILD|w32.WS_VISIBLE|0x00010000,
			230, 140, 85, 32, hwnd, w32.HMENU(idBtnSave), inst, nil)
		setFont(btnSave)

		btnCancel := w32.CreateWindowExStr(0, "BUTTON", "Cancel",
			w32.WS_CHILD|w32.WS_VISIBLE|0x00010000,
			325, 140, 85, 32, hwnd, w32.HMENU(idBtnCancel), inst, nil)
		setFont(btnCancel)

		w32.ShowWindow(hwnd, w32.SW_SHOW)
		w32.UpdateWindow(hwnd)

		var msg w32.MSG
		for w32.GetMessage(&msg, 0, 0, 0) > 0 {
			w32.TranslateMessage(&msg)
			w32.DispatchMessage(&msg)
		}
	}()
}

func settingsWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	h := w32.HWND(hwnd)
	m := uint32(msg)

	switch m {
	case w32.WM_COMMAND:
		switch int(wParam & 0xFFFF) {
		case idBtnSave:
			centerStr := w32.GetWindowText(hEditCenter)
			fixedStr := w32.GetWindowText(hEditFixed)
			marginStr := w32.GetWindowText(hEditMargin)

			if _, _, err := parseHotkey(centerStr); err != nil {
				w32.MessageBox(h, fmt.Sprintf("Invalid center hotkey: %v", err), "Centerist", w32.MB_ICONWARNING|w32.MB_OK)
				return 0
			}
			if _, _, err := parseHotkey(fixedStr); err != nil {
				w32.MessageBox(h, fmt.Sprintf("Invalid resize hotkey: %v", err), "Centerist", w32.MB_ICONWARNING|w32.MB_OK)
				return 0
			}
			margin, err := strconv.ParseInt(marginStr, 10, 32)
			if err != nil || margin < 0 {
				w32.MessageBox(h, "Invalid margin, enter a non-negative integer", "Centerist", w32.MB_ICONWARNING|w32.MB_OK)
				return 0
			}

			cfg := Config{
				Margin: int32(margin),
				Hotkeys: HotkeyConfig{
					Center:          centerStr,
					CenterFixedSize: fixedStr,
				},
			}
			if err := saveConfig(cfg); err != nil {
				w32.MessageBox(h, fmt.Sprintf("Failed to save: %v", err), "Centerist", w32.MB_ICONERROR|w32.MB_OK)
				return 0
			}

			w32ex.PostThreadMessage(mainThreadID, wmAppReload, 0, 0)
			w32.DestroyWindow(h)

		case idBtnCancel:
			w32.DestroyWindow(h)
		}
		return 0

	case w32.WM_CLOSE:
		w32.DestroyWindow(h)
		return 0

	case w32.WM_DESTROY:
		w32.PostQuitMessage(0)
		return 0
	}

	return w32.DefWindowProc(h, m, wParam, lParam)
}
