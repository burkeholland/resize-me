---
description: "Use when working in ResizeMe/ or on the original Go/Wails implementation. Covers the Wails frontend, Go bridge, and cross-platform app behavior."
applyTo: "ResizeMe/**"
---

# Go/Wails app guidelines

- Default to `ResizeMe/` only when the request explicitly references the original Go/Wails app or its frontend.
- If the task is about macOS menu-bar behavior, native SwiftUI, or Accessibility permission flow, switch to `ResizeMeMac/` instead.
- Keep the Go/Wails frontend and bridge logic in the existing project structure under `ResizeMe/`.
- Use the existing Wails build/dev workflow for this app rather than introducing macOS-only assumptions.
- If the user says “the app” or “the windows app” without naming the target, ask which one they mean before changing code.

See also:
- [ResizeMe/README.md](../../ResizeMe/README.md)
- [ResizeMe/build/README.md](../../ResizeMe/build/README.md)
