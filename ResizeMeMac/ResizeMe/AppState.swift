import AppKit
import Combine
import KeyboardShortcuts
import SwiftUI

@MainActor
final class AppState: ObservableObject {
    static let shared = AppState()

    @Published private(set) var config: AppConfig
    @Published private(set) var loadError: String?
    @Published var lastStatusMessage: String?
    @Published private(set) var frontmostAppName: String?

    let permissionService = PermissionService()
    let resizeService = ResizeService()
    let launchAtLoginService = LaunchAtLoginService()
    let hotkeyService = HotkeyService()
    let updateService = SparkleUpdateService()

    private let store: SettingsStore
    private var permissionCancellable: AnyCancellable?
    private var appActivationObserver: NSObjectProtocol?

    init(store: SettingsStore = SettingsStore()) {
        self.store = store
        let result = store.load()
        self.config = result.config
        self.loadError = result.loadError

        hotkeyService.onActivated = { [weak self] in
            self?.resizeNow()
        }
        permissionCancellable = permissionService.objectWillChange
            .sink { [weak self] _ in
                self?.objectWillChange.send()
            }

        hotkeyService.applyConfigString(config.hotkey)
        launchAtLoginService.reconcile(autoStart: config.autoStart)

        frontmostAppName = Self.externalAppName(NSWorkspace.shared.frontmostApplication)
        appActivationObserver = NSWorkspace.shared.notificationCenter.addObserver(
            forName: NSWorkspace.didActivateApplicationNotification,
            object: nil,
            queue: .main
        ) { [weak self] notification in
            let app = notification.userInfo?[NSWorkspace.applicationUserInfoKey] as? NSRunningApplication
            Task { @MainActor in
                guard let self else { return }
                // Ignore activations of ResizeMe itself (e.g. opening the menu or Settings),
                // so the menu keeps showing the app whose window would be resized.
                if let name = Self.externalAppName(app) {
                    self.frontmostAppName = name
                }
            }
        }
    }

    private static func externalAppName(_ app: NSRunningApplication?) -> String? {
        guard let app, app.processIdentifier != ProcessInfo.processInfo.processIdentifier else {
            return nil
        }
        return app.localizedName
    }

    func resizeNow() {
        guard let preset = config.activePreset else {
            lastStatusMessage = "No active preset"
            return
        }

        do {
            let outcome = try resizeService.resizeFrontmostWindow(to: preset, center: config.centerAfterResize)
            lastStatusMessage = outcome.isExact
                ? "Resized to \(preset.name)"
                : "Resized (constrained to \(Int(outcome.achieved.width))×\(Int(outcome.achieved.height)))"
        } catch let error as ResizeError {
            lastStatusMessage = friendlyMessage(for: error)
            if error == .permissionMissing {
                permissionService.refresh()
            }
        } catch {
            lastStatusMessage = "Resize failed"
        }
    }

    private func friendlyMessage(for error: ResizeError) -> String {
        switch error {
        case .permissionMissing:
            return "Accessibility permission required"
        case .noFrontmostApp:
            return "No frontmost app"
        case .noResizableWindow:
            return "No resizable window"
        case .windowFullscreen:
            return "Window is fullscreen"
        case .windowMinimized:
            return "Window is minimized"
        case .resizeRejected:
            return "App rejected the resize"
        }
    }

    func setActivePreset(_ id: String) {
        var next = config
        next.activePresetId = id
        _ = saveConfig(next)
    }

    func isFavoritePreset(_ id: String) -> Bool {
        config.favoritePresetIds.contains(id)
    }

    func toggleFavoritePreset(_ id: String) {
        guard config.hasPreset(id: id) else {
            lastStatusMessage = "Preset no longer exists"
            return
        }
        var next = config
        if let index = next.favoritePresetIds.firstIndex(of: id) {
            next.favoritePresetIds.remove(at: index)
        } else {
            next.favoritePresetIds.append(id)
        }
        _ = saveConfig(next)
    }

    @discardableResult
    func saveConfig(_ next: AppConfig) -> Bool {
        let current = config

        do {
            let normalized = try ConfigNormalizer.normalize(next, fallback: current)
            let previousHotkey = current.hotkey
            let previousAutoStart = current.autoStart

            hotkeyService.applyConfigString(normalized.hotkey)

            if normalized.autoStart != current.autoStart {
                do {
                    try launchAtLoginService.setEnabled(normalized.autoStart)
                } catch {
                    hotkeyService.applyConfigString(previousHotkey)
                    lastStatusMessage = error.localizedDescription
                    return false
                }
            }

            do {
                try store.save(normalized)
            } catch {
                hotkeyService.applyConfigString(previousHotkey)
                try? launchAtLoginService.setEnabled(previousAutoStart)
                lastStatusMessage = "Failed to save settings"
                return false
            }

            config = normalized
            loadError = nil
            return true
        } catch {
            lastStatusMessage = error.localizedDescription
            return false
        }
    }

    func completeFirstRun() {
        guard config.firstRun else { return }
        var next = config
        next.firstRun = false
        _ = saveConfig(next)
    }
}
