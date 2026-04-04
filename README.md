# Centerist

A lightweight Windows utility that centers the active window with global hotkeys, or centers it and resizes it to the work area minus a configurable margin.

## Features

- **Center window**: Keeps the current size and moves the window to the center of the monitor work area.
- **Center and resize**: Centers the window and sizes it to the work area minus the margin you set.

## Default hotkeys

| Action | Hotkey |
|--------|--------|
| Center window | `Alt + C` |
| Center and resize | `Alt + Enter` |

You can change hotkeys and margin from the tray menu (**Hotkey Settings...**). Changes apply immediately; no restart required.

## Usage

1. Run `Centerist.exe`. An icon appears in the system tray.
2. Use the hotkeys on the foreground window.
3. Right-click the tray icon to change hotkeys, toggle run on startup, or quit.

## Configuration

Settings are stored in `%APPDATA%\Centerist\config.json`:

```json
{
  "margin": 50,
  "hotkeys": {
    "center": "Alt+C",
    "centerFixedSize": "Alt+Enter"
  }
}
```

Supported modifiers: `Alt`, `Ctrl`, `Shift`, `Win`.  
Supported keys include `A`–`Z`, `0`–`9`, `F1`–`F12`, `Enter`, `Space`, `Tab`, `Escape`, and arrow keys.

## Build from source

Requires Go 1.17+:

```sh
go build -ldflags "-H=windowsgui" -o Centerist.exe .
```

## Windows installer

With [Inno Setup 6](https://jrsoftware.org/isinfo.php) installed:

```text
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" installer.iss
```

The installer is written to `dist/Centerist-Setup.exe`.

## License

[GNU General Public License v3.0](LICENSE) (GPL-3.0).
