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
	"syscall"

	"github.com/gonutz/w32/v2"
	"golang.org/x/sys/windows"
)

func EnumMonitors(f func(d w32.HMONITOR) bool) bool {
	callback := syscall.NewCallback(func(h, _, _, _ uintptr) uintptr {
		if f(w32.HMONITOR(h)) {
			return 1
		}
		return 0
	})
	return w32.EnumDisplayMonitors(0, nil, callback, 0)
}

func printMonitors() {
	i := 0
	EnumMonitors(func(d w32.HMONITOR) bool {
		var v w32.MONITORINFO
		if !w32.GetMonitorInfo(d, &v) {
			return false
		}
		fmt.Printf("> monitor#%d: 0x%x\n", i, d)
		i++
		fmt.Printf("       rcwork:%#v (w=%v,h=%v)\n", v.RcWork, v.RcWork.Width(), v.RcWork.Height())
		fmt.Printf("    rcmonitor:%#v (w=%v,h=%v)\n", v.RcMonitor, v.RcMonitor.Width(), v.RcWork.Height())
		fmt.Printf("      primary:%#v\n", v.DwFlags&w32.MONITORINFOF_PRIMARY > 0)

		ok, n := w32.GetNumberOfPhysicalMonitorsFromHMONITOR(d)
		if !ok {
			fmt.Printf("  physical monitors: failed to query count: %d\n", w32.GetLastError())
		} else {
			fmt.Printf("  physical monitors: %d\n", n)
			pMon := make([]w32.PHYSICAL_MONITOR, n)
			if !w32.GetPhysicalMonitorsFromHMONITOR(d, pMon) {
				fmt.Printf("  physical monitors: failed to get physical monitors: %d\n", w32.GetLastError())
			} else {
				for i, p := range pMon {
					name := windows.UTF16ToString(p.Description[:])
					fmt.Printf("  > physical monitor#%d: %s\n", i, name)
				}
			}
		}
		return true
	})
}
