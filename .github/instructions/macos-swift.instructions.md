---
description: "Use when working in ResizeMeMac/ or on the native Swift menu-bar app for macOS. Covers SwiftUI, accessibility, persistence, update behavior, and macOS-specific conventions."
applyTo: "ResizeMeMac/**"
---

# macOS Swift app guidelines

- Default to `ResizeMeMac/` for macOS-native work, menu-bar behavior, onboarding, and Accessibility permission flows.
- Keep the source of truth in `AppState` and `AppConfig`.
- Preserve `ConfigNormalizer` as the validation boundary for persisted settings.
- Reuse existing services (`ResizeService`, `PermissionService`, `SettingsStore`, `SparkleUpdateService`) instead of duplicating platform logic in views.
- Prefer explicit failure handling and user-visible status updates via `AppState.lastStatusMessage`.
- Use SwiftUI `popover(item:)` rather than `popover(isPresented:)` for data-dependent popovers to avoid state race conditions.
- Keep menu actions lightweight and delegate side effects to services or `AppState`.
- Maintain accessibility-related guidance and permission checks before resize actions.
- Treat `settings.json` compatibility as a hard requirement; load with tolerant decoding and normalize to the current schema.
- When extending config, include backward-compatible decoding defaults and update tests under `ResizeMeMac/Tests/`.
- Do not trigger Sparkle checks in debug builds.
- Keep update actions gated by `SparkleUpdateService.canCheckForUpdates` so debug/dev builds cannot accidentally hit release appcasts.
- Build and test with the existing Xcode workflow in `ResizeMeMac/`.

See also:
- [ResizeMeMac/README.md](../../ResizeMeMac/README.md)
- [docs/macos-build-sign-release.md](../../docs/macos-build-sign-release.md)
- [docs/macos-release-checklist.md](../../docs/macos-release-checklist.md)
