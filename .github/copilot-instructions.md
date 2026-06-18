# ResizeMe repository instructions

This repository contains two related apps:
- `ResizeMe/` — the original Go/Wails app
- `ResizeMeMac/` — the native Swift menu-bar app for macOS

If a request does not clearly say which app is meant, ask for clarification before making changes.

## Target selection
- Default to `ResizeMeMac/` for macOS-native work and menu-bar behavior.
- Use `ResizeMe/` only when the request explicitly references the original Go/Wails app or the cross-platform frontend.
- If the user says “the app” or “the window app” without naming the target, ask which one they mean.

## Architecture and state
- Keep the source of truth in `AppState` and `AppConfig` for the Swift app.
- Preserve `ConfigNormalizer` as the validation boundary for persisted settings.
- Reuse existing services (`ResizeService`, `PermissionService`, `SettingsStore`, `SparkleUpdateService`) instead of duplicating platform logic in views.
- Prefer explicit failure handling and user-visible status updates via `AppState.lastStatusMessage`.

## SwiftUI and menu bar behavior
- Keep menu actions lightweight and delegate side effects to services or `AppState`.
- For macOS UX, use menu-friendly controls and avoid adding complex modal flows for simple actions.
- Maintain accessibility-related guidance and permission checks before resize actions.

## Persistence and compatibility
- Treat `settings.json` compatibility as a hard requirement; load with tolerant decoding and normalize to the current schema.
- When extending config, include backward-compatible decoding defaults and update tests under `ResizeMeMac/Tests/`.

## Updates and build mode
- Do not trigger Sparkle checks in debug builds.
- Keep update actions gated by `SparkleUpdateService.canCheckForUpdates` so debug/dev builds cannot accidentally hit release appcasts.

## Quality bar
- Prefer surgical changes that match current naming/patterns.
- Add or update tests when behavior changes.
- Keep code ASCII unless an existing file already requires Unicode content.
