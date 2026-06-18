# ResizeMe

This repository contains two versions of ResizeMe:

- `ResizeMe/` — the original Go + Wails app
- `ResizeMeMac/` — the native Swift menu-bar app for macOS

If a request does not clearly say which app is meant, ask for clarification before making changes.

## Quick start

- For the native macOS app, see [ResizeMeMac/README.md](ResizeMeMac/README.md)
- For the original Go/Wails app, see [ResizeMe/README.md](ResizeMe/README.md)

## Download

Choose the build for your platform:

- [Download for macOS](https://github.com/burkeholland/resize-me/releases/latest)
- [Download for Windows](https://github.com/burkeholland/resize-me/releases/latest)

The website landing page in [docs/index.html](docs/index.html) also detects your OS and highlights the right download path automatically.

## Repository layout

- `ResizeMe/` — cross-platform Go/Wails implementation and frontend
- `ResizeMeMac/` — native SwiftUI menu-bar app for macOS
- `docs/` — release, signing, and distribution notes

## Notes

- The native macOS app is the default target for menu-bar and macOS-specific work.
- The Go/Wails app remains available for the original cross-platform implementation.
