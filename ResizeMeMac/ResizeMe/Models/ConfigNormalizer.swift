import Foundation

enum ConfigNormalizerError: LocalizedError, Equatable {
    case dimensionOutOfRange(presetName: String, dimension: String)

    var errorDescription: String? {
        switch self {
        case let .dimensionOutOfRange(presetName, dimension):
            return "\(presetName) \(dimension) must be between 100 and 10000"
        }
    }
}

enum ConfigNormalizer {
    static func normalize(_ config: AppConfig, fallback: AppConfig) throws -> AppConfig {
        var next = config
        next.schemaVersion = 1

        let trimmedHotkey = next.hotkey.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmedHotkey.isEmpty {
            next.hotkey = AppConfig.defaultHotkey
        }
        next.hotkey = normalizeHotkeyText(next.hotkey)
        if !isValidHotkeyText(next.hotkey) {
            next.hotkey = AppConfig.defaultHotkey
        }

        if next.presets.isEmpty {
            next.presets = fallback.presets
        }

        var seen = Set<String>()
        for index in next.presets.indices {
            var preset = next.presets[index]
            preset.name = preset.name.trimmingCharacters(in: .whitespacesAndNewlines)
            if preset.name.isEmpty {
                preset.name = "\(preset.width)x\(preset.height)"
            }
            if preset.width < 100 || preset.width > 10000 {
                throw ConfigNormalizerError.dimensionOutOfRange(presetName: preset.name, dimension: "width")
            }
            if preset.height < 100 || preset.height > 10000 {
                throw ConfigNormalizerError.dimensionOutOfRange(presetName: preset.name, dimension: "height")
            }
            if preset.id.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                preset.id = presetID(preset, index: index)
            }
            let baseID = preset.id
            var suffix = 2
            while seen.contains(preset.id) {
                preset.id = "\(baseID)-\(suffix)"
                suffix += 1
            }
            seen.insert(preset.id)
            next.presets[index] = preset
        }

        if !next.hasPreset(id: next.activePresetId) {
            if !fallback.activePresetId.isEmpty && next.hasPreset(id: fallback.activePresetId) {
                next.activePresetId = fallback.activePresetId
            } else if !next.presets.isEmpty {
                next.activePresetId = next.presets[0].id
            }
        }

        var seenFavoriteIDs = Set<String>()
        next.favoritePresetIds = next.favoritePresetIds.filter { id in
            guard next.hasPreset(id: id), !seenFavoriteIDs.contains(id) else {
                return false
            }
            seenFavoriteIDs.insert(id)
            return true
        }

        return next
    }

    static func normalizeHotkeyText(_ value: String) -> String {
        let parts = value.split(whereSeparator: { $0 == "+" || $0 == " " || $0 == "-" })
        var modifiers: Set<String> = []
        var key = ""
        for part in parts {
            let normalized = String(part).lowercased().trimmingCharacters(in: .whitespacesAndNewlines)
            switch normalized {
            case "", "plus":
                continue
            case "ctrl", "control":
                modifiers.insert("Ctrl")
            case "alt", "option":
                modifiers.insert("Alt")
            case "shift":
                modifiers.insert("Shift")
            case "win", "windows", "cmd", "meta":
                modifiers.insert("Win")
            default:
                key = normalized.uppercased()
            }
        }

        var ordered = [String]()
        for modifier in ["Ctrl", "Alt", "Shift", "Win"] {
            if modifiers.contains(modifier) {
                ordered.append(modifier)
            }
        }
        if !key.isEmpty {
            ordered.append(key)
        }
        return ordered.joined(separator: "+")
    }

    static func isValidHotkeyText(_ value: String) -> Bool {
        let parts = value.split(separator: "+")
        var hasModifier = false
        var hasKey = false

        for part in parts {
            let partString = String(part)
            switch partString {
            case "Ctrl", "Alt", "Shift", "Win":
                hasModifier = true
            default:
                if partString.count == 1,
                   let ch = partString.unicodeScalars.first,
                   (ch >= "A" && ch <= "Z") || (ch >= "0" && ch <= "9") {
                    hasKey = true
                } else if partString.hasPrefix("F") {
                    let rest = String(partString.dropFirst())
                    if let n = Int(rest), n >= 1 && n <= 24 {
                        hasKey = true
                    }
                }
            }
        }

        return hasModifier && hasKey
    }

    private static func presetID(_ preset: Preset, index: Int) -> String {
        var base = preset.name.lowercased().trimmingCharacters(in: .whitespacesAndNewlines)
        let pattern = "[^a-z0-9]+"
        if let regex = try? NSRegularExpression(pattern: pattern) {
            base = regex.stringByReplacingMatches(
                in: base,
                range: NSRange(base.startIndex..<base.endIndex, in: base),
                withTemplate: "-"
            )
        }
        base = base.trimmingCharacters(in: CharacterSet(charactersIn: "-"))
        if base.isEmpty {
            return "preset-\(index + 1)"
        }
        return base
    }
}
