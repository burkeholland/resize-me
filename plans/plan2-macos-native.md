## Plan: Native Swift macOS Menu Bar App for ResizeMe

**TL;DR**: Create a new Swift macOS menu bar application alongside the existing Go/Wails version. v1 is deliberately small: a menu bar app with one global hotkey, the existing preset list, Accessibility-based window resizing with a correct multi-display geometry model, a validated JSON settings store compatible with the existing schema, launch at login, and a clean permission flow. Extras (per-preset hotkeys, undo, percentage presets, anchors) and the full release pipeline (Sparkle, Homebrew) are explicitly staged for after the core is proven.

### Goals
1. Build a native macOS menu bar app that matches the core ResizeMe experience.
2. Prove the hard parts first: Accessibility permission, AX resize reliability, and multi-display geometry.
3. Keep the settings model schema-compatible with the existing Go app for future upstream contribution.
4. Stage distribution so signing/notarization come before Sparkle and Homebrew.

### Key decisions
- Create a new folder in this repo: `ResizeMeMac/`
- Use native SwiftUI/AppKit menu bar integration rather than Wails — a hybrid: SwiftUI for views/settings; an AppKit `NSApplicationDelegate`/app coordinator for activation policy, Settings window activation, About panel, and quit (required for a reliable `LSUIElement` app)
- Menu bar UI: start with SwiftUI `MenuBarExtra` (menu style); fall back to `NSStatusItem` + `NSMenu` if `MenuBarExtra` cannot support menu refresh-on-open, Sparkle menu items, or activation edge cases. Treat this as an explicit early checkpoint, not an assumption.
- Settings: a single `Codable` `AppConfig` persisted as JSON in `~/Library/Application Support/ResizeMe/settings.json` — NOT scattered `@AppStorage` keys. Field names and defaults stay compatible with the Go app's `settings.json` schema (`presets`, `activePresetId`, `centerAfterResize`, `hotkey`, `autoStart`, `firstRun`) plus a `schemaVersion` for migration. `@AppStorage` is allowed only for trivial UI-only preferences. The JSON `hotkey` field is the single source of truth for the shortcut (see Phase 4 for how KeyboardShortcuts is wired to it without using the library's own `UserDefaults` storage).
- **Sizing semantics**: preset width/height are applied as macOS **points** (logical units), not physical pixels. AX APIs operate in points; on Retina a 1920×1080 preset produces a 1920×1080-point window. Settings UI copy will say "window size (points)". A pixel-aware mode using `NSScreen.backingScaleFactor` is a possible later feature, not v1.
- Dependencies: KeyboardShortcuts (sindresorhus) only for v1 implementation. Sparkle is added in the release phase, not Phase 1.
- Target macOS 14 (Sonoma) as the baseline: `MenuBarExtra` and `SMAppService` require only 13+, and 14 gives a pragmatic floor for SwiftUI Settings/`@Observable` without restricting reach the way a 15-only floor would.
- Distribution: Developer ID signed, Hardened Runtime enabled, **App Sandbox disabled** (Accessibility control of other apps is incompatible with sandboxing), notarized direct download. Mac App Store is out of scope.

### Product identity
- Bundle ID: `com.resizeme.mac` placeholder — **confirm the final bundle ID with the upstream maintainer before asking any testers to grant Accessibility** (the permission is granted per signed bundle identity, so changing bundle ID or signing identity later resets every grant).
- `LSUIElement = YES` (no Dock icon).
- Hardened Runtime on; no App Sandbox entitlement; no app group.
- Signing: stable self-signed/dev certificate during development (avoids re-prompting for Accessibility after each rebuild), Developer ID for release.

### v1 scope (explicit)
**In:** menu bar app, one configurable global hotkey (default ⌃⌥R), default preset list matching the Go app, active preset selection, resize frontmost window, center-after-resize toggle, oversized-preset clamping, settings window, launch at login, Accessibility permission flow, About/Quit, OSLog diagnostics, signed+notarized .dmg/.zip release.

**Out (v1.1+):** per-preset hotkeys, undo last resize, percentage-of-screen presets, position anchors (left half, right half, corners), frontmost window *title* in the menu (frontmost *app name* via `NSWorkspace` is allowed as best-effort polish since it needs no AX), Sparkle auto-updates, Homebrew cask, pixel-aware sizing mode, notification-permission-based alerts.

### Phase 1 — Scaffold + permission + resize spike (the risky part first)
1. Create an Xcode project under `ResizeMeMac/` with the structure:
   - `ResizeMeApp.swift` (SwiftUI entry + AppKit delegate adaptor)
   - `Models/` (`AppConfig`, `Preset`)
   - `Services/` (`SettingsStore`, `ResizeService`, `WindowGeometryService`, `PermissionService`, `HotkeyService`, `LaunchAtLoginService`)
   - `Views/`, `Resources/`
2. Configure `LSUIElement`, Hardened Runtime, no sandbox; add the KeyboardShortcuts package.
3. Implement `PermissionService`:
   - `AXIsProcessTrusted()` checked on launch and before every resize.
   - `AXIsProcessTrustedWithOptions([kAXTrustedCheckOptionPrompt: true])` invoked only from an explicit user action.
   - "Open System Settings" deep link to Privacy & Security → Accessibility (`x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility`) — treat as best-effort since these URLs are not a stable API contract; if it fails to land on the right pane, fall back to opening Privacy & Security plus showing manual navigation text.
   - Re-check trust on `NSApplication.didBecomeActiveNotification` / menu open (no restart required after the user grants permission).
4. Implement the minimal resize spike — this includes throwaway versions of the core AX path so the checkpoint is actually buildable (Phase 2 then hardens these into the real services):
   - frontmost-app lookup → AX app element → focused-window read,
   - raw `kAXPositionAttribute`/`kAXSizeAttribute` read/write,
   - a disposable single-screen center/clamp helper.
   A temporary menu item "Resize Now" applies one hardcoded preset to the frontmost window. **Checkpoint:** verify against TextEdit, Safari, Finder, Terminal, and one Electron app (e.g., VS Code) on two displays before building anything else. Probe at this checkpoint whether target apps expose a usable raw `"AXFullScreen"` attribute (see Phase 2 — it is not a guaranteed public constant). Also validate that `MenuBarExtra` supports menu rebuild-on-open and Settings-window activation; if not, switch to `NSStatusItem` now.

### Phase 2 — Resize engine done properly
1. `ResizeService` algorithm:
   - Verify AX trust; surface `permissionMissing` if not.
   - `NSWorkspace.shared.frontmostApplication` → app AX element from PID.
   - Read `kAXFocusedWindowAttribute`; fall back to first standard window in `kAXWindowsAttribute`; validate role/subrole (`AXWindow`/`AXStandardWindow`).
   - Reject minimized and system/protected windows with specific errors. Fullscreen detection is **best-effort**: there is no guaranteed public AX constant for "is fullscreen" — probe the raw `"AXFullScreen"` attribute (validated per-app in the Phase 1 spike); when it isn't exposed, fall back to reporting `resizeRejected` when the geometry write fails or is ignored.
   - Read current `kAXPositionAttribute`/`kAXSizeAttribute`; compute target rect via `WindowGeometryService`; write size and position using `AXValueCreate(.cgSize/.cgPoint, …)`. Write ordering is a heuristic, not a rule: prefer position-then-size when growing and size-then-position when shrinking; for mixed cases or when the post-write re-read shows clamping, retry once with the alternate order.
   - Re-read actual geometry after the write; if the app constrained the size (min/max window limits), report partial success with the achieved size.
   - Error taxonomy: `permissionMissing`, `noFrontmostApp`, `noResizableWindow`, `windowFullscreen`, `windowMinimized`, `resizeRejected`, `sizeConstrained(actual:)`.
2. `WindowGeometryService` (single tested home for all coordinate math):
   - Coordinate-conversion invariant: AX uses a top-left-origin global space; AppKit uses bottom-left-origin. Convert with a **fixed global reference — the AppKit `frame.maxY` of the zero-origin screen** (the screen whose `frame.origin == .zero` in `NSScreen.screens`; **never `NSScreen.main`**, which follows the key window and is unreliable in an `LSUIElement` app). Recompute the reference on display-configuration-change notifications. Formula: `appKitY = primaryMaxY - axY - height`, and the inverse for writes. All incoming AX rects are converted to AppKit space *before* any screen-overlap math; the target rect is converted back to AX space only at the final write. Unit-test this against simulated displays left of, above, and below the primary.
   - Determine the screen containing the target window (greatest-overlap rule against `NSScreen.screens`, computed in AppKit space).
   - Use that screen's `visibleFrame` (excludes menu bar and Dock).
   - Centering: center within the window's current screen `visibleFrame`.
   - **Oversized presets** (e.g., 4K on a smaller display): apply the requested size, then clamp position so the window's top-left and title bar remain on-screen within `visibleFrame`. Predictable, no silent scaling, no warning dialogs.
   - Handle displays left of / above the primary (negative origins).
3. Error surfacing: inline menu state and a transient status (menu bar icon state or lightweight alert for permission errors) — **no** UserNotifications framework and no notification permission prompt in v1.
4. OSLog categories from day one: `permissions`, `hotkeys`, `resize`, `settings`, `loginItem`. Log AX error codes and app bundle IDs; do not log window titles.

### Phase 3 — Settings model with parity semantics
1. `AppConfig` (`Codable`): `schemaVersion`, `presets[] {id, name, width, height}`, `activePresetId`, `centerAfterResize`, `hotkey`, `autoStart`, `firstRun`. Same field names/defaults as the Go app; same default preset list (360p→4K landscape + portrait variants). Migration rule: a file with **missing `schemaVersion` is treated as v1** (a Go-written file); `loadError` is runtime-only and never persisted (matches Go's `omitempty` behavior).
2. `SettingsStore`:
   - Load → decode → normalize, porting the Go `NormalizeConfig` rules **exactly** (verified against `config_normalize.go`): empty/invalid hotkey → default; empty preset list → defaults; preset name trimmed, empty name → `"WxH"`; width/height outside 100–10000 is an **error** (the whole config is rejected, not clamped — load then keeps prior/default config and surfaces the error); empty preset ID → slugified-name ID or `preset-N`; **duplicate IDs get `-2`, `-3` suffixes** (never dropped); unknown `activePresetId` → fallback's active ID if still present, else first preset.
   - Load errors are non-fatal: keep defaults, surface the message in Settings UI (same as Go's `loadError`).
   - Save pipeline mirrors the Go transactional semantics: normalize → apply side effects (hotkey registration, login item) → persist via temp file + atomic rename → roll back side effects on persist failure → publish state.
3. `LaunchAtLoginService` via `SMAppService.mainApp`:
   - `register()`/`unregister()`, with UI driven by actual `status` (`.enabled`, `.requiresApproval`, `.notRegistered`).
   - Reconcile persisted `autoStart` with real registration status on launch; surface `.requiresApproval` in Settings using the dedicated API `SMAppService.openSystemSettingsLoginItems()` (not a hand-built URL).
4. Unit tests (XCTest): normalization rules, default-preset parity with the Go list, active-preset fallback, atomic save behavior, geometry math (target rect, screen selection, coordinate conversion, clamping) — geometry is pure math and fully testable without AX.

### Phase 4 — Hotkey + menu bar experience
1. `HotkeyService` using KeyboardShortcuts:
   - Single shortcut, default `⌃⌥R` (matches the Windows default Ctrl+Alt+R conceptually).
   - **Storage design preserves the transactional save:** do NOT use the `Name`-based `Recorder(name:)` API, which writes to `UserDefaults` immediately and would bypass the Phase 3 save/rollback pipeline. Instead use the binding-based `Recorder(shortcut: Binding<Shortcut?>)` so the recorded shortcut is staged in view state and flows through normalize → apply → persist. The JSON `hotkey` field is the single source of truth; a mapping layer converts the config string ↔ `KeyboardShortcuts.Shortcut`.
   - **Listener lifetime:** `KeyboardShortcuts.events(for:)` is an `AsyncStream`, so `HotkeyService` owns exactly one listener `Task`. Applying a shortcut cancels the existing task and starts a new `events(.keyUp, for: shortcut)` task (none when the shortcut is nil); if persistence fails, rollback restores the previous shortcut's task. This prevents stale/duplicate listeners after re-record, clear, or failed save.
   - Handle cleared/nil shortcuts (hotkey disabled state shown in menu and Settings); note in Settings help that global shortcuts can be suppressed by Secure Keyboard Entry.
2. Menu bar menu (rebuilt on open):
   - Frontmost app name (best-effort via `NSWorkspace`, no AX needed) as a disabled context line.
   - Active preset checkmark + preset picker (selecting a preset makes it active; does not resize).
   - "Resize Now" (disabled with explanatory state when permission is missing — routes to permission UI).
   - Settings…, About, Quit. (Check for Updates… appears only when Sparkle lands in the release phase.)
3. AppKit lifecycle: `NSApp.activate(ignoringOtherApps: true)` when opening Settings; standard About panel; verify Settings window focus, close-all-windows behavior, and ⌘Q handling under `LSUIElement`.

### Phase 5 — Settings window + first run
1. SwiftUI Settings with tabs: General (launch at login, center after resize), Presets (add/edit/delete with width/height validation in points), Shortcuts (KeyboardShortcuts recorder), About (version, link to repo). An Updates tab is added with Sparkle in the release phase.
2. First-run onboarding (because `firstRun` is in the schema): a single window explaining what the app does, why Accessibility permission is required, a button that triggers the AX prompt, live status that flips when permission is granted (re-check on activation), and a launch-at-login offer. Clean "permission denied" path: app remains usable for editing presets/settings; resize actions route to permission guidance.

### Phase 6 — Release (staged)
1. Stage 1 (v1 ship requirement): GitHub Actions build, Developer ID signing, Hardened Runtime, notarization + stapling, `.dmg`/`.zip` artifact on GitHub Releases.
2. Stage 2 (post-v1): Sparkle — add the package, EdDSA key generation (private key in repo secrets), appcast generation/hosting on GitHub Pages or Releases, `SUFeedURL`/`SUPublicEDKey` in Info.plist, Updates tab + "Check for Updates…" menu item, test update flow from a versioned build.
3. Stage 3 (post-v1): Homebrew cask once download URLs and versioning are stable and upstream ownership (bundle ID, signing identity, hosting) is settled.
4. The Tiny Clips project is *inspiration only* for the Actions workflow and appcast layout; every adopted piece (workflow file, signing steps, appcast script) is copied deliberately into this repo and documented — no implicit dependency on that project's structure.

### Verification checklist
1. Unit tests pass: normalization, preset parity, geometry math, coordinate conversion, clamping.
2. Manual AX matrix: TextEdit, Safari, Chrome, Terminal, Finder, VS Code/Electron — resize succeeds and centers correctly.
3. Multi-display: external display left of and above primary; Retina + non-Retina; Dock on left/bottom; menu bar on secondary display.
4. Edge cases: fullscreen window (clear error where detectable, otherwise graceful `resizeRejected`), minimized window (clear error), no windows (clear error), oversized preset on small display (clamped, title bar visible), app with min/max size constraints (partial-success report).
5. Permission lifecycle: fresh install prompt → grant without relaunch → revoke while running (next resize shows guidance) → re-grant.
6. Hotkey: default ⌃⌥R fires; re-record works; cleared shortcut shows disabled state; conflict with an OS shortcut is changeable.
7. Launch at login: toggle on/off reflects in System Settings Login Items; `.requiresApproval` state is surfaced; survives reboot.
8. Settings: an interrupted save never leaves malformed JSON at the final path (leftover temp files are ignored/cleaned on next launch); invalid JSON on disk loads defaults and surfaces the error in Settings.
9. `LSUIElement` behavior: no Dock icon, Settings window activates in front, About panel opens, ⌘Q from Settings quits cleanly.
10. Release: notarized artifact passes Gatekeeper (`spctl` assessment) on a clean machine.

### Risks and notes
- AX resizing is inconsistent across apps; the error taxonomy and partial-success reporting exist to make this visible rather than silent.
- Accessibility permission is tied to signed bundle identity — dev-build re-signing churn can repeatedly invalidate the grant (mitigate with a stable dev signing identity).
- Some apps (cross-platform toolkits, games, system apps) may reject AX writes entirely; documented as known limitations.
- `MenuBarExtra` may prove too limited; the Phase 1 checkpoint exists so a swap to `NSStatusItem` costs little.
- Sparkle/notarization/Homebrew need certificates and repo secrets owned by the upstream maintainer; that is why they are staged post-core and post-v1 respectively.

### Suggested next step
Start Phase 1: scaffold `ResizeMeMac/`, implement `PermissionService`, and ship the hardcoded-preset resize spike to validate AX + `MenuBarExtra` before investing in the rest.
