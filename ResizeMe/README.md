# ResizeMe (Go/Wails app)

This folder contains the original Go + Wails version of ResizeMe.

If a request does not clearly say whether you mean the native macOS Swift app in `ResizeMeMac/` or this Go/Wails app, ask which one they mean before proceeding.

## About

This project uses Wails to wrap the existing Go logic with a frontend UI. It is the original cross-platform app pipeline for ResizeMe.

## Live development

Run the app in development mode from this folder:

```sh
wails dev
```

This starts the Vite frontend and the Go/Wails bridge for live iteration.

## Build

To build a production package:

```sh
wails build
```

## Windows behavior and requirements

- Requires Windows 10 version 2004 (build 19041) or later and the Microsoft Edge WebView2 Runtime.
- Resizes only the foreground program using a user-configurable global hotkey.
- Stores presets and settings locally in `%APPDATA%\ResizeMe\settings.json`.
- Does not send settings, foreground-window metadata, telemetry, or analytics over the network.
- Runs in the system tray after setup. Launch at sign-in is optional and writes a `ResizeMe` value under the current user's `Run` registry key.
- Before uninstalling, turn off **Launch at startup** to remove that registry value.

## Notes

- Use this folder for the original Go/Wails implementation and frontend work.
- If the request is about macOS menu-bar behavior, accessibility permission flow, or native SwiftUI, switch to `ResizeMeMac/` instead.
- Presets support favorites: starred presets are grouped first in the tray menu and remain editable in the Wails settings UI.
