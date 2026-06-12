import Foundation
import KeyboardShortcuts

@MainActor
final class HotkeyService {
    private var listenerTask: Task<Void, Never>?
    private(set) var currentShortcut: KeyboardShortcuts.Shortcut?
    var onActivated: (() -> Void)?

    func apply(shortcut: KeyboardShortcuts.Shortcut?) {
        listenerTask?.cancel()
        listenerTask = nil
        currentShortcut = shortcut

        guard let shortcut else {
            Log.hotkeys.notice("Hotkey cleared; listener stopped")
            return
        }

        listenerTask = Task { [weak self] in
            for await _ in KeyboardShortcuts.events(.keyUp, for: shortcut) {
                guard !Task.isCancelled else { break }
                await MainActor.run {
                    self?.onActivated?()
                }
            }
        }

        Log.hotkeys.notice("Hotkey listener started")
    }

    func applyConfigString(_ value: String) {
        apply(shortcut: HotkeyMapper.shortcut(fromConfigString: value))
    }
}
