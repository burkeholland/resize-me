import XCTest
@testable import ResizeMe

final class ConfigNormalizerTests: XCTestCase {
    func testDefaultPresetParity() {
        let config = AppConfig.default

        XCTAssertEqual(config.presets.count, 18)
        XCTAssertEqual(config.presets.first?.id, "360p-landscape")
        XCTAssertEqual(config.presets.first?.width, 640)
        XCTAssertEqual(config.presets.first?.height, 360)
        XCTAssertEqual(config.presets.first(where: { $0.id == "1080p-landscape" })?.width, 1920)
        XCTAssertEqual(config.presets.first(where: { $0.id == "1080p-landscape" })?.height, 1080)
        XCTAssertEqual(config.activePresetId, "1080p-landscape")
        XCTAssertEqual(config.presets.first(where: { $0.id == "4k-portrait" })?.width, 2160)
        XCTAssertEqual(config.presets.first(where: { $0.id == "4k-portrait" })?.height, 3840)
        XCTAssertEqual(config.hotkey, "Ctrl+Alt+R")
        XCTAssertTrue(config.centerAfterResize)
        XCTAssertTrue(config.firstRun)
        XCTAssertFalse(config.autoStart)
    }

    func testEmptyHotkeyFallsBackToDefault() {
        let config = AppConfig(hotkey: "")
        let normalized = try? ConfigNormalizer.normalize(config, fallback: .default)
        XCTAssertEqual(normalized?.hotkey, AppConfig.defaultHotkey)
    }

    func testHotkeyNormalization() {
        XCTAssertEqual(ConfigNormalizer.normalizeHotkeyText("control option r"), "Ctrl+Alt+R")
        XCTAssertEqual(ConfigNormalizer.normalizeHotkeyText("cmd+shift+5"), "Shift+Win+5")
        XCTAssertEqual(ConfigNormalizer.normalizeHotkeyText("ctrl-alt-f11"), "Ctrl+Alt+F11")
    }

    func testInvalidHotkeyFallsBack() {
        XCTAssertEqual((try? ConfigNormalizer.normalize(AppConfig(hotkey: "Ctrl+Alt"), fallback: .default))?.hotkey, AppConfig.defaultHotkey)
        XCTAssertEqual((try? ConfigNormalizer.normalize(AppConfig(hotkey: "R"), fallback: .default))?.hotkey, AppConfig.defaultHotkey)
        XCTAssertEqual((try? ConfigNormalizer.normalize(AppConfig(hotkey: "Ctrl+Alt+F25"), fallback: .default))?.hotkey, AppConfig.defaultHotkey)
    }

    func testEmptyPresetsUsesFallback() {
        let normalized = try? ConfigNormalizer.normalize(AppConfig(presets: []), fallback: .default)
        XCTAssertEqual(normalized?.presets, AppConfig.default.presets)
    }

    func testEmptyNameBecomesWxH() {
        let normalized = try? ConfigNormalizer.normalize(AppConfig(presets: [Preset(id: "", name: "  ", width: 800, height: 600)]), fallback: .default)
        XCTAssertEqual(normalized?.presets.first?.name, "800x600")
    }

    func testWidthOutOfRangeThrows() {
        do {
            _ = try ConfigNormalizer.normalize(AppConfig(presets: [Preset(id: "x", name: "Bad", width: 99, height: 600)]), fallback: .default)
            XCTFail("Expected error")
        } catch let error as ConfigNormalizerError {
            XCTAssertTrue(error.localizedDescription.contains("width must be between 100 and 10000"))
        } catch {
            XCTFail("Unexpected error: \(error)")
        }

        do {
            _ = try ConfigNormalizer.normalize(AppConfig(presets: [Preset(id: "x", name: "Bad", width: 10001, height: 600)]), fallback: .default)
            XCTFail("Expected error")
        } catch let error as ConfigNormalizerError {
            XCTAssertTrue(error.localizedDescription.contains("width must be between 100 and 10000"))
        } catch {
            XCTFail("Unexpected error: \(error)")
        }

        do {
            _ = try ConfigNormalizer.normalize(AppConfig(presets: [Preset(id: "x", name: "Bad", width: 800, height: 99)]), fallback: .default)
            XCTFail("Expected error")
        } catch let error as ConfigNormalizerError {
            XCTAssertTrue(error.localizedDescription.contains("height must be between 100 and 10000"))
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testEmptyIDSlugified() {
        let normalized = try? ConfigNormalizer.normalize(AppConfig(presets: [Preset(id: "", name: "My Cool Preset!", width: 800, height: 600)]), fallback: .default)
        XCTAssertEqual(normalized?.presets.first?.id, "my-cool-preset")
    }

    func testEmptyIDAndUnsluggableName() {
        let normalized = try? ConfigNormalizer.normalize(AppConfig(presets: [Preset(id: "", name: "!!!", width: 800, height: 600)]), fallback: .default)
        XCTAssertEqual(normalized?.presets.first?.name, "!!!")
        XCTAssertEqual(normalized?.presets.first?.id, "preset-1")
    }

    func testDuplicateIDsSuffixed() {
        let normalized = try? ConfigNormalizer.normalize(AppConfig(presets: [
            Preset(id: "dup", name: "One", width: 800, height: 600),
            Preset(id: "dup", name: "Two", width: 900, height: 700),
            Preset(id: "dup", name: "Three", width: 1000, height: 800)
        ]), fallback: .default)

        XCTAssertEqual(normalized?.presets[1].id, "dup-2")
        XCTAssertEqual(normalized?.presets[2].id, "dup-3")
    }

    func testUnknownActivePresetFallsBack() {
        let config = AppConfig(presets: [Preset(id: "a", name: "A", width: 800, height: 600)], activePresetId: "nope")
        let normalized = try? ConfigNormalizer.normalize(config, fallback: .default)
        XCTAssertEqual(normalized?.activePresetId, "a")

        let config2 = AppConfig(presets: [Preset(id: "other", name: "Other", width: 800, height: 600)], activePresetId: "nope")
        let normalized2 = try? ConfigNormalizer.normalize(config2, fallback: AppConfig.default)
        XCTAssertEqual(normalized2?.activePresetId, "other")
    }

    func testValidConfigPassesThroughUnchanged() {
        let config = AppConfig(schemaVersion: 7, presets: [Preset(id: "x", name: "X", width: 800, height: 600)], activePresetId: "x", centerAfterResize: false, hotkey: "Ctrl+Shift+X", autoStart: true, firstRun: false)
        let normalized = try? ConfigNormalizer.normalize(config, fallback: .default)

        XCTAssertEqual(normalized?.schemaVersion, 1)
        XCTAssertEqual(normalized?.presets, config.presets)
        XCTAssertEqual(normalized?.activePresetId, "x")
        XCTAssertEqual(normalized?.hotkey, "Ctrl+Shift+X")
        XCTAssertFalse(normalized?.centerAfterResize ?? true)
    }
}
