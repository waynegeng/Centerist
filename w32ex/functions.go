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

package w32ex

import (
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

const (
	GA_PARENT    = 1
	GA_ROOT      = 2
	GA_ROOTOWNER = 3
)

var user32 = syscall.NewLazyDLL("user32.dll")

func RegisterHotKey(hwnd w32.HWND, id, mod, vk int) bool {
	r1, _, _ := user32.NewProc("RegisterHotKey").Call(uintptr(hwnd), uintptr(id), uintptr(mod), uintptr(vk))
	return r1 != 0
}

func GetDpiForWindow(hwnd w32.HWND) int32 {
	r1, _, _ := user32.NewProc("GetDpiForWindow").Call(uintptr(hwnd))
	return int32(r1)
}

func GetWindowModuleFileName(hwnd w32.HWND) string {
	var path [32768]uint16
	ret, _, _ := user32.NewProc("GetWindowModuleFileNameW").Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&path[0])),
		uintptr(len(path)),
	)
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(path[:])
}

func GetAncestor(hwnd w32.HWND, gaFlags uint) w32.HWND {
	r1, _, _ := user32.NewProc("GetAncestor").Call(uintptr(hwnd), uintptr(gaFlags))
	return w32.HWND(r1)
}

func GetShellWindow() (hwnd w32.HWND) {
	r1, _, _ := user32.NewProc("GetShellWindow").Call()
	return w32.HWND(r1)
}

func SetProcessDPIAware() bool {
	r1, _, _ := user32.NewProc("SetProcessDPIAware").Call()
	return r1 != 0
}

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

func UnregisterHotKey(hwnd w32.HWND, id int) bool {
	r1, _, _ := user32.NewProc("UnregisterHotKey").Call(uintptr(hwnd), uintptr(id))
	return r1 != 0
}

func GetCurrentThreadId() uint32 {
	r1, _, _ := kernel32.NewProc("GetCurrentThreadId").Call()
	return uint32(r1)
}

func PostThreadMessage(threadId uint32, msg uint32, wParam, lParam uintptr) bool {
	r1, _, _ := user32.NewProc("PostThreadMessageW").Call(uintptr(threadId), uintptr(msg), wParam, lParam)
	return r1 != 0
}
