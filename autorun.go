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
	"errors"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	AutoRunName = `Centerist`
	regKey      = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
)

func self() string {
	return os.Args[0]
}

func AutoRunEnabled() (bool, error) {
	rk, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer rk.Close()

	v, _, err := rk.GetStringValue(AutoRunName)
	if errors.Is(err, registry.ErrNotExist) || errors.Is(err, registry.ErrUnexpectedType) {
		return false, nil
	}
	return v == self(), err
}

func AutoRunDisable() error {
	rk, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.WRITE)
	if err != nil {
		return err
	}
	defer rk.Close()

	err = rk.DeleteValue(AutoRunName)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	return err
}

func AutoRunEnable() error {
	rk, err := registry.OpenKey(registry.CURRENT_USER, regKey, registry.WRITE)
	if err != nil {
		return err
	}
	defer rk.Close()
	return rk.SetStringValue(AutoRunName, self())
}
