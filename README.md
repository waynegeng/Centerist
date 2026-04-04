# Centerist

一个轻量级的 Windows 窗口居中工具。通过全局快捷键，快速将当前活动窗口居中显示，或居中并调整为指定大小。

## 功能

- **窗口居中**：保持窗口原有大小，移动到当前显示器工作区的正中央。
- **居中并调整大小**：将窗口居中，并调整为工作区大小减去指定边距。

## 默认快捷键


| 功能      | 快捷键           |
| ------- | ------------- |
| 窗口居中    | `Alt + C`     |
| 居中并调整大小 | `Alt + Enter` |


快捷键和边距均可通过托盘菜单的「快捷键设置...」修改，保存后立即生效，无需重启。

## 使用方法

1. 运行 `Centerist.exe`，程序图标会出现在系统托盘。
2. 按下快捷键即可操作当前活动窗口。
3. 右键托盘图标可以修改快捷键、设置开机启动或退出程序。

## 配置文件

配置文件位于 `%APPDATA%\Centerist\config.json`，格式如下：

```json
{
  "margin": 50,
  "hotkeys": {
    "center": "Alt+C",
    "centerFixedSize": "Alt+Enter"
  }
}
```

支持的修饰键：`Alt`、`Ctrl`、`Shift`、`Win`。
支持的按键：`A`-`Z`、`0`-`9`、`F1`-`F12`、`Enter`、`Space`、`Tab`、`Escape` 及方向键等。

## 从源码构建

需要 Go 1.17+：

```sh
go build -ldflags -H=windowsgui -o Centerist.exe .
```

## 许可证

Apache License 2.0，详见 [LICENSE](./LICENSE)。