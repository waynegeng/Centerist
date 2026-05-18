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

	hotkeyCapturePrompt = "Press shortcut..."
)

var (
	settingsOpenFlag int32
	hEditCenter      w32.HWND
	hEditFixed       w32.HWND
	hEditMargin      w32.HWND

	settingsWndProcCallback        = syscall.NewCallback(settingsWndProc)
	hotkeyEditSubclassProcCallback = syscall.NewCallback(hotkeyEditSubclassProc)

	hotkeyEditOrigProcs = map[w32.HWND]uintptr{}
	hotkeyEditSavedText string
)

func subclassHotkeyEdit(hwnd w32.HWND) {
	orig := w32.SetWindowLongPtr(hwnd, w32.GWLP_WNDPROC, hotkeyEditSubclassProcCallback)
	hotkeyEditOrigProcs[hwnd] = orig
}

func hotkeyEditOrigProc(hwnd w32.HWND) uintptr {
	if orig, ok := hotkeyEditOrigProcs[hwnd]; ok {
		return orig
	}
	return 0
}

func isModifierVK(vk int) bool {
	switch vk {
	case w32.VK_SHIFT, w32.VK_CONTROL, w32.VK_MENU,
		w32.VK_LSHIFT, w32.VK_RSHIFT,
		w32.VK_LCONTROL, w32.VK_RCONTROL,
		w32.VK_LMENU, w32.VK_RMENU,
		w32.VK_LWIN, w32.VK_RWIN:
		return true
	}
	return false
}

func keyboardMods() int {
	mod := 0
	if int16(w32.GetKeyState(w32.VK_CONTROL)) < 0 ||
		int16(w32.GetKeyState(w32.VK_LCONTROL)) < 0 ||
		int16(w32.GetKeyState(w32.VK_RCONTROL)) < 0 {
		mod |= MOD_CONTROL
	}
	if int16(w32.GetKeyState(w32.VK_SHIFT)) < 0 ||
		int16(w32.GetKeyState(w32.VK_LSHIFT)) < 0 ||
		int16(w32.GetKeyState(w32.VK_RSHIFT)) < 0 {
		mod |= MOD_SHIFT
	}
	if int16(w32.GetKeyState(w32.VK_MENU)) < 0 ||
		int16(w32.GetKeyState(w32.VK_LMENU)) < 0 ||
		int16(w32.GetKeyState(w32.VK_RMENU)) < 0 {
		mod |= MOD_ALT
	}
	if int16(w32.GetKeyState(w32.VK_LWIN)) < 0 ||
		int16(w32.GetKeyState(w32.VK_RWIN)) < 0 {
		mod |= MOD_WIN
	}
	return mod
}

func hotkeyEditSubclassProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	h := w32.HWND(hwnd)
	m := uint32(msg)
	orig := hotkeyEditOrigProc(h)

	switch m {
	case w32.WM_GETDLGCODE:
		if lParam != 0 {
			return w32.DLGC_WANTALLKEYS
		}
	case w32.WM_SETFOCUS:
		hotkeyEditSavedText = w32.GetWindowText(h)
		if orig != 0 {
			w32.CallWindowProc(orig, h, m, wParam, lParam)
		}
		w32.SetWindowText(h, hotkeyCapturePrompt)
		return 0
	case w32.WM_KILLFOCUS:
		if w32.GetWindowText(h) == hotkeyCapturePrompt {
			w32.SetWindowText(h, hotkeyEditSavedText)
		}
	case w32.WM_KEYDOWN, w32.WM_SYSKEYDOWN:
		vk := int(wParam)
		if vk == w32.VK_ESCAPE {
			w32.SetWindowText(h, hotkeyEditSavedText)
			return 0
		}
		if isModifierVK(vk) {
			return 0
		}
		mod := keyboardMods()
		if name, ok := formatHotkey(mod, vk); ok {
			w32.SetWindowText(h, name)
		}
		return 0
	case w32.WM_CHAR:
		return 0
	}

	if orig != 0 {
		return w32.CallWindowProc(orig, h, m, wParam, lParam)
	}
	return w32.DefWindowProc(h, m, wParam, lParam)
}

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
			430, 270,
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
		subclassHotkeyEdit(hEditCenter)

		lbl2 := w32.CreateWindowExStr(0, "STATIC", "Center & Resize:",
			w32.WS_CHILD|w32.WS_VISIBLE, 20, 57, 150, 20, hwnd, 0, inst, nil)
		setFont(lbl2)

		hEditFixed = w32.CreateWindowExStr(w32.WS_EX_CLIENTEDGE, "EDIT", cfg.Hotkeys.CenterFixedSize,
			w32.WS_CHILD|w32.WS_VISIBLE|0x00010000,
			180, 55, 220, 24, hwnd, 0, inst, nil)
		setFont(hEditFixed)
		subclassHotkeyEdit(hEditFixed)

		lblHint := w32.CreateWindowExStr(0, "STATIC", "Click a hotkey field, then press the key combination (Esc to cancel).",
			w32.WS_CHILD|w32.WS_VISIBLE, 20, 88, 380, 32, hwnd, 0, inst, nil)
		setFont(lblHint)

		lbl3 := w32.CreateWindowExStr(0, "STATIC", "Margin (px):",
			w32.WS_CHILD|w32.WS_VISIBLE, 20, 128, 150, 20, hwnd, 0, inst, nil)
		setFont(lbl3)

		hEditMargin = w32.CreateWindowExStr(w32.WS_EX_CLIENTEDGE, "EDIT", strconv.Itoa(int(cfg.Margin)),
			w32.WS_CHILD|w32.WS_VISIBLE|0x00010000,
			180, 126, 220, 24, hwnd, 0, inst, nil)
		setFont(hEditMargin)

		btnSave := w32.CreateWindowExStr(0, "BUTTON", "Save",
			w32.WS_CHILD|w32.WS_VISIBLE|0x00010000,
			230, 175, 85, 32, hwnd, w32.HMENU(idBtnSave), inst, nil)
		setFont(btnSave)

		btnCancel := w32.CreateWindowExStr(0, "BUTTON", "Cancel",
			w32.WS_CHILD|w32.WS_VISIBLE|0x00010000,
			325, 175, 85, 32, hwnd, w32.HMENU(idBtnCancel), inst, nil)
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

			if centerStr == hotkeyCapturePrompt || fixedStr == hotkeyCapturePrompt {
				w32.MessageBox(h, "Please set both hotkeys before saving.", "Centerist", w32.MB_ICONWARNING|w32.MB_OK)
				return 0
			}

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
