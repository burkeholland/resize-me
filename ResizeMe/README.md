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
- Applies preset width and height values as physical pixels.
- Global hotkeys use one or more modifiers plus A-Z, 0-9, or F1-F20 so shared settings work on both Windows and macOS.
- Stores presets and settings locally in `%APPDATA%\ResizeMe\settings.json`.
- Settings edits remain in the Settings window until you select **Save**. **Revert** discards the draft; if another ResizeMe action changes settings while a draft is open, saving the stale draft is rejected and the latest values are loaded.
- Does not send settings, foreground-window metadata, telemetry, or analytics over the network.
- Runs in the system tray after setup. Launch at sign-in is optional and writes a `ResizeMe` value under the current user's `Run` registry key.
- The tray menu shows a best-effort target application name from its executable. It does not show window titles, so document and page titles are not exposed in the menu; target lookup failures do not affect resizing.
- Before uninstalling, turn off **Launch at startup** to remove that registry value.
- Before resizing, ResizeMe restores minimized or maximized windows. Windows does not expose a universal fullscreen state, so fullscreen apps are resized when their window API permits it; otherwise the app reports the resize failure.
- ResizeMe keeps the title bar reachable on the target monitor when a preset is larger than its work area.
- Select **About ResizeMe** and then **Check for updates** to query published GitHub Windows releases. The check ignores draft and prerelease entries and requires the release asset that matches the running x64 or ARM64 binary.
- When an update is found, ResizeMe shows the safe winget command (`winget upgrade --id BurkeHolland.ResizeMe --exact`). It does not download or replace its own executable. GitHub release publication can precede winget availability, so retry the winget command later if winget has not indexed the release yet.
- A failed or incompatible update check is shown as an error in the About dialog; it is never reported as an up-to-date result.

## Notes

- Use this folder for the original Go/Wails implementation and frontend work.
- If the request is about macOS menu-bar behavior, accessibility permission flow, or native SwiftUI, switch to `ResizeMeMac/` instead.
- Presets support favorites and temporary hiding: starred visible presets are grouped first in the tray menu, while hidden presets remain saved in Settings and can be restored without losing their favorite state.
