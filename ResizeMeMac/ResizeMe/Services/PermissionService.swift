import AppKit
import ApplicationServices

@MainActor
final class PermissionService: ObservableObject {
    @Published private(set) var isTrusted: Bool

    private var observer: NSObjectProtocol?
    private var refreshTask: Task<Void, Never>?

    init() {
        self.isTrusted = AXIsProcessTrusted()

        observer = NotificationCenter.default.addObserver(
            forName: NSApplication.didBecomeActiveNotification,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            Task { @MainActor in
                self?.refresh()
            }
        }
    }

    func refresh() {
        let trusted = AXIsProcessTrusted()
        if trusted != isTrusted {
            Log.permissions.notice("Accessibility permission status changed: \(trusted ? "trusted" : "not trusted")")
            isTrusted = trusted
        }
    }

    private func startRefreshPolling() {
        refreshTask?.cancel()
        refreshTask = Task { @MainActor [weak self] in
            guard let self else { return }

            let deadline = Date().addingTimeInterval(30)
            while !Task.isCancelled && Date() < deadline && !self.isTrusted {
                self.refresh()
                try? await Task.sleep(nanoseconds: 1_000_000_000)
            }

            self.refresh()
            self.refreshTask = nil
        }
    }

    func requestPermission() {
        let options = [
            kAXTrustedCheckOptionPrompt.takeUnretainedValue() as String: true
        ] as CFDictionary

        _ = AXIsProcessTrustedWithOptions(options)
        refresh()
        startRefreshPolling()
    }

    func openSystemSettings() {
        let primary = URL(string: "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility")!
        if !NSWorkspace.shared.open(primary) {
            _ = NSWorkspace.shared.open(URL(string: "x-apple.systempreferences:com.apple.preference.security")!)
        }

        startRefreshPolling()
    }
}
