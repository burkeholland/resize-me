import Foundation

struct AppConfig: Codable, Equatable, Sendable {
    static let defaultHotkey = "Ctrl+Alt+R"

    var schemaVersion: Int
    var presets: [Preset]
    var activePresetId: String
    var favoritePresetIds: [String]
    var centerAfterResize: Bool
    var hotkey: String
    var autoStart: Bool
    var firstRun: Bool

    init(schemaVersion: Int = 1,
         presets: [Preset] = [],
         activePresetId: String = "",
         favoritePresetIds: [String] = [],
         centerAfterResize: Bool = true,
         hotkey: String = "",
         autoStart: Bool = false,
         firstRun: Bool = true) {
        self.schemaVersion = schemaVersion
        self.presets = presets
        self.activePresetId = activePresetId
        self.favoritePresetIds = favoritePresetIds
        self.centerAfterResize = centerAfterResize
        self.hotkey = hotkey
        self.autoStart = autoStart
        self.firstRun = firstRun
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        schemaVersion = try container.decodeIfPresent(Int.self, forKey: .schemaVersion) ?? 1
        presets = try container.decodeIfPresent([Preset].self, forKey: .presets) ?? []
        activePresetId = try container.decodeIfPresent(String.self, forKey: .activePresetId) ?? ""
        favoritePresetIds = try container.decodeIfPresent([String].self, forKey: .favoritePresetIds) ?? []
        centerAfterResize = try container.decodeIfPresent(Bool.self, forKey: .centerAfterResize) ?? true
        hotkey = try container.decodeIfPresent(String.self, forKey: .hotkey) ?? ""
        autoStart = try container.decodeIfPresent(Bool.self, forKey: .autoStart) ?? false
        firstRun = try container.decodeIfPresent(Bool.self, forKey: .firstRun) ?? true
    }

    enum CodingKeys: String, CodingKey {
        case schemaVersion
        case presets
        case activePresetId
        case favoritePresetIds
        case centerAfterResize
        case hotkey
        case autoStart
        case firstRun
    }

    static var `default`: AppConfig {
        AppConfig(
            schemaVersion: 1,
            presets: [
                Preset(id: "360p-landscape", name: "360p Landscape", width: 640, height: 360),
                Preset(id: "480p-landscape", name: "480p Landscape", width: 854, height: 480),
                Preset(id: "540p-landscape", name: "540p Landscape", width: 960, height: 540),
                Preset(id: "720p-landscape", name: "720p Landscape", width: 1280, height: 720),
                Preset(id: "900p-landscape", name: "900p Landscape", width: 1600, height: 900),
                Preset(id: "1080p-landscape", name: "1080p Landscape", width: 1920, height: 1080),
                Preset(id: "1440p-landscape", name: "1440p Landscape", width: 2560, height: 1440),
                Preset(id: "1800p-landscape", name: "1800p Landscape", width: 3200, height: 1800),
                Preset(id: "4k-landscape", name: "4K Landscape", width: 3840, height: 2160),
                Preset(id: "360p-portrait", name: "360p Portrait", width: 360, height: 640),
                Preset(id: "480p-portrait", name: "480p Portrait", width: 480, height: 854),
                Preset(id: "540p-portrait", name: "540p Portrait", width: 540, height: 960),
                Preset(id: "720p-portrait", name: "720p Portrait", width: 720, height: 1280),
                Preset(id: "900p-portrait", name: "900p Portrait", width: 900, height: 1600),
                Preset(id: "1080p-portrait", name: "1080p Portrait", width: 1080, height: 1920),
                Preset(id: "1440p-portrait", name: "1440p Portrait", width: 1440, height: 2560),
                Preset(id: "1800p-portrait", name: "1800p Portrait", width: 1800, height: 3200),
                Preset(id: "4k-portrait", name: "4K Portrait", width: 2160, height: 3840)
            ],
            activePresetId: "1080p-landscape",
            favoritePresetIds: [],
            centerAfterResize: true,
            hotkey: AppConfig.defaultHotkey,
            autoStart: false,
            firstRun: true
        )
    }

    func findPreset(id: String) -> Preset? {
        presets.first(where: { $0.id == id })
    }

    func hasPreset(id: String) -> Bool {
        findPreset(id: id) != nil
    }

    var activePreset: Preset? {
        findPreset(id: activePresetId)
    }
}
