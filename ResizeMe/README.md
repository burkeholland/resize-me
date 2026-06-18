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

## Notes

- Use this folder for the original Go/Wails implementation and frontend work.
- If the request is about macOS menu-bar behavior, accessibility permission flow, or native SwiftUI, switch to `ResizeMeMac/` instead.
