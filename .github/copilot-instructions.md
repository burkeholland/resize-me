# ResizeMe macOS app instructions

Focus your changes on `ResizeMeMac/` unless the user asks for another target.

## Architecture and state
- Keep the source of truth in `AppState` and `AppConfig`.
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
