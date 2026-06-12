import AppKit
import KeyboardShortcuts

enum HotkeyMapper {
    private static let keyLookup: [String: KeyboardShortcuts.Key] = [
        "A": .a, "B": .b, "C": .c, "D": .d, "E": .e, "F": .f, "G": .g, "H": .h, "I": .i, "J": .j,
        "K": .k, "L": .l, "M": .m, "N": .n, "O": .o, "P": .p, "Q": .q, "R": .r, "S": .s, "T": .t,
        "U": .u, "V": .v, "W": .w, "X": .x, "Y": .y, "Z": .z,
        "0": .zero, "1": .one, "2": .two, "3": .three, "4": .four, "5": .five, "6": .six, "7": .seven,
        "8": .eight, "9": .nine,
        "F1": .f1, "F2": .f2, "F3": .f3, "F4": .f4, "F5": .f5, "F6": .f6, "F7": .f7, "F8": .f8,
        "F9": .f9, "F10": .f10, "F11": .f11, "F12": .f12, "F13": .f13, "F14": .f14, "F15": .f15,
        "F16": .f16, "F17": .f17, "F18": .f18, "F19": .f19, "F20": .f20
    ]

    private static let reverseKeyLookup: [KeyboardShortcuts.Key: String] = Dictionary(uniqueKeysWithValues: keyLookup.map { ($0.value, $0.key) })

    static func shortcut(fromConfigString value: String) -> KeyboardShortcuts.Shortcut? {
        let parts = value
            .split(whereSeparator: { $0 == "+" || $0 == " " || $0 == "-" })
            .map(String.init)

        var modifiers = NSEvent.ModifierFlags()
        var keyToken: String?

        for part in parts {
            let token = part.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
            switch token {
            case "ctrl", "control":
                modifiers.insert(.control)
            case "alt", "option":
                modifiers.insert(.option)
            case "shift":
                modifiers.insert(.shift)
            case "win", "windows", "cmd", "command", "meta":
                modifiers.insert(.command)
            default:
                if keyToken == nil {
                    keyToken = part.trimmingCharacters(in: .whitespacesAndNewlines).uppercased()
                }
            }
        }

        guard let keyToken, !keyToken.isEmpty, !modifiers.isEmpty, let key = keyLookup[keyToken] else {
            return nil
        }

        return KeyboardShortcuts.Shortcut(key, modifiers: modifiers)
    }

    static func configString(from shortcut: KeyboardShortcuts.Shortcut?) -> String {
        guard let shortcut else { return "" }

        var tokens: [String] = []
        if shortcut.modifiers.contains(.control) { tokens.append("Ctrl") }
        if shortcut.modifiers.contains(.option) { tokens.append("Alt") }
        if shortcut.modifiers.contains(.shift) { tokens.append("Shift") }
        if shortcut.modifiers.contains(.command) { tokens.append("Win") }

        guard let keyToken = shortcut.key.flatMap({ reverseKeyLookup[$0] }) else {
            return ""
        }

        tokens.append(keyToken)
        return tokens.joined(separator: "+")
    }

    static func displayString(from shortcut: KeyboardShortcuts.Shortcut?) -> String {
        guard let shortcut else { return "None" }

        let symbols = [
            shortcut.modifiers.contains(.control) ? "⌃" : nil,
            shortcut.modifiers.contains(.option) ? "⌥" : nil,
            shortcut.modifiers.contains(.shift) ? "⇧" : nil,
            shortcut.modifiers.contains(.command) ? "⌘" : nil
        ].compactMap { $0 }.joined()

        guard let keyToken = shortcut.key.flatMap({ reverseKeyLookup[$0] }) else {
            return symbols.isEmpty ? "None" : symbols
        }

        return symbols + keyToken.uppercased()
    }
}
